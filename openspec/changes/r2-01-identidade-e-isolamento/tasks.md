# Tasks: Identidade, isolamento de tenants e administração segura

Status: **comportamentos, contratos e qualificação local concluídos; promoção regional/produtiva permanece aberta**. A fixture OIDC/MFA manual, os escopos administrativos, o isolamento e as rotas protegidas foram revalidados sem promover o laboratório a produção. Dependências de change: Sem pré-requisito de implementação para iniciar; integrações específicas descritas abaixo.

## 1. Contratos e preparação

- [x] 1.1 Revalidar snapshot, escopo e contratos compartilhados
  - Objective: Confrontar os achados desta change com HEAD, preservando evidências do SHA revisado; registrar deltas e fronteiras consumidor/produtor.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `docs/reviews/2026-09-07-r2`.
  - Depends on: nenhuma tarefa local; verificar dependências da change.
  - Validation: Inspeção do diff e contrato; review de responsáveis funcionais.
  - Completion criteria: Fatos atualizados, pré-condições e decisões pendentes identificados sem inventar aprovação.
  - Evidence: `openspec/changes/r2-01-identidade-e-isolamento/design.md`, `risk-matrix.md`, contratos OIDC/OpenAPI e `hub/evidence/r2/execution/identity-mfa-compatibility-latest.log` revisados no HEAD.

- [x] 1.2 Detalhar schemas e compatibilidade da fatia
  - Objective: Formalizar campos/estados/erros/permissões e exemplos sanitizados consumidos pelos requisitos abaixo antes da implementação.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `hub/api/openapi.yaml`; `hub/api/openapi-internal.yaml`; `hub/api/asyncapi.yaml`.
  - Depends on: 1.1.
  - Validation: Contract/schema review e casos inválidos; registrar quais contratos precisam nova versão.
  - Completion criteria: DTOs e versões acordados; nenhuma alteração incompatível implícita no perfil v1.
  - Evidence: `docs/06_CONFIGURACAO_E_SEGURANCA.md`, `hub/api/openapi.yaml`, `hub/api/openapi-internal.yaml` e `hub/api/asyncapi.yaml`; claims de sujeito, MFA, papel, escopo, tenant e erros protegidos permanecem explícitos.

- [x] 1.3 Preparar evolução aditiva e fixtures isoladas
  - Objective: Criar migrations adicionais quando aplicável, permissões, interfaces ou organização documental/UI necessária; separar fixtures de dados reais.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 1.2.
  - Validation: Aplicação em ambiente limpo e existente, rollback compatível, validação de unicidade/proveniência; não editar migration aplicada.
  - Completion criteria: Estrutura suporta as regras sem perda de histórico; para UI/qualificação, registrar explicitamente ausência de mudança de esquema quando confirmada.
  - Evidence: `hub/deploy/r2/scripts/reconcile_identity.py`, fixture OIDC/MFA local e migrations de RLS/auditoria; a reconciliação é idempotente e não usa credencial real.

## 2. Comportamentos

- [x] 2.1 Identidade autenticada e autorização por recurso
  - Objective: Entregar o comportamento R2-SEG-01 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline SEG-01, SEG-03, EXE-07, EXE-16.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `hub/internal/atlas/handlers.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-SEG-01; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.1; sem flags que apresentem mock como fluxo real.

- [x] 2.2 Workloads com menor privilégio
  - Objective: Entregar o comportamento R2-SEG-02 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline SEG-01, SEG-03, DAD-01, DAD-07.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `hub/internal/atlas/handlers.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-SEG-02; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.2; sem flags que apresentem mock como fluxo real.

- [x] 2.3 Leitura administrativa individual entre tenants
  - Objective: Entregar o comportamento R2-SEG-03 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline SEG-04, CFG-03.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `hub/internal/atlas/handlers.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-SEG-03; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.3; sem flags que apresentem mock como fluxo real.

- [x] 2.4 Destinos externos e callbacks protegidos
  - Objective: Entregar o comportamento R2-SEG-04 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline SEG-02, COM-02, CFG-05.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `hub/internal/atlas/handlers.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-SEG-04; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.4; sem flags que apresentem mock como fluxo real.

- [x] 2.5 Sessão administrativa e trilha de alterações
  - Objective: Entregar o comportamento R2-SEG-05 na autoridade correta, cobrindo os caminhos e falhas da spec; baseline CFG-01, CFG-02, SEG-01, SEG-03.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `hub/internal/atlas/handlers.go`.
  - Depends on: 1.3; contratos produtores listados no design disponíveis para integração.
  - Validation: Teste de domínio/contrato dos limites de R2-SEG-05; integração de transação/identidade onde relevante. Não usar apenas status HTTP como oráculo.
  - Completion criteria: Regra e erros observáveis implementados; prova completa será registrada na tarefa 3.5; sem flags que apresentem mock como fluxo real.

## 3. Provas por requisito

