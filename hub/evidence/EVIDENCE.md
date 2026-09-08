# Evidência de execução real — Hub de Interoperabilidade (fatia inicial)

**Data:** 6/7 de setembro de 2026. **Origem:** stack local completa
(`hub/deploy/docker-compose.yml`) efetivamente executada, com catálogo/produtos
configurados para provedores síncronos e assíncronos (polling e callback), tráfego real
gerado, e evidência capturada diretamente do banco de dados, das filas SQS/SNS, do S3,
dos logs de cada serviço e de capturas de tela do frontend e do Swagger. Nada aqui é
simulado ou reconstruído a posteriori — todo arquivo nesta pasta é a saída literal de um
comando executado contra a stack real.

**Estado corrente:** a importação HivePlace descrita ao final deste documento substitui
o catálogo de fixtures de teste. As seções de tráfego, filas e faturamento anteriores
continuam como evidência histórica da fatia inicial e não representam o catálogo
corrente do plano de controle.

Ver também `IMPLEMENTATION_AUDIT.md` (raiz do repositório) para a auditoria de
implementação e `openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md` para a
revisão requisito-a-requisito de todo o OpenSpec contra este código.

## Como reproduzir

```bash
cd hub/deploy && docker compose up -d --build   # sobe a stack; seed de fixtures é no-op
cd .. && deploy/import_hiveplace_collection.sh  # importa Hive: 34 produtos/232 endpoints
```

## Reexecução confirmada em 07/09/2026

A stack foi reconstruída no estado atual do `main` com `docker compose up -d --build`,
seguida de `go test -tags e2e ./test/e2e/...`, que passou, e
`go test -tags e2e ./internal/objectstore`, que também passou. O tráfego foi gerado
novamente por `generate_traffic.sh`; a saída desta execução está em
`traffic_output_current.txt`.

Os recortes atuais, capturados diretamente dos componentes, estão em:

- `db/hub_control_current.txt`, `db/hub_core_current.txt` e
  `db/hub_finance_current.txt` — catálogo, execução, entregas, outbox, inbox e
  fatos financeiros;
- `queues/s3_components_current.txt` — cinco filas SQS, dois tópicos SNS, bucket e
  metadados do objeto S3;
- `logs/current_stack.log` e `logs/current_trace_ids.txt` — logs dos cinco serviços e
  amostra dos `trace_id` propagados.

Nesta reexecução, o log capturado apresentou 59 mensagens DEBUG, 87 INFO e 1 WARN;
nenhum ERROR ocorreu no intervalo. O banco confirmou 11 operações `SUCCEEDED`, 4
`FAILED`, 14 entregas `DELIVERED`, 30 eventos de outbox publicados e `inbox_count=0`.
O resultado do S3 confirma o round-trip do objeto `evidence/roundtrip.json` com 76
bytes e `ContentType=application/json`.

As capturas das telas do admin-ui e do Swagger foram regeneradas pelo
`screenshots/shoot.mjs` contra a UI local e o container Swagger, incluindo todas as
cinco telas administrativas e as visões geral e expandida do endpoint.

O ensaio adicional multi-cliente/multi-provedor foi executado por
`multi_client_scenarios.sh`. Ele publicou duas APIs no catálogo (`consulta-cadastral-publica`
e `protocolo-publico`), criou contratos para `cliente-alpha`, `cliente-beta` e
`cliente-gamma`, consumiu provedores síncronos, polling e callback e registrou a saída
em `multi_client_scenarios_output.txt`. O relatório agregado de consumo, faturamento e
SLA está em `reports/multi_client_billing_sla_monitoring.txt`.

O cadastro de serviços agora rejeita código, versão, descrição, SLA, TTL ou modos
inválidos, com cobertura unitária em `internal/atlas/handlers_test.go`. A UI foi
exercitada de forma automatizada por `screenshots/validate.mjs`; todas as telas
existentes e o Swagger passaram. O endpoint `/metrics` passou a contar requisições por
método/status sem usar IDs de negócio como labels.

## Auditoria de logs e correção do PostgreSQL

