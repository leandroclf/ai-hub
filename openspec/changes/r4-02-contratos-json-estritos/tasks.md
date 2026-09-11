# Tasks: Contratos JSON estritos
## 1. Revalidação
- [x] 1.1 Confirmar HEAD/diff e fontes atuais.
  - Objective: preservar correções e verificar mudanças posteriores ao snapshot.
  - Likely files/components: explore.md e arquivos citados.
  - Depends on: AGENTS e prompt R4.
  - Validation: comparação de conteúdo e registro de evidência.
  - Completion criteria: cada achado tem estado atual verificável.

## 2. R4-CTR-01 — Documento JSON único e tipos sem coerção
- [x] 2.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-05 observável sem enfraquecer regra.
  - Likely files/components: hub/internal/atlas/offers.go, hub/internal/atlas/offers.go.
  - Depends on: 1.1.
  - Validation: R4-CTR-01-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
- [x] 2.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL aceitar exatamente um documento JSON completo e validar tipo JSON sem coerção por representação Go. null só é válido quando permitido explicitamente; strings numéricas não são números. Entrada deve ser totalmente consumida antes de retornar/persistir o payload, com mesma regra com/sem mapping.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 2.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
- [x] 2.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-CAT-01, R2-CAT-05.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 2.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.

## 3. R4-CTR-02 — Dialeto de schema publicado e semântica numérica
- [x] 3.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-06 observável sem enfraquecer regra.
  - Likely files/components: hub/internal/atlas/offers.go, hub/internal/atlas/offers.go, hub/internal/atlas/catalog.go.
  - Depends on: 1.1.
  - Validation: R4-CTR-02-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
- [x] 3.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL declarar dialeto/subconjunto e compilar/validar schemas na publicação, recusando keywords não suportadas em qualquer nível. Publicação e runtime usam a mesma semântica; equivalência numérica e tipo integer devem seguir o dialeto documentado com precisão exata. Limitar profundidade/custo sem ignorar regras.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 3.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
- [x] 3.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-CAT-01, R2-CAT-01, R2-CAT-05.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 3.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.

## 4. Rollout
- [ ] 4.1 Ensaiar upgrade/rollback e revisar o diff.
  - Objective: preservar dados, contratos e obrigações antigas.
  - Likely files/components: migrações, manifests e relatórios.
  - Depends on: qualificações anteriores.
  - Validation: ensaio compatível e checklist avaliador.
  - Completion criteria: riscos residuais explícitos; sem fechamento por inferência.
