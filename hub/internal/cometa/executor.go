package cometa

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/outbox"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
	"ai-hub/hub/internal/providersim"
)

// OperationFact e o payload publicado na outbox (aggregate_type
// "operation") quando Cometa observa um fato externo — consumido por
// Orbita (para finalizar protocolos ASYNC) e por Libra (para custo).
type OperationFact struct {
	ProtocolID        string `json:"protocol_id"`
	TraceID           string `json:"trace_id,omitempty"`
	OperationID       string `json:"operation_id"`
	ProviderAccountID string `json:"provider_account_id"`
	ProviderRequestID string `json:"provider_request_id,omitempty"`
	Kind              string `json:"kind"` // SUCCEEDED | FAILED | UNKNOWN
	ResponseBody      any    `json:"response_body,omitempty"`
	ErrorMessage      string `json:"error_message,omitempty"`
}

// Executor resolve credencial, chama o provedor (sincrono, poll ou
// callback conforme provider_mode homologado da conta — CAT-11) e
// persiste evidencia (EXE-04).
type Executor struct {
	store      *Store
	atlas      *atlasclient.Client
	log        *slog.Logger
	client     *http.Client
	selfURL    string
	tokenCache *providerauth.TokenCache
}

// NewExecutor cria um Executor. selfURL e a base URL publica deste
// Cometa, usada para montar o endereco de callback informado ao
// provedor simulado em modo async_callback.
func NewExecutor(store *Store, atlas *atlasclient.Client, log *slog.Logger, selfURL string, tokenCache *providerauth.TokenCache) *Executor {
	return &Executor{store: store, atlas: atlas, log: log, client: &http.Client{Timeout: 15 * time.Second}, selfURL: selfURL, tokenCache: tokenCache}
}

