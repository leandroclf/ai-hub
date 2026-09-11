package orbita

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/outbox"
	"ai-hub/hub/internal/platform/idgen"
)

// FinalBody e a representacao unica usada tanto pelo GET quanto pelo
// webhook (COM-05): mesmo schema, mesmos campos, para o mesmo
// protocolo/versao/contrato de cliente.
type FinalBody struct {
	ProtocolID    string `json:"protocol_id"`
	Status        string `json:"status"`
	ResultVersion int    `json:"result_version"`
	Result        any    `json:"result,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// ProtocolFinalizedFact e o payload publicado na outbox (aggregate_type
// "protocol") quando um protocolo e finalizado — consumido por Pulsar
// (webhook) e Libra (receita), conforme COM-01.
type ProtocolFinalizedFact struct {
	EconomicSnapshot    json.RawMessage                `json:"economic_snapshot"`
	EvidenceID          string                         `json:"evidence_id"`
	OccurredAt          time.Time                      `json:"occurred_at"`
	Representation      []byte                         `json:"representation"`
	CellID              string                         `json:"cell_id"`
	ProtocolID          string                         `json:"protocol_id"`
	TraceID             string                         `json:"trace_id,omitempty"`
	TenantID            string                         `json:"tenant_id"`
	ApplicationID       string                         `json:"application_id"`
	WebhookDestinations []dispatch.DestinationSnapshot `json:"webhook_destinations"`
	Status              string                         `json:"status"`
	FinalBody           FinalBody                      `json:"final_body"`
	EventID             string                         `json:"event_id"`
}

// Finalizer aplica a transicao terminal unica do protocolo (EXE-11) e
// publica o fato final na mesma transacao (COM-03).
type Finalizer struct {
	store *Store
	log   *slog.Logger
}

// NewFinalizer cria um Finalizer.
func NewFinalizer(store *Store, log *slog.Logger) *Finalizer {
	return &Finalizer{store: store, log: log}
}

// Finalize tenta a transicao terminal; retorna false se o protocolo ja
// havia sido finalizado por outro caminho (ex.: deadline concorrente),
// preservando a regra de uma unica transicao terminal serializavel.
func (f *Finalizer) Finalize(ctx context.Context, traceID, tenantID, protocolID string, expectedVersion int, status Status, body FinalBody, reason string) (bool, error) {
	eventID := idgen.New()
	body.Status = string(status)
	body.ProtocolID = protocolID
	body.ResultVersion = 1

	applied, err := f.store.Finalize(ctx, FinalizeParams{
		ProtocolID: protocolID, ExpectedVersion: expectedVersion, Status: status,
		FinalBody: body, TerminalReason: reason, FinalEventID: eventID,
	}, func(tx *sqlTx) error {
		fact := ProtocolFinalizedFact{ProtocolID: protocolID, TraceID: traceID, TenantID: tenantID, Status: string(status), FinalBody: body, EventID: eventID}
		fact.Representation = tx.Representation
		fact.EvidenceID = eventID
		// A command may legitimately omit an economic snapshot (for example
		// a synthetic/provider qualification request).  PostgreSQL returns
		// NULL for the JSON projection in that case, while json.RawMessage
		// is not a nullable database/sql destination.  Normalize the absence
		// to an empty object so finalization remains durable and the outbox
		// fact keeps a valid JSON contract.
		var commandRaw []byte
		if err := tx.Tx().QueryRowContext(ctx, `SELECT p.cell_id,p.application_id,COALESCE(i.command->'economic_snapshot','{}'::jsonb),i.command,clock_timestamp() FROM protocols p JOIN command_intents i ON i.command_id=p.command_id WHERE p.protocol_id=$1 AND p.tenant_id=$2`, protocolID, tenantID).Scan(&fact.CellID, &fact.ApplicationID, &fact.EconomicSnapshot, &commandRaw, &fact.OccurredAt); err != nil {
			return err
		}
		var command dispatch.Command
		if err := json.Unmarshal(commandRaw, &command); err != nil {
			return err
		}
		fact.WebhookDestinations = command.WebhookDestinations
		return outbox.Enqueue(ctx, tx.Tx(), "protocol", protocolID, "protocol.finalized", fact)
	})
	if errors.Is(err, ErrResultLate) {
		return f.Finalize(ctx, traceID, tenantID, protocolID, expectedVersion, StatusExpired, FinalBody{ProtocolID: protocolID, ErrorCode: "SLA_EXCEEDED", ErrorMessage: "prazo do cliente expirado"}, "REJECTED_LATE_SLA")
	}
	if err != nil {
		return false, err
	}
	if !applied {
		f.log.Info("finalizacao ignorada: protocolo ja possuia transicao terminal", "trace_id", traceID, "protocol_id", protocolID)
		return false, nil
	}
	f.log.Info("protocolo finalizado", "trace_id", traceID, "protocol_id", protocolID, "status", status, "reason", reason)
	return applied, nil
}
