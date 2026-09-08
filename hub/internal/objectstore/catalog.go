package objectstore

import (
	"ai-hub/hub/internal/contracts/files"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("objectstore: file not found")
var ErrIneligible = errors.New("objectstore: file is not eligible")
var ErrPolicy = errors.New("objectstore: approved data policy required")

type Catalog struct {
	db     *sql.DB
	client *Client
}

func NewCatalog(db *sql.DB, client *Client) *Catalog { return &Catalog{db: db, client: client} }

type FileRef = files.Reference

// PinReferenceTx participates in Órbita's acceptance transaction. Failure to
// acquire the exact immutable version rolls back the protocol and its intent.
func PinReferenceTx(ctx context.Context, tx *sql.Tx, tenant string, ref FileRef, obligation string) error {
	var state, version, hash string
	err := tx.QueryRowContext(ctx, "SELECT state,COALESCE(object_version,''),sha256 FROM file_refs WHERE id=$1 AND tenant_id=$2 FOR UPDATE", ref.ID, tenant).Scan(&state, &version, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if state != "READY" || ref.Version != version || ref.SHA256 != hash || obligation == "" {
		return ErrIneligible
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO object_retention_pins(file_id,tenant_id,obligation_id,reason) VALUES($1,$2,$3,'PROTOCOL_INPUT') ON CONFLICT DO NOTHING", ref.ID, tenant, obligation)
	return err
}

type UploadRequest struct {
	Size        int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	ContentType string `json:"content_type"`
	Purpose     string `json:"purpose"`
	Class       string `json:"class"`
	Region      string `json:"region"`
}
type UploadSession struct {
	FileRef   FileRef     `json:"file_ref"`
	ExpiresAt time.Time   `json:"expires_at"`
	Uploads   []UploadURL `json:"uploads"`
}

func (c *Catalog) policy(ctx context.Context, q UploadRequest) (time.Duration, error) {
	var seconds, maxSize int64
	var types []byte
	e := c.db.QueryRowContext(ctx, `SELECT retention_seconds,max_bytes,allowed_types FROM object_retention_policies WHERE class=$1 AND region=$2 AND purpose=$3 AND approved_by<>''`, q.Class, q.Region, q.Purpose).Scan(&seconds, &maxSize, &types)
	if errors.Is(e, sql.ErrNoRows) {
		return 0, ErrPolicy
	}
	if e != nil {
		return 0, e
	}
	var allowed []string
	if json.Unmarshal(types, &allowed) != nil || !slices.Contains(allowed, q.ContentType) || q.Size < 1 || q.Size > maxSize {
		return 0, ErrPolicy
	}
	return time.Duration(seconds) * time.Second, nil
}
func (c *Catalog) CreateSession(ctx context.Context, tenant string, q UploadRequest) (UploadSession, error) {
	var out UploadSession
	if tenant == "" {
		return out, ErrNotFound
	}
	sum, e := hex.DecodeString(q.SHA256)
	if e != nil || len(sum) != 32 {
		return out, ErrIntegrity
	}
	retention, e := c.policy(ctx, q)
	if e != nil {
		return out, e
	}
	id := uuid.NewString()
	key := "custody/" + id
	expires := time.Now().UTC().Add(15 * time.Minute)
	_, e = c.db.ExecContext(ctx, `INSERT INTO file_refs(id,tenant_id,object_key,sha256,size_bytes,content_type,purpose,class,region,state,expires_at,retention_until) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'UPLOADING',$10,$11)`, id, tenant, key, q.SHA256, q.Size, q.ContentType, q.Purpose, q.Class, q.Region, expires, time.Now().UTC().Add(retention))
	if e != nil {
		return out, e
	}
	uploadID, urls, e := c.client.UploadURLs(ctx, key, q.ContentType, q.SHA256, q.Size, 15*time.Minute)
	if e != nil {
		return out, e
	}
	if uploadID != "" {
		if _, e = c.db.ExecContext(ctx, `UPDATE file_refs SET upload_id=$3 WHERE tenant_id=$1 AND id=$2`, tenant, id, uploadID); e != nil {
			return out, e
		}
	}
	return UploadSession{FileRef: FileRef{ID: id, SHA256: q.SHA256, Size: q.Size, ContentType: q.ContentType, Purpose: q.Purpose, State: "UPLOADING"}, ExpiresAt: expires, Uploads: urls}, nil
}

type storedRef struct {
	FileRef
	Key, Tenant, UploadID string
	Expires, Retain       time.Time
	Obligation            string
}

func (c *Catalog) lookup(ctx context.Context, tenant, id string) (storedRef, error) {
	var v storedRef
	if _, e := uuid.Parse(id); e != nil {
		return v, ErrNotFound
	}
	e := c.db.QueryRowContext(ctx, `SELECT id,COALESCE(object_version,''),sha256,size_bytes,content_type,purpose,state,object_key,tenant_id,COALESCE(upload_id,''),expires_at,retention_until,COALESCE(obligation_id,'') FROM file_refs WHERE tenant_id=$1 AND id=$2`, tenant, id).Scan(&v.ID, &v.Version, &v.SHA256, &v.Size, &v.ContentType, &v.Purpose, &v.State, &v.Key, &v.Tenant, &v.UploadID, &v.Expires, &v.Retain, &v.Obligation)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrNotFound
	}
	return v, e
}
func (c *Catalog) Complete(ctx context.Context, tenant, id string, parts []CompletedPart) (FileRef, error) {
	v, e := c.lookup(ctx, tenant, id)
	if e != nil {
		return FileRef{}, e
	}
	if v.State == "READY" {
		return v.FileRef, nil
	}
	if v.State != "UPLOADING" && v.State != "VALIDATING" {
		return FileRef{}, ErrIneligible
	}
	if time.Now().After(v.Expires) {
		return FileRef{}, ErrIneligible
	}
	if v.UploadID != "" {
		e = c.client.CompleteMultipart(ctx, v.Key, v.UploadID, parts, int((v.Size+PartSize-1)/PartSize))
		if e != nil {
			if _, headErr := c.client.Head(ctx, v.Key, ""); headErr != nil {
				return FileRef{}, e
			}
		}
	}
	m, e := c.client.Verify(ctx, v.Key, "", v.Size, v.SHA256, v.ContentType)
	if e != nil {
		return FileRef{}, e
	}
	res, e := c.db.ExecContext(ctx, `UPDATE file_refs SET state='READY',object_version=$3,updated_at=clock_timestamp() WHERE tenant_id=$1 AND id=$2 AND state IN('UPLOADING','VALIDATING') AND expires_at>clock_timestamp()`, tenant, id, m.Version)
	if e != nil {
		return FileRef{}, e
	}
	n, e := res.RowsAffected()
	if e != nil {
		return FileRef{}, e
	}
	if n != 1 {
		current, e := c.Resolve(ctx, tenant, id)
		if e != nil {
			return FileRef{}, ErrIneligible
		}
		return current, nil
	}
	v.Version = m.Version
	v.State = "READY"
	return v.FileRef, nil
}
func (c *Catalog) Resolve(ctx context.Context, tenant, id string) (FileRef, error) {
	v, e := c.lookup(ctx, tenant, id)
	if e != nil {
		return FileRef{}, e
	}
	if v.State == "PURGED" || v.State == "PURGING" {
		return FileRef{}, ErrNotFound
	}
	if v.State != "READY" {
		return FileRef{}, ErrIneligible
	}
	return v.FileRef, nil
}
func (c *Catalog) Download(ctx context.Context, tenant, id string) (string, error) {
	v, e := c.lookup(ctx, tenant, id)
	if e != nil {
		return "", e
	}
	if v.State != "READY" {
		return "", ErrNotFound
	}
	return c.client.DownloadURL(ctx, v.Key, v.Version, 2*time.Minute)
}

// Pin serializes with the purge claim. A newly opened obligation cannot race a
// deletion that already claimed PURGING. Callers must pin before accepting work.
func (c *Catalog) Pin(ctx context.Context, tenant, id, obligation, reason string) error {
	if obligation == "" || reason == "" {
		return ErrIneligible
	}
	tx, e := c.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var state string
	e = tx.QueryRowContext(ctx, `SELECT state FROM file_refs WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenant, id).Scan(&state)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrNotFound
	}
	if e != nil {
		return e
	}
	if state != "READY" && state != "ORPHAN" {
		return ErrIneligible
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO object_retention_pins(file_id,tenant_id,obligation_id,reason) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, id, tenant, obligation, reason)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (c *Catalog) Unpin(ctx context.Context, tenant, id, obligation, actor string) error {
	if actor == "" {
		return ErrIneligible
	}
	tx, e := c.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, `DELETE FROM object_retention_pins WHERE tenant_id=$1 AND file_id=$2 AND obligation_id=$3`, tenant, id, obligation)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO object_actions(tenant_id,file_id,action,actor) VALUES($1,$2,'UNPIN',$3)`, tenant, id, actor)
	if e != nil {
		return e
	}
	return tx.Commit()
}

// StoreResult persists an obligation before S3. A later protocol commit failure
// leaves ORPHAN plus its operation identity for reconciliation, never success.
func (c *Catalog) StoreResult(ctx context.Context, tenant, obligation string, q UploadRequest, reader io.Reader) (FileRef, error) {
	if tenant == "" || obligation == "" {
		return FileRef{}, ErrIneligible
	}
	retention, e := c.policy(ctx, q)
	if e != nil {
		return FileRef{}, e
	}
	id := uuid.NewString()
	key := "results/" + id
	_, e = c.db.ExecContext(ctx, `INSERT INTO file_refs(id,tenant_id,object_key,sha256,size_bytes,content_type,purpose,class,region,state,expires_at,retention_until,obligation_id) VALUES($1,$2,$3,'PENDING',$4,$5,$6,$7,$8,'VALIDATING',clock_timestamp()+interval '1 hour',$9,$10)`, id, tenant, key, q.Size, q.ContentType, q.Purpose, q.Class, q.Region, time.Now().UTC().Add(retention), obligation)
	if e != nil {
		return FileRef{}, e
	}
	m, e := c.client.PutStream(ctx, key, q.ContentType, reader, q.Size)
	if e != nil {
		return FileRef{}, e
	}
	_, e = c.db.ExecContext(ctx, `UPDATE file_refs SET state='ORPHAN',object_version=$3,sha256=$4,size_bytes=$5,updated_at=clock_timestamp() WHERE tenant_id=$1 AND id=$2 AND state='VALIDATING'`, tenant, id, m.Version, m.SHA256, m.Size)
	if e != nil {
		return FileRef{}, e
	}
	return FileRef{ID: id, Version: m.Version, SHA256: m.SHA256, Size: m.Size, ContentType: m.ContentType, Purpose: q.Purpose, State: "ORPHAN"}, nil
}
func (c *Catalog) LinkResult(ctx context.Context, tenant, id, obligation string) (FileRef, error) {
	tx, e := c.db.BeginTx(ctx, nil)
	if e != nil {
		return FileRef{}, e
	}
	defer tx.Rollback()
	var state, owner string
	e = tx.QueryRowContext(ctx, `SELECT state,COALESCE(obligation_id,'') FROM file_refs WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenant, id).Scan(&state, &owner)
	if errors.Is(e, sql.ErrNoRows) {
		return FileRef{}, ErrNotFound
	}
	if e != nil {
		return FileRef{}, e
	}
	if obligation == "" || owner != obligation || (state != "ORPHAN" && state != "READY") {
		return FileRef{}, ErrIneligible
	}
	_, e = tx.ExecContext(ctx, `UPDATE file_refs SET state='READY' WHERE tenant_id=$1 AND id=$2`, tenant, id)
	if e != nil {
		return FileRef{}, e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO object_retention_pins(file_id,tenant_id,obligation_id,reason) VALUES($1,$2,$3,'RESULT_CUSTODY') ON CONFLICT DO NOTHING`, id, tenant, obligation)
	if e != nil {
		return FileRef{}, e
	}
	if e = tx.Commit(); e != nil {
		return FileRef{}, e
	}
	return c.Resolve(ctx, tenant, id)
}

// Reconcile inspects only stored objects and never invokes an external provider.
func (c *Catalog) Reconcile(ctx context.Context, tenant string) (int, error) {
	rows, e := c.db.QueryContext(ctx, `SELECT id FROM file_refs WHERE tenant_id=$1 AND state='VALIDATING' AND obligation_id IS NOT NULL ORDER BY created_at LIMIT 100`, tenant)
	if e != nil {
		return 0, e
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return 0, e
		}
		ids = append(ids, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return 0, e
	}
	n := 0
	for _, id := range ids {
		v, e := c.lookup(ctx, tenant, id)
		if e != nil {
			return n, e
		}
		m, e := c.client.Head(ctx, v.Key, "")
		if e != nil {
			continue
		}
		if m.Version == "" || m.Size < 1 || m.Size > v.Size {
			continue
		}
		m, e = c.client.Verify(ctx, v.Key, m.Version, m.Size, "", v.ContentType)
		if e != nil {
			continue
		}
		_, e = c.db.ExecContext(ctx, `UPDATE file_refs SET state='ORPHAN',object_version=$3,sha256=$4,size_bytes=$5,updated_at=clock_timestamp() WHERE tenant_id=$1 AND id=$2 AND state='VALIDATING'`, tenant, id, m.Version, m.SHA256, m.Size)
		if e != nil {
			return n, e
		}
		n++
	}
	return n, nil
}
