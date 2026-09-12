# Relatório da execução R4

## Atualização de qualificação — 11/09/2026

A sequência operacional local foi requalificada nesta rodada: catálogo
versionado, carga autenticada, SYNC/ASYNC/AUTO, polling, callback, negativas
de callback, portal administrativo, cache, ofertas, restore, Compose, RLS,
kind/HA, backend, frontend e OpenSpec strict. O Browser Harness passou como
ferramenta exploratória via `BU_CDP_URL` explícito, percorrendo as 16 rotas
administrativas em 390×844; a descoberta automática de Chrome headless ainda
não funciona neste host.

O relatório detalhado, com comandos, limites e proveniência, está em
[`EXECUTION-2026-09-10.md`](EXECUTION-2026-09-10.md). Os artefatos dinâmicos
estão em `hub/evidence/r2/execution/`.

O laboratório Kind também foi qualificado em perfil independente: o bootstrap
carrega as imagens offline, executa migrações/reconciliação OIDC e materializa
dependências, UI e gateway dentro do cluster. O gate de continuidade removeu
controladamente um pod de Cometa e um de Pulsar, observando retorno a 2/2 em
4,453 ms e 4,481 ms.

Na requalificação posterior, a suíte Go normal e com `-race`, o build do
portal, Playwright, RLS, carga autorizada e restore passaram novamente. O
bootstrap passou a preservar os IPs descobertos das fixtures após o build; as
pendências locais de capacidade foram reconciliadas somente quando o oráculo
sintético comprovou ausência de efeito (`404`), com confirmação explícita e
registro em `capacity-reconciliation-latest.log`. O teste concorrente de
admissão usa uma célula exclusiva por execução para não ser consumido pelo
worker Orbita do laboratório.

O produto composto também passou pela fronteira HTTP pública: o plano persistiu
duas etapas independentes, o provider-sim observou dois efeitos e a duplicata
da mesma chave não criou novo efeito. O editor administrativo persiste e relê
o mapeamento de entrada entre etapas. O budget efetivo passou nos limites de
snapshot, contexto, lease e margem.

Na qualificação complementar de R2-03/R2-02, duas credenciais dedicadas para a
mesma conta foram exercitadas simultaneamente sem cruzamento de tokens; mTLS
foi validado com certificado de cliente e CA fixada; polling/callback, fencing,
TTL zero, `Retry-After` incompatível, inbox/outbox, AUTO e reconciliação foram
reexecutados com PostgreSQL real. A correção do filtro UUID do relatório de SLA
também foi coberta: `MONITOR_ONLY` preserva o sucesso do cliente e registra
`MONITOR_ONLY_BREACH`, enquanto `REJECT_LATE` produz falha contratada.

Na qualificação complementar de R2-04/R2-07/R2-08, os testes focados cobriram
publicação imutável e conflito de revisão, DAG/fan-out, projeção JSON estrita,
importação em staging com diff, paginação por cursor, resultado volumoso em
`FileRef`, pins de retenção, RLS nas três bases e restore sem replay. O portal
foi reconstruído e repetiu o smoke autenticado; os lookups de referência agora
seguem cursores até 1.000 itens e transformam falhas em estado explícito.

### Adapter REST versionado executável — 11/09/2026

O Cometa deixou de construir o contrato HTTP diretamente para o `provider-sim`.
Foi criada a fronteira `ProviderAdapter`, com `rest-json-v1` compilado e
habilitado somente quando instalado pela registry. Submit, polling e
reconciliação obtêm método, path, payload e decodificação dessa fronteira;
autenticação, egress, capacidade, fencing e custódia permanecem nas camadas
do Cometa.

O teste de contrato usou um endpoint HTTP local independente do simulador:
validou `POST /v1/operations`, `GET /v1/operations/{id}` com escaping,
`idempotency_key`, `input`, `file_refs`, API Key por binding e o fluxo
`202/PENDING` → `200/SUCCEEDED` com marcador real. Campos desconhecidos,
documentos concatenados, status HTTP inesperado e estados de negócio inválidos
foram recusados. A imagem oficial do Cometa foi reconstruída e iniciou no
Compose `ai_hub_r3qual`; a evidência está em
`hub/evidence/r2/execution/r2-rest-adapter-latest.log`.

