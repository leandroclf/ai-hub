package cometa

import (
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestPostgresProviderReceiptIsSeparateFromNormalizedResult(t *testing.T) {
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
	id := idgen.New()
	cmd := dispatch.Command{
		CommandID: id, ProtocolID: idgen.New(), TenantID: "receipt-raw-" + idgen.New(),
		ApplicationID: "app-receipt", CellID: "r2-cell-a", StepDeadline: time.Now().Add(time.Minute),
	}
	if _, owned, err := store.PrepareSubmission(context.Background(), cmd, "binding", "v1"); err != nil || !owned {
		t.Fatalf("prepare owned=%v err=%v", owned, err)
	}
	raw := []byte(`{"provider_request_id":"provider-raw-1","status":"SUCCEEDED","detail":"raw-only"}`)
	normalized := json.RawMessage(`{"marker":"normalized"}`)
	_, err = store.ConserveObservation(context.Background(), cmd, dispatch.Result{
		Kind: dispatch.FactSucceeded, ProviderRequestID: "provider-raw-1", ResponseBody: normalized, RawResponse: raw,
	}, "PROVIDER", idgen.New())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM provider_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	})
	var body []byte
	var hash, receiptBody string
	if err := db.QueryRow("SELECT body,body_sha256 FROM provider_receipts WHERE operation_id=$1", id).Scan(&body, &hash); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if string(body) != string(raw) || hash != hex.EncodeToString(sum[:]) {
		t.Fatalf("raw receipt was altered: body=%s hash=%s", body, hash)
	}
	if err := db.QueryRow("SELECT body::text FROM operation_receipts WHERE operation_id=$1", id).Scan(&receiptBody); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(receiptBody, "normalized") || strings.Contains(receiptBody, "raw-only") {
		t.Fatalf("normalized and raw receipts were not separated: %s", receiptBody)
	}
}
