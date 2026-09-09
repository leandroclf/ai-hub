package orbita

import (
	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/contracts/economics"
	"ai-hub/hub/internal/platform/auth"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/libraclient"
	"ai-hub/hub/internal/objectstore"
	"ai-hub/hub/internal/platform/idgen"
)

// CreateRequest e o corpo de admissao (EXE-01/EXE-02).
type CreateRequest struct {
	FileRefs          []string `json:"file_refs,omitempty"`
	Mode              string   `json:"mode"` // SYNC | ASYNC | AUTO
	ProviderAccountID string   `json:"provider_account_id"`
	ServiceCode       string   `json:"service_code"`
	ServiceVersion    int      `json:"service_version"`
	Input             any      `json:"input"`
}

// Handlers expoe a API publica da Orbita: admissao e consulta
// unificada (EXE-01/EXE-07).
type Handlers struct {
	files      *objectstore.Catalog
	store      *Store
	atlas      *atlasclient.Client
	libra      *libraclient.Client
	dispatcher *Dispatcher
	finalizer  *Finalizer
	log        *slog.Logger
}

func (h *Handlers) SetFileCatalog(c *objectstore.Catalog) { h.files = c }

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
	principal, ok := auth.FromContext(r.Context())
	if !ok || !principal.Workload || !auth.Authorize(r.Context(), "protocols:read", "") {
		auth.Error(w, 403, "forbidden")
		return
	}
	if r.Method != http.MethodGet {
		auth.Error(w, 405, "method_not_allowed")
		return
	}
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
	if p.CellID != principal.CellID {
		auth.Error(w, 404, "not_found")
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
	tenantID, authorized := auth.PublicTenant(r.Context(), "protocols:write")
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if !authorized {
		writeErr(w, http.StatusUnauthorized, "forbidden", "identidade com permissão de criação necessária")
		return
	}
	if idempotencyKey == "" || len(idempotencyKey) > 200 {
		writeErr(w, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key e obrigatoria na criacao (EXE-01)")
		return
	}

	// trace_id correlaciona esta admissao ponta a ponta nos logs de
	// Orbita/Cometa/Libra/Pulsar (correlacao por log, nao span de
	// tracing distribuido real — ver internal/platform/logging).
	traceID := idgen.New()
	w.Header().Set("X-Trace-Id", traceID)
	h.log.Info("admissao recebida", "trace_id", traceID, "tenant_id", tenantID, "idempotency_key", idempotencyKey)

	var req CreateRequest
	rawBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 256*1024))
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
	req.Mode = mode

	// Hash semantico de idempotencia (EXE-01): nao inclui
	// timestamps de trace ou URLs temporarias — aqui, apenas o corpo
	// canonico decodificado.
	canon, _ := json.Marshal(req)
	sum := sha256.Sum256(canon)
	requestHash := hex.EncodeToString(sum[:])
	h.log.Debug("hash de idempotencia calculado", "trace_id", traceID, "request_hash", requestHash, "mode", mode)

	ctx := r.Context()

	if existing, err := h.store.FindByIdempotencyKey(ctx, tenantID, idempotencyKey, requestHash); err == nil {
		// EXE-01: mesma chave e mesmo hash retornam protocolo e estado
		// existentes — sem nova operacao.
		h.respondExistingAdmission(w, r, existing)
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

	principal, _ := auth.FromContext(ctx)
	if len(req.FileRefs) > 10 {
		writeErr(w, 422, "too_many_files", "máximo de dez referências por protocolo")
		return
	}
	var fileRefs []objectstore.FileRef
	for _, id := range req.FileRefs {
		if h.files == nil {
			writeErr(w, 503, "file_custody_unavailable", "catálogo de arquivos indisponível")
			return
		}
		ref, err := h.files.Resolve(ctx, tenantID, id)
		if err != nil {
			writeErr(w, 422, "file_not_eligible", "referência de arquivo não elegível")
			return
		}
		fileRefs = append(fileRefs, ref)
	}
	snapshot, err := h.atlas.Offer(ctx, tenantID, principal.ApplicationID, req.ServiceCode, req.ProviderAccountID)
	if err != nil {
		writeErr(w, 403, "offer_not_eligible", "oferta publicada indisponível para aplicação e serviço")
		return
	}
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	if err != nil || !slices.Contains(target.Modes, mode) || target.ClientSLASeconds <= 0 {
		writeErr(w, 422, "mode_not_eligible", "modo não elegível no contrato publicado")
		return
	}
	profile, err := atlas.DecodeCatalogData(snapshot.TechnicalProfile)
	if err != nil {
		writeErr(w, 503, "profile_unavailable", "perfil indisponível")
		return
	}
	input, _ := json.Marshal(req.Input)
	transformed, err := atlas.TransformJSON(input, profile.InputMapping, target.InputSchema)
	if err != nil {
		writeErr(w, 422, "invalid_input", err.Error())
		return
	}
	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		writeErr(w, 503, "snapshot_unavailable", "snapshot indisponível")
		return
	}
	economic, sale, err := economics.Freeze(snapshot.PurchaseContract.ID, snapshot.PurchaseContract.Version, snapshot.PurchaseContract.Data, snapshot.SaleContract.ID, snapshot.SaleContract.Version, snapshot.SaleContract.Data)
	if err != nil {
		writeErr(w, 422, "economic_policy_invalid", err.Error())
		return
	}
	economicBytes, err := json.Marshal(economic)
	if err != nil {
		writeErr(w, 503, "snapshot_unavailable", "contrato indisponível")
		return
	}

	now := time.Now().UTC()
	protocolID := idgen.New()
	commandID := idgen.New()
	clientDeadline := now.Add(time.Duration(target.ClientSLASeconds) * time.Second)

	dispatchMode := dispatch.DispatchQueued
	if mode == "SYNC" {
		dispatchMode = dispatch.DispatchDirect
	}

	p := Protocol{
		ProtocolID: protocolID, TenantID: tenantID, ApplicationID: principal.ApplicationID, CellID: os.Getenv("CELL_ID"), IdempotencyKey: idempotencyKey,
		RequestHash: requestHash, RequestBody: canon, Mode: mode, DispatchMode: string(dispatchMode),
		CommandID: commandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: clientDeadline,
	}
	cmd := dispatch.Command{
		FileRefs: fileRefs,
		TenantID: tenantID, ApplicationID: principal.ApplicationID, CellID: p.CellID, ProtocolID: protocolID, StepID: protocolID, CommandID: commandID,
		TraceID: traceID, DispatchMode: dispatchMode, Epoch: 1, ServiceCode: req.ServiceCode, ServiceVersion: snapshot.Target.Version,
		ProviderAccountID: snapshot.SelectedRoute.ProviderAccountID, RequestBody: json.RawMessage(transformed), StepDeadline: clientDeadline,
		AcceptedAt: now, RetryDeadline: now.Add(time.Duration(target.RetryTTLSeconds) * time.Second), RetryTTLSeconds: target.RetryTTLSeconds, ConfigSnapshot: snapshotBytes, EconomicSnapshot: economicBytes,
	}
	accepted, created, err := h.store.Admit(ctx, p, cmd, principal.Subject, sale.StrictBalance)
	if errors.Is(err, ErrIdempotencyConflict) {
		writeErr(w, 409, "idempotency_conflict", "chave reutilizada com outro pedido")
		return
	}
	if err != nil {
		writeErr(w, 503, "admission_unavailable", "aceite não confirmado; repita a mesma chave")
		return
	}
	if !created {
		h.respondExistingAdmission(w, r, accepted)
		return
	}
	w.Header().Set("Location", "/v1/protocols/"+protocolID)
	w.Header().Set("X-Protocol-Id", protocolID)
	// Once acceptance commits, client disconnection cannot cancel custody work.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
	defer cancel()
	if sale.StrictBalance {
		if err = h.libra.ReserveExact(ctx, tenantID, protocolID, string(sale.ReservationAmount), sale.Currency); err != nil {
			h.respondWithProtocol(w, p, http.StatusServiceUnavailable)
			return
		}
		if _, err = h.store.db.ExecContext(ctx, "UPDATE command_intents SET state='READY' WHERE command_id=$1 AND state='WAITING_RESERVATION'", commandID); err != nil {
			h.respondWithProtocol(w, p, 503)
			return
		}
	}

	h.log.Debug("comando de despacho construido", "trace_id", traceID, "protocol_id", protocolID,
		"command_id", commandID, "dispatch_mode", dispatchMode, "provider_account_id", req.ProviderAccountID)

	if mode == "SYNC" {
		h.handleSyncDispatch(w, r, tenantID, protocolID, commandID, cmd)
		return
	}

	// The durable intent publisher owns delivery and retry after admission.
	if mode == "AUTO" {
		h.respondAuto(w, r, accepted, target.AutoWaitSeconds)
		return
	}

	fresh, err := h.store.Get(ctx, tenantID, protocolID)
	if err != nil {
		acceptedError(w, protocolID, 503, "protocol_temporarily_unavailable")
		return
	}
	h.respondWithProtocol(w, fresh, http.StatusAccepted)
}

