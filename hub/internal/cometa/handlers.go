package cometa

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-hub/hub/internal/dispatch"
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var cmd dispatch.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_body"})
		return
	}
	ctx := r.Context()
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
	operationID := strings.TrimPrefix(r.URL.Path, "/internal/operations/")
	op, err := h.store.Get(r.Context(), operationID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(op)
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
