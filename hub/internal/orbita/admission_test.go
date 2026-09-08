package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"ai-hub/hub/internal/contracts/files"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

func TestAdmissionPostgresAtomicIdempotency(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(12)
	store := NewStore(db)
	ctx := context.Background()
	tenant := "admission-test-" + idgen.New()
	defer func() {
		db.Exec("DELETE FROM command_intents WHERE tenant_id=$1", tenant)
		db.Exec("DELETE FROM protocols WHERE tenant_id=$1", tenant)
	}()
	makeAdmission := func(application, hash string) (Protocol, dispatch.Command) {
		now := time.Now().UTC()
		p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: application, CellID: "r2-cell-a", IdempotencyKey: "same-key", RequestHash: hash, RequestBody: json.RawMessage(`{"marker":"synthetic"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
		cmd := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"version":1}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
		return p, cmd
	}
	var wg sync.WaitGroup
	ids := make(chan string, 24)
	errs := make(chan error, 24)
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, c := makeAdmission("app-a", "same-hash")
			got, _, e := store.Admit(ctx, p, c, "synthetic-subject", false)
			if e != nil {
				errs <- e
				return
			}
			ids <- got.ProtocolID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	var winner string
	for id := range ids {
		if winner == "" {
			winner = id
		}
		if winner != id {
			t.Errorf("different accepted protocols: %s %s", winner, id)
		}
	}
	var count int
	if err = db.QueryRow("SELECT count(*) FROM command_intents WHERE tenant_id=$1", tenant).Scan(&count); err != nil || count != 1 {
		t.Fatalf("intent count=%d error=%v", count, err)
	}
	// Competing publishers use a database lease. A previous epoch cannot
	// acknowledge a message after takeover, even if its network call returns.
	i, err := store.ClaimIntent(ctx, "r2-cell-a", "publisher-a")
	if err != nil || i.Command.ProtocolID != winner {
		t.Fatalf("claim: %+v %v", i, err)
	}
	if _, err = store.ClaimIntent(ctx, "r2-cell-a", "publisher-b"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("concurrent claim: %v", err)
	}
	if _, err = db.Exec("UPDATE command_intents SET lease_until=clock_timestamp()-interval '1 second' WHERE command_id=$1", i.Command.CommandID); err != nil {
		t.Fatal(err)
	}
	j, err := store.ClaimIntent(ctx, "r2-cell-a", "publisher-b")
	if err != nil || j.Epoch <= i.Epoch {
		t.Fatalf("takeover: %+v %v", j, err)
	}
	if ok, e := store.CompleteIntent(ctx, i, true); e != nil || ok {
		t.Fatalf("stale acknowledgement accepted: %v", e)
	}
	if ok, e := store.CompleteIntent(ctx, j, true); e != nil || !ok {
		t.Fatalf("current acknowledgement rejected: %v", e)
	}
	p, c := makeAdmission("app-a", "different-hash")
	if _, _, e := store.Admit(ctx, p, c, "synthetic-subject", false); !errors.Is(e, ErrIdempotencyConflict) {
		t.Fatalf("expected conflict, got %v", e)
	}
	p, c = makeAdmission("app-b", "same-hash")
	other, created, e := store.Admit(ctx, p, c, "synthetic-subject", true)
	if e != nil || !created || other.ProtocolID == winner {
		t.Fatalf("application isolation failed: %v", e)
	}
	var state string
	if e = db.QueryRow("SELECT state FROM command_intents WHERE command_id=$1", c.CommandID).Scan(&state); e != nil || state != "WAITING_RESERVATION" {
		t.Fatalf("reservation guard: %s %v", state, e)
	}
	// Force the second insert to fail after protocol insertion; no partial acceptance survives.
	existingCommand := c.CommandID
	p, c = makeAdmission("app-c", "same-hash")
	p.CommandID = existingCommand
	c.CommandID = existingCommand
	if _, _, e = store.Admit(ctx, p, c, "synthetic-subject", false); e == nil {
		t.Fatal("duplicate command accepted")
	}
	if e = db.QueryRow("SELECT count(*) FROM protocols WHERE protocol_id=$1", p.ProtocolID).Scan(&count); e != nil || count != 0 {
		t.Fatalf("failed admission survived: %d %v", count, e)
	}
	fileID := idgen.New()
	if _, e = db.Exec(`INSERT INTO file_refs(id,tenant_id,object_key,object_version,sha256,size_bytes,content_type,purpose,class,region,state,expires_at,retention_until) VALUES($1,$2,$3,'fixture-v1','fixture-sha',16,'application/octet-stream','TEST','SYNTHETIC','fixture','READY',clock_timestamp()+interval '1 hour',clock_timestamp()+interval '1 hour')`, fileID, tenant, "admission-metadata-fixture/"+fileID); e != nil {
		t.Fatal(e)
	}
	defer func() {
		db.Exec("DELETE FROM object_retention_pins WHERE file_id=$1", fileID)
		db.Exec("DELETE FROM file_refs WHERE id=$1", fileID)
	}()
	p, c = makeAdmission("app-file-failed", "same-hash")
	p.CommandID = existingCommand
	c.CommandID = existingCommand
	c.FileRefs = []files.Reference{{ID: fileID, Version: "fixture-v1", SHA256: "fixture-sha"}}
	if _, _, e = store.Admit(ctx, p, c, "fixture-subject", false); e == nil {
		t.Fatal("duplicate command with pin accepted")
	}
	if e = db.QueryRow("SELECT count(*) FROM object_retention_pins WHERE file_id=$1", fileID).Scan(&count); e != nil || count != 0 {
		t.Fatalf("pin leaked after rollback: %d %v", count, e)
	}
	p, c = makeAdmission("app-file-success", "same-hash")
	c.FileRefs = []files.Reference{{ID: fileID, Version: "fixture-v1", SHA256: "fixture-sha"}}
	if _, _, e = store.Admit(ctx, p, c, "fixture-subject", false); e != nil {
		t.Fatal(e)
	}
	if e = db.QueryRow("SELECT count(*) FROM object_retention_pins WHERE file_id=$1 AND obligation_id=$2", fileID, p.ProtocolID).Scan(&count); e != nil || count != 1 {
		t.Fatalf("pin missing from acceptance: %d %v", count, e)
	}
	t.Log("24 simultaneous requests: one protocol+intent; conflicting payload refused; application keys independent; reservation not released; failed intent insertion rolls back protocol")
}
