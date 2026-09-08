# Proposal: Identidade, isolamento de tenants e administração segura

## Change ID

`r2-01-identidade-e-isolamento`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Segurança e Engenharia de Plataforma**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Fechar a cadeia identidade → autorização por recurso → auditoria, da borda até os domínios internos.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-01 | P0 | Identidade do cliente controlada pelo próprio chamador | [hub/internal/orbita/handlers.go:95](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L95) |
| F-02 | P0 | APIs internas e administrativas sem identidade de workload | [hub/internal/orbita/handlers.go:53](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L53) |
| F-41 | P1 | Destinos configuráveis não possuem defesa SSRF | [hub/internal/atlas/handlers.go:100](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/handlers.go#L100) |

## Goals

- Identidade autenticada e autorização por recurso
- Workloads com menor privilégio
- Leitura administrativa individual entre tenants
- Destinos externos e callbacks protegidos
- Sessão administrativa e trilha de alterações

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Segurança e Engenharia de Plataforma.

## Scope

### In scope

- R2-SEG-01 — Identidade autenticada e autorização por recurso (baseline: SEG-01, SEG-03, EXE-07, EXE-16).
- R2-SEG-02 — Workloads com menor privilégio (baseline: SEG-01, SEG-03, DAD-01, DAD-07).
- R2-SEG-03 — Leitura administrativa individual entre tenants (baseline: SEG-04, CFG-03).
- R2-SEG-04 — Destinos externos e callbacks protegidos (baseline: SEG-02, COM-02, CFG-05).
- R2-SEG-05 — Sessão administrativa e trilha de alterações (baseline: CFG-01, CFG-02, SEG-01, SEG-03).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `identidade-e-isolamento-operacional` com 5 requisitos R2 e 15 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/identidade-e-isolamento-operacional/spec.md), mantendo a rastreabilidade com CFG-01, CFG-02, CFG-03, CFG-05, COM-02, DAD-01, DAD-07, EXE-07, EXE-16, SEG-01, SEG-02, SEG-03, SEG-04. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `identidade-e-isolamento-operacional` (ADDED).
- Baseline afetada por rastreabilidade: CFG-01, CFG-02, CFG-03, CFG-05, COM-02, DAD-01, DAD-07, EXE-07, EXE-16, SEG-01, SEG-02, SEG-03, SEG-04. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/deploy/kong/kong.yml`
- `hub/internal/platform/httpserver/httpserver.go`
- `hub/internal/orbita/handlers.go`
- `hub/internal/atlas/handlers.go`
- `hub/internal/libra/handlers.go`
- `hub/internal/pulsar/handlers.go`

### Data

Principais administrativos individuais, grants por aplicação/tenant, versões de política e trilha append-only de decisões de acesso; credenciais nunca na trilha.

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

P-09 define responsáveis e perfis nominais; P-04 define dados permitidos. Essas pendências não impedem eliminar confiança em headers arbitrários.
