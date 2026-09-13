# Explore: Isolamento e custódia autenticada

Snapshot a540b40007fe6b8ed523e17afe00e96ff8f8ad50. Leitura direta do delta R5→R6.

## F-R6-01 · P0 · ANALISE_ESTATICA
A migração 0049 remove o bypass nominal nas tabelas core contempladas. WithTenantTx continua sem chamada produtiva, os DSNs Compose usam hub e o contexto opcional continua fixo por workload. A prova rls-runtime-proof.sh continua executando ROLLBACK das fixtures antes das assertions negativas; a contagem dentro da transação só é impressa. Portanto, a execução histórica do script não prova isolamento de dados existentes nem adoção pelo runtime.

Fontes: [hub/migrations/core/0049_strict_runtime_rls.sql:5](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/migrations/core/0049_strict_runtime_rls.sql#L5), [hub/internal/platform/pg/pg.go:39](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/platform/pg/pg.go#L39), [hub/deploy/r2/tests/rls-runtime-proof.sh:17](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/rls-runtime-proof.sh#L17), [hub/deploy/r2/compose.yaml:203](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/compose.yaml#L203)

Vínculo anterior: R5-SEG-03.

## F-R6-02 · P0 · ANALISE_ESTATICA
A assinatura é agora verificada antes da custódia órfã e a identidade não depende do timestamp: preservar a correção. A reconciliação ainda revalida com CALLBACK_INGRESS_KEY atual, não com uma versão de chave congelada. O lock advisory de storeOrphan continua global, e três falhas de apply podem converter recibo em REJECTED. Órfãos sem operação não são selecionados pelo JOIN de reconciliação.

Fontes: [hub/internal/cometa/custody.go:228](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L228), [hub/internal/cometa/custody.go:195](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L195), [hub/internal/cometa/custody.go:385](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L385), [hub/internal/cometa/custody.go:411](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L411)

Vínculo anterior: R5-SEG-01.
