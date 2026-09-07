# Evidência de execução real — Hub de Interoperabilidade (fatia inicial)

**Data:** 6/7 de setembro de 2026. **Origem:** stack local completa
(`hub/deploy/docker-compose.yml`) efetivamente executada, com catálogo/produtos
configurados para provedores síncronos e assíncronos (polling e callback), tráfego real
gerado, e evidência capturada diretamente do banco de dados, das filas SQS/SNS, do S3,
dos logs de cada serviço e de capturas de tela do frontend e do Swagger. Nada aqui é
simulado ou reconstruído a posteriori — todo arquivo nesta pasta é a saída literal de um
comando executado contra a stack real.

Ver também `IMPLEMENTATION_AUDIT.md` (raiz do repositório) para a auditoria de
implementação e `openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md` para a
revisão requisito-a-requisito de todo o OpenSpec contra este código.

## Como reproduzir

```bash
cd hub/deploy && docker compose up -d --build   # sobe a stack e roda o seed
cd ../evidence && ./generate_traffic.sh          # gera os 10 cenarios abaixo
```

## Configuração de catálogo (serviços e provedores síncronos/assíncronos)

`hub/deploy/seed/seed.sh` (executado contra a stack real, saída completa em
`traffic_output.txt` e nos logs) registrou, via API real do Atlas:

- **2 serviços/produtos**: `consulta-cadastral` v1 (SYNC+ASYNC+AUTO, SLA 30s) e
  `protocolo-assincrono` v1 (ASYNC+AUTO apenas, SLA 45s) — ver
  `db/hub_control.txt`.
- **5 contas de provedor**: `prov-sync-1` (síncrono), `prov-poll-1` e `prov-poll-2`
  (assíncronos por polling, dois provedores distintos), `prov-callback-1` (assíncrono
  por callback), `prov-nocred-evid` (sem credencial, para o cenário de recusa).
- **5 vínculos de credencial**: 4 `SHARED_HUB` + 1 `TENANT_DEDICATED`.
- **3 contratos de tenant**: `acme` (padrão), `acme-strict` (saldo estrito), 
  `acme-dedicated` (credencial dedicada exigida).
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
- `admin-ui-01-servicos.png` — catálogo, com o serviço `consulta-cadastral` v1 real
  consultado ao vivo via API.
- `admin-ui-02-contas-provedor.png` — busca de `prov-sync-1`.
- `admin-ui-03-credenciais.png` — resolução real de credencial (SEG-05) para
  tenant `acme` / `prov-sync-1`, mostrando o vínculo `SHARED_HUB` resolvido (sem expor
  segredo — só `secret_ref`).
- `admin-ui-04-contratos.png` — contrato do tenant `acme`.
- `swagger-ui-01-overview.png` — Swagger UI real (container `swagger-ui`, porta 8092)
  servindo os dois documentos OpenAPI do hub.
- `swagger-ui-02-endpoint-expanded.png` — detalhe do endpoint `POST /v1/protocols`
  expandido, com schemas, respostas e descrições reais.

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
