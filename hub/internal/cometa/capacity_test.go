package cometa

import (
	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/platform/pg"
	"ai-hub/hub/internal/providerauth"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func capacityFixture(t *testing.T) (*sql.DB, *CapacityController, CapacityPolicy) {
	t.Helper()
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(12)
	p := CapacityPolicy{Domain: "synthetic-capacity-" + idgen.New(), Version: "fixture-v1", EvidenceRef: "synthetic-test-only", ValidUntil: time.Now().Add(time.Hour), MaxConcurrent: 8, MinConcurrent: 5, ReconciliationReserve: 1, MaxPending: 8, RatePerWindow: 200, WindowMillis: 1000, LeaseMillis: 5000, StableMillis: 100, LatencyThresholdMillis: 100, TenantLimits: map[string]int{"a": 2, "b": 2}, TenantPendingLimits: map[string]int{"a": 4, "b": 4}, TenantRateLimits: map[string]int{"a": 90, "b": 90}}
	t.Cleanup(func() {
		db.Exec("DELETE FROM capacity_feedback WHERE domain_id=$1", p.Domain)
		db.Exec("DELETE FROM capacity_permits WHERE domain_id=$1", p.Domain)
		db.Exec("DELETE FROM capacity_domains WHERE domain_id=$1", p.Domain)
		db.Close()
	})
	return db, NewCapacityController(db), p
}
func TestPostgresCapacityAggregateReplicasCells(t *testing.T) {
	db, c, p := capacityFixture(t)
	ctx := context.Background()
	if err := c.InstallPolicy(ctx, p); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var permits []CapacityPermit
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			replica := NewCapacityController(db)
			permit, err := replica.Acquire(ctx, p.Domain, idgen.New(), "a", fmt.Sprintf("cell-%d", i%2), fmt.Sprintf("replica-%d", i%3), "SUBMIT")
			if err == nil {
				mu.Lock()
				permits = append(permits, permit)
				mu.Unlock()
			} else if !errors.Is(err, ErrCapacityDenied) {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if len(permits) != 2 {
		t.Fatalf("60 noisy A contenders acquired %d, want 2", len(permits))
	}
	for i := 0; i < 2; i++ {
		v, err := c.Acquire(ctx, p.Domain, idgen.New(), "b", "cell-b", "healthy-b", "SUBMIT")
		if err != nil {
			t.Fatalf("B reserved capacity unavailable: %v", err)
		}
		permits = append(permits, v)
	}
	s, err := c.State(ctx, p.Domain)
	if err != nil || s.TransportOpen != 4 || s.PendingExternal != 4 {
		t.Fatalf("aggregate %+v %v", s, err)
	}
	if err = pg.WithAuditedScopeTx(ctx, db, "test-fixture:expire-lease", func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, `UPDATE capacity_permits SET lease_until=clock_timestamp()-interval '1 second' WHERE domain_id=$1`, p.Domain)
		return execErr
	}); err != nil {
		t.Fatal(err)
	}
	if err = c.Validate(ctx, permits[0]); !errors.Is(err, ErrCapacityFence) {
		t.Fatalf("expired validation %v", err)
	}
	if _, err = c.Acquire(ctx, p.Domain, idgen.New(), "a", "new-cell", "new-owner", "SUBMIT"); !errors.Is(err, ErrCapacityDenied) {
		t.Fatalf("expired permit recycled: %v", err)
	}
	if err = c.CompleteTransport(ctx, permits[0], "SUCCESS", time.Millisecond, false, "stale-receipt"); !errors.Is(err, ErrCapacityFence) {
		t.Fatalf("stale owner released quota: %v", err)
	}
	if err = c.ResolvePending(ctx, p.Domain, permits[0].ID, permits[0].Epoch, "provider-terminal-reconciliation-fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Acquire(ctx, p.Domain, idgen.New(), "a", "new-cell", "new-owner", "SUBMIT"); err != nil {
		t.Fatalf("positive reconciliation did not release quota: %v", err)
	}
	altered := p
	altered.Version = "rotated-secret-must-not-reset-budget"
	if err = c.InstallPolicy(ctx, altered); !errors.Is(err, ErrCapacityPolicy) {
		t.Fatalf("policy reset %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = c.Acquire(cancelled, p.Domain, idgen.New(), "b", "cell", "owner", "STATUS"); err == nil {
		t.Fatal("authority unavailable granted permit")
	}
	t.Log("partial S03: 60 contenders; 3 replica identities/2 cells; A=2, B=2; no expiry recycling; stale owner fenced; positive reconciliation releases one; cancelled authority request denies")
}
func TestPostgresCapacityAdaptiveAndPending(t *testing.T) {
	db, c, p := capacityFixture(t)
	ctx := context.Background()
	if err := c.InstallPolicy(ctx, p); err != nil {
		t.Fatal(err)
	}
	acquire := func(action string) CapacityPermit {
		t.Helper()
		v, err := c.Acquire(ctx, p.Domain, idgen.New(), "a", "cell", "replica", action)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	v := acquire("SUBMIT")
	if err := c.CompleteTransport(ctx, v, "TIMEOUT", 200*time.Millisecond, true, "timeout-fixture"); err != nil {
		t.Fatal(err)
	}
	s, err := c.State(ctx, p.Domain)
	if err != nil || s.Limit != 5 || s.TransportOpen != 0 || s.PendingExternal != 1 || s.LastFeedback != "TIMEOUT" {
		t.Fatalf("AIMD/UNKNOWN %+v %v", s, err)
	}
	for expected := 6; expected <= 8; expected++ {
		if _, err = db.Exec(`UPDATE capacity_domains SET stable_since=clock_timestamp()-interval '1 second' WHERE domain_id=$1`, p.Domain); err != nil {
			t.Fatal(err)
		}
		v = acquire("STATUS")
		if err = c.CompleteTransport(ctx, v, "SUCCESS", time.Millisecond, false, "stable-fixture"); err != nil {
			t.Fatal(err)
		}
		s, err = c.State(ctx, p.Domain)
		if err != nil || s.Limit != expected {
			t.Fatalf("gradual recovery expected %d got %+v %v", expected, s, err)
		}
	}
	v = acquire("FETCH")
	if err = c.CompleteTransport(ctx, v, "SUCCESS", 250*time.Millisecond, false, "slow-success-fixture"); err != nil {
		t.Fatal(err)
	}
	s, _ = c.State(ctx, p.Domain)
	if s.Limit != 5 || s.LastFeedback != "LATENCY" {
		t.Fatalf("latency feedback %+v", s)
	}
	for i := 0; i < 3; i++ {
		v = acquire("SUBMIT")
		if err = c.CompleteTransport(ctx, v, "SUCCESS", time.Millisecond, true, "pending-fixture"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = c.Acquire(ctx, p.Domain, idgen.New(), "a", "cell", "replica", "SUBMIT"); !errors.Is(err, ErrCapacityDenied) {
		t.Fatalf("A pending quota not enforced: %v", err)
	}
	v = acquire("STATUS")
	if err = c.CompleteTransport(ctx, v, "SUCCESS", time.Millisecond, false, "reconciliation-floor-fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Acquire(ctx, p.Domain, idgen.New(), "b", "cell", "replica", "SUBMIT"); err != nil {
		t.Fatalf("A pending pressure consumed B reserve: %v", err)
	}
	t.Log("partial S01/S02: timeout/slow success reduce 8 to 5; stable evidence increases 5 to 6 to 7 to 8; UNKNOWN retains pending; STATUS and B eligible")
}
func TestPostgresCapacityRollingRateIsolation(t *testing.T) {
	_, c, p := capacityFixture(t)
	ctx := context.Background()
	p.RatePerWindow = 5
	p.TenantRateLimits = map[string]int{"a": 2, "b": 2}
	p.WindowMillis = 60000
	if err := c.InstallPolicy(ctx, p); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		v, err := c.Acquire(ctx, p.Domain, idgen.New(), "a", "cell", "owner", "SUBMIT")
		if err != nil {
			t.Fatal(err)
		}
		if err = c.CompleteTransport(ctx, v, "SUCCESS", time.Millisecond, false, "fixture"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := c.Acquire(ctx, p.Domain, idgen.New(), "a", "cell", "owner", "SUBMIT"); !errors.Is(err, ErrCapacityDenied) {
		t.Fatalf("A rate exceeded: %v", err)
	}
	for i := 0; i < 2; i++ {
		v, err := c.Acquire(ctx, p.Domain, idgen.New(), "b", "cell", "owner", "SUBMIT")
		if err != nil {
			t.Fatalf("A stole B rate budget %v", err)
		}
		if err = c.CompleteTransport(ctx, v, "SUCCESS", time.Millisecond, false, "fixture"); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("rolling 60-second rate: A denied after 2; B admits 2; global domain")
}

func TestPostgresCapacityPartitionDeniesStaleReplicaAndPreservesHealthyTenant(t *testing.T) {
	db, c, p := capacityFixture(t)
	ctx := context.Background()
	p.TenantLimits = map[string]int{"a": 2, "b": 2}
	p.TenantPendingLimits = map[string]int{"a": 2, "b": 2}
	p.TenantRateLimits = map[string]int{"a": 40, "b": 40}
	if err := c.InstallPolicy(ctx, p); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	permits := make([]CapacityPermit, 0, 2)
	for replica := 0; replica < 3; replica++ {
		for attempt := 0; attempt < 20; attempt++ {
			wg.Add(1)
			go func(replica, attempt int) {
				defer wg.Done()
				v, err := NewCapacityController(db).Acquire(ctx, p.Domain, idgen.New(), "a", fmt.Sprintf("cell-%d", replica), fmt.Sprintf("replica-%d", replica), "SUBMIT")
				if err == nil {
					mu.Lock()
					permits = append(permits, v)
					mu.Unlock()
				} else if !errors.Is(err, ErrCapacityDenied) {
					t.Errorf("ruído da réplica retornou erro inesperado: %v", err)
				}
			}(replica, attempt)
		}
	}
	wg.Wait()
	if len(permits) != 2 {
		t.Fatalf("partilha do domínio concedeu %d permissões A; esperado 2", len(permits))
	}
	partitioned, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := c.Acquire(partitioned, p.Domain, idgen.New(), "a", "partitioned-cell", "partitioned-replica", "SUBMIT"); !errors.Is(err, context.Canceled) {
		t.Fatalf("réplica sem coordenador não foi interrompida: %v", err)
	}
	var total int
	if err := pg.WithAuditedScopeTx(ctx, db, "test-fixture:count-open-permits", func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT count(*) FROM capacity_permits WHERE domain_id=$1 AND transport_open`, p.Domain).Scan(&total)
	}); err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Fatalf("tentativa particionada alterou permissões válidas: %d", total)
	}
	if _, err := c.Acquire(ctx, p.Domain, idgen.New(), "b", "healthy-cell", "healthy-replica", "SUBMIT"); err != nil {
		t.Fatalf("tenant B perdeu sua reserva durante ruído/partição de A: %v", err)
	}
	t.Log("três réplicas disputaram o mesmo domínio sem multiplicar quota; a réplica particionada não enviou e B continuou elegível com reserva própria")
}

// TestPostgresCapacityPermitStuckAfterSettlementFailureResolvesOnlyWithEvidence
// prova R6-OPE-04-S03: um permit cujo settlement (CompleteTransport) nunca
// completou e cujo lease já venceu não libera a vaga sozinho — só uma
// reconciliação administrativa autorizada, com evidência terminal, fecha a
// obrigação; sem ela, a vaga continua ocupada.
func TestPostgresCapacityPermitStuckAfterSettlementFailureResolvesOnlyWithEvidence(t *testing.T) {
	db, c, p := capacityFixture(t)
	ctx := context.Background()
	if err := c.InstallPolicy(ctx, p); err != nil {
		t.Fatal(err)
	}
	permit, err := c.Acquire(ctx, p.Domain, idgen.New(), "a", "cell", "owner-crash", "SUBMIT")
	if err != nil {
		t.Fatal(err)
	}
	// Settlement never happens (simulating a crash between the provider
	// response and CompleteTransport), and the lease expires.
	if _, err = db.Exec(`UPDATE capacity_permits SET lease_until=clock_timestamp()-interval '1 second' WHERE domain_id=$1 AND permit_id=$2`, p.Domain, permit.ID); err != nil {
		t.Fatal(err)
	}
	state, err := c.State(ctx, p.Domain)
	if err != nil || state.TransportOpen != 1 {
		t.Fatalf("stuck permit not counted as open: %+v %v", state, err)
	}
	// A blind TTL-based reap must never run: the lease passing time alone
	// never resolves the permit. State stays open indefinitely without
	// evidence.
	time.Sleep(50 * time.Millisecond)
	state, err = c.State(ctx, p.Domain)
	if err != nil || state.TransportOpen != 1 {
		t.Fatalf("permit was recycled blindly by TTL: %+v %v", state, err)
	}

	store := NewStore(db)
	exec := NewExecutor(store, atlasclient.New("http://127.0.0.1:1", time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{})
	exec.SetCapacityController(c)
	h := NewHandlers(exec, store)
	mux := http.NewServeMux()
	h.RegisterAdmin(mux)

	principal := auth.Principal{Subject: "operator-recovery", Scopes: []string{"integrations:write"}, MFA: true, ExpiresAt: time.Now().Add(time.Hour)}
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/capacity-domains/"+p.Domain+"/permits/"+permit.ID+"/resolve", nil)
	req.Body = io.NopCloser(strings.NewReader(`{"reason":"provider confirmed terminal SUCCEEDED via out-of-band support ticket"}`))
	req = req.WithContext(auth.WithPrincipal(ctx, principal))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("resolve request: code=%d body=%s", w.Code, w.Body.String())
	}

	state, err = c.State(ctx, p.Domain)
	if err != nil || state.TransportOpen != 0 {
		t.Fatalf("obligation not released after authorized evidence: %+v %v", state, err)
	}
	var evidenceRef string
	if err = pg.WithAuditedScopeTx(ctx, db, "test-fixture:read-resolved-permit", func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT COALESCE(evidence_ref,'') FROM capacity_permits WHERE domain_id=$1 AND permit_id=$2`, p.Domain, permit.ID).Scan(&evidenceRef)
	}); err != nil {
		t.Fatal(err)
	}
	if evidenceRef != "operator-recovery:provider confirmed terminal SUCCEEDED via out-of-band support ticket" {
		t.Fatalf("resolution not attributed to the authorizing actor: %q", evidenceRef)
	}

	// Acquiring a fresh permit for the same tenant proves the slot the stuck
	// permit occupied was genuinely freed, not merely marked in a way that
	// left the count stale.
	if _, err = c.Acquire(ctx, p.Domain, idgen.New(), "a", "cell", "owner-recovered", "SUBMIT"); err != nil {
		t.Fatalf("slot not actually released: %v", err)
	}
}
