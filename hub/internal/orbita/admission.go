package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/objectstore"
	"ai-hub/hub/internal/platform/auth"
)

// Admit serializes a semantic idempotency key and commits the identity,
// snapshot and replayable obligation together. The caller must treat a commit
// error as uncertain and retry this same key; it must never invent an acceptance.
func (s *Store) Admit(ctx context.Context, p Protocol, cmd dispatch.Command, subject string, reservation bool) (Protocol, bool, error) {
	if p.ApplicationID == "" || p.CellID == "" || subject == "" || len(cmd.ConfigSnapshot) == 0 || cmd.CommandID != p.CommandID || cmd.TenantID != p.TenantID || cmd.ProtocolID != p.ProtocolID || cmd.ApplicationID != p.ApplicationID || cmd.CellID != p.CellID {
		return Protocol{}, false, errors.New("invalid admission identity or snapshot")
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		return Protocol{}, false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Protocol{}, false, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", p.TenantID+"\x1f"+p.ApplicationID+"\x1f"+p.IdempotencyKey); err != nil {
		return Protocol{}, false, err
	}
	var id, hash string
	err = tx.QueryRowContext(ctx, "SELECT protocol_id,request_hash FROM protocols WHERE tenant_id=$1 AND application_id=$2 AND idempotency_key=$3", p.TenantID, p.ApplicationID, p.IdempotencyKey).Scan(&id, &hash)
	if err == nil {
		if hash != p.RequestHash {
			return Protocol{}, false, ErrIdempotencyConflict
		}
		tx.Rollback()
		existing, e := s.Get(ctx, p.TenantID, id)
		return existing, false, e
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Protocol{}, false, err
	}
	var databaseNow time.Time
	if err = tx.QueryRowContext(ctx, "SELECT clock_timestamp()").Scan(&databaseNow); err != nil {
		return Protocol{}, false, err
	}
	if !databaseNow.Before(p.ClientDeadlineAt) {
		return Protocol{}, false, errors.New("admission deadline elapsed")
	}
	if len(cmd.FileRefs) > 10 {
		return Protocol{}, false, errors.New("too many file references")
	}
	for _, ref := range cmd.FileRefs {
		if err = objectstore.PinReferenceTx(ctx, tx, p.TenantID, ref, p.ProtocolID); err != nil {
			return Protocol{}, false, err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,admitted_subject,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,accepted_at,client_deadline_at,config_snapshot) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'ACCEPTED',$12,$13,$14)`, p.ProtocolID, p.TenantID, p.ApplicationID, p.CellID, subject, p.IdempotencyKey, p.RequestHash, []byte(p.RequestBody), p.Mode, p.DispatchMode, p.CommandID, p.AcceptedAt, p.ClientDeadlineAt, []byte(cmd.ConfigSnapshot))
	if err != nil {
		return Protocol{}, false, err
	}
	state := "READY"
	if reservation {
		state = "WAITING_RESERVATION"
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO command_intents(command_id,protocol_id,tenant_id,application_id,cell_id,dispatch_mode,command,state) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, p.CommandID, p.ProtocolID, p.TenantID, p.ApplicationID, p.CellID, p.DispatchMode, raw, state)
	if err != nil {
		return Protocol{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return Protocol{}, false, err
	}
	return p, true, nil
}

func applicationFromContext(ctx context.Context) string {
	p, ok := auth.FromContext(ctx)
	if ok && p.ApplicationID != "" {
		return p.ApplicationID
	}
	return "legacy-unverified"
}
