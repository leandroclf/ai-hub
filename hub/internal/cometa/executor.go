package cometa

import (
	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/platform/egress"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/outbox"
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
	clients    *egress.Pool
	selfURL    string
	tokenCache *providerauth.TokenCache
}

// NewExecutor cria um Executor. selfURL e a base URL publica deste
// Cometa, usada para montar o endereco de callback informado ao
// provedor simulado em modo async_callback.
func NewExecutor(store *Store, atlas *atlasclient.Client, log *slog.Logger, selfURL string, tokenCache *providerauth.TokenCache) *Executor {
	return &Executor{store: store, atlas: atlas, log: log, clients: egress.NewPool(egress.FromEnv()), selfURL: selfURL, tokenCache: tokenCache}
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
	if cmd.CellID != os.Getenv("CELL_ID") || cmd.TenantID == "" || cmd.CommandID == "" || cmd.StepDeadline.IsZero() {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactRejected, ErrorCode: "invalid_command"}
	}
	if _, err := e.store.Get(ctx, operationID); err == nil {
		result, err := e.store.DurableResult(ctx, cmd)
		if err == nil {
			return result
		}
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactUnknown, ErrorCode: "custody_unavailable"}
	} else if !errors.Is(err, ErrOperationNotFound) {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactUnknown, ErrorCode: "persistence_unavailable"}
	}
	var snapshot atlas.OfferSnapshot
	if json.Unmarshal(cmd.ConfigSnapshot, &snapshot) != nil || snapshot.Binding.ID == "" {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactRejected, ErrorCode: "invalid_snapshot"}
	}
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	if err != nil || target.AdapterID == "" {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactRejected, ErrorCode: "adapter_not_qualified"}
	}

	cred, err := e.atlas.BoundCredential(ctx, cmd.TenantID, snapshot.Binding.ID, snapshot.Binding.Version)
	if err != nil {
		// SEG-05: sem fallback implicito. Recusa antes do envio.
		e.log.Warn("credencial indisponivel, recusando antes do envio (SEG-05)", "trace_id", cmd.TraceID,
			"tenant_id", cmd.TenantID, "provider_account_id", cmd.ProviderAccountID, "error", err)

		return dispatch.Result{
			CommandID: cmd.CommandID, OperationID: operationID,
			Kind: dispatch.FactRejected, ErrorCode: "credential_unavailable", ErrorMessage: err.Error(),
		}
	}
	e.log.Debug("credencial resolvida (SEG-05)", "trace_id", cmd.TraceID, "binding_id", cred.BindingID,
		"credential_mode", cred.CredentialMode, "settlement_party", cred.SettlementParty)

	if cred.BindingID != snapshot.Binding.ID {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactRejected, ErrorCode: "binding_changed"}
	}

	var pa atlasclient.ProviderAccount
	if json.Unmarshal(snapshot.Account.Data, &pa) != nil || snapshot.Account.ID != cmd.ProviderAccountID {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactRejected, ErrorCode: "account_snapshot_invalid"}
	}
	pa.ProviderAccountID = snapshot.Account.ID

	binding, _ := atlas.DecodeCatalogData(snapshot.Binding)
	claim, owned, err := e.store.PrepareSubmission(ctx, cmd, cred.BindingID, binding.SecretVersion)
	if err != nil {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactUnknown, ErrorCode: "preparation_unavailable"}
	}
	if !owned {
		r, err := e.store.DurableResult(ctx, cmd)
		if err == nil {
			return r
		}
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactUnknown}
	}
	attemptID := claim.AttemptID
	sentAt := time.Now()
	client, err := e.clients.Client(pa.BaseURL, 15*time.Second)
	if err != nil {
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "egress_refused", err)
	}

	fail := shouldFail(cmd.RequestBody)
	delayMs := requestedDelayMs(cmd.RequestBody)

	req := providersim.SubmitRequest{
		ProtocolID:      cmd.ProtocolID,
		Mode:            providersim.Mode(pa.ProviderMode),
		DelayMs:         delayMs,
		Fail:            fail,
		DropAfterEffect: shouldDropAfterEffect(cmd.RequestBody),
	}
	if pa.ProviderMode == string(providersim.ModeAsyncCallback) {
		req.CallbackURL = e.callbackURLFor(operationID, claim.CallbackToken)
	}

	e.log.Debug("enviando chamada ao provedor", "trace_id", cmd.TraceID, "operation_id", operationID,
		"provider_base_url", pa.BaseURL, "provider_mode", pa.ProviderMode, "attempt_id", attemptID)

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, pa.BaseURL+"/v1/operations", bytes.NewReader(body))
	if err != nil {
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "build_request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if err := e.tokenCache.Apply(ctx, client, pa.ProviderAccountID, providerauth.Config{
		BindingID: cred.BindingID, TenantID: cmd.TenantID, Environment: os.Getenv("ENVIRONMENT"), SecretVersion: binding.SecretVersion,
		AuthType: pa.AuthType, Username: pa.AuthUsername, SecretRef: cred.SecretRef, APIKeyHeader: pa.APIKeyHeader,
		TokenURL: pa.OAuthTokenURL, ClientID: pa.OAuthClientID,
		ClientSecretRef: cred.SecretRef, MTLSCertificateRef: pa.MTLSCertificateRef,
		TokenTTLSeconds: pa.TokenTTLSeconds,
	}, httpReq); err != nil {
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "provider_authentication", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		// Falha de comunicacao: efeito possivelmente enviado e
		// desconhecido (EXE-09/EXE-04) — nao inventar resposta final.
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "transport_error", err)
	}
	defer resp.Body.Close()

	receivedAt := time.Now()
	var result providersim.OperationResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, 256*1024)).Decode(&result); err != nil || result.ProviderRequestID == "" || (result.Status != "SUCCEEDED" && result.Status != "FAILED" && result.Status != "PENDING") {
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, "invalid_provider_response", fmt.Errorf("invalid provider response"))
	}
	if _, err := e.store.db.ExecContext(ctx, "UPDATE attempts SET sent_at=$2,received_at=$3 WHERE attempt_id=$1", attemptID, sentAt, receivedAt); err != nil {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactUnknown, ErrorCode: "receipt_unavailable"}
	}

	switch resp.StatusCode {
	case http.StatusOK: // Provider final must have explicit terminal semantics.
		if result.Status == "PENDING" {
			return e.communicationFailure(ctx, operationID, attemptID, sentAt, "invalid_final", fmt.Errorf("pending result on final response"))
		}
		return e.finalize(ctx, cmd, operationID, result)
	case http.StatusAccepted: // provedor assincrono: pendente
		if result.Status != "PENDING" {
			return e.communicationFailure(ctx, operationID, attemptID, sentAt, "invalid_acceptance", fmt.Errorf("terminal result on pending response"))
		}
		saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		durable, err := e.store.ConserveAcceptance(saveCtx, cmd, result.ProviderRequestID, pa.ProviderMode == string(providersim.ModeAsyncPoll))
		if err != nil {
			return dispatch.Result{CommandID: cmd.CommandID, OperationID: operationID, Kind: dispatch.FactUnknown, ErrorCode: "acceptance_custody_unavailable"}
		}
		return durable
	default:
		return e.communicationFailure(ctx, operationID, attemptID, sentAt, fmt.Sprintf("status_%d", resp.StatusCode), fmt.Errorf("status inesperado"))
	}
}

