# Proposal: Qualificação integral
## Change ID
r3-07-qualificacao-integral
## Status
Draft — especificação pronta para implementação; nenhuma tarefa implementada por esta entrega.
## Why
Go/race passam nesta revisão com 18 testes SKIP. Relatório R2 admite integração incompleta e fixture OIDC bloqueada. Build da UI não detecta incompatibilidades de comandos.
## Context
Evolução do snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee; preservar mecanismos corretos v4/R2.
## Problem
As lacunas abaixo interrompem jornadas reais e deixam garantias normativas sem demonstração.
## Goals
- Qualificação integral do SHA sem skips ocultos
## Non-Goals
Reescrever stack, renomear componentes ou relaxar SLA/custódia para obter teste verde.
## Users / Actors Impacted
Clientes, provedores, operadores, desenvolvedores nominais. Responsável funcional: Engenharia e Qualidade.
## Scope
### In scope
Requisitos e cenários deste change mais a requalificação herdada vinculada.
### Out of scope
Migração tecnológica não justificada e contratação comercial não autorizada.
## Product Requirements Summary
- R3-QUA-01: A engenharia SHALL qualificar o SHA entregue com integração obrigatória sem skips, fixture OIDC reproduzível, navegador real e oráculos externos, cobrindo toda baseline v4/R2/R3. Evidência histórica não qualifica outro SHA; tarefas só fecham com provas correspondentes e bloqueios materiais permanecem explícitos.
## Business Rules
UUIDv7 recuperável após aceite, isolamento de tenant/aplicação, snapshot imutável, custódia antes de ACK, resultado terminal único e finanças exatas permanecem obrigatórios.
## Affected Capabilities
07-qualificacao-integral
## Expected Impact
### Code
docs/reviews/2026-09-07-r2/implementation/FINAL_REPORT.md, hub/internal/atlas/catalog_test.go, hub/internal/cometa/custody_test.go
### Data
Migrações aditivas e identidades/versionamento explícitos; sem perda de registros históricos.
### APIs / Contracts
Compatibilidade por versão, erros tipados, autorização em backend e sem sucesso fictício.
### Integrations
Caminho real com oráculos e credenciais de fixture, sem credenciais comerciais em artefatos.
### Operations
Métricas, recuperação, runbook e rollback devem acompanhar o comportamento.
### Security / Privacy
Menor privilégio; sem payload/segredo bruto em logs/evidência.
## Risks and Mitigations
Consultar risk-matrix.md; cada risco tem requisito, cenário e tarefa.
## Success Criteria
Todos os cenários deste change e regressão vinculada passam no SHA entregue; zero skip no gate obrigatório.
## Assumptions
Stack existente é mantida. Valores comerciais ausentes são fixtures explicitamente sintéticas.
## Open Questions
Decisões D-01…D-07/P-01…P-11 e T-R2-01 são preservadas; ver documento R3 05. Semântica não pode mudar por inferência.
