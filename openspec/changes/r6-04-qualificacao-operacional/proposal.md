# Proposal: Qualificação operacional e promoção
## Change ID
r6-04-qualificacao-operacional
## Status
Draft — próxima rodada de implementação, solicitada pelo responsável do projeto.
## Why
A ausência de manifesto agora bloqueia prd. Probe novo passou manifesto de um cenário, SHA de quarenta zeros, status PASS_WITHOUT_EXECUTION, command not executed e oráculos inventados iguais: o gate respondeu ALLOW. O validador exige forma e presença do ID no arquivo, não a identidade do candidato, integridade, cobertura obrigatória ou proveniência confiável.
O script de restore não mudou nesta rodada: compara lista parcial de tabelas, faz s3 sync/list-objects-v2 e consulta o oráculo somente antes de declarar PASS sem replay. Planos, etapas, novos settlements e versões históricas de objetos não têm cobertura explícita. O relatório da implementação reconhece o item não iniciado.
Kind independente já existe e EnsureTopology foi ligado aos quatro processos antes dos consumidores/relays: preservar. Não houve evolução de deploy além do gate nesta rodada. O gerador de dependências kind declara volumes efêmeros; isso é adequado ao laboratório, mas não prova recuperação ou ambientes remotos duráveis. Bootstrap único também não qualifica remoção posterior de assinatura SNS.
O executor passou a recusar controller nil quando domínio existe. Domínio vazio ainda retorna capacidade desabilitada. Há resolução por evidência, porém settle com lease vencido pode falhar e falhas de resolução são logadas; contagens sob lock percorrem histórico do domínio. Não afirmar ausência do controlador adaptativo, que já está implementado.
## What Changes
- R6-OPE-01 — Gate de promoção vinculado ao artefato e à cobertura
- R6-OPE-02 — Restore populado com versões e retomada reconciliada
- R6-OPE-03 — Ambientes e elasticidade com orçamento de dependências
- R6-OPE-04 — Capacidade obrigatória e feedback recuperável
## Context and Problem
Implementação atual a540b40007fe6b8ed523e17afe00e96ff8f8ad50; mudanças incrementais em fronteiras comprovadamente incompletas. Fontes e distinção entre reprodução e hipótese em explore.md.
## Goals
Cumprir os requisitos e cenários do delta, preservando correções já implementadas na R5.
## Non-Goals
Reescrita da plataforma, novo broker sem justificativa, deploy remoto ou aprovação comercial fictícia.
## Users / Actors Impacted
Clientes multi-tenant, provedores e administradores nominais. Responsável técnico: Plataforma e SRE.
## Scope
Requisitos listados, testes de falha, integração real, contratos, migrações e evidências. Os vínculos R5 continuam normativos.
## Business Rules
Custódia antes do aceite; efeito externo idempotente ou reconciliado; UUIDv7; final público imutável; isolamento entre tenants; valores financeiros exatos. Sem promessa de disponibilidade absoluta diante de falha de todas as autoridades duráveis.
## Affected Capabilities
r6-04-qualificacao-operacional
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
