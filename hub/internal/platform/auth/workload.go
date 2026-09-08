package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// WorkloadClient attaches a short-lived service credential only to its configured
// origin. Redirects are deliberately not followed. Secrets stay in private memory.
func WorkloadClient(baseURL string, timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, Transport: &workloadTransport{origin: baseURL, base: http.DefaultTransport}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

type workloadTransport struct {
	origin  string
	base    http.RoundTripper
	mu      sync.Mutex
	token   string
	expires time.Time
}

func (t *workloadTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	u, err := url.Parse(t.origin)
	if err != nil || u.Scheme != r.URL.Scheme || u.Host != r.URL.Host {
		return nil, errors.New("destino fora do escopo do workload")
	}
	token, err := t.accessToken(r.Context())
	if err != nil {
		return nil, err
	}
	clone := r.Clone(r.Context())
	clone.Header = r.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+token)
	return t.base.RoundTrip(clone)
}

func (t *workloadTransport) accessToken(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.token != "" && time.Now().Add(5*time.Second).Before(t.expires) {
		return t.token, nil
	}
	tokenURL := os.Getenv("OIDC_TOKEN_URL")
	clientID := os.Getenv("WORKLOAD_CLIENT_ID")
	secretPath := os.Getenv("WORKLOAD_CLIENT_SECRET_FILE")
	if tokenURL == "" || clientID == "" || secretPath == "" {
		return "", errors.New("credencial de workload indisponível")
	}
	secret, err := os.ReadFile(secretPath)
	if err != nil {
		return "", errors.New("credencial de workload indisponível")
	}
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {clientID}, "client_secret": {strings.TrimSpace(string(secret))}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.New("emissor de workload inválido")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("emissor de workload indisponível")
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Type        string `json:"token_type"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 32*1024)).Decode(&result) != nil || result.AccessToken == "" || result.ExpiresIn <= 5 || !strings.EqualFold(result.Type, "Bearer") {
		return "", errors.New("credencial de workload recusada")
	}
	t.token = result.AccessToken
	t.expires = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return t.token, nil
}
