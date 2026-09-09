package orbita

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

func TestPostgresRetryHorizonStartsOnceAndExpires(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	tenant, protocolID, commandID := "retry-horizon-"+idgen.New(), idgen.New(), idgen.New()
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM command_intents WHERE tenant_id=$1", tenant)
		_, _ = db.ExecContext(ctx, "DELETE FROM protocols WHERE tenant_id=$1", tenant)
	}()

	now := time.Now().UTC()
	p := Protocol{
		ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app-retry", CellID: "cell-retry",
		IdempotencyKey: "intent-1", RequestHash: "hash-1", RequestBody: []byte(`{"mode":"ASYNC"}`),
		Mode: "ASYNC", DispatchMode: string(dispatch.DispatchQueued), CommandID: commandID,
		Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute),
	}
	cmd := dispatch.Command{
		TenantID: tenant, ApplicationID: p.ApplicationID, CellID: p.CellID, ProtocolID: protocolID,
		StepID: "step-1", CommandID: commandID, DispatchMode: dispatch.DispatchQueued,
		ConfigSnapshot: []byte(`{"synthetic":true}`), RequestBody: map[string]any{"x": 1},
		AcceptedAt: now, StepDeadline: now.Add(time.Minute), RetryTTLSeconds: 1,
	}
	s := NewStore(db)
	if _, created, err := s.Admit(ctx, p, cmd, "subject-retry", false); err != nil || !created {
		t.Fatalf("admit created=%v err=%v", created, err)
	}

	first, err := s.ClaimIntent(ctx, p.CellID, "owner-1")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.CompleteIntent(ctx, first, false); err != nil || !ok {
		t.Fatalf("first failure ok=%v err=%v", ok, err)
	}
	var started, until time.Time
	if err := db.QueryRowContext(ctx, "SELECT retry_started_at,retry_until FROM command_intents WHERE command_id=$1", commandID).Scan(&started, &until); err != nil {
		t.Fatal(err)
	}
	if started.IsZero() || until.IsZero() || until.Sub(started) < 900*time.Millisecond || until.Sub(started) > 2*time.Second {
		t.Fatalf("invalid retry horizon started=%s until=%s", started, until)
	}

	if _, err := db.ExecContext(ctx, "UPDATE command_intents SET lease_until=clock_timestamp()-interval '1 second' WHERE command_id=$1", commandID); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	second, err := s.ClaimIntent(ctx, p.CellID, "owner-2")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.CompleteIntent(ctx, second, false); err != nil || !ok {
		t.Fatalf("expired failure ok=%v err=%v", ok, err)
	}
	var state string
	if err := db.QueryRowContext(ctx, "SELECT state FROM command_intents WHERE command_id=$1", commandID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "EXPIRED" {
		t.Fatalf("retry horizon renewed or intent remained publishable: state=%s", state)
	}
}
