package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
)

func TestReconciliationWorkerResolvesStatusWithoutReplay(t *testing.T) {
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
	db.SetMaxOpenConns(12)
	store := NewStore(db)
	ctx := context.Background()
	operationID, protocolID, requestID := idgen.New(), idgen.New(), idgen.New()
	cmd := dispatch.Command{CommandID: operationID, ProtocolID: protocolID, TenantID: "reconcile-synthetic", ApplicationID: "app-a", CellID: "reconcile-cell-" + operationID, ProviderAccountID: "account", StepDeadline: time.Now().Add(2 * time.Minute), RetryDeadline: time.Now().Add(time.Minute)}
	providerCalls := atomic.Int32{}
	submitCalls := atomic.Int32{}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			submitCalls.Add(1)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		providerCalls.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/v1/operations/provider-correlation" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = fmt.Fprint(w, `{"provider_request_id":"provider-correlation","status":"SUCCEEDED","detail":"reconciled"}`)
	}))
	defer provider.Close()
	providerURL, _ := url.Parse(provider.URL)
	t.Setenv("EGRESS_HTTP_ORIGINS", provider.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", providerURL.Host+"=127.0.0.1/32")
	identity := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"access_token":"fixture-workload","expires_in":60,"token_type":"Bearer"}`)
	}))
	defer identity.Close()
	catalog := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fixture-workload" || r.URL.Path != "/v1/credentials/binding/binding/1" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(atlasclient.CredentialBinding{BindingID: "binding", SecretRef: "fixture-secret", SecretVersion: "v1", TenantID: cmd.TenantID})
	}))
	defer catalog.Close()
	secretFile := t.TempDir() + "/workload-secret"
	if err = os.WriteFile(secretFile, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_TOKEN_URL", identity.URL)
	t.Setenv("WORKLOAD_CLIENT_ID", "fixture")
	t.Setenv("WORKLOAD_CLIENT_SECRET_FILE", secretFile)
	account, _ := json.Marshal(map[string]any{"base_url": provider.URL, "provider_mode": "async_poll", "auth_type": "NONE", "polling": PollPolicy{1, 8, 1, 0, 100}})
	snapshot := atlas.OfferSnapshot{Account: atlas.Resource{ID: "account", Data: account}, Binding: atlas.Resource{ID: "binding", Version: 1, Data: json.RawMessage(`{"secret_version":"v1"}`)}, Target: atlas.Resource{Data: json.RawMessage(`{"adapter_id":"synthetic-provider","output_schema":{"type":"object"}}`)}}
	cmd.ConfigSnapshot, _ = json.Marshal(snapshot)
	if _, _, err = store.PrepareSubmission(ctx, cmd, "binding", "v1"); err != nil {
		t.Fatal(err)
	}
	if _, err = store.ConserveAcceptance(ctx, cmd, "provider-correlation", true); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO protocol_reconciliation_requests(request_id,protocol_id,tenant_id,requested_by,reason) VALUES($1,$2,$3,'operator','external status required')`, requestID, protocolID, cmd.TenantID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM protocol_reconciliation_requests WHERE request_id=$1", requestID)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", operationID)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", operationID)
		db.Exec("DELETE FROM polling_schedule WHERE operation_id=$1", operationID)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", operationID)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", operationID)
	})

	exec := NewExecutor(store, atlasclient.New(catalog.URL, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{})
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go RunReconciliationWorker(workerCtx, store, exec, cmd.CellID, 5*time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil)))

	deadline := time.Now().Add(3 * time.Second)
	var state, operationState, evidenceID string
	for time.Now().Before(deadline) {
		if err = db.QueryRow(`SELECT state,COALESCE(evidence_id,'') FROM protocol_reconciliation_requests WHERE request_id=$1`, requestID).Scan(&state, &evidenceID); err != nil {
			t.Fatal(err)
		}
		if state == "RESOLVED" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if state != "RESOLVED" || evidenceID == "" {
		t.Fatalf("reconciliation did not resolve: state=%s evidence=%s", state, evidenceID)
	}
	if err = db.QueryRow(`SELECT state FROM operations WHERE operation_id=$1`, operationID).Scan(&operationState); err != nil {
		t.Fatal(err)
	}
	if operationState != "SUCCEEDED" || providerCalls.Load() != 1 || submitCalls.Load() != 0 {
		t.Fatalf("status reconciliation replayed or failed: operation=%s GET=%d POST=%d", operationState, providerCalls.Load(), submitCalls.Load())
	}
	var terminalFacts int
	if err = db.QueryRow(`SELECT count(*) FROM outbox WHERE aggregate_id=$1 AND event_type='operation.observed'`, operationID).Scan(&terminalFacts); err != nil {
		t.Fatal(err)
	}
	if terminalFacts != 2 {
		t.Fatalf("expected acceptance and one terminal fact, got %d", terminalFacts)
	}
	t.Logf("worker reconciled provider status with one GET, zero POST replay, durable evidence %s and fenced lease", evidenceID)
}