// handleSyncDispatch executa o percurso normativo de EXE-14: despacha
// diretamente, aguarda o fato externo na mesma chamada e decide o
// prazo antes de responder.
func (h *Handlers) handleSyncDispatch(w http.ResponseWriter, r *http.Request, tenantID, protocolID, commandID string, cmd dispatch.Command) {
	ctx, cancel := context.WithDeadline(context.WithoutCancel(r.Context()), cmd.StepDeadline)
	defer cancel()

	result, err := h.dispatcher.DispatchDirect(ctx, cmd)

	current, getErr := h.store.Get(r.Context(), tenantID, protocolID)
	if getErr != nil {
		acceptedError(w, protocolID, 503, "protocol_temporarily_unavailable")
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
		acceptedError(w, protocolID, 504, "provider_uncertain")
		return
	}

	status := StatusFailed
	if result.Kind == dispatch.FactSucceeded {
		status = StatusSucceeded
	}
	body := FinalBody{ProtocolID: protocolID, Result: result.ResponseBody, ErrorCode: result.ErrorCode, ErrorMessage: result.ErrorMessage}
	applied, err := h.finalizer.Finalize(r.Context(), cmd.TraceID, tenantID, protocolID, current.Version, status, body, "")
	if err != nil {
		acceptedError(w, protocolID, 503, "final_custody_unavailable")
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

func acceptedError(w http.ResponseWriter, id string, status int, code string) {
	w.Header().Set("Location", "/v1/protocols/"+id)
	writeJSON(w, status, map[string]any{"protocol_id": id, "query_url": "/v1/protocols/" + id, "error": code, "message": "Aceite conservado; consulte o protocolo ou repita a mesma chave."})
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
	if r.Method != http.MethodGet {
		auth.Error(w, 405, "method_not_allowed")
		return
	}
	tenantID, authorized := auth.PublicTenant(r.Context(), "protocols:read")
	if !authorized {
		writeErr(w, http.StatusUnauthorized, "forbidden", "identidade com permissão de leitura necessária")
		return
	}
	protocolID := strings.TrimPrefix(r.URL.Path, "/v1/protocols/")
	p, err := h.store.Get(r.Context(), tenantID, protocolID)
	principal, _ := auth.FromContext(r.Context())
	if err == nil && p.ApplicationID != principal.ApplicationID {
		writeErr(w, 404, "not_found", "protocolo não encontrado")
		return
	}
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
	w.Header().Set("Cache-Control", "no-store")
	if p.ProtocolID != "" {
		w.Header().Set("Location", "/v1/protocols/"+p.ProtocolID)
	}
	if len(p.FinalRepresentation) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		_, _ = w.Write(p.FinalRepresentation)
		return
	}

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
