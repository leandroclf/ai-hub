# Evidence index — R6 (estado vivo)

Todo item abaixo foi gerado nesta sessão, contra a implementação real (sem invenção de resultado). SHA de entrada: ver `evidence/SHA.txt`.

| Item | Comando | Ambiente | Log | Resultado |
|---|---|---|---|---|
| Prova RLS (SQL direto, role `hub_runtime`, sem ROLLBACK mascarando oráculo) | `bash hub/deploy/r2/tests/rls-runtime-proof.sh` | Postgres real do ecossistema Compose `hub-local` (já ativo, reaproveitado) | `evidence/rls-runtime-proof.log` | PASS — isolamento próprio/cruzado, recusa de escrita cruzada, `audited_scope` (SELECT-only, motivo obrigatório) nas 3 bases (core/control/finance) |
| Prova RLS via API Go (`pg.WithTenantTx`/`pg.WithAuditedScopeTx`) | `go test ./internal/platform/pg/... -run TestWithTenantTxAndAuditedScopeEnforceRLSAgainstRealPostgres -v` com `R2_RUNTIME_TEST_DSN` apontando para o mesmo Postgres real | idem | `evidence/go-build-vet-pg.log` | PASS |
| Build/vet completos | `go build ./...`, `go vet ./...` | Go local do checkout | `evidence/go-build-vet-pg.log` | PASS, sem novos avisos |
| Suíte completa de `internal/platform/pg` | `go test ./internal/platform/pg/... -v` | idem | `evidence/go-build-vet-pg.log` | 3/3 PASS (2 pré-existentes + 1 novo) |

## Limites explícitos desta evidência
- Não cobre os outros 14 requisitos R6 nem R6-SEG-01-S02/S03 completos (pool de conexão única alternando tenants; administrador nominal fim-a-fim).
- Não inclui `go test -race` nem a suíte completa de `internal/orbita`/`cometa`/`libra`/`atlas`/`pulsar` — não alterados nesta sessão além de `pg.go`/`pg_test.go`, migrações e o script de prova; não há necessidade de rodar toda a suíte para código não tocado, mas nenhum resultado histórico foi copiado como se fosse desta sessão.
- Docker/kind independente, browser, carga, HA e restore: não exercidos nesta sessão (nenhum requisito desta sessão dependeu disso).
