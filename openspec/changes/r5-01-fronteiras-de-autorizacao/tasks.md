# Tasks: Fronteiras de autorização e custódia
## 1. Estado inicial
- [ ] 1.1 Revalidar snapshot, fontes e ambiente.
  - Objective: preservar correções posteriores e trabalho local.
  - Likely files/components: AGENTS.md, explore.md, docs/reviews/2026-09-12-r5/.
  - Depends on: prompt e instruções do repositório.
  - Validation: git status/diff, inventário e classificação por achado.
  - Completion criteria: baseline/working tree e limitações identificadas.

## 2. R5-SEG-01
- [ ] 2.1 Reproduzir e fixar contrato de F-R5-01.
  - Objective: Autenticação antes de custódia de callback órfão.
  - Likely files/components: hub/internal/cometa/handlers.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go.
  - Depends on: 1.1.
  - Validation: R5-SEG-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 2.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL autenticar a origem antes de confirmar ou reservar custódia órfã. Cada recibo deve possuir escopo de origem autenticada, identidade estável de evento e disposição recuperável. Tentativas não autenticadas não podem consumir quota durável de obrigações válidas. Quotas, retenção e claims devem isolar contas/células; eventos irrelacionáveis devem chegar a disposição auditável sem bloqueio global.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 2.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 2.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-SEG-01 e vínculos R4-CBK-01, R4-CBK-02, R4-CBK-03.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 2.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 3. R5-SEG-02
- [ ] 3.1 Reproduzir e fixar contrato de F-R5-02.
  - Objective: Fallback de oferta respeita negação e revogação.
  - Likely files/components: hub/internal/atlasclient/client.go, hub/internal/atlas/offers.go.
  - Depends on: 1.1.
  - Validation: R5-SEG-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 3.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL distinguir indisponibilidade transitória de negação autoritativa ao resolver ofertas. Negação, revogação, conflito ou resposta inválida não podem ser convertidos em autorização por cache. Projeções válidas devem atender o caminho quente dentro de janela de autorização explicitamente publicada, com invalidação e limites de armazenamento mensuráveis.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 3.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 3.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-SEG-02 e vínculos R4-OPE-03, R3-CAT-04, R2-SEG-01.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 3.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 4. R5-SEG-03
- [ ] 4.1 Reproduzir e fixar contrato de F-R5-03.
  - Objective: Adoção real de RLS e autorização de dados auxiliares.
  - Likely files/components: hub/internal/platform/pg/pg.go, hub/deploy/r2/tests/rls-runtime-proof.sh, hub/migrations/core/0034_runtime_rls_migration_compat.sql, hub/internal/cometa/handlers.go.
  - Depends on: 1.1.
  - Validation: R5-SEG-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 4.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL aplicar identidade autenticada por transação/work item com roles runtime sem propriedade nem bypass, incluindo tabelas auxiliares e caminhos administrativos. Escopo global exige autorização nominal específica e auditoria. Provas de isolamento devem demonstrar dados próprios existentes, dados alheios existentes e invisíveis, escrita cruzada recusada e ausência de vazamento ao reutilizar conexões.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 4.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 4.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-SEG-03 e vínculos R3-OPE-04, R3-ADM-01, R2-DAD-05.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 4.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 5. Revisão e rollout
- [ ] 5.1 Revisar compatibilidade e rollback.
  - Objective: preparar entrega revisável sem implantação remota.
  - Likely files/components: design, migrações, manifests e relatório final.
  - Depends on: qualificações acima.
  - Validation: checklist avaliador e diff final.
  - Completion criteria: documentação e evidência consistentes; sem aprovação presumida.
