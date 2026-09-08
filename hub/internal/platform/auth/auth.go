// Package auth verifies access tokens at each service boundary. Caller supplied
// tenant headers never participate in identity or authorization.
package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

var ErrInvalidToken = errors.New("identidade inválida ou expirada")

type Principal struct {
	Subject       string    `json:"subject"`
	TenantID      string    `json:"tenant_id"`
	ApplicationID string    `json:"application_id"`
	CellID        string    `json:"cell_id"`
	Scopes        []string  `json:"scopes"`
	Roles         []string  `json:"roles"`
	MFA           bool      `json:"mfa"`
	Workload      bool      `json:"workload"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type principalKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok && p.Subject != "" && time.Now().Before(p.ExpiresAt)
}

func (p Principal) HasScope(scope string) bool { return slices.Contains(p.Scopes, scope) }
func (p Principal) HasRole(role string) bool   { return slices.Contains(p.Roles, role) }

// Authorize never grants cross-tenant access merely because a user is a reader.
// Global protocol diagnostics use a separate, audited domain operation.
func Authorize(ctx context.Context, scope, tenant string) bool {
	p, ok := FromContext(ctx)
	if !ok || !p.HasScope(scope) {
		return false
	}
	if p.Workload {
		return p.CellID != "" && p.CellID == os.Getenv("CELL_ID")
	}
	if tenant == "" {
		return p.MFA
	}
	return p.TenantID == tenant || (p.MFA && p.HasScope("admin:cross_tenant"))
}

func PublicTenant(ctx context.Context, scope string) (string, bool) {
	p, ok := FromContext(ctx)
	return p.TenantID, ok && !p.Workload && p.TenantID != "" && p.ApplicationID != "" && p.HasScope(scope)
}

type Config struct {
	Issuer     string
	JWKSURL    string
	Audience   string
	HTTPClient *http.Client
	KeyTTL     time.Duration
}

type Verifier struct {
	cfg       Config
	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	expires   time.Time
	lastFetch time.Time
}

func New(cfg Config) *Verifier {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	if cfg.KeyTTL <= 0 || cfg.KeyTTL > time.Minute {
		cfg.KeyTTL = time.Minute
	}
	return &Verifier{cfg: cfg, keys: map[string]*rsa.PublicKey{}}
}

func FromEnv() *Verifier {
	return New(Config{Issuer: os.Getenv("OIDC_ISSUER"), JWKSURL: os.Getenv("OIDC_JWKS_URL"), Audience: os.Getenv("OIDC_AUDIENCE")})
}

type claims struct {
	Issuer        string          `json:"iss"`
	Subject       string          `json:"sub"`
	Audience      json.RawMessage `json:"aud"`
	Expires       int64           `json:"exp"`
	NotBefore     int64           `json:"nbf"`
	Issued        int64           `json:"iat"`
	TokenType     string          `json:"typ"`
	TenantID      string          `json:"tenant_id"`
	ApplicationID string          `json:"application_id"`
	CellID        string          `json:"cell_id"`
	Scope         string          `json:"scope"`
	AMR           []string        `json:"amr"`
	Workload      bool            `json:"workload"`
	RealmAccess   struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

func (v *Verifier) Verify(ctx context.Context, token string) (Principal, error) {
	if len(token) > 16384 || v.cfg.Issuer == "" || v.cfg.Audience == "" || v.cfg.JWKSURL == "" {
		return Principal{}, ErrInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Principal{}, ErrInvalidToken
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	var h struct {
		Algorithm string   `json:"alg"`
		KeyID     string   `json:"kid"`
		Critical  []string `json:"crit"`
		Type      string   `json:"typ"`
	}
	if json.Unmarshal(header, &h) != nil || h.Algorithm != "RS256" || h.KeyID == "" || len(h.Critical) > 0 || (h.Type != "JWT" && h.Type != "at+jwt") {
		return Principal{}, ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	key, err := v.key(ctx, h.KeyID)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig) != nil {
		return Principal{}, ErrInvalidToken
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	var c claims
	if json.Unmarshal(body, &c) != nil {
		return Principal{}, ErrInvalidToken
	}
	now := time.Now().Unix()
	var audiences []string
	if json.Unmarshal(c.Audience, &audiences) != nil {
		var a string
		if json.Unmarshal(c.Audience, &a) != nil {
			return Principal{}, ErrInvalidToken
		}
		audiences = []string{a}
	}
	if c.Issuer != v.cfg.Issuer || c.Subject == "" || !slices.Contains(audiences, v.cfg.Audience) || c.Expires <= now || c.NotBefore > now || c.Issued > now || c.Expires <= c.Issued || c.TokenType == "ID" {
		return Principal{}, ErrInvalidToken
	}
	return Principal{Subject: c.Subject, TenantID: c.TenantID, ApplicationID: c.ApplicationID, CellID: c.CellID, Scopes: strings.Fields(c.Scope), Roles: c.RealmAccess.Roles, MFA: slices.Contains(c.AMR, "mfa") || slices.Contains(c.AMR, "otp"), Workload: c.Workload, ExpiresAt: time.Unix(c.Expires, 0)}, nil
}

func (v *Verifier) key(ctx context.Context, id string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	now := time.Now()
	if key, ok := v.keys[id]; ok && now.Before(v.expires) {
		return key, nil
	}
	// Unknown kids cannot force an unbounded request storm against the IdP.
	if now.Sub(v.lastFetch) < time.Second {
		return nil, ErrInvalidToken
	}
	v.lastFetch = now
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.cfg.JWKSURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidToken
	}
	var jwks struct {
		Keys []struct {
			KeyID string `json:"kid"`
			Type  string `json:"kty"`
			Use   string `json:"use"`
			Alg   string `json:"alg"`
			N     string `json:"n"`
			E     string `json:"e"`
		} `json:"keys"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 128*1024)).Decode(&jwks) != nil || len(jwks.Keys) > 32 {
		return nil, ErrInvalidToken
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range jwks.Keys {
		if k.Type != "RSA" || k.Use != "sig" || (k.Alg != "" && k.Alg != "RS256") || k.KeyID == "" {
			continue
		}
		n, ne := base64.RawURLEncoding.DecodeString(k.N)
		e, ee := base64.RawURLEncoding.DecodeString(k.E)
		if ne != nil || ee != nil || len(e) == 0 || len(e) > 4 || len(n) < 256 || len(n) > 1024 {
			continue
		}
		exponent := 0
		for _, b := range e {
			exponent = exponent*256 + int(b)
		}
		if exponent < 3 || exponent%2 == 0 {
			continue
		}
		if _, duplicate := keys[k.KeyID]; duplicate {
			return nil, ErrInvalidToken
		}
		keys[k.KeyID] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}
	}
	v.keys = keys
	v.expires = now.Add(v.cfg.KeyTTL)
	key, ok := keys[id]
	if !ok {
		return nil, ErrInvalidToken
	}
	return key, nil
}

func Error(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": http.StatusText(status)})
}

func (v *Verifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			Error(w, 401, "unauthorized")
			return
		}
		p, err := v.Verify(r.Context(), parts[1])
		if err != nil {
			Error(w, 401, "unauthorized")
			return
		}
		r.Header.Del("X-Tenant-Id")
		r.Header.Del("X-Application-Id")
		r.Header.Del("X-Cell-Id")
		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
	})
}

func Me(w http.ResponseWriter, r *http.Request) {
	p, ok := FromContext(r.Context())
	if !ok {
		Error(w, 401, "unauthorized")
		return
	}
	if r.Method != http.MethodGet {
		Error(w, 405, "method_not_allowed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(struct {
		Principal
		Environment string `json:"environment"`
	}{p, os.Getenv("ENVIRONMENT")})
}
