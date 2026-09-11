# Tasks: Custódia, execução única e resultado final correto

Status: **parcialmente concluído**. Comportamentos com todos os cenários comprovados foram sincronizados; itens com lacunas permanecem abertos. Dependências de change: r2-01-identidade-e-isolamento

## 1. Contratos e preparação

- [ ] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/internal/orbita`; `hub/internal/cometa/executor.go`; `hub/internal/cometa/store.go`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.

- [ ] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/internal/orbita`; `hub/internal/cometa/executor.go`; `hub/internal/cometa/store.go`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
  - Depends on: 1.1.
  - Validation: Contract/schema review e casos inválidos; registrar quais contratos precisam nova versão.
  - Completion criteria: DTOs e versões acordados; nenhuma alteração incompatível implícita no perfil v1.

- [ ] 1.3 Preparar evolução aditiva e fixtures isoladas
  - Objective: Criar migrations adicionais quando aplicável, permissões, interfaces ou organização documental/UI necessária; separar fixtures de dados reais.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 1.2.
  - Validation: Aplicação em ambiente limpo e existente, rollback compatível, validação de unicidade/proveniência; não editar migration aplicada.
  - Completion criteria: Estrutura suporta as regras sem perda de histórico; para UI/qualificação, registrar explicitamente ausência de mudança de esquema quando confirmada.

## 2. Comportamentos

- [x] 2.1 Aceite e intenção recuperáveis em todas as modalidades
  - Objective: Entregar o comportamento R2-EXE-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-01, EXE-15, DAD-03, COM-03, DAD-09, DAD-02.
  - Likely files/components: `hub/internal/orbita/handlers.go`; `hub/internal/orbita/store.go`; `hub/internal/outbox`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [x] 2.2 Idempotência concorrente com UUID recuperável
  - Objective: Entregar o comportamento R2-EXE-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-01, EXE-16, DAD-03.
  - Likely files/components: `hub/internal/orbita/handlers.go`; `hub/internal/orbita/store.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [x] 2.3 Um dono de despacho e tentativa anterior ao efeito
  - Objective: Entregar o comportamento R2-EXE-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-04, EXE-09, EXE-15, DAD-03, COM-06.
  - Likely files/components: `hub/internal/cometa/executor.go`; `hub/internal/cometa/store.go`; `hub/internal/dispatch/contract.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [x] 2.4 Resultado externo válido e durável
  - Objective: Entregar o comportamento R2-EXE-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-04, EXE-14, DAD-04, COM-06.
  - Likely files/components: `hub/internal/cometa/executor.go`; `hub/internal/cometa/store.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

- [x] 2.5 Inbox, outbox e confirmação de mensagens
  - Objective: Entregar o comportamento R2-EXE-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline COM-03, COM-04, OPE-11, FIN-04, EXE-08.
  - Likely files/components: `hub/internal/queue/queue.go`; `hub/internal/outbox`; `hub/internal/orbita/factconsumer.go`; `hub/internal/pulsar/consumer.go`; `hub/internal/libra/consumers.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.

- [x] 2.6 Deadline por confirmação durável e final tardio
  - Objective: Entregar o comportamento R2-EXE-06 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-11, EXE-12, FIN-10, OPE-10.
  - Likely files/components: `hub/internal/orbita/store.go`; `hub/internal/orbita/deadline_timer.go`; `hub/internal/orbita/finalize.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-06; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.6; sem flags que apresentem mock como fluxo real.

- [x] 2.7 SYNC direto e AUTO com espera limitada
  - Objective: Entregar o comportamento R2-EXE-07 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-02, EXE-14, COM-01, COM-06, CAT-11.
  - Likely files/components: `hub/internal/orbita/handlers.go`; `hub/internal/orbita/dispatcher.go`; `hub/internal/platform/httpserver/httpserver.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-07; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.7; sem flags que apresentem mock como fluxo real.