func (e *Executor) communicationFailure(ctx context.Context, operationID, attemptID string, sentAt time.Time, code string, cause error) dispatch.Result {
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	var cmd dispatch.Command
	var raw []byte
	if err := e.store.db.QueryRowContext(saveCtx, "SELECT command FROM operations WHERE operation_id=$1", operationID).Scan(&raw); err == nil {
		_ = json.Unmarshal(raw, &cmd)
	}
	result := dispatch.Result{CommandID: operationID, OperationID: operationID, Kind: dispatch.FactUnknown, ErrorCode: code}
	if cmd.CommandID == "" {
		return result
	}
	if _, err := e.store.db.ExecContext(saveCtx, "UPDATE attempts SET received_at=clock_timestamp(),error_code=$2 WHERE attempt_id=$1", attemptID, code); err != nil {
		return result
	}
	durable, err := e.store.ConserveObservation(saveCtx, cmd, result, "SUBMIT")
	if err != nil {
		return result
	}
	return durable
}

// finalize aplica um resultado conhecido (SUCCEEDED/FAILED), persiste
// o estado e publica o fato na outbox (COM-03) para Orbita/Libra.
func (e *Executor) finalize(ctx context.Context, cmd dispatch.Command, operationID string, result providersim.OperationResult) dispatch.Result {
	var snapshot atlas.OfferSnapshot
	if json.Unmarshal(cmd.ConfigSnapshot, &snapshot) != nil {
		return dispatch.Result{CommandID: operationID, Kind: dispatch.FactUnknown, ErrorCode: "snapshot_unavailable"}
	}
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	raw, _ := json.Marshal(result)
	if err == nil {
		_, err = atlas.TransformJSON(raw, nil, target.OutputSchema)
	}
	if err != nil {
		return dispatch.Result{CommandID: operationID, Kind: dispatch.FactUnknown, ErrorCode: "provider_output_contract_failed"}
	}
	kind := dispatch.FactSucceeded
	if result.Status == "FAILED" {
		kind = dispatch.FactFailed
	} else if result.Status != "SUCCEEDED" {
		return dispatch.Result{CommandID: operationID, Kind: dispatch.FactUnknown}
	}
	response := dispatch.Result{CommandID: operationID, OperationID: operationID, ProviderRequestID: result.ProviderRequestID, Kind: kind, ResponseBody: result}
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	durable, err := e.store.ConserveObservation(saveCtx, cmd, response, "PROVIDER")
	if err != nil {
		return dispatch.Result{CommandID: operationID, Kind: dispatch.FactUnknown, ErrorCode: "custody_unavailable"}
	}
	return durable
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

func (e *Executor) callbackURLFor(operationID, token string) string {
	return fmt.Sprintf("%s/callbacks/%s?token=%s", e.selfURL, operationID, url.QueryEscape(token))
}

// shouldFail permite injetar falha deterministica a partir do proprio
// payload de entrada (QUA-01: "falhas injetadas"), lendo um campo
// opcional "force_fail" quando o corpo chega como map[string]any.
func shouldFail(body any) bool {
	m, ok := inputObject(body)
	if !ok {
		return false
	}
	v, ok := m["force_fail"].(bool)
	return ok && v
}

// requestedDelayMs le um campo opcional "delay_ms" do corpo de
// entrada, usado para ensaiar deadlines e TTL de retry (EXE-10/EXE-11).
func requestedDelayMs(body any) int {
	m, ok := inputObject(body)
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

func shouldDropAfterEffect(body any) bool {
	m, ok := inputObject(body)
	if !ok {
		return false
	}
	v, ok := m["force_drop_after_effect"].(bool)
	return ok && v
}

func inputObject(body any) (map[string]any, bool) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, false
	}
	var m map[string]any
	err = json.Unmarshal(raw, &m)
	return m, err == nil && m != nil
}

