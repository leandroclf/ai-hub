package cometa

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"os"
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
			if json.Unmarshal(m.Envelope.Payload, &cmd) != nil || cmd.DispatchMode != dispatch.DispatchQueued || cmd.CellID != os.Getenv("CELL_ID") || m.Envelope.TenantID != cmd.TenantID || m.Envelope.EventID != cmd.CommandID || m.Envelope.Type != "command.dispatch" || m.Envelope.Producer != "orbita" || m.Envelope.SchemaVersion != 1 {
				sum := sha256.Sum256(m.RawBody)
				_, err := exec.store.db.ExecContext(ctx, `INSERT INTO message_quarantine(consumer,body_sha256,body,reason) VALUES('cometa', $1,$2,'invalid_command_envelope') ON CONFLICT(consumer,body_sha256) DO UPDATE SET last_seen_at=clock_timestamp(),occurrences=message_quarantine.occurrences+1`, hex.EncodeToString(sum[:]), m.RawBody)
				if err == nil {
					_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				}
				continue
			}
			execCtx := ctx
			var result dispatch.Result
			if !cmd.StepDeadline.IsZero() {
				var cancel func()
				execCtx, cancel = context.WithDeadline(ctx, cmd.StepDeadline)
				result = exec.Execute(execCtx, cmd)
				cancel()
			} else {
				result = exec.Execute(execCtx, cmd)
			}
			// Ack somente apos o efeito/intencao local estar
			// confirmado (COM-03): Execute ja persistiu antes de
			// retornar.
			if result.Durable {
				_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
			}
		}
	}
}
