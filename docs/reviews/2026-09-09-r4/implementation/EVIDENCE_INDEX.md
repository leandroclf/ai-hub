# Índice de evidências

## Rodada integrada de 10/09/2026

O índice desta rodada está em
[`EXECUTION-2026-09-10.md`](EXECUTION-2026-09-10.md). Ele supersede as
descrições de bloqueio operacional registradas abaixo, sem superseder os
limites arquiteturais.

| Evidência | Escopo | Resultado |
|---|---|---|
| `hub/evidence/r2/execution/authorized-load-latest.log` | Carga autenticada com seed versionado | PASS; oito caminhos, falha controlada e idempotência |
| `hub/evidence/r2/execution/browser-smoke.json` | Portal Playwright | PASS; OIDC, CRUD/readback, SLA, destinos versionados, logout e viewport móvel |
| Browser Harness `scenarios/admin-console.py` | 16 rotas administrativas em CDP | BLOCKED-ENVIRONMENT; executável presente, daemon sem `DevToolsActivePort` utilizável |
| `hub/internal/cometa/custody.go` + migração `0041` | Inbox de callback | PASS de implementação/testes focados; identidade por capability, limites de bytes/itens e retenção periódica; qualificação externa ainda aberta |
| `hub/internal/cometa/capacity.go` + executor/poller/reconciliation + `hub/internal/pulsar/worker.go` | Capacidade por domínio | PASS de integração local; concessões em `SUBMIT`/`STATUS`, reconciliação e `FETCH` de webhook; budgets integrais de todos os I/O ainda abertos |
| `hub/internal/orbita/admission.go` + `hub/internal/pulsar/custody.go` + `custody_test.go` | Snapshot de destino webhook e orçamento de entrega | PASS de teste PostgreSQL; versão aceita é congelada, entrega usa o snapshot persistido e o permit é encerrado com sinal/evidência |
| `internal/providerauth` | Revogação, expiração e limite de locks | PASS com `go test -race -count=1` |
| `restore-reconciliation.sh` | Bancos control/core/finance e S3 | PASS com sufixos `r4_sequence_20260910c` e `r4_20260911qual2`; digest/contagem sem replay e colisão de alvos recusada |
| `hub/evidence/r2/execution/capacity-reconciliation-latest.log` | Reconciliação controlada das concessões locais após falhas de transporte | PASS local; 12 ausências comprovadas pelo oráculo sintético fechadas, zero efeitos presentes protegidos; não aplicável a provedor comercial |
| `hub/internal/pulsar/custody_test.go` | Entrega concorrente com capacidade `FETCH` | PASS com PostgreSQL real; duas entregas HMAC concorrentes, `transport_open=0` e `pending_external=0` |
| `hub/deploy/r2/tests/webhook-capacity-runtime.sh` + `webhook-capacity-latest.log` | Entrega no worker Pulsar real do Compose | PASS; delivery sintética chegou a `DELIVERED` e o permit `r4-webhook` terminou `open=0`, `pending=0`, com limpeza do registro de teste |
| `hub/deploy/r2/tests/finance-runtime-proof.sh` + `finance-runtime-latest.log` | Incidência financeira do workload autorizado | PASS; outbox com `SUBMITTED`/`STATUS` carregou `attempt_id`, Libra recebeu fatos de custo/receita, inbox foi persistida, journal balanceou por moeda e não houve quarentena de incidência inválida |
| `hub/internal/cometa/polling_custody.go` + `hub/internal/libra/consumers.go` | Separação entre estado operacional e incidência econômica | PASS com PostgreSQL real; aceitação UNKNOWN publica `SUBMITTED`, polling publica `STATUS` por tentativa e fencing não publica observação não autorizada |
| `hub/internal/orbita/admission_test.go` | Isolamento do teste concorrente de admissão | PASS 20 repetições e suíte completa com o worker Orbita ativo; célula sintética não compartilhada com a execução do laboratório |
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
| `../evidence/openspec-strict-20260910.json` | OpenSpec strict reexecutado: 21 changes, 0 falhas; a revisão deve considerar o SHA registrado no artefato após o commit |
| `OPENSPEC-AUDIT-2026-09-10.md` | auditoria sequencial das quatro changes R4 e critérios para não encerrar por inferência |
| `hub/deploy/r2/tests/browser-harness/README.md` | procedimento permanente para validação exploratória do console com browser real |
| `hub/deploy/r2/tests/browser-harness/run.sh` e `scenarios/admin-console.py` | runner e cenário somente leitura do Browser Harness; resultado é exploratório, não substitui Playwright |
| 10/09/2026: kind, readiness, métricas/HPA e recuperação de pod | laboratório `ai-hub-r2` | PASS local; dependências completas ainda não estão dentro do cluster |
| 10/09/2026: smoke Chromium autenticado | `hub/evidence/r2/execution/browser-smoke.json` | PASS: OIDC PKCE, senha/OTP, CRUD, readback, logout e 390px |
| 10/09/2026: carga autorizada | `hub/deploy/r2/tests/authorized-load.mjs` | PASS: oito caminhos autorizados, falha controlada, polling/callback/AUTO, saldo estrito, credencial dedicada e quatro observações idempotentes |
| 10/09/2026: Browser Harness | `hub/deploy/r2/tests/browser-harness/run.sh` | BLOCKED-ENVIRONMENT: executável instalado, mas Chrome/daemon não expôs `DevToolsActivePort` utilizável |

Nenhuma evidência histórica foi promovida como PASS de integração.
