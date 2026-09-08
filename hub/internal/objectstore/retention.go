package objectstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
)

type PurgePlan struct {
	ID        string    `json:"id"`
	FileIDs   []string  `json:"file_ids"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Catalog) DryRun(ctx context.Context, tenant, actor string) (PurgePlan, error) {
	out := PurgePlan{ID: uuid.NewString(), FileIDs: []string{}, ExpiresAt: time.Now().UTC().Add(15 * time.Minute)}
	if actor == "" || tenant == "" {
		return out, ErrIneligible
	}
	rows, e := c.db.QueryContext(ctx, `SELECT id FROM file_refs f WHERE tenant_id=$1 AND state IN ('READY','ORPHAN','PURGING') AND retention_until<clock_timestamp() AND NOT EXISTS(SELECT 1 FROM object_retention_pins p WHERE p.file_id=f.id) ORDER BY id LIMIT 100`, tenant)
	if e != nil {
		return out, e
	}
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return out, e
		}
		out.FileIDs = append(out.FileIDs, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return out, e
	}
	b, _ := json.Marshal(out.FileIDs)
	_, e = c.db.ExecContext(ctx, `INSERT INTO object_gc_runs(id,tenant_id,actor,candidates,expires_at) VALUES($1,$2,$3,$4,$5)`, out.ID, tenant, actor, string(b), out.ExpiresAt)
	return out, e
}
func (c *Catalog) Purge(ctx context.Context, tenant, planID, id, actor string) error {
	if actor == "" {
		return ErrIneligible
	}
	tx, e := c.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var candidates []byte
	var expires time.Time
	e = tx.QueryRowContext(ctx, `SELECT candidates,expires_at FROM object_gc_runs WHERE id=$1 AND tenant_id=$2`, planID, tenant).Scan(&candidates, &expires)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrNotFound
	}
	if e != nil {
		return e
	}
	var ids []string
	if json.Unmarshal(candidates, &ids) != nil || !slices.Contains(ids, id) || time.Now().After(expires) {
		return ErrIneligible
	}
	var key, version, sha, state string
	var until time.Time
	e = tx.QueryRowContext(ctx, `SELECT object_key,COALESCE(object_version,''),sha256,state,retention_until FROM file_refs WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenant, id).Scan(&key, &version, &sha, &state, &until)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrNotFound
	}
	if e != nil {
		return e
	}
	if state == "PURGED" {
		return tx.Commit()
	}
	if (state != "READY" && state != "ORPHAN" && state != "PURGING") || time.Now().Before(until) {
		return ErrIneligible
	}
	var pinned bool
	if e = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM object_retention_pins WHERE file_id=$1)`, id).Scan(&pinned); e != nil {
		return e
	}
	if pinned {
		return ErrIneligible
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO object_tombstones(file_id,tenant_id,sha256,reason) VALUES($1,$2,$3,'RETENTION_POLICY') ON CONFLICT DO NOTHING`, id, tenant, sha)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `UPDATE file_refs SET state='PURGING' WHERE tenant_id=$1 AND id=$2`, tenant, id)
	if e != nil {
		return e
	}
	if e = tx.Commit(); e != nil {
		return e
	}
	if e = c.client.DeleteVersion(ctx, key, version); e != nil {
		return e
	}
	tx, e = c.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, `UPDATE file_refs SET state='PURGED',updated_at=clock_timestamp() WHERE tenant_id=$1 AND id=$2 AND state='PURGING'`, tenant, id)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO object_actions(tenant_id,file_id,action,actor) VALUES($1,$2,'PURGE',$3)`, tenant, id, actor)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (c *Catalog) IsTombstoned(ctx context.Context, tenant, id string) (bool, error) {
	var found bool
	e := c.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM object_tombstones WHERE tenant_id=$1 AND file_id=$2)`, tenant, id).Scan(&found)
	return found, e
}
