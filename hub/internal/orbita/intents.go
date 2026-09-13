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
	"ai-hub/hub/internal/platform/pg"
)

type Intent struct {
	Command dispatch.Command
	Owner   string
	Epoch   int64
	Expired bool
}

// RecoverOrphanedIntent torna novamente elegível uma intenção READY que ficou
// sem executor após a expiração de seu lease. A identidade do comando não é
// recriada: somente next_attempt_at/diagnóstico são atualizados, permitindo
// que o publisher normal faça a única entrega autorizada.
func (s *Store) RecoverOrphanedIntent(ctx context.Context, cell string, minAge time.Duration) (Intent, bool, error) {
	if cell == "" || minAge <= 0 {
		return Intent{}, false, errors.New("missing orphan recovery scope")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Intent{}, false, err
	}
	defer tx.Rollback()
	if err = pg.SetWorkerCellScope(ctx, tx, cell); err != nil {
		return Intent{}, false, err
	}
	var raw []byte
	var protocolID, tenantID string
	var result Intent
	err = tx.QueryRowContext(ctx, `
		SELECT i.command,i.epoch,p.protocol_id,p.tenant_id
		FROM command_intents i
		JOIN protocols p ON p.protocol_id=i.protocol_id AND p.tenant_id=i.tenant_id
		WHERE i.cell_id=$1 AND i.state='READY'
		  AND i.next_attempt_at<=clock_timestamp()
		  AND (i.lease_until IS NULL OR i.lease_until<=clock_timestamp())
		  AND i.created_at<=clock_timestamp()-($2 * interval '1 millisecond')
		  AND (p.client_deadline_at>clock_timestamp() OR i.continue_after_client_deadline)
		  AND (p.status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED') OR i.continue_after_client_deadline)
		ORDER BY i.next_attempt_at,i.command_id
		FOR UPDATE OF i SKIP LOCKED LIMIT 1`, cell, minAge.Milliseconds()).Scan(&raw, &result.Epoch, &protocolID, &tenantID)
	if errors.Is(err, sql.ErrNoRows) {
		return Intent{}, false, err
	}
	if err != nil {
		return Intent{}, false, err
	}
	if err = json.Unmarshal(raw, &result.Command); err != nil || result.Command.CommandID == "" || result.Command.ProtocolID != protocolID || result.Command.TenantID != tenantID || result.Command.CellID != cell {
		return Intent{}, false, errors.New("invalid orphan intent")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE command_intents SET next_attempt_at=clock_timestamp(),last_error='orphan_recovered',lease_owner=NULL,lease_until=NULL WHERE command_id=$1 AND state='READY'`, result.Command.CommandID); err != nil {
		return Intent{}, false, err
	}
	// protocol_audit não tem cell_id: worker_cell_scope não a alcança. Agora
	// que o tenant da linha reivindicada é conhecido, aplica também o escopo
	// de tenant (LOCAL, coexiste com o de célula já ativo) só para este
	// insert de auditoria.
	if err = pg.SetTenantScope(ctx, tx, tenantID); err != nil {
		return Intent{}, false, err
	}
	details, _ := json.Marshal(map[string]any{"command_id": result.Command.CommandID, "dispatch_mode": result.Command.DispatchMode, "previous_epoch": result.Epoch})
	if _, err = tx.ExecContext(ctx, `INSERT INTO protocol_audit(protocol_id,tenant_id,subject,action,details) VALUES($1,$2,'orphan-recovery','ORPHAN_DISPATCH_RECOVERED',$3)`, protocolID, tenantID, details); err != nil {
		return Intent{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return Intent{}, false, err
	}
	return result, true, nil
}

// RunOrphanRecovery mantém a recuperação de aceites órfãos separada do
// publisher: o scanner apenas registra o diagnóstico e reabre a mesma
// intenção; não há conversão de modo nem criação de novo comando.
func RunOrphanRecovery(ctx context.Context, s *Store, cell string, log *slog.Logger) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for n := 0; n < 16; n++ {
				_, recovered, err := s.RecoverOrphanedIntent(ctx, cell, 2*time.Second)
				if errors.Is(err, sql.ErrNoRows) || !recovered {
					break
				}
				if err != nil {
					log.Error("orphan intent recovery unavailable", "error", err)
					break
				}
			}
		}
	}
}

// ClaimIntent only takes QUEUED work. Direct recovery must query Cometa using
// the original command identity; it must never convert transport modes.
func (s *Store) ClaimIntent(ctx context.Context, cell, owner string) (Intent, error) {
	return s.claimIntentMode(ctx, cell, owner, dispatch.DispatchQueued)
}

// ClaimDirectIntent reivindica a intenção DIRECT criada para uma admissão
// SYNC. O request original e o recuperador usam a mesma autoridade de lease;
// assim, somente um deles pode atravessar o limite para o Cometa.
func (s *Store) ClaimDirectIntent(ctx context.Context, tenantID, commandID, owner string) (Intent, error) {
	if commandID == "" || owner == "" {
		return Intent{}, errors.New("missing direct intent identity")
	}
	var raw []byte
	var result Intent
	result.Owner = owner
	err := pg.WithTenantTx(ctx, s.db, tenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			UPDATE command_intents i
			SET lease_owner=$2,lease_until=clock_timestamp()+interval '15 seconds',epoch=i.epoch+1,attempts=i.attempts+1
			FROM protocols p
			WHERE i.command_id=$1 AND i.protocol_id=p.protocol_id
			  AND i.dispatch_mode=$3 AND i.state='READY'
			  AND i.next_attempt_at<=clock_timestamp()
			  AND (i.lease_until IS NULL OR i.lease_until<=clock_timestamp())
			  AND (p.client_deadline_at>clock_timestamp() OR i.continue_after_client_deadline)
			  AND (p.status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED') OR i.continue_after_client_deadline)
			RETURNING i.command,i.epoch,(i.retry_until IS NOT NULL AND i.retry_until<=clock_timestamp())`, commandID, owner, string(dispatch.DispatchDirect)).Scan(&raw, &result.Epoch, &result.Expired)
	})
	if err != nil {
		return Intent{}, err
	}
	if err = json.Unmarshal(raw, &result.Command); err != nil {
		return Intent{}, err
	}
	return result, nil
}

