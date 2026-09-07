package orbita

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/libraclient"
	"ai-hub/hub/internal/platform/idgen"
)

// CreateRequest e o corpo de admissao (EXE-01/EXE-02).
type CreateRequest struct {
	Mode              string `json:"mode"` // SYNC | ASYNC | AUTO
	ProviderAccountID string `json:"provider_account_id"`
	ServiceCode       string `json:"service_code"`
	ServiceVersion    int    `json:"service_version"`
	Input             any    `json:"input"`
}

// Handlers expoe a API publica da Orbita: admissao e consulta
// unificada (EXE-01/EXE-07).
type Handlers struct {
	store      *Store
	atlas      *atlasclient.Client
	libra      *libraclient.Client
	dispatcher *Dispatcher
	finalizer  *Finalizer
	log        *slog.Logger
}

// NewHandlers cria os handlers HTTP da Orbita.
func NewHandlers(store *Store, atlas *atlasclient.Client, libra *libraclient.Client, dispatcher *Dispatcher, finalizer *Finalizer, log *slog.Logger) *Handlers {
	return &Handlers{store: store, atlas: atlas, libra: libra, dispatcher: dispatcher, finalizer: finalizer, log: log}
}

// Register registra as rotas publicas da Orbita.
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/protocols", h.handleCreate)
	mux.HandleFunc("/v1/protocols/", h.handleGet)
}

// RegisterInternal registra rotas internas, chamadas apenas por outras
// aplicacoes do hub (ex.: Pulsar buscando a mesma representacao final
// usada pelo GET publico — COM-01/COM-05). Sem verificacao de tenant:
// confia na rede interna do dominio, nao na identidade do cliente.
func (h *Handlers) RegisterInternal(mux *http.ServeMux) {
	mux.HandleFunc("/internal/protocols/", h.handleGetInternal)
}

func (h *Handlers) handleGetInternal(w http.ResponseWriter, r *http.Request) {
	protocolID := strings.TrimPrefix(r.URL.Path, "/internal/protocols/")
	p, err := h.store.GetByID(r.Context(), protocolID)
	if errors.Is(err, ErrProtocolNotFound) {
		writeErr(w, http.StatusNotFound, "not_found", "protocolo nao encontrado")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	h.respondWithProtocol(w, p, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"error": code, "message": msg})
}

