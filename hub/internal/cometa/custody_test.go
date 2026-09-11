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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
	"ai-hub/hub/internal/providersim"
	_ "github.com/lib/pq"
)

func TestPostgresSubmissionAndObservationCustody(t *testing.T) {
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
	s := NewStore(db)
	ctx := context.Background()
	id := idgen.New()
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	}()
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "custody-synthetic", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "synthetic-account", RequestBody: map[string]any{"input": "different-from-output"}, StepDeadline: time.Now().Add(time.Minute), ConfigSnapshot: json.RawMessage(`{"version":1}`)}
	var winners atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 24; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, owned, e := s.PrepareSubmission(ctx, cmd, "binding-a", "v1")
			if e != nil {
				t.Error(e)
			}
			if owned {
				winners.Add(1)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("submit owners=%d", winners.Load())
	}
	var count int
	if e := db.QueryRow("SELECT count(*) FROM attempts WHERE operation_id=$1 AND prepared_at IS NOT NULL", id).Scan(&count); e != nil || count != 1 {
		t.Fatalf("attempt count %d %v", count, e)
	}
	response := dispatch.Result{Kind: dispatch.FactSucceeded, ResponseBody: map[string]string{"provider_marker": "actual-output"}, ProviderRequestID: "provider-id"}
	for n := 0; n < 12; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, e := s.ConserveObservation(ctx, cmd, response, "synthetic-race")
			if e != nil || !got.Durable || got.EvidenceID == "" {
				t.Errorf("custody %v %+v", e, got)
			}
		}()
	}
	wg.Wait()
	if e := db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1", id).Scan(&count); e != nil || count != 1 {
		t.Fatalf("terminal fact count %d %v", count, e)
	}
	got, e := s.DurableResult(ctx, cmd)
	if e != nil {
		t.Fatal(e)
	}
	body, _ := json.Marshal(got.ResponseBody)
	if string(body) != `{"provider_marker":"actual-output"}` {
		t.Fatalf("provider response lost: %s", body)
	}
	other := cmd
	other.TenantID = "other"
	if _, e = s.DurableResult(ctx, other); e == nil {
		t.Fatal("cross tenant result disclosed")
	}
	if _, _, e = s.PrepareSubmission(ctx, other, "binding-a", "v1"); e == nil {
		t.Fatal("conflicting command reused")
	}
	// UNKNOWN after an already committed final cannot overwrite the result.
	if _, e = s.ConserveObservation(ctx, cmd, dispatch.Result{Kind: dispatch.FactUnknown}, "late-timeout"); e != nil {
		t.Fatal(e)
	}
	again, e := s.DurableResult(ctx, cmd)
	if e != nil || again.EvidenceID != got.EvidenceID || again.Kind != dispatch.FactSucceeded {
		t.Fatalf("late observation changed final: %+v %v", again, e)
	}
	t.Log("24 submit contenders: one owner and prewritten attempt; 12 observations: one terminal fact; actual response retained; tenant conflict denied; late UNKNOWN preserved as receipt")
}

func TestCallbackCapabilityIsRandomAndStoredOnlyAsHash(t *testing.T) {
	first, err := newCallbackToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newCallbackToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) < 40 {
		t.Fatalf("callback capability is not an independent high-entropy value")
	}
	hash := callbackTokenHash(first)
	if hash == first || len(hash) != 64 || hash == callbackTokenHash(second) {
		t.Fatalf("callback capability hashing is invalid")
	}
}

func TestPostgresOrphanCapabilityHashDoesNotShadowValidCallback(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	operationID := idgen.New()
	body := []byte(`{"provider_request_id":"provider-correlation","status":"SUCCEEDED"}`)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM callback_inbox WHERE operation_id=$1", operationID) })
	if err = store.StoreOrphanCallback(context.Background(), operationID, "invalid-capability", body); err != nil {
		t.Fatal(err)
	}
	if err = store.StoreOrphanCallback(context.Background(), operationID, "invalid-capability", body); err != nil {
		t.Fatal(err)
	}
	if err = store.StoreOrphanCallback(context.Background(), operationID, "valid-capability", body); err != nil {
		t.Fatal(err)
	}
	var rows, occurrences int
	if err = db.QueryRow("SELECT count(*),coalesce(sum(occurrences),0) FROM callback_inbox WHERE operation_id=$1", operationID).Scan(&rows, &occurrences); err != nil {
		t.Fatal(err)
	}
	if rows != 2 || occurrences != 3 {
		t.Fatalf("orphan identity was collapsed incorrectly: rows=%d occurrences=%d", rows, occurrences)
	}
	t.Log("same callback bytes with different capability hashes remain distinct; exact duplicate increments occurrences")
}

