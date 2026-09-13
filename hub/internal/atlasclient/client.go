// Package atlasclient e o cliente HTTP que Orbita/Cometa/Libra usam
// para consultar projecoes publicadas pelo Atlas (catalogo, contas de
// provedor, credenciais, contratos). Mantem um cache local simples com
// TTL curto para nao consultar o Atlas a cada pedido (ARQ-02: "o
// trafego estabelecido nao consulta Atlas a cada pedido").
package atlasclient

import (
	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/platform/auth"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Offer resolves a published, tenant/application scoped snapshot. The local
// projection is the hot-path authority until its lease expires; an explicit
// refresh is available for catalog invalidation/revocation events.
func (c *Client) Offer(ctx context.Context, tenant, application, service string, serviceVersion int, account string) (atlas.OfferSnapshot, error) {
	key := fmt.Sprintf("offer:%s:%s:%s:%d:%s", tenant, application, service, serviceVersion, account)
	if cached, ok := c.getCached(key); ok {
		candidate, valid := cached.(atlas.OfferSnapshot)
		if valid && candidate.Hash != "" && time.Now().Before(candidate.ValidUntil) {
			return candidate, nil
		}
		c.invalidate(key)
	}
	return c.refreshOffer(ctx, key, tenant, application, service, serviceVersion, account)
}

// OfferAuthoritative forces a lookup at Atlas. Denials and invalid snapshots
// evict the projection and are never hidden by an older successful offer.
func (c *Client) OfferAuthoritative(ctx context.Context, tenant, application, service string, serviceVersion int, account string) (atlas.OfferSnapshot, error) {
	key := fmt.Sprintf("offer:%s:%s:%s:%d:%s", tenant, application, service, serviceVersion, account)
	return c.refreshOffer(ctx, key, tenant, application, service, serviceVersion, account)
}

// InvalidateOffer removes a projection after a catalog, account or policy
// event. Subsequent Offer calls must obtain a new published snapshot.
func (c *Client) InvalidateOffer(tenant, application, service string, serviceVersion int, account string) {
	c.invalidate(fmt.Sprintf("offer:%s:%s:%s:%d:%s", tenant, application, service, serviceVersion, account))
}

func (c *Client) refreshOffer(ctx context.Context, key, tenant, application, service string, serviceVersion int, account string) (atlas.OfferSnapshot, error) {
	q := url.Values{"tenant_id": {tenant}, "application_id": {application}, "service_code": {service}, "service_version": {strconv.Itoa(serviceVersion)}, "provider_account_id": {account}}
	var out atlas.OfferSnapshot
	status, err := c.getJSON(ctx, "/v1/offers/resolve?"+q.Encode(), &out)
	if err == nil {
		if !time.Now().Before(out.ValidUntil) || out.Hash == "" {
			c.invalidate(key)
			return atlas.OfferSnapshot{}, ErrOfferProjectionInvalid
		}
		c.setCached(key, out)
		return out, nil
	}
	if status == http.StatusForbidden || status == http.StatusConflict {
		c.invalidate(key)
	}
	return atlas.OfferSnapshot{}, err
}

// Client e o cliente HTTP do Atlas com cache local.
type Client struct {
	baseURL string
	http    *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	value   any
	expires time.Time
}

// maxCacheEntries bounds unbounded growth from an ever-widening set of
// tenant/application/service/account combinations. On overflow, expired
// entries are swept first; a still-full cache drops one arbitrary entry
// rather than grow past the limit.
const maxCacheEntries = 10000

var ErrOfferProjectionInvalid = fmt.Errorf("atlasclient: snapshot expired or invalid")

// HTTPError keeps the authoritative status available to callers that must
// distinguish denial/revocation from temporary Atlas unavailability.
type HTTPError struct {
	Status int
	Path   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("atlasclient: status %d de %s", e.Status, e.Path)
}

// New cria um cliente do Atlas.
func New(baseURL string, ttl time.Duration) *Client {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &Client{
		baseURL: baseURL,
		http:    auth.WorkloadClient(baseURL, 3*time.Second),
		cache:   make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// Contract e a projecao local do contrato do tenant.
type Contract struct {
	TenantID               string  `json:"tenant_id"`
	Plan                   string  `json:"plan"`
	UnitPrice              float64 `json:"unit_price"`
	StrictBalance          bool    `json:"strict_balance"`
	ClientSLASeconds       int     `json:"client_sla_seconds"`
	CredentialModeRequired string  `json:"credential_mode_required"`
}

// ProviderAccount e a projecao local de uma conta de provedor.
type ProviderAccount struct {
	ProviderAccountID    string `json:"provider_account_id"`
	ProviderID           string `json:"provider_id"`
	BaseURL              string `json:"base_url"`
	ProviderMode         string `json:"provider_mode"`
	AuthType             string `json:"auth_type"`
	APIKeyHeader         string `json:"api_key_header"`
	AuthUsername         string `json:"auth_username"`
	AuthSecretRef        string `json:"auth_secret_ref"`
	OAuthTokenURL        string `json:"oauth_token_url"`
	OAuthClientID        string `json:"oauth_client_id"`
	OAuthClientSecretRef string `json:"oauth_client_secret_ref"`
	MTLSCertificateRef   string `json:"mtls_certificate_ref"`
	TokenTTLSeconds      int    `json:"token_ttl_seconds"`
}

// CredentialBinding e a projecao local de um vinculo de credencial
// resolvido (nunca contem o segredo em si).
type CredentialBinding struct {
	SecretVersion     string `json:"secret_version"`
	BindingID         string `json:"binding_id"`
	CredentialMode    string `json:"credential_mode"`
	TenantID          string `json:"tenant_id"`
	ProviderAccountID string `json:"provider_account_id"`
	SecretRef         string `json:"secret_ref"`
	SettlementParty   string `json:"settlement_party"`
	State             string `json:"state"`
}

// ErrCredentialUnavailable espelha atlas.ErrCredentialUnavailable pelo
// canal HTTP (status 409).
type ErrCredentialUnavailable struct{ Tenant, ProviderAccount string }

func (e *ErrCredentialUnavailable) Error() string {
	return fmt.Sprintf("atlasclient: credencial indisponivel para tenant=%s provider_account=%s, sem fallback", e.Tenant, e.ProviderAccount)
}

func (c *Client) getCached(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.cache[key]
	if !ok || time.Now().After(e.expires) {
		return nil, false
	}
	return e.value, true
}

func (c *Client) setCached(key string, v any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.cache[key]; !exists && len(c.cache) >= maxCacheEntries {
		now := time.Now()
		for k, e := range c.cache {
			if now.After(e.expires) {
				delete(c.cache, k)
			}
		}
		if len(c.cache) >= maxCacheEntries {
			for k := range c.cache {
				delete(c.cache, k)
				break
			}
		}
	}
	c.cache[key] = cacheEntry{value: v, expires: time.Now().Add(c.ttl)}
}

func (c *Client) invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

func (c *Client) getJSON(ctx context.Context, path string, out any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode, &HTTPError{Status: resp.StatusCode, Path: path}
}

// Contract busca o contrato do tenant, usando o cache local quando
// disponivel.
func (c *Client) Contract(ctx context.Context, tenantID string) (Contract, error) {
	key := "contract:" + tenantID
	if v, ok := c.getCached(key); ok {
		return v.(Contract), nil
	}
	var out Contract
	if _, err := c.getJSON(ctx, "/v1/contracts/"+url.PathEscape(tenantID), &out); err != nil {
		return Contract{}, err
	}
	c.setCached(key, out)
	return out, nil
}

// ProviderAccount busca a conta de provedor pelo ID.
func (c *Client) ProviderAccount(ctx context.Context, id string) (ProviderAccount, error) {
	key := "provider_account:" + id
	if v, ok := c.getCached(key); ok {
		return v.(ProviderAccount), nil
	}
	var out ProviderAccount
	if _, err := c.getJSON(ctx, "/v1/provider-accounts/"+url.PathEscape(id), &out); err != nil {
		return ProviderAccount{}, err
	}
	c.setCached(key, out)
	return out, nil
}

// ResolveCredential resolve o vinculo de credencial elegivel (SEG-05),
// sem cache (a resolucao envolve estado que pode mudar por rotacao ou
// revogacao emergencial — CFG-02 exige distribuicao prioritaria dessas
// mudancas, entao esta chamada nao usa o cache do Contract/ProviderAccount).
func (c *Client) ResolveCredential(ctx context.Context, tenantID, providerAccountID string) (CredentialBinding, error) {
	var out CredentialBinding
	q := url.Values{"tenant_id": {tenantID}, "provider_account_id": {providerAccountID}}
	status, err := c.getJSON(ctx, "/v1/credentials/resolve?"+q.Encode(), &out)
	if status == http.StatusConflict {
		return CredentialBinding{}, &ErrCredentialUnavailable{Tenant: tenantID, ProviderAccount: providerAccountID}
	}
	if err != nil {
		return CredentialBinding{}, err
	}
	return out, nil
}

func (c *Client) BoundCredential(ctx context.Context, tenant, binding string, version int) (CredentialBinding, error) {
	var out CredentialBinding
	_, err := c.getJSON(ctx, fmt.Sprintf("/v1/credentials/binding/%s/%d?tenant_id=%s", url.PathEscape(binding), version, url.QueryEscape(tenant)), &out)
	return out, err
}
