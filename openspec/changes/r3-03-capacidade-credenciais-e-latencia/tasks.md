# Tasks: Capacidade, credenciais e latência

## 1. Preparação

- [ ] 1.1 Revalidar fontes e delta do HEAD.
  - Objective: confirmar achados no checkout atual sem sobrescrever trabalho.
  - Likely files/components: explore.md e fontes citadas.
  - Depends on: leitura do prompt R3 e instruções do repositório.
  - Validation: diff e atualização da matriz.
  - Completion criteria: cada achado tem evidência atual ou prova de já resolvido.

## 2. R3-INT-01 — Controle adaptativo conectado a todo I/O

- [ ] 2.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-10 e o comportamento normativo.
  - Likely files/components: hub/internal/cometa/capacity.go, hub/internal/cometa/executor.go, hub/internal/cometa/poller.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-INT-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 2.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL adquirir concessão global por domínio antes de SUBMIT/STATUS/FETCH, devolver métricas de resultado e adaptar concorrência com redução por erros, recuperação amortecida e Retry-After. Limites contratuais duros coexistem com adaptação e justiça entre tenants.
  - Likely files/components: hub/internal/cometa/capacity.go, hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, migrações/contratos/testes correlatos.
  - Depends on: 2.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 2.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-INT-06, R2-OPE-09.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 2.2.
  - Validation: cenários R3-INT-01-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 3. R3-INT-02 — Cache de autenticação isolado e sem tokens no Redis

- [ ] 3.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-11 e o comportamento normativo.
  - Likely files/components: hub/internal/providerauth/client.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-INT-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 3.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL manter segredos em cofre/cache de memória limitado com expiração/revogação, eliminar tokens do Redis e coordenar renovação por binding sem bloquear contas independentes. L1 válido deve dispensar chamada remota; falha não permite token expirado ou fallback de outro cliente.
  - Likely files/components: hub/internal/providerauth/client.go, migrações/contratos/testes correlatos.
  - Depends on: 3.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 3.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-INT-02, R2-INT-03, R2-OPE-06.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 3.2.
  - Validation: cenários R3-INT-02-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 4. R3-INT-03 — Pools HTTP e budgets de concorrência

- [ ] 4.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-12 e o comportamento normativo.
  - Likely files/components: hub/internal/platform/egress, hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, hub/internal/providerauth/client.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-INT-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 4.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL reutilizar pools limitados por origem/identidade TLS, manter timeout por operação e drenar pools obsoletos. Controlador global limita concorrência; certificados/bindings incompatíveis não compartilham estado de autenticação.
  - Likely files/components: hub/internal/platform/egress, hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, hub/internal/providerauth/client.go, migrações/contratos/testes correlatos.
  - Depends on: 4.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 4.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-INT-06, R2-OPE-09.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 4.2.
  - Validation: cenários R3-INT-03-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 5. Entrega

- [ ] 5.1 Ensaiar migração e rollback e concluir checklist.
  - Objective: habilitação reversível sem perder obrigações históricas.
  - Likely files/components: design.md, migrations, manifests e evidências.
  - Depends on: todas as qualificações anteriores; gate integral r3-07.
  - Validation: canário, rollback compatível e reconciliação.
  - Completion criteria: não há achado encerrado sem prova; bloqueios externos são precisos e não ocultam trabalho executável.
