package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/contracts/economics"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/libraclient"
	"ai-hub/hub/internal/platform/idgen"
)

type Intent struct {
	Command dispatch.Command
	Owner   string
	Epoch   int64
}

// ClaimIntent only takes QUEUED work. Direct recovery must query Cometa using
// the original command identity; it must never convert transport modes.
func (s *Store) ClaimIntent(ctx context.Context, cell, owner string) (Intent, error) {
	return s.claimIntentMode(ctx, cell, owner, dispatch.DispatchQueued)
}

// ClaimDirectIntent reivindica a intenção DIRECT criada para uma admissão
// SYNC. O request original e o recuperador usam a mesma autoridade de lease;
// assim, somente um deles pode atravessar o limite para o Cometa.
func (s *Store) ClaimDirectIntent(ctx context.Context, commandID, owner string) (Intent, error) {
	if commandID == "" || owner == "" {
		return Intent{}, errors.New("missing direct intent identity")
	}
	var raw []byte
	var result Intent
	result.Owner = owner
	err := s.db.QueryRowContext(ctx, `
		UPDATE command_intents i
		SET lease_owner=$2,lease_until=clock_timestamp()+interval '15 seconds',epoch=i.epoch+1,attempts=i.attempts+1
		FROM protocols p
		WHERE i.command_id=$1 AND i.protocol_id=p.protocol_id
		  AND i.dispatch_mode=$3 AND i.state='READY'
		  AND i.next_attempt_at<=clock_timestamp()
		  AND (i.lease_until IS NULL OR i.lease_until<=clock_timestamp())
		  AND p.client_deadline_at>clock_timestamp()
		  AND p.status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED')
		RETURNING i.command,i.epoch`, commandID, owner, string(dispatch.DispatchDirect)).Scan(&raw, &result.Epoch)
	if err != nil {
		return Intent{}, err
	}
	if err = json.Unmarshal(raw, &result.Command); err != nil {
		return Intent{}, err
	}
	return result, nil
}

func (s *Store) claimIntentMode(ctx context.Context, cell, owner string, mode dispatch.DispatchMode) (Intent, error) {
	var raw []byte
	var result Intent
	result.Owner = owner
	err := s.db.QueryRowContext(ctx, `WITH candidate AS (
 SELECT i.command_id FROM command_intents i JOIN protocols p USING(protocol_id)
 WHERE i.cell_id=$1 AND i.dispatch_mode=$3 AND i.state='READY'
 AND i.next_attempt_at<=clock_timestamp() AND (i.lease_until IS NULL OR i.lease_until<=clock_timestamp())
 AND p.client_deadline_at>clock_timestamp() AND p.status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED')
 ORDER BY i.next_attempt_at,i.command_id FOR UPDATE OF i SKIP LOCKED LIMIT 1)
 UPDATE command_intents i SET lease_owner=$2,lease_until=clock_timestamp()+interval '15 seconds',epoch=i.epoch+1,attempts=i.attempts+1
 FROM candidate c WHERE i.command_id=c.command_id RETURNING i.command,i.epoch`, cell, owner, string(mode)).Scan(&raw, &result.Epoch)
	if err != nil {
		return Intent{}, err
	}
	if err = json.Unmarshal(raw, &result.Command); err != nil {
		return Intent{}, err
	}
	return result, nil
}

// Direct recovery reuses the original operation identity over authenticated HTTP.
// It never publishes DIRECT to the command queue, and UNKNOWN remains pending.
func RunDirectRecovery(ctx context.Context, s *Store, d *Dispatcher, f *Finalizer, cell string, log *slog.Logger) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	owner := idgen.New()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			i, err := s.claimIntentMode(ctx, cell, owner, dispatch.DispatchDirect)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				log.Error("direct recovery claim unavailable")
				continue
			}
			call, cancel := context.WithTimeout(ctx, 5*time.Second)
			result, err := d.DispatchDirect(call, i.Command)
			cancel()
			completed := false
			if err == nil && result.Durable && (result.Kind == dispatch.FactSucceeded || result.Kind == dispatch.FactFailed) {
				p, e := s.Get(ctx, i.Command.TenantID, i.Command.ProtocolID)
				if e == nil {
					if IsTerminal(p.Status) {
						completed = true
					} else {
						status := StatusFailed
						if result.Kind == dispatch.FactSucceeded {
							status = StatusSucceeded
						}
						_, e = f.Finalize(ctx, i.Command.TraceID, p.TenantID, p.ProtocolID, p.Version, status, FinalBody{ProtocolID: p.ProtocolID, Result: result.ResponseBody, ErrorCode: result.ErrorCode, ErrorMessage: result.ErrorMessage}, "")
						completed = e == nil
					}
				}
			}
			if _, err = s.CompleteIntent(ctx, i, completed); err != nil {
				log.Error("direct recovery confirmation unavailable")
			}
		}
	}
}

