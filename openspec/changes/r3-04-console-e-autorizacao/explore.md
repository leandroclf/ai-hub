# Explore: Console e autorização

Snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee. Escopo brownfield v4/R2. Instruções: AGENTS.md e docs/openspec-docs/.

## F-R3-13
Leitura administrativa no mesmo tenant aceita protocols:read sem papel administrativo específico/MFA e lista todas as aplicações. Entre tenants já existe verificação MFA/papel; não se trata de afirmar bypass global irrestrito.

[hub/internal/orbita/admin.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admin.go), [hub/internal/platform/auth/auth.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/auth/auth.go)

## F-R3-14
Seleção usa protocol_id antes de delivery_id. Menu SLA aponta para rota não registrada; botão de reconciliação POST aponta para handler de leitura.

[hub/admin-ui/src/pages/OperationsPage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/OperationsPage.tsx), [hub/internal/pulsar/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/pulsar/handlers.go), [hub/internal/orbita/admin.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admin.go)

## F-R3-15
UI envia tenant_id/prepared_by a decoder estrito incompatível e datas YYYY-MM-DD para time.Time. Ações não mantêm tenant selecionado na query e criam idempotency key nova em cada execução.

[hub/admin-ui/src/pages/FinancePage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/FinancePage.tsx), [hub/internal/libra/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/handlers.go)

## F-R3-16
Editor de modos reduz array a seleção única; faltam campos condicionais de autenticação; lookups limitados à primeira centena. datetime-local corta UTC e o reinterpreta no fuso local ao salvar.

[hub/admin-ui/src/pages/CatalogPage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/CatalogPage.tsx), [hub/admin-ui/src](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src)

Não confundir presença do mecanismo com qualificação integrada. Ver relatório R3 e evidência de testes.
