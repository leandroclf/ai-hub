# Índice de evidências

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

Nenhuma evidência histórica foi promovida como PASS de integração.
