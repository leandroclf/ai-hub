# Tasks: Callbacks autenticados e recuperáveis

Status: **R4-CBK-02/03/04 implementados e qualificados no laboratório; R4-CBK-01 e promoção externa permanecem abertas**. A suíte atual confirma custódia, inbox, recuperação, correlação e validação comum; autenticação de origem por conta no gateway ainda exige homologação específica.
## 1. Revalidação
- [x] 1.1 Confirmar HEAD/diff e fontes atuais.
  - Objective: preservar correções e verificar mudanças posteriores ao snapshot.
  - Likely files/components: explore.md e arquivos citados.
  - Depends on: AGENTS e prompt R4.
  - Validation: comparação de conteúdo e registro de evidência.
  - Completion criteria: cada achado tem estado atual verificável.

## 2. R4-CBK-01 — Rota de callback compatível com a autenticação do provedor
- [ ] 2.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-01 observável sem enfraquecer regra.
  - Likely files/components: hub/cmd/cometa/main.go, hub/internal/cometa/handlers.go, hub/internal/cometa/executor.go, hub/internal/platform/httpserver/httpserver.go.
  - Depends on: 1.1.
  - Validation: R4-CBK-01-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
- [ ] 2.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL expor uma rota de callback com autenticação explicitamente homologada por conta e independente da credencial de workload interna. O caminho público/gateway/middleware/handler deve ser qualificado de ponta a ponta. Capability de operação pode complementar correlação, nunca substituir silenciosamente política da conta; segredos não devem aparecer em URLs registradas, logs ou traces.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 2.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
- [ ] 2.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-EXE-02, R2-SEG-04, R2-INT-04.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 2.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.

## 3. R4-CBK-02 — Admissão e deduplicação segura da inbox órfã
- [x] 3.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-02 observável sem enfraquecer regra.
  - Likely files/components: hub/internal/cometa/handlers.go, hub/internal/cometa/custody.go, hub/migrations/core/0032_callback_inbox.sql.
  - Depends on: 1.1.
  - Validation: R4-CBK-02-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
  - Evidence: `openspec/changes/r4-01-callbacks-autenticados-e-recuperaveis/risk-matrix.md`, `explore.md` e `hub/evidence/r2/execution/r4-cbk-regression-latest.log`; os contraexemplos de token, deduplicação, quota e escopo foram reproduzidos e delimitados.
- [ ] 3.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL autenticar a origem antes de admitir órfão e vincular custódia a tenant/conta/célula/política e identidade do evento. Uma tentativa inválida não pode ocupar a chave de deduplicação de recibo legítimo. Quota, retenção, autorização e disposição devem ser explícitas; conflito de identidade mantém evidência sem descartar o evento correto.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 3.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
- [ ] 3.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-EXE-02, R3-OPE-04, R2-SEG-04.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 3.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.

## 4. R4-CBK-03 — Recuperador autônomo limitado e independente do HTTP
- [x] 4.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-03 observável sem enfraquecer regra.
  - Likely files/components: hub/internal/cometa/custody.go, hub/internal/cometa/handlers.go.
  - Depends on: 1.1.
  - Validation: R4-CBK-03-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
  - Evidence: `hub/evidence/r2/execution/r4-cbk-regression-latest.log`; worker durável, lote limitado, claims, leases, epochs, isolamento e disposição de item inválido foram exercitados com PostgreSQL real.
- [x] 4.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL reconciliar inbox por worker durável com lote limitado, claim/lease/epoch e isolamento por célula/conta/tenant. Aceite de callback conhecido não deve aguardar drenagem global. A recuperação deve ocorrer após reinício/correlação sem depender de tráfego futuro; um item inválido deve ter disposição sem bloquear os seguintes.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 4.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
  - Evidence: `hub/internal/cometa/worker.go`, `custody.go`, migrations `0032_callback_inbox.sql`, `0036_callback_inbox_leases.sql` e `0041_callback_inbox_quota_retention.sql`; o recuperador independente do HTTP foi conectado ao processo do Cometa e passou a suíte `-race`.
- [x] 4.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-EXE-02, R3-EXE-03, R3-INT-01, R2-EXE-09.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 4.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.
  - Evidence: `hub/evidence/r2/execution/r4-cbk-regression-latest.log` e `r4-cbk-latest.log`; S01/S02/S03 passaram localmente, com limitação explícita para dependências externas.

## 5. R4-CBK-04 — Uma validação de resultado para SUBMIT, polling e callback
- [x] 5.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-04 observável sem enfraquecer regra.
  - Likely files/components: hub/internal/cometa/executor.go, hub/internal/cometa/executor.go, hub/internal/cometa/custody.go.
  - Depends on: 1.1.
  - Validation: R4-CBK-04-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
  - Evidence: `openspec/changes/r4-01-callbacks-autenticados-e-recuperaveis/risk-matrix.md`, `explore.md` e `hub/evidence/r2/execution/r4-cbk-regression-latest.log`; divergência de correlação/schema e a validação única foram cobertas.
- [x] 5.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL aplicar uma única validação de resultado externo para todas as modalidades, verificando conta, operação, correlação, schema e snapshot antes de alterar o estado. Conservar recibo original protegido e resultado normalizado versionado; divergência recebe disposição inválida sem substituir correlação/final. Não truncar a resposta a campos de simulador.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 5.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
  - Evidence: `hub/internal/cometa/executor.go` concentra a validação de correlação, snapshot, schema e resultado para polling/callback; recibo bruto e normalizado são preservados separadamente.
- [x] 5.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-EXE-01, R3-EXE-02, R3-CAT-02, R2-EXE-04.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 5.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.
  - Evidence: `hub/evidence/r2/execution/r4-cbk-regression-latest.log`, `representation-runtime-latest.json` e `webhook-retry-representation-smoke.json`; resultado válido, inválido, divergente e tardio mantiveram a custódia/finalização segura localmente.

## 6. Rollout
- [ ] 6.1 Ensaiar upgrade/rollback e revisar o diff.
  - Objective: preservar dados, contratos e obrigações antigas.
  - Likely files/components: migrações, manifests e relatórios.
  - Depends on: qualificações anteriores.
  - Validation: ensaio compatível e checklist avaliador.
  - Completion criteria: riscos residuais explícitos; sem fechamento por inferência.
