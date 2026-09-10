# Tasks: Upgrade, cache e crescimento
## 1. Revalidação
- [x] 1.1 Confirmar HEAD/diff e fontes atuais.
  - Objective: preservar correções e verificar mudanças posteriores ao snapshot.
  - Likely files/components: explore.md e arquivos citados.
  - Depends on: AGENTS e prompt R4.
  - Validation: comparação de conteúdo e registro de evidência.
  - Completion criteria: cada achado tem estado atual verificável.

## 2. R4-OPE-01 — L1 utilizável na falha de cofre e coordenação limitada
- [ ] 2.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-07 observável sem enfraquecer regra.
  - Likely files/components: hub/internal/providerauth/client.go, hub/internal/providerauth/client.go.
  - Depends on: 1.1.
  - Validation: R4-OPE-01-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
- [ ] 2.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL consultar cache válido por identidade/versionamento autorizado antes de buscar segredo remoto, respeitar expiração/revogação e coordenar renovação com espera cancelável por chave. Entradas de cache e estruturas de coordenação devem ter limites/evicção seguros sob rotação e crescimento, sem fallback para outro binding.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 2.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
- [ ] 2.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-INT-02, R3-OPE-04, R2-INT-03.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 2.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.

## 3. R4-OPE-02 — Upgrade com migração histórica imutável
- [x] 3.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-08 observável sem enfraquecer regra.
  - Likely files/components: hub/migrations/control/0002_provider_auth.sql, hub/migrations/control/0004_provider_api_key.sql, hub/deploy/r2/scripts/migrate.sh.
  - Depends on: 1.1.
  - Validation: R4-OPE-02-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
- [x] 3.2 Implementar fluxo, dados e integração.
  - Objective: A entrega SHALL preservar migrações aplicadas e implementar mudanças por migrações aditivas. Qualificar instalação limpa e upgrade de volume com checksum R2 anterior sem apagar dados ou desabilitar verificação. Bancos já inicializados com variante modificada precisam reconciliação explícita e restrita a hashes/estados conhecidos, não atualização cega do ledger.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 3.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
- [x] 3.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-OPE-02, R3-OPE-04, R2-OPE-08, R2-DAD-05.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 3.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.

## 4. R4-OPE-03 — Consulta de oferta seletiva sem materializar o portfólio
- [ ] 4.1 Fixar contrato e reproduzir contraexemplos.
  - Objective: tornar F-R4-09 observável sem enfraquecer regra.
  - Likely files/components: hub/internal/atlas/offers.go, hub/internal/atlas/catalog.go, hub/internal/atlasclient/client.go.
  - Depends on: 1.1.
  - Validation: R4-OPE-03-S01/S02/S03 e regressão herdada.
  - Completion criteria: falha atual ou correção existente demonstrada com oráculo.
- [ ] 4.2 Implementar fluxo, dados e integração.
  - Objective: O Hub SHALL resolver pelo conjunto elegível indexado de tenant/aplicação/alvo/versão/vigência, sem materializar catálogo inteiro por pedido, e usar projeção válida conforme a baseline. A prova de crescimento deve medir round-trips/memória/latência e ambiguidade verdadeira, não apenas capacidade de encontrar registro 101.
  - Likely files/components: fontes acima, migrações/contratos/harness correlatos.
  - Depends on: 4.1 e dependências do backlog R4.
  - Validation: integração real e negativas de autorização/erro.
  - Completion criteria: mecanismo conectado e todas as disposições preservam invariantes.
- [ ] 4.3 Qualificar e anexar evidência.
  - Objective: fechar cenários e requisitos herdados R3-CAT-04, R3-INT-03, R2-CAT-06.
  - Likely files/components: docs/reviews/2026-09-09-r4/implementation e hub/evidence/r4.
  - Depends on: 4.2.
  - Validation: cenários completos, sem skip obrigatório.
  - Completion criteria: comando, versão, hash/digest, resultado e logs saneados no índice.

## 5. Rollout
- [ ] 5.1 Ensaiar upgrade/rollback e revisar o diff.
  - Objective: preservar dados, contratos e obrigações antigas.
  - Likely files/components: migrações, manifests e relatórios.
  - Depends on: qualificações anteriores.
  - Validation: ensaio compatível e checklist avaliador.
  - Completion criteria: riscos residuais explícitos; sem fechamento por inferência.
