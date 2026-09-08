package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
)

func TestAutoPostgresBoundedWait(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	f := NewFinalizer(s, logger)
	h := NewHandlers(s, nil, nil, nil, f, logger)
	fixture := func(key string) Protocol {
		t.Helper()
		now := time.Now().UTC()
		p := Protocol{ProtocolID: idgen.New(), TenantID: "auto-test", ApplicationID: "app-a", CellID: "auto-cell", IdempotencyKey: key + idgen.New(), RequestHash: "hash", RequestBody: json.RawMessage(`{}`), Mode: "AUTO", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(10 * time.Second)}
		c := dispatch.Command{CommandID: p.CommandID, ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
		if _, _, err := s.Admit(context.Background(), p, c, "fixture", false); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
			db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
			db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
		})
		return p
	}
	t.Run("quick local final", func(t *testing.T) {
		p := fixture("quick")
		done := make(chan error, 1)
		go func() {
			time.Sleep(100 * time.Millisecond)
			current, err := s.Get(context.Background(), p.TenantID, p.ProtocolID)
			if err == nil {
				_, err = f.Finalize(context.Background(), "", p.TenantID, p.ProtocolID, current.Version, StatusFailed, FinalBody{ErrorCode: "FIXTURE_FINAL"}, "")
			}
			done <- err
		}()
		r := httptest.NewRequest("POST", "/v1/protocols", nil)
		w := httptest.NewRecorder()
		h.respondAuto(w, r, p, 1)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		current, err := s.Get(context.Background(), p.TenantID, p.ProtocolID)
		if err != nil || w.Code != 200 || w.Body.String() != string(current.FinalRepresentation) {
			t.Fatalf("quick response status=%d err=%v", w.Code, err)
		}
	})
	t.Run("slow returns accepted and retry does not renew wait", func(t *testing.T) {
		p := fixture("slow")
		r := httptest.NewRequest("POST", "/v1/protocols", nil)
		w := httptest.NewRecorder()
		start := time.Now()
		h.respondAuto(w, r, p, 1)
		if w.Code != 202 || time.Since(start) < 900*time.Millisecond || time.Since(start) > 2*time.Second {
			t.Fatalf("slow response: %d elapsed=%v", w.Code, time.Since(start))
		}
		start = time.Now()
		again := httptest.NewRecorder()
		h.respondAuto(again, r, p, 1)
		if again.Code != 202 || time.Since(start) > 300*time.Millisecond {
			t.Fatal("retry renewed AUTO wait")
		}
		var count int
		var mode, state string
		if err := db.QueryRow("SELECT count(*),min(dispatch_mode),min(state) FROM command_intents WHERE protocol_id=$1", p.ProtocolID).Scan(&count, &mode, &state); err != nil || count != 1 || mode != "QUEUED" || state != "READY" {
			t.Fatalf("intent mutated: %d %s %s %v", count, mode, state, err)
		}
	})
	t.Run("disconnect conserves intent", func(t *testing.T) {
		p := fixture("disconnect")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := s.AwaitAuto(ctx, p, 1); err == nil {
			t.Fatal("cancellation ignored")
		}
		var count int
		if err := db.QueryRow("SELECT count(*) FROM command_intents WHERE protocol_id=$1 AND state='READY'", p.ProtocolID).Scan(&count); err != nil || count != 1 {
			t.Fatal("disconnection removed obligation")
		}
	})
	t.Log("PostgreSQL + HTTP handler: fast local FAILED final uses persisted bytes; slow AUTO returns 202 after configured wait; retry does not renew wait; disconnect preserves QUEUED obligation. No broker/provider end-to-end or success deadline qualification.")
}