func (s *Store) claimIntentMode(ctx context.Context, cell, owner string, mode dispatch.DispatchMode) (Intent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Intent{}, err
	}
	defer tx.Rollback()
	if err = pg.SetWorkerCellScope(ctx, tx, cell); err != nil {
		return Intent{}, err
	}
	var raw []byte
	var protocolID string
	var result Intent
	result.Owner = owner
	// The intent row is serialized first. The plan row is locked and its
	// running_count is rechecked below before either the intent or the step is
	// published. This closes the read-count/update race between publishers.
	err = tx.QueryRowContext(ctx, `
		SELECT i.command,i.epoch,i.protocol_id,(i.retry_until IS NOT NULL AND i.retry_until<=clock_timestamp())
		FROM command_intents i
		JOIN protocols p ON p.protocol_id=i.protocol_id AND p.tenant_id=i.tenant_id
		WHERE i.cell_id=$1 AND i.dispatch_mode=$2 AND i.state='READY'
		  AND i.next_attempt_at<=clock_timestamp()
		  AND (i.lease_until IS NULL OR i.lease_until<=clock_timestamp())
		  AND (p.client_deadline_at>clock_timestamp() OR i.continue_after_client_deadline)
		  AND (p.status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED') OR i.continue_after_client_deadline)
		  AND (NOT EXISTS (SELECT 1 FROM operation_plans op WHERE op.protocol_id=i.protocol_id)
		       OR EXISTS (SELECT 1 FROM operation_steps os
		                  JOIN operation_plans op ON op.protocol_id=os.protocol_id
		                  WHERE os.protocol_id=i.protocol_id AND os.command_id=i.command_id
		                    AND ((os.state='READY' AND op.running_count<op.max_parallel)
		                         OR (os.state='RUNNING' AND os.lease_until<=clock_timestamp()))))
		ORDER BY (i.retry_until IS NOT NULL AND i.retry_until<=clock_timestamp()),i.next_attempt_at,i.command_id
		FOR UPDATE OF i SKIP LOCKED LIMIT 1`, cell, string(mode)).Scan(&raw, &result.Epoch, &protocolID, &result.Expired)
	if err != nil {
		return Intent{}, err
	}
	if err = json.Unmarshal(raw, &result.Command); err != nil {
		return Intent{}, err
	}
	var stepID, stepState string
	stepErr := tx.QueryRowContext(ctx, `SELECT step_id,state FROM operation_steps WHERE protocol_id=$1 AND command_id=$2 FOR UPDATE`, protocolID, result.Command.CommandID).Scan(&stepID, &stepState)
	product := stepErr == nil
	if stepErr != nil && !errors.Is(stepErr, sql.ErrNoRows) {
		return Intent{}, stepErr
	}
	if product && !result.Expired {
		var running, maxParallel int
		if err = tx.QueryRowContext(ctx, `SELECT running_count,max_parallel FROM operation_plans WHERE protocol_id=$1 FOR UPDATE`, protocolID).Scan(&running, &maxParallel); err != nil {
			return Intent{}, err
		}
		if stepState == "READY" {
			if running >= maxParallel {
				return Intent{}, sql.ErrNoRows
			}
			if _, err = tx.ExecContext(ctx, `UPDATE operation_plans SET running_count=running_count+1,updated_at=clock_timestamp() WHERE protocol_id=$1`, protocolID); err != nil {
				return Intent{}, err
			}
		} else if stepState != "RUNNING" {
			return Intent{}, sql.ErrNoRows
		}
		if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state='RUNNING',lease_owner=$2,lease_until=clock_timestamp()+interval '15 seconds',lease_epoch=lease_epoch+1,dispatch_started_at=COALESCE(dispatch_started_at,clock_timestamp()),version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND command_id=$3 AND (state='READY' OR (state='RUNNING' AND lease_until<=clock_timestamp()))`, protocolID, owner, result.Command.CommandID); err != nil {
			return Intent{}, err
		}
	}
	if err = tx.QueryRowContext(ctx, `UPDATE command_intents SET lease_owner=$2,lease_until=clock_timestamp()+interval '15 seconds',epoch=epoch+1,attempts=attempts+1 WHERE command_id=$1 RETURNING epoch`, result.Command.CommandID, owner).Scan(&result.Epoch); err != nil {
		return Intent{}, err
	}
	if err = tx.Commit(); err != nil {
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
			// R6-EXE-02: retry_until vencido barra nova submissão mesmo com o
			// SLA do cliente ainda futuro — o mesmo comando não pode ser
			// reenviado ao provedor sem saber se um efeito já ocorreu. A
			// reconciliação de um efeito possivelmente já realizado segue pelo
			// caminho de consulta autorizada (protocol_reconciliation_requests),
			// nunca por reenvio às cegas.
			if i.Expired {
				if _, err = s.CompleteIntent(ctx, i, false); err != nil {
					log.Error("expired direct intent finalization unavailable")
				}
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
			var commands []dispatch.Command
			err := pg.WithWorkerCellTx(ctx, s.db, cell, func(tx *sql.Tx) error {
				rows, queryErr := tx.QueryContext(ctx, `SELECT i.command FROM command_intents i JOIN protocols p USING(protocol_id) WHERE i.cell_id=$1 AND i.state='WAITING_RESERVATION' AND p.client_deadline_at>clock_timestamp() AND p.status IN ('ACCEPTED','RUNNING') ORDER BY i.created_at LIMIT 5`, cell)
				if queryErr != nil {
					return queryErr
				}
				defer rows.Close()
				for rows.Next() {
					var raw []byte
					var cmd dispatch.Command
					if rows.Scan(&raw) == nil && json.Unmarshal(raw, &cmd) == nil {
						commands = append(commands, cmd)
					}
				}
				return rows.Err()
			})
			if err != nil {
				continue
			}
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
				err = pg.WithTenantTx(ctx, s.db, cmd.TenantID, func(tx *sql.Tx) error {
					_, execErr := tx.ExecContext(ctx, "UPDATE command_intents SET state='READY' WHERE command_id=$1 AND state='WAITING_RESERVATION'", cmd.CommandID)
					return execErr
				})
				if err != nil {
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
	var n int64
	err := pg.WithTenantTx(ctx, s.db, i.Command.TenantID, func(tx *sql.Tx) error {
		res, execErr := tx.ExecContext(ctx, `
			UPDATE command_intents
			SET state = CASE
				WHEN retry_until IS NOT NULL AND retry_until <= clock_timestamp() THEN 'EXPIRED'
				WHEN $4 = 'DELIVERED' THEN 'DELIVERED'
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
					WHEN $4 = 'DELIVERED' AND (retry_until IS NULL OR retry_until>clock_timestamp()) THEN clock_timestamp()
					WHEN retry_until IS NOT NULL AND retry_until <= clock_timestamp() THEN clock_timestamp()
					WHEN $6 <= 0 THEN clock_timestamp()
					ELSE clock_timestamp()+make_interval(secs=>LEAST(2,$6))
				END
			WHERE command_id=$1 AND lease_owner=$2 AND epoch=$3 AND lease_until>clock_timestamp() AND state='READY'`, i.Command.CommandID, i.Owner, i.Epoch, state, code, i.Command.RetryTTLSeconds)
		if execErr != nil {
			return execErr
		}
		n, execErr = res.RowsAffected()
		if execErr != nil {
			return execErr
		}
		if n == 1 && !delivered {
			_, execErr = tx.ExecContext(ctx, `WITH released AS (
				UPDATE operation_steps SET state='READY',lease_owner=NULL,lease_until=NULL,version=version+1,updated_at=clock_timestamp()
				WHERE command_id=$1 AND state='RUNNING' RETURNING protocol_id
			) UPDATE operation_plans p SET running_count=GREATEST(0,p.running_count-1),updated_at=clock_timestamp()
			FROM released r WHERE p.protocol_id=r.protocol_id`, i.Command.CommandID)
		}
		return execErr
	})
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
				if i.Expired {
					if _, err = s.CompleteIntent(ctx, i, false); err != nil {
						log.Error("expired intent finalization unavailable")
					}
					continue
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
