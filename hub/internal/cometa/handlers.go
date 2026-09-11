package cometa

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/providersim"
	"github.com/google/uuid"
)

// Handlers expoe a API interna do Cometa: despacho direto (COM-06),
// consulta interna por operation_id (EXE-04) e recepcao de callback do
// provedor (EXE-08).
type Handlers struct {
	exec  *Executor
	store *Store
}

// NewHandlers cria os handlers HTTP do Cometa.
func NewHandlers(exec *Executor, store *Store) *Handlers {
	return &Handlers{exec: exec, store: store}
}

// Register registra as rotas do Cometa num ServeMux.
func (h *Handlers) Register(mux *http.ServeMux) {
	h.RegisterInternal(mux)
	h.RegisterCallback(mux)
}

// RegisterAdmin expõe somente diagnósticos de capacidade já persistidos. A
// tela administrativa não instala, altera ou recompõe políticas: limites são
// autoridade qualificada do runtime e mudanças exigem o fluxo operacional
// próprio.
func (h *Handlers) RegisterAdmin(mux *http.ServeMux) {
	mux.HandleFunc("/admin/v1/capacity-domains", h.handleAdminCapacityDomains)
}

type capacityDomainView struct {
	Domain                 string    `json:"domain"`
	Version                string    `json:"version"`
	EvidenceRef            string    `json:"evidence_ref"`
	ValidUntil             time.Time `json:"valid_until"`
	MaxConcurrent          int       `json:"max_concurrent"`
	MinConcurrent          int       `json:"min_concurrent"`
	EffectiveLimit         int       `json:"effective_limit"`
	LatencyThresholdMillis int64     `json:"latency_threshold_millis"`
	TransportOpen          int       `json:"transport_open"`
	PendingExternal        int       `json:"pending_external"`
	RateUsed               int       `json:"rate_used"`
	LastFeedback           string    `json:"last_feedback"`
	Epoch                  int64     `json:"epoch"`
}

func (h *Handlers) handleAdminCapacityDomains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	p, ok := auth.FromContext(r.Context())
	if !ok || p.Workload || !auth.Authorize(r.Context(), "integrations:read", "") {
		auth.Error(w, http.StatusForbidden, "forbidden")
		return
	}
	domain := strings.TrimSpace(r.URL.Query().Get("domain"))
	query := `SELECT d.domain_id,d.policy,d.current_limit,d.epoch,d.rate_used,d.last_feedback,
        (SELECT count(*) FROM capacity_permits p WHERE p.domain_id=d.domain_id AND p.transport_open),
        (SELECT count(*) FROM capacity_permits p WHERE p.domain_id=d.domain_id AND p.pending_external)
        FROM capacity_domains d`
	args := []any{}
	if domain != "" {
		query += " WHERE d.domain_id=$1"
		args = append(args, domain)
	}
	query += " ORDER BY d.domain_id"
	rows, err := h.store.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		auth.Error(w, http.StatusServiceUnavailable, "capacity_authority_unavailable")
		return
	}
	defer rows.Close()
	items := make([]capacityDomainView, 0)
	for rows.Next() {
		var (
			id, rawPolicy, lastFeedback           string
			currentLimit, open, pending, rateUsed int
			epoch                                 int64
		)
		if err := rows.Scan(&id, &rawPolicy, &currentLimit, &epoch, &rateUsed, &lastFeedback, &open, &pending); err != nil {
			auth.Error(w, http.StatusServiceUnavailable, "capacity_authority_unavailable")
			return
		}
		var policy CapacityPolicy
		if err := json.Unmarshal([]byte(rawPolicy), &policy); err != nil {
			auth.Error(w, http.StatusServiceUnavailable, "capacity_policy_invalid")
			return
		}
		items = append(items, capacityDomainView{Domain: id, Version: policy.Version, EvidenceRef: policy.EvidenceRef, ValidUntil: policy.ValidUntil, MaxConcurrent: policy.MaxConcurrent, MinConcurrent: policy.MinConcurrent, EffectiveLimit: currentLimit, LatencyThresholdMillis: policy.LatencyThresholdMillis, TransportOpen: open, PendingExternal: pending, RateUsed: rateUsed, LastFeedback: lastFeedback, Epoch: epoch})
	}
	if err := rows.Err(); err != nil {
		auth.Error(w, http.StatusServiceUnavailable, "capacity_authority_unavailable")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items, "next_cursor": ""})
}

func (h *Handlers) RegisterInternal(mux *http.ServeMux) {
	mux.HandleFunc("/internal/commands/direct", h.handleDirect)
	mux.HandleFunc("/internal/operations/", h.handleGetOperation)
}

