// populate_restore_fixture cria uma obrigação de arquivo real e verificável
// (file_refs + bytes em S3/localstack) no ambiente hub-local, para que
// restore-reconciliation.sh possa provar R6-OPE-02: um backup vazio nunca é
// aceito como "restore reconciliado", e toda versão de objeto referenciada
// precisa resolver para os bytes corretos após o restore.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"ai-hub/hub/internal/objectstore"

	_ "github.com/lib/pq"
)

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx := context.Background()
	dsn := envOr("CORE_DSN", "postgres://hub:hub@localhost:5432/hub_core?sslmode=disable")
	endpoint := envOr("S3_ENDPOINT", "http://localhost:4566")
	bucket := envOr("S3_BUCKET", "r2-custody")
	tenant := envOr("RESTORE_FIXTURE_TENANT", "restore-qualification")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open core db: %v", err)
	}
	defer db.Close()
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("ping core db: %v", err)
	}

	client, err := objectstore.New(ctx, endpoint, "us-east-1", bucket)
	if err != nil {
		log.Fatalf("s3 client: %v", err)
	}
	if err = client.EnsureBucket(ctx); err != nil {
		log.Fatalf("ensure bucket: %v", err)
	}

	class := "RESTORE-QUALIFICATION"
	if _, err = db.ExecContext(ctx, `INSERT INTO object_retention_policies(class,region,purpose,retention_seconds,max_bytes,allowed_types,approved_by) VALUES($1,'fixture-local','RESULT',3600,67108864,'["application/octet-stream"]','restore-qualification') ON CONFLICT DO NOTHING`, class); err != nil {
		log.Fatalf("install retention policy: %v", err)
	}

	catalog := objectstore.NewCatalog(db, client)
	// Two independently referenced objects/versions (R6-OPE-02-S01: "duas
	// versões referenciadas"), each a durable obligation with a pending step
	// still open (UNKNOWN-equivalent, not yet linked/READY) so restore must
	// preserve BOTH the bytes and the still-open obligation state.
	for i, content := range []string{
		"restore-qualification-object-v1-" + time.Now().UTC().Format(time.RFC3339Nano),
		"restore-qualification-object-v2-" + time.Now().UTC().Format(time.RFC3339Nano),
	} {
		obligation := fmt.Sprintf("restore-fixture-obligation-%d-%d", i, time.Now().UnixNano())
		ref, err := catalog.StoreResult(ctx, tenant, obligation, objectstore.UploadRequest{
			Size:        int64(len(content)),
			ContentType: "application/octet-stream",
			Purpose:     "RESULT",
			Class:       class,
			Region:      "fixture-local",
		}, bytes.NewBufferString(content))
		if err != nil {
			log.Fatalf("store result %d: %v", i, err)
		}
		sum := sha256.Sum256([]byte(content))
		fmt.Printf("populated file_ref id=%s version=%s sha256=%s tenant=%s obligation=%s\n", ref.ID, ref.Version, hex.EncodeToString(sum[:]), tenant, obligation)
	}
}
