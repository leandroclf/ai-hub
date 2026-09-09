package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

func TestPostgresSubmissionAndObservationCustody(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(12)
	s := NewStore(db)
	ctx := context.Background()
	id := idgen.New()
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	}()
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "custody-synthetic", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "synthetic-account", RequestBody: map[string]any{"input": "different-from-output"}, StepDeadline: time.Now().Add(time.Minute), ConfigSnapshot: json.RawMessage(`{"version":1}`)}
	var winners atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 24; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, owned, e := s.PrepareSubmission(ctx, cmd, "binding-a", "v1")
			if e != nil {
				t.Error(e)
			}
			if owned {
				winners.Add(1)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("submit owners=%d", winners.Load())
	}
	var count int
	if e := db.QueryRow("SELECT count(*) FROM attempts WHERE operation_id=$1 AND prepared_at IS NOT NULL", id).Scan(&count); e != nil || count != 1 {
		t.Fatalf("attempt count %d %v", count, e)
	}
	response := dispatch.Result{Kind: dispatch.FactSucceeded, ResponseBody: map[string]string{"provider_marker": "actual-output"}, ProviderRequestID: "provider-id"}
	for n := 0; n < 12; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, e := s.ConserveObservation(ctx, cmd, response, "synthetic-race")
			if e != nil || !got.Durable || got.EvidenceID == "" {
				t.Errorf("custody %v %+v", e, got)
			}
		}()
	}
	wg.Wait()
	if e := db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1", id).Scan(&count); e != nil || count != 1 {
		t.Fatalf("terminal fact count %d %v", count, e)
	}
	got, e := s.DurableResult(ctx, cmd)
	if e != nil {
		t.Fatal(e)
	}
	body, _ := json.Marshal(got.ResponseBody)
	if string(body) != `{"provider_marker":"actual-output"}` {
		t.Fatalf("provider response lost: %s", body)
	}
	other := cmd
	other.TenantID = "other"
	if _, e = s.DurableResult(ctx, other); e == nil {
		t.Fatal("cross tenant result disclosed")
	}
	if _, _, e = s.PrepareSubmission(ctx, other, "binding-a", "v1"); e == nil {
		t.Fatal("conflicting command reused")
	}
	// UNKNOWN after an already committed final cannot overwrite the result.
	if _, e = s.ConserveObservation(ctx, cmd, dispatch.Result{Kind: dispatch.FactUnknown}, "late-timeout"); e != nil {
		t.Fatal(e)
	}
	again, e := s.DurableResult(ctx, cmd)
	if e != nil || again.EvidenceID != got.EvidenceID || again.Kind != dispatch.FactSucceeded {
		t.Fatalf("late observation changed final: %+v %v", again, e)
	}
	t.Log("24 submit contenders: one owner and prewritten attempt; 12 observations: one terminal fact; actual response retained; tenant conflict denied; late UNKNOWN preserved as receipt")
}

func TestCallbackCapabilityIsRandomAndStoredOnlyAsHash(t *testing.T) {
	first, err := newCallbackToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newCallbackToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) < 40 {
		t.Fatalf("callback capability is not an independent high-entropy value")
	}
	hash := callbackTokenHash(first)
	if hash == first || len(hash) != 64 || hash == callbackTokenHash(second) {
		t.Fatalf("callback capability hashing is invalid")
	}
}

func TestPostgresPendingCustodyAtomic(t *testing.T) {
	t.Setenv("ENVIRONMENT", "local")
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)
	ctx := context.Background()
	id := idgen.New()
	defer func() {
		db.Exec("DELETE FROM polling_schedule WHERE operation_id=$1", id)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	}()
	// Missing durable polling budget is detected after tentative state/receipt
	// writes; rollback must leave neither a receipt nor a false acceptance.
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "pending-test", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "fixture", ConfigSnapshot: json.RawMessage(`{}`)}
	if _, owned, err := s.PrepareSubmission(ctx, cmd, "binding", "version"); err != nil || !owned {
		t.Fatalf("prepare: %v", err)
	}
	if r, err := s.ConserveAcceptance(ctx, cmd, "provider-pending", true); err == nil || r.Durable {
		t.Fatalf("false custody: %+v %v", r, err)
	}
	var state string
	var count int
	if err = db.QueryRow("SELECT state FROM operations WHERE operation_id=$1", id).Scan(&state); err != nil || state != "SUBMITTING" {
		t.Fatalf("partial state: %s %v", state, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM operation_receipts WHERE operation_id=$1", id).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial receipt: %d %v", count, err)
	}
	// Fixture adjustment of its persisted retry horizon, never a runtime backfill.
	cmd.RetryDeadline = time.Now().Add(time.Minute)
	raw, _ := json.Marshal(cmd)
	if _, err = db.Exec("UPDATE operations SET command=$2 WHERE operation_id=$1", id, raw); err != nil {
		t.Fatal(err)
	}
	got, err := s.ConserveAcceptance(ctx, cmd, "provider-pending", true)
	if err != nil || !got.Durable || got.EvidenceID == "" {
		t.Fatalf("acceptance: %+v %v", got, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM polling_schedule WHERE operation_id=$1 AND deadline_at=$2", id, cmd.RetryDeadline).Scan(&count); err != nil || count != 1 {
		t.Fatalf("polling obligation: %d %v", count, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1", id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("outbox: %d %v", count, err)
	}
	replay, err := s.DurableResult(ctx, cmd)
	if err != nil || !replay.Durable || replay.ProviderRequestID != "provider-pending" {
		t.Fatalf("replay: %+v %v", replay, err)
	}
	t.Log("real PostgreSQL: failed polling budget rolls back state and receipt; valid acceptance conserves correlation, receipt, schedule, result and outbox")
}
