# Checkpoint R4

Data: 2026-09-10 (America/Sao_Paulo)

HEAD: `7cd4a39fdbf4375cc0f246305c3220ec0b2bd9a3`

Hash SHA-256 do diff rastreado no checkpoint: `361273a9bc987914a16911acb7902a0be2f2002bd1ff98fbae1929df13fb02dd` (deve ser recalculado após qualquer edição rastreada).

Executado: OpenSpec strict `21/21` — PASS; `npm run build` no portal — PASS; `go test ./...`, `go test -race ./...`, `git diff --check` e validações de scripts — PASS. Compose `ai_hub_r3qual` foi qualificado com bootstrap, probes, migrações, restore e testes integrados; o restore passou com digests dos três bancos e S3 consistentes. O cluster kind existente chegou ao runtime, mas apresentou DNS/rede residual e workloads instáveis; browser posterior falhou por timeout; carga sem token retornou 401. Nenhum desses gates foi promovido como PASS integral.

Retomar por: confirmação para recriar somente o cluster kind residual; browser autenticado; carga com token de fixture; cenários externos de callback/inbox, revogação/crescimento de cache, ofertas em escala e regeneração integral da matriz.