// Execute processa um dispatch.Command: cria a operacao, resolve
// credencial (SEG-05), chama o provedor simulado segundo seu
// provider_mode homologado e persiste evidencia. Retorna o
// dispatch.Result que a Orbita usa (via HTTP em SYNC, via evento em
// ASYNC) para decidir a conclusao do passo (EXE-03: "Cometa decide
// apenas o fato externo; Orbita decide a conclusao").
func (e *Executor) Execute(ctx context.Context, cmd dispatch.Command) dispatch.Result {
	// EXE-15 (posse unica): a identidade da operacao deriva do
	// command_id, nao de um novo UUID a cada chamada. Isso torna
	// reentrega de transporte/mensagem idempotente: uma operacao ja
	// criada para este command_id nunca e reenviada ao provedor.
	operationID := cmd.CommandID
	e.log.Debug("comando recebido para execucao", "trace_id", cmd.TraceID, "protocol_id", cmd.ProtocolID,
		"operation_id", operationID, "dispatch_mode", cmd.DispatchMode, "provider_account_id", cmd.ProviderAccountID)
	if existing, err := e.store.Get(ctx, operationID); err == nil {
		switch existing.State {
		case StateSucceeded, StateFailed:
			e.log.Info("comando repetido: devolvendo operacao ja concluida", "operation_id", operationID)
			return dispatch.Result{
				CommandID: cmd.CommandID, OperationID: operationID,
				ProviderRequestID: existing.ProviderRequestID.String,
				Kind:              factKindFor(existing.State),
			}
		default:
			e.log.Info("comando repetido: operacao ja em andamento, sem novo envio", "operation_id", operationID, "state", existing.State)
			return dispatch.Result{CommandID: cmd.CommandID, OperationID: operationID, Kind: dispatch.FactUnknown}
		}
	}

	cred, err := e.atlas.ResolveCredential(ctx, cmd.TenantID, cmd.ProviderAccountID)
	if err != nil {
		// SEG-05: sem fallback implicito. Recusa antes do envio.
		e.log.Warn("credencial indisponivel, recusando antes do envio (SEG-05)", "trace_id", cmd.TraceID,
			"tenant_id", cmd.TenantID, "provider_account_id", cmd.ProviderAccountID, "error", err)
		_ = e.store.CreateOperation(ctx, Operation{
			OperationID: operationID, ProtocolID: cmd.ProtocolID,
			ProviderAccountID: cmd.ProviderAccountID, State: StateFailed,
		})
		return dispatch.Result{
			CommandID: cmd.CommandID, OperationID: operationID,
			Kind: dispatch.FactRejected, ErrorCode: "credential_unavailable", ErrorMessage: err.Error(),
		}
	}
	e.log.Debug("credencial resolvida (SEG-05)", "trace_id", cmd.TraceID, "binding_id", cred.BindingID,
		"credential_mode", cred.CredentialMode, "settlement_party", cred.SettlementParty)

	if err := e.store.CreateOperation(ctx, Operation{
		OperationID:         operationID,
		ProtocolID:          cmd.ProtocolID,
		ProviderAccountID:   cmd.ProviderAccountID,
		CredentialBindingID: nullable(cred.BindingID),
		State:               StatePrepared,
	}); err != nil {
		e.log.Error("falha ao persistir operacao antes do envio", "error", err)
		return dispatch.Result{CommandID: cmd.CommandID, OperationID: operationID, Kind: dispatch.FactUnknown, ErrorMessage: "falha ao persistir operacao"}
	}

	pa, err := e.atlas.ProviderAccount(ctx, cmd.ProviderAccountID)
	if err != nil {
		return dispatch.Result{CommandID: cmd.CommandID, OperationID: operationID, Kind: dispatch.FactUnknown, ErrorMessage: "conta de provedor nao resolvida"}
	}

	attemptID := idgen.New()
	sentAt := time.Now()
	_ = e.store.UpdateState(ctx, operationID, StateSubmitting, "")

	fail := shouldFail(cmd.RequestBody)
	delayMs := requestedDelayMs(cmd.RequestBody)

	req := providersim.SubmitRequest{
		ProtocolID: cmd.ProtocolID,
		Mode:       providersim.Mode(pa.ProviderMode),
		DelayMs:    delayMs,
		Fail:       fail,
	}
	if pa.ProviderMode == string(providersim.ModeAsyncCallback) {
		req.CallbackURL = e.callbackURLFor(operationID)
	}

	e.log.Debug("enviando chamada ao provedor", "trace_id", cmd.TraceID, "operation_id", operationID,
		"provider_base_url", pa.BaseURL, "provider_mode", pa.ProviderMode, "attempt_id", attemptID)

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, pa.BaseURL+"/v1/operations", bytes.NewReader(body))
	if err != nil {
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "build_request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if err := e.tokenCache.Apply(ctx, e.client, pa.ProviderAccountID, providerauth.Config{
		AuthType: pa.AuthType, Username: pa.AuthUsername, SecretRef: pa.AuthSecretRef,
		TokenURL: pa.OAuthTokenURL, ClientID: pa.OAuthClientID,
		ClientSecretRef: pa.OAuthClientSecretRef, MTLSCertificateRef: pa.MTLSCertificateRef,
		TokenTTLSeconds: pa.TokenTTLSeconds,
	}, httpReq); err != nil {
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "provider_authentication", err)
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		// Falha de comunicacao: efeito possivelmente enviado e
		// desconhecido (EXE-09/EXE-04) — nao inventar resposta final.
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "transport_error", err)
	}
	defer resp.Body.Close()

	receivedAt := time.Now()
	var result providersim.OperationResult
	_ = json.NewDecoder(resp.Body).Decode(&result)
	_ = e.store.RecordAttempt(ctx, attemptID, operationID, "SUBMIT", &sentAt, &receivedAt, "")

	switch resp.StatusCode {
	case http.StatusOK: // provedor sincrono: fato final ja conhecido
		return e.finalize(ctx, cmd, operationID, result)
	case http.StatusAccepted: // provedor assincrono: pendente
		_ = e.store.UpdateState(ctx, operationID, StateAcceptedExternal, result.ProviderRequestID)
		if pa.ProviderMode == string(providersim.ModeAsyncPoll) {
			deadline := cmd.StepDeadline
			_ = e.store.SchedulePolling(ctx, operationID, time.Now().Add(2*time.Second), deadline, 2)
		}
		return dispatch.Result{
			CommandID: cmd.CommandID, OperationID: operationID,
			ProviderRequestID: result.ProviderRequestID, Kind: dispatch.FactUnknown,
		}
	default:
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, fmt.Sprintf("status_%d", resp.StatusCode), fmt.Errorf("status inesperado"))
	}
}

