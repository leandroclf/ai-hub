// Package providerauth implementa autenticacao outbound dos chamados do Hub.
// Em local/dev os segredos sao referencias e o provider-sim valida o contrato
// sem receber segredo em claro. O cache de access token usa Redis com TTL.
package providerauth

import (
	"ai-hub/hub/internal/platform/egress"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	None                   = "NONE"
	APIKey                 = "API_KEY"
	Basic                  = "BASIC"
	OAuthClientCredentials = "OAUTH_CLIENT_CREDENTIALS"
	MTLSOAuth              = "MTLS_OAUTH"
)

type Config struct {
	AuthType           string
	SecretVersion      string
	BindingID          string
	TenantID           string
	Environment        string
	APIKeyHeader       string
	Username           string
	SecretRef          string
	TokenURL           string
	ClientID           string
	ClientSecretRef    string
	MTLSCertificateRef string
	TokenTTLSeconds    int
}

type TokenCache struct {
	client   *redis.Client
	Resolver SecretResolver
	mu       sync.Mutex
	locksMu  sync.Mutex
	locks    map[string]*sync.Mutex
	tokens   map[string]cachedToken
}
type cachedToken struct {
	value   string
	expires time.Time
}

func NewTokenCache(addr string) *TokenCache {
	return &TokenCache{client: redis.NewClient(&redis.Options{Addr: addr, DialTimeout: 100 * time.Millisecond, ReadTimeout: 100 * time.Millisecond, WriteTimeout: 100 * time.Millisecond, MaxRetries: -1}), Resolver: AWSVault{}, tokens: map[string]cachedToken{}, locks: map[string]*sync.Mutex{}}
}

func (c *TokenCache) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }

func (c *TokenCache) Close() error { return c.client.Close() }

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *TokenCache) bearer(ctx context.Context, httpClient *http.Client, providerAccountID string, cfg Config) (string, error) {
	if c == nil || c.Resolver == nil {
		return "", fmt.Errorf("secret resolver unavailable")
	}
	secret, err := c.Resolver.Resolve(ctx, cfg.ClientSecretRef, cfg.SecretVersion)
	if err != nil {
		return "", err
	}
	material := strings.Join([]string{cfg.Environment, cfg.TenantID, cfg.BindingID, providerAccountID, cfg.TokenURL, cfg.ClientID, secret.Version}, "\x1f")
	hash := sha256.Sum256([]byte(material))
	key := "hub:provider-token:" + hex.EncodeToString(hash[:])
	keyLock := c.lockFor(key)
	keyLock.Lock()
	defer keyLock.Unlock()
	c.mu.Lock()
	cached, ok := c.tokens[key]
	c.mu.Unlock()
	if ok && time.Now().Before(cached.expires) {
		return cached.value, nil
	}
	// Redis nunca armazena bearer tokens. Ele permanece disponível somente
	// para saúde/compatibilidade operacional; a autoridade do token é o L1
	// vinculado à conta, binding, tenant, ambiente e versão do segredo.
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {cfg.ClientID}, "client_secret": {secret.Value}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("invalid token endpoint")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenClient, err := egress.NewClient(cfg.TokenURL, 3*time.Second)
	if err != nil {
		return "", err
	}
	if cfg.AuthType == MTLSOAuth {
		if err = c.applyCertificate(ctx, tokenClient, cfg); err != nil {
			return "", err
		}
	}
	resp, err := tokenClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("provider token endpoint unavailable")
	}
	defer resp.Body.Close()
	var out tokenResponse
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 32*1024)).Decode(&out) != nil || out.AccessToken == "" || out.ExpiresIn <= 5 {
		return "", fmt.Errorf("invalid provider token response")
	}
	ttl := out.ExpiresIn - 5
	if cfg.TokenTTLSeconds > 0 && cfg.TokenTTLSeconds < ttl {
		ttl = cfg.TokenTTLSeconds
	}
	expiry := time.Now().Add(time.Duration(ttl) * time.Second)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tokens == nil {
		c.tokens = map[string]cachedToken{}
	}
	if len(c.tokens) >= 1024 {
		for k, v := range c.tokens {
			if !time.Now().Before(v.expires) {
				delete(c.tokens, k)
			}
		}
	}
	if len(c.tokens) < 1024 {
		c.tokens[key] = cachedToken{out.AccessToken, expiry}
	}
	return out.AccessToken, nil
}

func (c *TokenCache) lockFor(key string) *sync.Mutex {
	c.locksMu.Lock()
	defer c.locksMu.Unlock()
	if c.locks == nil {
		c.locks = map[string]*sync.Mutex{}
	}
	if lock, ok := c.locks[key]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	c.locks[key] = lock
	return lock
}

func (c *TokenCache) applyCertificate(ctx context.Context, client *http.Client, cfg Config) error {
	if c == nil || c.Resolver == nil {
		return fmt.Errorf("certificate resolver unavailable")
	}
	secret, err := c.Resolver.Resolve(ctx, cfg.MTLSCertificateRef, cfg.SecretVersion)
	if err != nil {
		return err
	}
	var pair struct {
		Certificate string `json:"certificate"`
		PrivateKey  string `json:"private_key"`
	}
	if json.Unmarshal([]byte(secret.Value), &pair) != nil {
		return fmt.Errorf("invalid certificate secret")
	}
	certificate, err := tls.X509KeyPair([]byte(pair.Certificate), []byte(pair.PrivateKey))
	if err != nil {
		return fmt.Errorf("invalid certificate pair")
	}
	return egress.WithCertificate(client, certificate, nil)
}

func (c *TokenCache) Apply(ctx context.Context, httpClient *http.Client, providerAccountID string, cfg Config, req *http.Request) error {
	switch cfg.AuthType {
	case "", None:
		return nil
	case Basic:
		if cfg.Username == "" || cfg.SecretRef == "" {
			return fmt.Errorf("basic auth incompleta")
		}
		if c == nil || c.Resolver == nil {
			return fmt.Errorf("secret resolver unavailable")
		}
		secret, err := c.Resolver.Resolve(ctx, cfg.SecretRef, cfg.SecretVersion)
		if err != nil {
			return err
		}
		req.SetBasicAuth(cfg.Username, secret.Value)
		return nil
	case APIKey:
		if c == nil || c.Resolver == nil {
			return fmt.Errorf("secret resolver unavailable")
		}
		if cfg.APIKeyHeader != "X-API-Key" && cfg.APIKeyHeader != "Authorization" {
			return fmt.Errorf("invalid API key header")
		}
		secret, err := c.Resolver.Resolve(ctx, cfg.SecretRef, cfg.SecretVersion)
		if err != nil {
			return err
		}
		req.Header.Set(cfg.APIKeyHeader, secret.Value)
		return nil
	case OAuthClientCredentials, MTLSOAuth:
		if c == nil {
			return fmt.Errorf("secret resolver unavailable")
		}
		token, err := c.bearer(ctx, httpClient, providerAccountID, cfg)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if cfg.AuthType == MTLSOAuth {
			if err := c.applyCertificate(ctx, httpClient, cfg); err != nil {
				return err
			}
		}

		return nil
	default:
		return fmt.Errorf("tipo de autenticacao não suportado: %s", cfg.AuthType)
	}
}
