# Proposal: Reconciliação financeira recuperável
## Change ID
r6-03-reconciliacao-financeira
## Status
Draft — próxima rodada de implementação, solicitada pelo responsável do projeto.
## Why
Payload e endpoints de replay foram adicionados. ProcessEnvelope retorna nil também quando Quarantine grava com sucesso; o handler então marca REPLAYED. Se o payload ainda inválido recria a mesma identidade, ON CONFLICT DO NOTHING conserva a linha e MarkQuarantineReplayed a retira da lista de pendências sem aplicação. Fato tardio é armazenado como EconomicEvent, enquanto o endpoint espera queue.Envelope.
A captura efetiva agora é chamada por ApplyEvent e registra diferença da reserva. SetWatermark permanece chamado apenas por teste; assim o mecanismo de fechamento carece de produtor durável de completude. Fatos tardios ainda entram em quarentena, não demonstram disputa/ajuste completo. A correção de captura deve ser preservada e testada com múltiplos fatos/ordem/escala decimal.
## What Changes
- R6-FIN-01 — Replay financeiro distingue aplicação de nova quarentena
- R6-FIN-02 — Completude e liquidação verificáveis no fluxo real
## Context and Problem
Implementação atual a540b40007fe6b8ed523e17afe00e96ff8f8ad50; mudanças incrementais em fronteiras comprovadamente incompletas. Fontes e distinção entre reprodução e hipótese em explore.md.
## Goals
Cumprir os requisitos e cenários do delta, preservando correções já implementadas na R5.
## Non-Goals
Reescrita da plataforma, novo broker sem justificativa, deploy remoto ou aprovação comercial fictícia.
## Users / Actors Impacted
Clientes multi-tenant, provedores e administradores nominais. Responsável técnico: Financeiro e Backend.
## Scope
Requisitos listados, testes de falha, integração real, contratos, migrações e evidências. Os vínculos R5 continuam normativos.
## Business Rules
Custódia antes do aceite; efeito externo idempotente ou reconciliado; UUIDv7; final público imutável; isolamento entre tenants; valores financeiros exatos. Sem promessa de disponibilidade absoluta diante de falha de todas as autoridades duráveis.
## Affected Capabilities
r6-03-reconciliacao-financeira
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