- [x] 3.1 Qualificar R2-SEG-01 com oráculos independentes
  - Objective: Executar R2-SEG-01-S01, R2-SEG-01-S02, R2-SEG-01-S03. Caso indispensável: Token inválido; esperado: o Hub responde 401 antes de qualquer efeito; token válido sem escopo recebe 403.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`.
  - Depends on: 2.1; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.2 Qualificar R2-SEG-02 com oráculos independentes
  - Objective: Executar R2-SEG-02-S01, R2-SEG-02-S02, R2-SEG-02-S03. Caso indispensável: Pool reutilizado; esperado: políticas e contexto transacional impedem acesso residual a A.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`.
  - Depends on: 2.2; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.3 Qualificar R2-SEG-03 com oráculos independentes
  - Objective: Executar R2-SEG-03-S01, R2-SEG-03-S02, R2-SEG-03-S03. Caso indispensável: Auditoria indisponível; esperado: essa leitura é negada de forma controlada sem interromper consultas públicas elegíveis.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`.
  - Depends on: 2.3; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.4 Qualificar R2-SEG-04 com oráculos independentes
  - Objective: Executar R2-SEG-04-S01, R2-SEG-04-S02, R2-SEG-04-S03. Caso indispensável: Rede privada legítima; esperado: a permissão é limitada ao destino/porta/identidade aprovados, sem exceção global a redes privadas.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`.
  - Depends on: 2.4; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

- [x] 3.5 Qualificar R2-SEG-05 com oráculos independentes
  - Objective: Executar R2-SEG-05-S01, R2-SEG-05-S02, R2-SEG-05-S03. Caso indispensável: Falha de backend; esperado: a resposta contém código tratável e correlação, sem DSN, token, stack ou segredo.
  - Likely files/components: `hub/test/e2e/e2e_test.go`; `hub/evidence`; `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`.
  - Depends on: 2.5; fixtures de r2-09; dependências de integração pertinentes.
  - Validation: Unitário/integrado/contrato/E2E conforme regra; falhas e concorrência reais quando normativas. Registrar ambiente, SHA, fixture, esperado/obtido e saída sanitizada.
  - Completion criteria: Todos os 3 cenários têm pass/fail/blocked explícito. Somente pass com evidência conclui esta tarefa; corrigir regressão em vez de reduzir assert.

## 4. Integração, migração e fechamento

- [x] 4.1 Ensaiar migração e recuo sem perda de obrigações
  - Objective: Criar rotas e identidades novas; migrar consumidores por perfil autenticado. O header legado pode continuar como metadado de correlação, mas nunca como autoridade. Isolar explicitamente o simulador local.
  - Likely files/components: `hub/migrations`; `docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md`.
  - Depends on: 2.x concluídas; plano de rollback do design.
  - Validation: Backfill em lotes, comparação de contagens/hashes e restart/rollback com estado em voo; adaptar ao escopo UI/documental sem inventar migração de dados.
  - Completion criteria: Histórico, identidade e obrigações preservados; procedimento e limitações registrados.
  - Evidence: `identity-mfa-compatibility-latest.log`, `browser-smoke.json`, `identity-global-reader.json`, `rls-runtime-proof.sh` e `compose-recreation-latest.log`; o fluxo manual OIDC/MFA, sessões, claims, escopos e isolamento sobreviveram à operação local. Rollback de IdP regional/produtivo permanece aberto.

- [x] 4.2 Qualificar a fatia integrada e observabilidade
  - Objective: Executar cenários IT pertinentes com consumidores/produtores reais de ensaio e jornadas administrativas afetadas; provar alerta/runbook de falha principal.
  - Likely files/components: `hub/deploy/kong/kong.yml`; `hub/internal/platform/httpserver/httpserver.go`; `hub/internal/orbita/handlers.go`; `hub/evidence`.
  - Depends on: 3.x e 4.1; changes produtoras/consumidoras necessárias integradas.
  - Validation: Contrato/E2E e observabilidade; gates G0–G7 conforme perfil; não exigir cloud para prova que cabe no laboratório.
  - Completion criteria: Evidência separa laboratório, homologação e perfil operacional; qualquer gate não executado continua aberto.
  - Evidence: `identity-mfa-compatibility-latest.log`, `browser-smoke.json`, `identity-global-reader.json` e `observability-alert-latest.log`; OIDC/PKCE, senha+OTP, leitura global auditada, escrita negada e labels/runbooks locais foram validados. IdP gerenciado, certificados e produção permanecem fora.

- [x] 4.3 Atualizar rastreabilidade e concluir apenas o comprovado
  - Objective: Atualizar matriz, auditoria, documentação e estado de tarefas conforme provas; preservar pendências e baseline.
  - Likely files/components: `IMPLEMENTATION_AUDIT.md`; `openspec/changes/r2-01-identidade-e-isolamento`; `docs/reviews/2026-09-07-r2`.
  - Depends on: 4.2.
  - Validation: OpenSpec strict, links/IDs e review de evidência.
  - Completion criteria: Nenhum requisito concluído por inferência; archive somente conforme plano de baseline e aceite real da change.
  - Evidence: `hub/evidence/r2/execution/EVIDENCE_INDEX.md` e `docs/reviews/2026-09-09-r4/implementation/CHECKPOINT.md`, atualizados com MFA manual, escopos e limites de promoção.
