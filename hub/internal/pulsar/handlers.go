package pulsar

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handlers expoe a API interna do Pulsar: cadastro de destino de
// webhook e consulta de entregas (EXE-07: "consultar entregas").
type Handlers struct {
	store *Store
}

// NewHandlers cria os handlers HTTP do Pulsar.
func NewHandlers(store *Store) *Handlers { return &Handlers{store: store} }

// Register registra as rotas internas do Pulsar.
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/internal/destinations", h.handleUpsertDestination)
	mux.HandleFunc("/internal/deliveries/", h.handleGetDeliveries)
}

func (h *Handlers) handleUpsertDestination(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var d Destination
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.store.UpsertDestination(r.Context(), d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"tenant_id": d.TenantID})
}

func (h *Handlers) handleGetDeliveries(w http.ResponseWriter, r *http.Request) {
	protocolID := strings.TrimPrefix(r.URL.Path, "/internal/deliveries/")
	rows, err := h.store.db.QueryContext(r.Context(), `
		SELECT delivery_id, protocol_id, event_id, destination_url, state, attempts_count, next_attempt_at
		FROM deliveries WHERE protocol_id = $1 ORDER BY created_at
	`, protocolID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var out []Delivery
	for rows.Next() {
		var d Delivery
		if err := rows.Scan(&d.DeliveryID, &d.ProtocolID, &d.EventID, &d.DestinationURL, &d.State, &d.AttemptsCount, &d.NextAttemptAt); err == nil {
			out = append(out, d)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
