package atlas

import (
	"ai-hub/hub/internal/platform/auth"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Handlers expoe a API HTTP administrativa minima do Atlas (CFG-01).
// Nao faz chamada de provedor por requisicao (ARQ-01).
type Handlers struct {
	store *Store
}

// NewHandlers cria os handlers HTTP do Atlas.
func NewHandlers(store *Store) *Handlers { return &Handlers{store: store} }

// Register registra as rotas do Atlas num ServeMux.
func (h *Handlers) Register(mux *http.ServeMux) {
	h.registerCatalog(mux)
	retired := func(w http.ResponseWriter, r *http.Request) { auth.Error(w, 410, "use_versioned_catalog") }
	mux.HandleFunc("/v1/services", retired)
	mux.HandleFunc("/v1/provider-accounts", retired)
	mux.HandleFunc("/v1/credential-bindings", retired)
	mux.HandleFunc("/v1/contracts", retired)
	mux.HandleFunc("/v1/services/", h.workloadRead("catalog:read", h.handleGetService))
	mux.HandleFunc("/v1/provider-accounts/", h.workloadRead("integrations:read", h.handleGetProviderAccount))
	mux.HandleFunc("/v1/credentials/resolve", h.workloadRead("integrations:read", h.handleResolveCredential))
	mux.HandleFunc("/v1/credentials/binding/", h.workloadRead("integrations:read", h.handleBoundCredential))
	mux.HandleFunc("/v1/contracts/", h.workloadRead("catalog:read", h.handleGetContract))
}

func (h *Handlers) handleBoundCredential(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/credentials/binding/"), "/")
	if len(parts) != 2 {
		auth.Error(w, 400, "invalid_binding_version")
		return
	}
	version, err := strconv.Atoi(parts[1])
	if err != nil {
		auth.Error(w, 400, "invalid_binding_version")
		return
	}
	tenant := r.URL.Query().Get("tenant_id")
	resource, err := h.store.GetResource(r.Context(), "credential-bindings", parts[0], version)
	if err != nil || tenant == "" || resource.State != "PUBLISHED" || (resource.TenantID != "" && resource.TenantID != tenant) {
		auth.Error(w, 409, "binding_unavailable")
		return
	}
	data, err := DecodeCatalogData(resource)
	if err != nil {
		auth.Error(w, 409, "binding_unavailable")
		return
	}
	if data.CredentialMode == "TENANT_DEDICATED" && resource.TenantID != tenant {
		auth.Error(w, 409, "binding_unavailable")
		return
	}
	now := time.Now()
	if (data.ValidFrom != nil && now.Before(*data.ValidFrom)) || (data.ValidUntil != nil && !now.Before(*data.ValidUntil)) {
		auth.Error(w, 409, "binding_unavailable")
		return
	}
	writeJSON(w, 200, map[string]any{"binding_id": resource.ID, "tenant_id": resource.TenantID, "provider_account_id": data.ProviderAccountID, "credential_mode": data.CredentialMode, "secret_ref": data.SecretRef, "secret_version": data.SecretVersion, "settlement_party": data.SettlementParty, "state": resource.State})
}

func (h *Handlers) workloadRead(scope string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := auth.FromContext(r.Context())
		if !ok || !p.Workload || !auth.Authorize(r.Context(), scope, "") {
			auth.Error(w, 403, "forbidden")
			return
		}
		if r.Method != http.MethodGet {
			auth.Error(w, 405, "method_not_allowed")
			return
		}
		tenant := r.URL.Query().Get("tenant_id")
		if strings.HasPrefix(r.URL.Path, "/v1/contracts/") {
			tenant = strings.TrimPrefix(r.URL.Path, "/v1/contracts/")
		}
		if tenant != "" {
			var cell string
			err := h.store.db.QueryRowContext(r.Context(), "SELECT cell_id FROM placements WHERE tenant_id=$1", tenant).Scan(&cell)
			if err != nil || cell != p.CellID {
				auth.Error(w, 403, "tenant_outside_cell")
				return
			}
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"error": code, "message": msg})
}

func (h *Handlers) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	var svc Service
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := validateService(svc); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_service", err.Error())
		return
	}
	if err := h.store.UpsertService(r.Context(), svc); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, svc)
}

func validateService(svc Service) error {
	if strings.TrimSpace(svc.Code) == "" {
		return errors.New("code e obrigatorio")
	}
	if svc.Version < 1 {
		return errors.New("version deve ser maior ou igual a 1")
	}
	if strings.TrimSpace(svc.Description) == "" {
		return errors.New("description e obrigatoria")
	}
	if svc.ClientSLASeconds <= 0 {
		return errors.New("client_sla_seconds deve ser maior que zero")
	}
	if svc.RetryTTLSeconds < 0 {
		return errors.New("retry_ttl_seconds nao pode ser negativo")
	}
	if len(svc.Modes) == 0 {
		return errors.New("modes deve conter ao menos um modo")
	}
	seen := make(map[string]struct{}, len(svc.Modes))
	for _, mode := range svc.Modes {
		mode = strings.ToUpper(strings.TrimSpace(mode))
		if mode != "SYNC" && mode != "ASYNC" && mode != "AUTO" {
			return errors.New("modes aceita somente SYNC, ASYNC ou AUTO")
		}
		if _, ok := seen[mode]; ok {
			return errors.New("modes nao pode conter duplicidades")
		}
		seen[mode] = struct{}{}
	}
	return nil
}