func TestPostgresExecutorDropAfterEffectIsUnknownAndNotReexecuted(t *testing.T) {
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
	id := idgen.New()
	capacity := NewCapacityController(db)
	capacityPolicy := CapacityPolicy{Domain: "synthetic-executor-capacity-" + id, Version: "fixture-v1", EvidenceRef: "synthetic-executor-capacity", ValidUntil: time.Now().Add(time.Hour), MaxConcurrent: 8, MinConcurrent: 5, ReconciliationReserve: 1, MaxPending: 8, RatePerWindow: 200, WindowMillis: 1000, LeaseMillis: 5000, StableMillis: 100, LatencyThresholdMillis: 100, TenantLimits: map[string]int{"drop-synthetic": 2}, TenantPendingLimits: map[string]int{"drop-synthetic": 4}, TenantRateLimits: map[string]int{"drop-synthetic": 90}}
	if err := capacity.InstallPolicy(context.Background(), capacityPolicy); err != nil {
		t.Fatal(err)
	}
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "drop-synthetic", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "account", StepDeadline: time.Now().Add(time.Minute), RequestBody: map[string]any{"force_drop_after_effect": true}}
	t.Cleanup(func() {
		db.Exec("DELETE FROM capacity_feedback WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_permits WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_domains WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	})
	provider := providersim.NewServer()
	providerMux := http.NewServeMux()
	provider.Routes(providerMux)
	providerHTTP := httptest.NewServer(providerMux)
	defer providerHTTP.Close()
	providerURL, _ := url.Parse(providerHTTP.URL)
	t.Setenv("EGRESS_HTTP_ORIGINS", providerHTTP.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", providerURL.Host+"=127.0.0.1/32")
	catalog := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fixture-workload", "expires_in": 60, "token_type": "Bearer"})
			return
		}
		if r.Header.Get("Authorization") != "Bearer fixture-workload" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(atlasclient.CredentialBinding{BindingID: "binding", SecretRef: "fixture-secret", SecretVersion: "v1", CredentialMode: "SHARED_HUB"})
	}))
	defer catalog.Close()
	account, _ := json.Marshal(map[string]any{"base_url": providerHTTP.URL, "provider_mode": "sync", "auth_type": "NONE"})
	snapshot := atlas.OfferSnapshot{Account: atlas.Resource{ID: "account", Data: account}, Binding: atlas.Resource{ID: "binding", Version: 1, Data: json.RawMessage(`{"secret_version":"v1"}`)}, Target: atlas.Resource{Data: json.RawMessage(`{"adapter_id":"synthetic-provider","output_schema":{"type":"object","properties":{}}}`)}, SelectedRoute: atlas.Route{CapacityDomain: capacityPolicy.Domain}}
	cmd.ConfigSnapshot, _ = json.Marshal(snapshot)
	tokenFile := t.TempDir() + "/secret"
	if err := os.WriteFile(tokenFile, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_TOKEN_URL", catalog.URL+"/token")
	t.Setenv("WORKLOAD_CLIENT_ID", "fixture")
	t.Setenv("WORKLOAD_CLIENT_SECRET_FILE", tokenFile)
	exec := NewExecutor(store, atlasclient.New(catalog.URL, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{Resolver: dropFixtureVault{}})
	exec.SetCapacityController(capacity)
	first := exec.Execute(context.Background(), cmd)
	t.Logf("first execution: %+v", first)
	if first.Kind != dispatch.FactUnknown || !first.Durable {
		t.Fatalf("first execution should be durable UNKNOWN: %+v", first)
	}
	second := exec.Execute(context.Background(), cmd)
	if second.Kind != dispatch.FactUnknown || !second.Durable || second.EvidenceID != first.EvidenceID {
		t.Fatalf("recovery changed durable UNKNOWN: first=%+v second=%+v", first, second)
	}
	response, err := providerHTTP.Client().Get(providerHTTP.URL + "/__qualification/effects")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var effects struct {
		Effects int `json:"effects"`
	}
	if err = json.NewDecoder(response.Body).Decode(&effects); err != nil {
		t.Fatal(err)
	}
	if effects.Effects != 1 {
		t.Fatalf("external effect was reexecuted: %d", effects.Effects)
	}
	state, err := capacity.State(context.Background(), capacityPolicy.Domain)
	if err != nil || state.TransportOpen != 0 || state.PendingExternal != 1 {
		t.Fatalf("capacity did not retain uncertain effect: %+v %v", state, err)
	}
	var permits int
	if err = db.QueryRow("SELECT count(*) FROM capacity_permits WHERE domain_id=$1", capacityPolicy.Domain).Scan(&permits); err != nil || permits != 1 {
		t.Fatalf("capacity permit count=%d error=%v", permits, err)
	}
}

type dropFixtureVault struct{}

func (dropFixtureVault) Resolve(context.Context, string, string) (providerauth.Secret, error) {
	return providerauth.Secret{Value: "fixture", Version: "v1"}, nil
}

func TestPostgresPendingCustodyAtomic(t *testing.T) {
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
	s := NewStore(db)
	ctx := context.Background()
	id := idgen.New()
	defer func() {
		db.Exec("DELETE FROM polling_schedule WHERE operation_id=$1", id)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	}()
	// Missing durable polling budget is detected after tentative state/receipt
	// writes; rollback must leave neither a receipt nor a false acceptance.
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "pending-test", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "fixture", ConfigSnapshot: json.RawMessage(`{}`)}
	if _, owned, err := s.PrepareSubmission(ctx, cmd, "binding", "version"); err != nil || !owned {
		t.Fatalf("prepare: %v", err)
	}
	if r, err := s.ConserveAcceptance(ctx, cmd, "provider-pending", true); err == nil || r.Durable {
		t.Fatalf("false custody: %+v %v", r, err)
	}
	var state string
	var count int
	if err = db.QueryRow("SELECT state FROM operations WHERE operation_id=$1", id).Scan(&state); err != nil || state != "SUBMITTING" {
		t.Fatalf("partial state: %s %v", state, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM operation_receipts WHERE operation_id=$1", id).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial receipt: %d %v", count, err)
	}
	// Fixture adjustment of its persisted retry horizon, never a runtime backfill.
	cmd.RetryDeadline = time.Now().Add(time.Minute)
	raw, _ := json.Marshal(cmd)
	if _, err = db.Exec("UPDATE operations SET command=$2 WHERE operation_id=$1", id, raw); err != nil {
		t.Fatal(err)
	}
	attemptID := idgen.New()
	got, err := s.ConserveAcceptance(ctx, cmd, "provider-pending", true, attemptID)
	if err != nil || !got.Durable || got.EvidenceID == "" {
		t.Fatalf("acceptance: %+v %v", got, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM polling_schedule WHERE operation_id=$1 AND deadline_at=$2", id, cmd.RetryDeadline).Scan(&count); err != nil || count != 1 {
		t.Fatalf("polling obligation: %d %v", count, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1", id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("outbox: %d %v", count, err)
	}
	var incidence, payloadAttempt string
	if err = db.QueryRow("SELECT payload->>'economic_kind',payload->>'attempt_id' FROM outbox WHERE aggregate_id=$1", id).Scan(&incidence, &payloadAttempt); err != nil || incidence != "SUBMITTED" || payloadAttempt != attemptID {
		t.Fatalf("economic acceptance: incidence=%q attempt=%q err=%v", incidence, payloadAttempt, err)
	}
	replay, err := s.DurableResult(ctx, cmd)
	if err != nil || !replay.Durable || replay.ProviderRequestID != "provider-pending" {
		t.Fatalf("replay: %+v %v", replay, err)
	}
	t.Log("real PostgreSQL: failed polling budget rolls back state and receipt; valid acceptance conserves correlation, receipt, schedule, result and outbox")
}
