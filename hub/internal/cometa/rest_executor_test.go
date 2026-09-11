package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
	_ "github.com/lib/pq"
)

// TestPostgresExecutorUsesQualifiedRESTAdapter qualifica R2-INT-01-S01:
// o executor usa o contrato REST publicado (método/path/corpo) e persiste o
// resultado sem chamar o endpoint do provider-sim.
func TestPostgresExecutorUsesQualifiedRESTAdapter(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	t.Setenv("CELL_ID", "r2-cell-a")
	t.Setenv("ENVIRONMENT", "local")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	store := NewStore(db)
	ctx := context.Background()
	operationID := idgen.New()
	tenant := "rest-executor-" + idgen.New()
	providerAccountID := "rest-account-" + idgen.New()
	bindingID := "rest-binding-" + idgen.New()

	var method, path string
	var received restJSONSubmitRequest
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("corpo REST inválido: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"provider_request_id":"rest-external-42","status":"SUCCEEDED","detail":"independent-rest"}`)
	}))
	defer provider.Close()
	providerURL, err := url.Parse(provider.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("EGRESS_HTTP_ORIGINS", provider.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", providerURL.Host+"=127.0.0.1/32")

	catalog := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "rest-fixture-token", "expires_in": 60, "token_type": "Bearer"})
			return
		}
		if r.URL.Path != "/v1/credentials/binding/"+bindingID+"/1" || r.URL.Query().Get("tenant_id") != tenant {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(atlasclient.CredentialBinding{
			BindingID: bindingID, TenantID: tenant, ProviderAccountID: providerAccountID,
			SecretRef: "vault://rest/fixture", SecretVersion: "v1", CredentialMode: "SHARED_HUB", State: "ACTIVE",
		})
	}))
	defer catalog.Close()
	secretPath := t.TempDir() + "/workload-secret"
	if err = os.WriteFile(secretPath, []byte("rest-workload-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_TOKEN_URL", catalog.URL+"/token")
	t.Setenv("WORKLOAD_CLIENT_ID", "rest-cometa")
	t.Setenv("WORKLOAD_CLIENT_SECRET_FILE", secretPath)

	account, _ := json.Marshal(atlasclient.ProviderAccount{ProviderAccountID: providerAccountID, ProviderID: "independent-rest", BaseURL: provider.URL, ProviderMode: "sync", AuthType: "NONE"})
	target, _ := json.Marshal(atlas.CatalogData{AdapterID: "rest-json-v1", AdapterContract: &atlas.AdapterContract{SubmitPath: "/analise", StatusPath: "/consulta/{id}"}, OutputSchema: json.RawMessage(`{"type":"object"}`)})
	snapshot := atlas.OfferSnapshot{
		Account:       atlas.Resource{ID: providerAccountID, Data: account},
		Binding:       atlas.Resource{ID: bindingID, Version: 1, Data: json.RawMessage(`{"secret_version":"v1"}`)},
		Target:        atlas.Resource{Kind: "services", ID: "rest-service", Version: 1, Data: target},
		SelectedRoute: atlas.Route{ProviderAccountID: providerAccountID, BindingID: bindingID},
	}
	config, _ := json.Marshal(snapshot)
	cmd := dispatch.Command{
		CommandID: operationID, ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app-rest",
		CellID: "r2-cell-a", ProviderAccountID: providerAccountID, DispatchMode: dispatch.DispatchDirect,
		ConfigSnapshot: config, RequestBody: map[string]any{"marker": "from-rest-executor"},
		StepDeadline: time.Now().Add(time.Minute),
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM provider_receipts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM attempts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM operations WHERE operation_id=$1", operationID)
	})

	executor := NewExecutor(store, atlasclient.New(catalog.URL, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{})
	result := executor.Execute(ctx, cmd)
	if result.Kind != dispatch.FactSucceeded || result.ProviderRequestID != "rest-external-42" || result.ResponseBody == nil {
		t.Fatalf("executor não custodiou o contrato REST: %+v", result)
	}
	if method != http.MethodPost || path != "/analise" || received.Mode != "sync" || received.Input.(map[string]any)["marker"] != "from-rest-executor" {
		t.Fatalf("executor não respeitou método/path/corpo REST: method=%s path=%s payload=%+v", method, path, received)
	}
	var state, providerRequestID string
	if err = db.QueryRowContext(ctx, "SELECT state,provider_request_id FROM operations WHERE operation_id=$1", operationID).Scan(&state, &providerRequestID); err != nil {
		t.Fatal(err)
	}
	if state != "SUCCEEDED" || providerRequestID != "rest-external-42" {
		t.Fatalf("custódia REST inesperada: state=%s provider_request_id=%s", state, providerRequestID)
	}
	t.Logf("PostgreSQL + HTTP REST independente: executor chamou POST /analise, persistiu SUCCEEDED e provider_request_id=%s sem provider-sim", providerRequestID)
}
