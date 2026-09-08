# Proposal: Adaptadores reais, credenciais, polling, callbacks e pressão

## Change ID

`r2-03-integracoes-credenciais-e-pressao`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Integrações e Engenharia de Core**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Executar contratos homologados do provedor com credencial correta e capacidade adaptativa compartilhada, preservando cada observação e entrega.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-03 | P0 | Callback confirma recebimento sem custódia comprovada | [hub/internal/cometa/handlers.go:73](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/handlers.go#L73) |
| F-10 | P1 | TTL de retry configurado não governa execução | [hub/internal/atlas/store.go:40](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/store.go#L40) |
| F-12 | P0 | Credencial dedicada resolvida não é a usada na chamada | [hub/internal/cometa/executor.go:83](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L83) |
| F-13 | P0 | Autenticação do simulador não equivale a integração segura real | [hub/internal/providerauth/client.go:57](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/providerauth/client.go#L57) |
| F-15 | P1 | Polling sem autenticação, lease e orçamento configurável | [hub/internal/cometa/poller.go:48](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/poller.go#L48) |
| F-16 | P1 | Não há amortecimento adaptativo nem isolamento de carga | [hub/internal/cometa/worker.go:34](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/worker.go#L34) |
| F-26 | P1 | Webhook usa URL como identidade do segredo | [hub/internal/pulsar/store.go:64](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/store.go#L64) |
| F-27 | P1 | Webhook sem claim, recibos completos e política por cliente | [hub/internal/pulsar/store.go:105](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/store.go#L105) |

## Goals

- Adapter executável homologado por capacidade
- Credencial efetiva vinculada ao cliente e à conta
- Segredos reais e cache dispensável
- Polling e callback combináveis e coordenados
- Janela absoluta de retry e classificação de falha
- Controle adaptativo global por domínio de capacidade
- Entrega ao cliente com identidade e política próprias
- SLA do provedor separado de SLA do cliente

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Integrações e Engenharia de Core.

## Scope

### In scope

- R2-INT-01 — Adapter executável homologado por capacidade (baseline: COM-02, CAT-06, CAT-11).
- R2-INT-02 — Credencial efetiva vinculada ao cliente e à conta (baseline: CFG-05, SEG-05, FIN-11, CAT-11).
- R2-INT-03 — Segredos reais e cache dispensável (baseline: SEG-05, DAD-10, OPE-13, ARQ-03).
- R2-INT-04 — Polling e callback combináveis e coordenados (baseline: EXE-05, EXE-06, SEG-02).
- R2-INT-05 — Janela absoluta de retry e classificação de falha (baseline: EXE-09, EXE-10, CFG-04).
- R2-INT-06 — Controle adaptativo global por domínio de capacidade (baseline: OPE-07, OPE-08, EXE-13, CAT-06).
- R2-INT-07 — Entrega ao cliente com identidade e política próprias (baseline: EXE-08, COM-05, SEG-02, CFG-03).
- R2-INT-08 — SLA do provedor separado de SLA do cliente (baseline: OPE-10, EXE-11, FIN-10, CAT-10).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `integracoes-credenciais-e-pressao` com 8 requisitos R2 e 28 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/integracoes-credenciais-e-pressao/spec.md), mantendo a rastreabilidade com ARQ-03, CAT-06, CAT-10, CAT-11, CFG-03, CFG-04, CFG-05, COM-02, COM-05, DAD-10, EXE-05, EXE-06, EXE-08, EXE-09, EXE-10, EXE-11, EXE-13, FIN-10, FIN-11, OPE-07, OPE-08, OPE-10, OPE-13, SEG-02, SEG-05. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `integracoes-credenciais-e-pressao` (ADDED).
- Baseline afetada por rastreabilidade: ARQ-03, CAT-06, CAT-10, CAT-11, CFG-03, CFG-04, CFG-05, COM-02, COM-05, DAD-10, EXE-05, EXE-06, EXE-08, EXE-09, EXE-10, EXE-11, EXE-13, FIN-10, FIN-11, OPE-07, OPE-08, OPE-10, OPE-13, SEG-02, SEG-05. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/internal/cometa`
- `hub/internal/providerauth/client.go`
- `hub/internal/pulsar`
- `hub/internal/atlasclient/client.go`
- `hub/migrations/control`
- `hub/migrations/core`

### Data

Perfil de adapter versionado, binding/secret_version, capacity_domain/lease, retry_until, agenda/recibo de polling, inbox de callback, delivery e key_id versionados; segredos resolvidos no cofre, tokens somente L1 privado.

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

P-05 contém garantias reais dos parceiros; P-07 define TTL/SLA; P-11 os envelopes máximos. Sem confirmação de idempotência não habilitar retry de efeito ambíguo.
