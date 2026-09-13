# Tasks: Console escalável e contratual
## 1. Estado inicial
- [ ] 1.1 Revalidar snapshot, fontes e ambiente.
  - Objective: preservar correções posteriores e trabalho local.
  - Likely files/components: AGENTS.md, explore.md, docs/reviews/2026-09-12-r5/.
  - Depends on: prompt e instruções do repositório.
  - Validation: git status/diff, inventário e classificação por achado.
  - Completion criteria: baseline/working tree e limitações identificadas.

## 2. R5-UX-01
- [ ] 2.1 Reproduzir e fixar contrato de F-R5-14.
  - Objective: Console preserva intenção e escala com catálogo.
  - Likely files/components: hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/pages/CapacityDomainsPage.tsx.
  - Depends on: 1.1.
  - Validation: R5-UX-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 2.2 Implementar a fronteira completa.
  - Objective: O console SHALL permitir selecionar referências por busca/paginação de servidor sem teto funcional de 1000, validar contratos de resposta por jornada e preservar identidade de intenção após resposta incerta. Configuração operacional suportada deve ter fluxo autenticado, versionado e auditável; o usuário deve distinguir pendente, confirmado e recusado.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 2.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 2.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-UX-01 e vínculos R3-ADM-02, R3-ADM-03, R3-ADM-04, R4-QUA-02.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 2.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 3. Revisão e rollout
- [ ] 3.1 Revisar compatibilidade e rollback.
  - Objective: preparar entrega revisável sem implantação remota.
  - Likely files/components: design, migrações, manifests e relatório final.
  - Depends on: qualificações acima.
  - Validation: checklist avaliador e diff final.
  - Completion criteria: documentação e evidência consistentes; sem aprovação presumida.
