# Checkpoint R4

Data: 2026-09-10 (America/Sao_Paulo)

HEAD da documentação deste checkpoint: será atualizado no commit de consolidação.

Hash SHA-256 do diff rastreado no checkpoint: `361273a9bc987914a16911acb7902a0be2f2002bd1ff98fbae1929df13fb02dd` (deve ser recalculado após qualquer edição rastreada).

Executado: OpenSpec strict `21/21` — PASS; `npm run build` no portal — PASS; `go test -race ./...`, `git diff --check` e validações de scripts — PASS. Compose `ai_hub_r3qual` foi qualificado com bootstrap, probes, migrações, restore e testes integrados; o restore passou com digests dos três bancos e S3 consistentes. O cluster kind foi recriado e passou readiness, métricas/HPA e recuperação de pod. O browser Chromium autenticado passou. A carga autenticada alcançou os serviços, mas está bloqueada por ausência de catálogo publicado; o Browser Harness opcional não está instalado.

Retomar por: seed/versionamento de catálogo via `admin/v1`; repetição da carga funcional; cenários externos de callback/inbox, revogação/crescimento de cache, ofertas em escala, HA e regeneração integral da matriz.
