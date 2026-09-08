# Tasks: Console administrativo completo e orientado às jornadas

Status: **todas abertas**. Este arquivo planeja implementação futura; esta revisão só produziu documentação. Dependências de change: r2-01-identidade-e-isolamento

## 1. Contratos e preparação

- [ ] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/admin-ui/src`; `hub/admin-ui/package.json`; `hub/admin-ui/vite.config.ts`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.

- [ ] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/admin-ui/src`; `hub/admin-ui/package.json`; `hub/admin-ui/vite.config.ts`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
  - Depends on: 1.1.
  - Validation: Contract/schema review e casos inválidos; registrar quais contratos precisam nova versão.
  - Completion criteria: DTOs e versões acordados; nenhuma alteração incompatível implícita no perfil v1.

- [ ] 1.3 Preparar evolução aditiva e fixtures isoladas
  - Objective: Criar migrations adicionais quando aplicável, permissões, interfaces ou organização documental/UI necessária; separar fixtures de dados reais.
  - Likely files/components: `hub/admin-ui/src`; `hub/admin-ui/package.json`; `hub/admin-ui/vite.config.ts`.
  - Depends on: 1.2.
  - Validation: Aplicação em ambiente limpo e existente, rollback compatível, validação de unicidade/proveniência; não editar migration aplicada.
  - Completion criteria: Estrutura suporta as regras sem perda de histórico; para UI/qualificação, registrar explicitamente ausência de mudança de esquema quando confirmada.

## 2. Comportamentos

- [ ] 2.1 Sessão, contexto e navegação administrativa
  - Objective: Entregar o comportamento R2-ADM-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-01, SEG-01, SEG-04.
  - Likely files/components: `hub/admin-ui/src/App.tsx`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [ ] 2.2 Listagens reais e detalhe editável
  - Objective: Entregar o comportamento R2-ADM-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-01, CAT-01, CFG-06.
  - Likely files/components: `hub/admin-ui/src/pages/shared.ts`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [ ] 2.3 Clientes, aplicações e ofertas contratadas
  - Objective: Entregar o comportamento R2-ADM-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-01, CFG-06, CAT-02, CAT-11.
  - Likely files/components: `hub/admin-ui/src/pages/ClientsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [ ] 2.4 Catálogo importado e serviços executáveis
  - Objective: Entregar o comportamento R2-ADM-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-01, CAT-02, CFG-02, COM-02.
  - Likely files/components: `hub/admin-ui/src/pages/ServicesPage.tsx`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

- [ ] 2.5 Produtos agregados e compostos
  - Objective: Entregar o comportamento R2-ADM-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-03, CAT-04, CAT-05, CAT-07, CAT-08.
  - Likely files/components: `hub/admin-ui/src/pages/ProductsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.

- [ ] 2.6 Provedores, vínculos e saúde de integração
  - Objective: Entregar o comportamento R2-ADM-06 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-05, SEG-05, OPE-07, CAT-06.
  - Likely files/components: `hub/admin-ui/src/pages/ProviderAccountsPage.tsx`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-06; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.6; sem flags que apresentem mock como fluxo real.

- [ ] 2.7 Contratos técnicos e políticas temporais
  - Objective: Entregar o comportamento R2-ADM-07 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CAT-09, CFG-04, FIN-03, EXE-10, EXE-11.
  - Likely files/components: `hub/admin-ui/src/pages/TechnicalProfilesPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-07; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.7; sem flags que apresentem mock como fluxo real.

- [ ] 2.8 Busca e timeline operacional de protocolos
  - Objective: Entregar o comportamento R2-ADM-08 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-03, SEG-04, EXE-07, DAD-04.
  - Likely files/components: `hub/admin-ui/src/pages/ProtocolsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-08; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.8; sem flags que apresentem mock como fluxo real.

- [ ] 2.9 Entregas e reprocessamento com segurança de efeito
  - Objective: Entregar o comportamento R2-ADM-09 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-03, EXE-08, EXE-09, SEG-04.
  - Likely files/components: `hub/admin-ui/src/pages/DeliveriesPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-09; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.9; sem flags que apresentem mock como fluxo real.

- [ ] 2.10 SLA bilateral e painéis operacionais consultáveis
  - Objective: Entregar o comportamento R2-ADM-10 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline OPE-06, OPE-10, CFG-03.
  - Likely files/components: `hub/admin-ui/src/pages/SLAReportsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-10; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.10; sem flags que apresentem mock como fluxo real.

