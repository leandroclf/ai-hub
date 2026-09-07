package orbita

import (
	"context"
	"encoding/json"
	"log/slog"

	"ai-hub/hub/internal/queue"
)

type operationFact struct {
	ProtocolID   string `json:"protocol_id"`
	TraceID      string `json:"trace_id,omitempty"`
	Kind         string `json:"kind"`
	ResponseBody any    `json:"response_body,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// RunOperationFactConsumer consome os fatos de operacao publicados por
// Cometa (COM-01) para finalizar protocolos ASYNC/AUTO — o mesmo
// caminho de conclusao que o SYNC direto usa, apenas disparado por
// evento em vez de resposta HTTP imediata (EXE-15: "watchdog, fila e
// resposta direta compartilham pre-condicoes").
func RunOperationFactConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, finalizer *Finalizer, log *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := q.Receive(ctx, queueURL, 5, 10)
		if err != nil {
			log.Error("fact consumer: falha ao receber fatos de operacao", "error", err)
			continue
		}
		for _, m := range msgs {
			var fact operationFact
			if err := json.Unmarshal(m.Envelope.Payload, &fact); err != nil {
				log.Error("fact consumer: payload invalido, descartando", "error", err)
				_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				continue
			}
			if fact.Kind == "UNKNOWN" || fact.ProtocolID == "" {
				// Ainda pendente/incerto: nao finaliza; deixa para
				// reconciliacao ou proxima observacao (EXE-09).
				log.Debug("fact consumer: fato UNKNOWN/incompleto, aguardando proxima observacao", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID)
				_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				continue
			}
			log.Debug("fact consumer: fato de operacao recebido", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID, "kind", fact.Kind)
			p, err := store.GetByID(ctx, fact.ProtocolID)
			if err != nil {
				log.Warn("fact consumer: protocolo desconhecido para fato recebido", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID)
				_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				continue
			}
			if !IsTerminal(p.Status) {
				status := StatusFailed
				if fact.Kind == "SUCCEEDED" {
					status = StatusSucceeded
				}
				body := FinalBody{ProtocolID: fact.ProtocolID, Result: fact.ResponseBody, ErrorMessage: fact.ErrorMessage}
				if _, err := finalizer.Finalize(ctx, fact.TraceID, p.TenantID, fact.ProtocolID, p.Version, status, body, ""); err != nil {
					log.Error("fact consumer: falha ao finalizar protocolo", "trace_id", fact.TraceID, "error", err, "protocol_id", fact.ProtocolID)
				}
			}
			// Ack somente apos a finalizacao (ou decisao de ignorar)
			// estar confirmada localmente (COM-03).
			_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
		}
	}
}
