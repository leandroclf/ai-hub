# Proposal: Evidência consolidada da implementação
## Change ID
r6-06-evidencia-consolidada
## Status
Draft — próxima rodada de implementação, solicitada pelo responsável do projeto.
## Why
O relatório R5 é explicitamente parcial. EVIDENCE_INDEX diz que os logs brutos ficaram no scroll, não foram persistidos. O commit menciona 246 testes, enquanto FINAL_REPORT/EVIDENCE_INDEX mencionam 245. SCENARIO_RESULTS contém apenas subconjunto e IDs de gate ad hoc. As matrizes originais da auditoria não foram regeneradas. Isto limita verificabilidade, não prova que testes históricos falharam.
## What Changes
- R6-QUA-01 — Evidência completa e vinculada ao código revisado
## Context and Problem
Implementação atual a540b40007fe6b8ed523e17afe00e96ff8f8ad50; mudanças incrementais em fronteiras comprovadamente incompletas. Fontes e distinção entre reprodução e hipótese em explore.md.
## Goals
Cumprir os requisitos e cenários do delta, preservando correções já implementadas na R5.
## Non-Goals
Reescrita da plataforma, novo broker sem justificativa, deploy remoto ou aprovação comercial fictícia.
## Users / Actors Impacted
Clientes multi-tenant, provedores e administradores nominais. Responsável técnico: Engenharia e Qualidade.
## Scope
Requisitos listados, testes de falha, integração real, contratos, migrações e evidências. Os vínculos R5 continuam normativos.
## Business Rules
Custódia antes do aceite; efeito externo idempotente ou reconciliado; UUIDv7; final público imutável; isolamento entre tenants; valores financeiros exatos. Sem promessa de disponibilidade absoluta diante de falha de todas as autoridades duráveis.
## Affected Capabilities
r6-06-evidencia-consolidada
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
