package libra

import (
	"context"
	"encoding/json"
	"log/slog"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/queue"
)

// placeholderProviderCost e um custo fixo por operacao usado nesta
// referencia local, pois a tabela de precos real do provedor (FIN-01)
// e uma decisao comercial pendente (P-03), nao modelada aqui.
const placeholderProviderCost = 0.20

type protocolFact struct {
	ProtocolID string `json:"protocol_id"`
	TraceID    string `json:"trace_id,omitempty"`
	TenantID   string `json:"tenant_id"`
	Status     string `json:"status"`
}

type operationFact struct {
	ProtocolID        string `json:"protocol_id"`
	TraceID           string `json:"trace_id,omitempty"`
	ProviderAccountID string `json:"provider_account_id"`
	Kind              string `json:"kind"`
}

// RunRevenueConsumer registra a receita do cliente quando um protocolo
// e finalizado com sucesso (FIN-04: "uma unidade elegivel apesar de
// multiplos passos") e captura/libera a reserva estrita conforme o
// resultado (FIN-06).
func RunRevenueConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, atlas *atlasclient.Client, log *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := q.Receive(ctx, queueURL, 5, 10)
		if err != nil {
			log.Error("libra: falha ao receber fatos de protocolo", "error", err)
			continue
		}
		for _, m := range msgs {
			var fact protocolFact
			if err := json.Unmarshal(m.Envelope.Payload, &fact); err == nil && fact.ProtocolID != "" {
				handleProtocolFact(ctx, store, atlas, fact, log)
			}
			_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
		}
	}
}

func handleProtocolFact(ctx context.Context, store *Store, atlas *atlasclient.Client, fact protocolFact, log *slog.Logger) {
	log.Debug("libra: fato de protocolo recebido", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID, "status", fact.Status)
	contract, err := atlas.Contract(ctx, fact.TenantID)
	if err != nil {
		log.Warn("libra: contrato nao encontrado para fato de protocolo", "trace_id", fact.TraceID, "tenant_id", fact.TenantID)
		return
	}
	switch fact.Status {
	case "SUCCEEDED", "PARTIALLY_SUCCEEDED":
		if err := store.RecordFact(ctx, fact.TenantID, fact.ProtocolID, "REVENUE", "product.success", contract.UnitPrice, "BRL"); err != nil {
			log.Error("libra: falha ao registrar receita", "trace_id", fact.TraceID, "error", err)
		} else {
			log.Info("libra: receita registrada", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID, "amount", contract.UnitPrice)
		}
		if contract.StrictBalance {
			_ = store.Capture(ctx, fact.ProtocolID)
			log.Debug("libra: reserva capturada", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID)
		}
	default: // FAILED, EXPIRED, CANCELLED
		// FIN-05: falha/expiracao nao gera receita do medidor de
		// sucesso; reserva estrita (se existir) e liberada.
		if contract.StrictBalance {
			_ = store.Release(ctx, fact.ProtocolID)
			log.Debug("libra: reserva liberada (sem sucesso)", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID)
		}
	}
}

// RunCostConsumer registra o custo do provedor quando Cometa observa
// um fato de operacao (FIN-04: "custo por operacao").
func RunCostConsumer(ctx context.Context, q *queue.Client, queueURL string, store *Store, log *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := q.Receive(ctx, queueURL, 5, 10)
		if err != nil {
			log.Error("libra: falha ao receber fatos de operacao", "error", err)
			continue
		}
		for _, m := range msgs {
			var fact operationFact
			if err := json.Unmarshal(m.Envelope.Payload, &fact); err == nil && fact.ProtocolID != "" && fact.Kind == "SUCCEEDED" {
				if err := store.RecordFact(ctx, "", fact.ProtocolID, "COST", "provider.operation", placeholderProviderCost, "BRL"); err != nil {
					log.Error("libra: falha ao registrar custo", "trace_id", fact.TraceID, "error", err)
				} else {
					log.Info("libra: custo registrado", "trace_id", fact.TraceID, "protocol_id", fact.ProtocolID, "amount", placeholderProviderCost)
				}
			}
			_ = q.Delete(ctx, queueURL, m.ReceiptHandle)
		}
	}
}
