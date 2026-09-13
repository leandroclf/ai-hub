# Proposal: Isolamento e custódia autenticada
## Change ID
r6-01-isolamento-e-custodia
## Status
Draft — próxima rodada de implementação, solicitada pelo responsável do projeto.
## Why
A migração 0049 remove o bypass nominal nas tabelas core contempladas. WithTenantTx continua sem chamada produtiva, os DSNs Compose usam hub e o contexto opcional continua fixo por workload. A prova rls-runtime-proof.sh continua executando ROLLBACK das fixtures antes das assertions negativas; a contagem dentro da transação só é impressa. Portanto, a execução histórica do script não prova isolamento de dados existentes nem adoção pelo runtime.
A assinatura é agora verificada antes da custódia órfã e a identidade não depende do timestamp: preservar a correção. A reconciliação ainda revalida com CALLBACK_INGRESS_KEY atual, não com uma versão de chave congelada. O lock advisory de storeOrphan continua global, e três falhas de apply podem converter recibo em REJECTED. Órfãos sem operação não são selecionados pelo JOIN de reconciliação.
## What Changes
- R6-SEG-01 — RLS no caminho real e prova com dados existentes
- R6-SEG-02 — Callback aceito sobrevive à rotação e à indisponibilidade
## Context and Problem
Implementação atual a540b40007fe6b8ed523e17afe00e96ff8f8ad50; mudanças incrementais em fronteiras comprovadamente incompletas. Fontes e distinção entre reprodução e hipótese em explore.md.
## Goals
Cumprir os requisitos e cenários do delta, preservando correções já implementadas na R5.
## Non-Goals
Reescrita da plataforma, novo broker sem justificativa, deploy remoto ou aprovação comercial fictícia.
## Users / Actors Impacted
Clientes multi-tenant, provedores e administradores nominais. Responsável técnico: Segurança e Core.
## Scope
Requisitos listados, testes de falha, integração real, contratos, migrações e evidências. Os vínculos R5 continuam normativos.
## Business Rules
Custódia antes do aceite; efeito externo idempotente ou reconciliado; UUIDv7; final público imutável; isolamento entre tenants; valores financeiros exatos. Sem promessa de disponibilidade absoluta diante de falha de todas as autoridades duráveis.
## Affected Capabilities
r6-01-isolamento-e-custodia
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
