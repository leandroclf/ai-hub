package orbita

import (
	"context"
	"log/slog"

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
	ProtocolID string    `json:"protocol_id"`
	TenantID   string    `json:"tenant_id"`
	Status     string    `json:"status"`
	FinalBody  FinalBody `json:"final_body"`
	EventID    string    `json:"event_id"`
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
func (f *Finalizer) Finalize(ctx context.Context, tenantID, protocolID string, expectedVersion int, status Status, body FinalBody, reason string) (bool, error) {
	eventID := idgen.New()
	body.Status = string(status)

	applied, err := f.store.Finalize(ctx, FinalizeParams{
		ProtocolID: protocolID, ExpectedVersion: expectedVersion, Status: status,
		FinalBody: body, TerminalReason: reason, FinalEventID: eventID,
	}, func(tx *sqlTx) error {
		fact := ProtocolFinalizedFact{ProtocolID: protocolID, TenantID: tenantID, Status: string(status), FinalBody: body, EventID: eventID}
		return outbox.Enqueue(ctx, tx.Tx(), "protocol", protocolID, "protocol.finalized", fact)
	})
	if err != nil {
		return false, err
	}
	if !applied {
		f.log.Info("finalizacao ignorada: protocolo ja possuia transicao terminal", "protocol_id", protocolID)
	}
	return applied, nil
}
