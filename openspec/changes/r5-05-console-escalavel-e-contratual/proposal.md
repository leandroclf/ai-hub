# Proposal: Console escalável e contratual
## Change ID
r5-05-console-escalavel-e-contratual
## Status
Draft — planejamento solicitado; código de produção não alterado nesta revisão.
## Why
Delivery ID, SLA/reconcile, DTO financeiro, destinos, OIDC e multimodalidade foram melhorados e têm smokes. Persistem lookup que carrega até 1000 e falha acima disso, validação runtime apenas isRecord e nova chave a cada nova chamada da ação após falha/reload. Capacidade é consulta somente leitura; onboarding de capacidade/qualificação segue seed direto no laboratório. UI TTL ainda descreve aceite (rastreado em R5-EXE-01).
## Context
Brownfield f87ce33034ae29c9431b1910dcc6a633b545e330. R4 remota contém quatro changes; avanços preservados.
## Problem
Falhas nas fronteiras reais impedem cumprir os requisitos vinculados.
## Goals
- Console preserva intenção e escala com catálogo
## Non-Goals
Reescrever stack, fabricar aprovação comercial ou executar produção nesta revisão.
## Users / Actors Impacted
Clientes, provedores, operadores nominais e equipes de suporte. Responsável: Frontend e Produto.
## Scope
### In scope
Comportamentos abaixo e fatias herdadas vinculadas, com testes de falha e integração.
### Out of scope
Provisionamento pago/deploy remoto e alteração de contratos normativos sem aprovação.
## Product Requirements Summary
- R5-UX-01: O console SHALL permitir selecionar referências por busca/paginação de servidor sem teto funcional de 1000, validar contratos de resposta por jornada e preservar identidade de intenção após resposta incerta. Configuração operacional suportada deve ter fluxo autenticado, versionado e auditável; o usuário deve distinguir pendente, confirmado e recusado.
## Business Rules
UUIDv7 após aceite, custódia antes de ACK, isolamento, resultado final único, snapshot contratual e valores exatos.
## Affected Capabilities
r5-05-console-escalavel-e-contratual
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
