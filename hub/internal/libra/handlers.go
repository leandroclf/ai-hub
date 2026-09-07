package libra

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// ReserveRequest e o corpo de POST /internal/reservations (FIN-06).
type ReserveRequest struct {
	TenantID   string  `json:"tenant_id"`
	ProtocolID string  `json:"protocol_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

// Handlers expoe a API interna do Libra: reserva de saldo estrito
// (FIN-06). A Orbita chama este endpoint com aprovacao financeira
// exigida antes de qualquer efeito externo quando o contrato requer.
type Handlers struct {
	store *Store
	log   *slog.Logger
}

// NewHandlers cria os handlers HTTP do Libra.
func NewHandlers(store *Store, log *slog.Logger) *Handlers { return &Handlers{store: store, log: log} }

// Register registra as rotas internas do Libra.
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/internal/reservations", h.handleReserve)
}

func (h *Handlers) handleReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Currency == "" {
		req.Currency = "BRL"
	}
	limit, err := h.store.CreditLimit(r.Context(), req.TenantID)
	if err != nil {
		h.log.Error("falha ao ler limite de credito", "error", err, "tenant_id", req.TenantID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = h.store.Reserve(r.Context(), req.TenantID, req.ProtocolID, req.Amount, req.Currency, limit)
	if errors.Is(err, ErrLimitExceeded) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "limit_exceeded", "message": "reserva excederia o limite estrito do tenant"})
		return
	}
	if err != nil {
		h.log.Error("falha ao reservar saldo", "error", err, "tenant_id", req.TenantID, "protocol_id", req.ProtocolID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "reserved"})
}