func (e *Executor) communicationFailure(ctx context.Context, operationID, attemptID string, sentAt time.Time, code string, err error) dispatch.Result {
	now := time.Now()
	_ = e.store.RecordAttempt(ctx, attemptID, operationID, "SUBMIT", &sentAt, &now, code)
	_ = e.store.UpdateState(ctx, operationID, StateUnknown, "")
	e.log.Warn("comunicacao com provedor falhou; estado permanece UNKNOWN", "operation_id", operationID, "error", err)
	return dispatch.Result{CommandID: "", OperationID: operationID, Kind: dispatch.FactUnknown, ErrorCode: code, ErrorMessage: err.Error()}
}

// finalize aplica um resultado conhecido (SUCCEEDED/FAILED), persiste
// o estado e publica o fato na outbox (COM-03) para Orbita/Libra.
func (e *Executor) finalize(ctx context.Context, cmd dispatch.Command, operationID string, result providersim.OperationResult) dispatch.Result {
	state := StateSucceeded
	kind := dispatch.FactSucceeded
	if result.Status == "FAILED" {
		state = StateFailed
		kind = dispatch.FactFailed
	}
	_ = e.store.UpdateState(ctx, operationID, state, result.ProviderRequestID)
	_ = e.store.ClearPolling(ctx, operationID)

	fact := OperationFact{
		ProtocolID: cmd.ProtocolID, TraceID: cmd.TraceID, OperationID: operationID,
		ProviderAccountID: cmd.ProviderAccountID, ProviderRequestID: result.ProviderRequestID,
		Kind: string(kind), ResponseBody: cmd.RequestBody, ErrorMessage: result.Detail,
	}
	e.log.Info("operacao finalizada", "trace_id", cmd.TraceID, "protocol_id", cmd.ProtocolID,
		"operation_id", operationID, "kind", kind)
	if err := e.publishOperationFact(ctx, operationID, fact); err != nil {
		e.log.Error("falha ao publicar fato de operacao na outbox", "trace_id", cmd.TraceID, "error", err)
	}

	return dispatch.Result{
		CommandID: cmd.CommandID, OperationID: operationID,
		ProviderRequestID: result.ProviderRequestID, Kind: kind, ResponseBody: cmd.RequestBody,
	}
}

func (e *Executor) publishOperationFact(ctx context.Context, operationID string, fact OperationFact) error {
	tx, err := e.store.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := outbox.Enqueue(ctx, tx, "operation", operationID, "operation.observed", fact); err != nil {
		return err
	}
	return tx.Commit()
}

func (e *Executor) callbackURLFor(operationID string) string {
	return fmt.Sprintf("%s/internal/callbacks/%s", e.selfURL, operationID)
}

// shouldFail permite injetar falha deterministica a partir do proprio
// payload de entrada (QUA-01: "falhas injetadas"), lendo um campo
// opcional "force_fail" quando o corpo chega como map[string]any.
func shouldFail(body any) bool {
	m, ok := body.(map[string]any)
	if !ok {
		return false
	}
	v, ok := m["force_fail"].(bool)
	return ok && v
}

// requestedDelayMs le um campo opcional "delay_ms" do corpo de
// entrada, usado para ensaiar deadlines e TTL de retry (EXE-10/EXE-11).
func requestedDelayMs(body any) int {
	m, ok := body.(map[string]any)
	if !ok {
		return 0
	}
	switch v := m["delay_ms"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

// ApplyExternalObservation aplica uma observacao recebida por callback
// ou por polling ao consolidado da operacao (EXE-06): se ja concluida,
// trata como evidencia duplicada.
func (e *Executor) ApplyExternalObservation(ctx context.Context, cmdCtx dispatch.Command, operationID string, result providersim.OperationResult) {
	op, err := e.store.Get(ctx, operationID)
	if err != nil {
		e.log.Warn("observacao para operacao desconhecida", "operation_id", operationID)
		return
	}
	if op.State == StateSucceeded || op.State == StateFailed {
		e.log.Info("observacao duplicada ignorada (uma transicao final ja aplicada)", "operation_id", operationID)
		return
	}
	if result.Status == "PENDING" {
		return
	}
	e.finalize(ctx, dispatch.Command{ProtocolID: op.ProtocolID, ProviderAccountID: op.ProviderAccountID}, operationID, result)
}

func factKindFor(state State) dispatch.FactKind {
	if state == StateSucceeded {
		return dispatch.FactSucceeded
	}
	return dispatch.FactFailed
}

func nullable(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