Foi coletado o log de todos os containers. O padrão `FATAL: database "hub" does not
exist` teve origem no healthcheck `pg_isready -U hub`, que omitia o banco bootstrap.
O Compose foi corrigido para `pg_isready -U hub -d postgres`, o PostgreSQL foi recriado
sem remover o volume e os serviços foram reiniciados de forma ordenada. A evidência
completa está em `logs/all_containers_final.log` e
`reports/container_log_audit_2026-09-07.md`.

Durante a substituição forçada foram registrados quatro erros transitórios de
reconexão (DNS/connection refused) nos loops do Pulsar, Órbita e Cometa. Após a
estabilização, a janela limpa de 30 segundos teve zero `FATAL`, `ERROR`, `WARN`,
`panic` ou `exception`; todos os healthchecks responderam com sucesso e a nova rodada
de e2e/S3/multi-cliente passou.

## Configuração de catálogo (serviços e provedores síncronos/assíncronos)

`hub/deploy/seed/seed.sh` (executado contra a stack real, saída completa em
`traffic_output.txt` e nos logs) registrou, via API real do Atlas:

- **2 serviços/produtos**: `consulta-cadastral` v1 (SYNC+ASYNC+AUTO, SLA 30s) e
  `protocolo-assincrono` v1 (ASYNC+AUTO apenas, SLA 45s) — ver
  `db/hub_control.txt`.
- **8 contas de provedor**: `prov-sync-1` (síncrono), `prov-poll-1` e `prov-poll-2`
  (assíncronos por polling, dois provedores distintos), `prov-callback-1` (assíncrono
  por callback), `prov-nocred-evid` (sem credencial, para o cenário de recusa),
  `prov-basic-1`, `prov-oauth-1` e `prov-mtls-oauth-1` (autenticação de saída).
- **9 vínculos de credencial**: 8 `SHARED_HUB` + 1 `TENANT_DEDICATED`, incluindo o
  vínculo dedicado `acme-dedicated-oauth` para OAuth.
- **4 contratos de tenant**: `acme` (padrão), `acme-strict` (saldo estrito),
  `acme-dedicated` e `acme-dedicated-oauth` (credenciais dedicadas exigidas).
- **1 destino de webhook** (`acme` → `webhook-sink`).

## Tráfego gerado (`traffic_output.txt`)

10 cenários reais, cobrindo todos os modos e caminhos de decisão principais:

