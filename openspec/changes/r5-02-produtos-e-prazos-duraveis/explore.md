# Explore — Produtos, prazos e compensação duráveis

Snapshot: f87ce33034ae29c9431b1910dcc6a633b545e330. Fontes lidas; hipóteses de concorrência são estáticas até ensaio.

## F-R5-04 (P0)
A primeira falha/horizonte persistem, mas ClaimIntent e ClaimDirectIntent ainda não excluem retry_until vencido; CompleteIntent avalia depois de I/O e DELIVERED prevalece. O polling já foi corrigido para priorizar StepDeadline em vez do TTL inicial e essa correção deve ser preservada. UI ainda explica TTL desde aceite, divergindo da regra de primeira falha. Fencing de submissão foi implementado e deve ser preservado; ele não elimina a diferença entre esses relógios.

[hub/internal/orbita/intents.go:135](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L135), [hub/internal/orbita/intents.go:255](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L255), [hub/internal/cometa/polling_custody.go:55](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/polling_custody.go#L55), [hub/admin-ui/src/pages/CatalogPage.tsx:6](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/CatalogPage.tsx#L6)

## F-R5-05 (P0)
Produtos agora têm plano/etapas e execução HTTP comprovada em laboratório. Porém o claim conta RUNNING/WAITING_PROVIDER sem cercar a linha do plano e atualiza etapa em outra transação: dois publishers podem observar vaga simultaneamente. Se crash ocorre após RUNNING antes de publicar, re-claim depende da mesma contagem e pode ficar bloqueado no próprio max_parallel. O helper antigo ExecuteDAG segue com falha de cancelamento, mas não é o runtime persistido; não confundir as implementações.

[hub/internal/orbita/intents.go:135](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L135), [hub/internal/orbita/product_store.go:55](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_store.go#L55), [hub/internal/atlas/executor.go:34](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlas/executor.go#L34)

## F-R5-06 (P0)
BuildProductPlan copia o snapshot do produto, troca Target e preserva SelectedRoute/Account/Binding e EconomicSnapshot de base. Isso executa duas etapas no mesmo provedor sintético, mas não demonstra composição de serviços com provedores, credenciais e contratos de compra distintos. Hash do filho é esvaziado. O plano guarda consolidation, porém consolidação final usa formato fixo de steps.

[hub/internal/orbita/product_plan.go:42](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_plan.go#L42), [hub/internal/orbita/product_plan.go:77](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_plan.go#L77), [hub/internal/atlas/offers.go:35](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlas/offers.go#L35)

## F-R5-07 (P0)
Compensações são persistidas, avanço sobre a R3. Porém são inseridas READY com depends_on vazio e herdam deadlines do comando original; o publisher exige protocolo não terminal e client_deadline futura. Assim, ordem reversa causal não está representada e compensação de obrigação externa pode deixar de executar depois de encerrar o atendimento.

[hub/internal/orbita/product_store.go:221](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_store.go#L221), [hub/internal/orbita/product_plan.go:116](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_plan.go#L116), [hub/internal/orbita/intents.go:48](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L48)

## F-R5-17 (P0)
Bootstrap assíncrono preserva disponibilidade HTTP durante falha de broker, mas cada processo confirma apenas sua parte. Órbita cria tópico de fatos finais e inicia relay sem comprovar assinaturas de Pulsar/Libra; Cometa publica fatos externos sem barreira comum de todas as assinaturas obrigatórias. Em ambiente limpo com startup fora de ordem, publicação SNS pode ser confirmada antes de assinaturas necessárias. Risco estático a ensaiar; tópicos já provisionados escondem essa janela.

[hub/cmd/orbita/main.go:67](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/cmd/orbita/main.go#L67), [hub/cmd/cometa/main.go:72](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/cmd/cometa/main.go#L72), [hub/cmd/libra/main.go:58](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/cmd/libra/main.go#L58), [hub/internal/queue/bootstrap.go:12](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/queue/bootstrap.go#L12)
