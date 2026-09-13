# EXECUTION_PLAN — R5

Baseline: `f87ce33034ae29c9431b1910dcc6a633b545e330`. Ordem herdada de
`docs/reviews/2026-09-12-r5/02-PLANO-DE-IMPLEMENTACAO.md`. Esta sessão
retomou um working tree que já continha implementação substancial em
progresso (não commitada) para praticamente todos os 17 requisitos —
o achado inicial (`docs/reviews/2026-09-12-r5/01-RELATORIO-REVISAO.md`)
descrevia o código **anterior** a essas mudanças; grande parte já estava
corrigida no disco antes desta sessão começar.

## Fatia 0 — Laboratório
- [x] Levantamento de estado real (working tree + documentação) via agentes de exploração.
- [x] Build (`go build ./...`), `go vet ./...`, `go test ./...` limpos.
- [x] Laboratório de dados local (Postgres/Redis/LocalStack) subido sem violar a regra de "um Compose por vez"; ver CHECKPOINT.md.
- [x] `go test ./... -race` executado com **todos** os DSNs/endpoints reais preenchidos: 245 testes, 0 falhas, **0 skips** (antes: 69 skipados por dependência ausente).
- [x] OpenSpec strict: 27/27 changes válidos (`@fission-ai/openspec@latest validate --all --strict`).
- [x] Frontend: `tsc --noEmit && vite build` limpo.

## R5-SEG-01 — Autenticação antes de custódia de callback órfão
- Estado: implementação já presente em `custody.go` (`callbackauth.Verify` antes de `storeOrphan`) — **mas com um bug bloqueante**: `ON CONFLICT DO UPDATE` mal formado (SQL inválido) na inserção com `provider_account_id`.
- [x] **Corrigido nesta sessão**: alvo de conflito ajustado para `(operation_id, body_sha256, token_hash)`, alinhado ao índice único `ux_callback_inbox_capability_identity` (migration 0041) e ao padrão do insert irmão da mesma função.
- [x] **Teste novo**: `TestPostgresOrphanAccountCallbackRequiresSignatureAndDeduplicates` — exercita exatamente o caminho antes quebrado (`StoreOrphanAccountCallback` com `provider_account_id`), prova dedupe por identidade estável em callback repetido (regressão do bug de SQL) e rejeição de assinatura forjada.
- [ ] Pendente: cenário S02 (rotação de chave durante correlação) e S03 completo (conta agressora + outra conta saudável em paralelo, quota isolada) ainda não replicados.

## R5-SEG-02 — Fallback de oferta respeita negação e revogação
- Estado: `atlasclient/client.go` já implementa janela de autorização (`ValidUntil`) desacoplada do TTL de cache, invalidação em 403/409, e `OfferAuthoritative`/`InvalidateOffer` para resolução forçada.
- [x] **Gap fechado nesta sessão**: cache sem limite de entradas (`maxCacheEntries`, sweep de expirados + eviction mínima).
- [x] **Teste novo**: `TestOfferDenialIsNeverMaskedByCache` (S01 — 403/409 nunca mascarado, mesmo após expirar e re-resolver) e `TestOfferCacheIsBounded`.
- [ ] Pendente: S03 (propagação de revogação entre múltiplas réplicas) — validado apenas por construção (TTL curto é por-processo, não há estado compartilhado entre réplicas); sem teste de duas réplicas reais.

## R5-SEG-03 — Adoção real de RLS e autorização de dados auxiliares
- Estado: migrations 0034 (preexistente) + 0049 (nova, `ENABLE`+`FORCE ROW LEVEL SECURITY` aditivo).
- [x] **Verificado nesta sessão com Postgres real**: `deploy/r2/tests/rls-runtime-proof.sh` — PASS. Prova negativa cross-tenant em `hub_core`, `hub_control` e `hub_finance` com role `hub_runtime` (`rolsuper=false, rolcreaterole=false, rolbypassrls=false`).
- [ ] Pendente: inventário completo tabela→owner→tenant→policy→grant citado no design (não refeito nesta sessão); prova de pool reutilizado/worker com tenant alternado (S02) e escopo operador local vs. global (S03) não exercitadas.

