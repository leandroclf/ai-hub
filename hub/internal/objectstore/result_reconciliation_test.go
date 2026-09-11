package objectstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestPostgresResultReconciliationLinksOnlyTheOriginalObligation qualifica
// R2-DAD-02-S03: um objeto já presente no storage, mas ainda VALIDATING por
// falha de commit posterior, é validado localmente e só pode ser ligado à
// obrigação original.
func TestPostgresResultReconciliationLinksOnlyTheOriginalObligation(t *testing.T) {
	dsn, endpoint := os.Getenv("R2_CORE_TEST_DSN"), os.Getenv("R2_S3_TEST_ENDPOINT")
	if dsn == "" || endpoint == "" {
		t.Skip("requires isolated PostgreSQL and S3")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	class := fmt.Sprintf("SYNTHETIC-RECON-%d", time.Now().UnixNano())
	tenant := "result-reconciliation-" + class
	obligation := "protocol-obligation-" + class
	otherObligation := "other-obligation-" + class
	policy := `INSERT INTO object_retention_policies(class,region,purpose,retention_seconds,max_bytes,allowed_types,approved_by) VALUES($1,'fixture-local','RESULT',3600,67108864,'["application/json"]','fixture-test')`
	if _, err = db.Exec(policy, class); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM object_retention_pins WHERE tenant_id=$1", tenant)
		_, _ = db.Exec("DELETE FROM file_refs WHERE tenant_id=$1", tenant)
		_, _ = db.Exec("DELETE FROM object_retention_policies WHERE class=$1", class)
	})

	bucket := fmt.Sprintf("r2-result-recon-%d", time.Now().UnixNano())
	client, err := New(ctx, endpoint, "us-east-1", bucket)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.EnsureBucket(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.s3.DeleteBucket(context.Background(), &s3.DeleteBucketInput{Bucket: aws.String(bucket)})

	body := []byte(`{"result":"durable-after-commit-gap","status":"SUCCEEDED"}`)
	key := "results/reconciliation-" + class
	if err = client.Put(ctx, key, body, "application/json"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if metadata, headErr := client.Head(context.Background(), key, ""); headErr == nil {
			_ = client.DeleteVersion(context.Background(), key, metadata.Version)
		}
	}()
	hash := sha256.Sum256(body)
	fileID := uuid.NewString()
	_, err = db.Exec(`
		INSERT INTO file_refs(id,tenant_id,object_key,sha256,size_bytes,content_type,purpose,class,region,state,expires_at,retention_until,obligation_id)
		VALUES($1,$2,$3,'PENDING',$4,$5,'RESULT',$6,'fixture-local','VALIDATING',clock_timestamp()+interval '1 hour',clock_timestamp()+interval '1 hour',$7)
	`, fileID, tenant, key, len(body), "application/json", class, obligation)
	if err != nil {
		t.Fatal(err)
	}

	catalog := NewCatalog(db, client)
	reconciled, err := catalog.Reconcile(ctx, tenant)
	if err != nil || reconciled != 1 {
		t.Fatalf("reconciliação do objeto presente: reconciled=%d err=%v", reconciled, err)
	}
	var state, storedHash, version string
	if err = db.QueryRow("SELECT state,sha256,object_version FROM file_refs WHERE id=$1", fileID).Scan(&state, &storedHash, &version); err != nil {
		t.Fatal(err)
	}
	if state != "ORPHAN" || storedHash != hex.EncodeToString(hash[:]) || version == "" {
		t.Fatalf("objeto não ficou reconciliável: state=%s hash=%s version=%s", state, storedHash, version)
	}
	if _, err = catalog.LinkResult(ctx, tenant, fileID, otherObligation); !errors.Is(err, ErrIneligible) {
		t.Fatalf("vínculo com obrigação diferente foi aceito: %v", err)
	}
	if _, err = catalog.LinkResult(ctx, tenant, fileID, obligation); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT state FROM file_refs WHERE id=$1", fileID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	var pins int
	if err = db.QueryRow("SELECT count(*) FROM object_retention_pins WHERE file_id=$1 AND obligation_id=$2", fileID, obligation).Scan(&pins); err != nil {
		t.Fatal(err)
	}
	if state != "READY" || pins != 1 {
		t.Fatalf("vínculo correto não consolidou custódia: state=%s pins=%d", state, pins)
	}
	t.Logf("PostgreSQL + LocalStack: objeto presente após gap de commit foi validado como ORPHAN, obrigação alheia recusada e vínculo original promoveu READY com um pin")
}
