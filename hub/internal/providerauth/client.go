// Package providerauth implementa autenticacao outbound dos chamados do Hub.
// Em local/dev os segredos sao referencias e o provider-sim valida o contrato
// sem receber segredo em claro. O cache de access token usa Redis com TTL.
package providerauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	None                   = "NONE"
	Basic                  = "BASIC"
	OAuthClientCredentials = "OAUTH_CLIENT_CREDENTIALS"
	MTLSOAuth              = "MTLS_OAUTH"
)

type Config struct {
	AuthType           string
	Username           string
	SecretRef          string
	TokenURL           string
	ClientID           string
	ClientSecretRef    string
	MTLSCertificateRef string
	TokenTTLSeconds    int
}

type TokenCache struct{ client *redis.Client }

func NewTokenCache(addr string) *TokenCache {
	return &TokenCache{client: redis.NewClient(&redis.Options{Addr: addr})}
}

func (c *TokenCache) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }

func (c *TokenCache) Close() error { return c.client.Close() }

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *TokenCache) bearer(ctx context.Context, httpClient *http.Client, providerAccountID string, cfg Config) (string, error) {
	key := "hub:provider-token:" + providerAccountID
	if token, err := c.client.Get(ctx, key).Result(); err == nil && token != "" {
		return token, nil
	}
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {cfg.ClientID}, "client_secret_ref": {cfg.ClientSecretRef}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("oauth token endpoint status %d", resp.StatusCode)
	}
	var out tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("oauth token response: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("oauth token response sem access_token")
	}
	ttl := cfg.TokenTTLSeconds
	if ttl <= 0 || (out.ExpiresIn > 0 && out.ExpiresIn < ttl) {
		ttl = out.ExpiresIn
	}
	if ttl <= 0 {
		ttl = 300
	}
	if ttl > 5 {
		ttl--
	} // margem para não usar token no instante da expiração
	if err := c.client.Set(ctx, key, out.AccessToken, time.Duration(ttl)*time.Second).Err(); err != nil {
		return "", fmt.Errorf("cache oauth token: %w", err)
	}
	return out.AccessToken, nil
}

func (c *TokenCache) Apply(ctx context.Context, httpClient *http.Client, providerAccountID string, cfg Config, req *http.Request) error {
	switch cfg.AuthType {
	case "", None:
		return nil
	case Basic:
		if cfg.Username == "" || cfg.SecretRef == "" {
			return fmt.Errorf("basic auth incompleta")
		}
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.SecretRef)))
		return nil
	case OAuthClientCredentials, MTLSOAuth:
		if c == nil {
			return fmt.Errorf("cache Redis não configurado")
		}
		token, err := c.bearer(ctx, httpClient, providerAccountID, cfg)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if cfg.AuthType == MTLSOAuth {
			if cfg.MTLSCertificateRef == "" {
				return fmt.Errorf("referencia de certificado mTLS ausente")
			}
			// O provider-sim valida a referência. TLS/mTLS criptográfico exige
			// certificados e transporte HTTPS em ambiente de homologação.
			req.Header.Set("X-MTLS-Certificate-Ref", cfg.MTLSCertificateRef)
		}
		return nil
	default:
		return fmt.Errorf("tipo de autenticacao não suportado: %s", cfg.AuthType)
	}
}
