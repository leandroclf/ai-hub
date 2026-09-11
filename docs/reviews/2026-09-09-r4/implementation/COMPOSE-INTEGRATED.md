# Evidência de qualificação integrada local

Data: 2026-09-10. Projeto Compose único: `ai_hub_r3qual`. Volumes preservados.

- `R2_COMPOSE_PROJECT=ai_hub_r3qual ./hub/deploy/r2/scripts/bootstrap.sh` — PASS após remover o campo Keycloak incompatível `userProfile`.
- Identidade/reconciliação — PASS.
- Upgrade do banco persistido — PASS: checksum R2 da 0002 foi reconciliado somente após verificar `api_key_header` e constraint API_KEY; 0036 e 0037 foram aplicadas.
- Probes HTTP `/healthz/ready` em portas 18080, 18081, 18082, 18083, 18084, 18090, 18091 e 13000 — HTTP 200.
- `go test -race ./...` com `R2_CORE_TEST_DSN`, `R2_FINANCE_DSN`, `ATLAS_TEST_DSN` e S3 local — PASS com o worker ativo; o teste de admissão usa uma célula sintética exclusiva por execução e não disputa a fila de `r2-cell-a`.
- `TestAWSVaultLocalStackVersion` com `SECRETS_ENDPOINT=http://127.0.0.1:14566` e segredo sintético — PASS.

Limites da evidência: o worker Orbita continua consumindo intenções da sua célula operacional por desenho; fixtures concorrentes devem usar célula isolada, como o teste faz. Browser Harness permanece bloqueado por CDP nesta máquina, e browser determinístico, carga, restore/HA e jornada comercial de ponta a ponta continuam limitados ao laboratório sintético, sem homologar produção.
