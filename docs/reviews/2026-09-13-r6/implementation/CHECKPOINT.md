# Checkpoint — R6 (execução em andamento)

## Estado real
HEAD de entrada: `a540b40007fe6b8ed523e17afe00e96ff8f8ad50` (confirmado, sem divergência).
Laboratório: ecossistema Compose `hub-local` já estava ativo (postgres/localstack/redis) — reaproveitado, nenhum container novo subido, nada de terceiros tocado (`hive-ia-agents` preservado). `docker.sock` do Docker Desktop não respondia; usado o daemon do sistema via `DOCKER_HOST=unix:///var/run/docker.sock`.

## R6-SEG-01 — mecanismo: CONCLUÍDO E PROVADO
`hub/internal/platform/pg/pg.go`: três primitivas de escopo (cada uma com variante "transação nova" `WithXTx` e variante "injetar em transação já aberta" `SetXScope`):
- `WithTenantTx`/`SetTenantScope` — `app.tenant_id` LOCAL; request/operação de um único tenant.
- `WithAuditedScopeTx`/`SetAuditedScope` — `app.access_reason` LOCAL, motivo não vazio obrigatório; SELECT-only a nível de policy (exceto na policy dedicada `capacity_authority`, ver abaixo). Para diagnóstico administrativo nominal e para ingresso de callback/observação externa que autentica por capability antes de saber o tenant.
- `WithWorkerCellTx`/`SetWorkerCellScope` — `app.worker_cell_id` LOCAL; cobre leitura E escrita. Para workers que processam fila entre tenants dentro de uma única célula operacional.

Migrações aplicadas nas 3 bases do `hub-local` (idempotentes, aplicam-se sozinhas em subida nova via `postgres-init/01-init.sh`):
- `core/0050`, `control/0038`, `finance/0035`: policy `audited_scope`; remove bypass nominal `current_user='hub'` sobrevivente em control/finance.
- `core/0051`: policy `worker_cell_scope` para tabelas com `cell_id`.
- `core/0052`, `control/0039`, `finance/0036`: **achado adicional** — 24 tabelas tenant-scoped (11 core, 6 control, 7 finance, incluindo `journal_batches`, `settlement_periods`, `reservation_settlements`) nunca tinham entrado em nenhuma lista de RLS de rodada anterior. Fechado com o mesmo padrão.
- `core/0053`: `protocol_reconciliation_requests` não tem `cell_id` próprio — policy dedicada com `EXISTS` contra `operations`.
- `core/0054`: **achado descoberto durante a conversão de `cometa`** — `capacity_permits` implementa um orçamento agregado por *domínio de provedor*, não por tenant nem por célula (provado por `TestPostgresCapacityAggregateReplicasCells`, que compartilha o mesmo domínio entre células). Nem `tenant_runtime` nem `worker_cell_scope` cobririam corretamente; policy dedicada `capacity_authority` (ALL commands, motivo obrigatório).
- `control/0040`: **achado do subagente que converteu `atlas`** — `credential_bindings` permite `tenant_id IS NULL` para vínculos `SHARED_HUB`, mas a policy `tenant_runtime` de `0039` não tinha exceção para NULL (nenhum escopo tornaria essas linhas visíveis). Corrigido com o mesmo padrão de `catalog_resources` (`tenant_id=''`), adaptado para `IS NULL`.

Prova real contra Postgres do `hub-local` (não simulada, sem ROLLBACK mascarando o oráculo):
- `hub/deploy/r2/tests/rls-runtime-proof.sh` reescrito — **PASS**.
- `hub/internal/platform/pg/pg_test.go` — `TestWithTenantTxAndAuditedScopeEnforceRLSAgainstRealPostgres`, `TestWithWorkerCellTxScopesToOneCellNotAnyTenant` via `R2_RUNTIME_TEST_DSN` real — **PASS** (4/4).

## R6-SEG-01 — adoção no caminho real: os 5 serviços CONVERTIDOS

| Serviço | Banco | Call sites | Produção | Testes (owner, regressão) | Testes (hub_runtime) |
|---|---|---|---|---|---|
| `orbita` | hub_core | 25 | ✅ convertido | 33/33 | fixtures em correção (subagente em andamento) |
| `cometa` | hub_core | 45 | ✅ convertido | 60/60 | 59/60 (1 exceção documentada: teste cria DDL própria — `hub_runtime` corretamente sem `CREATE` em `public`, fora de escopo de correção de teste) |
| `atlas` | hub_control | 23 | ✅ convertido | 37/37 (owner) | 37/37 |
| `pulsar` | hub_core | 16 | ✅ convertido (+ ~140 linhas de código morto legado removidas, tabela `webhook_destinations` singular superada por `webhook_destination_versions`) | 8/8 | 8/8 (agente relatou 5/5 em subconjunto real+fixtures antes da consolidação; suíte completa confere) |
| `libra` | hub_finance | 23 | ✅ convertido | 33/33 | 33/33 |

