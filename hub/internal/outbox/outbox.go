// Package outbox implementa o padrao outbox transacional (COM-03):
// publicacao de fato na mesma transacao do estado local; um relay
// separado publica no broker e so entao marca a linha como publicada.
// Compartilhado por Orbita e Cometa, que possuem a mesma base fisica
// hub_core mas dominios de escrita distintos (DAD-01) — cada app so
// enfileira/consome linhas do seu proprio aggregate_type.
package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Row e uma linha pendente/publicada da tabela outbox.
type Row struct {
	ID            int64
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       json.RawMessage
}

// Enqueue insere um novo evento na outbox dentro de uma transacao ja
// aberta pelo chamador, garantindo que o fato e o efeito/estado local
// sejam confirmados atomicamente (COM-03).
func Enqueue(ctx context.Context, tx *sql.Tx, aggregateType, aggregateID, eventType string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("outbox: serializar payload: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, published)
		VALUES ($1, $2, $3, $4, FALSE)
	`, aggregateType, aggregateID, eventType, body)
	if err != nil {
		return fmt.Errorf("outbox: inserir evento: %w", err)
	}
	return nil
}

// FetchPending busca ate limit linhas nao publicadas de um
// aggregate_type especifico, em ordem de insercao.
func FetchPending(ctx context.Context, db *sql.DB, aggregateType string, limit int) ([]Row, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload
		FROM outbox
		WHERE published = FALSE AND aggregate_type = $1
		ORDER BY id
		LIMIT $2
	`, aggregateType, limit)
	if err != nil {
		return nil, fmt.Errorf("outbox: buscar pendentes: %w", err)
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.AggregateType, &r.AggregateID, &r.EventType, &r.Payload); err != nil {
			return nil, fmt.Errorf("outbox: ler linha pendente: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// MarkPublished marca uma linha como publicada, somente apos
// confirmacao do broker.
func MarkPublished(ctx context.Context, db *sql.DB, id int64) error {
	_, err := db.ExecContext(ctx, `UPDATE outbox SET published = TRUE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("outbox: marcar publicado: %w", err)
	}
	return nil
}