func (h *Handlers) RegisterCallback(mux *http.ServeMux) {
	mux.HandleFunc("/internal/callbacks/", h.handleCallback)
	mux.HandleFunc("/callbacks/", h.handleCallback)
}

// handleDirect e o endpoint SYNC (COM-01: "Orbita -> Cometa em SYNC:
// HTTPS REST/JSON interno idempotente, mTLS e deadline propagado").
// mTLS fica fora do escopo desta referencia local (placeholder).
func (h *Handlers) handleDirect(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.FromContext(r.Context())
	if !ok || !p.Workload || !auth.Authorize(r.Context(), "cometa:execute", "") {
		auth.Error(w, 403, "forbidden")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var cmd dispatch.Command
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512*1024)).Decode(&cmd); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_body"})
		return
	}
	if cmd.CellID != p.CellID || cmd.DispatchMode != dispatch.DispatchDirect {
		auth.Error(w, 403, "command_outside_scope")
		return
	}
	ctx := context.WithoutCancel(r.Context())
	if !cmd.StepDeadline.IsZero() {
		var cancel func()
		ctx, cancel = withDeadline(ctx, cmd.StepDeadline)
		defer cancel()
	}
	result := h.exec.Execute(ctx, cmd)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handlers) handleGetOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, ok := auth.FromContext(r.Context())
	if !ok || !principal.Workload || !auth.Authorize(r.Context(), "protocols:read", "") {
		auth.Error(w, 403, "forbidden")
		return
	}
	operationID := strings.TrimPrefix(r.URL.Path, "/internal/operations/")
	tenant := r.URL.Query().Get("tenant_id")
	application := r.URL.Query().Get("application_id")
	if tenant == "" || application == "" {
		auth.Error(w, 400, "resource_scope_required")
		return
	}
	var result struct {
		OperationID string          `json:"operation_id"`
		ProtocolID  string          `json:"protocol_id"`
		State       string          `json:"state"`
		Result      json.RawMessage `json:"result,omitempty"`
	}
	var rawResult []byte
	err := h.store.db.QueryRowContext(r.Context(), `SELECT operation_id,protocol_id,state,result FROM operations WHERE operation_id=$1 AND tenant_id=$2 AND application_id=$3 AND cell_id=$4`, operationID, tenant, application, principal.CellID).Scan(&result.OperationID, &result.ProtocolID, &result.State, &rawResult)
	if errors.Is(err, sql.ErrNoRows) {
		auth.Error(w, 404, "operation_not_found")
		return
	}
	if err != nil {
		auth.Error(w, 503, "operation_authority_unavailable")
		return
	}
	result.Result = rawResult
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// handleCallback recebe o retorno assíncrono do provedor. A capability por
// operação protege o endpoint enquanto a política homologada por conta
// (assinatura, mTLS ou token do provedor) não está disponível neste contrato.
// A resposta 2xx só é emitida depois da custódia durável.
func (h *Handlers) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	operationID := strings.TrimPrefix(r.URL.Path, "/internal/callbacks/")
	operationID = strings.TrimPrefix(operationID, "/callbacks/")
	if operationID == "" || strings.Contains(operationID, "/") {
		http.Error(w, "invalid callback operation", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(operationID); err != nil {
		http.Error(w, "invalid callback operation", http.StatusBadRequest)
		return
	}
	if !secureCallbackKey(r.Header.Get("X-Provider-Callback-Key")) {
		http.Error(w, "callback unauthorized", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 512*1024))
	var result providersim.OperationResult
	if err != nil || len(body) == 0 || json.Unmarshal(body, &result) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Only a complete terminal observation is eligible for orphan custody.
	// This keeps malformed or merely pending traffic from consuming the
	// recoverable inbox quota before operation correlation exists.
	if result.ProviderRequestID == "" || (result.Status != "SUCCEEDED" && result.Status != "FAILED") {
		http.Error(w, "invalid callback observation", http.StatusBadRequest)
		return
	}
	token := r.URL.Query().Get("token")
	if err := h.store.AuthenticateCallback(r.Context(), operationID, token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := h.store.StoreOrphanCallback(r.Context(), operationID, token, body); err != nil {
				http.Error(w, "callback custody unavailable", http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if errors.Is(err, ErrCallbackCapabilityInvalid) {
			http.Error(w, "callback unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "callback custody unavailable", http.StatusServiceUnavailable)
		return
	}
	// 2xx is an acknowledgement of durable custody, never merely of parsing.
	if _, err := h.exec.ApplyExternalObservationRaw(r.Context(), operationID, result, body); err != nil {
		http.Error(w, "callback custody unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func secureCallbackKey(value string) bool {
	expected := os.Getenv("CALLBACK_INGRESS_KEY")
	return expected != "" && value != "" && subtle.ConstantTimeCompare([]byte(value), []byte(expected)) == 1
}
