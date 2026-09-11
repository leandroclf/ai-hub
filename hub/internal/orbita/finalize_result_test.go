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
	"ai-hub/hub/internal/objectstore"
	"ai-hub/hub/internal/platform/auth"
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
