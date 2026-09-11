package cometa

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/outbox"
	"ai-hub/hub/internal/platform/idgen"
)

type Submission struct {
	Command          dispatch.Command
	Owner, AttemptID string
	Epoch            int64
	// CallbackToken is a one-operation capability sent only to the provider
	// in its callback URL. Its hash, never the capability itself, is durable.
	CallbackToken string
}

var ErrCallbackInboxQuota = errors.New("callback inbox quota exceeded")

const callbackInboxMaxReceived = 10000

// PrepareSubmission grants the first submitter exclusive ownership and records
// its attempt before any external I/O. A duplicate never acquires permission to
// resubmit an operation whose effect may already exist.
func (s *Store) PrepareSubmission(ctx context.Context, cmd dispatch.Command, binding, secretVersion string) (Submission, bool, error) {
	raw, err := json.Marshal(cmd)
	if err != nil {
		return Submission{}, false, err
	}
	if cmd.TenantID == "" || cmd.CellID == "" || cmd.ApplicationID == "" || binding == "" {
		return Submission{}, false, errors.New("missing command identity")
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	callbackToken, err := newCallbackToken()
	if err != nil {
		return Submission{}, false, err
	}
	claim := Submission{Command: cmd, Owner: idgen.New(), AttemptID: idgen.New(), Epoch: 1, CallbackToken: callbackToken}
	tokenHash := callbackTokenHash(callbackToken)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Submission{}, false, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `INSERT INTO operations(operation_id,protocol_id,provider_account_id,credential_binding_id,secret_version_id,tenant_id,application_id,cell_id,command,command_hash,callback_token_hash,state,submit_owner,submit_epoch,submit_lease_until) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'SUBMITTING',$12,1,clock_timestamp()+interval '30 seconds') ON CONFLICT(operation_id) DO NOTHING`, cmd.CommandID, cmd.ProtocolID, cmd.ProviderAccountID, binding, secretVersion, cmd.TenantID, cmd.ApplicationID, cmd.CellID, raw, hash, tokenHash, claim.Owner)
	if err != nil {
		return Submission{}, false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Submission{}, false, err
	}
	if n == 0 {
		var original string
		if err = tx.QueryRowContext(ctx, "SELECT command_hash FROM operations WHERE operation_id=$1", cmd.CommandID).Scan(&original); err != nil {
			return Submission{}, false, err
		}
		if original != hash {
			return Submission{}, false, errors.New("command identity conflict")
		}
		return Submission{}, false, nil
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO attempts(attempt_id,operation_id,attempt_type,owner,epoch) VALUES($1,$2,'SUBMIT',$3,1)`, claim.AttemptID, cmd.CommandID, claim.Owner)
	if err != nil {
		return Submission{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return Submission{}, false, err
	}
	return claim, true, nil
}

func newCallbackToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func callbackTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// AuthenticateCallback verifies the per-operation callback capability without
// exposing it in storage or accepting a callback for a different operation.
func (s *Store) AuthenticateCallback(ctx context.Context, operationID, token string) error {
	if operationID == "" || token == "" {
		return errors.New("missing callback capability")
	}
	var expected string
	if err := s.db.QueryRowContext(ctx, `SELECT callback_token_hash FROM operations WHERE operation_id=$1`, operationID).Scan(&expected); err != nil {
		return err
	}
	actual := callbackTokenHash(token)
	if expected == "" || subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) != 1 {
		return errors.New("invalid callback capability")
	}
	return nil
}

// StoreOrphanCallback conserva um callback válido sintaticamente cuja
// operação ainda não está disponível nesta autoridade. O token é guardado
// somente como hash; a reconciliação posterior decide se ele pertence à
// operação que apareceu.
func (s *Store) StoreOrphanCallback(ctx context.Context, operationID, token string, body []byte) error {
	if operationID == "" || token == "" || len(body) == 0 || len(body) > 512*1024 {
		return errors.New("invalid orphan callback")
	}
	sum := sha256.Sum256(body)
	tokenHash := callbackTokenHash(token)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// A bounded global counter prevents an unknown operation id from being
	// used as an unbounded ingress queue. The advisory lock makes the quota
	// decision serializable without holding a row lock for every orphan.
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(72820311)"); err != nil {
		return err
	}
	var known bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM callback_inbox
		WHERE operation_id=$1 AND body_sha256=$2 AND token_hash=$3
	)`, operationID, hex.EncodeToString(sum[:]), tokenHash).Scan(&known); err != nil {
		return err
	}
	if !known {
		var received int
		if err = tx.QueryRowContext(ctx, "SELECT count(*) FROM callback_inbox WHERE disposition='RECEIVED'").Scan(&received); err != nil {
			return err
		}
		if received >= callbackInboxMaxReceived {
			return ErrCallbackInboxQuota
		}
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO callback_inbox(inbox_id,operation_id,token_hash,body_sha256,body)
		VALUES($1,$2,$3,$4,$5)
		ON CONFLICT(operation_id,body_sha256,token_hash)
		DO UPDATE SET occurrences=callback_inbox.occurrences+1
	`, idgen.New(), operationID, tokenHash, hex.EncodeToString(sum[:]), body); err != nil {
		return err
	}
	return tx.Commit()
}

// PruneCallbackInbox removes only processed evidence older than the explicit
// retention window. RECEIVED items are never deleted by maintenance, so a
// temporary outage cannot turn into silent loss of a callback obligation.
func (s *Store) PruneCallbackInbox(ctx context.Context, olderThan time.Duration, limit int) (int, error) {
	if olderThan < 24*time.Hour || limit < 1 || limit > 1000 {
		return 0, errors.New("invalid callback inbox retention")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM callback_inbox
		WHERE inbox_id IN (
			SELECT inbox_id FROM callback_inbox
			WHERE disposition IN ('APPLIED','REJECTED')
			  AND processed_at IS NOT NULL
			  AND processed_at < clock_timestamp()-make_interval(secs=>$1::double precision)
			ORDER BY processed_at
			LIMIT $2
		)`, olderThan.Seconds(), limit)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), nil
}

