# Índice de evidências R3

| evidência | SHA | digest | comando | resultado |
|---|---|---|---|---|
| teste de precisão/schema | working tree sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | `28ab0779fbf85b0ecfc957720ae6453a65ef95ab51b9ae79baff3d919e019a2e` | `cd hub && go test ./internal/atlas -run TestTechnicalProjection` | PASS; não é qualificação integral |
| baseline Go | `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | histórico R3 | `cd hub && go test ./...` | PASS unitário; integrações condicionais pendentes |
| OpenSpec estrito | working tree sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | a regenerar antes de commit | `npx --yes @fission-ai/openspec@latest validate --all --strict` | PASS: 17 changes, 0 falhas |
| regressão callback | working tree sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | a regenerar antes de commit | `cd hub && go test ./internal/cometa` | PASS unitário; capability por operação aleatória e somente hash persistível |
| migrations isoladas | working tree anterior a `0031_callback_custody.sql` sobre `a4a876a9f8e875db882f7ca45cf7dece24d57aee` | laboratório Docker `ai_hub_r3qual` | `docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml run --rm migrate` | PASS até core `0030`; não equivale à qualificação de fluxos. A aplicação de `0031` permanece pendente enquanto o PostgreSQL isolado recupera o WAL. |

Não há credenciais reais, payloads sensíveis ou tokens neste índice.

## Evidências da continuação em 2026-09-09

| evidência | conteúdo testado | digest/comando | resultado |
|---|---|---|---|
| correção de finalização nullable | `hub/internal/orbita/finalize.go` com `COALESCE` do snapshot econômico ausente | `572a1b67b1fd3e63d53d5e1ba26da3c9dd55c2db78bfc46de0d5b192df62f77b` / `cd hub && go test ./internal/orbita`; `go vet ./internal/orbita` | PASS unitário/vet; não substitui integração |
| migração callback inbox | `core/0032_callback_inbox.sql` aplicado no `ai_hub_r3qual` | `221629aa1c49f2755139c537da6bd6bbb4ba20727f52455713d457a5b798d238` / consulta `schema_migrations` | PASS DDL idempotente e checksum registrado |
| regressão local da continuação | pacote Go completo, frontend e contrato OpenSpec | `cd hub && go test ./...`; `go vet ./...`; `cd hub/admin-ui && npm run build`; `npx --yes @fission-ai/openspec@latest validate --all --strict` | PASS; 17 changes OpenSpec, 0 falhas |
| integração PostgreSQL concorrente | `TestAdmissionPostgresAtomicIdempotency` com DSN local real | `R2_CORE_TEST_DSN=postgres://...@127.0.0.1:15432/hub_core go test ./internal/orbita -run TestAdmissionPostgresAtomicIdempotency -count=1 -v` | NÃO QUALIFICADO: bloqueio observado em `WALWrite/WALSync`; encerrado sem apagar volume |

| admissão concorrente reexecutada | PostgreSQL real, 24 contenders, idempotência e rollback de intenção | `go test ./internal/orbita -run TestAdmissionPostgresAtomicIdempotency -count=1 -v` | PASS em 20,07s |
| custódia Cometa | capacidade, submissão, observação, polling, fencing e escopo HTTP | `go test ./internal/cometa -run TestPostgres -count=1 -v` e reproduções isoladas | PASS nas subprovas observadas; uma execução agregada teve flake de fencing sob WAL |
| polling HTTP autenticado | provider HTTP sintético, binding revogado, Basic Auth e egress privado | `go test ./internal/cometa -run TestPostgresPollingAuthenticatedHTTP -count=1 -v` | PASS em 7,40s |
| schema aberto de runtime | projeção JSON com `type=object` sem propriedades | `go test ./internal/atlas -run 'TestTechnicalProjection|TestTechnicalProjectionAcceptsOpenObjectRuntimeSchema' -count=1 -v` | PASS |
| financeira PostgreSQL | cenários R2-FIN contra `hub_finance` | `go test ./internal/libra -run TestFinancePostgresScenarios -count=1 -v` | NÃO QUALIFICADO: timeout em `COMMIT`, wait `WALSync` |

| lease e Retry-After de polling | lease mínima operacional e agenda após confirmação durável | `go test ./internal/cometa -run TestPostgresPollingClaimsFenceAndAbsoluteDeadline -count=1 -v` | PASS isolado em 13,43s; fencing, takeover, deadline e atraso persistido verificados |
| race/vet após correções | código Go completo | `cd hub && go test -race ./...`; `go vet ./...` | PASS |
| gate agregado Cometa | todos os testes PostgreSQL do pacote | `go test ./internal/cometa -run TestPostgres -count=1` | NÃO CONCLUÍDO: execução presa em `WALSync` do laboratório e encerrada; subprovas isoladas permanecem evidenciadas |

