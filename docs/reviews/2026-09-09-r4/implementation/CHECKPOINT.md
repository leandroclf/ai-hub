# Checkpoint R4

Data: 2026-09-10 (America/Sao_Paulo)

HEAD: `d93d86b8ba81b0e28bb852dba769966c87162c44`

Hash SHA-256 do diff rastreado no checkpoint: `47bd12ab65fb1863eeb83b2edd3e3fceeffc7bf18ea8727e94c2964a0b610657` (deve ser recalculado após qualquer edição rastreada).

Executado: `go test ./...` em `hub` — PASS; `sh -n hub/deploy/r2/scripts/migrate.sh` — PASS; `git diff --check` — PASS. `docker compose ls` não reportou projetos; o cluster kind `ai-hub-r2-*` estava ativo e containers `ai_hub_r3qual-*` estavam parados. Nenhum recurso foi alterado.

Retomar por: teste L1 com cofre indisponível; worker de inbox com claim/lease; validação de resultado compartilhada; índice seletivo do Atlas; ensaio PostgreSQL upgrade/restore; console e gate integral.
