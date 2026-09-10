# Checkpoint R4

Data: 2026-09-10 (America/Sao_Paulo)

HEAD: `b580022057f2dc2279b8ce00d97ac5a0cb479515`

Hash SHA-256 do diff rastreado no checkpoint: `361273a9bc987914a16911acb7902a0be2f2002bd1ff98fbae1929df13fb02dd` (deve ser recalculado após qualquer edição rastreada).

Executado: `go test ./...` em `hub` — PASS; `sh -n hub/deploy/r2/scripts/migrate.sh` — PASS; `git diff --check` — PASS. `docker compose ls` não reportou projetos; o cluster kind `ai-hub-r2-*` estava ativo e containers `ai_hub_r3qual-*` estavam parados. Nenhum recurso foi alterado.

Retomar por: qualificação PostgreSQL do worker/resultado/ofertas; ensaio PostgreSQL restore; console, browser, kind, carga e gate integral.
