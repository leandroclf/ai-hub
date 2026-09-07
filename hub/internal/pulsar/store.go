// Package pulsar implementa a entrega de webhook: agenda, assinatura
// HMAC, tentativas e recibos (EXE-08, COM-05).
package pulsar

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrDestinationNotFound e retornado quando o tenant nao tem destino
// de webhook cadastrado.
var ErrDestinationNotFound = errors.New("pulsar: destino de webhook nao cadastrado")

// Store encapsula o acesso as tabelas de dominio do Pulsar em hub_core:
// deliveries e webhook_destinations.
type Store struct {
	db *sql.DB
}

// NewStore cria um Store a partir de uma conexao ja aberta com hub_core.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Destination e o destino de webhook cadastrado de um tenant (EXE-08:
// "destino previamente cadastrado, verificado e vinculado ao tenant").
type Destination struct {
	TenantID   string `json:"tenant_id"`
	URL        string `json:"url"`
	HMACSecret string `json:"hmac_secret"`
}

// UpsertDestination cadastra/atualiza o destino de webhook de um tenant.
func (s *Store) UpsertDestination(ctx context.Context, d Destination) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO webhook_destinations (tenant_id, url, hmac_secret)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id) DO UPDATE SET url = EXCLUDED.url, hmac_secret = EXCLUDED.hmac_secret
	`, d.TenantID, d.URL, d.HMACSecret)
	if err != nil {
		return fmt.Errorf("pulsar: cadastrar destino: %w", err)
	}
	return nil
}

// GetDestination le o destino de webhook de um tenant.
func (s *Store) GetDestination(ctx context.Context, tenantID string) (Destination, error) {
	var d Destination
	row := s.db.QueryRowContext(ctx, `SELECT tenant_id, url, hmac_secret FROM webhook_destinations WHERE tenant_id = $1`, tenantID)
	err := row.Scan(&d.TenantID, &d.URL, &d.HMACSecret)
	if errors.Is(err, sql.ErrNoRows) {
		return Destination{}, ErrDestinationNotFound
	}
	if err != nil {
		return Destination{}, fmt.Errorf("pulsar: ler destino: %w", err)
	}
	return d, nil
}

// getDestinationByURL busca o destino cadastrado cuja URL corresponde
// a uma entrega ja criada — usado para recuperar o segredo HMAC no
// momento do envio, sem duplicar o segredo na tabela deliveries.
func (s *Store) getDestinationByURL(ctx context.Context, url string) (Destination, error) {
	var d Destination
	row := s.db.QueryRowContext(ctx, `SELECT tenant_id, url, hmac_secret FROM webhook_destinations WHERE url = $1`, url)
	err := row.Scan(&d.TenantID, &d.URL, &d.HMACSecret)
	if errors.Is(err, sql.ErrNoRows) {
		return Destination{}, ErrDestinationNotFound
	}
	if err != nil {
		return Destination{}, fmt.Errorf("pulsar: ler destino por url: %w", err)
	}
	return d, nil
}

// Delivery e a projecao persistida de uma entrega (DAD-02).
type Delivery struct {
	DeliveryID     string
	ProtocolID     string
	EventID        string
	DestinationURL string
	State          string
	AttemptsCount  int
	NextAttemptAt  time.Time
}

// CreateDelivery cria uma entrega por evento final e destino (EXE-08:
// "criar uma entrega por evento final e destino/versao"); unica por
// event_id/destino — reentrega do mesmo fato nao duplica a entrega.
func (s *Store) CreateDelivery(ctx context.Context, deliveryID, protocolID, eventID, destinationURL string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO deliveries (delivery_id, protocol_id, event_id, destination_url, state, next_attempt_at)
		VALUES ($1, $2, $3, $4, 'PENDING', now())
		ON CONFLICT (event_id, destination_url) DO NOTHING
	`, deliveryID, protocolID, eventID, destinationURL)
	if err != nil {
		return fmt.Errorf("pulsar: criar entrega: %w", err)
	}
	return nil
}

// DueDeliveries retorna entregas pendentes/agendadas cujo horario de
// proxima tentativa ja chegou.
func (s *Store) DueDeliveries(ctx context.Context, now time.Time, limit int) ([]Delivery, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT delivery_id, protocol_id, event_id, destination_url, state, attempts_count, next_attempt_at
		FROM deliveries
		WHERE state IN ('PENDING','RETRY_SCHEDULED') AND next_attempt_at <= $1
		ORDER BY next_attempt_at
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("pulsar: buscar entregas devidas: %w", err)
	}
	defer rows.Close()
	var out []Delivery
	for rows.Next() {
		var d Delivery
		if err := rows.Scan(&d.DeliveryID, &d.ProtocolID, &d.EventID, &d.DestinationURL, &d.State, &d.AttemptsCount, &d.NextAttemptAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// MarkDelivered marca uma entrega como concluida.
func (s *Store) MarkDelivered(ctx context.Context, deliveryID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE deliveries SET state = 'DELIVERED', updated_at = now() WHERE delivery_id = $1`, deliveryID)
	return err
}

// RetryOrExhaust reagenda a entrega com backoff ou a marca EXHAUSTED ao
// atingir o teto de tentativas (EXE-08: "esgotamento gera EXHAUSTED e
// alerta; resultado continua disponivel segundo retencao").
func (s *Store) RetryOrExhaust(ctx context.Context, deliveryID string, attemptsCount, maxAttempts int, nextAttemptAt time.Time) error {
	state := "RETRY_SCHEDULED"
	if attemptsCount >= maxAttempts {
		state = "EXHAUSTED"
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE deliveries SET state = $2, attempts_count = $3, next_attempt_at = $4, updated_at = now()
		WHERE delivery_id = $1
	`, deliveryID, state, attemptsCount, nextAttemptAt)
	return err
}

// Ping verifica a conectividade com hub_core (readiness).
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
