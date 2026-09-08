# Proposal: Custódia, execução única e resultado final correto

## Change ID

`r2-02-execucao-duravel-e-resultados`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Engenharia de Core e Dados**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Uma obrigação recuperável por aceite, um único dono de despacho e uma representação final proveniente do provedor e durável antes da resposta.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-04 | P0 | 202 após falha de envio de comando, sem retomada de despacho | [hub/internal/orbita/handlers.go:178](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L178) |
| F-05 | P0 | Consumidores removem mensagens mesmo quando o efeito falha | [hub/internal/cometa/worker.go:34](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/worker.go#L34) |
| F-06 | P0 | Resultado retornado não é o resultado do provedor | [hub/internal/cometa/executor.go:120](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L120) |
| F-07 | P0 | Concorrência pode disparar mais de uma operação externa | [hub/internal/cometa/store.go:32](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/store.go#L32) |
| F-08 | P0 | Estado externo e fato são gravados em transações separadas | [hub/internal/cometa/executor.go:192](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L192) |
| F-09 | P0 | Corrida temporal permite sucesso depois do deadline | [hub/internal/orbita/store.go:204](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/store.go#L204) |
| F-11 | P1 | SYNC, AUTO e UUID em erros não cumprem todo o contrato | [hub/internal/orbita/handlers.go:168](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L168) |
| F-28 | P1 | Resultado final não é materializado como representação imutável completa | [hub/internal/orbita/finalize.go:49](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/finalize.go#L49) |

## Goals

- Aceite e intenção recuperáveis em todas as modalidades
- Idempotência concorrente com UUID recuperável
- Um dono de despacho e tentativa anterior ao efeito
- Resultado externo válido e durável
- Inbox, outbox e confirmação de mensagens
- Deadline por confirmação durável e final tardio
- SYNC direto e AUTO com espera limitada
- Representação final única por contrato de cliente
- Reconciliação de obrigações sem reexecutar efeitos

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Engenharia de Core e Dados.

## Scope

### In scope

- R2-EXE-01 — Aceite e intenção recuperáveis em todas as modalidades (baseline: EXE-01, EXE-15, DAD-03, COM-03, DAD-09, DAD-02).
- R2-EXE-02 — Idempotência concorrente com UUID recuperável (baseline: EXE-01, EXE-16, DAD-03).
- R2-EXE-03 — Um dono de despacho e tentativa anterior ao efeito (baseline: EXE-04, EXE-09, EXE-15, DAD-03, COM-06).
- R2-EXE-04 — Resultado externo válido e durável (baseline: EXE-04, EXE-14, DAD-04, COM-06).
- R2-EXE-05 — Inbox, outbox e confirmação de mensagens (baseline: COM-03, COM-04, OPE-11, FIN-04, EXE-08).
- R2-EXE-06 — Deadline por confirmação durável e final tardio (baseline: EXE-11, EXE-12, FIN-10, OPE-10).
- R2-EXE-07 — SYNC direto e AUTO com espera limitada (baseline: EXE-02, EXE-14, COM-01, COM-06, CAT-11).
- R2-EXE-08 — Representação final única por contrato de cliente (baseline: COM-05, COM-04, DAD-04, CAT-09, EXE-07).
- R2-EXE-09 — Reconciliação de obrigações sem reexecutar efeitos (baseline: EXE-03, EXE-09, EXE-15, OPE-11, DAD-09).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `execucao-duravel-e-resultados` com 9 requisitos R2 e 29 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/execucao-duravel-e-resultados/spec.md), mantendo a rastreabilidade com CAT-09, CAT-11, COM-01, COM-03, COM-04, COM-05, COM-06, DAD-02, DAD-03, DAD-04, DAD-09, EXE-01, EXE-02, EXE-03, EXE-04, EXE-07, EXE-08, EXE-09, EXE-11, EXE-12, EXE-14, EXE-15, EXE-16, FIN-04, FIN-10, OPE-10, OPE-11. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `execucao-duravel-e-resultados` (ADDED).
- Baseline afetada por rastreabilidade: CAT-09, CAT-11, COM-01, COM-03, COM-04, COM-05, COM-06, DAD-02, DAD-03, DAD-04, DAD-09, EXE-01, EXE-02, EXE-03, EXE-04, EXE-07, EXE-08, EXE-09, EXE-11, EXE-12, EXE-14, EXE-15, EXE-16, FIN-04, FIN-10, OPE-10, OPE-11. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/internal/orbita`
- `hub/internal/cometa/executor.go`
- `hub/internal/cometa/store.go`
- `hub/internal/outbox`
- `hub/internal/queue`
- `hub/internal/dispatch/contract.go`
- `hub/migrations/core/0001_init.sql`

### Data

Comando durável, dispatch lease/epoch, tentativa pré-envio, observação externa, resultado por versão, inbox/outbox, relógio/decisão de deadline e obrigação de reconciliação. Tabelas novas são propostas, não existentes.

### APIs / Contracts

Contrato público versionado e APIs administrativas/internas conforme [contratos](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md). Métodos novos do console estão em [jornadas](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md); não são endpoints existentes presumidos.

### Integrations

Dependências: r2-01-identidade-e-isolamento Contratos compartilhados são acordados antes da implementação consumidora; mocks não contam como integração pronta.

### Operations

Provas por perfil, métricas/alertas e recuperação descritas no design; novas dependências de runtime devem ser declaradas no deploy completo.

### Security / Privacy

Identidade/tenant, mínimo privilégio, segredo por referência e evidência sanitizada em todos os caminhos, inclusive erro e operação administrativa.

## Risks and Mitigations

Ver [risk-matrix.md](risk-matrix.md), com evidência, responsável, mitigação e prova ainda pendente.

## Success Criteria

- Todos os cenários da capability executados com evidência vinculada ao SHA e ambiente; conflitos/falhas não omitidos.
- Migração, compatibilidade, rollback e interfaces dos consumidores verificados.
- Nenhuma tarefa fechada apenas por build, mock, texto ou manifesto não aplicado.

## Assumptions

Preservar as cinco autoridades e stack v4; valores sintéticos servem somente ensaios. Pendências não impedem correção genérica, mas bloqueiam ativação do perfil dependente.

## Open Questions

P-07 fecha SLAs e elegibilidade temporal comercial. A proposta conserva deadline rígido da v4; não inventa tolerância para resultado tardio.
