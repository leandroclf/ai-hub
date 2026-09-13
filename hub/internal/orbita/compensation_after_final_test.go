package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/platform/pg"
	_ "github.com/lib/pq"
)

// TestPostgresCompensationProgressesAfterPublicFinalState prova R6-EXE-03-S01:
// A e B concluem, C falha e o cliente expira (protocolo já finalizado como
// EXPIRED, um estado público terminal). Um recuperador que assume depois
// disso ainda precisa poder reivindicar e concluir a compensação devida — o
// encerramento voltado ao cliente não pode impedir uma obrigação
// compensatória pendente — e o final público EXPIRED não pode ser reaberto.
func TestPostgresCompensationProgressesAfterPublicFinalState(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cell := "compensation-after-final-cell-" + idgen.New()
	t.Setenv("CELL_ID", cell)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	protocolID := idgen.New()
	tenant := "compensation-after-final-tenant-" + idgen.New()
	commands := []dispatch.Command{
		productTestCommand(protocolID, tenant, cell, "A", "service-a", now),
		productTestCommand(protocolID, tenant, cell, "B", "service-b", now),
	}
	compensation := productTestCommand(protocolID, tenant, cell, "compensate_A", "compensate-a", now)
	compensation.ServiceVersion = 1
	plan := ProductPlan{TargetID: "product-final", TargetVersion: 1, MaxParallel: 2, AllowPartial: false, Consolidation: "ALL_REQUIRED", FailurePolicy: "COMPENSATE", Steps: []ProductPlanStep{
		{StepID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true, CompensationServiceID: "compensate-a", CommandID: commands[0].CommandID, CompensationCommand: compensation},
		{StepID: "B", ServiceID: "service-b", ServiceVersion: 1, Required: true, CommandID: commands[1].CommandID},
	}}
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, IdempotencyKey: "compensation-final-key", RequestHash: "compensation-final-hash", RequestBody: json.RawMessage(`{"marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commands[0].CommandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
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

	// The client SLA elapses and the deadline sweep finalizes the protocol as
	// EXPIRED — a public terminal state — before compensation for A has run.
	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	protocol, err := store.Get(ctx, tenant, protocolID)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := finalizer.Finalize(ctx, "", tenant, protocolID, protocol.Version, StatusExpired, FinalBody{ProtocolID: protocolID, ErrorCode: "SLA_EXCEEDED", ErrorMessage: "prazo do cliente expirado"}, "SLA_EXCEEDED")
	if err != nil || !applied {
		t.Fatalf("client-facing expiry finalize applied=%v err=%v", applied, err)
	}

	// The public state is now terminal, yet A's compensation intent must
	// still be claimable: encerrar o protocolo não pode impedir compensação
	// já devida.
	compensationIntent, err := store.ClaimIntent(ctx, cell, "publisher-compensate-a")
	if err != nil {
		t.Fatalf("compensation intent unclaimable after public final state: %v", err)
	}
	if compensationIntent.Command.StepID != "compensate_A" {
		t.Fatalf("unexpected claimed step: %s", compensationIntent.Command.StepID)
	}
	if _, err = store.CompleteIntent(ctx, compensationIntent, true); err != nil {
		t.Fatal(err)
	}
	outcome, err := store.ApplyProductFact(ctx, tenant, protocolID, compensation.CommandID, "SUCCEEDED", map[string]any{"compensated": true}, "")
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Finalize {
		// The store computed a natural FAILED consolidation, but finalizing it
		// must be a no-op: the public state already closed as EXPIRED.
		if reapplied, err := finalizer.Finalize(ctx, "", tenant, protocolID, outcome.ExpectedVersion, outcome.Status, outcome.Body, "COMPENSATED"); err != nil || reapplied {
			t.Fatalf("compensation reopened the public final state: reapplied=%v err=%v", reapplied, err)
		}
	}

	var stepState string
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='A'", protocolID).Scan(&stepState)
	}); err != nil || stepState != "COMPENSATED" {
		t.Fatalf("original step state=%s err=%v", stepState, err)
	}
	got, err := store.Get(ctx, tenant, protocolID)
	if err != nil || got.Status != StatusExpired {
		t.Fatalf("public final state was reopened: status=%s err=%v", got.Status, err)
	}
}
