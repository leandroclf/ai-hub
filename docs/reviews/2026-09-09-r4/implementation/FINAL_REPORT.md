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

## Resultado

Implementação e qualificação local integradas passaram nos gates declarados.
O catálogo é semeado pelos endpoints versionados do console; a
qualificação/célula de capacidade fica restrita ao fixture local autorizado. A
carga não imprime nem persiste o bearer. O callback exige chave de ingresso,
capability por operação e observação terminal; a inbox aplica identidade por
capability, limites e retenção limitada. A capacidade está ligada a
`SUBMIT`, `STATUS` e reconciliação. Destinos webhook são congelados no aceite
e o Pulsar entrega o snapshot persistido. O cache usa L1 privado,
expiração/revogação e locks limitados; Redis permanece opcional e vazio.

## Limite de conclusão

Este resultado não declara conclusão integral dos 201 requisitos/732 cenários
nem homologação de produção. O backlog herdado do R4-04 ainda contém cortes
arquiteturais abertos, incluindo DAG conectado, capacidade e budgets em todo
I/O, fencing geral, financeiro/webhooks/FileRefs completos, kind independente
de Compose, telemetria bilateral e matriz integral. O estado detalhado permanece em
`REQUIREMENTS_STATUS.csv`, `SCENARIO_RESULTS.csv`, `CHECKPOINT.md` e
`OPENSPEC-AUDIT-2026-09-10.md`; nenhum desses itens foi fechado somente por
documentação.
