package pulsar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

type webhookRetryRequest struct {
	body      []byte
	timestamp string
	eventID   string
	delivery  string
	signature string
}

// TestPostgresWebhookRetryPreservesFinalRepresentationAndHMAC qualifica
// R2-EXE-08-S01 em conjunto com o teste de resposta final da Orbita: duas
// tentativas da mesma entrega devem usar exatamente os bytes custodiados e
// recalcular o HMAC somente sobre esses bytes, sem projetar um novo corpo.
func TestPostgresWebhookRetryPreservesFinalRepresentationAndHMAC(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)

	ctx := context.Background()
	store := NewStore(db)
	cell := "representation-retry-" + idgen.New()
	tenant := "representation-retry-" + idgen.New()
	eventID := idgen.New()
	deliveryID := ""
	secret := "representation-retry-secret"
	body := []byte(`{"protocol_id":"01representation","status":"SUCCEEDED","result_version":1,"final_body":{"marker":"immutable-final"}}`)

	var mu sync.Mutex
	requests := make([]webhookRetryRequest, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Errorf("ler corpo do webhook: %v", readErr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		event := r.Header.Get("X-Hub-Event-Id")
		delivery := r.Header.Get("X-Hub-Delivery-Id")
		timestamp := r.Header.Get("X-Hub-Timestamp")
		prefix := timestamp + "." + event + "." + delivery + "."
		if !bytes.Equal(raw, body) || event != eventID || delivery == "" || r.Header.Get("X-Hub-Signature-256") != sign(secret, append([]byte(prefix), raw...)) {
			t.Errorf("corpo ou HMAC divergente: event=%q delivery=%q body=%q", event, delivery, raw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		requests = append(requests, webhookRetryRequest{body: append([]byte(nil), raw...), timestamp: timestamp, eventID: event, delivery: delivery, signature: r.Header.Get("X-Hub-Signature-256")})
		attempt := len(requests)
		mu.Unlock()
		if attempt == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	u := server.URL
	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("EGRESS_HTTP_ORIGINS", u)
	parsed, parseErr := serverURL(u)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	t.Setenv("EGRESS_PRIVATE_RULES", parsed.Host+"=127.0.0.1/32")

	destinationID := idgen.New()
	_, err = db.Exec(`
		INSERT INTO webhook_destination_versions(id,version,tenant_id,cell_id,url,secret_ref,secret_version,max_attempts,timeout_seconds,state,created_by)
		VALUES($1,1,$2,$3,$4,'representation-retry-secret','v1',2,2,'ACTIVE','fixture')
	`, destinationID, tenant, cell, u)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM webhook_attempts WHERE delivery_id IN (SELECT delivery_id FROM deliveries WHERE cell_id=$1)", cell)
		_, _ = db.Exec("DELETE FROM deliveries WHERE cell_id=$1", cell)
		_, _ = db.Exec("DELETE FROM webhook_destination_versions WHERE cell_id=$1", cell)
	}()

	fact := protocolFact{
		ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app-representation", CellID: cell,
		EventID: eventID, Representation: body, Status: "SUCCEEDED",
		WebhookDestinations: []dispatch.DestinationSnapshot{{ID: destinationID, Version: 1, URL: u, SecretRef: "representation-retry-secret", SecretVersion: "v1", MaxAttempts: 2, TimeoutSecond: 2}},
	}
	if err = store.ConserveFinal(ctx, idgen.New(), fact); err != nil {
		t.Fatal(err)
	}
	worker := NewDeliveryWorker(store, "", slog.New(slog.NewTextHandler(io.Discard, nil)))
	worker.resolver = fixtureSecrets{"representation-retry-secret": secret}
	first, err := store.ClaimDelivery(ctx, cell, "retry-worker-1")
	if err != nil {
		t.Fatal(err)
	}
	deliveryID = first.ID
	worker.attempt(ctx, first)
	if _, err = db.Exec("UPDATE deliveries SET next_attempt_at=clock_timestamp() WHERE delivery_id=$1", deliveryID); err != nil {
		t.Fatal(err)
	}
	second, err := store.ClaimDelivery(ctx, cell, "retry-worker-2")
	if err != nil {
		t.Fatal(err)
	}
	worker.attempt(ctx, second)

	mu.Lock()
	gotRequests := append([]webhookRetryRequest(nil), requests...)
	mu.Unlock()
	if len(gotRequests) != 2 {
		t.Fatalf("tentativas recebidas=%d, esperado 2", len(gotRequests))
	}
	if !bytes.Equal(gotRequests[0].body, gotRequests[1].body) || !bytes.Equal(gotRequests[0].body, body) {
		t.Fatal("reentrega não preservou os mesmos bytes da representação final")
	}
	if gotRequests[0].delivery != gotRequests[1].delivery || gotRequests[0].eventID != gotRequests[1].eventID {
		t.Fatal("reentrega perdeu a identidade da entrega/evento")
	}
	if gotRequests[0].signature == "" || gotRequests[1].signature == "" {
		t.Fatal("uma das tentativas não recebeu HMAC")
	}
	var state string
	var attempts int
	var hash string
	if err = db.QueryRow("SELECT state,attempts_count,body_sha256 FROM deliveries WHERE delivery_id=$1", deliveryID).Scan(&state, &attempts, &hash); err != nil {
		t.Fatal(err)
	}
	if state != "DELIVERED" || attempts != 2 {
		t.Fatalf("custódia final inesperada: state=%s attempts=%d", state, attempts)
	}
	if hash != sha256Hex(body) {
		t.Fatalf("hash custodiado=%s, esperado=%s", hash, sha256Hex(body))
	}
	t.Logf("PostgreSQL + HTTP: entrega %s repetiu os mesmos %d bytes em duas tentativas; HMAC e identidade verificaram ambas, estado final DELIVERED", deliveryID, len(body))
}

func serverURL(raw string) (*url.URL, error) {
	return url.Parse(raw)
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
