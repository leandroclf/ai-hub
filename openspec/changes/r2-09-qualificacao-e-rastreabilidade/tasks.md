# Tasks: Qualificação por requisito e evidência reproduzível

Status: **preparação e evidências parciais concluídas**. A revisão atualiza o snapshot, contratos compartilhados e fixtures isoladas; os comportamentos e a qualificação integral permanecem abertos. A matriz atual contém 732 cenários, com 231 associados a resultados/evidências e 501 `NAO_QUALIFICADO_NESTA_RODADA`. Dependências de change: Sem pré-requisito de implementação para iniciar; integrações específicas descritas abaixo.

## 1. Contratos e preparação

- [x] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação. Evidência: `docs/reviews/2026-09-09-r4/implementation/OPENSPEC-AUDIT-2026-09-10.md`, `INVENTORY-732-CENARIOS.csv` e `RESULT-MATRIX-732-CENARIOS.csv` no SHA atual.

- [x] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
  - Depends on: 1.1.
  - Validation: Contract/schema review e casos inválidos; registrar quais contratos precisam nova versão.
  - Completion criteria: DTOs e versões acordados; nenhuma alteração incompatível implícita no perfil v1. Evidência: revisão dos contratos `hub/api/openapi.yaml`, `openapi-internal.yaml` e `asyncapi.yaml`, com validação OpenSpec strict 21/21 e provas de contrato registradas no índice de evidências.

- [x] 1.3 Preparar evolução aditiva e fixtures isoladas
  - Objective: Criar migrations adicionais quando aplicável, permissões, interfaces ou organização documental/UI necessária; separar fixtures de dados reais.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`.
  - Depends on: 1.2.
  - Validation: Aplicação em ambiente limpo e existente, rollback compatível, validação de unicidade/proveniência; não editar migration aplicada.
  - Completion criteria: Estrutura suporta as regras sem perda de histórico; para UI/qualificação, registrar explicitamente ausência de mudança de esquema quando confirmada. Evidência: migrações aditivas e provas de recuo em `hub/evidence/r2/execution`, com fixtures de qualificação separadas do volume oficial.

## 2. Comportamentos

- [ ] 2.1 Rastreabilidade do alvo até evidência
  - Objective: Entregar o comportamento R2-QUA-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline QUA-03, QUA-04, DEC-02, DEC-05.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`; `hub/internal/atlas/handlers_test.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-QUA-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [ ] 2.2 Fixtures e oráculos independentes reproduzíveis
  - Objective: Entregar o comportamento R2-QUA-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline QUA-01, QUA-02, QUA-05, QUA-06.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`; `hub/internal/atlas/handlers_test.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-QUA-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [ ] 2.3 Gates de contrato, experiência e operação
  - Objective: Entregar o comportamento R2-QUA-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline QUA-04, QUA-06, COM-04, OPE-15.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`; `hub/internal/atlas/handlers_test.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-QUA-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [ ] 2.4 Evolução OpenSpec e baseline sem falso arquivamento
  - Objective: Entregar o comportamento R2-QUA-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline DEC-01, DEC-03, DEC-04, DEC-05, QUA-03, ARQ-04.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`; `hub/internal/atlas/handlers_test.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-QUA-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [ ] 3.1 Qualificar R2-QUA-01 com oráculos independentes
  - Objective: Executar R2-QUA-01-S01, R2-QUA-01-S02, R2-QUA-01-S03. Caso indispensável: Decisão ainda pendente; esperado: pendência permanece aberta; não presume avanço por commit ou lembrete.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.2 Qualificar R2-QUA-02 com oráculos independentes
  - Objective: Executar R2-QUA-02-S01, R2-QUA-02-S02, R2-QUA-02-S03. Caso indispensável: Regressão financeira; esperado: falha se terceira chamada externa ocorrer.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.3 Qualificar R2-QUA-03 com oráculos independentes
  - Objective: Executar R2-QUA-03-S01, R2-QUA-03-S02, R2-QUA-03-S03. Caso indispensável: P0 reaberto; esperado: ativação do fluxo afetado é bloqueada até corrigir ou manter explicitamente indisponível.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.4 Qualificar R2-QUA-04 com oráculos independentes
  - Objective: Executar R2-QUA-04-S01, R2-QUA-04-S02, R2-QUA-04-S03. Caso indispensável: Regressão posterior; esperado: status reabre com referência ao novo SHA, sem apagar prova histórica nem reduzir requisito.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Manter v4 como baseline normativo com lacunas abertas. R2 acrescenta deltas, não declara a v4 implementada. Reconciliação futura da árvore canônica deve preservar IDs e separar aceitação de especificação de pronto operacional.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4`; `IMPLEMENTATION_AUDIT.md`; `hub/test/e2e/e2e_test.go`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-09-qualificacao-e-rastreabilidade`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