func RunReservationRecovery(ctx context.Context, s *Store, client *libraclient.Client, cell string, log *slog.Logger) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, err := s.db.QueryContext(ctx, `SELECT i.command FROM command_intents i JOIN protocols p USING(protocol_id) WHERE i.cell_id=$1 AND i.state='WAITING_RESERVATION' AND p.client_deadline_at>clock_timestamp() AND p.status IN ('ACCEPTED','RUNNING') ORDER BY i.created_at LIMIT 5`, cell)
			if err != nil {
				continue
			}
			var commands []dispatch.Command
			for rows.Next() {
				var raw []byte
				var cmd dispatch.Command
				if rows.Scan(&raw) == nil && json.Unmarshal(raw, &cmd) == nil {
					commands = append(commands, cmd)
				}
			}
			rows.Close()
			for _, cmd := range commands {
				var snapshot atlas.OfferSnapshot
				if json.Unmarshal(cmd.ConfigSnapshot, &snapshot) != nil {
					continue
				}
				sale, _, err := economics.DecodePublished(snapshot.SaleContract.ID, snapshot.SaleContract.Version, snapshot.SaleContract.Data)
				if err != nil || !sale.StrictBalance {
					continue
				}
				call, cancel := context.WithTimeout(ctx, 3*time.Second)
				err = client.ReserveExact(call, cmd.TenantID, cmd.ProtocolID, string(sale.ReservationAmount), sale.Currency)
				cancel()
				if err != nil {
					continue
				}
				if _, err = s.db.ExecContext(ctx, "UPDATE command_intents SET state='READY' WHERE command_id=$1 AND state='WAITING_RESERVATION'", cmd.CommandID); err != nil {
					log.Error("reservation confirmation unavailable")
				}
			}
		}
	}
}

func (s *Store) CompleteIntent(ctx context.Context, i Intent, delivered bool) (bool, error) {
	state := "READY"
	code := "transport_unconfirmed"
	if delivered {
		state = "DELIVERED"
		code = ""
	}
	// The retry horizon starts at the first transient transport failure and
	// remains immutable across attempts and process restarts. Once exhausted,
	// the intent stops publishing and remains available to reconciliation.
	res, err := s.db.ExecContext(ctx, `
		UPDATE command_intents
		SET state = CASE
			WHEN $4 = 'DELIVERED' THEN 'DELIVERED'
			WHEN retry_until IS NOT NULL AND retry_until <= clock_timestamp() THEN 'EXPIRED'
			WHEN $6 <= 0 THEN 'EXPIRED'
			ELSE 'READY'
		END,
			retry_started_at = CASE
				WHEN $4 = 'DELIVERED' THEN retry_started_at
				WHEN retry_started_at IS NULL THEN clock_timestamp()
				ELSE retry_started_at
			END,
			retry_until = CASE
				WHEN $4 = 'DELIVERED' THEN retry_until
				WHEN retry_until IS NULL AND $6 > 0 THEN clock_timestamp() + make_interval(secs=>$6)
				ELSE retry_until
			END,
			last_error=NULLIF($5,''), lease_owner=NULL, lease_until=NULL,
			next_attempt_at = CASE
				WHEN $4 = 'DELIVERED' THEN clock_timestamp()
				WHEN retry_until IS NOT NULL AND retry_until <= clock_timestamp() THEN clock_timestamp()
				WHEN $6 <= 0 THEN clock_timestamp()
				ELSE clock_timestamp()+make_interval(secs=>LEAST(2,$6))
			END
		WHERE command_id=$1 AND lease_owner=$2 AND epoch=$3 AND lease_until>clock_timestamp() AND state='READY'`, i.Command.CommandID, i.Owner, i.Epoch, state, code, i.Command.RetryTTLSeconds)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func RunIntentPublisher(ctx context.Context, s *Store, d *Dispatcher, cell string, log *slog.Logger) {
	owner := idgen.New()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for n := 0; n < 16; n++ {
				i, err := s.ClaimIntent(ctx, cell, owner)
				if errors.Is(err, sql.ErrNoRows) {
					break
				}
				if err != nil {
					log.Error("intent claim unavailable")
					break
				}
				callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err = d.DispatchQueued(callCtx, i.Command)
				cancel()
				if _, e := s.CompleteIntent(ctx, i, err == nil); e != nil {
					log.Error("intent confirmation unavailable")
				}
				if err != nil {
					break
				}
			}
		}
	}
}
