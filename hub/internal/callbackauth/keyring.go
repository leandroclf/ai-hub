package callbackauth

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// Key e uma chave de assinatura de callback identificada por ID. O valor do
// segredo nunca e persistido — apenas o ID e gravado junto ao callback
// custodiado (R6-SEG-02), permitindo reconciliar contra a mesma chave que
// autenticou originalmente, mesmo depois de uma rotacao.
type Key struct {
	ID     string
	Secret string
}

// KeyRing mantem todas as chaves de ingresso simultaneamente validas.
// Rotacionar significa adicionar uma nova chave ao ring (com um ID novo) sem
// remover a anterior ate que nenhum callback pendente possa mais chegar
// assinado com ela — a mesma logica de sobreposicao usada por rotacao de
// chaves JWKS (ver internal/platform/auth.Verifier).
type KeyRing struct {
	keys    map[string]string
	current string
}

var ErrEmptyKeyRing = errors.New("callbackauth: nenhuma chave de ingresso configurada")

// NewKeyRing valida e monta o ring. currentID, se nao vazio, deve existir no
// ring e identifica a chave usada para assinar novos callbacks (usado pelo
// lado do provedor/simulador, nao pelo Hub, que so verifica).
func NewKeyRing(keys []Key, currentID string) (KeyRing, error) {
	if len(keys) == 0 {
		return KeyRing{}, ErrEmptyKeyRing
	}
	m := make(map[string]string, len(keys))
	for _, k := range keys {
		if k.ID == "" || k.Secret == "" {
			return KeyRing{}, errors.New("callbackauth: id e segredo da chave sao obrigatorios")
		}
		m[k.ID] = k.Secret
	}
	if currentID != "" {
		if _, ok := m[currentID]; !ok {
			return KeyRing{}, errors.New("callbackauth: chave atual nao presente no ring")
		}
	}
	return KeyRing{keys: m, current: currentID}, nil
}

// VerifyAny tenta cada chave do ring e devolve o ID da que autenticou —
// usado no ingresso, quando ainda nao se sabe qual chave assinou o callback.
func (r KeyRing) VerifyAny(accountID, operationID string, at time.Time, body []byte, value string) (string, bool) {
	for id, secret := range r.keys {
		if Verify(secret, accountID, operationID, at, body, value) {
			return id, true
		}
	}
	return "", false
}

// VerifyWithID verifica contra exatamente a chave indicada — usado pela
// reconciliacao, que deve validar contra a chave que originalmente aceitou o
// callback, nunca "a chave atual" (R6-SEG-02: rotacao nao pode invalidar
// custodia legitima).
func (r KeyRing) VerifyWithID(keyID, accountID, operationID string, at time.Time, body []byte, value string) bool {
	secret, ok := r.keys[keyID]
	if !ok || keyID == "" {
		return false
	}
	return Verify(secret, accountID, operationID, at, body, value)
}

// HasKey informa se um key_id persistido ainda esta presente no ring (ele
// pode ter sido definitivamente removido apos o fim da janela de rotacao).
func (r KeyRing) HasKey(keyID string) bool {
	_, ok := r.keys[keyID]
	return ok
}

// CurrentID devolve o ID da chave corrente para assinatura (lado do
// provedor/simulador). Vazio quando nao ha chave corrente definida.
func (r KeyRing) CurrentID() string { return r.current }

// SignCurrent assina com a chave corrente do ring — usado apenas pelo
// simulador de provedor local, nunca pelo Hub (que so verifica).
func (r KeyRing) SignCurrent(accountID, operationID string, at time.Time, body []byte) (signature, keyID string) {
	if r.current == "" {
		return "", ""
	}
	return Sign(r.keys[r.current], accountID, operationID, at, body), r.current
}

type keyRingEnvEntry struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

// LoadKeyRingFromEnv monta o ring a partir de CALLBACK_INGRESS_KEYS (JSON:
// [{"id":"...","secret":"..."}]) e CALLBACK_INGRESS_CURRENT_KEY_ID. Mantem
// compatibilidade com a variavel legada CALLBACK_INGRESS_KEY (uma unica
// chave, id fixo "legacy") quando CALLBACK_INGRESS_KEYS nao esta definida.
func LoadKeyRingFromEnv() (KeyRing, error) {
	if raw := os.Getenv("CALLBACK_INGRESS_KEYS"); raw != "" {
		var entries []keyRingEnvEntry
		if err := json.Unmarshal([]byte(raw), &entries); err != nil {
			return KeyRing{}, errors.New("callbackauth: CALLBACK_INGRESS_KEYS invalido: " + err.Error())
		}
		keys := make([]Key, 0, len(entries))
		for _, e := range entries {
			keys = append(keys, Key{ID: e.ID, Secret: e.Secret})
		}
		return NewKeyRing(keys, os.Getenv("CALLBACK_INGRESS_CURRENT_KEY_ID"))
	}
	if legacy := os.Getenv("CALLBACK_INGRESS_KEY"); legacy != "" {
		return NewKeyRing([]Key{{ID: "legacy", Secret: legacy}}, "legacy")
	}
	return KeyRing{}, ErrEmptyKeyRing
}
