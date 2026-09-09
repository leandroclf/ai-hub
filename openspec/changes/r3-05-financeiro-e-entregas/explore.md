# Explore: Financeiro e entregas

Snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee. Escopo brownfield v4/R2. Instruções: AGENTS.md e docs/openspec-docs/.

## F-R3-17
Reserva exata existe, mas captura não ajusta seu valor ao valor efetivo. DENY por franquia ocorre na incidência após execução externa.

[hub/internal/libra/store.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/store.go), [hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/contracts/economics/publication.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/contracts/economics/publication.go)

## F-R3-18
Fatos não transportam todas as unidades/identidades ATTEMPT; STATUS/FETCH carecem de incidência integrada. SetWatermark tem chamada em teste sem fluxo produtor de completude; ciclo de disputa/fechamento não está completo.

[hub/internal/cometa/custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/custody.go), [hub/internal/cometa/polling_custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/polling_custody.go), [hub/internal/libra/store.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/store.go), [hub/internal/libra/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/handlers.go)

## F-R3-19
Pulsar escolhe versões ACTIVE atuais por tenant no consumo, em vez de destinos/aplicação congelados na admissão.

[hub/internal/pulsar/custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/pulsar/custody.go), [hub/internal/orbita/finalize.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/finalize.go), [hub/internal/pulsar/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/pulsar/handlers.go)

Não confundir presença do mecanismo com qualificação integrada. Ver relatório R3 e evidência de testes.
