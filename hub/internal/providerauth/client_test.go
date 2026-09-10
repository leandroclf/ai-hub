package providerauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type testResolver map[string]Secret

func (r testResolver) Resolve(_ context.Context, ref, version string) (Secret, error) {
	s, ok := r[ref]
	if !ok || (version != "" && s.Version != version) {
		return Secret{}, errors.New("secret unavailable")
	}
	return s, nil
}

func TestBasicUsesDecryptedVersionWithoutFallback(t *testing.T) {
	c := NewTokenCache("127.0.0.1:1")
	defer c.Close()
	c.Resolver = testResolver{"dedicated": {Value: "synthetic-password", Version: "v1"}, "shared": {Value: "wrong-password", Version: "v1"}}
	req, _ := http.NewRequest("GET", "https://provider.example", nil)
	cfg := Config{AuthType: Basic, Username: "synthetic-user", SecretRef: "dedicated", SecretVersion: "v1"}
	if err := c.Apply(context.Background(), http.DefaultClient, "account", cfg, req); err != nil {
		t.Fatal(err)
	}
	user, password, ok := req.BasicAuth()
	if !ok || user != "synthetic-user" || password != "synthetic-password" {
		t.Fatal("decrypted credential not applied")
	}
	cfg.SecretVersion = "missing"
	next, _ := http.NewRequest("GET", "https://provider.example", nil)
	if err := c.Apply(context.Background(), http.DefaultClient, "account", cfg, next); err == nil || next.Header.Get("Authorization") != "" {
		t.Fatal("missing dedicated version was substituted")
	}
}

func TestOAuthWorksWithoutRedisAndScopesTokensByBinding(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_ = r.ParseForm()
		if r.Form.Get("client_secret") != "actual-client-secret" || r.Form.Get("client_secret_ref") != "" {
			t.Error("reference sent as credential")
			w.WriteHeader(401)
			return
		}
		fmt.Fprintf(w, `{"access_token":"synthetic-token-%d","expires_in":30}`, calls.Load())
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("EGRESS_HTTP_ORIGINS", server.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", u.Host+"=127.0.0.1/32")
	c := NewTokenCache("127.0.0.1:1")
	defer c.Close()
	c.Resolver = testResolver{"oauth-secret": {Value: "actual-client-secret", Version: "v1"}}
	cfg := Config{AuthType: OAuthClientCredentials, TokenURL: server.URL, ClientID: "fixture-client", ClientSecretRef: "oauth-secret", SecretVersion: "v1", BindingID: "binding-a", TenantID: "a", Environment: "local"}
	for _, binding := range []string{"binding-a", "binding-a", "binding-b"} {
		cfg.BindingID = binding
		req, _ := http.NewRequest("GET", "https://provider.example", nil)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := c.Apply(ctx, http.DefaultClient, "account", cfg, req)
		cancel()
		if err != nil || !strings.HasPrefix(req.Header.Get("Authorization"), "Bearer synthetic-token-") {
			t.Fatalf("OAuth without Redis: %v", err)
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("token requests=%d; wanted reuse only within binding", calls.Load())
	}
}

func TestOAuthCacheRevocationAndExpiry(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, `{"access_token":"revocable-token","expires_in":6}`)
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("EGRESS_HTTP_ORIGINS", server.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", u.Host+"=127.0.0.1/32")
	c := NewTokenCache("127.0.0.1:1")
	defer c.Close()
	c.Resolver = testResolver{"oauth-secret": {Value: "actual-client-secret", Version: "v1"}}
	cfg := Config{AuthType: OAuthClientCredentials, TokenURL: server.URL, ClientID: "fixture-client", ClientSecretRef: "oauth-secret", SecretVersion: "v1", BindingID: "binding-a", TenantID: "tenant-a", Environment: "local"}
	request := func() {
		req, _ := http.NewRequest("GET", "https://provider.example", nil)
		if err := c.Apply(context.Background(), http.DefaultClient, "account", cfg, req); err != nil {
			t.Fatal(err)
		}
	}
	request()
	request()
	if got := calls.Load(); got != 1 {
		t.Fatalf("token requests before revocation=%d; wanted 1", got)
	}
	c.InvalidateBinding("local", "tenant-a", "binding-a", "v1")
	request()
	if got := calls.Load(); got != 2 {
		t.Fatalf("token requests after revocation=%d; wanted 2", got)
	}
	time.Sleep(1100 * time.Millisecond)
	request()
	if got := calls.Load(); got != 3 {
		t.Fatalf("token requests after local expiry=%d; wanted 3", got)
	}
}

func TestOAuthLockTrackingIsBounded(t *testing.T) {
	c := NewTokenCache("127.0.0.1:1")
	defer c.Close()
	for i := 0; i < maxTrackedLocks*2; i++ {
		key := fmt.Sprintf("key-%d", i)
		lock := c.retainLock(key)
		c.releaseLock(lock)
	}
	c.locksMu.Lock()
	defer c.locksMu.Unlock()
	if len(c.locks) > maxTrackedLocks {
		t.Fatalf("tracked locks=%d; limit=%d", len(c.locks), maxTrackedLocks)
	}
}