// handleCreate implementa EXE-01 (admissao duravel) e EXE-02/EXE-14
// (modos de atendimento e SYNC direto). tenant_id vem de um header
// injetado pela borda confiavel (SEG-01); nesta referencia local, sem
// OAuth2/OIDC real, o header e aceito diretamente — placeholder de
// autenticacao, ver auditoria.
func (h *Handlers) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	tenantID := r.Header.Get("X-Tenant-Id")
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if tenantID == "" {
		writeErr(w, http.StatusUnauthorized, "missing_tenant", "X-Tenant-Id e obrigatorio")
		return
	}
	if idempotencyKey == "" {
		writeErr(w, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key e obrigatoria na criacao (EXE-01)")
		return
	}

	var req CreateRequest
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := json.Unmarshal(rawBody, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	mode := strings.ToUpper(req.Mode)
	if mode != "SYNC" && mode != "ASYNC" && mode != "AUTO" {
		writeErr(w, http.StatusBadRequest, "invalid_mode", "mode deve ser SYNC, ASYNC ou AUTO")
		return
	}

	// Hash semantico de idempotencia (EXE-01): nao inclui
	// timestamps de trace ou URLs temporarias — aqui, apenas o corpo
	// canonico decodificado.
	canon, _ := json.Marshal(req)
	sum := sha256.Sum256(canon)
	requestHash := hex.EncodeToString(sum[:])

	ctx := r.Context()

	if existing, err := h.store.FindByIdempotencyKey(ctx, tenantID, idempotencyKey, requestHash); err == nil {
		// EXE-01: mesma chave e mesmo hash retornam protocolo e estado
		// existentes — sem nova operacao.
		h.respondWithProtocol(w, existing, http.StatusOK)
		return
	} else if errors.Is(err, ErrIdempotencyConflict) {
		writeErr(w, http.StatusConflict, "idempotency_conflict", "Idempotency-Key reutilizada com payload diferente")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		// EXE-01/DAD-09: sem autoridade duravel disponivel para sequer
		// verificar idempotencia, o hub nao pode confirmar nem negar
		// aceite com seguranca — responde indisponibilidade, sem
		// vazar detalhe interno de infraestrutura ao cliente.
		h.log.Error("falha ao consultar idempotencia; autoridade duravel indisponivel", "error", err)
		writeErr(w, http.StatusServiceUnavailable, "admission_unavailable", "nao foi possivel verificar o aceite; tente novamente com a mesma chave")
		return
	}

	contract, err := h.atlas.Contract(ctx, tenantID)
	if err != nil {
		writeErr(w, http.StatusForbidden, "contract_not_found", "tenant sem contrato elegivel")
		return
	}

	now := time.Now().UTC()
	protocolID := idgen.New()
	commandID := idgen.New()
	clientDeadline := now.Add(time.Duration(contract.ClientSLASeconds) * time.Second)

	dispatchMode := dispatch.DispatchQueued
	if mode == "SYNC" {
		dispatchMode = dispatch.DispatchDirect
	}

	p := Protocol{
		ProtocolID: protocolID, TenantID: tenantID, IdempotencyKey: idempotencyKey,
		RequestHash: requestHash, RequestBody: canon, Mode: mode, DispatchMode: string(dispatchMode),
		CommandID: commandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: clientDeadline,
	}
	if err := h.store.Create(ctx, p); err != nil {
		// EXE-01: sem banco de admissao (ou falha equivalente),
		// responder indisponibilidade — nao confirmar aceite fictício.
		writeErr(w, http.StatusServiceUnavailable, "admission_unavailable", "nao foi possivel persistir o aceite")
		return
	}

	// FIN-06: para contrato com saldo estrito, nenhum passo externo e
	// liberado antes da reserva confirmada na autoridade financeira.
	if contract.StrictBalance {
		if err := h.libra.Reserve(ctx, tenantID, protocolID, contract.UnitPrice, "BRL"); err != nil {
			reason := "falha ao obter reserva financeira"
			if errors.Is(err, libraclient.ErrLimitExceeded) {
				reason = "limite estrito do tenant excedido"
			}
			body := FinalBody{ProtocolID: protocolID, ErrorCode: "STRICT_BALANCE_UNAVAILABLE", ErrorMessage: reason}
			_, _ = h.finalizer.Finalize(ctx, tenantID, protocolID, 0, StatusFailed, body, "STRICT_BALANCE_UNAVAILABLE")
			writeErr(w, http.StatusPaymentRequired, "strict_balance_unavailable", reason)
			return
		}
	}

	cmd := dispatch.Command{
		TenantID: tenantID, ProtocolID: protocolID, StepID: protocolID, CommandID: commandID,
		DispatchMode: dispatchMode, Epoch: 1, ServiceCode: req.ServiceCode, ServiceVersion: req.ServiceVersion,
		ProviderAccountID: req.ProviderAccountID, RequestBody: req.Input, StepDeadline: clientDeadline,
	}

	if mode == "SYNC" {
		h.handleSyncDispatch(w, r, tenantID, protocolID, commandID, cmd)
		return
	}

	// ASYNC/AUTO: 202 somente apos commit (ja ocorrido acima).
	if err := h.dispatcher.DispatchQueued(ctx, cmd); err != nil {
		h.log.Error("falha ao publicar comando em fila; protocolo permanece ACCEPTED para retomada", "error", err)
	}
	fresh, _ := h.store.Get(ctx, tenantID, protocolID)
	h.respondWithProtocol(w, fresh, http.StatusAccepted)
}

// handleSyncDispatch executa o percurso normativo de EXE-14: despacha
// diretamente, aguarda o fato externo na mesma chamada e decide o
// prazo antes de responder.
func (h *Handlers) handleSyncDispatch(w http.ResponseWriter, r *http.Request, tenantID, protocolID, commandID string, cmd dispatch.Command) {
	ctx, cancel := context.WithDeadline(r.Context(), cmd.StepDeadline)
	defer cancel()

	result, err := h.dispatcher.DispatchDirect(ctx, cmd)

	current, getErr := h.store.Get(r.Context(), tenantID, protocolID)
	if getErr != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", getErr.Error())
		return
	}
	if IsTerminal(current.Status) {
		// EXE-11: temporizador de deadline pode ja ter finalizado
		// (EXPIRED) enquanto aguardavamos o Cometa.
		h.respondWithProtocol(w, current, statusForTerminal(current.Status))
		return
	}

	if err != nil || result.Kind == dispatch.FactUnknown {
		// Timeout/incerteza: nao anunciar sucesso; se o deadline ja
		// passou, o timer o finalizara como EXPIRED; caso contrario,
		// devolve 504 sem sucesso fictício.
		writeErr(w, http.StatusGatewayTimeout, "provider_uncertain", "resultado do provedor incerto (UNKNOWN); consulte novamente com a mesma chave")
		return
	}

	status := StatusFailed
	if result.Kind == dispatch.FactSucceeded {
		status = StatusSucceeded
	}
	body := FinalBody{ProtocolID: protocolID, Result: result.ResponseBody, ErrorCode: result.ErrorCode, ErrorMessage: result.ErrorMessage}
	applied, err := h.finalizer.Finalize(r.Context(), tenantID, protocolID, current.Version, status, body, "")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if !applied {
		fresh, _ := h.store.Get(r.Context(), tenantID, protocolID)
		h.respondWithProtocol(w, fresh, statusForTerminal(fresh.Status))
		return
	}
	fresh, _ := h.store.Get(r.Context(), tenantID, protocolID)
	h.respondWithProtocol(w, fresh, http.StatusOK)
}

