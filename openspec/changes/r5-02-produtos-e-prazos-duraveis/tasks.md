# Tasks: Produtos, prazos e compensação duráveis
## 1. Estado inicial
- [ ] 1.1 Revalidar snapshot, fontes e ambiente.
  - Objective: preservar correções posteriores e trabalho local.
  - Likely files/components: AGENTS.md, explore.md, docs/reviews/2026-09-12-r5/.
  - Depends on: prompt e instruções do repositório.
  - Validation: git status/diff, inventário e classificação por achado.
  - Completion criteria: baseline/working tree e limitações identificadas.

## 2. R5-EXE-01
- [ ] 2.1 Reproduzir e fixar contrato de F-R5-04.
  - Objective: Horizonte de retry e SLA verificados antes do despacho.
  - Likely files/components: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/cometa/polling_custody.go, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depends on: 1.1.
  - Validation: R5-EXE-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 2.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL impedir novos despachos com possibilidade de efeito após o horizonte aplicável e conservar separadamente prazo do cliente, provedor, tentativa e retry de indisponibilidade desde primeira falha. Pendência assíncrona legítima não inicia TTL de falha. Takeover não renova prazos; observação tardia conserva evidência sem alterar final fechado.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 2.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 2.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-EXE-01 e vínculos R3-EXE-04, R2-EXE-05, R2-EXE-06.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 2.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 3. R5-EXE-02
- [ ] 3.1 Reproduzir e fixar contrato de F-R5-05.
  - Objective: Limite de execução de produto atômico e retomável.
  - Likely files/components: hub/internal/orbita/intents.go, hub/internal/orbita/product_store.go, hub/internal/atlas/executor.go.
  - Depends on: 1.1.
  - Validation: R5-EXE-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 3.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL reservar vaga de execução e posse da etapa atomicamente por produto, respeitando max_parallel entre réplicas. Uma etapa com posse expirada deve ser recuperável sem disputar uma segunda vaga nem repetir efeito confirmado. Cancelamento impede novo efeito e estado terminal não é ressuscitado.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 3.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 3.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-EXE-02 e vínculos R3-CAT-03, R2-EXE-01, R2-CAT-02.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 3.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 4. R5-EXE-03
- [ ] 4.1 Reproduzir e fixar contrato de F-R5-06.
  - Objective: Contrato, provedor e incidência próprios por etapa.
  - Likely files/components: hub/internal/orbita/product_plan.go, hub/internal/orbita/product_plan.go, hub/internal/atlas/offers.go.
  - Depends on: 1.1.
  - Validation: R5-EXE-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 4.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL congelar por etapa serviço/versão, rota homologada, conta, binding, contrato de compra, perfil técnico, prazo e identidade econômica coerentes. A venda do produto e os custos das etapas devem manter escopos próprios sem duplicação. Consolidação deve seguir política publicada e preservar proveniência do plano.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 4.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 4.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-EXE-03 e vínculos R3-CAT-02, R3-CAT-03, R3-FIN-02.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 4.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 5. R5-EXE-04
- [ ] 5.1 Reproduzir e fixar contrato de F-R5-07.
  - Objective: Compensação tem ordem causal e prazo próprio.
  - Likely files/components: hub/internal/orbita/product_store.go, hub/internal/orbita/product_plan.go, hub/internal/orbita/intents.go.
  - Depends on: 1.1.
  - Validation: R5-EXE-04-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 5.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL conservar e executar compensações em ordem causal reversa com prazo, retry e autorização próprios, independentemente do encerramento da resposta ao cliente. Falha ou incerteza de compensação deve permanecer reconciliável sem ser convertida em sucesso nem reabrir protocolo.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 5.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 5.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-EXE-04 e vínculos R3-CAT-03, R2-EXE-01.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 5.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 6. R5-EXE-05
- [ ] 6.1 Reproduzir e fixar contrato de F-R5-17.
  - Objective: Topologia de mensagens pronta antes de publicar obrigações.
  - Likely files/components: hub/cmd/orbita/main.go, hub/cmd/cometa/main.go, hub/cmd/libra/main.go, hub/internal/queue/bootstrap.go.
  - Depends on: 1.1.
  - Validation: R5-EXE-05-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 6.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL confirmar a topologia e políticas de todas as assinaturas obrigatórias antes de liberar publicação de fatos duráveis. Indisponibilidade da topologia deve manter obrigações no outbox sem bloquear rotas independentes. Alteração ou recriação de recurso exige revalidação antes de descartar custódia local.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 6.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 6.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-EXE-05 e vínculos R3-EXE-05, R2-EXE-09.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 6.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 7. Revisão e rollout
- [ ] 7.1 Revisar compatibilidade e rollback.
  - Objective: preparar entrega revisável sem implantação remota.
  - Likely files/components: design, migrações, manifests e relatório final.
  - Depends on: qualificações acima.
  - Validation: checklist avaliador e diff final.
  - Completion criteria: documentação e evidência consistentes; sem aprovação presumida.
