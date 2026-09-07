// Package libra implementa a medicao economica, reserva estrita de
// saldo e ledger gerencial (FIN-04, FIN-06, FIN-07).
package libra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrLimitExceeded e retornado quando uma reserva excederia o teto
// financeiro estrito do tenant (FIN-06).
var ErrLimitExceeded = errors.New("libra: reserva excederia o limite estrito do tenant")

// Store encapsula o acesso a hub_finance.
type Store struct {
	db *sql.DB
}

// NewStore cria um Store a partir de uma conexao ja aberta com hub_finance.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// RecordFact grava um fato economico imutavel com dedup semantica por
// protocolo/natureza/medidor (FIN-04): "uma unidade elegivel apesar de
// multiplos passos"; reentrega do evento de origem nao duplica.
func (s *Store) RecordFact(ctx context.Context, tenantID, protocolID, kind, meter string, amount float64, currency string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO economic_facts (tenant_id, protocol_id, kind, meter, amount, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (protocol_id, kind, meter) DO NOTHING
	`, tenantID, protocolID, kind, meter, amount, currency)
	if err != nil {
		return fmt.Errorf("libra: registrar fato economico: %w", err)
	}
	return nil
}

// Reserve tenta reservar atomicamente um valor para o protocolo,
// respeitando o teto estrito do tenant (FIN-06): "reserva atomica na
// autoridade financeira antes de qualquer efeito externo". Idempotente
// por protocol_id: uma segunda chamada para o mesmo protocolo apenas
// confirma a reserva ja existente (consulta idempotente resolve
// resposta de reserva perdida).
func (s *Store) Reserve(ctx context.Context, tenantID, protocolID string, amount float64, currency string, limitAmount float64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existing sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT state FROM reservations WHERE protocol_id = $1`, protocolID).Scan(&existing)
	if err == nil {
		// Ja existe reserva para este protocolo: idempotente.
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("libra: consultar reserva existente: %w", err)
	}

	// Serializa reservas concorrentes do mesmo tenant (FIN-06: "reserva
	// atomica"). FOR UPDATE nao e permitido sobre uma agregacao (SUM),
	// entao usamos um lock consultivo por tenant dentro da transacao.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, tenantID); err != nil {
		return fmt.Errorf("libra: obter lock de tenant: %w", err)
	}

	var reserved float64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM reservations
		WHERE tenant_id = $1 AND state = 'RESERVED'
	`, tenantID).Scan(&reserved)
	if err != nil {
		return fmt.Errorf("libra: somar reservas ativas: %w", err)
	}
	if reserved+amount > limitAmount {
		return ErrLimitExceeded
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO reservations (protocol_id, tenant_id, amount, currency, state)
		VALUES ($1, $2, $3, $4, 'RESERVED')
	`, protocolID, tenantID, amount, currency)
	if err != nil {
		return fmt.Errorf("libra: inserir reserva: %w", err)
	}
	return tx.Commit()
}

// Release libera uma reserva (ex.: protocolo encerrado sem captura).
func (s *Store) Release(ctx context.Context, protocolID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE reservations SET state = 'RELEASED', updated_at = now() WHERE protocol_id = $1 AND state = 'RESERVED'`, protocolID)
	if err != nil {
		return fmt.Errorf("libra: liberar reserva: %w", err)
	}
	return nil
}

// Capture confirma o consumo efetivo de uma reserva.
func (s *Store) Capture(ctx context.Context, protocolID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE reservations SET state = 'CAPTURED', updated_at = now() WHERE protocol_id = $1 AND state = 'RESERVED'`, protocolID)
	if err != nil {
		return fmt.Errorf("libra: capturar reserva: %w", err)
	}
	return nil
}

// CreditLimit le o teto financeiro estrito do tenant; ausencia de
// linha usa um teto alto padrao neste ambiente de referencia.
func (s *Store) CreditLimit(ctx context.Context, tenantID string) (float64, error) {
	var limit float64
	err := s.db.QueryRowContext(ctx, `SELECT limit_amount FROM credit_limits WHERE tenant_id = $1`, tenantID).Scan(&limit)
	if errors.Is(err, sql.ErrNoRows) {
		return 1_000_000, nil
	}
	if err != nil {
		return 0, fmt.Errorf("libra: ler limite de credito: %w", err)
	}
	return limit, nil
}

// Ping verifica a conectividade com hub_finance (readiness).
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
