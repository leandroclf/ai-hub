// Package providersim implementa um provedor externo simulado,
// deterministico e com falhas injetaveis, conforme QUA-01 ("Simulador
// local reproduz comportamento deterministico e falhas injetadas").
// Ele nao faz parte do dominio do hub: representa a contraparte
// externa que Cometa chama.
package providersim

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Mode escolhe o comportamento de entrega do provedor simulado.
type Mode string

const (
	ModeSync          Mode = "sync"           // responde no mesmo request (para SYNC direto)
	ModeAsyncPoll     Mode = "async_poll"     // 202 imediato, resultado disponivel depois via GET
	ModeAsyncCallback Mode = "async_callback" // 202 imediato, POST de callback depois
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
	protocols  map[string]string
	seq        int
	effects    int
	httpClient *http.Client
	tokens     map[string]time.Time
	stateFile  string
}

type persistedState struct {
	Ops       map[string]persistedOperation `json:"ops"`
	Protocols map[string]string             `json:"protocols"`
	Seq       int                           `json:"seq"`
	Effects   int                           `json:"effects"`
}

type persistedOperation struct {
	Result OperationResult `json:"result"`
	Ready  bool            `json:"ready"`
}

// NewServer cria um novo provedor simulado.
func NewServer() *Server {
	return NewServerWithState("")
}

// NewServerWithState habilita a persistência local do oráculo externo. O
// arquivo deve estar em um volume exclusivo do simulador; estado comercial
// real nunca é usado nesta fixture.
func NewServerWithState(stateFile string) *Server {
	s := newServer(stateFile)
	s.loadState()
	return s
}

func newServer(stateFile string) *Server {
	return &Server{
		ops:        make(map[string]*operation),
		protocols:  make(map[string]string),
		httpClient: &http.Client{Timeout: 5 * time.Second},
		tokens:     make(map[string]time.Time),
		stateFile:  stateFile,
	}
}

func (s *Server) loadState() {
	if s.stateFile == "" {
		return
	}
	b, err := os.ReadFile(s.stateFile)
	if err != nil {
		return
	}
	var state persistedState
	if json.Unmarshal(b, &state) != nil {
		return
	}
	for id, op := range state.Ops {
		s.ops[id] = &operation{result: op.Result, ready: op.Ready}
	}
	s.protocols, s.seq, s.effects = state.Protocols, state.Seq, state.Effects
}

func (s *Server) saveStateLocked() {
	if s.stateFile == "" {
		return
	}
	state := persistedState{Ops: make(map[string]persistedOperation, len(s.ops)), Protocols: s.protocols, Seq: s.seq, Effects: s.effects}
	for id, op := range s.ops {
		state.Ops[id] = persistedOperation{Result: op.result, Ready: op.ready}
	}
	b, err := json.Marshal(state)
	if err != nil {
		return
	}
	if err = os.MkdirAll(filepath.Dir(s.stateFile), 0700); err != nil {
		return
	}
	tmp := s.stateFile + ".tmp"
	if err = os.WriteFile(tmp, b, 0600); err == nil {
		_ = os.Rename(tmp, s.stateFile)
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
	mux.HandleFunc("/oauth/token", s.handleToken)
	mux.HandleFunc("/__qualification/effects", s.handleEffects)
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil || r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_id") == "" || !strings.HasPrefix(r.Form.Get("client_secret_ref"), "vault://") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	token := "sim-" + base64.RawURLEncoding.EncodeToString(b)
	s.mu.Lock()
	s.tokens[token] = time.Now().Add(90 * time.Second)
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 90})
}

func (s *Server) authenticated(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Basic ") {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		return err == nil && strings.Contains(string(raw), ":vault://")
	}
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		s.mu.Lock()
		expiry, ok := s.tokens[token]
		s.mu.Unlock()
		if !ok || time.Now().After(expiry) {
			return false
		}
		if cert := r.Header.Get("X-MTLS-Certificate-Ref"); cert != "" && !strings.HasPrefix(cert, "vault://") {
			return false
		}
		return true
	}
	return false
}

func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Contas autenticadas enviam Basic, Bearer ou Bearer+mTLS simulado.
	// Contas NONE continuam permitidas para preservar os cenários legados.
	if r.Header.Get("Authorization") != "" && !s.authenticated(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	if id := s.protocols[req.ProtocolID]; id != "" {
		op := s.ops[id]
		var existing OperationResult
		if op != nil {
			existing = op.result
		}
		s.mu.Unlock()
		if op == nil {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if existing.Status == "PENDING" {
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(existing)
		return
	}
	s.seq++
	id := fmt.Sprintf("prov-req-%06d", s.seq)
	s.protocols[req.ProtocolID] = id
	s.effects++
	// Acknowledge the external effect only after its durable fixture oracle
	// has recorded the stable protocol mapping.
	s.ops[id] = &operation{result: OperationResult{ProviderRequestID: id, Status: "PENDING"}}
	s.saveStateLocked()
	s.mu.Unlock()
	status := "SUCCEEDED"
	if req.Fail {
		status = "FAILED"
	}
	result := OperationResult{ProviderRequestID: id, Status: status, Detail: "simulado"}

	switch req.Mode {
	case ModeAsyncPoll:
		s.mu.Lock()
		s.ops[id] = &operation{result: OperationResult{ProviderRequestID: id, Status: "PENDING"}}
		s.saveStateLocked()
		s.mu.Unlock()
		go s.resolveAfter(id, req.DelayMs, result)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(OperationResult{ProviderRequestID: id, Status: "PENDING"})
	case ModeAsyncCallback:
		s.mu.Lock()
		s.ops[id] = &operation{result: OperationResult{ProviderRequestID: id, Status: "PENDING"}}
		s.saveStateLocked()
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
		s.saveStateLocked()
		s.mu.Unlock()
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	}
}

func (s *Server) handleEffects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"effects": s.effects, "protocols": len(s.protocols)})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) resolveAfter(id string, delayMs int, result OperationResult) {
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	s.mu.Lock()
	s.ops[id] = &operation{result: result, ready: true}
	s.saveStateLocked()
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
