# Tasks: Financeiro e entregas

## 1. Preparação

- [ ] 1.1 Revalidar fontes e delta do HEAD.
  - Objective: confirmar achados no checkout atual sem sobrescrever trabalho.
  - Likely files/components: explore.md e fontes citadas.
  - Depends on: leitura do prompt R3 e instruções do repositório.
  - Validation: diff e atualização da matriz.
  - Completion criteria: cada achado tem evidência atual ou prova de já resolvido.

## 2. R3-FIN-01 — Reserva estrita e franquia antes do efeito

- [ ] 2.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-17 e o comportamento normativo.
  - Likely files/components: hub/internal/libra/store.go, hub/internal/orbita/handlers.go, hub/internal/contracts/economics/publication.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-FIN-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 2.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL autorizar saldo/franquia antes do efeito, reservar exposição máxima ou obter autorização incremental prévia e conciliar captura/liberação com valor efetivo exato. UNKNOWN mantém retenção até evidência positiva; expiração do cliente não prova ausência de custo.
  - Likely files/components: hub/internal/libra/store.go, hub/internal/orbita/handlers.go, hub/internal/contracts/economics/publication.go, migrações/contratos/testes correlatos.
  - Depends on: 2.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 2.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-FIN-03, R2-FIN-04.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 2.2.
  - Validation: cenários R3-FIN-01-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 3. R3-FIN-02 — Incidência completa e fechamento operacional

- [ ] 3.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-18 e o comportamento normativo.
  - Likely files/components: hub/internal/cometa/custody.go, hub/internal/cometa/polling_custody.go, hub/internal/libra/store.go, hub/internal/libra/handlers.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-FIN-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 3.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL medir unidades de compra/venda do snapshot, incluindo SUBMIT/STATUS/FETCH contratados, com attempt_id/evidence_id e deduplicação por unidade. Watermarks derivam de completude demonstrável; disputa, conciliação e exportação devem funcionar sem editar banco manualmente.
  - Likely files/components: hub/internal/cometa/custody.go, hub/internal/cometa/polling_custody.go, hub/internal/libra/store.go, hub/internal/libra/handlers.go, migrações/contratos/testes correlatos.
  - Depends on: 3.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 3.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-FIN-01, R2-FIN-02, R2-FIN-06.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 3.2.
  - Validation: cenários R3-FIN-02-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 4. R3-FIN-03 — Destino de webhook congelado por aplicação

- [ ] 4.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-19 e o comportamento normativo.
  - Likely files/components: hub/internal/pulsar/custody.go, hub/internal/orbita/finalize.go, hub/internal/pulsar/handlers.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-FIN-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 4.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL congelar destinos autorizados e contrato de entrega, conservar representação final imutável e distinguir reentrega de execução. Ausência de destino tem disposição explícita. Tentativas, leases e comando manual respeitam budgets auditáveis.
  - Likely files/components: hub/internal/pulsar/custody.go, hub/internal/orbita/finalize.go, hub/internal/pulsar/handlers.go, migrações/contratos/testes correlatos.
  - Depends on: 4.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 4.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-INT-07, R2-EXE-08, R2-ADM-09.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 4.2.
  - Validation: cenários R3-FIN-03-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 5. Entrega

- [ ] 5.1 Ensaiar migração e rollback e concluir checklist.
  - Objective: habilitação reversível sem perder obrigações históricas.
  - Likely files/components: design.md, migrations, manifests e evidências.
  - Depends on: todas as qualificações anteriores; gate integral r3-07.
  - Validation: canário, rollback compatível e reconciliação.
  - Completion criteria: não há achado encerrado sem prova; bloqueios externos são precisos e não ocultam trabalho executável.
