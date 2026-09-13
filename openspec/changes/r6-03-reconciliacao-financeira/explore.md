# Explore: Reconciliação financeira recuperável

Snapshot a540b40007fe6b8ed523e17afe00e96ff8f8ad50. Leitura direta do delta R5→R6.

## F-R6-07 · P0 · ANALISE_ESTATICA
Payload e endpoints de replay foram adicionados. ProcessEnvelope retorna nil também quando Quarantine grava com sucesso; o handler então marca REPLAYED. Se o payload ainda inválido recria a mesma identidade, ON CONFLICT DO NOTHING conserva a linha e MarkQuarantineReplayed a retira da lista de pendências sem aplicação. Fato tardio é armazenado como EconomicEvent, enquanto o endpoint espera queue.Envelope.

Fontes: [hub/internal/libra/consumers.go:33](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/consumers.go#L33), [hub/internal/libra/handlers.go:419](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/handlers.go#L419), [hub/internal/libra/store.go:389](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L389), [hub/internal/libra/store.go:420](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L420)

Vínculo anterior: R5-DAD-04.

## F-R6-08 · P0 · ANALISE_ESTATICA
A captura efetiva agora é chamada por ApplyEvent e registra diferença da reserva. SetWatermark permanece chamado apenas por teste; assim o mecanismo de fechamento carece de produtor durável de completude. Fatos tardios ainda entram em quarentena, não demonstram disputa/ajuste completo. A correção de captura deve ser preservada e testada com múltiplos fatos/ordem/escala decimal.

Fontes: [hub/internal/libra/store.go:144](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L144), [hub/internal/libra/store.go:453](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L453), [hub/internal/libra/settlement.go:217](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/settlement.go#L217)

Vínculo anterior: R5-DAD-03.
