# Explore — Integridade de dados e reconciliação

Snapshot: f87ce33034ae29c9431b1910dcc6a633b545e330. Fontes lidas; hipóteses de concorrência são estáticas até ensaio.

## F-R5-08 (P0)
TransformJSON agora passa todos os probes anteriores. A perda reaparece depois: operationFact.ResponseBody é any e json.Unmarshal usa float64. Probe reproduziu 9007199254740993→9007199254740992 na desserialização/serialização desse fato. Consolidação de produto também usa any; frontend converte fact_id com Number. Correção do validador não protege essas fronteiras.

[hub/internal/orbita/factconsumer.go:25](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/factconsumer.go#L25), [hub/internal/orbita/factconsumer.go:35](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/factconsumer.go#L35), [hub/internal/orbita/product_store.go:289](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_store.go#L289), [hub/admin-ui/src/pages/FinancePage.tsx:58](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/FinancePage.tsx#L58)

## F-R5-09 (P0)
Harness agora evita reutilizar alvo e compara digests SQL; permanece s3 sync/list-objects-v2 e uma única consulta ao oráculo antes de PASS. Não comprova versões históricas referenciadas, pins/tombstones, retomada nem reconciliação financeira/externa após replay. Lista de tabelas comparadas não inclui operation_plans/steps. Evidência histórica de cópia não encerra restore do produto atualizado.

[hub/deploy/r2/tests/restore-reconciliation.sh:57](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/restore-reconciliation.sh#L57), [hub/deploy/r2/tests/restore-reconciliation.sh:63](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/restore-reconciliation.sh#L63), [hub/deploy/r2/tests/restore-reconciliation.sh:21](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/restore-reconciliation.sh#L21)

## F-R5-10 (P0)
Incidências SUBMITTED/STATUS, dedupe e ledger balanceado avançaram. Captura ainda muda estado da reserva sem apurar liberação da diferença para o valor real; SetWatermark aparece chamado em teste, sem produtor integrado de completude. Evento tardio insere finance_quarantine, apesar do comentário prometer disputa. São lacunas de fechamento operacional, não ausência de ledger.

[hub/internal/libra/store.go:126](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L126), [hub/internal/libra/store.go:337](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L337), [hub/internal/libra/settlement.go:231](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/settlement.go#L231), [hub/internal/libra/store.go:300](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L300)

## F-R5-16 (P0)
ProcessEnvelope valida e chama Quarantine para envelope inválido; Quarantine persiste apenas payload_hash. O consumidor apaga a mensagem após retorno nil. Portanto o conteúdo inválido pode deixar de ser recuperável depois do ACK. O avanço do worker Cometa, que conserva bytes, não foi aplicado à autoridade financeira.

[hub/internal/libra/store.go:331](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L331), [hub/internal/libra/consumers.go:58](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/consumers.go#L58)
