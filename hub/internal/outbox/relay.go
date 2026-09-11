package outbox

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"ai-hub/hub/internal/platform/httpserver"
)

// PublishFunc publica uma linha da outbox no broker (SNS). Retornar
// erro mantem a linha pendente para nova tentativa (COM-03: o
// publicador pode repetir o mesmo event_id).
type PublishFunc func(ctx context.Context, row Row) error

// RunRelay consome a outbox de um aggregate_type especifico em loop,
// publicando cada linha pendente e marcando-a publicada somente apos
// sucesso. Bloqueia ate o contexto ser cancelado — deve rodar em sua
// propria goroutine.
func RunRelay(ctx context.Context, db *sql.DB, aggregateType string, publish PublishFunc, interval time.Duration, log *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, err := FetchPending(ctx, db, aggregateType, 50)
			if err != nil {
				log.Error("outbox: falha ao buscar pendentes", "error", err, "aggregate_type", aggregateType)
				continue
			}
			for _, row := range rows {
				if err := publish(ctx, row); err != nil {
					log.Error("outbox: falha ao publicar, mantendo pendente", "error", err, "id", row.ID)
					continue
				}
				if err := MarkPublished(ctx, db, row.ID); err != nil {
					log.Error("outbox: falha ao marcar publicado", "error", err, "id", row.ID)
				}
			}
		}
	}
}

// RunPendingAgeMetric publica a idade da obrigação não publicada mais antiga
// consultando a mesma autoridade PostgreSQL que o relay protege. O valor é
// um gauge de baixa cardinalidade, separado por tipo de agregado e serviço.
func RunPendingAgeMetric(ctx context.Context, db *sql.DB, aggregateType, component string, interval time.Duration, registry *httpserver.Registry, log *slog.Logger) {
	update := func() {
		var age sql.NullFloat64
		err := db.QueryRowContext(ctx, `SELECT COALESCE(EXTRACT(EPOCH FROM (clock_timestamp()-MIN(created_at))),0) FROM outbox WHERE published=FALSE AND aggregate_type=$1`, aggregateType).Scan(&age)
		if err != nil {
			log.Warn("outbox: falha ao medir idade da obrigação", "error", err, "aggregate_type", aggregateType)
			return
		}
		value := 0.0
		if age.Valid && age.Float64 > 0 {
			value = age.Float64
		}
		registry.SetGauge("hub_obligation_age_seconds", map[string]string{"kind": "outbox", "component": component}, value)
	}
	update()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			update()
		}
	}
}
