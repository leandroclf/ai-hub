# Proposal: Console administrativo operável
## Change ID
r6-05-console-operavel
## Status
Draft — próxima rodada de implementação, solicitada pelo responsável do projeto.
## Why
Lookup agora busca no servidor com limite 50, resolvendo o antigo teto de 1000 para campos com pesquisa. StepsEditor e RoutesEditor recebem apenas essa primeira lista e não expõem pesquisa/cursor próprios; serviços, contas e bindings fora dos primeiros 50 ficam inalcançáveis nesses editores. Referência selecionada fora da página também não é carregada pontualmente.
Guards runtime foram introduzidos, mas são opcionais e usados sobretudo nas listagens; detalhe, save e command ainda fazem cast sem guard. mappingErrors é local ao StepsEditor e não participa do bloqueio save do pai: JSON de mapping inválido pode deixar salvo o valor anterior. command mantém chave por duas tentativas, mas uma nova ação/reload cria outra identidade.
## What Changes
- R6-UX-01 — Seletores completos para produtos e rotas
- R6-UX-02 — Validação e intenção preservadas em toda mutação
## Context and Problem
Implementação atual a540b40007fe6b8ed523e17afe00e96ff8f8ad50; mudanças incrementais em fronteiras comprovadamente incompletas. Fontes e distinção entre reprodução e hipótese em explore.md.
## Goals
Cumprir os requisitos e cenários do delta, preservando correções já implementadas na R5.
## Non-Goals
Reescrita da plataforma, novo broker sem justificativa, deploy remoto ou aprovação comercial fictícia.
## Users / Actors Impacted
Clientes multi-tenant, provedores e administradores nominais. Responsável técnico: Frontend e Produto.
## Scope
Requisitos listados, testes de falha, integração real, contratos, migrações e evidências. Os vínculos R5 continuam normativos.
## Business Rules
Custódia antes do aceite; efeito externo idempotente ou reconciliado; UUIDv7; final público imutável; isolamento entre tenants; valores financeiros exatos. Sem promessa de disponibilidade absoluta diante de falha de todas as autoridades duráveis.
## Affected Capabilities
r6-05-console-operavel
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
