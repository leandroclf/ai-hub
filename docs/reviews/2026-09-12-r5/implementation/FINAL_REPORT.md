# FINAL_REPORT — R5 (parcial e delimitado)

## Veredito
**Implementação parcial.** Esta sessão não fecha os 17 requisitos R5 nem as
25 fatias herdadas por completo — o escopo total (segurança distribuída,
execução transacional concorrente, financeiro, HA/k8s, console) excede o que
é honesto entregar como "integral" em uma única sessão. O que foi possível
fazer foi: (a) confirmar que grande parte da implementação dos 17 requisitos
já existia, não commitada, no working tree ao início da sessão; (b) encontrar
e corrigir um bug bloqueante real; (c) fechar dois gaps concretos de "helper
implementado mas nunca chamado"; (d) provar, com infraestrutura real (não
mocks), que a suíte inteira de testes do repositório passa com **zero
skips**, e que dois gates de segurança/promoção funcionam de fato.

## O que foi corrigido e verificado nesta sessão
1. **Bug bloqueante corrigido**: `hub/internal/cometa/custody.go` —
   `storeOrphan` tinha um `INSERT ... ON CONFLICT DO UPDATE / DO UPDATE SET`
   inválido (SQL malformado) que quebraria em runtime todo callback órfão com
   `provider_account_id` preenchido. Corrigido para o alvo de conflito
   correto, já usado no índice único existente. Teste novo
   (`TestPostgresOrphanAccountCallbackRequiresSignatureAndDeduplicates`) prova
   com Postgres real que o caminho antes quebrado agora deduplica
   corretamente e rejeita assinatura forjada (R5-SEG-01-S01).
2. **R5-SEG-02** (fallback de oferta): cache local do `atlasclient` agora tem
   limite de entradas (evita crescimento ilimitado). Dois testes novos provam
   que uma negação 403/409 nunca é mascarada pelo cache, mesmo após reresolver.
3. **R5-SEG-03** (RLS real): a prova negativa de isolamento cross-tenant
   (`rls-runtime-proof.sh`) foi executada contra um Postgres real nesta
   sessão e **passou** nas três bases (`hub_core`, `hub_control`,
   `hub_finance`), com o papel runtime confirmado sem superpoderes.
4. **R5-DAD-03** (captura por valor efetivo): `CaptureEffective`, que existia
   como código morto, agora é chamado de dentro de `ApplyEvent` sempre que um
   fato de receita é confirmado (`SUCCEEDED`/`PARTIALLY_SUCCEEDED`), usando o
   valor efetivo somado dos fatos econômicos já gravados e liberando a
   diferença da reserva.
5. **R5-DAD-04** (quarentena recuperável): endpoints administrativos novos
   (`GET/POST /admin/v1/finance/quarantine...`) conectam `ReplayQuarantine`/
   `MarkQuarantineReplayed` (também código morto até então) ao fluxo real,
   reprocessando o fato original via `ProcessEnvelope` — nunca reenviando
   SUBMIT ao provedor.
6. **R5-OPE-02** (gate de promoção): confirmado com execução real que `prd`
   sem manifesto de evidência bloqueia (`PROMOTION_GATE=BLOCK`).
7. **Gate obrigatório sem skip**: laboratório local (Postgres 16.11 real,
   LocalStack 3.8 real com SQS/SNS/S3/Secrets Manager, Redis 7.4) subido sem
   violar a regra de "um ecossistema Compose por vez" (verificado antes via
   `docker compose ls`/`docker ps -a`; único outro projeto ativo na máquina,
   `hive-ia-agents`, preservado intocado). Com isso, `go test ./... -race`
   passou de 137 PASS/69 SKIP (auditoria original) para **245 PASS / 0 SKIP**.
8. OpenSpec strict (27/27), `go vet`, e build do frontend (`tsc --noEmit &&
   vite build`) confirmados limpos.

## O que NÃO foi feito (limitação material, não fabricação de sucesso)
- **R5-DAD-02** (restore populado/reconciliação) e **R5-OPE-03** (kind/HA
  regional): não iniciados — exigem infraestrutura e tempo além do possível
  nesta sessão.
- **R5-SEG-01/SEG-03**: implementação real existe e um bug foi corrigido, mas
  os cenários adversariais S01-S03 completos (dedupe sob ataque, rotação de
  chave, pool reutilizado com tenant alternado, escopo operador global) não
  foram replicados individualmente com oráculo dedicado.
- **R5-EXE-01..05/DAD-01**: já implementados no working tree antes desta
  sessão; confirmados apenas indiretamente (suíte de testes existente passa),
  sem reprodução isolada de cada cenário do achado original.
- **Imagem candidata** (`hub/deploy/Dockerfile`, Go 1.26): não requalificada
  nesta sessão; toolchain local permanece Go 1.24.0.
- **QUA-01**: inventário/CSVs de requisitos e cenários da auditoria original
  não foram regenerados a partir das specs atuais.
- Nenhum push, merge, deploy remoto, aprovação comercial ou dado de produção
  foi tocado, conforme instrução.

## Caminhos dos artefatos
- `docs/reviews/2026-09-12-r5/implementation/{EXECUTION_PLAN,FINDINGS_STATUS,DECISIONS_AND_BLOCKERS,CHECKPOINT,EVIDENCE_INDEX,FINAL_REPORT}.md`
- `docs/reviews/2026-09-12-r5/implementation/{REQUIREMENTS_STATUS,SCENARIO_RESULTS}.csv`
- Código alterado: `hub/internal/cometa/custody.go`, `hub/internal/atlasclient/client.go`,
  `hub/internal/atlasclient/client_test.go`, `hub/internal/libra/store.go`,
  `hub/internal/libra/handlers.go` (diffs sobre o working tree já em progresso,
  não commitados nesta sessão).

## Recomendação de continuidade
Seguir a ordem do `EXECUTION_PLAN.md`: fechar os cenários S01-S03 de SEG-01/
SEG-03 com teste dedicado, depois atacar DAD-02 (restore) e OPE-03 (kind) —
os dois maiores blocos de esforço de infraestrutura restantes — e por fim
regenerar o inventário QUA-01 a partir do estado real do código.
