# Evidência de execução integrada R4 — 10/09/2026

Rodada realizada no ambiente local `ai_hub_r3qual`, com dados sintéticos,
branch `main` e o único ecossistema Compose do AI Hub ativo. Senhas, OTPs,
bearers, cookies e chaves privadas não fazem parte desta evidência.

## Gates executados

| Gate | Comando/artefato | Resultado |
|---|---|---|
| Catálogo versionado | `node hub/deploy/r2/tests/catalog-seed.mjs` | PASS; seed idempotente via `admin/v1`, validação/publicação com `If-Match`, qualificação/célula sintéticas somente no fixture autorizado |
| Carga autorizada | `R2_COMPOSE_PROJECT=ai_hub_r3qual node hub/deploy/r2/tests/authorized-load.mjs` | PASS; oito caminhos, idempotência e log saneado em `hub/evidence/r2/execution/authorized-load-latest.log` (prefixo `r4-authorized-1789123114890`) |
| SLO de referência | `R2_SLO_SAMPLES=30 R2_SLO_CONCURRENCY=5 node hub/deploy/r2/tests/slo-load-proof.mjs` | PASS local; 30 admissões ASYNC e 30 GETs, payload de 59989 bytes, admissão p95=55,07 ms/p99=57,06 ms e GET p95=5,16 ms/p99=5,23 ms; artefato saneado em `hub/evidence/r2/execution/slo-load-latest.json` |
| Broker fora / reinício | `node hub/deploy/r2/tests/broker-outage-runtime-proof.mjs` | PASS local; `localstack` interrompido, Órbita/Cometa recriados sem dependências, readiness preservado, SYNC e GET `SUCCEEDED`, persistência independente confirmada e broker restaurado; reconciliação da fixture sem efeito em `broker-outage-reconciliation-latest.log` |
| Ambientes e promoção | `bash hub/deploy/r2/tests/environment-boundary-proof.sh` | PASS estrutural; cinco overlays em namespaces distintos, referências de workload/runtime sem valores de segredo, gate `prd` bloqueado sem perfil/aprovações e desenvolvimento permitido; isolamento cross-environment com IdP real permanece parcial |
| Dependência opcional | `bash hub/deploy/r2/tests/optional-dependency-runtime-proof.sh` | PASS local; Alloy parado, readiness/liveness de Órbita/Cometa preservadas e Atlas/Pulsar/Libra/provider-sim/sink sem alteração de container ou reinício; Alloy restaurado no cleanup |
| Provedor fora | `node hub/deploy/r2/tests/provider-outage-runtime-proof.mjs` | PASS local; provider-sim parado, SYNC respondeu 504 sem sucesso fictício, GET preservou o protocolo, capacidade ficou `transport_open=false/pending_external=true`, workloads não afetados mantiveram containers e provider foi reconciliado/restaurado |
| Callback | Carga acima + testes focados de inbox/validação terminal | PASS parcial; callback concluiu, ingress key/capability inválidos retornaram 401, deduplicação por capability, quota e retenção têm implementação e testes focados; cenários externos completos ainda abertos |
| Portal Playwright | `node hub/deploy/r2/tests/browser-smoke.mjs` | PASS; OIDC PKCE, senha+OTP, erro de JSON inválido sem falso sucesso, editor de contrato REST declarativo, filtro local de referências, criação/leitura durável, editor de produto com mapeamento, SLA, redelivery administrativo com recibo/auditoria durável, reconciliação segura sem correlação externa, destinos versionados, preparação de ajuste financeiro com bloqueio de autoaprovação, navegação, logout, HTTP 401 pós-logout sem credencial, storage sem sessão persistida, viewport móvel sem overflow e `tenant_reader` sem formulário de escrita |
| Entrega webhook com capacidade | `bash hub/deploy/r2/tests/webhook-capacity-runtime.sh` | PASS; worker Pulsar real entregou no sink local e encerrou o permit `r4-webhook` sem concessão aberta ou obrigação pendente; fixture removida ao final |
| Incidência financeira integrada | `bash hub/deploy/r2/tests/finance-runtime-proof.sh` após carga autorizada | PASS; `SUBMITTED`/`STATUS` com `attempt_id`, 12 fatos (`cost=4`, `revenue=8`), inbox=26, journal balanceado por moeda e zero quarentena de incidência inválida; log em `hub/evidence/r2/execution/finance-runtime-latest.log` |
| Browser Harness | `BU_CDP_URL=http://127.0.0.1:9222 hub/deploy/r2/tests/browser-harness/run.sh` | PASS-EXPLORATORY; 16 rotas em viewport 390×844; descoberta automática do daemon headless requer CDP explícito |
| Cache de autenticação | `go test -race -count=1 ./internal/providerauth` | PASS; isolamento por binding, revogação, expiração, coordenação cancelável e máximo de 1024 locks rastreados |
| L1 durante falha do cofre | `hub/evidence/r2/execution/r4-ope-01-latest.log` | PASS; token válido reutilizado em L1 sem nova consulta ao resolver, espera cancelada por contexto e rotação limitada a 1.024 locks |
| Redis opcional | Compose profile `cache`, `redis-cli ping` e `INFO keyspace` | PASS; `PONG`, keyspace vazio; Cometa permanece sem dependência autoritativa de Redis e sem token persistido |
| Ofertas | consulta `LIMIT 2` + `EXPLAIN (ANALYZE, BUFFERS)` | PASS de caminho; índice parcial `catalog_offer_eligibility_lookup` confirmado quando o planner usa índice; catálogo pequeno pode escolher seq scan por custo |
| Restore | `R2_COMPOSE_PROJECT=ai_hub_r3qual R2_RESTORE_SUFFIX=r4_20260911qual2 bash hub/deploy/r2/tests/restore-reconciliation.sh` | PASS; contagem/digest de control, core, finance e objetos S3 iguais; efeitos observados sem replay e re-admissão mantida desabilitada; alvos já existentes agora são recusados antes do restore |
| Kind/HA | `hub/deploy/r2/kind/bootstrap-independent.sh` + `continuity-runtime-proof.sh` | PASS local independente; PostgreSQL, LocalStack, Keycloak, Alloy, provider-sim, webhook-sink, observabilidade, Kong e UI materializados no cluster; cinco workloads em duas réplicas; Cometa/Pulsar recuperados após exclusão controlada com RTO observado de 4,453 ms/4,481 ms |
| Produto/DAG durável | `R2_CORE_TEST_DSN=... go test -race -count=1 ./internal/orbita -run 'TestProductPlan'` | PASS; etapas e dependências persistidas, limite de paralelismo, consolidação por fatos e compensação separada; evidência em `hub/evidence/r2/execution/product-dag-latest.log` |
| Produto/DAG via HTTP | `R2_COMPOSE_PROJECT=ai_hub_r3qual node hub/deploy/r2/tests/product-runtime-proof.mjs` | PASS; produto com duas etapas independentes, duas operações/efeitos no provider-sim, GET final `SUCCEEDED` e duplicata idempotente sem novo efeito; `hub/evidence/r2/execution/product-http-latest.log` |
| Budget efetivo de I/O | `go test -count=1 ./internal/cometa -run 'TestProviderHTTPBudgetUsesFrozenTargetBudget\|TestEffectiveHTTPBudgetNeverExceedsLeaseOrContext' -v` | PASS; budget do snapshot limitado por contexto, lease e margem; `hub/evidence/r2/execution/capacity-budget-latest.log` |
| Resultado volumoso/FileRef | `go test -count=1 ./internal/orbita -run TestFinalizeMaterializesLargeResultAndLinksAfterCommit` + `go test -tags=e2e ... ./internal/objectstore -run TestPostgresS3MultipartFileRef` | PASS; resultado acima de 1 MiB usa upload streaming, fica `ORPHAN` até o protocolo confirmar e é exposto como `FileRef`; LocalStack real comprovou multipart, verificação/download e retenção |
| Backend | `go test -race -count=1 ./...` e `go vet ./...` | PASS; entrega Pulsar com domínio `r4-webhook` e lease seguro também coberta por teste PostgreSQL |
| Frontend | `npm run build` em `hub/admin-ui` | PASS; TypeScript e Vite, 41 módulos |
| Especificações | `openspec validate --all --strict --no-interactive --json` | PASS; 21/21 changes válidas, 0 falhas |
| Contratos JSON estritos | `hub/evidence/r2/execution/r4-ctr-02-latest.log` | PASS local; dialeto `ai-hub-json-schema-subset/v1`, enum numérico matemático, integer exato e limites de schema; rollout/rollback ainda aberto |
| R2-03/R2-02 complementar | `hub/evidence/r2/execution/r2-int-exe-qualification.log` | PASS seletivo; credenciais dedicadas, mTLS, polling/callback, prazos, SLA bilateral, idempotência, fencing, custódia e reconciliação; cenários fora do envelope local permanecem não qualificados |
| R2-04/R2-07 complementar | `hub/evidence/r2/execution/r2-cat-dad-qualification.log` | PASS seletivo; catálogo versionado, conflito de revisão, DAG, JSON estrito, importação/diff, paginação, FileRef, pins e restore controlado; limites de storage comercial/regional permanecem abertos |
| R2-08/R2-09 complementar | `hub/evidence/r2/execution/rls-runtime-latest.log` + `kind-continuity-latest.log` + `slo-load-latest.json` + `broker-outage-runtime-latest.json` + `environment-boundary-latest.log` + `optional-dependency-runtime-latest.log` + `provider-outage-runtime-latest.json` | PASS seletivo; RLS nas três bases, bootstrap Kind independente, UI/gateway, recuperação de workloads, perfil de referência SLO, reinício com broker fora, gate de promoção, dependência opcional sem cascata e provedor fora contido por capacidade; manutenção sob ruído, isolamento cross-environment integrado, ambientes remotos e HA regional permanecem abertos |

