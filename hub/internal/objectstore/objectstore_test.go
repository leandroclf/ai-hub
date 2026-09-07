//go:build e2e

// Teste de integracao real contra o S3 emulado (LocalStack), provando
// que o cliente objectstore (DAD-05) funciona. Nao esta encadeado a
// nenhum fluxo de admissao real ainda — ver IMPLEMENTATION_AUDIT.md
// ("Cache (DAD-10)" / objetos): este e um teste do cliente em si, nao
// evidencia de uso em producao.
package objectstore

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestPutGetRoundTrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := New(ctx, "http://localhost:4566", "us-east-1", "hub-objects-test")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := client.EnsureBucket(ctx); err != nil {
		t.Fatalf("EnsureBucket: %v", err)
	}

	key := "evidence/roundtrip.json"
	body := []byte(`{"protocol_id":"evidence-test","result":"objeto grande de exemplo (DAD-05)"}`)

	if err := client.Put(ctx, key, body, "application/json"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("conteudo divergente: got=%s want=%s", got, body)
	}
}