- [ ] 2.11 Financeiro de compra, venda e conciliação
  - Objective: Entregar o comportamento R2-ADM-11 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline FIN-01, FIN-02, FIN-06, FIN-07, FIN-08, FIN-10.
  - Likely files/components: `hub/admin-ui/src/pages/FinancePage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-11; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.11; sem flags que apresentem mock como fluxo real.

- [ ] 2.12 Qualidade de uso, acessibilidade e integração real
  - Objective: Entregar o comportamento R2-ADM-12 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-01, QUA-03, QUA-04.
  - Likely files/components: `hub/admin-ui/src/index.css`; `hub/admin-ui/src/api/atlasClient.ts`; `hub/api/openapi-internal.yaml`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-ADM-12; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.12; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [ ] 3.1 Qualificar R2-ADM-01 com oráculos independentes
  - Objective: Executar R2-ADM-01-S01, R2-ADM-01-S02, R2-ADM-01-S03. Caso indispensável: Logout; esperado: não reaparecem dados, filtros sensíveis ou permissões do anterior.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/App.tsx`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.2 Qualificar R2-ADM-02 com oráculos independentes
  - Objective: Executar R2-ADM-02-S01, R2-ADM-02-S02, R2-ADM-02-S03. Caso indispensável: Conflito de edição; esperado: UI apresenta conflito e diff/recarregamento sem sobrescrita silenciosa.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/shared.ts`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.3 Qualificar R2-ADM-03 com oráculos independentes
  - Objective: Executar R2-ADM-03-S01, R2-ADM-03-S02, R2-ADM-03-S03. Caso indispensável: Suspensão; esperado: confirma impacto e razão; timeline histórica continua acessível ao papel autorizado.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/ClientsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.4 Qualificar R2-ADM-04 com oráculos independentes
  - Objective: Executar R2-ADM-04-S01, R2-ADM-04-S02, R2-ADM-04-S03. Caso indispensável: Sem adapter; esperado: estado deixa claro que não está disponível para consumo.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/ServicesPage.tsx`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.5 Qualificar R2-ADM-05 com oráculos independentes
  - Objective: Executar R2-ADM-05-S01, R2-ADM-05-S02, R2-ADM-05-S03. Caso indispensável: Simulação sem efeitos; esperado: nenhuma chamada faturável real é feita; relatório identifica fixture e versões.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/ProductsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.6 Qualificar R2-ADM-06 com oráculos independentes
  - Objective: Executar R2-ADM-06-S01, R2-ADM-06-S02, R2-ADM-06-S03. Caso indispensável: Pressão reduzida; esperado: vê limite efetivo, teto, latência, erro, backlog e motivo da redução, sem tratá-lo como limite fixo eterno.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/ProviderAccountsPage.tsx`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.6; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.7 Qualificar R2-ADM-07 com oráculos independentes
  - Objective: Executar R2-ADM-07-S01, R2-ADM-07-S02, R2-ADM-07-S03. Caso indispensável: SYNC incompatível; esperado: relatório aponta incompatibilidade e bloqueia publicação.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/TechnicalProfilesPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.7; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.8 Qualificar R2-ADM-08 com oráculos independentes
  - Objective: Executar R2-ADM-08-S01, R2-ADM-08-S02, R2-ADM-08-S03. Caso indispensável: Filtro entre tenants; esperado: API e UI não revelam protocolos alheios.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/ProtocolsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.8; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.9 Qualificar R2-ADM-09 com oráculos independentes
  - Objective: Executar R2-ADM-09-S01, R2-ADM-09-S02, R2-ADM-09-S03. Caso indispensável: Leitor administrativo; esperado: servidor nega; possuir leitura global não concede operação.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/DeliveriesPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.9; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.10 Qualificar R2-ADM-10 com oráculos independentes
  - Objective: Executar R2-ADM-10-S01, R2-ADM-10-S02, R2-ADM-10-S03. Caso indispensável: Consulta limitada; esperado: arquivo contém somente registros permitidos e exportação é auditada.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/SLAReportsPage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.10; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.11 Qualificar R2-ADM-11 com oráculos independentes
  - Objective: Executar R2-ADM-11-S01, R2-ADM-11-S02, R2-ADM-11-S03. Caso indispensável: Estorno; esperado: lançamento compensatório aparece no histórico; original não desaparece.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/pages/FinancePage.tsx (novo proposto)`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.11; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [ ] 3.12 Qualificar R2-ADM-12 com oráculos independentes
  - Objective: Executar R2-ADM-12-S01, R2-ADM-12-S02, R2-ADM-12-S03. Caso indispensável: Jornada integrada; esperado: estado persiste no servidor; evidência inclui papel, SHA, requests e screenshot sanitizado, não só renderização.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/admin-ui/src/index.css`; `hub/admin-ui/src/api/atlasClient.ts`.
  - Depends on: 2.12; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [ ] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Preservar acessos aos quatro cadastros com rotas novas e listagens reais; manter UI antiga restrita até migração, sem permitir escrita anônima. Introduzir por jornadas disponíveis e feature flags por permissão.
  - Likely files/components: `hub/admin-ui/src`; `hub/admin-ui/package.json`; `hub/admin-ui/vite.config.ts`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.

- [ ] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/admin-ui/src`; `hub/admin-ui/package.json`; `hub/admin-ui/vite.config.ts`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.

- [ ] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-05-console-administrativo`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
