package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/providersim"
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
	mux.HandleFunc("/internal/commands/direct", h.handleDirect)
	mux.HandleFunc("/internal/operations/", h.handleGetOperation)
	mux.HandleFunc("/internal/callbacks/", h.handleCallback)
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

// handleCallback recebe o retorno assincrono do provedor simulado em
// modo async_callback (EXE-08: autenticar, persistir recibo, so
// depois responder 2xx — autenticacao de callback e um placeholder
// nesta referencia local, ja que nao ha segredo real de provedor).
func (h *Handlers) handleCallback(w http.ResponseWriter, r *http.Request) {
	operationID := strings.TrimPrefix(r.URL.Path, "/internal/callbacks/")
	var result providersim.OperationResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 2xx confirma recebimento HTTP; a aplicacao do fato ocorre de
	// forma duravel antes desta resposta (EXE-08).
	h.exec.ApplyExternalObservation(r.Context(), dispatch.Command{}, operationID, result)
	w.WriteHeader(http.StatusOK)
}
