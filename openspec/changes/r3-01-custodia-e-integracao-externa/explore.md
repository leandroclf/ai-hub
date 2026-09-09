# Explore: Custódia e integração externa

Snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee. Escopo brownfield v4/R2. Instruções: AGENTS.md e docs/openspec-docs/.

## F-R3-01
O executor recusa AdapterID diferente de synthetic-provider. Configurar endpoint não cria integração executável. Config não recebe APIKeyHeader, embora Apply o exija.

[hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/cometa/poller.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/poller.go), [hub/internal/atlasclient/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlasclient/client.go)

## F-R3-02
Callback chama ApplyExternalObservation sem retorno de erro e responde 200; operação desconhecida, erro de consulta e observação após terminal não têm confirmação de recibo nesse caminho. A rota está sob JWT interno do Hub, não sob autenticação específica da conta provedora.

[hub/internal/cometa/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/handlers.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/cmd/cometa/main.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/cmd/cometa/main.go)

## F-R3-03
Após SUBMITTING, replay pode devolver UNKNOWN durável sem recuperador geral desse estado. Envio não transmite chave idempotente homologada. Consulta inicial de replay antecede validação completa de hash/aplicação.

[hub/internal/cometa/custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/custody.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/orbita/intents.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/intents.go)

## F-R3-04
RetryDeadline deriva do aceite e limita polling; provider_mode mantém polling/callback exclusivos. ADR R2 registra contraexemplo de commit posterior ao deadline e requisito não qualificado.

[hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/cometa/polling_custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/polling_custody.go), [hub/internal/orbita/finalize.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/finalize.go), [docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md)

## F-R3-05
Bootstraps independentes não estabelecem barreira de todas as assinaturas obrigatórias antes do relay. Quarentena financeira conserva hash/motivo sem payload para replay e então confirma a mensagem.

[hub/internal/queue/bootstrap.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/queue/bootstrap.go), [hub/internal/queue/queue.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/queue/queue.go), [hub/internal/libra/consumers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/consumers.go), [hub/internal/libra/store.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/store.go)

Não confundir presença do mecanismo com qualificação integrada. Ver relatório R3 e evidência de testes.
