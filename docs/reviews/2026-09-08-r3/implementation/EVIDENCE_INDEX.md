# Índice de evidências R3

| evidência | SHA | digest | comando | resultado |
|---|---|---|---|---|
| teste de precisão/schema | working tree sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | `28ab0779fbf85b0ecfc957720ae6453a65ef95ab51b9ae79baff3d919e019a2e` | `cd hub && go test ./internal/atlas -run TestTechnicalProjection` | PASS; não é qualificação integral |
| baseline Go | `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | histórico R3 | `cd hub && go test ./...` | PASS unitário; integrações condicionais pendentes |
| OpenSpec estrito | working tree sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | a regenerar antes de commit | `npx --yes @fission-ai/openspec@latest validate --all --strict` | PASS: 17 changes, 0 falhas |
| regressão callback | working tree sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | a regenerar antes de commit | `cd hub && go test ./internal/cometa` | PASS unitário; capability por operação aleatória e somente hash persistível |
| migrations isoladas | working tree anterior a `0031_callback_custody.sql` sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | laboratório Docker `ai_hub_r3qual` | `docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml run --rm migrate` | PASS até core `0030`; não equivale à qualificação de fluxos. A aplicação de `0031` permanece pendente enquanto o PostgreSQL isolado recupera o WAL. |

Não há credenciais reais, payloads sensíveis ou tokens neste índice.
