package cometa

import (
	"context"
	"crypto/sha256"
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
}

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
	claim := Submission{Command: cmd, Owner: idgen.New(), AttemptID: idgen.New(), Epoch: 1}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Submission{}, false, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `INSERT INTO operations(operation_id,protocol_id,provider_account_id,credential_binding_id,secret_version_id,tenant_id,application_id,cell_id,command,command_hash,state,submit_owner,submit_epoch,submit_lease_until) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'SUBMITTING',$11,1,clock_timestamp()+interval '30 seconds') ON CONFLICT(operation_id) DO NOTHING`, cmd.CommandID, cmd.ProtocolID, cmd.ProviderAccountID, binding, secretVersion, cmd.TenantID, cmd.ApplicationID, cmd.CellID, raw, hash, claim.Owner)
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
func (s *Store) ConserveObservation(ctx context.Context, cmd dispatch.Command, r dispatch.Result, source string) (dispatch.Result, error) {
	return s.conserveObservation(ctx, cmd, r, source, false, false)
}

// ConserveAcceptance keeps provider correlation, receipt, result, polling
// obligation and outbox in one commit. No caller can acknowledge a partial save.
func (s *Store) ConserveAcceptance(ctx context.Context, cmd dispatch.Command, providerRequestID string, poll bool) (dispatch.Result, error) {
	if providerRequestID == "" {
		return dispatch.Result{}, errors.New("missing provider correlation")
	}
	return s.conserveObservation(ctx, cmd, dispatch.Result{Kind: dispatch.FactUnknown, ProviderRequestID: providerRequestID}, "SUBMIT_ACCEPTED", true, poll)
}

func (s *Store) conserveObservation(ctx context.Context, cmd dispatch.Command, r dispatch.Result, source string, pending, poll bool) (dispatch.Result, error) {
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
	fact := map[string]any{"protocol_id": cmd.ProtocolID, "tenant_id": cmd.TenantID, "application_id": cmd.ApplicationID, "cell_id": cmd.CellID, "step_id": cmd.StepID, "operation_id": cmd.CommandID, "provider_account_id": cmd.ProviderAccountID, "provider_request_id": r.ProviderRequestID, "kind": r.Kind, "response_body": r.ResponseBody, "error_message": r.ErrorMessage, "evidence_id": r.EvidenceID, "occurred_at": time.Now().UTC(), "economic_snapshot": cmd.EconomicSnapshot}
	if err = outbox.Enqueue(ctx, tx, "operation", cmd.CommandID, "operation.observed", fact); err != nil {
		return dispatch.Result{}, err
	}
	if err = tx.Commit(); err != nil {
		return dispatch.Result{}, err
	}
	return r, nil
}
