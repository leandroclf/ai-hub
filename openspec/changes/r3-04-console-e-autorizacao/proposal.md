# Proposal: Console e autorização
## Change ID
r3-04-console-e-autorizacao
## Status
Draft — especificação pronta para implementação; nenhuma tarefa implementada por esta entrega.
## Why
Leitura administrativa no mesmo tenant aceita protocols:read sem papel administrativo específico/MFA e lista todas as aplicações. Entre tenants já existe verificação MFA/papel; não se trata de afirmar bypass global irrestrito.
Seleção usa protocol_id antes de delivery_id. Menu SLA aponta para rota não registrada; botão de reconciliação POST aponta para handler de leitura.
UI envia tenant_id/prepared_by a decoder estrito incompatível e datas YYYY-MM-DD para time.Time. Ações não mantêm tenant selecionado na query e criam idempotency key nova em cada execução.
Editor de modos reduz array a seleção única; faltam campos condicionais de autenticação; lookups limitados à primeira centena. datetime-local corta UTC e o reinterpreta no fuso local ao salvar.
## Context
Evolução do snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee; preservar mecanismos corretos v4/R2.
## Problem
As lacunas abaixo interrompem jornadas reais e deixam garantias normativas sem demonstração.
## Goals
- Fronteira administrativa e aplicação
- Ações operacionais com API efetiva
- Comandos financeiros compatíveis
- Formulários completos e tempo local
## Non-Goals
Reescrever stack, renomear componentes ou relaxar SLA/custódia para obter teste verde.
## Users / Actors Impacted
Clientes, provedores, operadores, desenvolvedores nominais. Responsável funcional: Frontend e Segurança.
## Scope
### In scope
Requisitos e cenários deste change mais a requalificação herdada vinculada.
### Out of scope
Migração tecnológica não justificada e contratação comercial não autorizada.
## Product Requirements Summary
- R3-ADM-01: O Hub SHALL distinguir consumidor de operador e aplicar escopo por tenant/aplicação em listagens, detalhe e payload. Leitura entre tenants exige identidade nominal autorizada, MFA, auditoria e mascaramento; desenvolvedor não usa conta compartilhada nem herda escrita financeira.
- R3-ADM-02: O console SHALL usar contratos validados por operação, delivery_id para entregas e endpoints efetivos para reconciliação e SLA. Ação não implementada não pode simular disponibilidade. Redelivery repete bytes da entrega e não executa novamente provedor.
- R3-ADM-03: O console SHALL emitir contrato financeiro válido, propagar tenant autorizado, converter períodos com fuso explícito e preservar chave da intenção em retries. Autor deriva da identidade; aprovação, disputa e exportação devem completar jornadas com segregação.
- R3-ADM-04: O console SHALL preservar multimodalidade, oferecer campos condicionais de autenticação, lookup pesquisável/paginado e conversão correta de fuso. Respostas devem ser validadas em runtime e jornadas acessíveis por teclado com erros vinculados aos campos.
## Business Rules
UUIDv7 recuperável após aceite, isolamento de tenant/aplicação, snapshot imutável, custódia antes de ACK, resultado terminal único e finanças exatas permanecem obrigatórios.
## Affected Capabilities
04-console-e-autorizacao
## Expected Impact
### Code
hub/admin-ui/src, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/FinancePage.tsx, hub/admin-ui/src/pages/OperationsPage.tsx, hub/internal/libra/handlers.go, hub/internal/orbita/admin.go, hub/internal/platform/auth/auth.go, hub/internal/pulsar/handlers.go
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