## Detalhes observados

- A carga final produziu `SUCCEEDED` para SYNC, dois polling, callback, AUTO,
  saldo estrito e credencial dedicada; o caso de falha forçada produziu
  `FAILED` conforme esperado.
- A repetição da chave do SYNC de sucesso retornou o mesmo `protocol_id` em
  quatro observações.
- A inbox de callback terminou sem itens `RECEIVED`; os efeitos do simulador
  foram observados sem reenvio durante o restore.
- O ingresso de callback rejeitou observações sem `provider_request_id` ou não
  terminais antes de criar custódia órfã; a inbox usa `(operation_id,
  body_sha256, token_hash)`, limite de bytes/itens e limpeza limitada de itens
  processados.
- A capacidade está ligada às concessões de `SUBMIT`, `STATUS` de polling,
  reconciliação e `FETCH` de webhook. Cada caminho valida o fencing antes da
  autenticação e imediatamente antes do I/O externo; submissão conserva
  `UNKNOWN` sem reenviar quando a lease cai, enquanto observadores e webhook
  liberam o permit para retomada segura. O Pulsar falha fechado quando o lease
  não cobre o timeout mais a margem de segurança; respostas/timeout fecham-no
  com sinal, latência e evidência. O runtime Compose instala o domínio
  `r4-webhook` sem alterar os volumes existentes.
