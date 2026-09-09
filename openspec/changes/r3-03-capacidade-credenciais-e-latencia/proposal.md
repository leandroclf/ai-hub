# Proposal: Capacidade, credenciais e latência
## Change ID
r3-03-capacidade-credenciais-e-latencia
## Status
Draft — especificação pronta para implementação; nenhuma tarefa implementada por esta entrega.
## Why
Capacity possui código/testes, mas Execute/requestPoll não adquirem concessão nem realimentam o controlador. Relatório R2 admite a desconexão.
Bearer resolve segredo antes do L1, mantém mutex do TokenCache durante Redis/OAuth e grava bearer token no Redis, contrariando a restrição R2.
NewClient cria Transport por submit/poll/OAuth. MaxConnsPerHost em transports distintos não limita consumo agregado e impede reaproveitamento eficiente.
## Context
Evolução do snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee; preservar mecanismos corretos v4/R2.
## Problem
As lacunas abaixo interrompem jornadas reais e deixam garantias normativas sem demonstração.
## Goals
- Controle adaptativo conectado a todo I/O
- Cache de autenticação isolado e sem tokens no Redis
- Pools HTTP e budgets de concorrência
## Non-Goals
Reescrever stack, renomear componentes ou relaxar SLA/custódia para obter teste verde.
## Users / Actors Impacted
Clientes, provedores, operadores, desenvolvedores nominais. Responsável funcional: Integrações e Plataforma.
## Scope
### In scope
Requisitos e cenários deste change mais a requalificação herdada vinculada.
### Out of scope
Migração tecnológica não justificada e contratação comercial não autorizada.
## Product Requirements Summary
- R3-INT-01: O Hub SHALL adquirir concessão global por domínio antes de SUBMIT/STATUS/FETCH, devolver métricas de resultado e adaptar concorrência com redução por erros, recuperação amortecida e Retry-After. Limites contratuais duros coexistem com adaptação e justiça entre tenants.
- R3-INT-02: O Hub SHALL manter segredos em cofre/cache de memória limitado com expiração/revogação, eliminar tokens do Redis e coordenar renovação por binding sem bloquear contas independentes. L1 válido deve dispensar chamada remota; falha não permite token expirado ou fallback de outro cliente.
- R3-INT-03: O Hub SHALL reutilizar pools limitados por origem/identidade TLS, manter timeout por operação e drenar pools obsoletos. Controlador global limita concorrência; certificados/bindings incompatíveis não compartilham estado de autenticação.
## Business Rules
UUIDv7 recuperável após aceite, isolamento de tenant/aplicação, snapshot imutável, custódia antes de ACK, resultado terminal único e finanças exatas permanecem obrigatórios.
## Affected Capabilities
03-capacidade-credenciais-e-latencia
## Expected Impact
### Code
hub/internal/cometa/capacity.go, hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, hub/internal/platform/egress, hub/internal/providerauth/client.go
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
