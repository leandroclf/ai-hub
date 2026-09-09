package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/lib/pq"
)

func TestPostgresS3MultipartFileRef(t *testing.T) {
	dsn, endpoint := os.Getenv("R2_CORE_TEST_DSN"), os.Getenv("R2_S3_TEST_ENDPOINT")
	if dsn == "" || endpoint == "" {
		t.Skip("requires isolated PostgreSQL and S3")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	schema := fmt.Sprintf("r2_objects_%d", time.Now().UnixNano())
	if _, err = root.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	defer root.Exec("DROP SCHEMA " + schema + " CASCADE")
	u, _ := url.Parse(dsn)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	migration, err := os.ReadFile("../../migrations/core/0030_objects_retention.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO object_retention_policies VALUES('SYNTHETIC','fixture-local','TEST',3600,67108864,'["application/octet-stream"]','FIXTURE_ONLY_NOT_PRODUCTION')`)
	if err != nil {
		t.Fatal(err)
	}
	bucket := fmt.Sprintf("r2-files-test-%d", time.Now().UnixNano())
	client, err := New(ctx, endpoint, "us-east-1", bucket)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.EnsureBucket(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.s3.DeleteBucket(context.Background(), &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
	catalog := NewCatalog(db, client)
	file, err := os.Create(filepath.Join(t.TempDir(), "synthetic.bin"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	part := bytes.Repeat([]byte{0x52}, int(PartSize))
	hash := sha256.New()
	writer := io.MultiWriter(file, hash)
	for i := 0; i < 2; i++ {
		if _, err = writer.Write(part); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = writer.Write(part[:128]); err != nil {
		t.Fatal(err)
	}
	size := PartSize*2 + 128
	sha := hex.EncodeToString(hash.Sum(nil))
	session, err := catalog.CreateSession(ctx, "tenant-a", UploadRequest{Size: size, SHA256: sha, ContentType: "application/octet-stream", Purpose: "TEST", Class: "SYNTHETIC", Region: "fixture-local"})
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Uploads) != 3 {
		t.Fatalf("multipart urls=%d", len(session.Uploads))
	}
	if _, err = catalog.Resolve(ctx, "tenant-b", session.FileRef.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("cross tenant upload visible")
	}
	var parts []CompletedPart
	var offset int64
	for _, upload := range session.Uploads {
		request, err := http.NewRequestWithContext(ctx, "PUT", upload.URL, io.NewSectionReader(file, offset, upload.Size))
		if err != nil {
			t.Fatal(err)
		}
		request.ContentLength = upload.Size
		request.Header = upload.Headers.Clone()
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatalf("upload status=%d", response.StatusCode)
		}
		parts = append(parts, CompletedPart{PartNumber: upload.PartNumber, ETag: response.Header.Get("ETag")})
		offset += upload.Size
	}
	ref, err := catalog.Complete(ctx, "tenant-a", session.FileRef.ID, parts)
	if err != nil {
		t.Fatal(err)
	}
	if ref.Version == "" || ref.SHA256 != sha || ref.Size != size {
		t.Fatalf("invalid FileRef: %+v", ref)
	}
	stored, err := catalog.lookup(ctx, "tenant-a", ref.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteVersion(context.Background(), stored.Key, ref.Version)
	if _, err = catalog.Download(ctx, "tenant-b", ref.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("foreign download authorized")
	}
	signed, err := catalog.Download(ctx, "tenant-a", ref.ID)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Get(signed)
	if err != nil {
		t.Fatal(err)
	}
	downloadHash := sha256.New()
	n, err := io.Copy(downloadHash, response.Body)
	response.Body.Close()
	if err != nil || n != size || hex.EncodeToString(downloadHash.Sum(nil)) != sha {
		t.Fatal("versioned download differs")
	}
	if _, err = db.Exec("UPDATE file_refs SET sha256='changed' WHERE id=$1", ref.ID); err == nil {
		t.Fatal("immutable checksum modified")
	}
	if err = catalog.Pin(ctx, "tenant-a", ref.ID, "protocol-obligation", "synthetic-test"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE file_refs SET retention_until=clock_timestamp()-interval '1 second' WHERE id=$1", ref.ID); err != nil {
		t.Fatal(err)
	}
	plan, err := catalog.DryRun(ctx, "tenant-a", "fixture-operator")
	if err != nil || len(plan.FileIDs) != 0 {
		t.Fatal("pinned obligation selected for purge")
	}
	if err = catalog.Unpin(ctx, "tenant-a", ref.ID, "protocol-obligation", "fixture-operator"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE file_refs SET retention_until=clock_timestamp()-interval '1 second' WHERE id=$1", ref.ID); err != nil {
		t.Fatal(err)
	}
	batch, err := catalog.RunPurgeBatch(ctx, "tenant-a", "fixture-operator")
	if err != nil || batch.Candidates != 1 || batch.Purged != 1 || batch.Failed != 0 {
		t.Fatalf("retention batch: %+v, err=%v", batch, err)
	}
	tombstoned, err := catalog.IsTombstoned(ctx, "tenant-a", ref.ID)
	if err != nil || !tombstoned {
		t.Fatalf("purge tombstone missing: %v", err)
	}
	if _, err = catalog.Download(ctx, "tenant-a", ref.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("purged object remained readable")
	}
	t.Logf("direct multipart upload and streaming verify/download: %d bytes, 3 parts; pin excluded first plan, then durable retention batch tombstoned and removed the unpinned version", size)
}