| readiness do runtime | Atlas, Orbita, Cometa, Pulsar e Libra ativos | `curl --max-time 5 http://127.0.0.1:{18081,18080,18082,18083,18084}/healthz/ready` | PASS: HTTP 200 nas cinco capacidades; imagens em execução são históricas, não o diff atual |

| reconstrução pós-reboot | serviços Go e admin-ui construídos a partir do working tree atual | `docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml build --parallel atlas orbita cometa pulsar libra provider-sim webhook-sink admin-ui` | PASS; bases Go/distroless pinned obtidas e imagens exportadas |
| migração idempotente pós-reboot | control/core/finance, incluindo callback inbox | `docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml run --rm migrate` | PASS; versões existentes reconhecidas, sem alterar volumes |
| Kong após limitação de workers | gateway declarativo com um worker e rota protegida | `docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml up -d --no-build ...; curl -i http://127.0.0.1:18000/v1/protocols` | PASS; healthcheck saudável e HTTP 401 sem credencial |
| readiness com imagens atuais | Atlas, Orbita, Cometa, Pulsar e Libra reconstruídos | `curl --max-time 5 http://127.0.0.1:{18081,18080,18082,18083,18084}/healthz/ready` | PASS: HTTP 200 nas cinco capacidades; digests registrados no checkpoint |
| Órbita PostgreSQL pós-reboot | admissão atômica concorrente e inbox/custódia de fatos | `R2_CORE_TEST_DSN=postgres://...@127.0.0.1:15432/hub_core go test ./internal/orbita -run 'TestAdmissionPostgresAtomicIdempotency|TestOperationFactPostgresCustody' -count=1 -v` | PASS: 24 contenders, rollback de intenção, redelivery e quarentena |
| Cometa PostgreSQL pós-reboot | capacidade, custódia, escopo, fencing, polling e callback | `R2_CORE_TEST_DSN=postgres://...@127.0.0.1:15432/hub_core go test ./internal/cometa -run TestPostgres -count=1 -v` | PASS: todas as subprovas, 0,713s |
| Atlas PostgreSQL pós-reboot | concorrência/publicação/isolamento e paginação/staging | `ATLAS_TEST_DSN=postgres://...@127.0.0.1:15432/hub_control go test ./internal/atlas -run TestCatalogPostgres -count=1 -v` | PASS: ambos os cenários, 0,787s |
| Pulsar PostgreSQL pós-reboot | custódia versionada de webhook e recibo HTTP 204 | `R2_CORE_TEST_DSN=postgres://...@127.0.0.1:15432/hub_core go test ./internal/pulsar -run TestPostgres -count=1 -v` | PASS: bytes/hash, claims exclusivas e duas chaves por tenant |
| objetos PostgreSQL+LocalStack pós-reboot | multipart streaming, hash, autorização e retenção | `R2_CORE_TEST_DSN=postgres://...@127.0.0.1:15432/hub_core R2_S3_TEST_ENDPOINT=http://127.0.0.1:14566 go test ./internal/objectstore -run TestPostgresS3 -count=1 -v` | PASS: 16 MiB, 3 partes, pin e purge protegido, 0,617s |
| Libra PostgreSQL pós-reboot | financeiro completo R2-FIN-01 a R2-FIN-06 | `R2_FINANCE_DSN=postgres://...@127.0.0.1:15432/hub_finance go test ./internal/libra -run TestFinancePostgresScenarios -count=1 -v` | PASS: todos os subcenários, 0,550s; WALSync não reproduzido após reinício |
| cofre LocalStack pós-reboot | resolução SigV4, versão fixada e rejeição de versão ausente | `R2_TEST_SECRET_REF=r2/provider/fixture SECRETS_ENDPOINT=http://127.0.0.1:14566 go test ./internal/providerauth -run TestAWSVaultLocalStackVersion -count=1 -v` | PASS; fixture exclusivamente local e segredo omitido |
| smoke browser administrativo | carregamento real do bundle servido pelo container e presença da entrada de sessão | `node -e "...playwright-core... goto('http://127.0.0.1:13000/') ..."` | PASS parcial: título e botão Entrar; jornada autenticada com persistência ainda não qualificada |

### Digests da continuação atual

