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

## Continuação em 2026-09-09 — fatia de finalização e migração

- Estado de entrada: branch de continuidade `codex/r3-implementacao-integral`, criada a partir do merge `b9d0f90`; working tree estava limpo. O relatório histórico ainda referencia o SHA pré-merge e a branch antiga.
- Falha reproduzida no runtime: `orbita` repetia `sql.Scan error ... storing driver.Value type <nil> into type *json.RawMessage` no timer de deadline. A causa era a projeção nullable de `command->'economic_snapshot'` em `Finalizer`.
- Correção aplicada em `hub/internal/orbita/finalize.go`: `COALESCE(..., '{}'::jsonb)` antes do scan, mantendo fato final/outbox com JSON válido quando o comando não possui snapshot econômico. `gofmt`, `go test ./internal/orbita`, `go vet ./internal/orbita` e `git diff --check` passaram.
- Migração `core/0032_callback_inbox.sql`: DDL e índice já existiam no laboratório, mas a linha de controle estava ausente; a aplicação idempotente foi concluída com checksum `221629aa1c49f2755139c537da6bd6bbb4ba20727f52455713d457a5b798d238`. O `migrate` one-shot do Compose ficou preso no cliente Docker Desktop e foi encerrado; o container de execução não contém volume.
- Integração PostgreSQL: tentativa com `R2_CORE_TEST_DSN` alcançou a base, mas a concorrência de admissão ficou bloqueada em `WALWrite/WALSync` durante rollback/limpeza. Foi encerrada após janela controlada; não contar como PASS. Nenhum volume foi removido.
- Estado operacional ao checkpoint: `ai_hub_r3qual` continua com PostgreSQL, LocalStack, Keycloak, provider-sim e webhook-sink ativos; serviços de domínio permanecem parados da execução anterior; Kong está unhealthy. Compose/kind/Browser/observabilidade real continuam NÃO QUALIFICADOS.
- Próximo passo: reabrir a integração em uma janela PostgreSQL saudável, reconstruir imagens do working tree, validar callback/inbox e depois atacar recuperação SUBMITTING/UNKNOWN e fencing. A correção de `economic_snapshot` não encerra T-R2-01 nem transforma a qualificação integrada em PASS.

## Continuação em 2026-09-09 — integração PostgreSQL e polling

- A compilação local dos comandos Go (`CGO_ENABLED=0 go build ./cmd/orbita ./cmd/cometa ./cmd/atlas ./cmd/pulsar ./cmd/libra`) passou. O BuildKit/registry não concluiu a construção das imagens em 180s; o último ponto foi o pull da base pinned `golang:1.24-bookworm`. Não contar imagens Docker atuais como conteúdo do working tree.
- `TestAdmissionPostgresAtomicIdempotency` passou contra PostgreSQL real com 24 contenders concorrentes em 20,07s.
- Cometa passou em PostgreSQL real: capacidade agregada/adaptativa/rolling rate, custódia de submissão e observação, custódia pendente, escopo HTTP por tenant/application/cell e conflito polling/callback. O teste de fencing/absolute deadline passou isoladamente; a primeira execução foi flakey sob carga de WAL.
- `TestPostgresPollingAuthenticatedHTTP` passou após a correção de schema aberto: GET HTTP real com Basic Auth derivado do binding, token de workload no catálogo, egress pinado e revogação impedindo nova chamada.
- Corrigido `TransformJSON` para aceitar schema de runtime `{\"type\":\"object\"}` como objeto aberto quando `additionalProperties` não for falso; publicação continua exigindo propriedades tipadas. Regressão unitária adicionada em `hub/internal/atlas/catalog_test.go`.
- Atlas aprovou a prova de concorrência/publicação/isolamento. A prova de paginação staging ficou além da janela por `DataFileImmediateSync`; deve ser repetida em PostgreSQL sem saturação.
- Libra financeiro foi interrompido por timeout com backend em `WALSync` durante `COMMIT`; não é PASS nem falha funcional concluída.

## Continuação em 2026-09-09 — correções de polling