// ReconcileCallbackInbox reapplies callbacks received before operation
// correlation. Invalid capabilities are retained as rejected evidence and
// never reach the external-observation state machine.
func (s *Store) ReconcileCallbackInbox(ctx context.Context, apply func(context.Context, string, dispatch.Result) error) (int, error) {
	return s.ReconcileCallbackInboxBatch(ctx, idgen.New(), 50, apply)
}

// ReconcileCallbackInboxBatch claims a bounded set before doing any nested
// SQL/application work. Expired leases can be reclaimed by another replica;
// epoch prevents a stale worker from disposing a newer claim.
func (s *Store) ReconcileCallbackInboxBatch(ctx context.Context, owner string, limit int, apply func(context.Context, string, dispatch.Result) error) (int, error) {
	if owner == "" || limit < 1 {
		return 0, errors.New("claim de inbox inválido")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `
		SELECT i.inbox_id, i.operation_id, i.token_hash, i.body,
		       o.callback_token_hash, o.command, i.claim_epoch + 1, i.processing_attempts + 1
		FROM callback_inbox i
		JOIN operations o ON o.operation_id=i.operation_id
		WHERE i.disposition='RECEIVED' AND (i.lease_until IS NULL OR i.lease_until < clock_timestamp())
		ORDER BY i.received_at
		FOR UPDATE OF i SKIP LOCKED LIMIT $1`, limit)
	if err != nil {
		return 0, err
	}
	type item struct {
		inboxID, operationID, receivedHash, expectedHash string
		body, commandRaw                                 []byte
		epoch, processingAttempts                        int64
	}
	items := make([]item, 0, limit)
	for rows.Next() {
		var v item
		if err := rows.Scan(&v.inboxID, &v.operationID, &v.receivedHash, &v.body, &v.expectedHash, &v.commandRaw, &v.epoch, &v.processingAttempts); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, v)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	for _, v := range items {
		if _, err := tx.ExecContext(ctx, `UPDATE callback_inbox SET claim_owner=$2,claim_epoch=$3,lease_until=clock_timestamp()+interval '30 seconds',processing_attempts=processing_attempts+1 WHERE inbox_id=$1`, v.inboxID, owner, v.epoch); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	count := 0
	for _, v := range items {
		var inboxID, operationID, receivedHash, expectedHash string
		inboxID, operationID, receivedHash, expectedHash = v.inboxID, v.operationID, v.receivedHash, v.expectedHash
		body, commandRaw := v.body, v.commandRaw
		dispose := func(disposition, lastError string) error {
			_, err := s.db.ExecContext(ctx, `UPDATE callback_inbox SET disposition=$1,processed_at=CASE WHEN $1<>'RECEIVED' THEN clock_timestamp() ELSE processed_at END,claim_owner=NULL,lease_until=NULL,last_error=$4 WHERE inbox_id=$2 AND disposition='RECEIVED' AND claim_owner=$3 AND claim_epoch=$5`, disposition, inboxID, owner, lastError, v.epoch)
			return err
		}
		if subtle.ConstantTimeCompare([]byte(receivedHash), []byte(expectedHash)) != 1 {
			if err := dispose("REJECTED", "callback capability mismatch"); err != nil {
				return count, err
			}
			continue
		}
		var result providersimOperationResult
		if err := json.Unmarshal(body, &result); err != nil {
			_ = dispose("REJECTED", "invalid callback JSON")
			continue
		}
		if result.Status != "SUCCEEDED" && result.Status != "FAILED" {
			if err := dispose("REJECTED", "non-terminal callback status"); err != nil {
				return count, err
			}
			continue
		}
		var command dispatch.Command
		if err := json.Unmarshal(commandRaw, &command); err != nil {
			_ = dispose("REJECTED", "invalid stored command")
			continue
		}
		if err := apply(ctx, operationID, result.toDispatchResult()); err != nil {
			disposition := "RECEIVED"
			if v.processingAttempts >= 3 {
				disposition = "REJECTED"
			}
			_ = dispose(disposition, err.Error())
			continue
		}
		if err := dispose("APPLIED", ""); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// providersimOperationResult is kept local to avoid coupling custody to the
// provider simulator package in production reconciliation code.
type providersimOperationResult struct {
	ProviderRequestID string `json:"provider_request_id"`
	Status            string `json:"status"`
	Detail            string `json:"detail,omitempty"`
}

func (r providersimOperationResult) toDispatchResult() dispatch.Result {
	kind := dispatch.FactUnknown
	if r.Status == "SUCCEEDED" {
		kind = dispatch.FactSucceeded
	}
	if r.Status == "FAILED" {
		kind = dispatch.FactFailed
	}
	return dispatch.Result{Kind: kind, ProviderRequestID: r.ProviderRequestID, ResponseBody: map[string]any{"detail": r.Detail}}
}

func (s *Store) DurableResult(ctx context.Context, cmd dispatch.Command) (dispatch.Result, error) {
	var raw []byte
	var tenant, cell string
	err := s.db.QueryRowContext(ctx, "SELECT result,tenant_id,cell_id FROM operations WHERE operation_id=$1", cmd.CommandID).Scan(&raw, &tenant, &cell)
	if err != nil {
		return dispatch.Result{}, err
	}
	if tenant != cmd.TenantID || cell != cmd.CellID {
		return dispatch.Result{}, errors.New("operation outside scope")
	}
	if len(raw) == 0 {
		return dispatch.Result{CommandID: cmd.CommandID, OperationID: cmd.CommandID, Kind: dispatch.FactUnknown, Durable: true}, nil
	}
	var r dispatch.Result
	if err = json.Unmarshal(raw, &r); err != nil {
		return dispatch.Result{}, err
	}
	return r, nil
}

// ConserveObservation commits the real observation, aggregate result and outbox
// together. Concurrent polling/callback replies produce only one terminal fact.
func (s *Store) ConserveObservation(ctx context.Context, cmd dispatch.Command, r dispatch.Result, source string, attemptID ...string) (dispatch.Result, error) {
	return s.conserveObservation(ctx, cmd, r, source, false, false, firstString(attemptID))
}

// ConserveAcceptance keeps provider correlation, receipt, result, polling
// obligation and outbox in one commit. No caller can acknowledge a partial save.
func (s *Store) ConserveAcceptance(ctx context.Context, cmd dispatch.Command, providerRequestID string, poll bool, attemptID ...string) (dispatch.Result, error) {
	if providerRequestID == "" {
		return dispatch.Result{}, errors.New("missing provider correlation")
	}
	return s.conserveObservation(ctx, cmd, dispatch.Result{Kind: dispatch.FactUnknown, ProviderRequestID: providerRequestID}, "SUBMIT_ACCEPTED", true, poll, firstString(attemptID))
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (s *Store) conserveObservation(ctx context.Context, cmd dispatch.Command, r dispatch.Result, source string, pending, poll bool, attemptID string) (dispatch.Result, error) {
	if r.Kind != dispatch.FactSucceeded && r.Kind != dispatch.FactFailed && r.Kind != dispatch.FactUnknown {
		return dispatch.Result{}, errors.New("invalid observed kind")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dispatch.Result{}, err
	}
	defer tx.Rollback()
	var state, tenant, cell string
	var original, result []byte
	err = tx.QueryRowContext(ctx, "SELECT state,tenant_id,cell_id,command,result FROM operations WHERE operation_id=$1 FOR UPDATE", cmd.CommandID).Scan(&state, &tenant, &cell, &original, &result)
	if err != nil {
		return dispatch.Result{}, err
	}
	if tenant != cmd.TenantID || cell != cmd.CellID {
		return dispatch.Result{}, errors.New("observation outside scope")
	}
	if json.Unmarshal(original, &cmd) != nil {
		return dispatch.Result{}, errors.New("invalid stored command")
	}
	// A terminal operation may still receive a callback or poll result. Keep
	// the receipt, but classify a different terminal fact explicitly so the
	// reconciler/finance pipeline can investigate it without reopening the
	// protocol or treating it as a new effect.
	if state == string(StateSucceeded) || state == string(StateFailed) || state == string(StateCancelled) {
		var existing dispatch.Result
		if len(result) > 0 && json.Unmarshal(result, &existing) == nil &&
			(existing.Kind != r.Kind || (existing.ProviderRequestID != "" && r.ProviderRequestID != "" && existing.ProviderRequestID != r.ProviderRequestID)) {
			source += "_CONFLICT"
		}
	}
	r.CommandID = cmd.CommandID
	r.OperationID = cmd.CommandID
	r.Durable = true
	r.EvidenceID = idgen.New()
	raw, err := json.Marshal(r)
	if err != nil {
		return dispatch.Result{}, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO operation_receipts(evidence_id,operation_id,tenant_id,source,body) VALUES($1,$2,$3,$4,$5)", r.EvidenceID, cmd.CommandID, cmd.TenantID, source, raw)
	if err != nil {
		return dispatch.Result{}, err
	}
	if state == string(StateSucceeded) || state == string(StateFailed) || state == string(StateCancelled) {
		if json.Unmarshal(result, &r) != nil {
			return dispatch.Result{}, errors.New("legacy result unavailable")
		}
		if err = tx.Commit(); err != nil {
			return dispatch.Result{}, err
		}
		return r, nil
	}
	next := string(StateUnknown)
	if pending {
		next = string(StateAcceptedExternal)
	}
	if r.Kind == dispatch.FactSucceeded {
		next = string(StateSucceeded)
	}
	if r.Kind == dispatch.FactFailed {
		next = string(StateFailed)
	}
	_, err = tx.ExecContext(ctx, `UPDATE operations SET state=$2,result=$3,evidence_id=$4,provider_request_id=COALESCE(NULLIF($5,''),provider_request_id),updated_at=clock_timestamp() WHERE operation_id=$1`, cmd.CommandID, next, raw, r.EvidenceID, r.ProviderRequestID)
	if err != nil {
		return dispatch.Result{}, err
	}
	if r.Kind != dispatch.FactUnknown {
		if _, err = tx.ExecContext(ctx, "DELETE FROM polling_schedule WHERE operation_id=$1", cmd.CommandID); err != nil {
			return dispatch.Result{}, err
		}
	}
	if pending && poll {
		if err = ScheduleAcceptedPollTx(ctx, tx, cmd); err != nil {
			return dispatch.Result{}, err
		}
	}
	fact := map[string]any{"protocol_id": cmd.ProtocolID, "tenant_id": cmd.TenantID, "application_id": cmd.ApplicationID, "cell_id": cmd.CellID, "step_id": cmd.StepID, "operation_id": cmd.CommandID, "attempt_id": attemptID, "provider_account_id": cmd.ProviderAccountID, "provider_request_id": r.ProviderRequestID, "kind": r.Kind, "economic_kind": economicKindForSource(source), "response_body": r.ResponseBody, "error_message": r.ErrorMessage, "evidence_id": r.EvidenceID, "occurred_at": time.Now().UTC(), "economic_snapshot": cmd.EconomicSnapshot}
	if err = outbox.Enqueue(ctx, tx, "operation", cmd.CommandID, "operation.observed", fact); err != nil {
		return dispatch.Result{}, err
	}
	if err = tx.Commit(); err != nil {
		return dispatch.Result{}, err
	}
	return r, nil
}

func economicKindForSource(source string) string {
	if source == "SUBMIT_ACCEPTED" {
		return "SUBMITTED"
	}
	return ""
}
