# Relatório da execução R4

## Atualização de qualificação — 10/09/2026

A sequência operacional local foi requalificada nesta rodada: catálogo
versionado, carga autenticada, SYNC/ASYNC/AUTO, polling, callback, negativas
de callback, portal administrativo, cache, ofertas, restore, Compose, RLS,
kind/HA, backend, frontend e OpenSpec strict. O Browser Harness permanece
disponível como ferramenta exploratória, mas está bloqueado neste host por
ausência de um endpoint CDP/`DevToolsActivePort` utilizável.

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

## Limite de conclusão

Este resultado não declara conclusão integral dos 201 requisitos/732 cenários
nem homologação de produção. O backlog herdado do R4-04 ainda contém cortes
arquiteturais abertos, incluindo DAG conectado, capacidade e budgets em todo
I/O, fencing geral, financeiro/webhooks/FileRefs completos, telemetria bilateral,
IaC/HA regional e matriz integral. O estado detalhado permanece em
`REQUIREMENTS_STATUS.csv`, `SCENARIO_RESULTS.csv`, `CHECKPOINT.md` e
`OPENSPEC-AUDIT-2026-09-10.md`; nenhum desses itens foi fechado somente por
documentação.
