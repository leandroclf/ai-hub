# Tasks: Adaptadores reais, credenciais, polling, callbacks e pressão

Status: **todas abertas**. Este arquivo planeja implementação futura; esta revisão só produziu documentação. Dependências de change: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## 1. Contratos e preparação

- [ ] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/internal/cometa`; `hub/internal/providerauth/client.go`; `hub/internal/pulsar`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.

- [ ] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/internal/cometa`; `hub/internal/providerauth/client.go`; `hub/internal/pulsar`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
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

- [ ] 2.1 Adapter executável homologado por capacidade
  - Objective: Entregar o comportamento R2-INT-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline COM-02, CAT-06, CAT-11.
  - Likely files/components: `hub/internal/cometa/executor.go`; `hub/internal/providersim/providersim.go`; `hub/internal/atlas/store.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [ ] 2.2 Credencial efetiva vinculada ao cliente e à conta
  - Objective: Entregar o comportamento R2-INT-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-05, SEG-05, FIN-11, CAT-11.
  - Likely files/components: `hub/internal/cometa/executor.go`; `hub/internal/providerauth/client.go`; `hub/internal/atlas/store.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [ ] 2.3 Segredos reais e cache dispensável
  - Objective: Entregar o comportamento R2-INT-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline SEG-05, DAD-10, OPE-13, ARQ-03.
  - Likely files/components: `hub/internal/providerauth/client.go`; `hub/cmd/cometa/main.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [ ] 2.4 Polling e callback combináveis e coordenados
  - Objective: Entregar o comportamento R2-INT-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-05, EXE-06, SEG-02.
  - Likely files/components: `hub/internal/cometa/poller.go`; `hub/internal/cometa/handlers.go`; `hub/internal/cometa/store.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

- [ ] 2.5 Janela absoluta de retry e classificação de falha
  - Objective: Entregar o comportamento R2-INT-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-09, EXE-10, CFG-04.
  - Likely files/components: `hub/internal/cometa`; `hub/internal/atlasclient/client.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.

- [ ] 2.6 Controle adaptativo global por domínio de capacidade
  - Objective: Entregar o comportamento R2-INT-06 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-07, OPE-08, EXE-13, CAT-06.
  - Likely files/components: `hub/internal/cometa`; `hub/internal/atlas/store.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-06; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.6; sem flags que apresentem mock como fluxo real.

- [ ] 2.7 Entrega ao cliente com identidade e política próprias
  - Objective: Entregar o comportamento R2-INT-07 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline EXE-08, COM-05, SEG-02, CFG-03.
  - Likely files/components: `hub/internal/pulsar`; `hub/migrations/core`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-07; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.7; sem flags que apresentem mock como fluxo real.

- [ ] 2.8 SLA do provedor separado de SLA do cliente
  - Objective: Entregar o comportamento R2-INT-08 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-10, EXE-11, FIN-10, CAT-10.
  - Likely files/components: `hub/internal/cometa`; `hub/internal/orbita`; `hub/internal/platform`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-INT-08; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.8; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [x] 3.1 Qualificar R2-INT-01 com oráculos independentes
  - Objective: Executar R2-INT-01-S01, R2-INT-01-S02, R2-INT-01-S03. Caso indispensável: Protocolo especializado; esperado: capacidade fica não disponível; suporte futuro exige adapter e ensaio específico, sem simulação silenciosa via REST.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa/executor.go`; `hub/internal/providersim/providersim.go`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.2 Qualificar R2-INT-02 com oráculos independentes
  - Objective: Executar R2-INT-02-S01, R2-INT-02-S02, R2-INT-02-S03. Caso indispensável: Rotação e cache; esperado: cache antigo não é reutilizado indevidamente; operação já aceita conserva conta, snapshot e pagador.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa/executor.go`; `hub/internal/providerauth/client.go`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.3 Qualificar R2-INT-03 com oráculos independentes
  - Objective: Executar R2-INT-03-S01, R2-INT-03-S02, R2-INT-03-S03. Caso indispensável: Cofre indisponível; esperado: operação afetada aguarda/recusa dentro da política sem credencial inválida; outras contas elegíveis continuam.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/providerauth/client.go`; `hub/cmd/cometa/main.go`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.4 Qualificar R2-INT-04 com oráculos independentes
  - Objective: Executar R2-INT-04-S01, R2-INT-04-S02, R2-INT-04-S03, R2-INT-04-S04. Caso indispensável: Falha de recibo; esperado: ACK 2xx não é emitido; o parceiro pode retransmitir conforme contrato.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa/poller.go`; `hub/internal/cometa/handlers.go`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 4 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.5 Qualificar R2-INT-05 com oráculos independentes
  - Objective: Executar R2-INT-05-S01, R2-INT-05-S02, R2-INT-05-S03, R2-INT-05-S04. Caso indispensável: Timeout após envio; esperado: somente consulta/reconciliação segura é permitida antes de autorizar nova execução.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa`; `hub/internal/atlasclient/client.go`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 4 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.6 Qualificar R2-INT-06 com oráculos independentes
  - Objective: Executar R2-INT-06-S01, R2-INT-06-S02, R2-INT-06-S03, R2-INT-06-S04. Caso indispensável: Ruído de cliente; esperado: B mantém SLO do perfil e reservas; A recebe contenção/recusa explícita em vez de consumir capacidade de B.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa`; `hub/internal/atlas/store.go`.
  - Depends on: 2.6; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 4 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.7 Qualificar R2-INT-07 com oráculos independentes
  - Objective: Executar R2-INT-07-S01, R2-INT-07-S02, R2-INT-07-S03, R2-INT-07-S04. Caso indispensável: Duplicata no receptor; esperado: event_id/result_version identificam o mesmo final e assinatura/timestamp são verificáveis.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/pulsar`; `hub/migrations/core`.
  - Depends on: 2.7; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 4 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.8 Qualificar R2-INT-08 com oráculos independentes
  - Objective: Executar R2-INT-08-S01, R2-INT-08-S02, R2-INT-08-S03. Caso indispensável: Coorte ainda aberta; esperado: painel mostra elegíveis, abertos, cumpridos, vencidos e exclusões, sem contar abertos como sucesso.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/cometa`; `hub/internal/orbita`.
  - Depends on: 2.8; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Introduzir adapters por capacidade e manter provider-sim só no perfil de ensaio. Migrar bindings explícitos e invalidar cache por conta antigo; em operações existentes, fixar a identidade da conta sem trocar o pagador durante rotação.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/internal/cometa`; `hub/internal/providerauth/client.go`; `hub/internal/pulsar`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-03-integracoes-credenciais-e-pressao`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
