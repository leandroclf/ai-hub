package pulsar

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"ai-hub/hub/internal/cometa"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
	_ "github.com/lib/pq"
)

type fixtureSecrets map[string]string

func (f fixtureSecrets) Resolve(_ context.Context, ref, version string) (providerauth.Secret, error) {
	value, ok := f[ref]
	if !ok || version != "v1" {
		return providerauth.Secret{}, errors.New("missing fixture key")
	}
	return providerauth.Secret{Value: value, Version: version}, nil
}

func TestPostgresVersionedWebhookCustody(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	s := NewStore(db)
	ctx := context.Background()
	cell := "delivery-test-" + idgen.New()
	capacityDomain := "delivery-test-capacity-" + idgen.New()
	defer func() {
		db.Exec("DELETE FROM capacity_feedback WHERE domain_id=$1", capacityDomain)
		db.Exec("DELETE FROM capacity_permits WHERE domain_id=$1", capacityDomain)
		db.Exec("DELETE FROM capacity_domains WHERE domain_id=$1", capacityDomain)
		db.Exec("DELETE FROM webhook_attempts WHERE delivery_id IN (SELECT delivery_id FROM deliveries WHERE cell_id=$1)", cell)
		db.Exec("DELETE FROM deliveries WHERE cell_id=$1", cell)
		db.Exec("DELETE FROM webhook_destination_versions WHERE cell_id=$1", cell)
	}()
	body := []byte("{\n  \"result\": \"PROVIDER_ACTUAL\", \"status\": \"SUCCEEDED\"\n}")
	eventA, eventB := idgen.New(), idgen.New()
	keys := map[string]string{eventA: "synthetic-a", eventB: "synthetic-b"}
	var mu sync.Mutex
	received := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		event := r.Header.Get("X-Hub-Event-Id")
		prefix := r.Header.Get("X-Hub-Timestamp") + "." + event + "." + r.Header.Get("X-Hub-Delivery-Id") + "."
		if !bytes.Equal(b, body) || r.Header.Get("X-Hub-Signature-256") != sign(keys[event], append([]byte(prefix), b...)) {
			t.Error("body or tenant signing key differs")
			w.WriteHeader(400)
			return
		}
		mu.Lock()
		received[event] = true
		mu.Unlock()
		w.WriteHeader(204)
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("EGRESS_HTTP_ORIGINS", server.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", u.Host+"=127.0.0.1/32")
	var facts []protocolFact
	for index, event := range []string{eventA, eventB} {
		tenant := cell + string(rune('a'+index))
		ref := []string{"key-a", "key-b"}[index]
		destinationID := idgen.New()
		_, err = db.Exec(`INSERT INTO webhook_destination_versions(id,version,tenant_id,cell_id,url,secret_ref,secret_version,max_attempts,timeout_seconds,state,created_by) VALUES($1,1,$2,$3,$4,$5,'v1',3,2,'ACTIVE','fixture')`, destinationID, tenant, cell, server.URL, ref)
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 {
			if _, err = db.Exec(`INSERT INTO webhook_destination_versions(id,version,tenant_id,cell_id,url,secret_ref,secret_version,max_attempts,timeout_seconds,state,created_by) VALUES($1,2,$2,$3,$4,$5,'v1',3,2,'ACTIVE','fixture')`, destinationID, tenant, cell, server.URL+"/new", ref); err != nil {
				t.Fatal(err)
			}
		}
		facts = append(facts, protocolFact{ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app-a", CellID: cell, EventID: event, Representation: body, Status: "SUCCEEDED", WebhookDestinations: []dispatch.DestinationSnapshot{{ID: destinationID, Version: 1, URL: server.URL, SecretRef: ref, SecretVersion: "v1", MaxAttempts: 3, TimeoutSecond: 2}}})
		defer db.Exec("DELETE FROM inbox WHERE consumer='pulsar' AND event_id=$1", event)
	}
	for _, fact := range facts {
		for i := 0; i < 2; i++ {
			if err = s.ConserveFinal(ctx, fact.EventID, fact); err != nil {
				t.Fatal(err)
			}
		}
	}
	var count int
	if err = db.QueryRow("SELECT count(*) FROM deliveries WHERE cell_id=$1", cell).Scan(&count); err != nil || count != 2 {
		t.Fatalf("deliveries=%d %v", count, err)
	}
	first, err := s.ClaimDelivery(ctx, cell, "worker-a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.ClaimDelivery(ctx, cell, "worker-b")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("same delivery claimed twice")
	}
	if first.DestinationVersion != 1 || second.DestinationVersion != 1 {
		t.Fatalf("delivery ignored acceptance snapshot and selected versions %d/%d", first.DestinationVersion, second.DestinationVersion)
	}
	if _, err = s.ClaimDelivery(ctx, cell, "worker-c"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("leased delivery reclaimed: %v", err)
	}
	var workerLogs bytes.Buffer
	worker := NewDeliveryWorker(s, "", slog.New(slog.NewTextHandler(&workerLogs, nil)))
	worker.resolver = fixtureSecrets{"key-a": "synthetic-a", "key-b": "synthetic-b"}
	worker.capacity = cometa.NewCapacityController(db)
	worker.capacityDomain = capacityDomain
	if err = worker.capacity.InstallPolicy(ctx, cometa.CapacityPolicy{
		Domain:                 capacityDomain,
		Version:                "delivery-test-v1",
		EvidenceRef:            "delivery-test-capacity",
		ValidUntil:             time.Now().Add(time.Hour),
		MaxConcurrent:          8,
		MinConcurrent:          5,
		ReconciliationReserve:  1,
		MaxPending:             8,
		RatePerWindow:          200,
		WindowMillis:           1000,
		LeaseMillis:            5000,
		StableMillis:           100,
		LatencyThresholdMillis: 100,
		TenantLimits:           map[string]int{cell + "a": 1, cell + "b": 1},
		TenantPendingLimits:    map[string]int{cell + "a": 2, cell + "b": 2},
		TenantRateLimits:       map[string]int{cell + "a": 90, cell + "b": 90},
	}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, d := range []ClaimedDelivery{first, second} {
		wg.Add(1)
		go func(d ClaimedDelivery) { defer wg.Done(); worker.attempt(ctx, d) }(d)
	}
	wg.Wait()
	if len(received) != 2 {
		t.Fatalf("received events=%d", len(received))
	}
	if err = db.QueryRow("SELECT count(*) FROM deliveries WHERE cell_id=$1 AND state='DELIVERED'", cell).Scan(&count); err != nil || count != 2 {
		t.Fatalf("acknowledged receipts=%d %v; worker log=%s", count, err, workerLogs.String())
	}
	var open, pending int
	if err = db.QueryRow("SELECT count(*) FILTER (WHERE transport_open), count(*) FILTER (WHERE pending_external) FROM capacity_permits WHERE domain_id=$1", capacityDomain).Scan(&open, &pending); err != nil {
		t.Fatal(err)
	}
	if open != 0 || pending != 0 {
		t.Fatalf("webhook capacity permits not settled: open=%d pending=%d", open, pending)
	}
	t.Log("two tenants share URL with distinct pinned keys; original bytes conserved across JSONB event processing; duplicate facts create two total deliveries; exclusive claims and actual HTTP204 receipts committed")
}

func TestPostgresSlowWebhookTenantDoesNotBlockHealthyDelivery(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	ctx := context.Background()
	store := NewStore(db)
	cell := "delivery-slow-isolation-" + idgen.New()
	tenantSlow, tenantHealthy := cell+"-slow", cell+"-healthy"
	eventSlow, eventHealthy := idgen.New(), idgen.New()
	body := []byte(`{"status":"SUCCEEDED","result_version":1,"marker":"same-final"}`)
	var mu sync.Mutex
	received := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		event := r.Header.Get("X-Hub-Event-Id")
		_, _ = io.ReadAll(r.Body)
		mu.Lock()
		received[event]++
		mu.Unlock()
		if event == eventSlow {
			time.Sleep(1200 * time.Millisecond)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("EGRESS_HTTP_ORIGINS", server.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", parsedURL.Host+"=127.0.0.1/32")
	for _, fixture := range []struct {
		tenant, event, key string
	}{
		{tenantSlow, eventSlow, "slow-key"},
		{tenantHealthy, eventHealthy, "healthy-key"},
	} {
		destinationID := idgen.New()
		if _, err := db.Exec(`INSERT INTO webhook_destination_versions(id,version,tenant_id,cell_id,url,secret_ref,secret_version,max_attempts,timeout_seconds,state,created_by) VALUES($1,1,$2,$3,$4,$5,'v1',2,1,'ACTIVE','fixture')`, destinationID, fixture.tenant, cell, server.URL, fixture.key); err != nil {
			t.Fatal(err)
		}
		fact := protocolFact{ProtocolID: idgen.New(), TenantID: fixture.tenant, ApplicationID: "app-delivery", CellID: cell, EventID: fixture.event, Representation: body, Status: "SUCCEEDED", WebhookDestinations: []dispatch.DestinationSnapshot{{ID: destinationID, Version: 1, URL: server.URL, SecretRef: fixture.key, SecretVersion: "v1", MaxAttempts: 2, TimeoutSecond: 1}}}
		if err := store.ConserveFinal(ctx, fixture.event, fact); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM inbox WHERE consumer='pulsar' AND event_id IN ($1,$2)", eventSlow, eventHealthy)
		db.Exec("DELETE FROM webhook_attempts WHERE delivery_id IN (SELECT delivery_id FROM deliveries WHERE cell_id=$1)", cell)
		db.Exec("DELETE FROM deliveries WHERE cell_id=$1", cell)
		db.Exec("DELETE FROM webhook_destination_versions WHERE cell_id=$1", cell)
	})

	claimed := make([]ClaimedDelivery, 0, 2)
	for i := 0; i < 2; i++ {
		delivery, err := store.ClaimDelivery(ctx, cell, fmt.Sprintf("isolation-worker-%d", i))
		if err != nil {
			t.Fatal(err)
		}
		claimed = append(claimed, delivery)
	}
	worker := NewDeliveryWorker(store, "", slog.New(slog.NewTextHandler(io.Discard, nil)))
	worker.resolver = fixtureSecrets{"slow-key": "slow-secret", "healthy-key": "healthy-secret"}
	started := time.Now()
	var wg sync.WaitGroup
	for _, delivery := range claimed {
		wg.Add(1)
		go func(delivery ClaimedDelivery) {
			defer wg.Done()
			worker.attempt(ctx, delivery)
		}(delivery)
	}
	wg.Wait()
	if elapsed := time.Since(started); elapsed >= 2*time.Second {
		t.Fatalf("entrega saudável foi serializada pelo tenant lento: duração=%s", elapsed)
	}
	var slowState, healthyState string
	if err := db.QueryRow("SELECT state FROM deliveries WHERE event_id=$1", eventSlow).Scan(&slowState); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT state FROM deliveries WHERE event_id=$1", eventHealthy).Scan(&healthyState); err != nil {
		t.Fatal(err)
	}
	if slowState != "RETRY_SCHEDULED" || healthyState != "DELIVERED" {
		t.Fatalf("isolamento de entrega incorreto: lento=%s saudável=%s", slowState, healthyState)
	}
	if _, err := db.Exec("UPDATE deliveries SET next_attempt_at=clock_timestamp() WHERE event_id=$1", eventSlow); err != nil {
		t.Fatal(err)
	}
	second, err := store.ClaimDelivery(ctx, cell, "isolation-retry-worker")
	if err != nil || second.EventID != eventSlow {
		t.Fatalf("retry do tenant lento não foi agendado: event=%s err=%v", second.EventID, err)
	}
	worker.attempt(ctx, second)
	if err := db.QueryRow("SELECT state FROM deliveries WHERE event_id=$1", eventSlow).Scan(&slowState); err != nil {
		t.Fatal(err)
	}
	if slowState != "EXHAUSTED" {
		t.Fatalf("esgotamento do destino lento não ficou visível: %s", slowState)
	}
	mu.Lock()
	defer mu.Unlock()
	if received[eventHealthy] != 1 || received[eventSlow] != 2 {
		t.Fatalf("tentativas por tenant inesperadas: %+v", received)
	}
	t.Logf("destino lento ficou RETRY_SCHEDULED e depois EXHAUSTED; entrega saudável foi DELIVERED em paralelo; tentativas=%+v", received)
}
