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

O backlog local diagnosticado continha 15 comandos expirados e quatro permits
com `pending_external`. Os comandos foram quarentenados e os permits só foram
fechados após o provider-sim confirmar HTTP 404, com `closed=4 protected=0`.
Na sequência, a prova HTTP do produto passou novamente com duas etapas,
dois efeitos independentes, consolidação `SUCCEEDED` e repetição idempotente
sem novo efeito; o resultado está em
`hub/evidence/r2/execution/product-http-latest.log`.

## Limite de conclusão

Este resultado não declara conclusão integral dos 201 requisitos/732 cenários
nem homologação de produção. O backlog herdado do R4-04 ainda contém cortes
arquiteturais abertos, incluindo provedor comercial, carga e budgets completos
em todo I/O, fencing geral, financeiro/webhooks/FileRefs produtivos, telemetria
bilateral, IaC/HA regional e qualificação dos cenários sem evidência. A matriz
integral de rastreabilidade está em `RESULT-MATRIX-732-CENARIOS.csv`: 64 linhas
estão associadas a resultados existentes e 668 permanecem
`NAO_QUALIFICADO_NESTA_RODADA`. O estado detalhado permanece em
`REQUIREMENTS_STATUS.csv`, `SCENARIO_RESULTS.csv`, `RESULT-MATRIX-732-CENARIOS.csv`, `CHECKPOINT.md` e
`OPENSPEC-AUDIT-2026-09-10.md`; nenhum desses itens foi fechado somente por
documentação.
