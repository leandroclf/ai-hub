# Tasks: Ambientes completos, disponibilidade, escala e observabilidade

Status: **todas abertas**. Este arquivo planeja implementação futura; esta revisão só produziu documentação. Dependências de change: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## 1. Contratos e preparação

- [ ] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.

- [ ] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
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

- [ ] 2.1 Plataforma local completa e persistente
  - Objective: Entregar o comportamento R2-OPE-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-04, QUA-01, DAD-07.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [ ] 2.2 Kubernetes e cinco ambientes coerentes
  - Objective: Entregar o comportamento R2-OPE-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-04, OPE-05, ARQ-06, ARQ-01.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [ ] 2.3 Probes por capacidade e drenagem
  - Objective: Entregar o comportamento R2-OPE-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-05, OPE-13, OPE-14.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [ ] 2.4 Escala de pods, nós e dados sem chamados rotineiros
  - Objective: Entregar o comportamento R2-OPE-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-12, OPE-08, CFG-06, DAD-11, ARQ-06.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

- [ ] 2.5 Placement, fencing e recursos por ambiente e célula
  - Objective: Entregar o comportamento R2-OPE-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline DAD-11, ARQ-06, CFG-06, OPE-08.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.

- [ ] 2.6 Fallback seletivo e dependências mínimas
  - Objective: Entregar o comportamento R2-OPE-06 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-13, DAD-09, DAD-10, ARQ-02.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-06; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.6; sem flags que apresentem mock como fluxo real.

- [ ] 2.7 Telemetria e SLA com investigação acionável
  - Objective: Entregar o comportamento R2-OPE-07 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-03, OPE-06, OPE-10.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-07; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.7; sem flags que apresentem mock como fluxo real.

- [ ] 2.8 Segurança e reprodutibilidade de infraestrutura
  - Objective: Entregar o comportamento R2-OPE-08 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline ARQ-05, SEG-01, OPE-04, QUA-04.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-08; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.8; sem flags que apresentem mock como fluxo real.

- [ ] 2.9 SLO, isolamento e continuidade qualificados
  - Objective: Entregar o comportamento R2-OPE-09 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-01, OPE-02, OPE-03, OPE-09, OPE-11, OPE-15.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/internal/atlasclient`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-OPE-09; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.9; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [ ] 3.1 Qualificar R2-OPE-01 com oráculos independentes
  - Objective: Executar R2-OPE-01-S01, R2-OPE-01-S02, R2-OPE-01-S03. Caso indispensável: Cache desligado; esperado: serviços não falham por dependência do cache opcional.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.2 Qualificar R2-OPE-02 com oráculos independentes
  - Objective: Executar R2-OPE-02-S01, R2-OPE-02-S02, R2-OPE-02-S03. Caso indispensável: Produção sem perfil; esperado: gate bloqueia ativação e mostra pendências sem impedir desenvolvimento local.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.3 Qualificar R2-OPE-03 com oráculos independentes
  - Objective: Executar R2-OPE-03-S01, R2-OPE-03-S02, R2-OPE-03-S03. Caso indispensável: Provedor fora; esperado: não reinicia todos os pods; circuito/quotas contêm apenas a capacidade afetada.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.4 Qualificar R2-OPE-04 com oráculos independentes
  - Objective: Executar R2-OPE-04-S01, R2-OPE-04-S02, R2-OPE-04-S03. Caso indispensável: Quota esgotada; esperado: estado registra impedimento e novas admissões são limitadas; não rouba reserva de clientes existentes.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.5 Qualificar R2-OPE-05 com oráculos independentes
  - Objective: Executar R2-OPE-05-S01, R2-OPE-05-S02, R2-OPE-05-S03. Caso indispensável: Falha parcial de provisionamento; esperado: não atribui tráfego até recursos e canário estarem qualificados; retry de reconcile não duplica dono.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.6 Qualificar R2-OPE-06 com oráculos independentes
  - Objective: Executar R2-OPE-06-S01, R2-OPE-06-S02, R2-OPE-06-S03. Caso indispensável: Todas cópias autoritativas indisponíveis; esperado: Hub informa indisponibilidade sem 202 fictício; preserva repetição pela mesma chave quando voltar.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.6; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.7 Qualificar R2-OPE-07 com oráculos independentes
  - Objective: Executar R2-OPE-07-S01, R2-OPE-07-S02, R2-OPE-07-S03. Caso indispensável: Backend de observabilidade fora; esperado: buffer/export limitado não bloqueia thread de negócio; perda de telemetria é sinalizada e não altera ledger/auditoria duráveis.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.7; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.8 Qualificar R2-OPE-08 com oráculos independentes
  - Objective: Executar R2-OPE-08-S01, R2-OPE-08-S02, R2-OPE-08-S03. Caso indispensável: Atualização de dependência; esperado: registra suporte/licença, lock/digest, scan e ensaios de contrato; referência de versão não vira aprovação automática.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.8; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.9 Qualificar R2-OPE-09 com oráculos independentes
  - Objective: Executar R2-OPE-09-S01, R2-OPE-09-S02, R2-OPE-09-S03. Caso indispensável: Uso crítico; esperado: ativação é bloqueada para esse perfil, com contingência e pendências explicitadas.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy`; `hub/cmd`.
  - Depends on: 2.9; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Criar overlays e papéis de runtime preservando cinco aplicações lógicas; implantar controllers antes de seus recursos; migrar filas por célula sem duas autoridades. Executar prova local/kind antes de promoção.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/deploy`; `hub/cmd`; `hub/internal/platform`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-08-operacao-elastica-e-observavel`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
