# Índice de evidências

## Rodada integrada de 10/09/2026

O índice desta rodada está em
[`EXECUTION-2026-09-10.md`](EXECUTION-2026-09-10.md). Ele supersede as
descrições de bloqueio operacional registradas abaixo, sem superseder os
limites arquiteturais.

| Evidência | Escopo | Resultado |
|---|---|---|
| `hub/evidence/r2/execution/authorized-load-latest.log` | Carga autenticada com seed versionado | PASS; oito caminhos, falha controlada e idempotência |
| `hub/evidence/r2/execution/browser-smoke.json` | Portal Playwright | PASS; OIDC, CRUD/readback, logout e viewport móvel |
| Browser Harness `scenarios/admin-console.py` | 16 rotas administrativas em CDP | PASS-EXPLORATORY; 390×844 sem overflow |
| `internal/providerauth` | Revogação, expiração e limite de locks | PASS com `go test -race -count=1` |
| `restore-reconciliation.sh` | Bancos control/core/finance e S3 | PASS; digest/contagem sem replay |
| Kind `ai-hub-r2` | Prontidão, métricas, HPA/KEDA e recuperação | PASS local; dependências ainda Compose-linked |

## Evidências históricas

| Evidência | Escopo |
|---|---|
| `go test ./...` | compilação e testes unitários atuais |
| `hub/internal/atlas/catalog_test.go` | null/string, string/integer, EOF, minimum, dialeto e inteiro exato |
| `hub/migrations/control/0002_provider_auth.sql` | checksum histórico R2 `37241e604376471efd7de394e9394673feaec0a32476a517c38363890dad1541` |
| `hub/deploy/r2/scripts/migrate.sh` | reconciliação restrita e falha para hash desconhecido |
| `hub/migrations/core/0036_callback_inbox_leases.sql` | lease/epoch/claim aditivo para recuperação da inbox |
| `hub/internal/cometa/worker.go` | worker autônomo, lote limitado e disposição de poison item |
| `hub/migrations/control/0037_catalog_offer_lookup.sql` | índice parcial do conjunto elegível de ofertas |
| `hub/internal/atlas/offers.go` | consulta seletiva com limite de ambiguidade |
| `COMPOSE-INTEGRATED.md` | bootstrap, OIDC, upgrade, probes e testes integrados locais |
| `git diff --check` | higiene do diff |
| `docker compose ls`, `docker ps -a` | proveniência do runtime local |
| `../evidence/openspec-strict-20260910.json` | OpenSpec strict atual: 21 changes, 0 falhas, SHA `bd8b63f` |
| `OPENSPEC-AUDIT-2026-09-10.md` | auditoria sequencial das quatro changes R4 e critérios para não encerrar por inferência |
| `hub/deploy/r2/tests/browser-harness/README.md` | procedimento permanente para validação exploratória do console com browser real |
| `hub/deploy/r2/tests/browser-harness/run.sh` e `scenarios/admin-console.py` | runner e cenário somente leitura do Browser Harness; resultado é exploratório, não substitui Playwright |
| 10/09/2026: kind, readiness, métricas/HPA e recuperação de pod | laboratório `ai-hub-r2` | PASS local; dependências completas ainda não estão dentro do cluster |
| 10/09/2026: smoke Chromium autenticado | `hub/evidence/r2/execution/browser-smoke.json` | PASS: OIDC PKCE, senha/OTP, CRUD, readback, logout e 390px |
| 10/09/2026: carga autorizada | `hub/evidence/generate_traffic.sh` | BLOCKED: token aceito, mas catálogo sem recursos/publicações e retorno `403 offer_not_eligible` |
| 10/09/2026: Browser Harness | `hub/deploy/r2/tests/browser-harness/run.sh` | BLOCKED-ENVIRONMENT: comando não instalado |

Nenhuma evidência histórica foi promovida como PASS de integração.
