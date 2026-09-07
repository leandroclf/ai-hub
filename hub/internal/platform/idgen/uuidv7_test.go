package idgen

import (
	"testing"

	"github.com/google/uuid"
)

// TestNewIsUUIDv7 valida EXE-16: protocol_id (e demais identificadores
// derivados) usa UUIDv7 conforme RFC 9562, nao UUIDv4/ULID.
func TestNewIsUUIDv7(t *testing.T) {
	id := New()
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("New() nao produziu um UUID valido: %v", err)
	}
	if parsed.Version() != 7 {
		t.Fatalf("versao = %d, want 7 (UUIDv7)", parsed.Version())
	}
}

// TestNewIsUnique garante que chamadas sucessivas nao colidem — a
// unicidade e protegida pela persistencia (EXE-16), mas o gerador em
// si tambem nao deve repetir sob uso normal.
func TestNewIsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := New()
		if seen[id] {
			t.Fatalf("UUID repetido na iteracao %d: %s", i, id)
		}
		seen[id] = true
	}
}
