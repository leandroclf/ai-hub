# Tasks: Portfólio e contratos efetivos

## 1. Preparação

- [ ] 1.1 Revalidar fontes e delta do HEAD.
  - Objective: confirmar achados no checkout atual sem sobrescrever trabalho.
  - Likely files/components: explore.md e fontes citadas.
  - Depends on: leitura do prompt R3 e instruções do repositório.
  - Validation: diff e atualização da matriz.
  - Completion criteria: cada achado tem evidência atual ou prova de já resolvido.

## 2. R3-CAT-01 — Precisão numérica e validação de schemas

- [ ] 2.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-06 e o comportamento normativo.
  - Likely files/components: hub/internal/atlas/offers.go, hub/internal/orbita/handlers.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-CAT-01-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 2.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL preservar valores numéricos exatos, validar integralmente o dialeto declarado e rejeitar publicação de construções de schema não suportadas. Transformação deve ter limites de profundidade/tamanho/custo e não executar código ou acessar rede/segredos.
  - Likely files/components: hub/internal/atlas/offers.go, hub/internal/orbita/handlers.go, migrações/contratos/testes correlatos.
  - Depends on: 2.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 2.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-CAT-05, R2-EXE-04.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 2.2.
  - Validation: cenários R3-CAT-01-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 3. R3-CAT-02 — Política efetiva e representação por cliente

- [ ] 3.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-07 e o comportamento normativo.
  - Likely files/components: hub/internal/orbita/handlers.go, hub/internal/orbita/finalize.go, hub/internal/atlas/offers.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-CAT-02-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 3.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL definir precedência serviço→oferta→perfil dentro de limites contratuais, validar versão solicitada e congelar política efetiva/hash por aplicação. Representação final transformada deve ser persistida uma vez e reutilizada com bytes idênticos em GET e corpo de webhook.
  - Likely files/components: hub/internal/orbita/handlers.go, hub/internal/orbita/finalize.go, hub/internal/atlas/offers.go, migrações/contratos/testes correlatos.
  - Depends on: 3.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 3.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-CAT-02, R2-CAT-05, R2-CAT-06, R2-EXE-08.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 3.2.
  - Validation: cenários R3-CAT-02-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 4. R3-CAT-03 — Agregação e composição com executor de DAG

- [ ] 4.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-08 e o comportamento normativo.
  - Likely files/components: hub/internal/orbita/handlers.go, hub/internal/atlas/catalog.go, hub/internal/atlas/offers.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-CAT-03-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 4.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL executar DAG versionado com dependências, mapeamentos, paralelismo limitado, parcialidade, compensações e estado durável por etapa. Sucesso do produto depende do critério contratado; compensação não é rollback automático de efeito externo.
  - Likely files/components: hub/internal/orbita/handlers.go, hub/internal/atlas/catalog.go, hub/internal/atlas/offers.go, migrações/contratos/testes correlatos.
  - Depends on: 4.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 4.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-CAT-03, R2-CAT-04, R2-ADM-05.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 4.2.
  - Validation: cenários R3-CAT-03-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 5. R3-CAT-04 — Resolução indexada e projeção disponível

- [ ] 5.1 Fixar contrato e teste de regressão.
  - Objective: tornar observável o defeito F-R3-09 e o comportamento normativo.
  - Likely files/components: hub/internal/atlas/offers.go, hub/internal/atlasclient/client.go, hub/internal/atlas/catalog.go.
  - Depends on: 1.1; contratos correlatos v4/R2.
  - Validation: três cenários R3-CAT-04-S01/S02/S03 com oráculo independente.
  - Completion criteria: regressão reproduz o problema ou demonstra resolução existente, sem modificar regra para passar.
- [ ] 5.2 Implementar e conectar o fluxo completo.
  - Objective: O Hub SHALL resolver ofertas por chave/vigência indexada sem teto artificial sobre o portfólio e distribuir projeções versionadas ao data plane. Snapshot válido sustenta caminho quente durante falha do controle dentro da validade/revogação definidas; cache frio não autoriza contrato desconhecido.
  - Likely files/components: hub/internal/atlas/offers.go, hub/internal/atlasclient/client.go, hub/internal/atlas/catalog.go, migrações/contratos/testes correlatos.
  - Depends on: 5.1; ordem global definida no roteiro R3.
  - Validation: testes de contrato, integração e falha pertinentes.
  - Completion criteria: mecanismo é chamado no caminho real e todas as disposições de erro são recuperáveis.
- [ ] 5.3 Qualificar e registrar evidência.
  - Objective: comprovar cenários e regressão dos requisitos R2-CAT-06, R2-CAT-08, R2-OPE-06.
  - Likely files/components: evidence/r3 e matrizes em docs/reviews/2026-09-08-r3/implementation.
  - Depends on: 5.2.
  - Validation: cenários R3-CAT-04-S01/S02/S03, sem skip no gate obrigatório.
  - Completion criteria: SHA/digests/comandos/oráculos/resultados registrados; risco residual explicitado.

## 6. Entrega

- [ ] 6.1 Ensaiar migração e rollback e concluir checklist.
  - Objective: habilitação reversível sem perder obrigações históricas.
  - Likely files/components: design.md, migrations, manifests e evidências.
  - Depends on: todas as qualificações anteriores; gate integral r3-07.
  - Validation: canário, rollback compatível e reconciliação.
  - Completion criteria: não há achado encerrado sem prova; bloqueios externos são precisos e não ocultam trabalho executável.
