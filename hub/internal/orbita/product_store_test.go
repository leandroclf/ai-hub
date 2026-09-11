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

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/queue"
)

func TestProductPlanPersistsDependenciesAndConsolidatesAfterAllSteps(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cell := "product-plan-cell-" + idgen.New()
	t.Setenv("CELL_ID", cell)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	protocolID := idgen.New()
	tenant := "product-plan-tenant-" + idgen.New()
	commands := []dispatch.Command{
		productTestCommand(protocolID, tenant, cell, "A", "service-a", now),
		productTestCommand(protocolID, tenant, cell, "B", "service-b", now),
		productTestCommand(protocolID, tenant, cell, "C", "service-c", now),
	}
	plan := ProductPlan{TargetID: "product-1", TargetVersion: 1, MaxParallel: 2, AllowPartial: false, Consolidation: "ALL_REQUIRED", FailurePolicy: "STOP", Steps: []ProductPlanStep{
		{StepID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true, CommandID: commands[0].CommandID},
		{StepID: "B", ServiceID: "service-b", ServiceVersion: 1, Required: true, CommandID: commands[1].CommandID},
		{StepID: "C", ServiceID: "service-c", ServiceVersion: 1, Required: true, DependsOn: []string{"A", "B"}, InputMapping: map[string]string{"marker": "A.marker"}, CommandID: commands[2].CommandID},
	}}
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, IdempotencyKey: "product-key", RequestHash: "product-hash", RequestBody: json.RawMessage(`{"marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commands[0].CommandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	if _, created, err := store.AdmitProduct(ctx, p, commands, plan, "fixture", false); err != nil || !created {
		t.Fatalf("admit product: created=%v err=%v", created, err)
	}
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", protocolID)
		db.Exec("DELETE FROM operation_steps WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM operation_plans WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
	}()
	first, err := store.ClaimIntent(ctx, cell, "publisher-a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.ClaimIntent(ctx, cell, "publisher-b")
	if err != nil {
		t.Fatal(err)
	}
	if first.Command.StepID == second.Command.StepID {
		t.Fatalf("parallel claims selected the same step: %s", first.Command.StepID)
	}
	if _, err = store.ClaimIntent(ctx, cell, "publisher-c"); err != sql.ErrNoRows {
		t.Fatalf("max_parallel was not enforced: %v", err)
	}
	if _, err = store.CompleteIntent(ctx, first, true); err != nil {
		t.Fatal(err)
	}
	if _, err = store.CompleteIntent(ctx, second, true); err != nil {
		t.Fatal(err)
	}
	if outcome, err := store.ApplyProductFact(ctx, protocolID, commands[0].CommandID, "SUCCEEDED", map[string]any{"marker": "a"}, ""); err != nil || outcome.Finalize {
		t.Fatalf("first step outcome=%+v err=%v", outcome, err)
	}
	var state string
	if err = db.QueryRowContext(ctx, "SELECT state FROM command_intents WHERE command_id=$1", commands[2].CommandID).Scan(&state); err != nil || state != "PENDING" {
		t.Fatalf("dependent step released too early: state=%s err=%v", state, err)
	}
	if outcome, err := store.ApplyProductFact(ctx, protocolID, commands[1].CommandID, "SUCCEEDED", map[string]any{"marker": "b"}, ""); err != nil || outcome.Finalize {
		t.Fatalf("second step outcome=%+v err=%v", outcome, err)
	}
	var rawCommand []byte
	if err = db.QueryRowContext(ctx, "SELECT state,command FROM command_intents WHERE command_id=$1", commands[2].CommandID).Scan(&state, &rawCommand); err != nil || state != "READY" {
		t.Fatalf("dependent step was not released: state=%s err=%v", state, err)
	}
	var dependent dispatch.Command
	if err = json.Unmarshal(rawCommand, &dependent); err != nil {
		t.Fatalf("dependent command JSON invalid: %v", err)
	}
	requestBody, _ := json.Marshal(dependent.RequestBody)
	if string(requestBody) != `{"marker":"a"}` {
		t.Fatalf("dependency input was not materialized: %+v err=%v", dependent, err)
	}
	fact := operationFact{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, StepID: "C", OperationID: commands[2].CommandID, EvidenceID: idgen.New(), Kind: "SUCCEEDED", ResponseBody: map[string]any{"marker": "c"}}
	factRaw, _ := json.Marshal(fact)
	envelope := queue.Envelope{EventID: idgen.New(), Type: "operation.observed", SchemaVersion: 1, Producer: "cometa", TenantID: tenant, ProtocolID: protocolID, OccurredAt: now, RecordedAt: now, Payload: factRaw}
	envelopeRaw, _ := json.Marshal(envelope)
	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err = store.ConsumeOperationFact(ctx, queue.ReceivedMessage{Envelope: envelope, RawBody: envelopeRaw}, finalizer); err != nil {
		t.Fatalf("consume final product fact: %v", err)
	}
	got, err := store.Get(ctx, tenant, protocolID)
	if err != nil || got.Status != StatusSucceeded {
		t.Fatalf("persisted product status=%s err=%v", got.Status, err)
	}
}

func TestProductPlanCompensationIsASeparateTrackedOperation(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cell := "product-compensation-cell-" + idgen.New()
	t.Setenv("CELL_ID", cell)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	protocolID := idgen.New()
	tenant := "product-compensation-tenant-" + idgen.New()
	commands := []dispatch.Command{
		productTestCommand(protocolID, tenant, cell, "A", "service-a", now),
		productTestCommand(protocolID, tenant, cell, "B", "service-b", now),
	}
	compensation := productTestCommand(protocolID, tenant, cell, "compensate_A", "compensate-a", now)
	compensation.ServiceVersion = 1
	plan := ProductPlan{TargetID: "product-2", TargetVersion: 1, MaxParallel: 2, AllowPartial: false, Consolidation: "ALL_REQUIRED", FailurePolicy: "COMPENSATE", Steps: []ProductPlanStep{
		{StepID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true, CompensationServiceID: "compensate-a", CommandID: commands[0].CommandID, CompensationCommand: compensation},
		{StepID: "B", ServiceID: "service-b", ServiceVersion: 1, Required: true, CommandID: commands[1].CommandID},
	}}
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, IdempotencyKey: "compensation-key", RequestHash: "compensation-hash", RequestBody: json.RawMessage(`{"marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commands[0].CommandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	if _, created, err := store.AdmitProduct(ctx, p, commands, plan, "fixture", false); err != nil || !created {
		t.Fatalf("admit product: created=%v err=%v", created, err)
	}
	defer func() {
		db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM operation_steps WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM operation_plans WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
	}()
	for _, command := range commands {
		intent, claimErr := store.ClaimIntent(ctx, cell, "publisher-"+command.StepID)
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		if _, completeErr := store.CompleteIntent(ctx, intent, true); completeErr != nil {
			t.Fatal(completeErr)
		}
	}
	if outcome, err := store.ApplyProductFact(ctx, protocolID, commands[0].CommandID, "SUCCEEDED", map[string]any{"marker": "a"}, ""); err != nil || outcome.Finalize {
		t.Fatalf("success before failure outcome=%+v err=%v", outcome, err)
	}
	if outcome, err := store.ApplyProductFact(ctx, protocolID, commands[1].CommandID, "FAILED", map[string]any{"detail": "broke"}, "provider failure"); err != nil || outcome.Finalize {
		t.Fatalf("compensation should remain pending: outcome=%+v err=%v", outcome, err)
	}
	var state string
	if err = db.QueryRowContext(ctx, "SELECT state FROM command_intents WHERE command_id=$1", compensation.CommandID).Scan(&state); err != nil || state != "READY" {
		t.Fatalf("compensation intent state=%s err=%v", state, err)
	}
	outcome, err := store.ApplyProductFact(ctx, protocolID, compensation.CommandID, "SUCCEEDED", map[string]any{"compensated": true}, "")
	if err != nil || !outcome.Finalize || outcome.Status != StatusFailed {
		t.Fatalf("compensation outcome=%+v err=%v", outcome, err)
	}
	if err = db.QueryRowContext(ctx, "SELECT state FROM operation_steps WHERE protocol_id=$1 AND step_id='A'", protocolID).Scan(&state); err != nil || state != "COMPENSATED" {
		t.Fatalf("original step state=%s err=%v", state, err)
	}
}

func TestProductOptionalFailureFinalizesPartialWithDurableStepStates(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cell := "product-partial-cell-" + idgen.New()
	t.Setenv("CELL_ID", cell)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	protocolID := idgen.New()
	tenant := "product-partial-tenant-" + idgen.New()
	commands := []dispatch.Command{
		productTestCommand(protocolID, tenant, cell, "A", "service-a", now),
		productTestCommand(protocolID, tenant, cell, "B", "service-b", now),
	}
	plan := ProductPlan{TargetID: "product-partial", TargetVersion: 1, MaxParallel: 2, AllowPartial: true, Consolidation: "ALL_REQUIRED", FailurePolicy: "STOP", Steps: []ProductPlanStep{
		{StepID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true, CommandID: commands[0].CommandID},
		{StepID: "B", ServiceID: "service-b", ServiceVersion: 1, Required: false, CommandID: commands[1].CommandID},
	}}
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, IdempotencyKey: "partial-key", RequestHash: "partial-hash", RequestBody: json.RawMessage(`{"marker":"input"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commands[0].CommandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	if _, created, err := store.AdmitProduct(ctx, p, commands, plan, "fixture", false); err != nil || !created {
		t.Fatalf("admit product: created=%v err=%v", created, err)
	}
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", protocolID)
		db.Exec("DELETE FROM orbita_fact_inbox WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM operation_steps WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM operation_plans WHERE protocol_id=$1", protocolID)
		db.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
	}()
	for _, command := range commands {
		intent, claimErr := store.ClaimIntent(ctx, cell, "publisher-"+command.StepID)
		if claimErr != nil {
			t.Fatal(claimErr)
		}
		if _, completeErr := store.CompleteIntent(ctx, intent, true); completeErr != nil {
			t.Fatal(completeErr)
		}
	}
	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	consume := func(command dispatch.Command, kind string, response map[string]any, message string) error {
		fact := operationFact{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, StepID: command.StepID, OperationID: command.CommandID, EvidenceID: idgen.New(), Kind: kind, ResponseBody: response, ErrorMessage: message}
		payload, marshalErr := json.Marshal(fact)
		if marshalErr != nil {
			return marshalErr
		}
		envelope := queue.Envelope{EventID: idgen.New(), Type: "operation.observed", SchemaVersion: 1, Producer: "cometa", TenantID: tenant, ProtocolID: protocolID, OccurredAt: now, RecordedAt: now, Payload: payload}
		raw, marshalErr := json.Marshal(envelope)
		if marshalErr != nil {
			return marshalErr
		}
		return store.ConsumeOperationFact(ctx, queue.ReceivedMessage{Envelope: envelope, RawBody: raw}, finalizer)
	}
	if err := consume(commands[0], "SUCCEEDED", map[string]any{"marker": "a"}, ""); err != nil {
		t.Fatalf("consume required success: %v", err)
	}
	if err := consume(commands[1], "FAILED", map[string]any{"detail": "optional failure"}, "optional step failed"); err != nil {
		t.Fatalf("consume optional failure: %v", err)
	}
	got, err := store.Get(ctx, tenant, protocolID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusPartiallySucceeded {
		t.Fatalf("product status=%s, want PARTIALLY_SUCCEEDED", got.Status)
	}
	var body FinalBody
	if err := json.Unmarshal(got.FinalBody, &body); err != nil {
		t.Fatal(err)
	}
	result, ok := body.Result.(map[string]any)
	if !ok || result["steps"] == nil || result["failed_steps"] == nil {
		t.Fatalf("partial representation omitted step outcomes: %+v", body)
	}
	steps, ok := result["steps"].(map[string]any)
	if !ok || steps["A"] == nil {
		t.Fatalf("required successful step absent from partial representation: %+v", result)
	}
	failed, ok := result["failed_steps"].([]any)
	if !ok || len(failed) != 1 || failed[0] != "B" {
		t.Fatalf("optional failure not represented exactly once: %+v", result["failed_steps"])
	}
	var states string
	if err := db.QueryRowContext(ctx, "SELECT string_agg(step_id || ':' || state, ',' ORDER BY step_id) FROM operation_steps WHERE protocol_id=$1", protocolID).Scan(&states); err != nil {
		t.Fatal(err)
	}
	if states != "A:SUCCEEDED,B:FAILED" {
		t.Fatalf("durable optional step states=%s", states)
	}
}

func productTestCommand(protocolID, tenant, cell, step, serviceID string, accepted time.Time) dispatch.Command {
	data, _ := json.Marshal(atlas.CatalogData{InputSchema: json.RawMessage(`{"type":"object","properties":{"marker":{"type":"string"}},"required":["marker"]}`)})
	snapshot, _ := json.Marshal(atlas.OfferSnapshot{Target: atlas.Resource{Kind: "services", ID: serviceID, Version: 1, Data: data}})
	return dispatch.Command{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: cell, StepID: step, CommandID: idgen.New(), DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: snapshot, EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: accepted, StepDeadline: accepted.Add(time.Minute), RetryTTLSeconds: 30}
}
