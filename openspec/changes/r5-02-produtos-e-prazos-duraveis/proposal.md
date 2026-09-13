# Proposal: Produtos, prazos e compensação duráveis
## Change ID
r5-02-produtos-e-prazos-duraveis
## Status
Draft — planejamento solicitado; código de produção não alterado nesta revisão.
## Why
A primeira falha/horizonte persistem, mas ClaimIntent e ClaimDirectIntent ainda não excluem retry_until vencido; CompleteIntent avalia depois de I/O e DELIVERED prevalece. O polling já foi corrigido para priorizar StepDeadline em vez do TTL inicial e essa correção deve ser preservada. UI ainda explica TTL desde aceite, divergindo da regra de primeira falha. Fencing de submissão foi implementado e deve ser preservado; ele não elimina a diferença entre esses relógios.
Produtos agora têm plano/etapas e execução HTTP comprovada em laboratório. Porém o claim conta RUNNING/WAITING_PROVIDER sem cercar a linha do plano e atualiza etapa em outra transação: dois publishers podem observar vaga simultaneamente. Se crash ocorre após RUNNING antes de publicar, re-claim depende da mesma contagem e pode ficar bloqueado no próprio max_parallel. O helper antigo ExecuteDAG segue com falha de cancelamento, mas não é o runtime persistido; não confundir as implementações.
BuildProductPlan copia o snapshot do produto, troca Target e preserva SelectedRoute/Account/Binding e EconomicSnapshot de base. Isso executa duas etapas no mesmo provedor sintético, mas não demonstra composição de serviços com provedores, credenciais e contratos de compra distintos. Hash do filho é esvaziado. O plano guarda consolidation, porém consolidação final usa formato fixo de steps.
Compensações são persistidas, avanço sobre a R3. Porém são inseridas READY com depends_on vazio e herdam deadlines do comando original; o publisher exige protocolo não terminal e client_deadline futura. Assim, ordem reversa causal não está representada e compensação de obrigação externa pode deixar de executar depois de encerrar o atendimento.
Bootstrap assíncrono preserva disponibilidade HTTP durante falha de broker, mas cada processo confirma apenas sua parte. Órbita cria tópico de fatos finais e inicia relay sem comprovar assinaturas de Pulsar/Libra; Cometa publica fatos externos sem barreira comum de todas as assinaturas obrigatórias. Em ambiente limpo com startup fora de ordem, publicação SNS pode ser confirmada antes de assinaturas necessárias. Risco estático a ensaiar; tópicos já provisionados escondem essa janela.
## Context
Brownfield f87ce33034ae29c9431b1910dcc6a633b545e330. R4 remota contém quatro changes; avanços preservados.
## Problem
Falhas nas fronteiras reais impedem cumprir os requisitos vinculados.
## Goals
- Horizonte de retry e SLA verificados antes do despacho
- Limite de execução de produto atômico e retomável
- Contrato, provedor e incidência próprios por etapa
- Compensação tem ordem causal e prazo próprio
- Topologia de mensagens pronta antes de publicar obrigações
## Non-Goals
Reescrever stack, fabricar aprovação comercial ou executar produção nesta revisão.
## Users / Actors Impacted
Clientes, provedores, operadores nominais e equipes de suporte. Responsável: Core e Integrações.
## Scope
### In scope
Comportamentos abaixo e fatias herdadas vinculadas, com testes de falha e integração.
### Out of scope
Provisionamento pago/deploy remoto e alteração de contratos normativos sem aprovação.
## Product Requirements Summary
- R5-EXE-01: O Hub SHALL impedir novos despachos com possibilidade de efeito após o horizonte aplicável e conservar separadamente prazo do cliente, provedor, tentativa e retry de indisponibilidade desde primeira falha. Pendência assíncrona legítima não inicia TTL de falha. Takeover não renova prazos; observação tardia conserva evidência sem alterar final fechado.
- R5-EXE-02: O Hub SHALL reservar vaga de execução e posse da etapa atomicamente por produto, respeitando max_parallel entre réplicas. Uma etapa com posse expirada deve ser recuperável sem disputar uma segunda vaga nem repetir efeito confirmado. Cancelamento impede novo efeito e estado terminal não é ressuscitado.
- R5-EXE-03: O Hub SHALL congelar por etapa serviço/versão, rota homologada, conta, binding, contrato de compra, perfil técnico, prazo e identidade econômica coerentes. A venda do produto e os custos das etapas devem manter escopos próprios sem duplicação. Consolidação deve seguir política publicada e preservar proveniência do plano.
- R5-EXE-04: O Hub SHALL conservar e executar compensações em ordem causal reversa com prazo, retry e autorização próprios, independentemente do encerramento da resposta ao cliente. Falha ou incerteza de compensação deve permanecer reconciliável sem ser convertida em sucesso nem reabrir protocolo.
- R5-EXE-05: O Hub SHALL confirmar a topologia e políticas de todas as assinaturas obrigatórias antes de liberar publicação de fatos duráveis. Indisponibilidade da topologia deve manter obrigações no outbox sem bloquear rotas independentes. Alteração ou recriação de recurso exige revalidação antes de descartar custódia local.
## Business Rules
UUIDv7 após aceite, custódia antes de ACK, isolamento, resultado final único, snapshot contratual e valores exatos.
## Affected Capabilities
r5-02-produtos-e-prazos-duraveis
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