- O executor de produto usa uma chave externa estável por comando de etapa,
  preservando a chave do protocolo nas operações simples. O probe HTTP criou
  duas etapas independentes, observou dois efeitos distintos e confirmou que a
  repetição da requisição retornou o mesmo protocolo sem novo efeito.
- A reconciliação administrativa rejeita `UNKNOWN` sem `provider_request_id`
  com estado `REJECTED`, auditoria e indicação explícita de ausência de
  correlação externa. O finalizador consegue materializar o fato a partir do
  snapshot do protocolo quando o intent histórico não está disponível.
- Resultado acima de 1 MiB não é colocado no JSONB do protocolo: o finalizador
  faz upload streaming pela autoridade de objetos, conserva a referência como
  `ORPHAN` antes do commit e executa `LinkResult` depois da representação final.
  Falha de vínculo permanece uma obrigação reconciliável, sem desfazer o
  protocolo já confirmado.
- O aceite congelou destinos de webhook por tenant/aplicação e o Pulsar
  consumiu somente o snapshot persistido no fato, sem consultar uma versão
  dinâmica posterior.
- A incidência econômica foi separada do estado operacional: a aceitação
  externa permanece `UNKNOWN` para a Órbita, mas publica `economic_kind=SUBMITTED`;
  cada polling publica `economic_kind=STATUS` com `attempt_id`. O oráculo
  financeiro confirmou fatos de custo/receita e journal balanceado no mesmo
  workload.
- O cenário exploratório do Browser Harness passou com CDP explícito e percorreu
  16 rotas; a descoberta automática de Chrome headless continua indisponível.
  Ele permanece auxiliar, e o gate determinístico do frontend é o Playwright.
- A prova complementar de R2-03/R2-02 usou endpoint OAuth HTTPS de teste local,
  PostgreSQL real e o cofre LocalStack versionado. O teste de `Retry-After`
  superior à janela preservou a espera e não permitiu nova claim além do
  deadline. O relatório administrativo de SLA foi exercitado por ID após a
  correção do filtro UUID.

## Atualização operacional de 11/09/2026

A primeira repetição da prova de entrega encontrou deriva de configuração no
runtime: o Pulsar estava com `EGRESS_PRIVATE_RULES` em `127.0.0.1/32`, embora os
nomes internos resolvessem para a rede Compose `172.19.0.0/16`. O diagnóstico
foi confirmado comparando IPs dos containers e a configuração efetiva do
Pulsar; a entrega foi recusada antes do POST e terminou como
`transport_unconfirmed`. O bootstrap oficial recalculou as três CIDRs e a
recriação controlada do Compose corrigiu a configuração sem remover volumes.

