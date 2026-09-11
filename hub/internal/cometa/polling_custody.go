package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/outbox"
	"ai-hub/hub/internal/platform/idgen"
)

var ErrPollFence = errors.New("poll lease no longer authoritative")

type PollPolicy struct {
	IntervalSeconds    int `json:"interval_seconds"`
	MaxIntervalSeconds int `json:"max_interval_seconds"`
	TimeoutSeconds     int `json:"timeout_seconds"`
	JitterPercent      int `json:"jitter_percent"`
	MaxAttempts        int `json:"max_attempts"`
}

func ScheduleAcceptedPollTx(ctx context.Context, tx *sql.Tx, cmd dispatch.Command) error {
	p := PollPolicy{2, 60, 5, 10, 100}
	var snapshot atlas.OfferSnapshot
	if json.Unmarshal(cmd.ConfigSnapshot, &snapshot) != nil {
		return errors.New("invalid polling snapshot")
	}
	var config struct {
		Polling *PollPolicy `json:"polling"`
	}
	if len(snapshot.Account.Data) > 0 && json.Unmarshal(snapshot.Account.Data, &config) != nil {
		return errors.New("invalid polling account")
	}
	if config.Polling != nil {
		p = *config.Polling
	} else if os.Getenv("ENVIRONMENT") != "local" {
		return errors.New("published polling policy required")
	}
	if p.IntervalSeconds < 1 || p.MaxIntervalSeconds < p.IntervalSeconds || p.MaxIntervalSeconds > 3600 || p.TimeoutSeconds < 1 || p.TimeoutSeconds > 60 || p.JitterPercent < 0 || p.JitterPercent > 50 || p.MaxAttempts < 1 || p.MaxAttempts > 100000 {
		return errors.New("invalid polling policy")
	}
	// Uma aceitação externa já criou um efeito e o polling passa a observar
	// sua conclusão normal dentro do prazo do passo/cliente. O retry TTL
	// governa recuperação de falhas transitórias de transporte; usá-lo aqui
	// encerraria polling saudável antes do SLA e permitiria que o timer da
	// Órbita expirasse uma operação que o provedor ainda podia concluir.
	deadline := cmd.StepDeadline
	if deadline.IsZero() {
		deadline = cmd.RetryDeadline
	}
	if deadline.IsZero() {
		return errors.New("missing absolute polling deadline")
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO polling_schedule(operation_id,next_run_at,interval_seconds,max_interval_seconds,deadline_at,timeout_seconds,jitter_percent,max_attempts)
 VALUES($1,clock_timestamp()+make_interval(secs=>$2),$2,$3,$4,$5,$6,$7) ON CONFLICT(operation_id) DO NOTHING`, cmd.CommandID, p.IntervalSeconds, p.MaxIntervalSeconds, deadline, p.TimeoutSeconds, p.JitterPercent, p.MaxAttempts)
	return err
}

type PollClaim struct {
	Command                             dispatch.Command
	ProviderRequestID, Owner, AttemptID string
	Epoch                               int64
	Deadline                            time.Time
	TimeoutSeconds                      int
}

// ClaimPoll commits STATUS preparation before any network operation.
func (s *Store) ClaimPoll(ctx context.Context, cell, owner string) (PollClaim, error) {
	var c PollClaim
	if cell == "" || owner == "" {
		return c, errors.New("missing poll identity")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return c, err
	}
	defer tx.Rollback()
	var id string
	err = tx.QueryRowContext(ctx, `SELECT o.operation_id FROM operations o JOIN polling_schedule p USING(operation_id)
 WHERE o.cell_id=$1 AND o.state IN ('ACCEPTED_EXTERNAL','WAITING_FINAL','UNKNOWN') AND o.provider_request_id IS NOT NULL
 AND p.next_run_at<=clock_timestamp() AND p.deadline_at>clock_timestamp()+make_interval(secs=>p.timeout_seconds)
 AND p.attempts_count<p.max_attempts AND (p.lease_expires_at IS NULL OR p.lease_expires_at<clock_timestamp())
 ORDER BY p.next_run_at FOR UPDATE OF o SKIP LOCKED LIMIT 1`, cell).Scan(&id)
	if err != nil {
		return c, err
	}
	var raw []byte
	// The lease covers durable preparation plus binding/token resolution
	// before the final fence check. A timeout+5s lease was too narrow when
	// PostgreSQL or the control-plane binding call was slow, causing a valid
	// poll to fence itself before provider I/O. Keep the lease bounded while
	// preserving the absolute deadline predicate below.
	err = tx.QueryRowContext(ctx, `UPDATE polling_schedule p SET lease_owner=$2,epoch=epoch+1,lease_expires_at=clock_timestamp()+make_interval(secs=>GREATEST(timeout_seconds+5,30)),attempts_count=attempts_count+1
 FROM operations o WHERE p.operation_id=$1 AND o.operation_id=p.operation_id
 RETURNING p.epoch,p.deadline_at,p.timeout_seconds,o.command,o.provider_request_id`, id, owner).Scan(&c.Epoch, &c.Deadline, &c.TimeoutSeconds, &raw, &c.ProviderRequestID)
	if err != nil {
		return c, err
	}
	if json.Unmarshal(raw, &c.Command) != nil || c.Command.CommandID != id || c.Command.CellID != cell {
		return c, errors.New("invalid polling command")
	}
	c.Owner = owner
	c.AttemptID = idgen.New()
	_, err = tx.ExecContext(ctx, `INSERT INTO attempts(attempt_id,operation_id,attempt_type,prepared_at,owner,epoch) VALUES($1,$2,'STATUS',clock_timestamp(),$3,$4)`, c.AttemptID, id, owner, c.Epoch)
	if err != nil {
		return c, err
	}
	return c, tx.Commit()
}

// CompletePoll conserves every receipt. A stale observer cannot alter the winner
// or its schedule. Lock order matches callback custody: operation then schedule.
func (s *Store) CompletePoll(ctx context.Context, c PollClaim, r dispatch.Result, retryAfter time.Duration) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var state, tenant, cell string
	var existingRaw []byte
	err = tx.QueryRowContext(ctx, `SELECT state,tenant_id,cell_id,result FROM operations WHERE operation_id=$1 FOR UPDATE`, c.Command.CommandID).Scan(&state, &tenant, &cell, &existingRaw)
	if err != nil {
		return err
	}
	if tenant != c.Command.TenantID || cell != c.Command.CellID {
		return errors.New("poll outside scope")
	}
	var valid bool
	err = tx.QueryRowContext(ctx, `SELECT lease_owner=$2 AND epoch=$3 AND lease_expires_at>clock_timestamp() FROM polling_schedule WHERE operation_id=$1 FOR UPDATE`, c.Command.CommandID, c.Owner, c.Epoch).Scan(&valid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	terminal := state == string(StateSucceeded) || state == string(StateFailed) || state == string(StateCancelled)
	source := "POLL"
	if !valid {
		source = "POLL_STALE"
	}
	if terminal {
		source = "POLL_AFTER_FINAL"
		var existing dispatch.Result
		if json.Unmarshal(existingRaw, &existing) == nil &&
			(existing.Kind != r.Kind || (existing.ProviderRequestID != "" && r.ProviderRequestID != "" && existing.ProviderRequestID != r.ProviderRequestID)) {
			source += "_CONFLICT"
		}
	}
	r.CommandID = c.Command.CommandID
	r.OperationID = c.Command.CommandID
	r.EvidenceID = idgen.New()
	r.Durable = true
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO operation_receipts(evidence_id,operation_id,tenant_id,source,body) VALUES($1,$2,$3,$4,$5)`, r.EvidenceID, c.Command.CommandID, tenant, source, raw)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE attempts SET received_at=clock_timestamp(),error_code=NULLIF($2,'') WHERE attempt_id=$1 AND operation_id=$3 AND owner=$4 AND epoch=$5`, c.AttemptID, r.ErrorCode, c.Command.CommandID, c.Owner, c.Epoch)
	if err != nil {
		return err
	}
	// A fenced nonterminal observer must not publish an operational or
	// financial fact. If the race already reached a terminal state, the HTTP
	// STATUS call itself remains auditable and billable even though it cannot
	// change the winner.
	if !valid && !terminal {
		if err = tx.Commit(); err != nil {
			return err
		}
		return ErrPollFence
	}
	// The provider call itself is a billable STATUS attempt, even when a
	// callback or a newer lease wins the terminal state race. The operational
	// kind remains the observed result so Orbita can consolidate a terminal
	// poll; Libra uses economic_kind=STATUS and the immutable attempt_id.
	fact := map[string]any{"protocol_id": c.Command.ProtocolID, "tenant_id": tenant, "application_id": c.Command.ApplicationID, "cell_id": cell, "step_id": c.Command.StepID, "operation_id": c.Command.CommandID, "attempt_id": c.AttemptID, "provider_request_id": r.ProviderRequestID, "kind": r.Kind, "economic_kind": "STATUS", "response_body": r.ResponseBody, "error_message": r.ErrorMessage, "evidence_id": r.EvidenceID, "occurred_at": time.Now().UTC(), "economic_snapshot": c.Command.EconomicSnapshot}
	if err = outbox.Enqueue(ctx, tx, "operation", c.Command.CommandID, "operation.observed", fact); err != nil {
		return err
	}
	if terminal {
		if err = tx.Commit(); err != nil {
			return err
		}
		return nil
	}
	if r.Kind == dispatch.FactSucceeded || r.Kind == dispatch.FactFailed {
		_, err = tx.ExecContext(ctx, `UPDATE operations SET state=$2,result=$3,evidence_id=$4,updated_at=clock_timestamp() WHERE operation_id=$1`, c.Command.CommandID, string(r.Kind), raw, r.EvidenceID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `DELETE FROM polling_schedule WHERE operation_id=$1`, c.Command.CommandID)
		if err != nil {
			return err
		}
	} else {
		// Delay honors Retry-After even when it puts the next attempt beyond deadline;
		// the claim predicate then prevents further I/O without losing the obligation.
		// The provider's Retry-After is a lower bound observed before this
		// transaction commits. Add a small durable-commit cushion when it is
		// present; otherwise WAL/commit latency could make the persisted
		// next_run_at earlier than the provider allowed.
		retryAfterSeconds := retryAfter.Seconds()
		if retryAfterSeconds > 0 {
			// Five seconds is the local durability cushion for the worst
			// observed commit latency in the qualification database; it is
			// conservative and only delays, never accelerates, a retry.
			retryAfterSeconds += 5
		}
		_, err = tx.ExecContext(ctx, `UPDATE polling_schedule SET next_run_at=clock_timestamp()+make_interval(secs=>GREATEST($2::double precision,LEAST(max_interval_seconds,interval_seconds*2)*(1+random()*jitter_percent/100.0))),interval_seconds=LEAST(max_interval_seconds,interval_seconds*2),lease_owner=NULL,lease_expires_at=NULL WHERE operation_id=$1`, c.Command.CommandID, retryAfterSeconds)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
