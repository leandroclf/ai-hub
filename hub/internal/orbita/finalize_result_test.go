package orbita

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/libra"
	"ai-hub/hub/internal/objectstore"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/queue"
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

func TestFinalizeUsesProtocolSnapshotWhenIntentIsMissing(t *testing.T) {
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
	tenant := "missing-intent-tenant-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "missing-intent-cell", IdempotencyKey: "missing-intent-key", RequestHash: "missing-intent-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"fixture":true}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	}()
	if _, err = db.ExecContext(ctx, "DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID); err != nil {
		t.Fatal(err)
	}

	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if applied, err := finalizer.Finalize(ctx, "trace", tenant, p.ProtocolID, 0, StatusExpired, FinalBody{ErrorCode: "SLA_EXCEEDED"}, "SLA_EXCEEDED"); err != nil || !applied {
		t.Fatalf("finalize without intent: applied=%v err=%v", applied, err)
	}
	var outboxCount int
	if err = db.QueryRowContext(ctx, "SELECT count(*) FROM outbox WHERE aggregate_id=$1 AND event_type='protocol.finalized'", p.ProtocolID).Scan(&outboxCount); err != nil {
		t.Fatal(err)
	}
	if outboxCount != 1 {
		t.Fatalf("expected one durable finalization fact, got %d", outboxCount)
	}
}

func TestFinalizeLateSuccessBecomesExpired(t *testing.T) {
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
	tenant := "late-success-tenant-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "late-success-cell", IdempotencyKey: "late-success-key", RequestHash: "late-success-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"fixture":true}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	}()
	if _, err = db.ExecContext(ctx, "UPDATE protocols SET client_deadline_at=clock_timestamp()-interval '1 second' WHERE protocol_id=$1", p.ProtocolID); err != nil {
		t.Fatal(err)
	}

	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if applied, err := finalizer.Finalize(ctx, "late-trace", tenant, p.ProtocolID, 0, StatusSucceeded, FinalBody{Result: map[string]any{"result_marker": "provider"}}, ""); err != nil || !applied {
		t.Fatalf("late finalization: applied=%v err=%v", applied, err)
	}
	got, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusExpired || !strings.Contains(string(got.FinalBody), "SLA_EXCEEDED") || strings.Contains(string(got.FinalBody), "result_marker") {
		t.Fatalf("late success escaped as terminal success: status=%s body=%s", got.Status, got.FinalBody)
	}
	t.Logf("late success was arbitrated as EXPIRED with SLA_EXCEEDED and no provider result: protocol=%s", p.ProtocolID)
}