Após a correção, as provas foram repetidas: entrega webhook
`f61bda52-c97f-4c5d-9c71-70b8d363b86a` terminou com `open=0` e `pending=0`, a
incidência financeira terminou balanceada e sem quarentena inválida, e o RLS
confirmou isolamento cross-tenant nas três bases. O produto HTTP
`r4-product-http-1789104098588` finalizou com duas operações e dois efeitos
independentes. O smoke Chromium criou `browser-client-1789104122199` e
`browser-product-1789104180968`, com persistência do mapeamento entre etapas.

O incidente demonstra que a execução deve usar `bootstrap.sh` ou fornecer as
CIDRs descobertas no ambiente do Compose; `docker compose up` isolado pode
reintroduzir os defaults de loopback e não é um procedimento de qualificação.

### Correção de fila e recuperação de comandos — 11/09/2026

Uma repetição posterior da prova encontrou uma segunda causa independente:
`CreateQueue` do LocalStack devolve uma URL com hostname externo, enquanto o
worker dentro do Compose precisa usar `http://localstack:4566`. A mudança
`0a57029` normaliza essa autoridade somente no ambiente local, mantém URLs de
produção intocadas e registra a URL efetiva no log de prontidão do broker.

O mesmo commit passou a impedir que um comando com `StepDeadline` vencido
entre no executor. O envelope é conservado em `message_quarantine` antes do
ACK com razão `expired_command_deadline`; envelopes inválidos e rejeições
determinísticas recebem a mesma proteção. Falhas recuperáveis, como
`capacity_unavailable`, não são descartadas e seguem para redelivery.

Durante o diagnóstico, 15 mensagens antigas expiradas foram isoladas. Quatro
permits de `r4-poll-1` ainda tinham `pending_external`; o runner
`reconcile-local-pending.sh` consultou o provider-sim e fechou somente os
casos com HTTP 404 (`closed=4 protected=0`). Nenhum efeito externo presente
foi encerrado por essa rotina.

Após a reconciliação, o probe HTTP `product-runtime-proof.mjs` passou com a
chave `r4-product-http-1789109019646`: admissão `202`, duas operações e dois
efeitos independentes, consolidação pública `SUCCEEDED` e repetição da chave
sem novo plano ou efeito. O log saneado está em
`hub/evidence/r2/execution/product-http-latest.log`.

### Qualificação adicional de catálogo, dados e portal — 11/09/2026

O runner `identity-scope-proof.py` confirmou a fronteira do leitor global no
runtime persistido: `auditor-global` emite sujeito, `pwd+otp`, papel
`hub_protocol_reader` e `admin:cross_tenant`; leituras administrativas de
Atlas e Órbita entre tenants retornaram 200 sem campos de segredo, a tentativa
de escrita retornou 403 e `leitor-a` permaneceu restrito ao próprio tenant.
Consultas globais continuam exigindo finalidade e são registradas como
`READ` na tabela de auditoria. O fixture não representa uma concessão de
produção.

Os testes focados de Atlas e Órbita foram reexecutados com `-race` usando
PostgreSQL real: publicação imutável, conflito de revisão, DAG/fan-out,
projeção JSON estrita, importação em staging com diff, sanitização e paginação
por cursor passaram. A prova de objetos confirmou resultado volumoso em
multipart/streaming com `FileRef`, pin de retenção e expurgo somente após a
obrigação ser liberada. O `rls-runtime-proof.sh` confirmou isolamento
cross-tenant para as três bases usando o papel `hub_runtime`.

O portal foi reconstruído no container oficial e repetiu o Playwright com
backend real. Além dos checks anteriores, os selects de referência agora
percorrem cursores de catálogo até um limite seguro de 1.000 itens e sinalizam
falha de lookup; não apresentam uma primeira página incompleta como catálogo
vazio. O editor também filtra localmente as opções já carregadas por nome ou
identificador, com check dedicado no smoke para garantir que a lista não seja
apresentada como uma seleção vazia sem explicação.

O mesmo smoke validou a troca de identidade após logout OIDC com
`id_token_hint`: `operadora-a` retorna ao portal, `auditor-global` entra em
seguida, informa tenant e finalidade, consulta protocolo de outro tenant e
abre seu detalhe sem receber ação de escrita. A finalidade é preservada na
rota do detalhe/timeline e a API continua exigindo a mesma auditoria.

