package cometa

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providersim"
	"ai-hub/hub/internal/queue"
)

// RunCallbackInboxWorker recupera callbacks órfãos sem depender de novo
// tráfego HTTP. Cada ciclo reivindica lote limitado; leases/epochs impedem
// que réplicas concorrentes ou workers atrasados apliquem o mesmo item.
func RunCallbackInboxWorker(ctx context.Context, store *Store, exec *Executor, interval time.Duration, log *slog.Logger) {
	if interval <= 0 {
		interval = time.Second
	}
	owner := "callback-reconciler-" + idgen.New()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	lastPrune := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if lastPrune.IsZero() || time.Since(lastPrune) >= time.Hour {
				if removed, err := store.PruneCallbackInbox(ctx, 7*24*time.Hour, 100); err != nil {
					log.Error("callback inbox retention failed", "error", err)
				} else if removed > 0 {
					log.Info("callback inbox retention applied", "removed", removed)
				}
				lastPrune = time.Now()
			}
			_, err := store.ReconcileCallbackInboxBatch(ctx, owner, 25, func(ctx context.Context, id string, observed dispatch.Result) error {
				status := "FAILED"
				if observed.Kind == dispatch.FactSucceeded {
					status = "SUCCEEDED"
				}
				if observed.Kind != dispatch.FactSucceeded && observed.Kind != dispatch.FactFailed {
					return errors.New("invalid reconciled callback status")
				}
				providerResult := providersim.OperationResult{ProviderRequestID: observed.ProviderRequestID, Status: status, Detail: observed.ErrorMessage}
				raw := observed.RawResponse
				if len(raw) > 0 {
					if err := json.Unmarshal(raw, &providerResult); err != nil {
						return err
					}
				} else {
					raw, _ = json.Marshal(providerResult)
				}
				_, err := exec.ApplyExternalObservationRaw(ctx, id, providerResult, raw)
				return err
			})
			if err != nil {
				log.Error("callback inbox reconciliation failed", "error", err)
			}
		}
	}
}

// RunCommandWorker consome comandos QUEUED (ASYNC/AUTO) da fila
// dedicada da celula (COM-01) e executa exatamente a mesma logica do
// despacho DIRETO (EXE-15: "watchdog, fila e resposta direta
// compartilham pre-condicoes e chaves economicas"). O resultado nao
// retorna por HTTP: e publicado como fato (outbox -> SNS), consumido
// pela Orbita para finalizar o protocolo.
func RunCommandWorker(ctx context.Context, q *queue.Client, queueURL string, exec *Executor, log *slog.Logger) {
	log.Info("worker de comandos iniciado", "queue_url", queueURL)
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
		if len(msgs) > 0 {
			log.Info("worker recebeu lote de comandos", "count", len(msgs))
		}
		for _, m := range msgs {
			var cmd dispatch.Command
			if json.Unmarshal(m.Envelope.Payload, &cmd) != nil || cmd.DispatchMode != dispatch.DispatchQueued || cmd.CellID != os.Getenv("CELL_ID") || m.Envelope.TenantID != cmd.TenantID || m.Envelope.EventID != cmd.CommandID || m.Envelope.Type != "command.dispatch" || m.Envelope.Producer != "orbita" || m.Envelope.SchemaVersion != 1 {
				if err := quarantineCommand(ctx, exec.store.db, m, "invalid_command_envelope"); err == nil {
					_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				} else {
					log.Error("worker: quarentena do envelope indisponivel", "error", err)
				}
				continue
			}
			if commandDeadlineExpired(cmd, time.Now()) {
				if err := quarantineCommand(ctx, exec.store.db, m, "expired_command_deadline"); err == nil {
					_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				} else {
					log.Error("worker: quarentena do comando expirado indisponivel", "error", err)
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
			log.Info("worker processou comando", "trace_id", cmd.TraceID, "protocol_id", cmd.ProtocolID, "kind", result.Kind, "durable", result.Durable, "error_code", result.ErrorCode)
			if reason := commandResultQuarantineReason(result); reason != "" {
				if err := quarantineCommand(ctx, exec.store.db, m, reason); err != nil {
					log.Error("worker: quarentena da rejeicao indisponivel", "error", err, "error_code", result.ErrorCode)
					continue
				}
				if err := q.Delete(ctx, queueURL, m.ReceiptHandle); err != nil {
					log.Warn("worker: ACK da rejeicao indisponivel; redelivery e segura", "error", err, "error_code", result.ErrorCode)
				}
				continue
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

func commandResultQuarantineReason(result dispatch.Result) string {
	if result.Kind != dispatch.FactRejected {
		return ""
	}
	switch result.ErrorCode {
	case "invalid_command", "invalid_snapshot", "adapter_not_qualified", "adapter_unavailable", "binding_changed", "account_snapshot_invalid":
		return "rejected_command:" + result.ErrorCode
	default:
		// Capacity and credential failures can recover without changing the
		// accepted command. Preserve them for redelivery instead of dropping
		// an obligation that may still be executable.
		return ""
	}
}

func commandDeadlineExpired(cmd dispatch.Command, now time.Time) bool {
	return !cmd.StepDeadline.IsZero() && !now.Before(cmd.StepDeadline)
}

func quarantineCommand(ctx context.Context, db *sql.DB, m queue.ReceivedMessage, reason string) error {
	sum := sha256.Sum256(m.RawBody)
	_, err := db.ExecContext(ctx, `INSERT INTO message_quarantine(consumer,body_sha256,body,reason) VALUES('cometa', $1,$2,$3) ON CONFLICT(consumer,body_sha256) DO UPDATE SET last_seen_at=clock_timestamp(),occurrences=message_quarantine.occurrences+1,reason=EXCLUDED.reason`, hex.EncodeToString(sum[:]), m.RawBody, reason)
	return err
}
