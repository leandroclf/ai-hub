# Explore: Capacidade, credenciais e latência

Snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee. Escopo brownfield v4/R2. Instruções: AGENTS.md e docs/openspec-docs/.

## F-R3-10
Capacity possui código/testes, mas Execute/requestPoll não adquirem concessão nem realimentam o controlador. Relatório R2 admite a desconexão.

[hub/internal/cometa/capacity.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/capacity.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/cometa/poller.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/poller.go)

## F-R3-11
Bearer resolve segredo antes do L1, mantém mutex do TokenCache durante Redis/OAuth e grava bearer token no Redis, contrariando a restrição R2.

[hub/internal/providerauth/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/providerauth/client.go)

## F-R3-12
NewClient cria Transport por submit/poll/OAuth. MaxConnsPerHost em transports distintos não limita consumo agregado e impede reaproveitamento eficiente.

[hub/internal/platform/egress](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/egress), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/cometa/poller.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/poller.go), [hub/internal/providerauth/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/providerauth/client.go)

Não confundir presença do mecanismo com qualificação integrada. Ver relatório R3 e evidência de testes.
