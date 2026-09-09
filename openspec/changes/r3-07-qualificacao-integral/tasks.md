# Tasks: Qualificação integral

## 1. Preparação

- [ ] 1.1 Revalidar fontes e delta do HEAD.
  - Objective: confirmar achados no checkout atual sem sobrescrever trabalho.
  - Likely files/components: explore.md e fontes citadas.
  - Depends on: leitura do prompt R3 e instruções do repositório.
  - Validation: diff e atualização da matriz.
  - Completion criteria: cada achado tem evidência atual ou prova de já resolvido.

## 2. R3-QUA-01 — Qualificação integral do SHA sem skips ocultos

- [ ] 2.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-25 e o comportamento normativo.
  - Likely files/components: hub/internal/atlas/catalog_test.go, hub/internal/cometa/custody_test.go, docs/reviews/2026-09-07-r2/implementation/FINAL_REPORT.md.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-QUA-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 2.2 Implementar e conectar o fluxo completo.
  - Objective: A engenharia SHALL qualificar o SHA entregue com integração obrigatória sem skips, fixture OIDC reproduzível, navegador real e oráculos externos, cobrindo toda baseline v4/R2/R3. Evidência histórica não qualifica outro SHA; tarefas só fecham com provas correspondentes e bloqueios materiais permanecem explícitos.
  - Likely files/components: hub/internal/atlas/catalog_test.go, hub/internal/cometa/custody_test.go, docs/reviews/2026-09-07-r2/implementation/FINAL_REPORT.md, migrações/contratos/testes correlatos.
  - Depends on: 2.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 2.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-QUA-01, R2-QUA-02, R2-QUA-03, R2-QUA-04.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 2.2.
  - Validation: cenários R3-QUA-01-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 3. Entrega

- [ ] 3.1 Ensaiar migração e rollback e concluir checklist.
  - Objective: habilitação reversível sem perder obrigações históricas.
  - Likely files/components: design.md, migrations, manifests e evidências.
  - Depends on: todas as qualificações anteriores; gate integral r3-07.
  - Validation: canário, rollback compatível e reconciliação.
  - Completion criteria: não há achado encerrado sem prova; bloqueios externos são precisos e não ocultam trabalho executável.
