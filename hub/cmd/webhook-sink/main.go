// Comando webhook-sink e um destino de webhook de teste usado apenas
// nos ensaios locais desta referencia (nao faz parte do dominio do
// hub): recebe e lista as entregas do Pulsar para verificacao manual
// e automatizada (QUA-03: "comparar hash/schema/media type do corpo
// final servido pelo GET com o corpo enviado em todas as tentativas de
// webhook").
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"ai-hub/hub/internal/platform/config"
	"ai-hub/hub/internal/platform/logging"
)

type received struct {
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

func main() {
	log := logging.New("webhook-sink")
	addr := config.Env("HTTP_ADDR", ":8091")

	var mu sync.Mutex
	var items []received

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		items = append(items, received{
			Headers: map[string]string{
				"X-Hub-Signature-256": r.Header.Get("X-Hub-Signature-256"),
				"X-Hub-Event-Id":      r.Header.Get("X-Hub-Event-Id"),
			},
			Body: json.RawMessage(body),
		})
		mu.Unlock()
		log.Info("webhook recebido", "event_id", r.Header.Get("X-Hub-Event-Id"))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/received", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
	})

	log.Info("webhook-sink starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("server stopped", "error", err)
	}
}
