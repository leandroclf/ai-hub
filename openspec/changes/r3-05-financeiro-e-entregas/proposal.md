# Proposal: Financeiro e entregas
## Change ID
r3-05-financeiro-e-entregas
## Status
Draft — especificação pronta para implementação; nenhuma tarefa implementada por esta entrega.
## Why
Reserva exata existe, mas captura não ajusta seu valor ao valor efetivo. DENY por franquia ocorre na incidência após execução externa.
Fatos não transportam todas as unidades/identidades ATTEMPT; STATUS/FETCH carecem de incidência integrada. SetWatermark tem chamada em teste sem fluxo produtor de completude; ciclo de disputa/fechamento não está completo.
Pulsar escolhe versões ACTIVE atuais por tenant no consumo, em vez de destinos/aplicação congelados na admissão.
## Context
Evolução do snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee; preservar mecanismos corretos v4/R2.
## Problem
As lacunas abaixo interrompem jornadas reais e deixam garantias normativas sem demonstração.
## Goals
- Reserva estrita e franquia antes do efeito
- Incidência completa e fechamento operacional
- Destino de webhook congelado por aplicação
## Non-Goals
Reescrever stack, renomear componentes ou relaxar SLA/custódia para obter teste verde.
## Users / Actors Impacted
Clientes, provedores, operadores, desenvolvedores nominais. Responsável funcional: Financeiro e Core.
## Scope
### In scope
Requisitos e cenários deste change mais a requalificação herdada vinculada.
### Out of scope
Migração tecnológica não justificada e contratação comercial não autorizada.
## Product Requirements Summary
- R3-FIN-01: O Hub SHALL autorizar saldo/franquia antes do efeito, reservar exposição máxima ou obter autorização incremental prévia e conciliar captura/liberação com valor efetivo exato. UNKNOWN mantém retenção até evidência positiva; expiração do cliente não prova ausência de custo.
- R3-FIN-02: O Hub SHALL medir unidades de compra/venda do snapshot, incluindo SUBMIT/STATUS/FETCH contratados, com attempt_id/evidence_id e deduplicação por unidade. Watermarks derivam de completude demonstrável; disputa, conciliação e exportação devem funcionar sem editar banco manualmente.
- R3-FIN-03: O Hub SHALL congelar destinos autorizados e contrato de entrega, conservar representação final imutável e distinguir reentrega de execução. Ausência de destino tem disposição explícita. Tentativas, leases e comando manual respeitam budgets auditáveis.
## Business Rules
UUIDv7 recuperável após aceite, isolamento de tenant/aplicação, snapshot imutável, custódia antes de ACK, resultado terminal único e finanças exatas permanecem obrigatórios.
## Affected Capabilities
05-financeiro-e-entregas
## Expected Impact
### Code
hub/internal/cometa/custody.go, hub/internal/cometa/polling_custody.go, hub/internal/contracts/economics/publication.go, hub/internal/libra/handlers.go, hub/internal/libra/store.go, hub/internal/orbita/finalize.go, hub/internal/orbita/handlers.go, hub/internal/pulsar/custody.go, hub/internal/pulsar/handlers.go
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
