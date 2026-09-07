package cometa

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/queue"
)

// RunCommandWorker consome comandos QUEUED (ASYNC/AUTO) da fila
// dedicada da celula (COM-01) e executa exatamente a mesma logica do
// despacho DIRETO (EXE-15: "watchdog, fila e resposta direta
// compartilham pre-condicoes e chaves economicas"). O resultado nao
// retorna por HTTP: e publicado como fato (outbox -> SNS), consumido
// pela Orbita para finalizar o protocolo.
func RunCommandWorker(ctx context.Context, q *queue.Client, queueURL string, exec *Executor, log *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := q.Receive(ctx, queueURL, 5, 10)
		if err != nil {
			log.Error("worker: falha ao receber comandos", "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		for _, m := range msgs {
			var cmd dispatch.Command
			if err := json.Unmarshal(m.Envelope.Payload, &cmd); err != nil {
				log.Error("worker: comando invalido, descartando", "error", err)
				_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				continue
			}
			if cmd.DispatchMode != dispatch.DispatchQueued {
				// EXE-15: o despachante de outbox so publica comandos
				// elegiveis QUEUED; uma intencao DIRECT nunca deveria
				// chegar aqui.
				log.Warn("worker: comando com dispatch_mode inesperado, descartando", "dispatch_mode", cmd.DispatchMode)
				_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				continue
			}
			execCtx := ctx
			if !cmd.StepDeadline.IsZero() {
				var cancel func()
				execCtx, cancel = context.WithDeadline(ctx, cmd.StepDeadline)
				exec.Execute(execCtx, cmd)
				cancel()
			} else {
				exec.Execute(execCtx, cmd)
			}
			// Ack somente apos o efeito/intencao local estar
			// confirmado (COM-03): Execute ja persistiu antes de
			// retornar.
			_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
		}
	}
}
