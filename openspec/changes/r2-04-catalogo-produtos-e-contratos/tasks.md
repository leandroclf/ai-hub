# Tasks: Catálogo operável, produtos compostos e contratos por cliente

Status: **parcialmente concluído**. Catálogo versionado e DAG/compensação têm prova integral; ofertas, agregação, perfis e onboarding permanecem abertos onde faltam cenários. Dependências de change: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## 1. Contratos e preparação

- [ ] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.

- [ ] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
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

- [x] 2.1 Catálogo versionado com publicação governada
  - Objective: Entregar o comportamento R2-CAT-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-01, CAT-02, CFG-02, CFG-04.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [x] 2.2 Ofertas por cliente e roteamento autorizado
  - Objective: Entregar o comportamento R2-CAT-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-02, CAT-06, CAT-10, CAT-11, FIN-02.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [x] 2.3 Agregação paralela limitada
  - Objective: Entregar o comportamento R2-CAT-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-03, CAT-05, CAT-08, EXE-13.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [x] 2.4 Composição por DAG e compensações
  - Objective: Entregar o comportamento R2-CAT-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-04, CAT-05, CAT-08, EXE-03, EXE-13.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

- [ ] 2.5 Perfis técnicos de entrada e saída por cliente
  - Objective: Entregar o comportamento R2-CAT-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-09, COM-02, COM-04, COM-05.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.

- [ ] 2.6 Snapshot e projeções independentes de consulta por pedido
  - Objective: Entregar o comportamento R2-CAT-06 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-03, CFG-02, DAD-08, DAD-10, ARQ-02.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-06; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.6; sem flags que apresentem mock como fluxo real.

- [ ] 2.7 Importação segura e inventário executável distinto
  - Objective: Entregar o comportamento R2-CAT-07 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-01, CAT-02, CFG-02, COM-02, QUA-01.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-07; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.7; sem flags que apresentem mock como fluxo real.

- [x] 2.8 Consulta administrativa persistente e onboarding
  - Objective: Entregar o comportamento R2-CAT-08 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-01, CFG-06, CAT-11, DAD-11.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/migrations/control`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-CAT-08; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.8; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [x] 3.1 Qualificar R2-CAT-01 com oráculos independentes
  - Objective: Executar R2-CAT-01-S01, R2-CAT-01-S02, R2-CAT-01-S03. Caso indispensável: Edição concorrente; esperado: servidor detecta versão desatualizada e oferece diff/recarregamento sem sobrescrever trabalho alheio.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.2 Qualificar R2-CAT-02 com oráculos independentes
  - Objective: Executar R2-CAT-02-S01, R2-CAT-02-S02, R2-CAT-02-S03. Caso indispensável: Sem combinação de SLA viável; esperado: recebe bloqueios específicos e oferta não é ativada.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.3 Qualificar R2-CAT-03 com oráculos independentes
  - Objective: Executar R2-CAT-03-S01, R2-CAT-03-S02, R2-CAT-03-S03. Caso indispensável: Fan-out abusivo; esperado: configuração/pedido é recusado antes de exceder recursos de outros clientes.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.4 Qualificar R2-CAT-04 com oráculos independentes
  - Objective: Executar R2-CAT-04-S01, R2-CAT-04-S02, R2-CAT-04-S03. Caso indispensável: Compensação incompleta; esperado: protocolo informa estado contratado e gera obrigação de reconciliação, sem declarar rollback integral fictício.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.5 Qualificar R2-CAT-05 com oráculos independentes
  - Objective: Executar R2-CAT-05-S01, R2-CAT-05-S02, R2-CAT-05-S03. Caso indispensável: Upgrade de schema; esperado: mudança incompatível exige nova versão e migração; protocolos v1 conservam seu resultado.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.6 Qualificar R2-CAT-06 com oráculos independentes
  - Objective: Executar R2-CAT-06-S01, R2-CAT-06-S02, R2-CAT-06-S03. Caso indispensável: Revogação de segurança; esperado: política de revogação de segurança prevalece para novos acessos ao segredo; preserva histórico sem reutilizar credencial proibida.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.6; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.7 Qualificar R2-CAT-07 com oráculos independentes
  - Objective: Executar R2-CAT-07-S01, R2-CAT-07-S02, R2-CAT-07-S03. Caso indispensável: Credenciais na collection; esperado: valores secretos não aparecem no lote, logs ou UI; somente referências e metadados permitidos são conservados.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.7; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.8 Qualificar R2-CAT-08 com oráculos independentes
  - Objective: Executar R2-CAT-08-S01, R2-CAT-08-S02, R2-CAT-08-S03. Caso indispensável: Capacidade insuficiente; esperado: registra provisionamento automático e só ativa após qualificação; não aponta tráfego para célula incompleta.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/atlas`; `hub/internal/atlasclient`.
  - Depends on: 2.8; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Transformar catálogo atual em rascunhos de origem rastreável; manter IDs e consultas históricas. Campos herdados por tenant viram defaults explícitos, nunca autorização a qualquer serviço. Reconciliar importação com diff sem limpeza global.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/internal/atlas`; `hub/internal/atlasclient`; `hub/internal/orbita`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-04-catalogo-produtos-e-contratos`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
