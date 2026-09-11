package cometa

import (
	"errors"
	"os"
	"strings"
)

var ErrAdapterUnavailable = errors.New("cometa: adapter não habilitado no runtime")

// AdapterRegistry é a fronteira entre a homologação publicada no Atlas e a
// capacidade efetivamente carregada no Cometa. O catálogo pode conservar
// versões históricas, mas somente IDs explicitamente instalados chegam ao
// transporte externo.
type AdapterRegistry struct {
	allowed map[string]struct{}
}

var compiledAdapters = map[string]struct{}{
	"provider-sim":       {},
	"synthetic-provider": {},
}

func NewAdapterRegistry(ids ...string) *AdapterRegistry {
	r := &AdapterRegistry{allowed: map[string]struct{}{}}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if _, compiled := compiledAdapters[id]; compiled {
			r.allowed[id] = struct{}{}
		}
	}
	return r
}

// AdapterRegistryFromEnv mantém somente fixtures explicitamente instaladas
// quando COMETA_ADAPTERS existe. Sem configuração, os adapters embutidos da
// referência local permanecem disponíveis para o provider-sim e seus testes.
// Um adapter comercial só pode ser habilitado depois de ser registrado no
// binário; adicionar um ID na variável de ambiente não cria transporte.
func AdapterRegistryFromEnv() *AdapterRegistry {
	raw := strings.TrimSpace(os.Getenv("COMETA_ADAPTERS"))
	if raw == "" {
		return NewAdapterRegistry("provider-sim", "synthetic-provider")
	}
	return NewAdapterRegistry(strings.Split(raw, ",")...)
}

func (r *AdapterRegistry) Supports(id string) bool {
	if r == nil {
		return false
	}
	_, ok := r.allowed[strings.TrimSpace(id)]
	return ok
}