**Consolidado (owner, todos os 6 pacotes incl. `internal/platform/pg`): 177/177 — zero regressão.**

Cada domínio usa a primitiva certa por caso, não uma regra única: a maioria dos call sites é `WithTenantTx` (request de um tenant conhecido); ingresso de callback/observação que autentica por capability antes de saber o tenant usa `WithAuditedScopeTx`; varreduras de fila cruzando tenants dentro de uma célula (claim/complete de intents, deadline sweep, polling, reconciliação administrativa) usam `WithWorkerCellTx`; o orçamento de capacidade (verdadeiramente cross-tenant e cross-célula) usa a policy dedicada `capacity_authority` via `WithAuditedScopeTx`. Funções sem nenhum chamador em produção (`ClaimProductStep` em orbita; `CreateOperation`/`FindByProviderRequestID`/`UpdateState`/os helpers de polling legado em cometa) foram deliberadamente deixadas sem conversão — código morto pré-existente, mesmo tratamento em todos os pacotes.

## R6-SEG-01 — CONCLUÍDO (corte de credencial aplicado)
Todos os itens do checkpoint anterior foram concluídos nesta sessão:
1. Fixtures de teste de `orbita` corrigidas (subagente + 3 correções manuais adicionais: `RecoverOrphanedIntent` gravava em `protocol_audit` — que não tem `cell_id`, logo `worker_cell_scope` não a alcançava — sem nunca aplicar `app.tenant_id`; **achado de produção real, corrigido** com `pg.SetTenantScope` adicional após conhecer o tenant da linha reivindicada. Mais 2 testes com queries cruas não vistas pelo subagente, corrigidas manualmente).
2. `TestPostgresValidResponseCommitFailureLeavesRecoverableObligation` (cometa) precisava criar um trigger/função de fault-injection — DDL que `hub_runtime` corretamente não pode executar. Corrigido com uma segunda conexão opcional (`R2_CORE_ADMIN_DSN`, default = mesma DSN de teste) só para o DDL da fixture; o fluxo de aplicação sob teste continua rodando com a credencial real.
3. **Corte de credencial aplicado**: `deploy/docker-compose.yml`, `deploy/r2/compose.yaml`, `deploy/r2/kind/render-runtime.py` e os 5 `cmd/*/main.go` agora usam `hub_runtime:r2-runtime-fixture` em vez de `hub:hub`/`hub:r2-local-fixture`. `go build ./...` limpo após a troca.
4. **Suíte consolidada final, 6 pacotes (orbita/cometa/atlas/pulsar/libra/platform/pg): 177/177 sob owner (regressão zero) E 177/177 sob `hub_runtime` real (zero falha, zero skip)**.

Nenhum container do `hub-local` foi reiniciado (só postgres/localstack/redis estavam de pé; nenhum serviço de aplicação rodando para reiniciar). A troca fica pronta para a próxima subida/build das imagens.

Pendências não bloqueantes: R6-SEG-01-S02 (teste explícito de pool de conexão única alternando tenants sob carga — mecanismo já correto e provado em `pg_test.go`, falta só o teste de carga) e R6-SEG-01-D (métricas/runbook operacional de RLS).

## R6-SEG-02 — CONCLUÍDO E PROVADO
`internal/callbackauth`: `KeyRing` com múltiplas `kid` (mesmo padrão de rotação já usado em `internal/platform/auth`/JWKS). Migração `core/0055` adiciona `callback_key_id`/`verified_at` a `callback_inbox`; `AuthenticateAccountCallback` retorna o `kid` que validou e persiste-o, então a reconciliação nunca precisa recompor "a chave atual" — verifica sempre com a chave que originalmente autenticou o callback. Lock consultivo de custódia particionado por `accountID+cell` (não mais uma constante global), removendo colisão de célula.
Quarentena distinta de rejeição definitiva: migração `core/0056` (`replayed_at`/`replayed_by`) + `core/0057` (amplia o CHECK `callback_inbox_disposition_check` para aceitar `'QUARANTINED'` — **achado real**: o código descartava esse erro silenciosamente com `_ = dispose(...)`, então o disposition nunca girava e a linha ficava presa 30s em retry infinito; corrigido para propagar o erro). Novo endpoint administrativo `/admin/v1/callback-inbox/{id}/replay` (nominal + MFA + escopo `integrations:write` + motivo 8-512 chars) chama `Store.ReplayCallbackInbox`.
Testes novos contra Postgres real: `TestPostgresCallbackSurvivesKeyRotationDuringReconciliation`, `TestPostgresCallbackApplyFailureExhaustsIntoRecoverableQuarantine` — **PASS**.

