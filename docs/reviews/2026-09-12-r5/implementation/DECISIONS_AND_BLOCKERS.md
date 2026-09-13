# DECISIONS_AND_BLOCKERS — R5

## Decisões técnicas tomadas nesta sessão
1. **Alvo de conflito do insert de callback órfão** (`custody.go`): usar
   `(operation_id, body_sha256, token_hash)` — o mesmo índice único já usado
   pelo insert irmão — em vez de inventar uma nova constraint. Alternativa
   descartada: adicionar uma constraint nova sobre
   `(operation_id, body_sha256, provider_account_id)`, que exigiria migração
   extra sem necessidade, já que `token_hash` já é a identidade determinística
   do evento para contas conhecidas (`callbackAccountIdentity`).
2. **Limite do cache de ofertas** (`atlasclient`): eviction simples (varre
   expirados; se ainda cheio, remove 1 entrada arbitrária) em vez de LRU. É a
   solução mínima que impede crescimento ilimitado sem adicionar dependência;
   suficiente porque o TTL por entrada já é curto.
3. **Captura efetiva** (`libra/store.go`): extraí a lógica de
   `CaptureEffective` para `captureEffectiveTx` (roda dentro da transação já
   aberta por `ApplyEvent`, sob o mesmo `accountLock`) em vez de chamar o
   método público (que abriria uma segunda transação aninhada). O valor
   efetivo usado é a soma dos `economic_facts` de kind `REVENUE` já gravados
   para o protocolo nesta mesma transação — é o único valor monetário
   autoritativo disponível no momento da captura.
4. **Endpoint de replay de quarentena** (`libra/handlers.go`): restrito a
   `hub_admin` porque `finance_quarantine` não tem coluna `tenant_id` (alguns
   motivos de quarentena, como `INVALID_JSON`, nunca chegam a ter uma
   identidade de tenant parseável) — não dá para aplicar o filtro por tenant
   usado nos demais endpoints administrativos.

## Bloqueios reais (não contornáveis nesta sessão)
- **R5-DAD-02 (restore populado/reconciliação)**: exige um ambiente de
  restore real (banco recriado a partir de backup/snapshot, comparação antes
  e depois). Não executado — falta de tempo/escopo de infraestrutura na
  sessão, não indisponibilidade da ferramenta em si (Postgres real está
  disponível no laboratório subido).
- **R5-OPE-03 (kind/k8s, HA regional)**: exige cluster kind e provisionamento
  de perfis local/dev/hom/ppd/prd — não iniciado nesta sessão.
- **Imagem candidata (Dockerfile Go 1.26)**: toolchain local é Go 1.24.0; a
  imagem de produção declarada (`hub/deploy/Dockerfile`) usa Go 1.26. Não
  fiz `docker build` da imagem candidata nesta sessão para requalificá-la —
  gap já registrado na auditoria original, permanece aberto.
- **Decisões comerciais D-01..D-07 e T-R2-01**: fora do escopo técnico;
  nenhuma foi assumida ou fabricada nesta sessão.

## Responsável e próximo passo por bloqueio
| Bloqueio | Próximo passo concreto | Gate reexecutável |
|---|---|---|
| Restore/reconciliação (DAD-02) | Restaurar `hub_core`/`hub_finance` a partir de um dump do laboratório atual, comparar `operation_plans/steps/compensations` e saldo financeiro antes/depois | `hub/deploy/r2/tests/restore-reconciliation.sh` (citado no change r5-03, não executado) |
| kind/HA (OPE-03) | `hub/deploy/r2/kind/bootstrap-independent.sh` já existe e deve ser reaproveitado, não recriado | Scripts em `hub/deploy/r2/kind/` |
| Imagem candidata | `docker build -f hub/deploy/Dockerfile --build-arg SERVICE=orbita .` e requalificar com Go 1.26 real | `hub/deploy/r2/tests/promotion-gate.sh` com manifesto real |
