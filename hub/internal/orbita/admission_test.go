package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"reflect"
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
	cell := "admission-test-cell-" + idgen.New()
	defer func() {
		db.Exec("DELETE FROM command_intents WHERE tenant_id=$1", tenant)
		db.Exec("DELETE FROM protocols WHERE tenant_id=$1", tenant)
	}()
	makeAdmission := func(application, hash string) (Protocol, dispatch.Command) {
		now := time.Now().UTC()
		p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: application, CellID: cell, IdempotencyKey: "same-key", RequestHash: hash, RequestBody: json.RawMessage(`{"marker":"synthetic"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
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
	i, err := store.ClaimIntent(ctx, cell, "publisher-a")
	if err != nil || i.Command.ProtocolID != winner {
		t.Fatalf("claim: %+v %v", i, err)
	}
	if _, err = store.ClaimIntent(ctx, cell, "publisher-b"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("concurrent claim: %v", err)
	}
	if _, err = db.Exec("UPDATE command_intents SET lease_until=clock_timestamp()-interval '1 second' WHERE command_id=$1", i.Command.CommandID); err != nil {
		t.Fatal(err)
	}
	j, err := store.ClaimIntent(ctx, cell, "publisher-b")
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

func TestAdmissionFreezesConfigSnapshotAtAcceptance(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	ctx := context.Background()
	tenant := "snapshot-admission-" + idgen.New()
	protocolID := idgen.New()
	now := time.Now().UTC()
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app-snapshot", CellID: "snapshot-cell", IdempotencyKey: "snapshot-key", RequestHash: "snapshot-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	v1 := json.RawMessage(`{"profile_version":1,"output_field":"legacy_status"}`)
	command := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: v1, AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, command, "snapshot-fixture", false); err != nil || !created {
		t.Fatalf("admit snapshot: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", protocolID)
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
	})
	command.ConfigSnapshot = json.RawMessage(`{"profile_version":2,"output_field":"new_status"}`)
	var protocolSnapshot, intentRaw []byte
	if err := db.QueryRowContext(ctx, `SELECT config_snapshot FROM protocols WHERE protocol_id=$1`, protocolID).Scan(&protocolSnapshot); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT command FROM command_intents WHERE protocol_id=$1`, protocolID).Scan(&intentRaw); err != nil {
		t.Fatal(err)
	}
	var protocolValue, expectedValue map[string]any
	var intent struct {
		ConfigSnapshot map[string]any `json:"config_snapshot"`
	}
	if err := json.Unmarshal(protocolSnapshot, &protocolValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(v1, &expectedValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(intentRaw, &intent); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(protocolValue, expectedValue) || !reflect.DeepEqual(intent.ConfigSnapshot, expectedValue) {
		t.Fatalf("snapshot pós-aceite foi reinterpretado: protocol=%s intent=%s", protocolSnapshot, intentRaw)
	}
	t.Logf("PostgreSQL congelou config_snapshot v1 em protocol=%s; alterar a fonte local para v2 não reescreveu protocolo nem intent", protocolID)
}

func TestRecoverOrphanedIntentRequeuesSameDurableObligation(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	ctx := context.Background()
	tenant := "orphan-recovery-" + idgen.New()
	cell := "orphan-recovery-cell-" + idgen.New()
	now := time.Now().UTC()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app-orphan", CellID: cell, IdempotencyKey: "orphan-recovery-key", RequestHash: "orphan-recovery-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"fixture":true}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "orphan-fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM protocol_audit WHERE protocol_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	})
	if _, err = db.ExecContext(ctx, `UPDATE command_intents SET created_at=clock_timestamp()-interval '5 seconds',next_attempt_at=clock_timestamp()-interval '1 second',lease_owner='dead-executor',lease_until=clock_timestamp()-interval '1 second' WHERE command_id=$1`, p.CommandID); err != nil {
		t.Fatal(err)
	}

	recovered, ok, err := store.RecoverOrphanedIntent(ctx, cell, 2*time.Second)
	if err != nil || !ok || recovered.Command.CommandID != p.CommandID || recovered.Command.ProtocolID != p.ProtocolID {
		t.Fatalf("orphan recovery did not preserve identity: ok=%v intent=%+v err=%v", ok, recovered, err)
	}
	var state, lastError string
	if err = db.QueryRowContext(ctx, "SELECT state,last_error FROM command_intents WHERE command_id=$1", p.CommandID).Scan(&state, &lastError); err != nil {
		t.Fatal(err)
	}
	if state != "READY" || lastError != "orphan_recovered" {
		t.Fatalf("orphan intent was not requeued diagnostically: state=%s error=%s", state, lastError)
	}
	var audits int
	if err = db.QueryRowContext(ctx, "SELECT count(*) FROM protocol_audit WHERE protocol_id=$1 AND action='ORPHAN_DISPATCH_RECOVERED'", p.ProtocolID).Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("orphan recovery audit count=%d err=%v", audits, err)
	}
	claimed, err := store.ClaimIntent(ctx, cell, "publisher-after-orphan")
	if err != nil || claimed.Command.CommandID != p.CommandID || claimed.Epoch <= recovered.Epoch {
		t.Fatalf("publisher did not reclaim same intent: claim=%+v err=%v", claimed, err)
	}
	t.Logf("PostgreSQL: aceito órfão foi diagnosticado e reencaminhado com o mesmo command_id=%s, sem criar nova obrigação", p.CommandID)
}
