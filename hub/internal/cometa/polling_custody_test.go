package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
)

func pollDB(t *testing.T) (*Store, dispatch.Command) {
	t.Helper()
	t.Setenv("ENVIRONMENT", "local")
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(12)
	id := idgen.New()
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "poll-synthetic", ApplicationID: "app", CellID: "poll-" + id, ProviderAccountID: "account", StepDeadline: time.Now().Add(2 * time.Minute), RetryDeadline: time.Now().Add(time.Minute), ConfigSnapshot: json.RawMessage(`{"account":{"data":{"polling":{"interval_seconds":1,"max_interval_seconds":8,"timeout_seconds":1,"jitter_percent":0,"max_attempts":100}}}}`)}
	t.Cleanup(func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM polling_schedule WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
		db.Close()
	})
	s := NewStore(db)
	if _, _, err = s.PrepareSubmission(context.Background(), cmd, "binding", "v1"); err != nil {
		t.Fatal(err)
	}
	return s, cmd
}
func acceptPoll(t *testing.T, s *Store, cmd dispatch.Command) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.ConserveAcceptance(ctx, cmd, "provider-correlation", true); err != nil {
		t.Fatal(err)
	}
	// Current and pre-integration callers both get the intended policy in this fixture.
	if _, err := s.db.Exec(`UPDATE polling_schedule SET next_run_at=clock_timestamp()-interval '1 second',timeout_seconds=1,interval_seconds=1,max_interval_seconds=8,jitter_percent=0 WHERE operation_id=$1`, cmd.CommandID); err != nil {
		t.Fatal(err)
	}
}
func TestPostgresPollingClaimsFenceAndAbsoluteDeadline(t *testing.T) {
	s, cmd := pollDB(t)
	acceptPoll(t, s, cmd)
	ctx := context.Background()
	var won atomic.Int32
	var mu sync.Mutex
	var claim PollClaim
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, err := s.ClaimPoll(ctx, cmd.CellID, fmt.Sprint(i))
			if errors.Is(err, sql.ErrNoRows) {
				return
			}
			if err != nil {
				t.Error(err)
				return
			}
			won.Add(1)
			mu.Lock()
			claim = c
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	if won.Load() != 1 {
		t.Fatalf("owners=%d", won.Load())
	}
	var prepared int
	if err := s.db.QueryRow(`SELECT count(*) FROM attempts WHERE operation_id=$1 AND attempt_type='STATUS' AND prepared_at IS NOT NULL AND sent_at IS NULL`, cmd.CommandID).Scan(&prepared); err != nil || prepared != 1 {
		t.Fatalf("prepared=%d %v", prepared, err)
	}
	if _, err := s.ClaimPoll(ctx, "other-cell", "other"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-cell claim %v", err)
	}
	if _, err := s.db.Exec(`UPDATE polling_schedule SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_id=$1`, cmd.CommandID); err != nil {
		t.Fatal(err)
	}
	newer, err := s.ClaimPoll(ctx, cmd.CellID, "takeover")
	if err != nil {
		t.Fatal(err)
	}
	if newer.Epoch <= claim.Epoch {
		t.Fatal("epoch did not advance")
	}
	final := dispatch.Result{Kind: dispatch.FactSucceeded, ProviderRequestID: "provider-correlation", ResponseBody: map[string]string{"result_marker": "provider"}}
	if err = s.CompletePoll(ctx, claim, final, 0); !errors.Is(err, ErrPollFence) {
		t.Fatalf("stale update %v", err)
	}
	var state string
	s.db.QueryRow(`SELECT state FROM operations WHERE operation_id=$1`, cmd.CommandID).Scan(&state)
	if state != "ACCEPTED_EXTERNAL" {
		t.Fatalf("stale changed state %s", state)
	}
	if err = s.CompletePoll(ctx, newer, dispatch.Result{Kind: dispatch.FactUnknown, ErrorCode: "poll_pending"}, 5*time.Second); err != nil {
		t.Fatal(err)
	}
	var interval int
	var original time.Time
	var delay float64
	if err = s.db.QueryRow(`SELECT interval_seconds,deadline_at,extract(epoch from next_run_at-clock_timestamp()) FROM polling_schedule WHERE operation_id=$1`, cmd.CommandID).Scan(&interval, &original, &delay); err != nil {
		t.Fatal(err)
	}
	if interval != 2 || delay < 4 || original.Sub(cmd.RetryDeadline) > time.Microsecond || cmd.RetryDeadline.Sub(original) > time.Microsecond {
		t.Fatalf("policy interval=%d delay=%f deadline=%s", interval, delay, original)
	}
	if _, err = s.db.Exec(`UPDATE polling_schedule SET next_run_at=clock_timestamp()-interval '1 second',deadline_at=clock_timestamp()+interval '500 milliseconds' WHERE operation_id=$1`, cmd.CommandID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ClaimPoll(ctx, cmd.CellID, "late"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("insufficient budget claim %v", err)
	}
	t.Log("24 contenders: one committed STATUS preparation; takeover fenced stale final while retaining receipt; backoff and Retry-After persisted; original deadline unchanged; insufficient budget and foreign cell denied")
}
func TestPostgresPollingCallbackConflictRetainsBoth(t *testing.T) {
	s, cmd := pollDB(t)
	acceptPoll(t, s, cmd)
	ctx := context.Background()
	claim, err := s.ClaimPoll(ctx, cmd.CellID, "poll")
	if err != nil {
		t.Fatal(err)
	}
	var before int
	s.db.QueryRow(`SELECT count(*) FROM outbox WHERE aggregate_id=$1`, cmd.CommandID).Scan(&before)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := s.CompletePoll(ctx, claim, dispatch.Result{Kind: dispatch.FactSucceeded, ResponseBody: map[string]string{"marker": "poll"}}, 0); err != nil {
			t.Error(err)
		}
	}()
	go func() {
		defer wg.Done()
		if _, err := s.ConserveObservation(ctx, cmd, dispatch.Result{Kind: dispatch.FactFailed, ResponseBody: map[string]string{"marker": "callback"}}, "CALLBACK_TEST"); err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
	var facts, receipts int
	s.db.QueryRow(`SELECT count(*) FROM outbox WHERE aggregate_id=$1`, cmd.CommandID).Scan(&facts)
	s.db.QueryRow(`SELECT count(*) FROM operation_receipts WHERE operation_id=$1`, cmd.CommandID).Scan(&receipts)
	if facts != before+1 || receipts != 3 {
		t.Fatalf("facts=%d before=%d receipts=%d", facts, before, receipts)
	}
	t.Log("Conflicting poll/callback conserved both observations and produced exactly one terminal outbox fact")
}

type pollFixtureVault struct{}

func (pollFixtureVault) Resolve(_ context.Context, ref, version string) (providerauth.Secret, error) {
	if ref != "dedicated-secret" || version != "v1" {
		return providerauth.Secret{}, errors.New("wrong binding")
	}
	return providerauth.Secret{Value: "synthetic-password", Version: "v1"}, nil
}

func TestPostgresPollingAuthenticatedHTTP(t *testing.T) {
	s, cmd := pollDB(t)
	ctx := context.Background()
	var calls atomic.Int32
	var revoked atomic.Bool
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		user, password, ok := r.BasicAuth()
		if !ok || user != "tenant-user" || password != "synthetic-password" || r.Method != "GET" || r.URL.Path != "/v1/operations/provider-correlation" {
			w.WriteHeader(401)
			return
		}
		fmt.Fprint(w, `{"provider_request_id":"provider-correlation","status":"SUCCEEDED","result":{"marker":"actual-provider"}}`)
	}))
	defer provider.Close()
	identity := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"fixture-workload","expires_in":60,"token_type":"Bearer"}`)
	}))
	defer identity.Close()
	catalog := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if revoked.Load() {
			w.WriteHeader(409)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fixture-workload" || r.URL.Path != "/v1/credentials/binding/binding/1" || r.URL.Query().Get("tenant_id") != cmd.TenantID {
			w.WriteHeader(403)
			return
		}
		json.NewEncoder(w).Encode(atlasclient.CredentialBinding{BindingID: "binding", SecretRef: "dedicated-secret", SecretVersion: "v1", TenantID: cmd.TenantID})
	}))
	defer catalog.Close()
	file := filepath.Join(t.TempDir(), "fixture-secret")
	if err := os.WriteFile(file, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_TOKEN_URL", identity.URL)
	t.Setenv("WORKLOAD_CLIENT_ID", "fixture")
	t.Setenv("WORKLOAD_CLIENT_SECRET_FILE", file)
	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("EGRESS_HTTP_ORIGINS", provider.URL)
	u, _ := url.Parse(provider.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", u.Host+"=127.0.0.1/32")
	account, _ := json.Marshal(map[string]any{"base_url": provider.URL, "auth_type": "BASIC", "auth_username": "tenant-user", "polling": PollPolicy{1, 8, 1, 0, 100}})
	snapshot := atlas.OfferSnapshot{Account: atlas.Resource{ID: "account", Data: account}, Binding: atlas.Resource{ID: "binding", Version: 1, Data: json.RawMessage(`{"secret_version":"v1"}`)}, Target: atlas.Resource{Data: json.RawMessage(`{"adapter_id":"synthetic-provider","output_schema":{"type":"object"}}`)}}
	cmd.ConfigSnapshot, _ = json.Marshal(snapshot)
	raw, _ := json.Marshal(cmd)
	if _, err := s.db.Exec(`UPDATE operations SET command=$2 WHERE operation_id=$1`, cmd.CommandID, raw); err != nil {
		t.Fatal(err)
	}
	acceptPoll(t, s, cmd)
	claim, err := s.ClaimPoll(ctx, cmd.CellID, "http")
	if err != nil {
		t.Fatal(err)
	}
	exec := NewExecutor(s, atlasclient.New(catalog.URL, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{Resolver: pollFixtureVault{}})
	result, _ := exec.requestPoll(ctx, claim)
	if result.Kind != dispatch.FactSucceeded || calls.Load() != 1 {
		t.Fatalf("HTTP observation=%+v calls=%d", result, calls.Load())
	}
	revoked.Store(true)
	denied, _ := exec.requestPoll(ctx, claim)
	if denied.Kind != dispatch.FactUnknown || denied.ErrorCode != "poll_binding_unavailable" || calls.Load() != 1 {
		t.Fatalf("revocation did not stop request %+v", denied)
	}
	if err := s.CompletePoll(ctx, claim, result, 0); err != nil {
		t.Fatal(err)
	}
	t.Log("Actual local HTTP GET used dedicated decrypted Basic secret, workload-authenticated current binding and pinned egress; revoked binding prevented next provider call; terminal receipt persisted")
}
