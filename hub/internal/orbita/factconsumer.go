package orbita

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"ai-hub/hub/internal/queue"
)

type operationFact struct {
	ProtocolID    string `json:"protocol_id"`
	TenantID      string `json:"tenant_id"`
	ApplicationID string `json:"application_id"`
	CellID        string `json:"cell_id"`
	OperationID   string `json:"operation_id"`
	EvidenceID    string `json:"evidence_id"`
	TraceID       string `json:"trace_id,omitempty"`
	Kind          string `json:"kind"`
	ResponseBody  any    `json:"response_body,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// ConsumeOperationFact acknowledges only durable quarantine, a conserved
// nonterminal observation, or confirmed terminal consolidation. A lost ACK
// replays the same inbox identity and cannot replace the original observation.
func (s *Store) ConsumeOperationFact(ctx context.Context, m queue.ReceivedMessage, finalizer *Finalizer) error {
	e := m.Envelope
	var fact operationFact
	if json.Unmarshal(e.Payload, &fact) != nil || e.EventID == "" || e.Type != "operation.observed" || e.SchemaVersion != 1 || e.Producer != "cometa" || fact.ProtocolID == "" || fact.ProtocolID != e.ProtocolID || fact.TenantID == "" || fact.TenantID != e.TenantID || fact.ApplicationID == "" || fact.CellID == "" || fact.CellID != os.Getenv("CELL_ID") || fact.OperationID == "" || fact.EvidenceID == "" || (fact.Kind != "UNKNOWN" && fact.Kind != "SUCCEEDED" && fact.Kind != "FAILED") {
		return s.quarantineFact(ctx, m, "invalid_operation_envelope")
	}
	p, err := s.Get(ctx, fact.TenantID, fact.ProtocolID)
	if errors.Is(err, ErrProtocolNotFound) {
		return s.quarantineFact(ctx, m, "unknown_protocol_or_tenant")
	}
	if err != nil {
		return err
	}
	if p.ApplicationID != fact.ApplicationID || p.CellID != fact.CellID || p.CommandID != fact.OperationID {
		return s.quarantineFact(ctx, m, "operation_scope_mismatch")
	}
	raw, err := json.Marshal(e)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	_, err = s.db.ExecContext(ctx, `INSERT INTO orbita_fact_inbox(event_id,body_sha256,envelope,protocol_id,tenant_id,application_id,cell_id) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(event_id) DO NOTHING`, e.EventID, hash, raw, p.ProtocolID, p.TenantID, p.ApplicationID, p.CellID)
	if err != nil {
		return err
	}
	var original, disposition string
	if err = s.db.QueryRowContext(ctx, `SELECT body_sha256,disposition FROM orbita_fact_inbox WHERE event_id=$1`, e.EventID).Scan(&original, &disposition); err != nil {
		return err
	}
	if original != hash {
		return s.quarantineFact(ctx, m, "event_identity_conflict")
	}
	if disposition != "RECEIVED" {
		return nil
	}
	disposition = "OBSERVED"
	if fact.Kind != "UNKNOWN" {
		if !IsTerminal(p.Status) {
			status := StatusFailed
			if fact.Kind == "SUCCEEDED" {
				status = StatusSucceeded
			}
			_, err = finalizer.Finalize(ctx, fact.TraceID, p.TenantID, p.ProtocolID, p.Version, status, FinalBody{Result: fact.ResponseBody, ErrorMessage: fact.ErrorMessage}, "")
			if err != nil {
				return err
			}
		}
		// A version conflict alone is not proof of a terminal effect.
		p, err = s.Get(ctx, fact.TenantID, fact.ProtocolID)
		if err != nil {
			return err
		}
		if !IsTerminal(p.Status) {
			return errors.New("terminal consolidation pending")
		}
		disposition = "APPLIED"
	}
	_, err = s.db.ExecContext(ctx, `UPDATE orbita_fact_inbox SET disposition=$2,processed_at=clock_timestamp() WHERE event_id=$1 AND disposition='RECEIVED'`, e.EventID, disposition)
	return err
}

func (s *Store) quarantineFact(ctx context.Context, m queue.ReceivedMessage, reason string) error {
	raw := m.RawBody
	if len(raw) == 0 {
		var err error
		raw, err = json.Marshal(m.Envelope)
		if err != nil {
			return err
		}
	}
	sum := sha256.Sum256(raw)
	_, err := s.db.ExecContext(ctx, `INSERT INTO message_quarantine(consumer,body_sha256,body,reason) VALUES('orbita',$1,$2,$3) ON CONFLICT(consumer,body_sha256) DO UPDATE SET occurrences=message_quarantine.occurrences+1,last_seen_at=clock_timestamp()`, hex.EncodeToString(sum[:]), raw, reason)
	return err
}

func RunOperationFactConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, finalizer *Finalizer, log *slog.Logger) {
	for ctx.Err() == nil {
		msgs, err := q.Receive(ctx, queueURL, 5, 10)
		if err != nil {
			log.Error("orbita: operation facts receive failed")
			continue
		}
		for _, m := range msgs {
			if err = store.ConsumeOperationFact(ctx, m, finalizer); err != nil {
				log.Error("orbita: operation fact custody unavailable")
				continue
			}
			if err = q.Delete(ctx, queueURL, m.ReceiptHandle); err != nil {
				log.Warn("orbita: operation fact ACK unavailable; redelivery is safe")
			}
		}
	}
}
