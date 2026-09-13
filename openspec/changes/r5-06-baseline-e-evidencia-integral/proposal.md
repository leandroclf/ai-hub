# Proposal: Baseline e evidência integral
## Change ID
r5-06-baseline-e-evidencia-integral
## Status
Draft — planejamento solicitado; código de produção não alterado nesta revisão.
## Why
Repositório contém R4 original (12 requisitos/36 cenários), não o quinto change da edição regenerada (quatro requisitos adicionais). Baseline real é 201/732; matriz reconhece 231 linhas associadas e 501 sem qualificação. Gerador associa resultado por scenario_id, sem exigir digest do código/spec da execução na linha de origem; vínculo textual não prova atualidade nem PASS. R4 adicional precisa ponte explícita, sem copiar 744 como contagem remota.
## Context
Brownfield f87ce33034ae29c9431b1910dcc6a633b545e330. R4 remota contém quatro changes; avanços preservados.
## Problem
Falhas nas fronteiras reais impedem cumprir os requisitos vinculados.
## Goals
- Baseline reconciliada e resultados com proveniência
## Non-Goals
Reescrever stack, fabricar aprovação comercial ou executar produção nesta revisão.
## Users / Actors Impacted
Clientes, provedores, operadores nominais e equipes de suporte. Responsável: Engenharia e Qualidade.
## Scope
### In scope
Comportamentos abaixo e fatias herdadas vinculadas, com testes de falha e integração.
### Out of scope
Provisionamento pago/deploy remoto e alteração de contratos normativos sem aprovação.
## Product Requirements Summary
- R5-QUA-01: A engenharia SHALL inventariar specs reais e reconciliar revisões não incorporadas com rastreio explícito, sem perder requisito nem duplicar identidade. Resultado deve carregar proveniência de cenário/conteúdo/artefato e oráculo, distinguindo PASS, FAIL, NOT_RUN e bloqueio externo. Associação a arquivo não é conformidade; mudança material invalida a prova afetada.
## Business Rules
UUIDv7 após aceite, custódia antes de ACK, isolamento, resultado final único, snapshot contratual e valores exatos.
## Affected Capabilities
r5-06-baseline-e-evidencia-integral
## Expected Impact
### Code
Fontes em explore.md e tarefas discriminadas por requisito.
### Data
Migrações aditivas; obrigações antigas preservadas, backfill validado e sem atualização cega do ledger.
### APIs / Contracts
Semântica explícita, versionada e compatível; erros não viram sucesso.
### Integrations
Qualificação externa por conta/adapter; fixture não equivale a contrato comercial.
### Operations
Scripts reproduzíveis e evidências do conteúdo atual; um Compose conforme AGENTS.
### Security / Privacy
Identidade do token/work item, menor privilégio, sem segredo em artefatos.
## Risks and Mitigations
Ver risk-matrix.md, design.md e cenários negativos.
## Success Criteria
Todos os cenários deste change e herdados afetados comprovados com código, oráculo, comando e digest.
## Assumptions
Defaults sintéticos autorizados no registro R3 podem apoiar laboratório; não aprovam produção.
## Open Questions
D-01…D-07/P-01…P-11/T-R2-01: conferir registro real. Somente gates dependentes ficam bloqueados.
