package cometa

import (
	"testing"
)

func TestAdapterRegistryFailsClosedForUnknownRuntimeAdapter(t *testing.T) {
	r := NewAdapterRegistry("provider-sim")
	if !r.Supports("provider-sim") {
		t.Fatal("adapter instalado foi recusado")
	}
	if r.Supports("provider-not-installed") {
		t.Fatal("adapter não instalado foi aceito")
	}
}

func TestAdapterRegistryReadsExplicitRuntimeAllowlist(t *testing.T) {
	t.Setenv("COMETA_ADAPTERS", "provider-acme-v1, provider-sim")
	r := AdapterRegistryFromEnv()
	if r.Supports("provider-acme-v1") || !r.Supports("provider-sim") || r.Supports("synthetic-provider") {
		t.Fatal("allowlist explícita não foi aplicada")
	}
}

func TestAdapterRegistryDoesNotTurnCatalogIDIntoTransport(t *testing.T) {
	if NewAdapterRegistry("provider-acme-v1").Supports("provider-acme-v1") {
		t.Fatal("ID de catálogo sem implementação compilada foi habilitado")
	}
}
