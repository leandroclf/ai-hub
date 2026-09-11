# Tasks: Arquivos, propriedade de dados, retenção e recuperação

Status: **parcialmente concluído**. Upload, custódia de resultado e retenção/replay (R2-DAD-03) têm prova local atualizada; RLS/continuidade, restore completo e cenários regionais permanecem abertos onde faltam provas completas. Dependências de change: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## 1. Contratos e preparação

- [ ] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.

- [ ] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
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

- [x] 2.1 Upload direto e referência imutável por tenant
  - Objective: Entregar o comportamento R2-DAD-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline DAD-05, COM-02, SEG-02.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `hub/deploy/postgres-init/01-init.sh`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-DAD-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [x] 2.2 Custódia de resultado volumoso
  - Objective: Entregar o comportamento R2-DAD-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline DAD-05, DAD-04, DAD-09, OPE-11.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `hub/deploy/postgres-init/01-init.sh`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-DAD-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [x] 2.3 Retenção por classe e obrigações abertas
  - Objective: Entregar o comportamento R2-DAD-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline DAD-06, DAD-08, OPE-11.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `hub/deploy/postgres-init/01-init.sh`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-DAD-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.
  - Evidence: `hub/evidence/r2/execution/r2-dad-03-runtime-latest.log`; retenção por pin/tombstone, obrigação aberta, falha de storage e replay antigo foram exercitados com PostgreSQL/LocalStack.

- [x] 2.4 Propriedade, RLS e continuidade de leitura
  - Objective: Entregar o comportamento R2-DAD-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline DAD-01, DAD-07, DAD-10, SEG-03.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `hub/deploy/postgres-init/01-init.sh`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-DAD-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.
  - Evidence: `hub/evidence/r2/execution/r2-dad-04-read-authority-latest.log`; leitura pendente/expirada, indisponibilidade explícita e RLS nas três bases foram comprovados.

- [x] 2.5 Restauração reconciliada sem duplicar efeito
  - Objective: Entregar o comportamento R2-DAD-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline DAD-07, OPE-11, OPE-15.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `hub/deploy/postgres-init/01-init.sh`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-DAD-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.
  - Evidence: `hub/evidence/r2/execution/restore-reconciliation-latest.log`; restore isolado comparou contagens/digests de control/core/finance, observou o oráculo externo sem replay e manteve a readmissão bloqueada.

## 3. Provas por requisito

- [x] 3.1 Qualificar R2-DAD-01 com oráculos independentes
  - Objective: Executar R2-DAD-01-S01, R2-DAD-01-S02, R2-DAD-01-S03. Caso indispensável: Multipart incompleto; esperado: referência não fica READY; sessão pode ser retomada/expirada conforme política.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/objectstore`; `hub/internal/orbita`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.2 Qualificar R2-DAD-02 com oráculos independentes
  - Objective: Executar R2-DAD-02-S01, R2-DAD-02-S02, R2-DAD-02-S03. Caso indispensável: Objeto sem commit do protocolo; esperado: liga objeto somente à obrigação correta ou marca órfão para política segura, sem reexecutar provedor.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/objectstore`; `hub/internal/orbita`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.3 Qualificar R2-DAD-03 com oráculos independentes
  - Objective: Executar R2-DAD-03-S01, R2-DAD-03-S02, R2-DAD-03-S03. Caso indispensável: Mensagem antiga; esperado: tombstone/dedup impede novo efeito e produz diagnóstico.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/objectstore`; `hub/internal/orbita`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.
  - Evidence: `hub/evidence/r2/execution/r2-dad-03-runtime-latest.log`; S01, S02 e S03 = PASS, sem novo efeito no replay e com tombstone/quarentena observáveis.

- [ ] 3.4 Qualificar R2-DAD-04 com oráculos independentes
  - Objective: Executar R2-DAD-04-S01, R2-DAD-04-S02, R2-DAD-04-S03. Caso indispensável: Isolamento de domínio; esperado: nega escrita fora da autoridade mesmo que aplicação tenha bug.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/objectstore`; `hub/internal/orbita`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.5 Qualificar R2-DAD-05 com oráculos independentes
  - Objective: Executar R2-DAD-05-S01, R2-DAD-05-S02, R2-DAD-05-S03. Caso indispensável: Falha regional; esperado: perfil não é qualificado nem ativado até prova de custódia e fencing compatíveis.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/internal/objectstore`; `hub/internal/orbita`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Criar catálogo de objetos e permissões sem reescrever finais históricos; materializar referência somente depois de validar existência e dono. Backfill separado e retomável; nunca criar file_ref com objeto não confirmado.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/internal/objectstore`; `hub/internal/orbita`; `hub/migrations`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-07-objetos-dados-e-retencao`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
