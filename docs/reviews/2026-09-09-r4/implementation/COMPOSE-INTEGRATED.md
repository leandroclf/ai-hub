# Evidência de qualificação integrada local

Data: 2026-09-11. Projeto Compose único: `ai_hub_r3qual`. Volumes preservados.

- `R2_COMPOSE_PROJECT=ai_hub_r3qual ./hub/deploy/r2/scripts/bootstrap.sh` — PASS após remover o campo Keycloak incompatível `userProfile`.
- Identidade/reconciliação — PASS.
- Upgrade do banco persistido — PASS: checksum R2 da 0002 foi reconciliado somente após verificar `api_key_header` e constraint API_KEY; 0036 e 0037 foram aplicadas.
- Probes HTTP `/healthz/ready` em portas 18080, 18081, 18082, 18083, 18084, 18090, 18091 e 13000 — HTTP 200.
- `go test -race ./...` com `R2_CORE_TEST_DSN`, `R2_FINANCE_DSN`, `ATLAS_TEST_DSN` e S3 local — PASS com o worker ativo; o teste de admissão usa uma célula sintética exclusiva por execução e não disputa a fila de `r2-cell-a`.
- `TestAWSVaultLocalStackVersion` com `SECRETS_ENDPOINT=http://127.0.0.1:14566` e segredo sintético — PASS.
- Produto composto via API pública — PASS: duas etapas independentes,
  duas operações/efeitos no provider-sim, finalização `SUCCEEDED` e repetição
  idempotente sem novo efeito.
- Entrega webhook com `r4-webhook` e incidência financeira — PASS após
  recalcular as CIDRs das fixtures; permit terminou `open=0/pending=0`, journal
  balanceou por moeda e não houve incidência inválida.
- Portal administrativo — PASS no smoke Chromium: OIDC/OTP, CRUD/readback,
  editor de produto com mapeamento entre etapas, destinos, SLA, financeiro,
  sessão sem armazenamento persistente e viewport 390px.

O runtime deve ser iniciado pelo bootstrap oficial. Uma execução isolada de
`docker compose up` pode aplicar os defaults de loopback em
`EGRESS_PRIVATE_RULES`, que não correspondem aos IPs internos das fixtures e
fazem a entrega webhook falhar fechado antes do POST. O bootstrap descobre as
CIDRs atuais e recria a configuração sem remover volumes.

Limites da evidência: o worker Orbita continua consumindo intenções da sua célula operacional por desenho; fixtures concorrentes devem usar célula isolada, como o teste faz. Browser Harness passou como exploração com `BU_CDP_URL` explícito, enquanto browser determinístico, carga, restore/HA e jornada comercial de ponta a ponta continuam limitados ao laboratório sintético, sem homologar produção.
