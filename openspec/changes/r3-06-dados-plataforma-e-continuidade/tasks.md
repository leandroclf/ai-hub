# Tasks: Dados, plataforma e continuidade

## 1. Preparação

- [ ] 1.1 Revalidar fontes e delta do HEAD.
  - Objective: confirmar achados no checkout atual sem sobrescrever trabalho.
  - Likely files/components: explore.md e fontes citadas.
  - Depends on: leitura do prompt R3 e instruções do repositório.
  - Validation: diff e atualização da matriz.
  - Completion criteria: cada achado tem evidência atual ou prova de já resolvido.

## 2. R3-OPE-01 — Objetos conectados ao fluxo e retenção

- [ ] 2.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-20 e o comportamento normativo.
  - Likely files/components: hub/internal/objectstore/catalog.go, hub/internal/objectstore/retention.go, hub/internal/cometa/executor.go, hub/internal/orbita/admission.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-OPE-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 2.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL consumir referências imutáveis por streaming limitado, conservar resultado volumoso antes da publicação e integrar retenção/pins a protocolos, entregas e finanças. Classe/região restringem processamento e acesso.
  - Likely files/components: hub/internal/objectstore/catalog.go, hub/internal/objectstore/retention.go, hub/internal/cometa/executor.go, hub/internal/orbita/admission.go, migrações/contratos/testes correlatos.
  - Depends on: 2.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 2.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-DAD-01, R2-DAD-02, R2-DAD-03.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 2.2.
  - Validation: cenários R3-OPE-01-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 3. R3-OPE-02 — Kind completo e ambientes reprodutíveis

- [ ] 3.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-21 e o comportamento normativo.
  - Likely files/components: hub/deploy/r2/kind/render-runtime.py, hub/deploy/r2/k8s/base/workloads.yaml, hub/deploy/r2/k8s/overlays/prd/kustomization.yaml.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-OPE-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 3.2 Implementar e conectar o fluxo completo.
  - Objective: A entrega SHALL oferecer Compose local completo e kind completo com dependências no cluster, UI/gateway/identidade/dados/mensageria/cofre emulável/observabilidade. Serviços gerenciados em dev/hom/ppd/prd exigem IaC e referências explícitas; imagens e endpoints devem ser reproduzíveis sem IP efêmero.
  - Likely files/components: hub/deploy/r2/kind/render-runtime.py, hub/deploy/r2/k8s/base/workloads.yaml, hub/deploy/r2/k8s/overlays/prd/kustomization.yaml, migrações/contratos/testes correlatos.
  - Depends on: 3.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 3.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-OPE-01, R2-OPE-02, R2-OPE-08.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 3.2.
  - Validation: cenários R3-OPE-02-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 4. R3-OPE-03 — Escala e continuidade com envelope

- [ ] 4.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-22 e o comportamento normativo.
  - Likely files/components: hub/deploy/r2/k8s/base/workloads.yaml, hub/internal/platform/httpserver.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-OPE-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 4.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL automatizar onboarding/placement e escala de pods/nós/dados dentro de envelope aprovado, separar readiness por capacidade e drenar workers/leases. Saturação além do envelope requer admissão controlada; escala infinita e failover instantâneo não são promessas válidas.
  - Likely files/components: hub/deploy/r2/k8s/base/workloads.yaml, hub/internal/platform/httpserver, migrações/contratos/testes correlatos.
  - Depends on: 4.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 4.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-OPE-03, R2-OPE-04, R2-OPE-05, R2-OPE-09.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 4.2.
  - Validation: cenários R3-OPE-03-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 5. R3-OPE-04 — Autoridade durável, isolamento e restore

- [ ] 5.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-23 e o comportamento normativo.
  - Likely files/components: hub/migrations/core, hub/migrations/control, hub/internal/platform/pg, hub/internal/orbita/admission.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-OPE-04-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 5.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL preservar autoridade durável replicada, papéis mínimos e isolamento tenant/aplicação/célula, e qualificar restore reconciliando mensagens/efeitos/objetos/finanças. Redis é dispensável. Falha total da autoridade exige recusa segura; fallback durável alternativo exige consistência/fencing comprovados.
  - Likely files/components: hub/migrations/core, hub/migrations/control, hub/internal/platform/pg, hub/internal/orbita/admission.go, migrações/contratos/testes correlatos.
  - Depends on: 5.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 5.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-DAD-04, R2-DAD-05, R2-OPE-06.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 5.2.
  - Validation: cenários R3-OPE-04-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 6. R3-OPE-05 — SLA bilateral e telemetria verificável

- [ ] 6.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-24 e o comportamento normativo.
  - Likely files/components: hub/internal/platform/httpserver/telemetry.go, hub/admin-ui/src/pages/OperationsPage.tsx, hub/deploy/r2.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-OPE-05-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 6.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL emitir métricas reais de idade/custódia/lag/capacidade e SLA cliente-Hub/Hub-provedor, permitir consulta autorizada de violações e correlacionar logs/traces/protocolos sem segredos. Prometheus/Loki/Grafana devem ser testados com falhas injetadas e cardinalidade limitada.
  - Likely files/components: hub/internal/platform/httpserver/telemetry.go, hub/admin-ui/src/pages/OperationsPage.tsx, hub/deploy/r2, migrações/contratos/testes correlatos.
  - Depends on: 6.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 6.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-OPE-07, R2-INT-08, R2-ADM-10.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 6.2.
  - Validation: cenários R3-OPE-05-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 7. Entrega

- [ ] 7.1 Ensaiar migração e rollback e concluir checklist.
  - Objective: habilitação reversível sem perder obrigações históricas.
  - Likely files/components: design.md, migrations, manifests e evidências.
  - Depends on: todas as qualificações anteriores; gate integral r3-07.
  - Validation: canário, rollback compatível e reconciliação.
  - Completion criteria: não há achado encerrado sem prova; bloqueios externos são precisos e não ocultam trabalho executável.