Na mesma rodada, `leitor-a` consultou destinos webhook, protocolos e financeiro
com as permissões de leitura. O portal exibiu os dados autorizados em modo
somente leitura e não renderizou `Publicar versão`, `Solicitar reconciliação` ou
`Verificar e fechar período`; as regras de escrita continuam aplicadas também
pelas APIs, que exigem os escopos correspondentes. Isso cobre a fronteira de
autorização na experiência do usuário sem tratar a visibilidade da interface
como substituta da autorização do servidor.

O smoke também selecionou um `batch_id` do journal persistido, preparou um
ajuste idempotente como `operadora-a` e releu o registro na mesma jornada. Como
o `prepared_by` era o sujeito autenticado, o portal não mostrou `Aprovar` e
exibiu `Aguardando aprovador distinto`. O backend continua sendo a autoridade
final: `ApproveAdjustment` rejeita o mesmo sujeito e exige um aprovador
distinto; esta evidência cobre a prevenção de erro na experiência do usuário,
sem alegar que toda a jornada de aprovação cross-tenant esteja homologada.

Para a ação operacional, o smoke inseriu uma entrega sintética `EXHAUSTED`
com representação persistida e sem destino executável, abriu o detalhe pelo
`delivery_id` real, acionou `Reentregar mesmos bytes do webhook` e recebeu
`RETRY_SCHEDULED`. Uma consulta independente confirmou a auditoria
`REDELIVER_SAME_BYTES` com tenant, recurso, ação e subject não vazio; a fixture
foi removida ao final. Durante o ensaio foi corrigido o handler Pulsar para
serializar `destination_id`/`destination_version` nulos como `null`, em vez de
falhar com 503 ao ler entregas legadas sem destino versionado. O gate prova a
transição administrativa e a auditoria, não uma entrega posterior em endpoint
comercial.

Na mesma jornada, a reconciliação selecionou um protocolo local cuja operação
estava `UNKNOWN` sem `provider_request_id`. O portal enviou a justificativa,
recebeu HTTP 409 com a mensagem pública de ausência de correlação e a consulta
independente confirmou a solicitação `REJECTED` e o evento
`RECONCILIATION_REJECTED` auditado. Nenhum replay de provedor foi executado;
isso qualifica a disposição segura do caso incerto e não a recuperação positiva
quando uma correlação externa perdida é posteriormente encontrada.

O cliente HTTP administrativo também passou a rejeitar corpo vazio, JSON
malformado ou tipo JSON incompatível, inclusive em respostas HTTP 200. O smoke
intercepta uma resposta inválida, confirma a mensagem de autoridade e verifica
que a lista previamente carregada permanece disponível.

### Adapter REST versionado executável — 11/09/2026

O teste `r2-rest-adapter-latest.log` percorreu a fronteira `ProviderAdapter`
com um endpoint HTTP local independente do `provider-sim`. O adapter
`rest-json-v1` agora constrói os paths declarativos limitados da oferta (na prova,
`POST /analise` e `GET /consulta/{id}`; sem host, query, fragmento ou código),
preservou `idempotency_key`, `mode`, `input` e `file_refs`, aplicou API Key
resolvida pelo binding e validou o fluxo `202/PENDING` → `200/SUCCEEDED` com
marcador de resposta externo. O decoder recusou campos desconhecidos,
documentos concatenados, estados inválidos e status HTTP fora do contrato.

Executor, polling e reconciliação passaram a consumir essa fronteira para
submit/status/decodificação. A imagem `ai-hub-r2-cometa:r2` foi reconstruída a
partir do commit `ac9b346`, o container oficial foi recriado com `--no-deps` e
`/metrics` respondeu HTTP 200. A prova é seletiva de contrato e integração
local; não cobre provedor comercial, autenticação regional, crash/partição ou
produção.

## Limites e pendências explícitas

Esta rodada fecha a sequência operacional local de seed, carga, callback,
portal, cache, ofertas, restore e HA do laboratório. Ela não equivale à
conclusão integral dos 201 requisitos/732 cenários nem à homologação de
provedores reais, AWS regional, multi-célula ou produção.

A execução durável do DAG agora está conectada à admissão de produtos e a
jornada HTTP local com provider-sim tem evidência dedicada. Permanecem fora do
fechamento integral, entre outros, carga prolongada e expiração durante I/O,
qualificação de adapter/provedor real, vinculação produtiva de
financeiro/webhooks, fencing positivo de efeitos cujo `provider_request_id` já
foi perdido, projeção de catálogo em escala e a matriz integral de
requisitos/cenários. O perfil Kind independente fecha o gate local de
dependências e endpoints; não substitui IaC/HA regional dos ambientes remotos.
Esses itens continuam marcados como abertos no backlog R4 e não foram
convertidos em PASS por inferência a partir desta execução local.