Esse avanço cobre a fronteira executável e o contrato de homologação local,
mas não transforma o fixture em provedor comercial nem fecha sozinho a
homologação de produção, os ensaios de crash/partição ou a matriz integral.

## Resultado

Implementação e qualificação local integradas passaram nos gates declarados.
O catálogo é semeado pelos endpoints versionados do console; a
qualificação/célula de capacidade fica restrita ao fixture local autorizado. A
carga não imprime nem persiste o bearer. O callback exige chave de ingresso,
capability por operação e observação terminal; a inbox aplica identidade por
capability, limites e retenção limitada. A capacidade está ligada a
`SUBMIT`, `STATUS`, reconciliação e `FETCH` de webhook. O Pulsar só inicia a
entrega quando o lease cobre o timeout configurado com margem de segurança e
fecha o permit com sinal, latência e evidência; destinos webhook são congelados
no aceite e o Pulsar entrega o snapshot persistido. O cache usa L1 privado,
expiração/revogação e locks limitados; Redis permanece opcional e vazio.
Na mesma rodada, a incidência econômica foi desacoplada do estado operacional:
aceitação externa `UNKNOWN` gera incidência `SUBMITTED`, cada polling gera
incidência `STATUS` por `attempt_id`, e o oráculo financeiro confirmou fatos de
custo/receita, inbox e journal balanceado no workload autorizado.
O polling passou a exigir o `provider_request_id` congelado no claim; uma
resposta divergente fica como recibo de investigação e não produz estado ou
outbox da operação protegida. O Browser Smoke também confirmou que, após o
logout, uma chamada sem credencial à API administrativa retorna HTTP 401.

## Atualização operacional — 11/09/2026

Após o bootstrap oficial do único Compose, o worker Cometa foi corrigido para
normalizar a URL externa devolvida pelo LocalStack para a autoridade interna
do serviço, detectar `StepDeadline` vencido antes da execução e conservar o
envelope completo em quarentena antes do ACK. Rejeições determinísticas são
isoladas; falhas recuperáveis permanecem sujeitas a redelivery. A submissão
também valida a posse durável da operação por tenant, aplicação, célula, epoch,
lease e hash do comando antes da autenticação e imediatamente antes do POST,
sem reenvio quando a posse é perdida.

O backlog local diagnosticado continha 15 comandos expirados e dez permits
com `pending_external`. Os comandos foram quarentenados e os permits só foram
fechados após o provider-sim confirmar HTTP 404, com `closed=10 protected=0`.
Na sequência, a prova HTTP do produto passou novamente com duas etapas,
dois efeitos independentes, consolidação `SUCCEEDED` e repetição idempotente
sem novo efeito; o resultado está em
`hub/evidence/r2/execution/product-http-latest.log`.

A carga autorizada subsequente passou com o prefixo
`r4-authorized-1789127281502`; financeiro, webhook, RLS, restore por digest,
continuidade Kind e o smoke Playwright também passaram. O Browser Harness
percorreu as 16 rotas com sessão OIDC/OTP qualificada, viewport 390×844 e
resultado `PASS-EXPLORATORY`. Durante a repetição foram renovados o upstream
do `admin-ui` após rebuild da Órbita e as CIDRs internas do Pulsar; nenhuma
dessas correções alterou dados produtivos ou versionou segredos.

## Limite de conclusão

Este resultado não declara conclusão integral dos 201 requisitos/732 cenários
nem homologação de produção. O backlog herdado do R4-04 ainda contém cortes
arquiteturais abertos, incluindo provedor comercial, carga e budgets completos
em todo I/O, fencing geral, financeiro/webhooks/FileRefs produtivos, telemetria
bilateral, IaC/HA regional e qualificação dos cenários sem evidência. A matriz
integral de rastreabilidade está em `RESULT-MATRIX-732-CENARIOS.csv`: 231 linhas
estão associadas a resultados existentes e 501 permanecem
`NAO_QUALIFICADO_NESTA_RODADA`. O estado detalhado permanece em
`REQUIREMENTS_STATUS.csv`, `SCENARIO_RESULTS.csv`, `RESULT-MATRIX-732-CENARIOS.csv`, `CHECKPOINT.md` e
`OPENSPEC-AUDIT-2026-09-10.md`; nenhum desses itens foi fechado somente por
documentação.
