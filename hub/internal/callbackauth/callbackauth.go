// Package callbackauth implementa a autenticação de callbacks externos por
// conta, sem transportar credencial na URL. A chave raiz é fornecida pelo
// boundary do provedor e a derivação inclui a conta, impedindo reutilização
// silenciosa entre contas.
package callbackauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"time"
)

const Version = "v1"

// Sign produz uma assinatura HMAC-SHA256 para a identidade completa do
// callback. O corpo é incluído por hash para evitar ambiguidade de framing.
func Sign(root, accountID, operationID string, at time.Time, body []byte) string {
	if root == "" || accountID == "" || operationID == "" || at.IsZero() || len(body) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, deriveKey(root, accountID))
	mac.Write([]byte(canonical(operationID, at, body)))
	return Version + "=" + hex.EncodeToString(mac.Sum(nil))
}

// Verify compara uma assinatura sem aceitar escopo de outra conta/operação
// ou alteração do corpo recebido.
func Verify(root, accountID, operationID string, at time.Time, body []byte, value string) bool {
	expected := Sign(root, accountID, operationID, at, body)
	if expected == "" || len(value) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(value)) == 1
}

func deriveKey(root, accountID string) []byte {
	mac := hmac.New(sha256.New, []byte(root))
	mac.Write([]byte("ai-hub/callback-account/"))
	mac.Write([]byte(accountID))
	return mac.Sum(nil)
}

func canonical(operationID string, at time.Time, body []byte) string {
	digest := sha256.Sum256(body)
	return strings.Join([]string{Version, operationID, at.UTC().Format("2006-01-02T15:04:05Z"), hex.EncodeToString(digest[:])}, "\n")
}
