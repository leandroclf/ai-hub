# Proposal: Qualificação por requisito e evidência reproduzível

## Change ID

`r2-09-qualificacao-e-rastreabilidade`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Qualidade, Liderança técnica e SRE**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Converter o alvo de 98 requisitos em entregas demonstráveis e impedir que teste superficial, manifesto de referência ou screenshot se torne prova de produção.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-37 | P1 | Ensaios de HA, isolamento e recuperação não foram demonstrados | [hub/deploy/k8s/README.md:106](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/README.md#L106) |
| F-38 | P1 | Cobertura existente não prova as garantias que os nomes dos testes sugerem | [hub/test/e2e/e2e_test.go:78](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/test/e2e/e2e_test.go#L78) |
| F-39 | P2 | Rastreabilidade histórica contém conclusões que não refletem o snapshot | [openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md:68](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md#L68) |
| F-40 | P2 | Versões e contratos de ferramenta carecem de qualificação operacional | [hub/go.mod:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/go.mod#L3) |

## Goals

- Rastreabilidade do alvo até evidência
- Fixtures e oráculos independentes reproduzíveis
- Gates de contrato, experiência e operação
- Evolução OpenSpec e baseline sem falso arquivamento

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Qualidade, Liderança técnica e SRE.

## Scope

### In scope

- R2-QUA-01 — Rastreabilidade do alvo até evidência (baseline: QUA-03, QUA-04, DEC-02, DEC-05).
- R2-QUA-02 — Fixtures e oráculos independentes reproduzíveis (baseline: QUA-01, QUA-02, QUA-05, QUA-06).
- R2-QUA-03 — Gates de contrato, experiência e operação (baseline: QUA-04, QUA-06, COM-04, OPE-15).
- R2-QUA-04 — Evolução OpenSpec e baseline sem falso arquivamento (baseline: DEC-01, DEC-03, DEC-04, DEC-05, QUA-03, ARQ-04).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `qualificacao-integrada-r2` com 4 requisitos R2 e 12 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/qualificacao-integrada-r2/spec.md), mantendo a rastreabilidade com ARQ-04, COM-04, DEC-01, DEC-02, DEC-03, DEC-04, DEC-05, OPE-15, QUA-01, QUA-02, QUA-03, QUA-04, QUA-05, QUA-06. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `qualificacao-integrada-r2` (ADDED).
- Baseline afetada por rastreabilidade: ARQ-04, COM-04, DEC-01, DEC-02, DEC-03, DEC-04, DEC-05, OPE-15, QUA-01, QUA-02, QUA-03, QUA-04, QUA-05, QUA-06. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `openspec/changes/hub-interoperabilidade-v4`
- `IMPLEMENTATION_AUDIT.md`
- `hub/test/e2e/e2e_test.go`
- `hub/internal/atlas/handlers_test.go`
- `hub/evidence`
- `docs/openspec-docs`

### Data

Matriz requisito → gap → change → cenário → tarefa → execução/artefato/SHA; fixtures sintéticas; evidências sanitizadas; decisões com estado e responsável.

### APIs / Contracts

Contrato público versionado e APIs administrativas/internas conforme [contratos](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md). Métodos novos do console estão em [jornadas](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md); não são endpoints existentes presumidos.

### Integrations

Dependências: Sem pré-requisito de implementação para iniciar; integrações específicas descritas abaixo. Contratos compartilhados são acordados antes da implementação consumidora; mocks não contam como integração pronta.

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

P-01 a P-11 permanecem com responsáveis de função, sem nomes inventados. Falta de decisão limita ativação específica; não é justificativa para declarar funcionalidade incompleta como pronta.
