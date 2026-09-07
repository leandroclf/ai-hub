// Package providersim implementa um provedor externo simulado,
// deterministico e com falhas injetaveis, conforme QUA-01 ("Simulador
// local reproduz comportamento deterministico e falhas injetadas").
// Ele nao faz parte do dominio do hub: representa a contraparte
// externa que Cometa chama.
package providersim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Mode escolhe o comportamento de entrega do provedor simulado.
type Mode string

const (
	ModeSync           Mode = "sync"            // responde no mesmo request (para SYNC direto)
	ModeAsyncPoll      Mode = "async_poll"       // 202 imediato, resultado disponivel depois via GET
	ModeAsyncCallback  Mode = "async_callback"   // 202 imediato, POST de callback depois
)

// SubmitRequest e o corpo aceito por POST /v1/operations.
type SubmitRequest struct {
	ProtocolID  string `json:"protocol_id"`
	Mode        Mode   `json:"mode"`
	DelayMs     int    `json:"delay_ms"`
	Fail        bool   `json:"fail"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// OperationResult e o formato de resultado devolvido pelo provedor,
// tanto na resposta sincrona quanto no callback/poll.
type OperationResult struct {
	ProviderRequestID string `json:"provider_request_id"`
	Status            string `json:"status"` // SUCCEEDED | FAILED | PENDING
	Detail            string `json:"detail,omitempty"`
}

type operation struct {
	result OperationResult
	ready  bool
}

// Server e o servidor HTTP do provedor simulado.
type Server struct {
	mu         sync.Mutex
	ops        map[string]*operation
	seq        int
	httpClient *http.Client
}

// NewServer cria um novo provedor simulado.
func NewServer() *Server {
	return &Server{
		ops:        make(map[string]*operation),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *Server) nextID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return fmt.Sprintf("prov-req-%06d", s.seq)
}

// Routes registra as rotas do provedor simulado num ServeMux.
func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/operations", s.handleSubmit)
	mux.HandleFunc("/v1/operations/", s.handleGet)
}

func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := s.nextID()
	status := "SUCCEEDED"
	if req.Fail {
		status = "FAILED"
	}
	result := OperationResult{ProviderRequestID: id, Status: status, Detail: "simulado"}

	switch req.Mode {
	case ModeAsyncPoll:
		s.mu.Lock()
		s.ops[id] = &operation{result: OperationResult{ProviderRequestID: id, Status: "PENDING"}}
		s.mu.Unlock()
		go s.resolveAfter(id, req.DelayMs, result)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(OperationResult{ProviderRequestID: id, Status: "PENDING"})
	case ModeAsyncCallback:
		s.mu.Lock()
		s.ops[id] = &operation{result: OperationResult{ProviderRequestID: id, Status: "PENDING"}}
		s.mu.Unlock()
		go s.callbackAfter(id, req.DelayMs, result, req.CallbackURL)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(OperationResult{ProviderRequestID: id, Status: "PENDING"})
	default: // ModeSync
		if req.DelayMs > 0 {
			time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
		}
		s.mu.Lock()
		s.ops[id] = &operation{result: result, ready: true}
		s.mu.Unlock()
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	}
}

func (s *Server) resolveAfter(id string, delayMs int, result OperationResult) {
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	s.mu.Lock()
	s.ops[id] = &operation{result: result, ready: true}
	s.mu.Unlock()
}

func (s *Server) callbackAfter(id string, delayMs int, result OperationResult, callbackURL string) {
	s.resolveAfter(id, delayMs, result)
	if callbackURL == "" {
		return
	}
	body, _ := json.Marshal(result)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v1/operations/"):]
	s.mu.Lock()
	op, ok := s.ops[id]
	s.mu.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(op.result)
}
