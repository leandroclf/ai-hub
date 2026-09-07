package pulsar

import (
	"context"
	"encoding/json"
	"log/slog"

	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/queue"
)

type protocolFact struct {
	ProtocolID string          `json:"protocol_id"`
	TraceID    string          `json:"trace_id,omitempty"`
	TenantID   string          `json:"tenant_id"`
	Status     string          `json:"status"`
	FinalBody  json.RawMessage `json:"final_body"`
	EventID    string          `json:"event_id"`
}

// RunFactConsumer consome "protocol.finalized" (COM-01: "Orbita ->
// Pulsar e Libra: fato final via SNS/SQS") e cria uma entrega por
// evento final e destino, quando o tenant possui destino cadastrado.
// Ausencia de destino nao e erro: apenas nao ha o que entregar (o GET
// do cliente continua servindo o resultado local).
func RunFactConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, log *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := q.Receive(ctx, queueURL, 5, 10)
		if err != nil {
			log.Error("pulsar: falha ao receber fatos de protocolo", "error", err)
			continue
		}
		for _, m := range msgs {
			var fact protocolFact
			if err := json.Unmarshal(m.Envelope.Payload, &fact); err == nil && fact.ProtocolID != "" {
				log.Debug("pulsar: fato de protocolo recebido", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID, "status", fact.Status)
				dest, derr := store.GetDestination(ctx, fact.TenantID)
				if derr == nil {
					deliveryID := idgen.New()
					if err := store.CreateDelivery(ctx, deliveryID, fact.ProtocolID, fact.EventID, dest.URL); err != nil {
						log.Error("pulsar: falha ao criar entrega", "trace_id", fact.TraceID, "error", err)
					} else {
						log.Info("pulsar: entrega agendada", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID, "delivery_id", deliveryID, "destination", dest.URL)
					}
				} else {
					log.Debug("pulsar: tenant sem destino de webhook cadastrado", "trace_id", fact.TraceID, "tenant_id", fact.TenantID)
				}
			}
			_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
		}
	}
}
