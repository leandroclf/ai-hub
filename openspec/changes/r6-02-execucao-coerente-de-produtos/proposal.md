# Proposal: Execução coerente de produtos e prazos
## Change ID
r6-02-execucao-coerente-de-produtos
## Status
Draft — próxima rodada de implementação, solicitada pelo responsável do projeto.
## Why
BuildProductPlan agora gera hash do filho, mas stepSnapshot só substitui SelectedRoute; Account, Binding, contratos e perfil permanecem da oferta base, e command := base conserva ProviderAccountID e EconomicSnapshot. Probe reproduziu rota B com conta A, binding B/A e comando A. Hash válido certifica bytes, não coerência semântica.
claimIntentMode retorna Expired e RunIntentPublisher evita envio quando true. RunDirectRecovery chama o mesmo claim, mas executa DispatchDirect sem verificar Expired; ClaimDirectIntent também devolve o campo sem consumidor no handler. O guard do modo QUEUED não fecha a fronteira DIRECT.
O DAG reverso e um prazo próprio foram adicionados. Os claims ainda exigem p.status NOT IN estados terminais mesmo quando continue_after_client_deadline=true. Portanto, a exceção de deadline não autoriza continuar após EXPIRED/CANCELLED. O TTL mínimo de compensação também é fixado em 30 segundos, sem política independente demonstrada.
O claim ganhou lock do plano e running_count: não repetir o antigo achado de contagem desprotegida. CompleteIntent ainda atualiza command_intents em uma transação implícita e, depois, redefine a etapa e decrementa o plano em outro Exec. Uma interrupção entre essas escritas pode deixar estado e contador divergentes, sobretudo se o intent já ficou EXPIRED e não poderá ser reclamado.
## What Changes
- R6-EXE-01 — Snapshot de etapa coerente em todas as identidades
- R6-EXE-02 — Expiração bloqueia também recuperação DIRECT
- R6-EXE-03 — Compensação progride após final público
- R6-EXE-04 — Liberação de slot e intent na mesma transação
## Context and Problem
Implementação atual a540b40007fe6b8ed523e17afe00e96ff8f8ad50; mudanças incrementais em fronteiras comprovadamente incompletas. Fontes e distinção entre reprodução e hipótese em explore.md.
## Goals
Cumprir os requisitos e cenários do delta, preservando correções já implementadas na R5.
## Non-Goals
Reescrita da plataforma, novo broker sem justificativa, deploy remoto ou aprovação comercial fictícia.
## Users / Actors Impacted
Clientes multi-tenant, provedores e administradores nominais. Responsável técnico: Core e Integrações.
## Scope
Requisitos listados, testes de falha, integração real, contratos, migrações e evidências. Os vínculos R5 continuam normativos.
## Business Rules
Custódia antes do aceite; efeito externo idempotente ou reconciliado; UUIDv7; final público imutável; isolamento entre tenants; valores financeiros exatos. Sem promessa de disponibilidade absoluta diante de falha de todas as autoridades duráveis.
## Affected Capabilities
r6-02-execucao-coerente-de-produtos
## Impact
### Code
Arquivos citados em explore.md e seus chamadores reais.
### Data
Migrações aditivas quando necessárias; validar em base populada; preservar identidades e custódia existente.
### APIs / Contracts
Contratos versionados com erro explícito; igualdade da representação final GET/webhook preservada.
### Operations
Rollback compatível com dados novos, métricas e runbook incluídos.
## Risks and Mitigations
Risco de correção isolada sem wiring: teste pela API/worker produtivo com oráculo. Risco de regressão histórica: matriz integral e qualificação por impacto.
## Success Criteria
Todos os cenários deste change têm resultado individual e evidência do artefato candidato; nenhum P0 vinculado encerra por build isolado.
## Assumptions / Open Questions
Valores de SLA/capacidade/retenção são parametrizados; decisões externas D-01…D-07 e T-R2-01 não são presumidas aprovadas. Fixtures locais permitem implementar mecanismos técnicos.
