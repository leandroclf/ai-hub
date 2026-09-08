package pulsar

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

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
	defer func() {
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
		_, err = db.Exec(`INSERT INTO webhook_destination_versions(id,version,tenant_id,cell_id,url,secret_ref,secret_version,max_attempts,timeout_seconds,state,created_by) VALUES($1,1,$2,$3,$4,$5,'v1',3,2,'ACTIVE','fixture')`, idgen.New(), tenant, cell, server.URL, ref)
		if err != nil {
			t.Fatal(err)
		}
		facts = append(facts, protocolFact{ProtocolID: idgen.New(), TenantID: tenant, CellID: cell, EventID: event, Representation: body, Status: "SUCCEEDED"})
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
	if _, err = s.ClaimDelivery(ctx, cell, "worker-c"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("leased delivery reclaimed: %v", err)
	}
	var workerLogs bytes.Buffer
	worker := NewDeliveryWorker(s, "", slog.New(slog.NewTextHandler(&workerLogs, nil)))
	worker.resolver = fixtureSecrets{"key-a": "synthetic-a", "key-b": "synthetic-b"}
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
	t.Log("two tenants share URL with distinct pinned keys; original bytes conserved across JSONB event processing; duplicate facts create two total deliveries; exclusive claims and actual HTTP204 receipts committed")
}
