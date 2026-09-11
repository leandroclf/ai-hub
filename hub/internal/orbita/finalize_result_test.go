package orbita

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/objectstore"
	"ai-hub/hub/internal/platform/idgen"
)

type resultCatalogFixture struct {
	ref              objectstore.FileRef
	stored           []byte
	query            objectstore.UploadRequest
	linked           bool
	linkedID         string
	linkedObligation string
}

func (f *resultCatalogFixture) StoreResult(_ context.Context, _, _ string, q objectstore.UploadRequest, reader io.Reader) (objectstore.FileRef, error) {
	f.query = q
	var err error
	f.stored, err = io.ReadAll(reader)
	if err != nil {
		return objectstore.FileRef{}, err
	}
	return f.ref, nil
}

func (f *resultCatalogFixture) LinkResult(_ context.Context, _, id, obligation string) (objectstore.FileRef, error) {
	f.linked = true
	f.linkedID = id
	f.linkedObligation = obligation
	return f.ref, nil
}

func TestFinalizeMaterializesLargeResultAndLinksAfterCommit(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	tenant := "large-result-tenant-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "large-result-cell", IdempotencyKey: "large-result-key", RequestHash: "large-result-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"fixture":true}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	}()
	fixture := &resultCatalogFixture{ref: objectstore.FileRef{ID: "result-file-1", Version: "v1", SHA256: "fixture-sha", Size: 2, ContentType: "application/json", Purpose: "RESULT", State: "ORPHAN"}}
	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	finalizer.SetResultCatalog(fixture)
	large := strings.Repeat("x", int(objectstore.InlineLimit)+1)
	if applied, err := finalizer.Finalize(ctx, "trace", tenant, p.ProtocolID, 0, StatusSucceeded, FinalBody{Result: map[string]any{"large": large}}, ""); err != nil || !applied {
		t.Fatalf("finalize large result: applied=%v err=%v", applied, err)
	}
	if len(fixture.stored) <= int(objectstore.InlineLimit) || !bytes.Contains(fixture.stored, []byte(`"large"`)) || fixture.query.Size != int64(len(fixture.stored)) || !fixture.linked || fixture.linkedID != fixture.ref.ID || fixture.linkedObligation != p.ProtocolID {
		t.Fatalf("object custody not completed: size=%d query=%+v linked=%v id=%s obligation=%s", len(fixture.stored), fixture.query, fixture.linked, fixture.linkedID, fixture.linkedObligation)
	}
	got, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	var final FinalBody
	if err = json.Unmarshal(got.FinalBody, &final); err != nil {
		t.Fatal(err)
	}
	result, ok := final.Result.(map[string]any)
	if !ok || result["file_id"] != fixture.ref.ID || result["version"] != fixture.ref.Version {
		t.Fatalf("final body does not expose stable FileRef: %#v", final.Result)
	}
}
