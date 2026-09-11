package libra

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/queue"
)

// ProcessEnvelope returns only after financial custody or durable quarantine.
// Contract changes in Atlas cannot affect historical events: no runtime lookup.
func (s *Store) ProcessEnvelope(ctx context.Context, consumer string, env queue.Envelope) error {
	raw, _ := json.Marshal(env)
	var fact EconomicEvent
	reason := ""
	if json.Unmarshal(env.Payload, &fact) != nil {
		reason = "INVALID_JSON"
	} else if env.EventID == "" || env.SchemaVersion != 1 || env.TenantID == "" || fact.TenantID != env.TenantID || fact.ProtocolID != env.ProtocolID {
		reason = "INVALID_ENVELOPE_IDENTITY"
	} else if (consumer == "cost" && env.Producer != "cometa") || (consumer == "revenue" && env.Producer != "orbita") {
		reason = "INVALID_PRODUCER"
	} else if fact.EconomicSnapshot.Validate() != nil {
		reason = "MISSING_OR_INVALID_SNAPSHOT"
	} else if fact.EvidenceID == "" || fact.OccurredAt.IsZero() {
		reason = "MISSING_ECONOMIC_EVIDENCE"
	} else if fact.EconomicKind != "" && !validEconomicKind(fact.EconomicKind) {
		reason = "INVALID_ECONOMIC_INCIDENCE"
	}
	if reason != "" {
		return s.Quarantine(ctx, consumer, env.EventID, reason, raw)
	}
	// The operational fact and the economic incidence are intentionally
	// separate. UNKNOWN is a valid operational state for Orbita, while a
	// durable SUBMIT acceptance or STATUS poll is still billable by Libra.
	if fact.EconomicKind != "" {
		fact.Kind = fact.EconomicKind
	}
	return s.ApplyEvent(ctx, consumer, env.EventID, fact)
}

func validEconomicKind(kind string) bool {
	switch kind {
	case "SUBMITTED", "STATUS", "FETCH", "SUCCEEDED", "PARTIALLY_SUCCEEDED", "FAILED":
		return true
	default:
		return false
	}
}
func RunRevenueConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, _ *atlasclient.Client, log *slog.Logger) {
	runConsumer(ctx, q, queueURL, store, "revenue", log)
}
func RunCostConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, log *slog.Logger) {
	runConsumer(ctx, q, queueURL, store, "cost", log)
}
func runConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, consumer string, log *slog.Logger) {
	for ctx.Err() == nil {
		msgs, e := q.Receive(ctx, queueURL, 5, 10)
		if e != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error("finance receive failed", "consumer", consumer, "error", e)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}
		for _, m := range msgs {
			if e = store.ProcessEnvelope(ctx, consumer, m.Envelope); e != nil {
				log.Error("finance custody failed; message remains unacknowledged", "consumer", consumer, "event_id", m.Envelope.EventID, "error", e)
				continue
			}
			if e = q.Delete(ctx, queueURL, m.ReceiptHandle); e != nil {
				log.Warn("finance acknowledgment failed; replay is idempotent", "consumer", consumer, "error", e)
			}
		}
	}
}
