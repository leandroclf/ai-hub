# Explore — Fronteiras de autorização e custódia

Snapshot: f87ce33034ae29c9431b1910dcc6a633b545e330. Fontes lidas; hipóteses de concorrência são estáticas até ensaio.

## F-R5-01 (P0)
A rota pública e HMAC por conta foram implementados. Porém AuthenticateAccountCallback consulta operations antes de Verify. Para operação inexistente, o handler chama StoreOrphanAccountCallback, que só valida campos não vazios e grava via storeOrphan sem Verify. Assim, uma assinatura arbitrária com timestamp recente pode atingir a custódia órfã e 202, se o banco estiver disponível. A quota é global; variação da assinatura muda token_hash e identidade de dedupe. A reconciliação só seleciona órfãos com operação existente e prune não remove RECEIVED. Evidência estática de caminho; não foi executada exploração externa.

[hub/internal/cometa/handlers.go:206](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/handlers.go#L206), [hub/internal/cometa/custody.go:228](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/custody.go#L228), [hub/internal/cometa/custody.go:240](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/custody.go#L240)

## F-R5-02 (P0)
Offer agora tem cache de fallback, mas chama Atlas antes dele em toda requisição e usa cache após qualquer erro. Probe em httptest reproduziu retorno de oferta antiga após HTTP 403 offer_not_eligible. Suspensão/revogação explícita pode ser ocultada até vencer o cache; o caminho quente ainda paga a consulta remota.

[hub/internal/atlasclient/client.go:23](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlasclient/client.go#L23), [hub/internal/atlas/offers.go:202](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlas/offers.go#L202)

## F-R5-03 (P0)
WithTenantTx permanece sem chamada nos stores; RuntimeDSN é opt-in de tenant fixo. Script RLS continua revertendo fixtures antes da assertion negativa e apenas imprime a contagem dentro da transação. Zero após rollback não prova isolamento. Migração de compatibilidade permite hub e grants abrangem tabelas auxiliares sem cobertura completa. Endpoint capacity-domains consulta todos os domínios para integrations:read sem escopo por recurso explícito.

[hub/internal/platform/pg/pg.go:39](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/platform/pg/pg.go#L39), [hub/deploy/r2/tests/rls-runtime-proof.sh:17](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/rls-runtime-proof.sh#L17), [hub/migrations/core/0034_runtime_rls_migration_compat.sql:5](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/migrations/core/0034_runtime_rls_migration_compat.sql#L5), [hub/internal/cometa/handlers.go:65](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/handlers.go#L65)