## R5-EXE-01..05 e R5-DAD-01 — Produtos, prazos e precisão numérica
- Estado: já implementados no working tree (`intents.go`, `product_store.go`, `product_plan.go`, migrations 0047/0048, `dispatch/contract.go`, `factconsumer.go`, `libra/settlement.go`, `libra/handlers.go`, admin-ui). `go test ./... -race` cobre isso sem falha.
- [ ] Não reexecutei cenário de concorrência dedicado (dois publishers + crash RUNNING/publicação) além do que já existe em `intents_test.go`/`product_store_test.go` — não craquei se cobre literalmente o cenário do achado F-R5-05.

## R5-DAD-02 — Restore populado e reconciliação
- [ ] **Não trabalhado nesta sessão.** Requer ambiente de restore/kind, fora do escopo do que foi possível validar no tempo disponível.

## R5-DAD-03 — Captura por valor efetivo
- Estado anterior: `CaptureEffective`/`SetWatermark` existiam mas **não eram chamados por nenhum fluxo real** (dead code confirmado por grep).
- [x] **Corrigido nesta sessão**: `ApplyEvent` (consumer `revenue`, status SUCCEEDED/PARTIALLY_SUCCEEDED) agora soma os fatos `REVENUE` já gravados para o protocolo e chama `captureEffectiveTx` (nova função interna, extraída de `CaptureEffective` para rodar na mesma transação/lock já aberta), liberando a diferença como `released_amount` sem perder o `UNCERTAIN_HOLD` para os demais estados.
- [ ] Pendente: `SetWatermark` continua sem produtor real de completude (nenhum caller); fechamento de período (`ClosePeriod`) não foi conectado a essa watermark nesta sessão.

## R5-DAD-04 — Quarentena financeira recuperável
- Estado anterior: `ReplayQuarantine`/`MarkQuarantineReplayed` existiam sem endpoint HTTP.
- [x] **Corrigido nesta sessão**: `GET /admin/v1/finance/quarantine` (escopo `hub_admin`, tabela não tem `tenant_id`) e `POST /admin/v1/finance/quarantine/{id}/replay` — decodifica o payload cru em `queue.Envelope` e **reprocessa via `ProcessEnvelope`** (nunca reenvia SUBMIT ao provedor), só marca `REPLAYED` após aceite/re-quarentena durável.
- [ ] Pendente: teste HTTP de integração do endpoint (hoje só compila e passa `go vet`; não escrevi um `handlers_test.go` cobrindo o replay ponta a ponta).

## R5-OPE-01 — Capacidade sem bypass
- [ ] **Não trabalhado nesta sessão** além do que já estava no working tree (`executor.go: ErrCapacityPolicy` quando domínio existe sem `e.capacity`).

## R5-OPE-02 — Gate de promoção com evidência real
- [x] **Verificado nesta sessão**: `promotion-gate.sh` sem `R2_QUALIFICATION_EVIDENCE_MANIFEST` em `prd` → `PROMOTION_GATE=BLOCK missing=qualification_evidence_manifest`, exit 1. Fix já presente no working tree; comportamento confirmado com execução real do script.
- [ ] Pendente: gerar um manifesto real e provar o caminho `ALLOW` com `validate-qualification-evidence.py` (só testei o caminho de bloqueio).

## R5-OPE-03 — Ambientes elásticos duráveis
- [ ] **Não trabalhado.** Requer kind/k8s, fora do tempo desta sessão.

## R5-UX-01 — Console escalável
- Estado: já implementado no working tree (`CatalogPage.tsx` busca server-side, `admin.ts` valida contrato).
- [x] Build TS/Vite limpo confirma compilação; não há teste de browser/E2E executado nesta sessão.

## R5-QUA-01 — Baseline e evidência íntegra
- [x] OpenSpec strict PASS (27/27).
- [ ] Regeneração do inventário/matriz de requisitos e cenários **não refeita** nesta sessão — os CSVs em `docs/reviews/2026-09-12-r5/*.csv` são os da auditoria original, ainda não atualizados com os fixes desta sessão.

## Resumo do que falta para "implementação integral"
Ver `FINAL_REPORT.md` para o veredito consolidado e `DECISIONS_AND_BLOCKERS.md` para os itens que dependem de infraestrutura (kind/k8s, restore, HA regional, decisões comerciais D-01..D-07) não exercitável neste ambiente/tempo.
