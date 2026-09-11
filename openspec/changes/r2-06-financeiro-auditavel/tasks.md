# Tasks: Medição, saldo estrito, ledger e fechamento

Status: **comportamentos e provas concluídos; integração/migração e fechamento permanecem abertos**. Dependências de change: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados, r2-04-catalogo-produtos-e-contratos

## 1. Contratos e preparação

- [x] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.

- [x] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
  - Depends on: 1.1.
  - Validation: Contract/schema review e casos inválidos; registrar quais contratos precisam nova versão.
  - Completion criteria: DTOs e versões acordados; nenhuma alteração incompatível implícita no perfil v1.

- [x] 1.3 Preparar evolução aditiva e fixtures isoladas
  - Objective: Criar migrations adicionais quando aplicável, permissões, interfaces ou organização documental/UI necessária; separar fixtures de dados reais.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 1.2.
  - Validation: Aplicação em ambiente limpo e existente, rollback compatível, validação de unicidade/proveniência; não editar migration aplicada.
  - Completion criteria: Estrutura suporta as regras sem perda de histórico; para UI/qualificação, registrar explicitamente ausência de mudança de esquema quando confirmada.

## 2. Comportamentos

- [x] 2.1 Compra, venda e incidência por snapshot
  - Objective: Entregar o comportamento R2-FIN-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-01, FIN-02, FIN-03, FIN-05, FIN-11.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/migrations/finance`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-FIN-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [x] 2.2 Medição exata e deduplicação econômica
  - Objective: Entregar o comportamento R2-FIN-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-04, FIN-05, FIN-09.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/migrations/finance`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-FIN-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [x] 2.3 Saldo estrito e retenção de incerteza
  - Objective: Entregar o comportamento R2-FIN-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-06, FIN-10, DAD-03, DAD-11.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/migrations/finance`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-FIN-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [x] 2.4 Planos, franquias e política de produto
  - Objective: Entregar o comportamento R2-FIN-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-02, FIN-05, FIN-09, CAT-07.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/migrations/finance`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-FIN-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

- [x] 2.5 Ledger imutável e ajustes compensatórios
  - Objective: Entregar o comportamento R2-FIN-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-07, FIN-10.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/migrations/finance`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-FIN-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.

- [x] 2.6 Fechamento, reconciliação e integração financeira
  - Objective: Entregar o comportamento R2-FIN-06 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-08, FIN-10, FIN-11.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/migrations/finance`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-FIN-06; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.6; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [x] 3.1 Qualificar R2-FIN-01 com oráculos independentes
  - Objective: Executar R2-FIN-01-S01, R2-FIN-01-S02, R2-FIN-01-S03. Caso indispensável: Compra por submit; esperado: custo e ausência de receita seguem marcos distintos, sem preço constante de simulador.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/libra`; `hub/internal/libraclient/client.go`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.2 Qualificar R2-FIN-02 com oráculos independentes
  - Objective: Executar R2-FIN-02-S01, R2-FIN-02-S02, R2-FIN-02-S03. Caso indispensável: Precisão; esperado: resultado decimal é exato e arredondamento é único no marco contratado.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/libra`; `hub/internal/libraclient/client.go`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.3 Qualificar R2-FIN-03 com oráculos independentes
  - Objective: Executar R2-FIN-03-S01, R2-FIN-03-S02, R2-FIN-03-S03, R2-FIN-03-S04. Caso indispensável: Timeout com efeito incerto; esperado: hold permanece até evidência de cancelamento/ausência/custo; duplicata não libera nem captura duas vezes.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/libra`; `hub/internal/libraclient/client.go`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 4 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.4 Qualificar R2-FIN-04 com oráculos independentes
  - Objective: Executar R2-FIN-04-S01, R2-FIN-04-S02, R2-FIN-04-S03. Caso indispensável: Franquia final concorrente; esperado: uma usa franquia e a outra segue regra de excedente ou recusa explicitamente contratada.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/libra`; `hub/internal/libraclient/client.go`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.5 Qualificar R2-FIN-05 com oráculos independentes
  - Objective: Executar R2-FIN-05-S01, R2-FIN-05-S02, R2-FIN-05-S03. Caso indispensável: Estorno autorizado; esperado: partidas compensatórias preservam original, razão e correlação.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/libra`; `hub/internal/libraclient/client.go`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.6 Qualificar R2-FIN-06 com oráculos independentes
  - Objective: Executar R2-FIN-06-S01, R2-FIN-06-S02, R2-FIN-06-S03. Caso indispensável: Contestação de SLA; esperado: valor, evidência, disputa e ajuste ficam separados da resposta final imutável do cliente.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/libra`; `hub/internal/libraclient/client.go`.
  - Depends on: 2.6; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Migrar decimal sem converter por float; carregar fatos antigos em lote de abertura com proveniência LEGACY_UNVERIFIED onde snapshot faltar. Não inventar tarifas históricas. Bloquear fechamento definitivo até reconciliação e aprovação.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/internal/libra`; `hub/internal/libraclient/client.go`; `hub/internal/atlas/store.go`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-06-financeiro-auditavel`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