- [ ] 2.8 Representação final única por contrato de cliente
  - Objective: Entregar o comportamento R2-EXE-08 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline COM-05, COM-04, DAD-04, CAT-09, EXE-07.
  - Likely files/components: `hub/internal/orbita/finalize.go`; `hub/internal/orbita/handlers.go`; `hub/internal/pulsar/worker.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-08; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.8; sem flags que apresentem mock como fluxo real.

- [x] 2.9 Reconciliação de obrigações sem reexecutar efeitos
  - Objective: Entregar o comportamento R2-EXE-09 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-03, EXE-09, EXE-15, OPE-11, DAD-09.
  - Likely files/components: `hub/internal/orbita`; `hub/internal/cometa`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-EXE-09; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.9; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [x] 3.1 Qualificar R2-EXE-01 com oráculos independentes
  - Objective: Executar R2-EXE-01-S01, R2-EXE-01-S02, R2-EXE-01-S03. Caso indispensável: Commit incerto; esperado: o Hub resolve pela autoridade de idempotência, sem afirmar não execução nem criar outro efeito.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/orbita/handlers.go`; `hub/internal/orbita/store.go`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.2 Qualificar R2-EXE-02 com oráculos independentes
  - Objective: Executar R2-EXE-02-S01, R2-EXE-02-S02, R2-EXE-02-S03. Caso indispensável: Erro após aceite; esperado: o erro expõe UUID e consulta quando o aceite é conhecido; retry usa a mesma chave.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/orbita/handlers.go`; `hub/internal/orbita/store.go`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.3 Qualificar R2-EXE-03 com oráculos independentes
  - Objective: Executar R2-EXE-03-S01, R2-EXE-03-S02, R2-EXE-03-S03. Caso indispensável: Fencing antigo; esperado: sua ação é rejeitada; lease expirada não prova ausência de efeito já enviado.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa/executor.go`; `hub/internal/cometa/store.go`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.4 Qualificar R2-EXE-04 com oráculos independentes
  - Objective: Executar R2-EXE-04-S01, R2-EXE-04-S02, R2-EXE-04-S03. Caso indispensável: Commit falha; esperado: Cometa não declara fato durável; mantém obrigação recuperável e sem novo submit automático.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa/executor.go`; `hub/internal/cometa/store.go`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.5 Qualificar R2-EXE-05 com oráculos independentes
  - Objective: Executar R2-EXE-05-S01, R2-EXE-05-S02, R2-EXE-05-S03. Caso indispensável: Schema desconhecido; esperado: original sanitizado e metadados ficam em quarentena; alerta e replay autorizado preservam a identidade de origem.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/queue/queue.go`; `hub/internal/outbox`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.6 Qualificar R2-EXE-06 com oráculos independentes
  - Objective: Executar R2-EXE-06-S01, R2-EXE-06-S02, R2-EXE-06-S03, R2-EXE-06-S04. Caso indispensável: Resultado tardio custoso; esperado: GET e webhook conservam o erro final; Libra considera apenas o custo elegível e a contestação.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/orbita/store.go`; `hub/internal/orbita/deadline_timer.go`.
  - Depends on: 2.6; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 4 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.7 Qualificar R2-EXE-07 com oráculos independentes
  - Objective: Executar R2-EXE-07-S01, R2-EXE-07-S02, R2-EXE-07-S03, R2-EXE-07-S04. Caso indispensável: Cliente desconecta; esperado: custódia/reconciliação continua com contexto interno limitado; retry e GET recuperam a mesma identidade.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/orbita/handlers.go`; `hub/internal/orbita/dispatcher.go`.
  - Depends on: 2.7; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 4 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.8 Qualificar R2-EXE-08 com oráculos independentes
  - Objective: Executar R2-EXE-08-S01, R2-EXE-08-S02, R2-EXE-08-S03. Caso indispensável: Resultado pendente ou expirado; esperado: recebe 200 com estado contratado local; consulta não faz polling no provedor nem mascara atraso como sucesso.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/orbita/finalize.go`; `hub/internal/orbita/handlers.go`.
  - Depends on: 2.8; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.9 Qualificar R2-EXE-09 com oráculos independentes
  - Objective: Executar R2-EXE-09-S01, R2-EXE-09-S02, R2-EXE-09-S03. Caso indispensável: Callback antes da associação; esperado: recibo é reconciliado sem perda ou segunda operação.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/orbita`; `hub/internal/cometa`.
  - Depends on: 2.9; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Adicionar estruturas por migration nova, sem editar 0001 aplicado. Backfill com proveniência dos resultados legados; não reconstruir resposta do provedor a partir da entrada. Pedidos sem prova viram LEGACY_UNVERIFIED e seguem tratamento assistido.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/internal/orbita`; `hub/internal/cometa/executor.go`; `hub/internal/cometa/store.go`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-02-execucao-duravel-e-resultados`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
