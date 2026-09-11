package pulsar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"os"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/queue"
)

type protocolFact struct {
	Representation      []byte                         `json:"representation"`
	CellID              string                         `json:"cell_id"`
	ProtocolID          string                         `json:"protocol_id"`
	TraceID             string                         `json:"trace_id,omitempty"`
	TenantID            string                         `json:"tenant_id"`
	ApplicationID       string                         `json:"application_id"`
	WebhookDestinations []dispatch.DestinationSnapshot `json:"webhook_destinations"`
	Status              string                         `json:"status"`
	FinalBody           json.RawMessage                `json:"final_body"`
	EventID             string                         `json:"event_id"`
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
			if json.Unmarshal(m.Envelope.Payload, &fact) != nil || m.Envelope.Type != "protocol.finalized" || m.Envelope.SchemaVersion != 1 || m.Envelope.Producer != "orbita" || fact.CellID != os.Getenv("CELL_ID") || fact.TenantID != m.Envelope.TenantID || fact.ApplicationID == "" || len(fact.WebhookDestinations) > 100 || len(fact.Representation) == 0 {
				sum := sha256.Sum256(m.RawBody)
				_, err := store.db.ExecContext(ctx, `INSERT INTO message_quarantine(consumer,body_sha256,body,reason) VALUES('pulsar',$1,$2,'invalid_final_envelope') ON CONFLICT(consumer,body_sha256) DO UPDATE SET occurrences=message_quarantine.occurrences+1,last_seen_at=clock_timestamp()`, hex.EncodeToString(sum[:]), m.RawBody)
				if err == nil {
					_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
				}
				continue
			}
			if err := store.ConserveFinal(ctx, m.Envelope.EventID, fact); err != nil {
				log.Error("pulsar final custody unavailable")
				continue
			}
			_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
		}
	}
}
