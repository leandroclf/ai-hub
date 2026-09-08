package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/queue"
)

func TestOperationFactPostgresCustody(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	t.Setenv("CELL_ID", "r2-fact-test")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(12)
	ctx := context.Background()
	s := NewStore(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	f := NewFinalizer(s, logger)
	now := time.Now().UTC()
	p := Protocol{ProtocolID: idgen.New(), TenantID: "fact-test-" + idgen.New(), ApplicationID: "app-a", CellID: "r2-fact-test", IdempotencyKey: "key", RequestHash: "hash", RequestBody: json.RawMessage(`{"input_marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{CommandID: p.CommandID, ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, _, err = s.Admit(ctx, p, c, "fixture", false); err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Exec(`DELETE FROM message_quarantine WHERE consumer='orbita' AND convert_from(body,'UTF8') LIKE $1`, "%"+p.ProtocolID+"%")
		db.Exec(`DELETE FROM orbita_fact_inbox WHERE protocol_id=$1`, p.ProtocolID)
		db.Exec(`DELETE FROM outbox WHERE aggregate_id=$1`, p.ProtocolID)
		db.Exec(`DELETE FROM command_intents WHERE protocol_id=$1`, p.ProtocolID)
		db.Exec(`DELETE FROM protocols WHERE protocol_id=$1`, p.ProtocolID)
	}()
	fact := operationFact{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, OperationID: p.CommandID, EvidenceID: idgen.New(), Kind: "SUCCEEDED", ResponseBody: map[string]any{"result_marker": "real-result"}}
	message := func(fact operationFact, id string) queue.ReceivedMessage {
		raw, _ := json.Marshal(fact)
		e := queue.Envelope{EventID: id, Type: "operation.observed", SchemaVersion: 1, Producer: "cometa", TenantID: fact.TenantID, ProtocolID: fact.ProtocolID, OccurredAt: now, RecordedAt: now, Payload: raw}
		body, _ := json.Marshal(e)
		return queue.ReceivedMessage{Envelope: e, RawBody: body}
	}
	m := message(fact, idgen.New())
	// A real closed SQL pool simulates unavailable final persistence, while the
	// independent receipt connection remains live. Returning an error forbids ACK.
	unavailable, _ := sql.Open("postgres", dsn)
	unavailable.Close()
	if err = s.ConsumeOperationFact(ctx, m, NewFinalizer(NewStore(unavailable), logger)); err == nil {
		t.Fatal("ack permitted after failed final persistence")
	}
	var disposition string
	if err = db.QueryRow(`SELECT disposition FROM orbita_fact_inbox WHERE event_id=$1`, m.Envelope.EventID).Scan(&disposition); err != nil || disposition != "RECEIVED" {
		t.Fatalf("missing recoverable receipt: %s %v", disposition, err)
	}
	got, err := s.Get(ctx, p.TenantID, p.ProtocolID)
	if err != nil || IsTerminal(got.Status) {
		t.Fatalf("failed effect changed protocol: %+v %v", got, err)
	}
	// Invalid kinds and cross-application observations cannot finalize anything.
	invalid := fact
	invalid.Kind = "UNRECOGNIZED"
	if err = s.ConsumeOperationFact(ctx, message(invalid, idgen.New()), f); err != nil {
		t.Fatal(err)
	}
	invalid = fact
	invalid.ApplicationID = "app-other"
	if err = s.ConsumeOperationFact(ctx, message(invalid, idgen.New()), f); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Get(ctx, p.TenantID, p.ProtocolID)
	if IsTerminal(got.Status) {
		t.Fatal("invalid observation finalized protocol")
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.ConsumeOperationFact(ctx, m, f); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	got, err = s.Get(ctx, p.TenantID, p.ProtocolID)
	if err != nil || got.Status != StatusSucceeded {
		t.Fatalf("recovery: %s %v", got.Status, err)
	}
	original := string(got.FinalRepresentation)
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM outbox WHERE aggregate_id=$1`, p.ProtocolID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicate final outbox: %d %v", count, err)
	}
	// Reusing an event identity with different content is quarantined.
	conflict := fact
	conflict.ResponseBody = map[string]any{"result_marker": "tampered"}
	if err = s.ConsumeOperationFact(ctx, message(conflict, m.Envelope.EventID), f); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Get(ctx, p.TenantID, p.ProtocolID)
	if string(got.FinalRepresentation) != original {
		t.Fatal("historical final replaced")
	}
	if err = db.QueryRow(`SELECT count(*) FROM message_quarantine WHERE consumer='orbita' AND convert_from(body,'UTF8') LIKE $1`, "%"+p.ProtocolID+"%").Scan(&count); err != nil || count != 3 {
		t.Fatalf("quarantine count=%d err=%v", count, err)
	}
	t.Log("real PostgreSQL: failed final persistence forbids ACK; receipt survives; 12 redeliveries produce one final/outbox; unknown kind, application mismatch and identity conflict quarantined")
}
