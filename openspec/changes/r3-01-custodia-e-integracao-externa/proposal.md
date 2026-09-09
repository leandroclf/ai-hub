# Proposal: Custódia e integração externa
## Change ID
r3-01-custodia-e-integracao-externa
## Status
Draft — especificação pronta para implementação; nenhuma tarefa implementada por esta entrega.
## Why
O executor recusa AdapterID diferente de synthetic-provider. Configurar endpoint não cria integração executável. Config não recebe APIKeyHeader, embora Apply o exija.
Callback chama ApplyExternalObservation sem retorno de erro e responde 200; operação desconhecida, erro de consulta e observação após terminal não têm confirmação de recibo nesse caminho. A rota está sob JWT interno do Hub, não sob autenticação específica da conta provedora.
Após SUBMITTING, replay pode devolver UNKNOWN durável sem recuperador geral desse estado. Envio não transmite chave idempotente homologada. Consulta inicial de replay antecede validação completa de hash/aplicação.
RetryDeadline deriva do aceite e limita polling; provider_mode mantém polling/callback exclusivos. ADR R2 registra contraexemplo de commit posterior ao deadline e requisito não qualificado.
Bootstraps independentes não estabelecem barreira de todas as assinaturas obrigatórias antes do relay. Quarentena financeira conserva hash/motivo sem payload para replay e então confirma a mensagem.
## Context
Evolução do snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee; preservar mecanismos corretos v4/R2.
## Problem
As lacunas abaixo interrompem jornadas reais e deixam garantias normativas sem demonstração.
## Goals
- Adapters executáveis e autenticação completa
- Callback com confirmação de custódia
- Recuperação de efeito incerto e fencing
- Relógios de retry, polling e prazo final
- Topologia de mensagens e quarentena recuperável
## Non-Goals
Reescrever stack, renomear componentes ou relaxar SLA/custódia para obter teste verde.
## Users / Actors Impacted
Clientes, provedores, operadores, desenvolvedores nominais. Responsável funcional: Core e Integrações.
## Scope
### In scope
Requisitos e cenários deste change mais a requalificação herdada vinculada.
### Out of scope
Migração tecnológica não justificada e contratação comercial não autorizada.
## Product Requirements Summary
- R3-EXE-01: O Hub SHALL executar adapters versionados homologados, transmitir configuração completa de autenticação e distinguir inventário importado de capacidade executável. Publicação deve rejeitar capacidades não suportadas; resultado sintético nunca substitui resposta real.
- R3-EXE-02: O Hub SHALL autenticar callback pela política da conta, limitar método/tamanho/replay e emitir 2xx somente após custódia durável. Recibos órfãos, duplicados e tardios autenticados devem ter disposição recuperável sem substituir final terminal.
- R3-EXE-03: O Hub SHALL manter obrigação recuperável para SUBMITTING/UNKNOWN, validar identidade completa/hash em todo replay e verificar lease/epoch/prazo antes de I/O. Reenvio de SUBMIT exige idempotência homologada ou prova positiva de ausência de efeito; UNKNOWN exige consulta ou reconciliação.
- R3-EXE-04: O Hub SHALL separar SLA do cliente, SLA do provedor, TTL desde primeira falha transitória, timeout por tentativa e horizonte de reconciliação. Polling/callback devem coexistir; espera saudável não consome TTL de indisponibilidade. Não substituir confirmação durável no prazo por checagem SQL anterior ao commit sem mudança normativa autorizada.
- R3-EXE-05: O Hub SHALL verificar topologia durável obrigatória antes de liberar publicação, conservar cada obrigação até handoff comprovado e guardar bytes originais limitados ou referência imutável protegida em quarentena. Hash sozinho não é custódia recuperável; reconciliação cobre fan-out e retenção.
## Business Rules
UUIDv7 recuperável após aceite, isolamento de tenant/aplicação, snapshot imutável, custódia antes de ACK, resultado terminal único e finanças exatas permanecem obrigatórios.
## Affected Capabilities
01-custodia-e-integracao-externa
## Expected Impact
### Code
docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md, hub/cmd/cometa/main.go, hub/internal/atlasclient/client.go, hub/internal/cometa/custody.go, hub/internal/cometa/executor.go, hub/internal/cometa/handlers.go, hub/internal/cometa/poller.go, hub/internal/cometa/polling_custody.go, hub/internal/libra/consumers.go, hub/internal/libra/store.go, hub/internal/orbita/finalize.go, hub/internal/orbita/handlers.go, hub/internal/orbita/intents.go, hub/internal/queue/bootstrap.go, hub/internal/queue/queue.go
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