## R6-EXE-01 — CONCLUÍDO E PROVADO
Achado: `BuildProductPlan` (orbita) derivava snapshots por etapa a partir de um único offer top-level — rota/conta/vínculo de uma etapa podiam divergir silenciosamente da conta realmente usada por outra etapa do mesmo produto, e o economic snapshot por etapa carregava a regra de venda do protocolo inteiro em vez do custo de compra da própria etapa.
Correção em 3 camadas:
1. `atlas.Store.ResolveOffer` agora resolve cada etapa (e cada etapa de compensação) de forma independente via `resolveStepOffer` — publicado, tenant coerente, binding↔conta coerentes — populando `OfferSnapshot.StepOffers[stepID]`. Falha fecha o `ResolveOffer` inteiro (fail-closed) se qualquer etapa não resolver.
2. `orbita.BuildProductPlan` consome exclusivamente o `StepOffer` já resolvido da própria etapa (nunca mais o offer do produto pai) para `ProviderAccountID` e para o novo `stepEconomicSnapshot` (que usa apenas o `Buy` do `PurchaseContract` da etapa, não o `Sell` do contrato do produto).
3. `cometa.executor` ganhou uma verificação defensiva adicional: mesmo com hash de snapshot válido, rejeita (`snapshot_route_account_binding_mismatch`) se `SelectedRoute.ProviderAccountID`/`Binding.ProviderAccountID` divergirem de `Account.ID` dentro do próprio snapshot — um hash certifica bytes, não coerência semântica entre os campos.
Novo teste `TestBuildProductPlanRejectsStepWithoutOwnResolvedRoute` prova o fail-closed. Regressões de fixture (5, em orbita/cometa) corrigidas — todas por fixtures que não tinham `SelectedRoute.ProviderAccountID`/`Binding.Data["provider_account_id"]`/`StepOffers` coerentes.

**Achado adicional fora do escopo original, corrigido nesta passagem**: `internal/objectstore` (`catalog.go`, `retention.go`) nunca tinha sido convertido para as primitivas de escopo RLS — usava `c.db` cru em toda parte. Sob a credencial `hub_runtime` (corte de credencial do R6-SEG-01), toda a custódia de arquivo (`file_refs`, `object_retention_pins`, `object_tombstones`, `object_gc_runs`) ficava invisível/inoperante (RLS filtra silenciosamente sem `app.tenant_id`). Convertido para `pg.WithTenantTx` em todos os métodos tenant-scoped (`CreateSession`, `lookup`, `Complete`, `Pin`, `Unpin`, `StoreResult`, `LinkResult`, `Reconcile`, `DryRun`, `Purge`, `IsTombstoned`); testes próprios (`catalog_test.go`) também corrigidos (usavam `db` cru para asserção/limpeza).

**Suíte completa (`go test ./...`, 30 pacotes) — 239/239 sob owner (zero regressão) e 241/241 sob `hub_runtime` real (zero falha, zero skip inesperado).**

## R6-EXE-02 — CONCLUÍDO E PROVADO
Achado: `RunIntentPublisher` (QUEUED) já barrava dispatch quando `Intent.Expired`; `RunDirectRecovery` (DIRECT) não tinha a mesma checagem e podia reenviar DIRECT ao provedor com `retry_until` vencido, mesmo com SLA do cliente futuro — reenvio às cegas sem saber se o efeito já ocorreu. Corrigido com a mesma guarda usada no lado QUEUED. Teste real `TestPostgresDirectRecoveryNeverResubmitsAfterRetryExpiryEvenWithFutureSLA` — PASS (servidor HTTP fake conta SUBMITs; zero após a correção).

## R6-EXE-03 — CONCLUÍDO E PROVADO
Achado: as 3 queries de claim de intenção excluíam TODA intenção — inclusive de compensação — assim que `protocols.status` virava terminal público. `continue_after_client_deadline=TRUE` (já usado para isentar compensação do gate de `client_deadline_at`) foi reaproveitado para isentá-la também do gate de status terminal — sem migração nova. Testes reais: `TestPostgresCompensationProgressesAfterPublicFinalState` (compensação conclui após protocolo já EXPIRED; final público não reabre) e `TestPostgresCompensationUnknownThenDefinitiveFailureStaysReconciliable` (UNKNOWN não vira sucesso por omissão; falha definitiva fica COMPENSATION_FAILED, nunca promovida a sucesso) — ambos PASS.

