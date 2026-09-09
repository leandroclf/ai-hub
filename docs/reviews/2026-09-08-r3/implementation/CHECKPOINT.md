# Checkpoint R3

- Último passo validado: `TransformJSON` preserva `9007199254740993`, rejeita `enum`/campos extras aninhados; API Key e autorização administrativa estão conectadas. O callback passou a exigir uma capability aleatória por operação, persistida somente como SHA-256, com método/corpo limitados e 2xx exclusivamente após custódia durável.
- Testes: `go test ./...`, `go vet ./...`, `npm run build`, `git diff --check` e `npx --yes @fission-ai/openspec@latest validate --all --strict` — PASS (17 changes OpenSpec).
- Ambiente: branch local `codex/r3-implementacao-integral`; laboratório Compose isolado `ai_hub_r3qual`, com PostgreSQL/LocalStack/Keycloak e migrações aplicadas. A pilha foi construída a partir do working tree.
- Próximo passo: concluir a política de callback por conta (assinatura/mTLS/token homologado) e inbox órfã; revalidar recuperação SUBMITTING/UNKNOWN/fencing e o harness PostgreSQL/OIDC; depois executar os testes condicionais com dependências reais.
- Pendências: 189 requisitos, cenários v4/R2/R3, 25 achados R3, 42 históricos e qualificação operacional. O ensaio PostgreSQL concorrente foi interrompido por bloqueio de commits/WAL no laboratório Docker; não foi contado como PASS.

## Continuação em 2026-09-09

- Implementada a inbox durável de callbacks órfãos (`callback_inbox`), com corpo limitado pelo handler, hash do corpo, hash da capability, deduplicação, disposição `RECEIVED/APPLIED/REJECTED` e método de reconciliação posterior.
- Callback conhecido continua exigindo capability por operação antes da aplicação; callback de operação ainda não encontrada é conservado como obrigação recuperável e respondido com `202` somente após persistência. IDs inválidos são rejeitados.
- Bootstrap inicial do Compose passou a aplicar todas as migrações ordenadas de control/core/finance, incluindo `0032_callback_inbox.sql`.
- Validação desta fatia: `go test ./...`, `go vet ./...`, `bash -n hub/deploy/postgres-init/01-init.sh`, `git diff --check`, `docker compose config --quiet` e OpenSpec strict (17/17) passaram. Integração PostgreSQL ainda em execução nesta retomada.
- Runtime: tentativas de subir projeto Compose exclusivo (`ai_hub_r3_inbox`) e container PostgreSQL exclusivo (`ai_hub_r3_inbox_pg2`) ficaram bloqueadas durante pull/criação no Docker Desktop; não houve container persistido nem alteração em volumes existentes. PostgreSQL integrado permanece NÃO EXECUTADO.
