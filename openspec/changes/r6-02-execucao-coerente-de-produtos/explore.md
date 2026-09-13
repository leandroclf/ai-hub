# Explore: Execução coerente de produtos e prazos

Snapshot a540b40007fe6b8ed523e17afe00e96ff8f8ad50. Leitura direta do delta R5→R6.

## F-R6-03 · P0 · PROBE_REPRODUZIDO
BuildProductPlan agora gera hash do filho, mas stepSnapshot só substitui SelectedRoute; Account, Binding, contratos e perfil permanecem da oferta base, e command := base conserva ProviderAccountID e EconomicSnapshot. Probe reproduziu rota B com conta A, binding B/A e comando A. Hash válido certifica bytes, não coerência semântica.

Fontes: [hub/internal/orbita/product_plan.go:126](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_plan.go#L126), [hub/internal/orbita/product_plan.go:83](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_plan.go#L83), [hub/internal/cometa/executor.go:199](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L199)

Vínculo anterior: R5-EXE-03.

## F-R6-04 · P0 · ANALISE_ESTATICA
claimIntentMode retorna Expired e RunIntentPublisher evita envio quando true. RunDirectRecovery chama o mesmo claim, mas executa DispatchDirect sem verificar Expired; ClaimDirectIntent também devolve o campo sem consumidor no handler. O guard do modo QUEUED não fecha a fronteira DIRECT.

Fontes: [hub/internal/orbita/intents.go:208](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L208), [hub/internal/orbita/intents.go:363](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L363), [hub/internal/orbita/handlers.go:337](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/handlers.go#L337)

Vínculo anterior: R5-EXE-01.

## F-R6-05 · P0 · ANALISE_ESTATICA
O DAG reverso e um prazo próprio foram adicionados. Os claims ainda exigem p.status NOT IN estados terminais mesmo quando continue_after_client_deadline=true. Portanto, a exceção de deadline não autoriza continuar após EXPIRED/CANCELLED. O TTL mínimo de compensação também é fixado em 30 segundos, sem política independente demonstrada.

Fontes: [hub/internal/orbita/intents.go:50](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L50), [hub/internal/orbita/product_store.go:264](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_store.go#L264), [hub/migrations/core/0048_compensation_independent_deadline.sql:4](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/migrations/core/0048_compensation_independent_deadline.sql#L4)

Vínculo anterior: R5-EXE-04.

## F-R6-06 · P0 · ANALISE_ESTATICA
O claim ganhou lock do plano e running_count: não repetir o antigo achado de contagem desprotegida. CompleteIntent ainda atualiza command_intents em uma transação implícita e, depois, redefine a etapa e decrementa o plano em outro Exec. Uma interrupção entre essas escritas pode deixar estado e contador divergentes, sobretudo se o intent já ficou EXPIRED e não poderá ser reclamado.

Fontes: [hub/internal/orbita/intents.go:295](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L295), [hub/internal/orbita/intents.go:336](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L336), [hub/internal/orbita/intents.go:180](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L180)

Vínculo anterior: R5-EXE-02.