- `TransformJSON` com schema de objeto aberto foi corrigido e o fluxo HTTP de polling passou isoladamente.
- Lease de polling foi ampliada para no mínimo 30s, cobrindo preparação durável e resolução de binding/token antes do fence imediatamente anterior ao I/O; o deadline absoluto continua obrigatório.
- `Retry-After` agora recebe margem conservadora de durabilidade de 5s quando presente, evitando que a agenda persistida fique abaixo do mínimo do provedor após latência de commit/WAL.
- `TestPostgresPollingClaimsFenceAndAbsoluteDeadline` passou isoladamente em 13,43s após a correção; `TestPostgresPollingAuthenticatedHTTP` passou em 0,51s na execução combinada.
- `go test -race ./...` e `go vet ./...` passaram. A execução agregada de todos os testes PostgreSQL do Cometa foi encerrada após ficar presa em `WALSync`; não contar como PASS agregado.

## Continuação em 2026-09-09 — runtime restaurado

- O ecossistema `ai_hub_r3qual` voltou a executar Atlas, Orbita, Cometa, Pulsar e Libra com readiness HTTP `200` em todas as cinco portas; PostgreSQL, LocalStack, Keycloak, provider-sim e webhook-sink permanecem ativos.
- Foram removidos somente os cinco containers duplicados em estado `Created/Dead` criados pela tentativa desta sessão; nenhum volume foi removido.
- As imagens em execução continuam sendo as imagens Docker anteriores (`created` antes desta continuação), pois o pull/build da base Go pinned não terminou. Readiness positivo não é prova de execução do diff atual.

## Continuação após reinício — reconstrução do laboratório

- O Docker Desktop foi reiniciado durante a reconstrução; os containers da validação encerraram com `Exited (255)`, sem remoção de volumes.
- As bases pinned foram obtidas e as imagens dos serviços foram reconstruídas a partir do working tree atual: `docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml build --parallel atlas orbita cometa pulsar libra provider-sim webhook-sink admin-ui` — PASS.
- O bootstrap idempotente foi executado novamente. Todas as migrações de `hub_control`, `hub_core` (incluindo `0032_callback_inbox`) e `hub_finance` foram reconhecidas como aplicadas; volumes existentes foram reutilizados.
- Digests executados: Atlas `sha256:21a2986c83d9f62420219ed650d47c7012290c0a0cb17ded86387db378807c5b`; Órbita `sha256:d628569139c0b0a41b97cb8f68c9468a1a61aa9c001d2b59ce5f972008bdfa79`; Cometa `sha256:08334b546f2d38db49e6f2fdb462a6adb258eba9e091f3b750b4f74b87a45c14`; Pulsar `sha256:b4da0adff5ad8c38a25a0b67a0ba4987a5382be1192f8b792a60adafe2efaa8b`; Libra `sha256:f45fe00aa5998af35c5914fc5b132ea78ecf8b49c5e59729f56933f5a263305b`.
- Corrigido o bootstrap local do Kong com `KONG_NGINX_WORKER_PROCESSES=1`; após recriação limpa, o healthcheck passou e `/v1/protocols` pelo gateway retornou `401` sem credencial, confirmando a rota protegida.
- Readiness após reconstrução: Atlas, Órbita, Cometa, Pulsar e Libra retornaram HTTP 200. A qualificação financeira PostgreSQL continua aberta por bloqueio `WALSync` reproduzido no cenário concorrente; não foi contado como PASS.
- Após o reboot e a reconstrução, a integração PostgreSQL passou: Órbita (admissão concorrente e custódia de fatos), Cometa (todos os `TestPostgres`), Atlas (publicação/paginação/staging), Pulsar (custódia de webhook), objetos (multipart S3/LocalStack) e Libra (todos os cenários financeiros R2-FIN-01 a R2-FIN-06). O bloqueio `WALSync` não foi reproduzido nesta janela limpa.
- A fixture sintética de segredo foi criada somente no LocalStack com referência `r2/provider/fixture`; `TestAWSVaultLocalStackVersion` passou com SigV4, pin de versão e rejeição de versão inexistente. O valor não foi registrado na evidência.
- Smoke de navegador com Chrome local e `playwright-core` contra `http://127.0.0.1:13000/`: título administrativo carregado e botão `Entrar` presente. A jornada autenticada de salvar/recarregar ainda não está qualificada porque exige uma identidade nominal/MFA de fixture; não marcar como concluída por esse smoke.

## Continuação — OIDC e idempotência do console

