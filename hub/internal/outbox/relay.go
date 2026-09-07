package outbox

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
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