| # | Cenário | Resultado observado |
|---|---|---|
| 1 | SYNC sucesso | 200 SUCCEEDED |
| 2 | SYNC falha forçada | 200 FAILED |
| 3 | ASYNC polling, provedor A | 202 → SUCCEEDED (confirmado depois) |
| 4 | ASYNC polling, provedor B | 202 → SUCCEEDED |
| 5 | ASYNC callback | 202 → SUCCEEDED |
| 6 | AUTO | 202 → SUCCEEDED |
| 7 | Saldo estrito dentro do limite | 200 SUCCEEDED, reserva CAPTURED |
| 8 | Credencial dedicada (TENANT_DEDICATED) | 200 SUCCEEDED |
| 9 | Credencial indisponível (SEG-05) | 200 FAILED, sem fallback |
| 10 | Idempotência (repetição do #1) | Mesmo protocol_id, sem nova execução |

## Banco de dados (`db/`)

- `hub_control.txt` — catálogo, contas de provedor, credenciais, contratos reais.
- `hub_core.txt` — 9 protocolos (todos os modos), 9 operações, 8 tentativas físicas,
  agenda de polling **vazia** (tudo resolvido), 7 entregas de webhook `DELIVERED`,
  outbox com 17 eventos, todos `published = t`. Inclui a contagem da tabela `inbox`
  (0 — gap conhecido, ver auditoria).
- `hub_finance.txt` — 14 fatos econômicos (6 REVENUE + 8 COST) com dedup por
  `(protocol_id, kind, meter)`, 1 reserva `CAPTURED`.

## Filas SQS / tópicos SNS (`queues/`)

- `sqs_sns.txt` — as 5 filas e 2 tópicos reais, com as 4 assinaturas de fan-out
  exatamente como descrito em COM-01/ARQ-02.
- `queue_message_demo.txt` — **demonstração real de acúmulo e drenagem**: o Cometa foi
  parado (`docker stop`), uma admissão ASYNC foi feita, a fila `cometa-commands`
  mostrou 1 mensagem pendente por quase 2 minutos (corpo real capturado, incluindo o
  `trace_id` da requisição original), e ao religar o Cometa a mensagem foi consumida e a
  fila voltou a 0. Também registra um achado: como o protocolo já havia expirado por
  SLA quando o Cometa retomou, a chamada de resolução de credencial falhou por
  "context deadline exceeded" e foi rotulada como "credencial indisponível" — a causa
  raiz real é o deadline vencido, não a credencial; nenhum efeito inseguro ocorreu, mas
  o código de erro nesse caso específico é impreciso (registrado em
  `IMPLEMENTATION_AUDIT.md`).

## S3 (`s3/`)

DAD-05 (arquivos em S3) **não está integrado a nenhum fluxo real de admissão** — gap já
conhecido. Para não deixar isso sem evidência, um teste de integração real
(`hub/internal/objectstore/objectstore_test.go`, tag `e2e`) grava e lê de volta um
objeto no S3 emulado usando o próprio cliente Go do hub (`internal/objectstore`), não
apenas o AWS CLI:

- `objectstore_test_output.txt` — `go test -tags e2e` passando.
- `s3_evidence.txt` — bucket `hub-objects-test`, objeto `evidence/roundtrip.json` (76
  bytes) e seu conteúdo real, listados via `awslocal s3api`.

## Logs — INFO, DEBUG, WARN e ERROR (`logs/`)

Nesta sessão, `internal/platform/logging` ganhou nível configurável (`LOG_LEVEL=debug`,
já setado em todos os serviços no `docker-compose.yml`) e um `trace_id` gerado na
admissão e propagado por todo o ciclo (Órbita → comando/fila → Cometa → fato → Libra/
Pulsar), visível em todos os níveis de log abaixo:

| Arquivo | INFO | DEBUG | WARN | ERROR |
|---|---:|---:|---:|---:|
| `atlas_full.log` | ✅ | — | — | — |
| `orbita_full.log` | ✅ | ✅ | — | ✅ (Postgres parado de propósito) |
| `cometa_full.log` | ✅ | ✅ | ✅ (credencial indisponível) | ✅ |
| `pulsar_full.log` | ✅ | ✅ | — | ✅ |
| `libra_full.log` | ✅ | ✅ | — | — |

As linhas `ERROR` foram provocadas deliberadamente (parada temporária do PostgreSQL) e
confirmam que o sistema responde 503 sem vazar detalhe interno ao cliente, embora
registre o detalhe completo no log do servidor.

**Nota sobre "trace"**: não é OpenTelemetry/tracing distribuído com spans reais — é
correlação por `trace_id` propagado manualmente e logado em cada etapa. Documentado
assim no próprio código (`internal/platform/logging`), sem alegar mais do que isso.

## Capturas de tela (`screenshots/`)

Geradas via Playwright (Chromium do sistema) contra a stack real e o admin-ui em modo
dev, script em `shoot.mjs`:

- `admin-ui-00-inicial.png` — tela inicial.
- `admin-ui-01-servicos.png` — catálogo, com o produto `hiveplace-01-token-request` v1 real
  consultado ao vivo via API.
- `admin-ui-02-contas-provedor.png` — configuração de `provider-hiveplace-hml`, exibindo
  HivePlace, OAuth Client Credentials e TTL sem exibir segredo.
- `admin-ui-03-credenciais.png` — resolução real de credencial (SEG-05) para
  tenant `hiveplace-sandbox` / `provider-hiveplace-hml`, mostrando o vínculo `SHARED_HUB` resolvido (sem expor
  segredo — só `secret_ref`).
- `admin-ui-04-contratos.png` — contrato do tenant `hiveplace-sandbox`.
- `swagger-ui-01-overview.png` — Swagger UI real (container `swagger-ui`, porta 8092)
  servindo os dois documentos OpenAPI do hub.
- `swagger-ui-02-endpoint-expanded.png` — detalhe do endpoint `POST /v1/protocols`
  expandido, com schemas, respostas e descrições reais.

## Autenticação de provedores e cache Redis (07/09/2026)

Foi executado `auth_scenarios.sh` contra a stack local reconstruída. Todos os cinco
cenários retornaram HTTP 200 e `SUCCEEDED`:

- `BASIC` com `prov-basic-1` e credencial compartilhada;
- `OAUTH_FIRST` com Client Credentials, obtendo token no endpoint do provedor;
- `OAUTH_CACHE_HIT` com a mesma conta, reutilizando o token armazenado no Redis;
- `MTLS_OAUTH` com OAuth e referência de certificado validada pelo simulador local;
- `DEDICATED_OAUTH` com contrato `TENANT_DEDICATED` e referência dedicada.

Os protocolos e perfis podem ser reproduzidos em `auth_scenarios_output.txt` e
`db/provider_auth_profiles.txt`. O arquivo `db/redis_token_cache.txt` contém somente
nomes de chave e TTL, sem token ou segredo. A execução final observou duas chaves
`hub:provider-token:*` com TTL de aproximadamente 89 segundos; `DBSIZE=2`.

Os logs completos desta rodada estão em `logs/auth_integration_final.log`, com DEBUG
habilitado e sem `FATAL`, `ERROR` ou `WARN` na janela coletada. A janela limpa posterior
está em `logs/auth_clean_window.log`. O primeiro ensaio anterior registrou um 404/504
transitório porque a rota `/oauth/token` ainda não estava publicada pelo `provider-sim`;
isso foi corrigido, o serviço foi reconstruído e a rodada final passou.

Limite da evidência: `MTLS_OAUTH` demonstra a política e a associação da referência de
certificado no simulador, mas não substitui handshake TLS/mTLS nem uma PKI real. Também
não há exposição de segredo: os campos persistidos são apenas referências `vault://`.

## Importação HivePlace a partir da collection Postman

Em 07/09/2026, a collection `Hive` fornecida pelo usuário foi tratada como fonte de
metadados e importada pelo script `../deploy/import_hiveplace_collection.sh`. O resultado
foi uma conta `provider-hiveplace-hml` (`provider_id=HivePlace`, ambiente `HML`), 34
produtos e 232 endpoints no catálogo `provider_api_catalog`. A distribuição dos endpoints
por autenticação é: 150 `NONE`, 69 `BEARER`, 11 `API_KEY` e 2 `BASIC`.

Os valores do environment não foram persistidos. O banco guarda somente referências
`postman-Hive-HML-userName`, `postman-Hive-HML-userPass` e a referência agregada do
vínculo compartilhado. Os registros de teste do plano de controle foram removidos;
protocolos e operações históricas em `hub_core` foram preservados. A saída da consulta
está em `db/hiveplace_catalog_import.txt`.

O request principal de `hiveplace-06-antispoofing` foi corrigido para usar a imagem JPEG
Base64 válida do cenário `POST RAND` e Bearer `{{access_token}}`. A execução no sandbox
retornou HTTP 200, score 98 e status `Concluido`; saída em
`hiveplace_antispoofing_output.txt`.

## Achados desta sessão (novos, além dos 5 já registrados na auditoria anterior)

1. **Tabela `inbox` confirmada vazia** (0 linhas) — nenhum consumidor a usa; dedup
   depende de chaves específicas de cada domínio, não de um mecanismo genérico de
   inbox (COM-03 parcialmente satisfeito).
2. **Erro de credencial mal atribuído** quando um comando QUEUED é processado após o
   deadline do protocolo já ter vencido: a causa raiz é o `context.WithDeadline` já
   expirado, mas o erro é reportado como "credencial indisponível". Nenhum efeito
   inseguro resultou disso (nenhuma chamada ao provedor, nenhum sucesso tardio), mas a
   classificação do erro é imprecisa. Não corrigido nesta sessão — registrado para
   trabalho futuro.