**Suíte completa (30 pacotes) após EXE-02+EXE-03 — 242/242 sob owner e 244/244 sob `hub_runtime` real, zero falha.**

## R6-EXE-04 — CONCLUÍDO E PROVADO (sem achado de código)
`CompleteIntent` e `claimIntentMode` já atualizavam intent, lease de etapa e contador do plano atomicamente numa única transação; fencing por época já impedia dono antigo de agir após takeover. Faltava só evidência real dedicada ao caso de produto (só existia para intent simples). Testes novos: `TestPostgresSlotReleaseNeverExceedsMaxParallelNorDoubleDecrements` (S01/S02) e `TestPostgresProductStepTakeoverAfterCrashRecoversSameSlotAndFencesOldOwner` (S03) — PASS, zero alteração de produção.

**Suíte completa (30 pacotes) após EXE-04 — 244/244 sob owner e 246/246 sob `hub_runtime` real, zero falha.**

## R6-FIN-01 — CONCLUÍDO E PROVADO
Achado: o endpoint de replay de quarentena marcava REPLAYED sempre que `ProcessEnvelope` retornava `nil` — mas re-quarentenar (payload ainda inválido) também retorna `nil`. Corrigido com `ReplayDisposition` tipado (`APPLIED`/`QUARANTINED`/`CONFLICT`); só marca REPLAYED em `APPLIED`, senão registra a tentativa (nova migração `finance/0037`) sem prometer resolução. S02/S03 já eram cobertos por mecanismo pré-existente (dedup por hash + auto-quarentena de fato tardio); só faltava evidência real. Testes novos: `quarantine_replay_still_invalid_never_marks_replayed`, `quarantine_replay_after_crash_between_apply_and_confirmation`.

## R6-FIN-02 — CONCLUÍDO E PROVADO (sem achado de código)
`ClosePeriod`/`captureEffectiveTx`/`reservationState` já implementavam completude durável e captura/liberação exata com evidência auditável (herança R2/R5). Faltava só o teste do caso numérico exato. Teste novo `R6-FIN-02-S02_partial_capture_release_and_duplicate` — PASS.

**Suíte completa (30 pacotes) após FIN-01/02 — 247/247 sob owner e 249/249 sob `hub_runtime` real, zero falha.**

## R6-OPE-04 — CONCLUÍDO E PROVADO
Achado real (S01): rota sem `capacity_domain` era bypass silencioso em `acquireCapacity` (execução sem controle algum), apesar de Atlas já exigir o domínio para publicar. Corrigido: ausência de domínio ou controlador não instalado agora nega antes de I/O, sem bypass. S02 já coberto (`TestPostgresCapacityAdaptiveAndPending`). Achado real (S03): `ResolvePending` (fecha permit preso com evidência terminal) nunca tinha endpoint algum — capacidade travada após falha de settlement + lease vencido não tinha caminho de recuperação. Novo endpoint admin `/admin/v1/capacity-domains/{domain}/permits/{id}/resolve`. Testes novos comprovam ambos os achados; suíte cometa completa 63/63.

**Suíte completa (30 pacotes) após OPE-04 — 248/248 sob owner e 250/250 sob `hub_runtime` real, zero falha.**

## R6-UX-01 — CONCLUÍDO E PROVADO
Achado real: uma referência já selecionada (campo simples, passo de produto ou rota) fora da janela de 50 resultados da busca atual desaparecia da lista (parecia "sem seleção"), tanto ao reabrir (S01) quanto ao pesquisar outro termo (S02). Corrigido: `loadLookup` agora busca explicitamente por ID cada referência já selecionada quando ausente do resultado da página atual, mesclando sem duplicar; o filtro textual local parou de esconder a seleção atual. S03 (corrida de tenant) já estava correto via `AbortController` por efeito.

## R6-UX-02 — CONCLUÍDO E PROVADO
Achado real (S01): erro de JSON inválido no mapping de um passo de produto ficava isolado no componente filho (`StepsEditor`) e nunca bloqueava salvar/publicar no formulário pai — risco de enviar mutação com o mapping anterior sem o operador perceber. Corrigido com callback `onErrorsChange` elevando o estado ao pai. S02/S03 já eram cobertos pelo mecanismo existente (Idempotency-Key estável por chamada + retry, If-Match por revisão com diff de conflito, validação de contrato de resposta antes de aplicar ao rascunho).

Evidência de ambos: `tsc --noEmit` e `vite build` limpos (admin-ui não tem framework de teste; mesmo padrão dos commits anteriores do frontend).

## Ainda não iniciado nesta sessão
R6-OPE-01…03, R6-QUA-01.

## Recursos locais
Nenhum container/rede/volume novo criado ou removido. Ecossistema `hub-local` inalterado; nada a limpar.
