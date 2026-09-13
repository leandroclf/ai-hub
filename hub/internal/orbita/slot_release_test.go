package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/platform/pg"
	_ "github.com/lib/pq"
)

// TestPostgresSlotReleaseNeverExceedsMaxParallelNorDoubleDecrements prova
// R6-EXE-04-S01/S02: intent, lease da etapa e contador do plano mudam na
// mesma transação (CompleteIntent), então uma "vaga" nunca fica perdida nem
// duplicada quando o retry de um passo expira ou quando uma conclusão é
// repetida (fato final concorrente/reenviado) depois de dois publishers
// disputarem a última vaga.
func TestPostgresSlotReleaseNeverExceedsMaxParallelNorDoubleDecrements(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cell := "slot-release-cell-" + idgen.New()
	t.Setenv("CELL_ID", cell)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	protocolID := idgen.New()
	tenant := "slot-release-tenant-" + idgen.New()
	commands := []dispatch.Command{
		productTestCommand(protocolID, tenant, cell, "A", "service-a", now),
		productTestCommand(protocolID, tenant, cell, "B", "service-b", now),
	}
	commands[0].RetryTTLSeconds = 0 // R6-EXE-04-S01: falha definitiva no primeiro retry expirado
	plan := ProductPlan{TargetID: "product-slots", TargetVersion: 1, MaxParallel: 1, AllowPartial: true, Consolidation: "ALL_REQUIRED", FailurePolicy: "STOP", Steps: []ProductPlanStep{
		{StepID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: false, CommandID: commands[0].CommandID},
		{StepID: "B", ServiceID: "service-b", ServiceVersion: 1, Required: false, CommandID: commands[1].CommandID},
	}}
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, IdempotencyKey: "slot-key", RequestHash: "slot-hash", RequestBody: json.RawMessage(`{"marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commands[0].CommandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	if _, created, err := store.AdmitProduct(ctx, p, commands, plan, "fixture", false); err != nil || !created {
		t.Fatalf("admit product: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", protocolID)
		pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
			tx.Exec("DELETE FROM command_intents WHERE protocol_id=$1", protocolID)
			tx.Exec("DELETE FROM operation_steps WHERE protocol_id=$1", protocolID)
			tx.Exec("DELETE FROM operation_plans WHERE protocol_id=$1", protocolID)
			tx.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
			return nil
		})
	})

	// max_parallel=1: A takes the only slot, B must not be claimable while it
	// is held — running_count nunca excede máximo.
	a, err := store.ClaimIntent(ctx, cell, "publisher-a")
	if err != nil {
		t.Fatal(err)
	}
	if a.Command.StepID != "A" {
		t.Fatalf("unexpected first claim: %s", a.Command.StepID)
	}
	if _, err = store.ClaimIntent(ctx, cell, "publisher-b"); err != sql.ErrNoRows {
		t.Fatalf("max_parallel not enforced while slot is held: %v", err)
	}
	assertRunningCount(t, ctx, db, tenant, protocolID, 1)

	// A's retry horizon is exhausted on the first transient failure
	// (RetryTTLSeconds=0): the intent, the step's RUNNING lease and the
	// plan's running_count all release in the same transaction.
	if ok, err := store.CompleteIntent(ctx, a, false); err != nil || !ok {
		t.Fatalf("complete a (exhausted retry) ok=%v err=%v", ok, err)
	}
	assertRunningCount(t, ctx, db, tenant, protocolID, 0)
	var stepAState string
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='A'", protocolID).Scan(&stepAState)
	}); err != nil || stepAState != "READY" {
		t.Fatalf("step A lease was not released atomically with the intent: state=%s err=%v", stepAState, err)
	}

	// A repeated/duplicate final confirmation for the very same claim (a
	// concurrent final fact arriving twice) must be a no-op, never a second
	// decrement — the slot A released must not be double-counted.
	if ok, err := store.CompleteIntent(ctx, a, false); err != nil || ok {
		t.Fatalf("duplicate completion was not a no-op: ok=%v err=%v", ok, err)
	}
	assertRunningCount(t, ctx, db, tenant, protocolID, 0)

	// The freed slot is usable exactly once: B can now take it.
	b, err := store.ClaimIntent(ctx, cell, "publisher-b")
	if err != nil || b.Command.StepID != "B" {
		t.Fatalf("released slot not reusable: %+v err=%v", b, err)
	}
	assertRunningCount(t, ctx, db, tenant, protocolID, 1)
	if ok, err := store.CompleteIntent(ctx, b, true); err != nil || !ok {
		t.Fatalf("complete b ok=%v err=%v", ok, err)
	}
	// Transport delivery alone does not release the slot: the step stays
	// RUNNING until the real observation arrives (ApplyProductFact), so a
	// transient duplicate delivery confirmation never frees a slot the
	// business outcome hasn't confirmed yet.
	assertRunningCount(t, ctx, db, tenant, protocolID, 1)
	// A never received a business observation (its transport retry merely
	// expired, and expiration never fabricates an outcome), so the product
	// stays open even once B's real fact arrives.
	if _, err := store.ApplyProductFact(ctx, tenant, protocolID, commands[1].CommandID, "SUCCEEDED", map[string]any{"marker": "b"}, ""); err != nil {
		t.Fatal(err)
	}
	assertRunningCount(t, ctx, db, tenant, protocolID, 0)
}

// TestPostgresProductStepTakeoverAfterCrashRecoversSameSlotAndFencesOldOwner
// prova R6-EXE-04-S03: a lease RUNNING de uma etapa de produto expira após um
// crash pré-publicação; o takeover com época nova reivindica exatamente a
// mesma vaga (running_count não muda, pois a vaga já estava contada) e o dono
// antigo, ao retornar, não consegue mais alterar o estado (fencing por
// época).
func TestPostgresProductStepTakeoverAfterCrashRecoversSameSlotAndFencesOldOwner(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cell := "step-takeover-cell-" + idgen.New()
	t.Setenv("CELL_ID", cell)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	protocolID := idgen.New()
	tenant := "step-takeover-tenant-" + idgen.New()
	commands := []dispatch.Command{productTestCommand(protocolID, tenant, cell, "A", "service-a", now)}
	plan := ProductPlan{TargetID: "product-takeover", TargetVersion: 1, MaxParallel: 1, AllowPartial: true, Consolidation: "ALL_REQUIRED", FailurePolicy: "STOP", Steps: []ProductPlanStep{
		{StepID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true, CommandID: commands[0].CommandID},
	}}
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, IdempotencyKey: "takeover-key", RequestHash: "takeover-hash", RequestBody: json.RawMessage(`{"marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commands[0].CommandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	if _, created, err := store.AdmitProduct(ctx, p, commands, plan, "fixture", false); err != nil || !created {
		t.Fatalf("admit product: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", protocolID)
		pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
			tx.Exec("DELETE FROM command_intents WHERE protocol_id=$1", protocolID)
			tx.Exec("DELETE FROM operation_steps WHERE protocol_id=$1", protocolID)
			tx.Exec("DELETE FROM operation_plans WHERE protocol_id=$1", protocolID)
			tx.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
			return nil
		})
	})

	oldOwner, err := store.ClaimIntent(ctx, cell, "publisher-crashed")
	if err != nil {
		t.Fatal(err)
	}
	assertRunningCount(t, ctx, db, tenant, protocolID, 1)

	// Simulate the crash: both the intent's and the step's RUNNING leases
	// expire pre-publication.
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		if _, execErr := tx.ExecContext(ctx, "UPDATE operation_steps SET lease_until=clock_timestamp()-interval '1 second' WHERE protocol_id=$1 AND step_id='A'", protocolID); execErr != nil {
			return execErr
		}
		_, execErr := tx.ExecContext(ctx, "UPDATE command_intents SET lease_until=clock_timestamp()-interval '1 second' WHERE command_id=$1", oldOwner.Command.CommandID)
		return execErr
	}); err != nil {
		t.Fatal(err)
	}

	newOwner, err := store.ClaimIntent(ctx, cell, "publisher-takeover")
	if err != nil {
		t.Fatal(err)
	}
	if newOwner.Epoch <= oldOwner.Epoch {
		t.Fatalf("takeover did not advance epoch: old=%d new=%d", oldOwner.Epoch, newOwner.Epoch)
	}
	// The very same slot was recovered, not a second one.
	assertRunningCount(t, ctx, db, tenant, protocolID, 1)

	// The old (crashed) owner returning must not alter any state.
	if ok, err := store.CompleteIntent(ctx, oldOwner, true); err != nil || ok {
		t.Fatalf("stale owner altered state after takeover: ok=%v err=%v", ok, err)
	}
	assertRunningCount(t, ctx, db, tenant, protocolID, 1)
	var stepState string
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='A'", protocolID).Scan(&stepState)
	}); err != nil || stepState != "RUNNING" {
		t.Fatalf("stale owner changed step state: state=%s err=%v", stepState, err)
	}

	// The current owner still closes the step normally.
	if ok, err := store.CompleteIntent(ctx, newOwner, true); err != nil || !ok {
		t.Fatalf("current owner completion rejected: ok=%v err=%v", ok, err)
	}
}

func assertRunningCount(t *testing.T, ctx context.Context, db *sql.DB, tenant, protocolID string, want int) {
	t.Helper()
	var got int
	if err := pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT running_count FROM operation_plans WHERE protocol_id=$1", protocolID).Scan(&got)
	}); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("running_count=%d want=%d", got, want)
	}
}
