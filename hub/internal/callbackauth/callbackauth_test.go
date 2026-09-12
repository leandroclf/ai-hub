package callbackauth

import (
	"testing"
	"time"
)

func TestSignatureIsScopedToAccountOperationAndBody(t *testing.T) {
	when := time.Unix(1789173000, 0).UTC()
	signature := Sign("fixture-root", "provider-account-a", "operation-a", when, []byte(`{"status":"SUCCEEDED"}`))
	if signature == "" {
		t.Fatal("assinatura vazia")
	}
	if !Verify("fixture-root", "provider-account-a", "operation-a", when, []byte(`{"status":"SUCCEEDED"}`), signature) {
		t.Fatal("assinatura válida foi recusada")
	}
	for _, tc := range []struct {
		name, account, operation, body string
	}{
		{"outra conta", "provider-account-b", "operation-a", `{"status":"SUCCEEDED"}`},
		{"outra operação", "provider-account-a", "operation-b", `{"status":"SUCCEEDED"}`},
		{"corpo adulterado", "provider-account-a", "operation-a", `{"status":"FAILED"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if Verify("fixture-root", tc.account, tc.operation, when, []byte(tc.body), signature) {
				t.Fatal("assinatura foi aceita fora do escopo")
			}
		})
	}
}

func TestSignatureRejectsMissingRootAndMalformedValue(t *testing.T) {
	when := time.Unix(1789173000, 0).UTC()
	if Verify("", "provider-account-a", "operation-a", when, []byte("body"), "sha256=deadbeef") {
		t.Fatal("raiz vazia autorizou callback")
	}
	if Verify("fixture-root", "provider-account-a", "operation-a", when, []byte("body"), "not-a-signature") {
		t.Fatal("assinatura malformada foi aceita")
	}
}
