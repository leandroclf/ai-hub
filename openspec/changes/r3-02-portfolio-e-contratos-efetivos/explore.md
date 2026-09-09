# Explore: Portfólio e contratos efetivos

Snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee. Escopo brownfield v4/R2. Instruções: AGENTS.md e docs/openspec-docs/.

## F-R3-06
Provas executadas: TransformJSON altera 9007199254740993 para 9007199254740992 e aceita INVALID contra enum [OK]. Usa float64 e validação parcial.

[hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go), [hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go)

## F-R3-07
Admissão usa tempos/modos do target sem composição efetiva de overrides. service_version recebido não participa da resolução. Finalizer usa FinalBody fixo sem OutputMapping.

[hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/orbita/finalize.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/finalize.go), [hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go)

## F-R3-08
Catálogo contém steps, mas admissão produz um comando com StepID igual ao protocolo; não há executor integrado do DAG/agregação nesse caminho.

[hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/atlas/catalog.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/catalog.go), [hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go)

## F-R3-09
ResolveOffer recusa tenant com mais de 100 ofertas antes de filtrar aplicação/serviço. Caminho Offer consulta Atlas em cada requisição.

[hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go), [hub/internal/atlasclient/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlasclient/client.go), [hub/internal/atlas/catalog.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/catalog.go)

Não confundir presença do mecanismo com qualificação integrada. Ver relatório R3 e evidência de testes.
