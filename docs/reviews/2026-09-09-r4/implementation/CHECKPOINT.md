# Checkpoint R4

Data: 2026-09-10 (America/Sao_Paulo)

HEAD: `6c22a562a7d71519205ab17967c57ad3690a6351`

Hash SHA-256 do diff rastreado no checkpoint: `b12d1f810da91e179ebe1ef522d749e083bba154ee1f03b04749da063f8ca719` (deve ser recalculado após qualquer edição rastreada).

Executado: `go test ./...` em `hub` — PASS; `sh -n hub/deploy/r2/scripts/migrate.sh` — PASS; `git diff --check` — PASS. `docker compose ls` não reportou projetos; o cluster kind `ai-hub-r2-*` estava ativo e containers `ai_hub_r3qual-*` estavam parados. Nenhum recurso foi alterado.

Retomar por: teste L1 com cofre indisponível; qualificação PostgreSQL do worker/resultado; índice seletivo do Atlas; ensaio PostgreSQL upgrade/restore; console e gate integral.