func (h *Handlers) handleGetService(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/services/"), "/")
	if len(parts) != 2 {
		writeErr(w, http.StatusBadRequest, "invalid_path", "esperado /v1/services/{code}/{version}")
		return
	}
	version, err := strconv.Atoi(parts[1])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_version", "versao deve ser inteira")
		return
	}
	svc, err := h.store.GetService(r.Context(), parts[0], version)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not_found", "servico/versao nao encontrado")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, svc)
}

func (h *Handlers) handleProviderAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	var pa ProviderAccount
	if err := json.NewDecoder(r.Body).Decode(&pa); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := validateProviderAccount(&pa); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_provider_account", err.Error())
		return
	}
	if err := h.store.UpsertProviderAccount(r.Context(), pa); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, pa)
}

func validateProviderAccount(pa *ProviderAccount) error {
	if strings.TrimSpace(pa.ProviderAccountID) == "" || strings.TrimSpace(pa.ProviderID) == "" {
		return errors.New("provider_account_id e provider_id sao obrigatorios")
	}
	baseURL, err := url.ParseRequestURI(strings.TrimSpace(pa.BaseURL))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return errors.New("base_url deve ser uma URL valida")
	}
	pa.AuthType = strings.ToUpper(strings.TrimSpace(pa.AuthType))
	if pa.AuthType == "" {
		pa.AuthType = "NONE"
	}
	if pa.TokenTTLSeconds == 0 {
		pa.TokenTTLSeconds = 300
	}
	if pa.TokenTTLSeconds < 1 {
		return errors.New("token_ttl_seconds deve ser maior que zero")
	}
	switch pa.AuthType {
	case "NONE":
	case "BASIC":
		if strings.TrimSpace(pa.AuthUsername) == "" || strings.TrimSpace(pa.AuthSecretRef) == "" {
			return errors.New("BASIC exige auth_username e auth_secret_ref")
		}
	case "OAUTH_CLIENT_CREDENTIALS", "MTLS_OAUTH":
		tokenURL, parseErr := url.ParseRequestURI(strings.TrimSpace(pa.OAuthTokenURL))
		if parseErr != nil || tokenURL.Scheme == "" || tokenURL.Host == "" ||
			strings.TrimSpace(pa.OAuthClientID) == "" || strings.TrimSpace(pa.OAuthClientSecretRef) == "" {
			return errors.New("OAuth exige oauth_token_url, oauth_client_id e oauth_client_secret_ref")
		}
		if pa.AuthType == "MTLS_OAUTH" && strings.TrimSpace(pa.MTLSCertificateRef) == "" {
			return errors.New("MTLS_OAUTH exige mtls_certificate_ref")
		}
	default:
		return errors.New("auth_type aceita somente NONE, BASIC, OAUTH_CLIENT_CREDENTIALS ou MTLS_OAUTH")
	}
	return nil
}

func (h *Handlers) handleGetProviderAccount(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/provider-accounts/")
	pa, err := h.store.GetProviderAccount(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not_found", "conta de provedor nao encontrada")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pa)
}

func (h *Handlers) handleCredentialBindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	var cb CredentialBinding
	if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if cb.State == "" {
		cb.State = "ATIVO"
	}
	if err := h.store.UpsertCredentialBinding(r.Context(), cb); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cb)
}

// handleResolveCredential implementa a resolucao determinística de
// SEG-05: GET /v1/credentials/resolve?tenant_id=&provider_account_id=
func (h *Handlers) handleResolveCredential(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	providerAccountID := r.URL.Query().Get("provider_account_id")
	if tenantID == "" || providerAccountID == "" {
		writeErr(w, http.StatusBadRequest, "missing_params", "tenant_id e provider_account_id sao obrigatorios")
		return
	}
	cb, err := h.store.ResolveCredential(r.Context(), tenantID, providerAccountID)
	if errors.Is(err, ErrCredentialUnavailable) {
		writeErr(w, http.StatusConflict, "credential_unavailable", "credencial indisponivel para o modo exigido pelo contrato; sem fallback")
		return
	}
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "contract_not_found", "contrato do tenant nao encontrado")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cb)
}

func (h *Handlers) handleContracts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	var c Contract
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if c.CredentialModeRequired == "" {
		c.CredentialModeRequired = "SHARED_HUB"
	}
	if err := h.store.UpsertContract(r.Context(), c); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handlers) handleGetContract(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimPrefix(r.URL.Path, "/v1/contracts/")
	c, err := h.store.GetContract(r.Context(), tenantID)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not_found", "contrato nao encontrado")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, c)
}
