# Proposal: Fronteiras de autorização e custódia
## Change ID
r5-01-fronteiras-de-autorizacao
## Status
Draft — planejamento solicitado; código de produção não alterado nesta revisão.
## Why
A rota pública e HMAC por conta foram implementados. Porém AuthenticateAccountCallback consulta operations antes de Verify. Para operação inexistente, o handler chama StoreOrphanAccountCallback, que só valida campos não vazios e grava via storeOrphan sem Verify. Assim, uma assinatura arbitrária com timestamp recente pode atingir a custódia órfã e 202, se o banco estiver disponível. A quota é global; variação da assinatura muda token_hash e identidade de dedupe. A reconciliação só seleciona órfãos com operação existente e prune não remove RECEIVED. Evidência estática de caminho; não foi executada exploração externa.
Offer agora tem cache de fallback, mas chama Atlas antes dele em toda requisição e usa cache após qualquer erro. Probe em httptest reproduziu retorno de oferta antiga após HTTP 403 offer_not_eligible. Suspensão/revogação explícita pode ser ocultada até vencer o cache; o caminho quente ainda paga a consulta remota.
WithTenantTx permanece sem chamada nos stores; RuntimeDSN é opt-in de tenant fixo. Script RLS continua revertendo fixtures antes da assertion negativa e apenas imprime a contagem dentro da transação. Zero após rollback não prova isolamento. Migração de compatibilidade permite hub e grants abrangem tabelas auxiliares sem cobertura completa. Endpoint capacity-domains consulta todos os domínios para integrations:read sem escopo por recurso explícito.
## Context
Brownfield f87ce33034ae29c9431b1910dcc6a633b545e330. R4 remota contém quatro changes; avanços preservados.
## Problem
Falhas nas fronteiras reais impedem cumprir os requisitos vinculados.
## Goals
- Autenticação antes de custódia de callback órfão
- Fallback de oferta respeita negação e revogação
- Adoção real de RLS e autorização de dados auxiliares
## Non-Goals
Reescrever stack, fabricar aprovação comercial ou executar produção nesta revisão.
## Users / Actors Impacted
Clientes, provedores, operadores nominais e equipes de suporte. Responsável: Segurança e Core.
## Scope
### In scope
Comportamentos abaixo e fatias herdadas vinculadas, com testes de falha e integração.
### Out of scope
Provisionamento pago/deploy remoto e alteração de contratos normativos sem aprovação.
## Product Requirements Summary
- R5-SEG-01: O Hub SHALL autenticar a origem antes de confirmar ou reservar custódia órfã. Cada recibo deve possuir escopo de origem autenticada, identidade estável de evento e disposição recuperável. Tentativas não autenticadas não podem consumir quota durável de obrigações válidas. Quotas, retenção e claims devem isolar contas/células; eventos irrelacionáveis devem chegar a disposição auditável sem bloqueio global.
- R5-SEG-02: O Hub SHALL distinguir indisponibilidade transitória de negação autoritativa ao resolver ofertas. Negação, revogação, conflito ou resposta inválida não podem ser convertidos em autorização por cache. Projeções válidas devem atender o caminho quente dentro de janela de autorização explicitamente publicada, com invalidação e limites de armazenamento mensuráveis.
- R5-SEG-03: O Hub SHALL aplicar identidade autenticada por transação/work item com roles runtime sem propriedade nem bypass, incluindo tabelas auxiliares e caminhos administrativos. Escopo global exige autorização nominal específica e auditoria. Provas de isolamento devem demonstrar dados próprios existentes, dados alheios existentes e invisíveis, escrita cruzada recusada e ausência de vazamento ao reutilizar conexões.
## Business Rules
UUIDv7 após aceite, custódia antes de ACK, isolamento, resultado final único, snapshot contratual e valores exatos.
## Affected Capabilities
r5-01-fronteiras-de-autorizacao
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
