package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/platform/pg"
	_ "github.com/lib/pq"
)

// TestPostgresCompensationUnknownThenDefinitiveFailureStaysReconciliable
// prova R6-EXE-03-S03: a compensação em si pode apresentar UNKNOWN (mantém-se
// pendente, nunca vira sucesso por omissão) e, ao esgotar sua política em uma
// falha definitiva, permanece reconciliável como COMPENSATION_FAILED — a
// falha nunca é promovida a sucesso do produto.
func TestPostgresCompensationUnknownThenDefinitiveFailureStaysReconciliable(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cell := "compensation-reconcile-cell-" + idgen.New()
	t.Setenv("CELL_ID", cell)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	protocolID := idgen.New()
	tenant := "compensation-reconcile-tenant-" + idgen.New()
	commands := []dispatch.Command{
		productTestCommand(protocolID, tenant, cell, "A", "service-a", now),
		productTestCommand(protocolID, tenant, cell, "B", "service-b", now),
	}
	compensation := productTestCommand(protocolID, tenant, cell, "compensate_A", "compensate-a", now)
	compensation.ServiceVersion = 1
	plan := ProductPlan{TargetID: "product-reconcile", TargetVersion: 1, MaxParallel: 2, AllowPartial: false, Consolidation: "ALL_REQUIRED", FailurePolicy: "COMPENSATE", Steps: []ProductPlanStep{
		{StepID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true, CompensationServiceID: "compensate-a", CommandID: commands[0].CommandID, CompensationCommand: compensation},
		{StepID: "B", ServiceID: "service-b", ServiceVersion: 1, Required: true, CommandID: commands[1].CommandID},
	}}
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, IdempotencyKey: "compensation-reconcile-key", RequestHash: "compensation-reconcile-hash", RequestBody: json.RawMessage(`{"marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commands[0].CommandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
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
	for _, command := range commands {
		intent, claimErr := store.ClaimIntent(ctx, cell, "publisher-"+command.StepID)
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		if _, completeErr := store.CompleteIntent(ctx, intent, true); completeErr != nil {
			t.Fatal(completeErr)
		}
	}
	if outcome, err := store.ApplyProductFact(ctx, tenant, protocolID, commands[0].CommandID, "SUCCEEDED", map[string]any{"marker": "a"}, ""); err != nil || outcome.Finalize {
		t.Fatalf("A succeeded outcome=%+v err=%v", outcome, err)
	}
	if outcome, err := store.ApplyProductFact(ctx, tenant, protocolID, commands[1].CommandID, "FAILED", map[string]any{"detail": "broke"}, "provider failure"); err != nil || outcome.Finalize {
		t.Fatalf("B failed outcome=%+v err=%v", outcome, err)
	}

	compensationIntent, err := store.ClaimIntent(ctx, cell, "publisher-compensate-a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.CompleteIntent(ctx, compensationIntent, true); err != nil {
		t.Fatal(err)
	}

	// The compensation call itself comes back UNKNOWN (transport uncertainty):
	// it must not be silently promoted to success or failure.
	if outcome, err := store.ApplyProductFact(ctx, tenant, protocolID, compensation.CommandID, "UNKNOWN", nil, ""); err != nil || outcome.Finalize {
		t.Fatalf("UNKNOWN compensation outcome=%+v err=%v", outcome, err)
	}
	var compensationState string
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='compensate_A'", protocolID).Scan(&compensationState)
	}); err != nil || compensationState != "WAITING_PROVIDER" {
		t.Fatalf("UNKNOWN compensation was not kept pending: state=%s err=%v", compensationState, err)
	}
	var originalState string
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='A'", protocolID).Scan(&originalState)
	}); err != nil || originalState != "COMPENSATING" {
		t.Fatalf("original step advanced past COMPENSATING while its compensation is still UNKNOWN: state=%s err=%v", originalState, err)
	}

	// The compensation policy exhausts into a definitive failure.
	outcome, err := store.ApplyProductFact(ctx, tenant, protocolID, compensation.CommandID, "FAILED", map[string]any{"detail": "compensation provider down"}, "compensation exhausted")
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Finalize || outcome.Status != StatusFailed {
		t.Fatalf("definitive compensation failure was not reconciliable as product failure: outcome=%+v", outcome)
	}
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='compensate_A'", protocolID).Scan(&compensationState)
	}); err != nil || compensationState != "FAILED" {
		t.Fatalf("definitive compensation failure state=%s err=%v", compensationState, err)
	}
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='A'", protocolID).Scan(&originalState)
	}); err != nil || originalState != "COMPENSATION_FAILED" {
		t.Fatalf("original step did not stay reconciliable as COMPENSATION_FAILED: state=%s err=%v", originalState, err)
	}
	if _, err = NewFinalizer(store, slog.Default()).Finalize(ctx, "", tenant, protocolID, outcome.ExpectedVersion, outcome.Status, outcome.Body, "COMPENSATION_FAILED"); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(ctx, tenant, protocolID)
	if err != nil || got.Status != StatusFailed {
		t.Fatalf("compensation failure was promoted away from FAILED: status=%s err=%v", got.Status, err)
	}
}