func statusForTerminal(s Status) int {
	if s == StatusExpired {
		return http.StatusGatewayTimeout
	}
	return http.StatusOK
}

// handleGet implementa o GET unificado de estado/resultado (EXE-07,
// COM-05): mesma representacao usada pelo webhook.
func (h *Handlers) handleGet(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-Id")
	if tenantID == "" {
		writeErr(w, http.StatusUnauthorized, "missing_tenant", "X-Tenant-Id e obrigatorio")
		return
	}
	protocolID := strings.TrimPrefix(r.URL.Path, "/v1/protocols/")
	p, err := h.store.Get(r.Context(), tenantID, protocolID)
	if errors.Is(err, ErrProtocolNotFound) {
		// EXE-16: inexistente ou de outro tenant recebem a mesma
		// resposta, sem revelar existencia (SEG-04).
		writeErr(w, http.StatusNotFound, "not_found", "protocolo nao encontrado")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	status := http.StatusOK
	if p.Status == StatusExpired {
		status = http.StatusOK // corpo de erro contratado, nao 5xx (EXE-07)
	}
	h.respondWithProtocol(w, p, status)
}

func (h *Handlers) respondWithProtocol(w http.ResponseWriter, p Protocol, status int) {
	resp := map[string]any{
		"protocol_id":    p.ProtocolID,
		"status":         p.Status,
		"result_version": p.ResultVersion,
		"mode":           p.Mode,
	}
	if len(p.FinalBody) > 0 {
		var fb any
		_ = json.Unmarshal(p.FinalBody, &fb)
		resp["final_body"] = fb
	}
	writeJSON(w, status, resp)
}
