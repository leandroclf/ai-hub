package providerauth

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestAWSVaultLocalStackVersion(t *testing.T) {
	ref := os.Getenv("R2_TEST_SECRET_REF")
	if ref == "" {
		t.Skip("requires isolated LocalStack secret fixture")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	vault := AWSVault{}
	secret, err := vault.Resolve(ctx, ref, "")
	if err != nil {
		t.Fatal(err)
	}
	if secret.Value != "r2-synthetic-provider-password" || secret.Version == "" {
		t.Fatal("unexpected synthetic secret")
	}
	same, err := vault.Resolve(ctx, ref, secret.Version)
	if err != nil || same.Value != secret.Value || same.Version != secret.Version {
		t.Fatal("version pin failed")
	}
	if _, err = vault.Resolve(ctx, ref, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("missing version fell back")
	}
	t.Log("AWS SigV4 GetSecretValue against isolated LocalStack: decrypted synthetic value, pinned version, missing version denied; secret omitted from evidence")
}