| arquivo | SHA-256 |
|---|---|
| `hub/internal/orbita/finalize.go` | `572a1b67b1fd3e63d53d5e1ba26da3c9dd55c2db78bfc46de0d5b192df62f77b` |
| `hub/internal/atlas/offers.go` | `5367032ee34c97cce9aa8f172835b074295b7c619af2593a9d9fc4c832d0a38f` |
| `hub/internal/atlas/catalog_test.go` | `864d4785bce7fd4b929f6bf4543f982e4b64b6afa4e0520a0b3c8d7de52f5b17` |
| `hub/internal/cometa/polling_custody.go` | `d47eecae1028d46ae436c06b31906a8c92cbc5d22854bbd2a7a7c9fdfe1010df` |

| bootstrap/identidade/cofre | `hub/deploy/r2/scripts/bootstrap.sh`, `hub/deploy/r2/scripts/reconcile_identity.py` | `0d388ac1f03bed2aceb36f2cc616ddeb1381dcee065144b1654337e6f2d3acf8`, `219bf38aafd2b350c7ac60ee3cdbcc7b6c3f0d21c83a13685191d0edf4426410` | bootstrap com `R2_COMPOSE_PROJECT`; reconciliação Keycloak; teste SigV4 versionado contra LocalStack | PASS |
| Kind real | runtime atual `ai-hub-r2` | `c4ee5ab7a95abf40f1a502fc438c95a7b2f019f16cb42d4a2293abc4cd64fb7b` | `R2_COMPOSE_PROJECT=ai_hub_r3qual bash hub/deploy/r2/kind/bootstrap.sh`; `kubectl get pods/top/hpa` | PASS; 3 nós, 5 pods `1/1`, métricas e HPA/KEDA |
| observabilidade real | Alloy/Docker→Loki, Prometheus, OTLP→Tempo, Grafana | `ee05b9f32dd5192d0d554ac970e2021fafb96251687232a5f4dbbeffd3f5017b` | queries Prometheus/Loki/Tempo e métricas Alloy | PASS; 9.181 entradas Docker, zero erros de envio |
| navegador autenticado | OIDC PKCE+MFA, save/reload/readback, storage e 390px | `bf09b657d7fe5cc817dba5018b1326bb8d0a98e93e8b650296b84261c835562d` | `PLAYWRIGHT_MODULE=.../playwright-core/index.js node hub/deploy/r2/tests/browser-smoke.mjs` | PASS; JSON e screenshot atualizados |
| integração de domínio | SUBMITTING/UNKNOWN, fencing, polling/callback, isolamento, catálogo, webhook, multipart e R2-FIN-01..06 | working tree atual | suíte com DSNs reais PostgreSQL/LocalStack | PASS; restore externo ainda não coberto |

| executor DAG | execução por camadas, mapeamento, limite de paralelismo e compensação | working tree atual | `cd hub && go test -race ./internal/atlas -run 'TestExecuteDAG' -count=1` | PASS unitário; ainda não conectado ao DAG persistido do produto |
| oráculo do provider-sim | deduplicação de submissão por protocolo e contagem externa de efeitos | working tree atual; imagem provider-sim reconstruída | duas chamadas a `/v1/operations` + `curl /__qualification/effects` | PASS sintético: um `provider_request_id`, um efeito; contador não persistido entre reinícios |
| RLS runtime negativo | role não proprietária `hub_runtime`, `FORCE ROW LEVEL SECURITY` e isolamento tenant A/B em control/core/finance | working tree atual; migrações 0033/0034 aplicadas | `bash hub/deploy/r2/tests/rls-runtime-proof.sh` | PASS: tenant A não lê linha de tenant B nos três bancos e flags `false:false:false`; adoção dos DSNs nos serviços permanece aberta |
| restore cercado | pg_dump/pg_restore de control/core/finance, contagens e cópia isolada de objetos | working tree atual; execução `autonomous2` | `bash hub/deploy/r2/tests/restore-reconciliation.sh` | PASS de integridade/count/no-replay com admissão desabilitada; não é ainda reconciliação populada de efeitos/financeiro |
| checksum de migração | divergência de checksum convertida em erro real do migrador | `hub/deploy/r2/scripts/migrate.sh` | `bash -n hub/deploy/r2/scripts/migrate.sh` + execução do migrador no Compose | PASS; versões 0033/0034 reconhecidas sem mascarar erro |
| 42 achados históricos | reavaliação individual F-01…F-42 com condição de encerramento | `REAVALIACAO-42-INDIVIDUAL.md` | inspeção documental e vínculo a evidências | 42 reavaliados individualmente; nenhum encerrado sem prova específica |
