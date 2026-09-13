# Tasks: Capacidade, ambientes e promoção verificáveis
## 1. Estado inicial
- [ ] 1.1 Revalidar snapshot, fontes e ambiente.
  - Objective: preservar correções posteriores e trabalho local.
  - Likely files/components: AGENTS.md, explore.md, docs/reviews/2026-09-12-r5/.
  - Depends on: prompt e instruções do repositório.
  - Validation: git status/diff, inventário e classificação por achado.
  - Completion criteria: baseline/working tree e limitações identificadas.

## 2. R5-OPE-01
- [ ] 2.1 Reproduzir e fixar contrato de F-R5-11.
  - Objective: Capacidade sem bypass e recuperação de permits.
  - Likely files/components: hub/internal/cometa/executor.go, hub/internal/cometa/capacity.go, hub/internal/cometa/capacity.go, hub/internal/cometa/executor.go.
  - Depends on: 1.1.
  - Validation: R5-OPE-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 2.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL exigir política de capacidade qualificada para cada rota ativa e recuperar concessões pendentes com evidência durável de transporte/efeito, sem reciclar apenas por timeout. Crescimento de histórico e número de tenants não deve violar orçamento de concessão publicado; isolamento e limite agregado devem valer entre réplicas.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 2.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 2.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-OPE-01 e vínculos R3-INT-01, R3-INT-03, R2-INT-05.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 2.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 3. R5-OPE-02
- [ ] 3.1 Reproduzir e fixar contrato de F-R5-12.
  - Objective: Promoção exige evidência vinculada ao artefato.
  - Likely files/components: hub/deploy/r2/tests/promotion-gate.sh, hub/deploy/r2/tests/validate-qualification-evidence.py, hub/deploy/Dockerfile, hub/deploy/r2/Dockerfile.ui.
  - Depends on: 1.1.
  - Validation: R5-OPE-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 3.2 Implementar a fronteira completa.
  - Objective: O processo de promoção SHALL recusar artefato sem evidência íntegra e compatível com SHA/conteúdo/imagens efetivos e sem aprovações verificáveis do ambiente aplicável. Uma string PASS ou nome de aprovação não constitui prova. Mudança de toolchain/base exige requalificação material antes da promoção.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 3.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 3.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-OPE-02 e vínculos R3-QUA-01, R2-QUA-04, R2-OPE-08.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 3.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 4. R5-OPE-03
- [ ] 4.1 Reproduzir e fixar contrato de F-R5-13.
  - Objective: Ambientes elásticos com dados duráveis e isolamento completo.
  - Likely files/components: hub/deploy/r2/kind/render-independent-dependencies.py, hub/deploy/r2/kind/bootstrap-independent.sh, hub/deploy/r2/k8s/overlays/prd/kustomization.yaml.
  - Depends on: 1.1.
  - Validation: R5-OPE-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 4.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL disponibilizar perfis local/dev/hom/ppd/prd reproduzíveis com identidade, dados, segredos, rede, observabilidade e capacidade isolados. Escala automática deve abranger pods/nós/placement e budgets das dependências dentro de quotas explícitas; continuidade é aferida por requisições/obrigações reconciliadas, não somente readiness.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 4.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 4.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-OPE-03 e vínculos R3-OPE-02, R3-OPE-03, R3-OPE-05, R2-OPE-01.
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
