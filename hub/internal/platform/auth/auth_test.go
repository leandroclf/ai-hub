package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T) (*Verifier, func(map[string]any) string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]string{"kid": "fixture", "kty": "RSA", "use": "sig", "alg": "RS256", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": "AQAB"}}})
	}))
	t.Cleanup(srv.Close)
	v := New(Config{Issuer: "https://fixture.invalid", Audience: "ai-hub", JWKSURL: srv.URL})
	sign := func(c map[string]any) string {
		b, _ := json.Marshal(c)
		h := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"fixture","typ":"JWT"}`))
		body := h + "." + base64.RawURLEncoding.EncodeToString(b)
		hash := sha256.Sum256([]byte(body))
		sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		if err != nil {
			t.Fatal(err)
		}
		return body + "." + base64.RawURLEncoding.EncodeToString(sig)
	}
	return v, sign
}

func validClaims() map[string]any {
	return map[string]any{"iss": "https://fixture.invalid", "aud": []string{"ai-hub"}, "sub": "person-a", "tenant_id": "acme", "application_id": "app-acme", "iat": time.Now().Unix(), "exp": time.Now().Add(time.Minute).Unix(), "scope": "protocols:read protocols:write", "amr": []string{"pwd", "otp"}}
}

func TestAccessTokenClaimsAndSignature(t *testing.T) {
	v, sign := fixture(t)
	for _, tc := range []struct {
		name  string
		field string
		value any
	}{{"expired", "exp", time.Now().Unix()}, {"issuer", "iss", "https://evil.invalid"}, {"audience", "aud", []string{"another-service"}}, {"future", "nbf", time.Now().Add(time.Hour).Unix()}, {"subject", "sub", ""}, {"ID token", "typ", "ID"}} {
		t.Run(tc.name, func(t *testing.T) {
			c := validClaims()
			c[tc.field] = tc.value
			if _, err := v.Verify(context.Background(), sign(c)); err == nil {
				t.Fatal("invalid claim accepted")
			}
		})
	}
	valid := sign(validClaims())
	p, err := v.Verify(context.Background(), valid)
	if err != nil || p.TenantID != "acme" || !p.MFA {
		t.Fatalf("valid identity rejected: %v", err)
	}
	parts := strings.Split(valid, ".")
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"tenant_id":"beta"}`))
	if _, err := v.Verify(context.Background(), strings.Join(parts, ".")); err == nil {
		t.Fatal("forged identity accepted")
	}
	if _, err := v.Verify(context.Background(), "eyJhbGciOiJub25lIn0.e30."); err == nil {
		t.Fatal("unsigned token accepted")
	}
}

func TestHeaderCannotGrantTenantAndScopes(t *testing.T) {
	v, sign := fixture(t)
	called := false
	h := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		tenant, ok := PublicTenant(r.Context(), "protocols:read")
		if !ok || tenant != "acme" || r.Header.Get("X-Tenant-Id") != "" {
			t.Fatal("header affected trusted identity")
		}
		if Authorize(r.Context(), "protocols:read", "beta") {
			t.Fatal("cross-tenant granted")
		}
		if Authorize(r.Context(), "finance:write", "acme") {
			t.Fatal("scope escalation")
		}
	}))
	r := httptest.NewRequest("GET", "/v1/protocols/id", nil)
	r.Header.Set("X-Tenant-Id", "beta")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 || called {
		t.Fatal("unauthenticated request reached handler")
	}
	r.Header.Set("Authorization", "Bearer "+sign(validClaims()))
	h.ServeHTTP(httptest.NewRecorder(), r)
	if !called {
		t.Fatal("valid token did not reach handler")
	}
}

func TestReaderDoesNotGrantWritesAndWorkloadIsCellBound(t *testing.T) {
	t.Setenv("CELL_ID", "r2-cell-a")
	p := Principal{Subject: "individual", MFA: true, Roles: []string{"hub_protocol_reader"}, Scopes: []string{"protocols:read"}, TenantID: "acme", ExpiresAt: time.Now().Add(time.Minute)}
	ctx := WithPrincipal(context.Background(), p)
	if Authorize(ctx, "catalog:publish", "acme") || Authorize(ctx, "protocols:read", "beta") {
		t.Fatal("reader role expanded authority")
	}
	p.Workload = true
	p.CellID = "r2-cell-b"
	p.Scopes = []string{"cometa:execute"}
	if Authorize(WithPrincipal(context.Background(), p), "cometa:execute", "acme") {
		t.Fatal("foreign cell accepted")
	}
	p.CellID = "r2-cell-a"
	if !Authorize(WithPrincipal(context.Background(), p), "cometa:execute", "acme") {
		t.Fatal("authorized workload rejected")
	}
	p.ExpiresAt = time.Now().Add(-time.Second)
	if Authorize(WithPrincipal(context.Background(), p), "cometa:execute", "acme") {
		t.Fatal("expired context accepted")
	}
}

func TestPublicErrorsCarryCorrelationWithoutInternalDetails(t *testing.T) {
	recorder := httptest.NewRecorder()
	ErrorWithMessage(recorder, http.StatusServiceUnavailable, "catalog_unavailable", "catálogo indisponível; tente novamente")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", recorder.Code)
	}
	correlation := recorder.Header().Get("X-Trace-Id")
	if correlation == "" || recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("resposta sem correlação/cache-control: headers=%v", recorder.Header())
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "catalog_unavailable" || body["correlation_id"] != correlation {
		t.Fatalf("envelope inválido: %#v", body)
	}
	for _, secret := range []string{"postgres://", "token", "stack", "password"} {
		if strings.Contains(recorder.Body.String(), secret) {
			t.Fatalf("detalhe interno exposto: %q", secret)
		}
	}

	withExisting := httptest.NewRecorder()
	withExisting.Header().Set("X-Trace-Id", "trace-fixture")
	Error(withExisting, http.StatusForbidden, "forbidden")
	if withExisting.Header().Get("X-Trace-Id") != "trace-fixture" {
		t.Fatal("correlação previamente estabelecida foi substituída")
	}
}

func TestR2Seg05S03BackendErrorScenario(t *testing.T) {
	recorder := httptest.NewRecorder()
	ErrorWithMessage(recorder, http.StatusServiceUnavailable, "catalog_unavailable", "catálogo indisponível; tente novamente")
	body := recorder.Body.String()
	for _, forbidden := range []string{"postgres://", "Bearer ", "stack", "password", "secret"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Fatalf("detalhe interno vazou no erro público: %q", forbidden)
		}
	}
	if recorder.Code != http.StatusServiceUnavailable || recorder.Header().Get("X-Trace-Id") == "" || !strings.Contains(body, `"correlation_id"`) {
		t.Fatalf("envelope de erro sem correlação: status=%d headers=%v body=%s", recorder.Code, recorder.Header(), body)
	}
}