func TestLateCostlyResultKeepsErrorRepresentationAndFinanceContest(t *testing.T) {
	coreDSN := os.Getenv("R2_CORE_TEST_DSN")
	financeDSN := os.Getenv("R2_FINANCE_DSN")
	if coreDSN == "" || financeDSN == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN and R2_FINANCE_DSN")
	}
	coreDB, err := sql.Open("postgres", coreDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer coreDB.Close()
	financeDB, err := sql.Open("postgres", financeDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = financeDB.Close() })
	if err = coreDB.Ping(); err != nil {
		t.Fatal(err)
	}
	if err = financeDB.Ping(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	core := NewStore(coreDB)
	finance := libra.NewStore(financeDB)
	now := time.Now().UTC()
	tenant := "late-costly-" + idgen.New()
	protocolID := idgen.New()
	commandID := idgen.New()
	economic := libra.Snapshot{
		ContractID: "late-costly-v1", Version: 1, Currency: "BRL", SettlementParty: "HUB",
		Buy:  []libra.PricingRule{{Meter: "external-execution", Amount: "0.2", Incidence: []string{"SUCCEEDED"}, UnitScope: "OPERATION"}},
		Sell: []libra.PricingRule{{Meter: "customer-result", Amount: "1", Incidence: []string{"SUCCEEDED"}, UnitScope: "PROTOCOL"}},
	}
	economicRaw, _ := json.Marshal(economic)
	p := Protocol{ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app", CellID: "late-costly-cell", IdempotencyKey: "late-costly-key", RequestHash: "late-costly-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: commandID, Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{ProtocolID: protocolID, TenantID: tenant, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: commandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"fixture":true}`), EconomicSnapshot: economicRaw, AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := core.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	coreCleanup := func() {
		_, _ = coreDB.Exec("DELETE FROM outbox WHERE aggregate_id=$1", protocolID)
		_, _ = coreDB.Exec("DELETE FROM command_intents WHERE protocol_id=$1", protocolID)
		_, _ = coreDB.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
	}
	t.Cleanup(coreCleanup)
	financeCleanup := func() {
		_, _ = financeDB.Exec("DELETE FROM finance_disputes WHERE tenant_id=$1", tenant)
		_, _ = financeDB.Exec("DELETE FROM finance_quarantine WHERE tenant_id=$1", tenant)
		_, _ = financeDB.Exec("DELETE FROM finance_inbox WHERE tenant_id=$1", tenant)
		_, _ = financeDB.Exec("DELETE FROM ledger_entries WHERE tenant_id=$1", tenant)
		_, _ = financeDB.Exec("DELETE FROM economic_facts WHERE tenant_id=$1", tenant)
		_, _ = financeDB.Exec("DELETE FROM finance_snapshots WHERE tenant_id=$1", tenant)
		_, _ = financeDB.Exec("DELETE FROM reservations WHERE tenant_id=$1", tenant)
		_, _ = financeDB.Exec("DELETE FROM credit_limits WHERE tenant_id=$1", tenant)
	}
	t.Cleanup(financeCleanup)
	if _, err = coreDB.ExecContext(ctx, "UPDATE protocols SET client_deadline_at=clock_timestamp()-interval '1 second' WHERE protocol_id=$1", protocolID); err != nil {
		t.Fatal(err)
	}

	finalizer := NewFinalizer(core, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if applied, err := finalizer.Finalize(ctx, "late-costly-trace", tenant, protocolID, 0, StatusSucceeded, FinalBody{Result: map[string]any{"provider_result": "late"}}, ""); err != nil || !applied {
		t.Fatalf("late finalization: applied=%v err=%v", applied, err)
	}
	got, err := core.Get(ctx, tenant, protocolID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusExpired || !strings.Contains(string(got.FinalBody), "SLA_EXCEEDED") || strings.Contains(string(got.FinalBody), "provider_result") {
		t.Fatalf("resultado tardio escapou na consulta GET: status=%s body=%s", got.Status, got.FinalBody)
	}

	var finalPayload []byte
	if err = coreDB.QueryRowContext(ctx, "SELECT payload FROM outbox WHERE aggregate_id=$1 AND event_type='protocol.finalized'", protocolID).Scan(&finalPayload); err != nil {
		t.Fatal(err)
	}
	var finalFact ProtocolFinalizedFact
	if err = json.Unmarshal(finalPayload, &finalFact); err != nil {
		t.Fatal(err)
	}
	if finalFact.Status != string(StatusExpired) || !strings.Contains(string(finalFact.Representation), "SLA_EXCEEDED") || string(finalFact.Representation) != string(got.FinalRepresentation) {
		t.Fatalf("representação para webhook divergiu do GET: status=%s representation=%s", finalFact.Status, finalFact.Representation)
	}

	if _, err = financeDB.Exec(`INSERT INTO credit_limits(tenant_id,limit_amount,currency) VALUES($1,'2','BRL')`, tenant); err != nil {
		t.Fatal(err)
	}
	if err = finance.ReserveExact(ctx, tenant, protocolID, "1", "BRL"); err != nil {
		t.Fatal(err)
	}
	lateCost := libra.EconomicEvent{ProtocolID: protocolID, TenantID: tenant, OperationID: idgen.New(), AttemptID: idgen.New(), Status: "SUCCEEDED", Kind: "SUCCEEDED", EvidenceID: idgen.New(), OccurredAt: now, EconomicSnapshot: economic}
	lateCostPayload, _ := json.Marshal(lateCost)
	if err = finance.ProcessEnvelope(ctx, "cost", queue.Envelope{EventID: idgen.New(), Type: "operation.observed", SchemaVersion: 1, Producer: "cometa", TenantID: tenant, ProtocolID: protocolID, OccurredAt: now, RecordedAt: now, Payload: lateCostPayload}); err != nil {
		t.Fatalf("custo tardio elegível: %v", err)
	}
	lateRevenue := libra.EconomicEvent{ProtocolID: protocolID, TenantID: tenant, OperationID: lateCost.OperationID, Status: "EXPIRED", Kind: "EXPIRED", EvidenceID: finalFact.EvidenceID, OccurredAt: now, ExternalState: "SUCCEEDED", SafeToRelease: true, EconomicSnapshot: economic}
	lateRevenuePayload, _ := json.Marshal(lateRevenue)
	if err = finance.ProcessEnvelope(ctx, "revenue", queue.Envelope{EventID: finalFact.EventID, Type: "protocol.finalized", SchemaVersion: 1, Producer: "orbita", TenantID: tenant, ProtocolID: protocolID, OccurredAt: finalFact.OccurredAt, RecordedAt: finalFact.OccurredAt, Payload: lateRevenuePayload}); err != nil {
		t.Fatalf("final expirado na receita: %v", err)
	}

	costFacts, err := finance.Facts(ctx, tenant, "COST", 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	revenueFacts, err := finance.Facts(ctx, tenant, "REVENUE", 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(costFacts) != 1 || costFacts[0].Amount != "0.20000000" || len(revenueFacts) != 0 {
		t.Fatalf("incidência tardia incorreta: custos=%+v receitas=%+v", costFacts, revenueFacts)
	}
	if _, err = finance.Dispute(ctx, tenant, costFacts[0].ID, "0.2", "resultado tardio fora do SLA", finalFact.EvidenceID, "auditor-late"); err != nil {
		t.Fatalf("contestação do custo tardio: %v", err)
	}
	var disputes, ledgerEntries int
	if err = financeDB.QueryRowContext(ctx, "SELECT count(*) FROM finance_disputes WHERE tenant_id=$1", tenant).Scan(&disputes); err != nil {
		t.Fatal(err)
	}
	if err = financeDB.QueryRowContext(ctx, "SELECT count(*) FROM ledger_entries WHERE tenant_id=$1", tenant).Scan(&ledgerEntries); err != nil {
		t.Fatal(err)
	}
	if disputes != 1 || ledgerEntries != 2 {
		t.Fatalf("contestação/razão financeira inesperadas: disputes=%d ledger_entries=%d", disputes, ledgerEntries)
	}
	t.Logf("resultado tardio custoso: GET/webhook preservaram EXPIRED/SLA_EXCEEDED, Libra registrou somente custo de 0.2 BRL e uma contestação, sem receita de sucesso")
}

func TestFinalizeCommitCrossingDeadlineDoesNotEscapeAsSuccess(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	tenant := "commit-crosses-deadline-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "commit-crosses-deadline-cell", IdempotencyKey: "commit-crosses-deadline-key", RequestHash: "commit-crosses-deadline-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(350 * time.Millisecond)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"fixture":true}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	})

	lockTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lockTx.Rollback()
	if err = lockTx.QueryRowContext(ctx, "SELECT protocol_id FROM protocols WHERE protocol_id=$1 FOR UPDATE", p.ProtocolID).Scan(new(string)); err != nil {
		t.Fatal(err)
	}

	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	result := make(chan struct {
		applied bool
		err     error
	}, 1)
	go func() {
		applied, finalizeErr := finalizer.Finalize(ctx, "commit-crosses-deadline", tenant, p.ProtocolID, 0, StatusSucceeded, FinalBody{Result: map[string]any{"provider": "started-before-deadline"}}, "")
		result <- struct {
			applied bool
			err     error
		}{applied: applied, err: finalizeErr}
	}()

	time.Sleep(600 * time.Millisecond)
	if err = lockTx.Commit(); err != nil {
		t.Fatal(err)
	}
	completed := <-result
	if completed.err != nil || !completed.applied {
		t.Fatalf("finalização bloqueada pelo commit: applied=%v err=%v", completed.applied, completed.err)
	}
	got, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusExpired || got.TerminalReason.String != "REJECTED_LATE_SLA" || strings.Contains(string(got.FinalBody), "started-before-deadline") {
		t.Fatalf("commit após deadline escapou como sucesso: status=%s reason=%s body=%s", got.Status, got.TerminalReason.String, got.FinalBody)
	}
	var facts int
	if err = db.QueryRowContext(ctx, "SELECT count(*) FROM outbox WHERE aggregate_id=$1 AND event_type='protocol.finalized'", p.ProtocolID).Scan(&facts); err != nil {
		t.Fatal(err)
	}
	if facts != 1 {
		t.Fatalf("commit após deadline publicou %d fatos terminais, esperado 1", facts)
	}
	t.Logf("PostgreSQL: lock iniciado antes do deadline e liberado depois dele produziu somente EXPIRED/REJECTED_LATE_SLA, com um fato terminal")
}

func TestFinalizeAndExpiryRaceProducesOneTerminalFact(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	tenant := "finalize-expiry-race-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "finalize-expiry-race-cell", IdempotencyKey: "finalize-expiry-race-key", RequestHash: "finalize-expiry-race-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"fixture":true}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	})

	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	start := make(chan struct{})
	results := make(chan error, 2)
	go func() {
		<-start
		_, err := finalizer.Finalize(ctx, "success-race", tenant, p.ProtocolID, 0, StatusSucceeded, FinalBody{Result: map[string]any{"winner": "success"}}, "")
		results <- err
	}()
	go func() {
		<-start
		_, err := finalizer.Finalize(ctx, "expiry-race", tenant, p.ProtocolID, 0, StatusExpired, FinalBody{ErrorCode: "SLA_EXCEEDED"}, "SLA_EXCEEDED")
		results <- err
	}()
	close(start)
	if first, second := <-results, <-results; first != nil || second != nil {
		t.Fatalf("corrida retornou erro: first=%v second=%v", first, second)
	}

	got, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	if !IsTerminal(got.Status) || got.ResultVersion != 1 || got.Version != 1 {
		t.Fatalf("corrida deixou estado não terminal/inconsistente: status=%s result_version=%d version=%d", got.Status, got.ResultVersion, got.Version)
	}
	var facts int
	if err = db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1 AND event_type='protocol.finalized'", p.ProtocolID).Scan(&facts); err != nil {
		t.Fatal(err)
	}
	if facts != 1 {
		t.Fatalf("corrida publicou %d fatos terminais, esperado 1", facts)
	}
	t.Logf("PostgreSQL: finalização e expiração concorrentes convergiram para uma única transição %s, result_version=1 e um protocol.finalized", got.Status)
}

func TestFinalRepresentationCannotBeReplacedAfterTerminalCommit(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	tenant := "final-representation-immutable-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "final-representation-cell", IdempotencyKey: "final-representation-key", RequestHash: "final-representation-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"contract_version":"v1"}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	})
	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	firstBody := FinalBody{Result: map[string]any{"contract_version": "v1", "marker": "historical-bytes"}}
	if applied, err := finalizer.Finalize(ctx, "immutable-v1", tenant, p.ProtocolID, 0, StatusSucceeded, firstBody, ""); err != nil || !applied {
		t.Fatalf("finalização inicial: applied=%v err=%v", applied, err)
	}
	initial, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	if applied, err := finalizer.Finalize(ctx, "immutable-v2", tenant, p.ProtocolID, initial.Version, StatusFailed, FinalBody{Result: map[string]any{"contract_version": "v2", "marker": "replacement"}}, "CONTRACT_UPDATED"); err != nil || applied {
		t.Fatalf("finalização posterior substituiu histórico: applied=%v err=%v", applied, err)
	}
	current, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != StatusSucceeded || current.ResultVersion != 1 || string(current.FinalRepresentation) != string(initial.FinalRepresentation) || strings.Contains(string(current.FinalRepresentation), "replacement") {
		t.Fatalf("representação histórica mudou após segunda finalização: status=%s version=%d body=%s", current.Status, current.ResultVersion, current.FinalRepresentation)
	}
	var facts int
	if err = db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1 AND event_type='protocol.finalized'", p.ProtocolID).Scan(&facts); err != nil || facts != 1 {
		t.Fatalf("fatos finais após tentativa de substituição=%d err=%v", facts, err)
	}
	t.Log("PostgreSQL: representação final v1 permaneceu imutável diante de tentativa v2; status, result_version, bytes e protocol.finalized foram preservados")
}

func TestFinalizeRejectsSuccessAfterProviderSLABreach(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC().Add(-2 * time.Second)
	tenant := "provider-sla-tenant-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "provider-sla-cell", IdempotencyKey: "provider-sla-key", RequestHash: "provider-sla-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: time.Now().UTC().Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"target":{"data":{"provider_sla_seconds":1,"provider_sla_policy":"REJECT_LATE"}}}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM protocols WHERE protocol_id=$1", p.ProtocolID)
	}()

	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if applied, err := finalizer.Finalize(ctx, "trace", tenant, p.ProtocolID, 0, StatusSucceeded, FinalBody{Result: map[string]any{"provider": "late"}}, ""); err != nil || !applied {
		t.Fatalf("late provider finalization: applied=%v err=%v", applied, err)
	}
	got, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusFailed || got.TerminalReason.String != "REJECTED_PROVIDER_SLA" {
		t.Fatalf("late provider result became status=%s reason=%s", got.Status, got.TerminalReason.String)
	}
	var body FinalBody
	if err := json.Unmarshal(got.FinalBody, &body); err != nil {
		t.Fatal(err)
	}
	if body.ErrorCode != "PROVIDER_SLA_EXCEEDED" || body.Status != string(StatusFailed) {
		t.Fatalf("late provider result was not rejected: %+v", body)
	}
	t.Log("finalização após o prazo do provedor sob REJECT_LATE foi materializada como FAILED, preservando o prazo do cliente e sem sucesso fictício")
}

func TestFinalizeMonitorOnlyPreservesClientSuccessAndReportsProviderBreach(t *testing.T) {
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
	now := time.Now().UTC().Add(-2 * time.Second)
	tenant := "provider-monitor-tenant-" + idgen.New()
	p := Protocol{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "provider-monitor-cell", IdempotencyKey: "provider-monitor-key", RequestHash: "provider-monitor-hash", RequestBody: json.RawMessage(`{"input":"fixture"}`), Mode: "ASYNC", DispatchMode: "QUEUED", CommandID: idgen.New(), Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: time.Now().UTC().Add(time.Minute)}
	c := dispatch.Command{ProtocolID: p.ProtocolID, TenantID: p.TenantID, ApplicationID: p.ApplicationID, CellID: p.CellID, CommandID: p.CommandID, DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"target":{"data":{"provider_sla_seconds":1,"provider_sla_policy":"MONITOR_ONLY"}}}`), EconomicSnapshot: json.RawMessage(`{}`), AcceptedAt: now, StepDeadline: p.ClientDeadlineAt}
	if _, created, err := store.Admit(ctx, p, c, "fixture", false); err != nil || !created {
		t.Fatalf("admit: created=%v err=%v", created, err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM protocol_access_audit WHERE requested_tenant=$1", tenant)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM command_intents WHERE protocol_id=$1", p.ProtocolID)
		db.Exec("DELETE FROM protocols WHERE tenant_id=$1", tenant)
	})

	finalizer := NewFinalizer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if applied, err := finalizer.Finalize(ctx, "trace", tenant, p.ProtocolID, 0, StatusSucceeded, FinalBody{Result: map[string]any{"provider": "late-but-accepted"}}, ""); err != nil || !applied {
		t.Fatalf("monitor-only finalization: applied=%v err=%v", applied, err)
	}
	got, err := store.Get(ctx, tenant, p.ProtocolID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusSucceeded {
		t.Fatalf("monitor-only provider breach changed client result to %s", got.Status)
	}
	principal := auth.Principal{Subject: "sla-reader", TenantID: tenant, MFA: true, Roles: []string{"hub_protocol_reader"}, Scopes: []string{"protocols:read"}, ExpiresAt: time.Now().Add(time.Hour)}
	mux := http.NewServeMux()
	NewHandlers(store, nil, nil, nil, nil, nil).RegisterAdmin(mux)
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/sla-reports/"+p.ProtocolID+"?tenant_id="+tenant, nil).WithContext(auth.WithPrincipal(ctx, principal))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("SLA report status=%d body=%s", response.Code, response.Body.String())
	}
	var report struct {
		Status              string `json:"status"`
		ProviderSLAOutcome  string `json:"provider_sla_outcome"`
		ProviderSLABreached bool   `json:"provider_sla_breached"`
		ClientSLABreached   bool   `json:"client_sla_breached"`
		ProviderSLAPolicy   string `json:"provider_sla_policy"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != string(StatusSucceeded) || report.ProviderSLAPolicy != "MONITOR_ONLY" || report.ProviderSLAOutcome != "MONITOR_ONLY_BREACH" || !report.ProviderSLABreached || report.ClientSLABreached {
		t.Fatalf("monitor-only report lost the bilateral distinction: %+v", report)
	}
	exportRequest := httptest.NewRequest(http.MethodGet, "/admin/v1/sla-reports/export?tenant_id="+tenant+"&status=SUCCEEDED&limit=100", nil).WithContext(auth.WithPrincipal(ctx, principal))
	exportResponse := httptest.NewRecorder()
	mux.ServeHTTP(exportResponse, exportRequest)
	if exportResponse.Code != http.StatusOK {
		t.Fatalf("SLA export status=%d body=%s", exportResponse.Code, exportResponse.Body.String())
	}
	var exported struct {
		CSV     string `json:"csv"`
		Rows    int    `json:"rows"`
		Limited bool   `json:"limited"`
		MaxRows int    `json:"max_rows"`
	}
	if err := json.Unmarshal(exportResponse.Body.Bytes(), &exported); err != nil {
		t.Fatal(err)
	}
	if exported.Rows != 1 || !exported.Limited || exported.MaxRows != 100 || !strings.Contains(exported.CSV, p.ProtocolID) || !strings.Contains(exported.CSV, "provider_sla_outcome") {
		t.Fatalf("SLA export perdeu recorte ou cabeçalho: %+v", exported)
	}
	var exportAudits int
	if err := db.QueryRow("SELECT count(*) FROM protocol_access_audit WHERE subject=$1 AND requested_tenant=$2 AND resource='sla-reports' AND action='EXPORT'", principal.Subject, tenant).Scan(&exportAudits); err != nil || exportAudits != 1 {
		t.Fatalf("SLA exportação não foi auditada: count=%d err=%v", exportAudits, err)
	}
	t.Log("atraso do provedor sob MONITOR_ONLY preservou SUCCEEDED do cliente e o painel separou MONITOR_ONLY_BREACH de client_sla_breached")
}