- O cliente OIDC do console deixou de solicitar o escopo opcional `profile`, que não é publicado pelo realm de laboratório; `openid` continua suficiente para o contrato administrativo. O admin-ui foi recompilado e recriado a partir do working tree.
- O `realm.json` passou a declarar de forma reproduzível o User Profile para `tenant_id`, `application_id` e `cell_id`, com edição restrita a administradores. A instância existente foi ajustada via Admin REST sem remover volume; claims sintéticas foram verificadas em token local sem registrar segredo.
- `command()` do console agora reutiliza a mesma chave de idempotência ao repetir uma tentativa transitória (rede/5xx), evitando que um retry da mesma intenção gere um segundo efeito. Erros 4xx não são repetidos.
- OpenSpec estrito foi executado novamente: 17 mudanças aprovadas, 0 falhas. `git diff --check`, `npm run build`, `go test ./...` e JSON do realm passaram.
- A reconstrução de uma fixture nominal após reinício mostrou que credenciais/atributos importados no Keycloak existente não são atualizados por `--import-realm` quando o realm já existe. O caso está isolado como problema de bootstrap/reconciliação da fixture; ainda não promover a jornada autenticada de salvar/recarregar a PASS.

## Continuação — gates locais executados após o reinício

- O bootstrap local agora aceita `R2_COMPOSE_PROJECT`, usa a rede/nome do projeto dinamicamente, executa `identity-reconcile` antes das APIs e semeia idempotentemente o segredo sintético `r2/provider/fixture` no LocalStack. Nenhum volume foi removido.
- `reconcile_identity.py` foi executado contra o realm já persistente e retornou `identity reconciliation: PASS`; o mapper `sub`, User Profile, senha, OTP, atributos de tenant/aplicação/célula e papéis das quatro fixtures foram reconciliados.
- Jornada Chromium real: Authorization Code PKCE + senha + OTP, criação autenticada de cliente, reload, novo login com OTP de janela posterior, leitura durável do mesmo recurso, navegação de provedores, ausência de token em storage e viewport 390px sem overflow — PASS. Recurso sintético: `browser-client-1788967926311`.
- Kind real `ai-hub-r2`: três nós, cinco deployments `1/1 Running`, métricas de nós disponíveis; HPA com alvos CPU reais e KEDA/Pulsar pronto.
- Observabilidade real: Prometheus `up=1`, Loki com 9.181 entradas via `loki.source.docker` sem erros de parsing/entrega, Tempo com traces dos serviços e Grafana/Alloy/Loki/Prometheus/Tempo saudáveis.
- Integrações reais pós-reboot: Órbita, Cometa, Atlas, Pulsar, objetos PostgreSQL+LocalStack, Libra R2-FIN-01..06 e cofre SigV4 com versão fixada passaram; reinicialização controlada da Órbita preservou readiness 200 nas cinco APIs.
- Gates ainda abertos: restore reconciliado com efeitos externos/financeiro, RLS com credencial não proprietária, executor DAG/composição além do planejador, fluxos administrativos restantes, reavaliação individual dos 42 achados históricos, ensaios completos de falha/restore e decisões externas D-01…D-07/T-R2-01.

## Continuação autônoma — custódia, isolamento e rastreabilidade

- Foi implementado `ExecuteDAG` em `hub/internal/atlas/executor.go`, com camadas do planejador, limite de paralelismo, mapeamento de entradas, modo parcial e compensação reversa. `go test -race ./internal/atlas` passou. A API ainda não está conectada a um produto/DAG persistido de produção; portanto o gate de composição permanece aberto.
- O provider-sim ganhou deduplicação por `protocol_id` e o oráculo `/__qualification/effects`. Duas submissões do mesmo protocolo retornaram o mesmo `provider_request_id` e produziram um único efeito no ensaio sintético. O contador é em memória e não prova recuperação do provedor após reinício.
- Foram adicionadas as migrações `0033_runtime_rls` e `0034_runtime_rls_migration_compat` para control/core/finance, com a role `hub_runtime` sem privilégios de proprietário e `FORCE ROW LEVEL SECURITY`. A prova negativa passou nos três bancos, incluindo verificação de `rolsuper=false`, `rolcreaterole=false` e `rolbypassrls=false`; os serviços legados ainda usam a compatibilidade do usuário `hub`, então a adoção integral do DSN runtime continua aberta.
- `restore-reconciliation.sh` executou dump/restore em bancos isolados, comparou contagens de tabelas e objetos e observou o oráculo externo sem re-admissão (`RESTORE_RECONCILIATION=PASS`, sufixo `autonomous2`). Esse é um ensaio de restore cercado e sem replay; ainda falta um cenário populado que reconcilie efeitos externos e obrigações financeiras após restore.
- `migrate.sh` passou a falhar de fato em divergência de checksum; o fluxo anterior ignorava o código de `\quit 3`. A correção foi validada com migrações 0033/0034 aplicadas no ecossistema oficial.
- A reavaliação individual dos 42 achados está em `REAVALIACAO-42-INDIVIDUAL.md`. Nenhum achado foi encerrado por inferência; cada item possui estado, evidência atual e condição objetiva de encerramento.

