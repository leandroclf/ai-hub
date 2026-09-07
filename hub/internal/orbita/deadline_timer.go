package orbita

import (
	"context"
	"log/slog"
	"time"
)

// RunDeadlineTimer encerra protocolos abertos ao atingir o deadline,
// mesmo sem erro tecnico e mesmo com polling saudavel (EXE-11): "um
// temporizador duravel da Orbita encerra protocolo aberto ao atingir o
// deadline". Disputa a mesma transicao terminal serializavel usada
// pelo caminho de sucesso (Store.Finalize com ExpectedVersion).
func RunDeadlineTimer(ctx context.Context, store *Store, finalizer *Finalizer, interval time.Duration, log *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			due, err := store.DueDeadlines(ctx, time.Now().UTC(), 100)
			if err != nil {
				log.Error("deadline timer: falha ao buscar protocolos vencidos", "error", err)
				continue
			}
			for _, d := range due {
				body := FinalBody{
					ProtocolID:   d.ProtocolID,
					ErrorCode:    "SLA_EXCEEDED",
					ErrorMessage: "prazo do cliente esgotado antes da conclusao (EXE-11)",
				}
				applied, err := finalizer.Finalize(ctx, "", d.TenantID, d.ProtocolID, d.Version, StatusExpired, body, "SLA_EXCEEDED")
				if err != nil {
					log.Error("deadline timer: falha ao finalizar por expiracao", "error", err, "protocol_id", d.ProtocolID)
					continue
				}
				if applied {
					log.Info("protocolo expirado por SLA", "protocol_id", d.ProtocolID)
				}
			}
		}
	}
}
