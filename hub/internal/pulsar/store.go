// Package pulsar implementa a entrega de webhook: agenda, assinatura
// HMAC, tentativas e recibos (EXE-08, COM-05).
package pulsar

import (
	"context"
	"database/sql"
)

// Store encapsula o acesso as tabelas de dominio do Pulsar em hub_core:
// deliveries e webhook_destination_versions.
type Store struct {
	db *sql.DB
}

// NewStore cria um Store a partir de uma conexao ja aberta com hub_core.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Ping verifica a conectividade com hub_core (readiness).
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