### Próxima retomada

Prioridade: conectar o executor a um DAG persistido e concluir o cenário de restore populado com oráculos independentes; depois substituir a compatibilidade RLS pela credencial de runtime nos serviços, executar a matriz de falhas/reinícios e tratar os fluxos administrativos/comerciais restantes. D-01…D-07 e T-R2-01 permanecem gates externos/decisórios, sem bloquear a implementação técnica independente.

## Baseline de decisões adotada — 2026-09-09

- D-01…D-07 receberam defaults executáveis para laboratório/homologação em `05-DECISOES-E-MIGRACAO.md`: envelope sintético, SLA e retry, unidade financeira, retenção, plataforma, capacidade de provedor e catálogo versionado.
- T-R2-01 recebeu a regra operacional conservadora: só anunciar sucesso após confirmação durável; em dúvida, preservar `UNKNOWN` e reconciliar; resultado tardio não reabre protocolo. A prova formal da fronteira relógio/commit continua obrigatória.
- As decisões habilitam a execução técnica, mas não substituem ratificação nominal de produção, contrato comercial, orçamento, retenção regulatória, matriz de capacidades de provedor ou aceite de RPO/RTO.

## Continuação — worker durável de retenção

- `Catalog.RunPurgeBatch` agora executa o plano persistido e tenta cada candidato individualmente; falha no S3 deixa o item em `PURGING` para retomada posterior e não impede os demais candidatos.
- `RunRetentionWorker` foi conectado ao processo Órbita e executa imediatamente e em intervalo configurável. O escopo é somente a lista explícita `RETENTION_TENANTS`; configuração vazia desabilita o expurgo e não existe wildcard.
- `TenantsFromEnv` remove espaços, entradas vazias e duplicatas; a regressão unitária passou. O worker preserva obrigações por meio dos pins e do estado durável existente.

## Continuação — oráculo externo persistente

- O provider-sim passou a persistir `protocol_id`, `provider_request_id`, operações e contador de efeitos em `/var/lib/provider-sim/state.json`, com gravação atômica e volume exclusivo `provider-sim-data`.
- Ensaio real: submissão sintética retornou `prov-req-000001` e `effects=1`; após `docker compose restart provider-sim`, a mesma submissão retornou o mesmo identificador e `effects=1, protocols=1`. Isso comprova não reexecução do simulador após reinício.

## Continuação — horizonte absoluto de retry

- `command_intents` passou a persistir `retry_started_at` e `retry_until` pela migração `core/0035_retry_horizon.sql`.
- `dispatch.Command` carrega `retry_ttl_seconds`; a primeira falha transitória inicia a janela uma única vez. Tentativas posteriores não renovam o horizonte; TTL zero encerra a intenção como `EXPIRED` para reconciliação.
- O teste PostgreSQL `TestPostgresRetryHorizonStartsOnceAndExpires` passou com a instância real: primeira falha, takeover após lease, backoff elegível, expiração e estado final `EXPIRED`.
- A imagem atual de Órbita foi reconstruída e recriada no Compose oficial; `/healthz/ready` retornou `200`.

## Continuação — conflito de observações externas

- Callback/polling que chegam após um estado terminal e divergem do fato vencedor agora recebem fonte `CALLBACK_CONFLICT` ou `POLL_AFTER_FINAL_CONFLICT` em `operation_receipts`.
- A evidência conflitante é conservada, mas não reabre a operação, não substitui o resultado durável e não cria novo outbox/efeito.
- `TestPostgresPollingCallbackConflictRetainsBoth` passou no PostgreSQL real: duas observações, exatamente um fato terminal e uma evidência explicitamente conflitante.
