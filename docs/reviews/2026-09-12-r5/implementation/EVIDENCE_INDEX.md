# EVIDENCE_INDEX — R5

SHA base: `f87ce33034ae29c9431b1910dcc6a633b545e330`. Working tree com as
edições descritas em `EXECUTION_PLAN.md`/`FINDINGS_STATUS.md` (não
commitadas — ver `git status`/`git diff` para o diff exato no momento desta
evidência).

Toolchain: Go `1.24.0 linux/amd64`; Node/Vite conforme `hub/admin-ui/package.json`;
PostgreSQL `16.11` (imagem `postgres:16-alpine`, digest
`sha256:d7ead3a9d3fe4f2906d95e00edc8b36b65749a00a49c47285a35d2be95e876dd`);
LocalStack `3.8`; Redis `7.4-alpine`.

| # | Comando | Ambiente | Resultado | Observação |
|---|---|---|---|---|
| 1 | `go build ./...` (em `hub/`) | local, sem infra | PASS, sem saída | |
| 2 | `go vet ./...` (em `hub/`) | local, sem infra | PASS, sem saída | |
| 3 | `go test ./internal/atlasclient/... ./internal/cometa/... -v` | local, sem infra | 31 passed | antes das correções de cache/teste |
| 4 | `go test ./internal/atlasclient/... -v -run TestOffer` | local, sem infra | 3 passed (2 preexistentes + 2 novos, sendo `TestOfferCacheIsBounded` e `TestOfferDenialIsNeverMaskedByCache`) | valida R5-SEG-02-S01 |
| 5 | `docker --context default compose up -d postgres localstack redis` (`hub/deploy/docker-compose.yml`, projeto `hub-local`) | laboratório local | 3 serviços healthy | localstack recriado manualmente depois para incluir `secretsmanager` (ver CHECKPOINT.md) |
| 6 | `psql` init automático (`hub/deploy/postgres-init/01-init.sh`) | container `hub-local-postgres-1` | Cria `hub_control/hub_core/hub_finance` e aplica todas as migrations, incluindo as 5 novas (0047-0049 core, 0031-0032 finance) sem erro | confirmado via `\dt`/`information_schema.columns` em `hub_finance.reservations` (colunas `reserved_amount/captured_amount/released_amount/settlement_evidence` presentes) |
| 7 | `go test ./... -race -v` com `R2_FINANCE_DSN`, `R2_CORE_TEST_DSN`, `R2_CONTROL_TEST_DSN`, `ATLAS_TEST_DSN`, `R2_S3_TEST_ENDPOINT`, `R2_TEST_SECRET_REF`, `SECRETS_ENDPOINT`, `ENVIRONMENT=local`, credenciais AWS fake | laboratório local (Postgres/LocalStack/Redis reais) | **245 passed, 0 failed, 0 skipped** (30 pacotes) | antes: 69 skips na auditoria original por dependência ausente |
| 8 | `go vet ./...` (pós-integração) | idem | PASS, sem saída | |
| 9 | `DOCKER_CONTEXT=default R2_PG_CONTAINER=hub-local-postgres-1 bash hub/deploy/r2/tests/rls-runtime-proof.sh` | laboratório local | `RLS runtime negative cross-tenant proof (control/core/finance, non-owner role): PASS` | prova real de RLS (R5-SEG-03), não simulada |
| 10 | `R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE=x R2_PROMOTION_APPROVALS="P-01,P-08,P-10" R2_ENVIRONMENT_ISOLATION_PROOF=PASS bash hub/deploy/r2/tests/promotion-gate.sh` (sem manifesto) | local, sem infra | `PROMOTION_GATE=BLOCK environment=prd missing=qualification_evidence_manifest`, exit 1 | confirma fix de R5-OPE-02 |
| 11 | `cd hub/admin-ui && npm run build` (`tsc --noEmit && vite build`) | local, sem infra | PASS — `dist/` gerado, sem erro de tipo | |
| 12 | `npx --yes @fission-ai/openspec@latest validate --all --strict --no-interactive --json` | local, sem infra (rede liberada para npm registry) | `{"items":27,"passed":27,"failed":0}` | inclui os 6 changes r5-0X e a baseline v4/R2/R3/R4 |

## Logs completos
Os logs brutos de cada comando acima ficaram no scroll desta sessão (não
persistidos como arquivo separado por não haver diretório `hub/evidence/r5`
pré-existente para isso). Recomenda-se, na próxima sessão, redirecionar cada
execução para `hub/evidence/r5/<comando>.log` para satisfazer literalmente o
requisito de "logs saneados" do checklist do avaliador.
