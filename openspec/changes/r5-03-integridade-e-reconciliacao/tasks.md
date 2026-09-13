# Tasks: Integridade de dados e reconciliação
## 1. Estado inicial
- [ ] 1.1 Revalidar snapshot, fontes e ambiente.
  - Objective: preservar correções posteriores e trabalho local.
  - Likely files/components: AGENTS.md, explore.md, docs/reviews/2026-09-12-r5/.
  - Depends on: prompt e instruções do repositório.
  - Validation: git status/diff, inventário e classificação por achado.
  - Completion criteria: baseline/working tree e limitações identificadas.

## 2. R5-DAD-01
- [ ] 2.1 Reproduzir e fixar contrato de F-R5-08.
  - Objective: Precisão do resultado preservada em todas as fronteiras.
  - Likely files/components: hub/internal/orbita/factconsumer.go, hub/internal/orbita/factconsumer.go, hub/internal/orbita/product_store.go, hub/admin-ui/src/pages/FinancePage.tsx.
  - Depends on: 1.1.
  - Validation: R5-DAD-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 2.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL preservar valor e tipo de números/identificadores em toda cadeia provedor→fato→estado→produto→representação→GET/webhook/console, sem coerção imprecisa. IDs inteiros fora da faixa segura do consumidor devem ter contrato textual explícito; resultados inválidos não são persistidos como sucesso.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 2.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 2.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-DAD-01 e vínculos R4-CTR-01, R4-CTR-02, R3-CAT-01.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 2.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 3. R5-DAD-02
- [ ] 3.1 Reproduzir e fixar contrato de F-R5-09.
  - Objective: Restore populado e reconciliação antes de retomada.
  - Likely files/components: hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh.
  - Depends on: 1.1.
  - Validation: R5-DAD-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 3.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL demonstrar restore não vazio e retomada cercada de todas as autoridades e obrigações, incluindo planos, etapas, compensações, recibos, versões de objetos e financeiro. Antes de liberar tráfego deve reconciliar identidades/bytes/efeitos/valores; ausência de fixture ou divergência impede aprovação.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 3.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 3.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-DAD-02 e vínculos R3-OPE-04, R3-OPE-01, R2-OPE-08.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 3.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 4. R5-DAD-03
- [ ] 4.1 Reproduzir e fixar contrato de F-R5-10.
  - Objective: Liquidação por valor efetivo e completude demonstrável.
  - Likely files/components: hub/internal/libra/store.go, hub/internal/libra/store.go, hub/internal/libra/settlement.go, hub/internal/libra/store.go.
  - Depends on: 1.1.
  - Validation: R5-DAD-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 4.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL liquidar saldo pelo valor contratado efetivo, preservando reserva/hold de incerteza e liberando excedente de forma auditável. Fechamento exige evidência durável de completude dos produtores, sem watermarks fabricados. Fatos tardios devem gerar disposição financeira consultável e ajustável sem alterar exportação fechada.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 4.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 4.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-DAD-03 e vínculos R3-FIN-01, R3-FIN-02, R2-FIN-03, R2-FIN-05.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 4.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 5. R5-DAD-04
- [ ] 5.1 Reproduzir e fixar contrato de F-R5-16.
  - Objective: Quarentena financeira conserva mensagem recuperável.
  - Likely files/components: hub/internal/libra/store.go, hub/internal/libra/consumers.go.
  - Depends on: 1.1.
  - Validation: R5-DAD-04-S01/S02/S03 com oráculo independente.
  - Completion criteria: comportamento atual demonstrado e solução compatível desenhada.
- [ ] 5.2 Implementar a fronteira completa.
  - Objective: O Hub SHALL conservar conteúdo recuperável e proveniência de mensagens financeiras não aplicáveis antes de confirmar sua remoção do transporte. Reprocessamento exige autorização, idempotência e trilha de disposição; hash isolado não constitui custódia do conteúdo.
  - Likely files/components: fontes acima e migrações/contratos/scripts relacionados.
  - Depends on: 5.1 e ordem do plano de integração.
  - Validation: integração real, concorrência/erro e recuperação.
  - Completion criteria: código conectado ao fluxo, não só helper; sem perda das obrigações antigas.
- [ ] 5.3 Qualificar e fechar por evidência.
  - Objective: comprovar R5-DAD-04 e vínculos R3-EXE-05, R2-DAD-01, R2-FIN-01.
  - Likely files/components: hub/evidence/r5 e docs/reviews/2026-09-12-r5/implementation.
  - Depends on: 5.2.
  - Validation: cenários, código/digests, logs saneados, upgrade/rollback pertinente.
  - Completion criteria: evidência atual com resultado, oráculo e zero skip obrigatório afetado.

## 6. Revisão e rollout
- [ ] 6.1 Revisar compatibilidade e rollback.
  - Objective: preparar entrega revisável sem implantação remota.
  - Likely files/components: design, migrações, manifests e relatório final.
  - Depends on: qualificações acima.
  - Validation: checklist avaliador e diff final.
  - Completion criteria: documentação e evidência consistentes; sem aprovação presumida.
