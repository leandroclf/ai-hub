# Tasks: Custódia e integração externa

## 1. Preparação

- [ ] 1.1 Revalidar fontes e delta do HEAD.
  - Objective: confirmar achados no checkout atual sem sobrescrever trabalho.
  - Likely files/components: explore.md e fontes citadas.
  - Depends on: leitura do prompt R3 e instruções do repositório.
  - Validation: diff e atualização da matriz.
  - Completion criteria: cada achado tem evidência atual ou prova de já resolvido.

## 2. R3-EXE-01 — Adapters executáveis e autenticação completa

- [ ] 2.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-01 e o comportamento normativo.
  - Likely files/components: hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, hub/internal/atlasclient/client.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-EXE-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 2.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL executar adapters versionados homologados, transmitir configuração completa de autenticação e distinguir inventário importado de capacidade executável. Publicação deve rejeitar capacidades não suportadas; resultado sintético nunca substitui resposta real.
  - Likely files/components: hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, hub/internal/atlasclient/client.go, migrações/contratos/testes correlatos.
  - Depends on: 2.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 2.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-INT-01, R2-INT-02, R2-ADM-04.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 2.2.
  - Validation: cenários R3-EXE-01-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 3. R3-EXE-02 — Callback com confirmação de custódia

- [ ] 3.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-02 e o comportamento normativo.
  - Likely files/components: hub/internal/cometa/handlers.go, hub/internal/cometa/executor.go, hub/cmd/cometa/main.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-EXE-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 3.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL autenticar callback pela política da conta, limitar método/tamanho/replay e emitir 2xx somente após custódia durável. Recibos órfãos, duplicados e tardios autenticados devem ter disposição recuperável sem substituir final terminal.
  - Likely files/components: hub/internal/cometa/handlers.go, hub/internal/cometa/executor.go, hub/cmd/cometa/main.go, migrações/contratos/testes correlatos.
  - Depends on: 3.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 3.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-SEG-04, R2-EXE-04, R2-EXE-05, R2-INT-04.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 3.2.
  - Validation: cenários R3-EXE-02-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 4. R3-EXE-03 — Recuperação de efeito incerto e fencing

- [ ] 4.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-03 e o comportamento normativo.
  - Likely files/components: hub/internal/cometa/custody.go, hub/internal/cometa/executor.go, hub/internal/orbita/intents.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-EXE-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 4.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL manter obrigação recuperável para SUBMITTING/UNKNOWN, validar identidade completa/hash em todo replay e verificar lease/epoch/prazo antes de I/O. Reenvio de SUBMIT exige idempotência homologada ou prova positiva de ausência de efeito; UNKNOWN exige consulta ou reconciliação.
  - Likely files/components: hub/internal/cometa/custody.go, hub/internal/cometa/executor.go, hub/internal/orbita/intents.go, migrações/contratos/testes correlatos.
  - Depends on: 4.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 4.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-EXE-03, R2-EXE-09, R2-INT-05.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 4.2.
  - Validation: cenários R3-EXE-03-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 5. R3-EXE-04 — Relógios de retry, polling e prazo final

- [ ] 5.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-04 e o comportamento normativo.
  - Likely files/components: hub/internal/orbita/handlers.go, hub/internal/cometa/polling_custody.go, hub/internal/orbita/finalize.go, docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-EXE-04-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 5.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL separar SLA do cliente, SLA do provedor, TTL desde primeira falha transitória, timeout por tentativa e horizonte de reconciliação. Polling/callback devem coexistir; espera saudável não consome TTL de indisponibilidade. Não substituir confirmação durável no prazo por checagem SQL anterior ao commit sem mudança normativa autorizada.
  - Likely files/components: hub/internal/orbita/handlers.go, hub/internal/cometa/polling_custody.go, hub/internal/orbita/finalize.go, docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md, migrações/contratos/testes correlatos.
  - Depends on: 5.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 5.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-EXE-06, R2-INT-04, R2-INT-05, R2-INT-08.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 5.2.
  - Validation: cenários R3-EXE-04-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 6. R3-EXE-05 — Topologia de mensagens e quarentena recuperável

- [ ] 6.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-05 e o comportamento normativo.
  - Likely files/components: hub/internal/queue/bootstrap.go, hub/internal/queue/queue.go, hub/internal/libra/consumers.go, hub/internal/libra/store.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-EXE-05-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 6.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL verificar topologia durável obrigatória antes de liberar publicação, conservar cada obrigação até handoff comprovado e guardar bytes originais limitados ou referência imutável protegida em quarentena. Hash sozinho não é custódia recuperável; reconciliação cobre fan-out e retenção.
  - Likely files/components: hub/internal/queue/bootstrap.go, hub/internal/queue/queue.go, hub/internal/libra/consumers.go, hub/internal/libra/store.go, migrações/contratos/testes correlatos.
  - Depends on: 6.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 6.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-EXE-05, R2-QUA-02.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 6.2.
  - Validation: cenários R3-EXE-05-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 7. Entrega

- [ ] 7.1 Ensaiar migração e rollback e concluir checklist.
  - Objective: habilitação reversível sem perder obrigações históricas.
  - Likely files/components: design.md, migrations, manifests e evidências.
  - Depends on: todas as qualificações anteriores; gate integral r3-07.
  - Validation: canário, rollback compatível e reconciliação.
  - Completion criteria: não há achado encerrado sem prova; bloqueios externos são precisos e não ocultam trabalho executável.