// ApplyExternalObservation aplica uma observacao recebida por callback
// ou por polling ao consolidado da operacao (EXE-06): se ja concluida,
// trata como evidencia duplicada.
func (e *Executor) ApplyExternalObservation(ctx context.Context, operationID string, result providersim.OperationResult) (dispatch.Result, error) {
	_, err := e.store.Get(ctx, operationID)
	if err != nil {
		e.log.Warn("observacao para operacao desconhecida", "operation_id", operationID)
		return dispatch.Result{}, err
	}
	if result.Status == "PENDING" {
		return dispatch.Result{}, errors.New("callback pending is not a final observation")
	}
	var raw []byte
	var storedCorrelation string
	if err = e.store.db.QueryRowContext(ctx, "SELECT command,COALESCE(provider_request_id,'') FROM operations WHERE operation_id=$1", operationID).Scan(&raw, &storedCorrelation); err != nil {
		return dispatch.Result{}, err
	}
	if storedCorrelation != "" && storedCorrelation != result.ProviderRequestID {
		return dispatch.Result{}, errors.New("callback provider correlation mismatch")
	}
	var command dispatch.Command
	if json.Unmarshal(raw, &command) != nil {
		return dispatch.Result{}, errors.New("invalid stored callback command")
	}
	var snapshot atlas.OfferSnapshot
	if json.Unmarshal(command.ConfigSnapshot, &snapshot) != nil {
		return dispatch.Result{}, errors.New("callback snapshot unavailable")
	}
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	if err != nil {
		return dispatch.Result{}, errors.New("callback target unavailable")
	}
	resultRaw, err := json.Marshal(result)
	if err != nil || len(target.OutputSchema) == 0 {
		return dispatch.Result{}, errors.New("callback result unavailable")
	}
	if _, err := atlas.TransformJSON(resultRaw, nil, target.OutputSchema); err != nil {
		return dispatch.Result{}, errors.New("callback output contract failed")
	}
	var fullResult map[string]any
	if json.Unmarshal(resultRaw, &fullResult) != nil {
		return dispatch.Result{}, errors.New("callback result encoding failed")
	}
	response := dispatch.Result{CommandID: operationID, OperationID: operationID, ProviderRequestID: result.ProviderRequestID, ResponseBody: fullResult}
	if result.Status == "SUCCEEDED" {
		response.Kind = dispatch.FactSucceeded
	} else if result.Status == "FAILED" {
		response.Kind = dispatch.FactFailed
	} else {
		return dispatch.Result{}, errors.New("invalid callback status")
	}
	return e.store.ConserveObservation(ctx, command, response, "CALLBACK")
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
