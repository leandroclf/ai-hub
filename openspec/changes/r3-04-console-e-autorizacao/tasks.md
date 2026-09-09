# Tasks: Console e autorização

## 1. Preparação

- [ ] 1.1 Revalidar fontes e delta do HEAD.
  - Objective: confirmar achados no checkout atual sem sobrescrever trabalho.
  - Likely files/components: explore.md e fontes citadas.
  - Depends on: leitura do prompt R3 e instruções do repositório.
  - Validation: diff e atualização da matriz.
  - Completion criteria: cada achado tem evidência atual ou prova de já resolvido.

## 2. R3-ADM-01 — Fronteira administrativa e aplicação

- [ ] 2.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-13 e o comportamento normativo.
  - Likely files/components: hub/internal/orbita/admin.go, hub/internal/platform/auth/auth.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-ADM-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 2.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL distinguir consumidor de operador e aplicar escopo por tenant/aplicação em listagens, detalhe e payload. Leitura entre tenants exige identidade nominal autorizada, MFA, auditoria e mascaramento; desenvolvedor não usa conta compartilhada nem herda escrita financeira.
  - Likely files/components: hub/internal/orbita/admin.go, hub/internal/platform/auth/auth.go, migrações/contratos/testes correlatos.
  - Depends on: 2.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 2.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-SEG-01, R2-SEG-03, R2-ADM-01.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 2.2.
  - Validation: cenários R3-ADM-01-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 3. R3-ADM-02 — Ações operacionais com API efetiva

- [ ] 3.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-14 e o comportamento normativo.
  - Likely files/components: hub/admin-ui/src/pages/OperationsPage.tsx, hub/internal/pulsar/handlers.go, hub/internal/orbita/admin.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-ADM-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 3.2 Implementar e conectar o fluxo completo.
  - Objective: O console SHALL usar contratos validados por operação, delivery_id para entregas e endpoints efetivos para reconciliação e SLA. Ação não implementada não pode simular disponibilidade. Redelivery repete bytes da entrega e não executa novamente provedor.
  - Likely files/components: hub/admin-ui/src/pages/OperationsPage.tsx, hub/internal/pulsar/handlers.go, hub/internal/orbita/admin.go, migrações/contratos/testes correlatos.
  - Depends on: 3.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 3.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-ADM-08, R2-ADM-09, R2-ADM-10.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 3.2.
  - Validation: cenários R3-ADM-02-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 4. R3-ADM-03 — Comandos financeiros compatíveis

- [ ] 4.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-15 e o comportamento normativo.
  - Likely files/components: hub/admin-ui/src/pages/FinancePage.tsx, hub/internal/libra/handlers.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-ADM-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 4.2 Implementar e conectar o fluxo completo.
  - Objective: O console SHALL emitir contrato financeiro válido, propagar tenant autorizado, converter períodos com fuso explícito e preservar chave da intenção em retries. Autor deriva da identidade; aprovação, disputa e exportação devem completar jornadas com segregação.
  - Likely files/components: hub/admin-ui/src/pages/FinancePage.tsx, hub/internal/libra/handlers.go, migrações/contratos/testes correlatos.
  - Depends on: 4.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 4.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-ADM-11, R2-FIN-05, R2-FIN-06.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 4.2.
  - Validation: cenários R3-ADM-03-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 5. R3-ADM-04 — Formulários completos e tempo local

- [ ] 5.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-16 e o comportamento normativo.
  - Likely files/components: hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-ADM-04-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 5.2 Implementar e conectar o fluxo completo.
  - Objective: O console SHALL preservar multimodalidade, oferecer campos condicionais de autenticação, lookup pesquisável/paginado e conversão correta de fuso. Respostas devem ser validadas em runtime e jornadas acessíveis por teclado com erros vinculados aos campos.
  - Likely files/components: hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src, migrações/contratos/testes correlatos.
  - Depends on: 5.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 5.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-ADM-02, R2-ADM-03, R2-ADM-06, R2-ADM-07, R2-ADM-12.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 5.2.
  - Validation: cenários R3-ADM-04-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 6. Entrega

- [ ] 6.1 Ensaiar migração e rollback e concluir checklist.
  - Objective: habilitação reversível sem perder obrigações históricas.
  - Likely files/components: design.md, migrations, manifests e evidências.
  - Depends on: todas as qualificações anteriores; gate integral r3-07.
  - Validation: canário, rollback compatível e reconciliação.
  - Completion criteria: não há achado encerrado sem prova; bloqueios externos são precisos e não ocultam trabalho executável.
