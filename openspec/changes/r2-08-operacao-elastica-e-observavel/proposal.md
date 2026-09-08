# Proposal: Ambientes completos, disponibilidade, escala e observabilidade

## Change ID

`r2-08-operacao-elastica-e-observavel`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Plataforma, SRE e Engenharia**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Subir a plataforma completa, escalar com isolamento e manter operações elegíveis durante falhas de dependências não autoritativas.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-14 | P1 | Redis e broker bloqueiam inclusive reinício do caminho SYNC | [hub/cmd/cometa/main.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/cometa/main.go#L43) |
| F-31 | P1 | Docker local não inclui toda a plataforma e não garante persistência na recriação | [hub/deploy/docker-compose.yml:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L3) |
| F-32 | P1 | Kubernetes é referência incompleta, sem os cinco ambientes | [hub/deploy/k8s/README.md:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/README.md#L3) |
| F-33 | P1 | Autoscaling de células e reconciliação de IaC não fecham o circuito | [hub/migrations/control/0001_init.sql:62](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/control/0001_init.sql#L62) |
| F-34 | P1 | Clientes AWS e bootstrap são fixos ao ambiente local | [hub/internal/queue/queue.go:58](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/queue/queue.go#L58) |
| F-35 | P1 | Probes não acompanham capacidades e não há drenagem graciosa | [hub/internal/platform/httpserver/httpserver.go:86](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/httpserver.go#L86) |
| F-36 | P1 | Observabilidade não mede SLA, pressão nem custódia | [hub/internal/platform/httpserver/metrics.go:16](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/metrics.go#L16) |
| F-42 | P1 | Recepção e pools não limitam custo por tenant | [hub/internal/orbita/handlers.go:114](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L114) |

## Goals

- Plataforma local completa e persistente
- Kubernetes e cinco ambientes coerentes
- Probes por capacidade e drenagem
- Escala de pods, nós e dados sem chamados rotineiros
- Placement, fencing e recursos por ambiente e célula
- Fallback seletivo e dependências mínimas
- Telemetria e SLA com investigação acionável
- Segurança e reprodutibilidade de infraestrutura
- SLO, isolamento e continuidade qualificados

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Plataforma, SRE e Engenharia.

## Scope

### In scope

- R2-OPE-01 — Plataforma local completa e persistente (baseline: OPE-04, QUA-01, DAD-07).
- R2-OPE-02 — Kubernetes e cinco ambientes coerentes (baseline: OPE-04, OPE-05, ARQ-06, ARQ-01).
- R2-OPE-03 — Probes por capacidade e drenagem (baseline: OPE-05, OPE-13, OPE-14).
- R2-OPE-04 — Escala de pods, nós e dados sem chamados rotineiros (baseline: OPE-12, OPE-08, CFG-06, DAD-11, ARQ-06).
- R2-OPE-05 — Placement, fencing e recursos por ambiente e célula (baseline: DAD-11, ARQ-06, CFG-06, OPE-08).
- R2-OPE-06 — Fallback seletivo e dependências mínimas (baseline: OPE-13, DAD-09, DAD-10, ARQ-02).
- R2-OPE-07 — Telemetria e SLA com investigação acionável (baseline: OPE-03, OPE-06, OPE-10).
- R2-OPE-08 — Segurança e reprodutibilidade de infraestrutura (baseline: ARQ-05, SEG-01, OPE-04, QUA-04).
- R2-OPE-09 — SLO, isolamento e continuidade qualificados (baseline: OPE-01, OPE-02, OPE-03, OPE-09, OPE-11, OPE-15).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `operacao-elastica-e-observavel` com 9 requisitos R2 e 27 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/operacao-elastica-e-observavel/spec.md), mantendo a rastreabilidade com ARQ-01, ARQ-02, ARQ-05, ARQ-06, CFG-06, DAD-07, DAD-09, DAD-10, DAD-11, OPE-01, OPE-02, OPE-03, OPE-04, OPE-05, OPE-06, OPE-08, OPE-09, OPE-10, OPE-11, OPE-12, OPE-13, OPE-14, OPE-15, QUA-01, QUA-04, SEG-01. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `operacao-elastica-e-observavel` (ADDED).
- Baseline afetada por rastreabilidade: ARQ-01, ARQ-02, ARQ-05, ARQ-06, CFG-06, DAD-07, DAD-09, DAD-10, DAD-11, OPE-01, OPE-02, OPE-03, OPE-04, OPE-05, OPE-06, OPE-08, OPE-09, OPE-10, OPE-11, OPE-12, OPE-13, OPE-14, OPE-15, QUA-01, QUA-04, SEG-01. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/deploy`
- `hub/cmd`
- `hub/internal/platform`
- `hub/internal/atlasclient`
- `hub/internal/queue`
- `hub/internal/objectstore`

### Data

Placement e epoch por tenant/contrato/célula; estado de provisionamento e qualificação; métricas/alertas provisionados; parâmetros por ambiente com segredo externo; orçamento de pools e reservas por célula.

### APIs / Contracts

Contrato público versionado e APIs administrativas/internas conforme [contratos](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md). Métodos novos do console estão em [jornadas](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md); não são endpoints existentes presumidos.

### Integrations

Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados Contratos compartilhados são acordados antes da implementação consumidora; mocks não contam como integração pronta.

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

P-01/P-08/P-10/P-11 bloqueiam ativação cloud/crítica, não a entrega do laboratório completo. Distinguir kind como laboratório de Kubernetes gerenciado remoto.
