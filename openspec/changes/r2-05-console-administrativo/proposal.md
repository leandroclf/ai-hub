# Proposal: Console administrativo completo e orientado às jornadas

## Change ID

`r2-05-console-administrativo`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Produto, Frontend e UX**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Permitir que operadores configurem, publiquem, diagnostiquem e conciliem o Hub por jornadas completas, com dados do servidor e permissões efetivas.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-20 | P1 | Console limitado a quatro formulários e histórico efêmero | [hub/admin-ui/src/App.tsx:7](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L7) |
| F-21 | P1 | Faltam jornadas administrativas de produto, operação e financeiro | [hub/admin-ui/src/App.tsx:9](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L9) |
| F-22 | P2 | Formulários expõem detalhes técnicos e estados pouco guiados | [hub/admin-ui/src/pages/ServicesPage.tsx:81](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/ServicesPage.tsx#L81) |

## Goals

- Sessão, contexto e navegação administrativa
- Listagens reais e detalhe editável
- Clientes, aplicações e ofertas contratadas
- Catálogo importado e serviços executáveis
- Produtos agregados e compostos
- Provedores, vínculos e saúde de integração
- Contratos técnicos e políticas temporais
- Busca e timeline operacional de protocolos
- Entregas e reprocessamento com segurança de efeito
- SLA bilateral e painéis operacionais consultáveis
- Financeiro de compra, venda e conciliação
- Qualidade de uso, acessibilidade e integração real

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Produto, Frontend e UX.

## Scope

### In scope

- R2-ADM-01 — Sessão, contexto e navegação administrativa (baseline: CFG-01, SEG-01, SEG-04).
- R2-ADM-02 — Listagens reais e detalhe editável (baseline: CFG-01, CAT-01, CFG-06).
- R2-ADM-03 — Clientes, aplicações e ofertas contratadas (baseline: CFG-01, CFG-06, CAT-02, CAT-11).
- R2-ADM-04 — Catálogo importado e serviços executáveis (baseline: CAT-01, CAT-02, CFG-02, COM-02).
- R2-ADM-05 — Produtos agregados e compostos (baseline: CAT-03, CAT-04, CAT-05, CAT-07, CAT-08).
- R2-ADM-06 — Provedores, vínculos e saúde de integração (baseline: CFG-05, SEG-05, OPE-07, CAT-06).
- R2-ADM-07 — Contratos técnicos e políticas temporais (baseline: CAT-09, CFG-04, FIN-03, EXE-10, EXE-11).
- R2-ADM-08 — Busca e timeline operacional de protocolos (baseline: CFG-03, SEG-04, EXE-07, DAD-04).
- R2-ADM-09 — Entregas e reprocessamento com segurança de efeito (baseline: CFG-03, EXE-08, EXE-09, SEG-04).
- R2-ADM-10 — SLA bilateral e painéis operacionais consultáveis (baseline: OPE-06, OPE-10, CFG-03).
- R2-ADM-11 — Financeiro de compra, venda e conciliação (baseline: FIN-01, FIN-02, FIN-06, FIN-07, FIN-08, FIN-10).
- R2-ADM-12 — Qualidade de uso, acessibilidade e integração real (baseline: CFG-01, QUA-03, QUA-04).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `console-administrativo-operavel` com 12 requisitos R2 e 36 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/console-administrativo-operavel/spec.md), mantendo a rastreabilidade com CAT-01, CAT-02, CAT-03, CAT-04, CAT-05, CAT-06, CAT-07, CAT-08, CAT-09, CAT-11, CFG-01, CFG-02, CFG-03, CFG-04, CFG-05, CFG-06, COM-02, DAD-04, EXE-07, EXE-08, EXE-09, EXE-10, EXE-11, FIN-01, FIN-02, FIN-03, FIN-06, FIN-07, FIN-08, FIN-10, OPE-06, OPE-07, OPE-10, QUA-03, QUA-04, SEG-01, SEG-04, SEG-05. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `console-administrativo-operavel` (ADDED).
- Baseline afetada por rastreabilidade: CAT-01, CAT-02, CAT-03, CAT-04, CAT-05, CAT-06, CAT-07, CAT-08, CAT-09, CAT-11, CFG-01, CFG-02, CFG-03, CFG-04, CFG-05, CFG-06, COM-02, DAD-04, EXE-07, EXE-08, EXE-09, EXE-10, EXE-11, FIN-01, FIN-02, FIN-03, FIN-06, FIN-07, FIN-08, FIN-10, OPE-06, OPE-07, OPE-10, QUA-03, QUA-04, SEG-01, SEG-04, SEG-05. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/admin-ui/src`
- `hub/admin-ui/package.json`
- `hub/admin-ui/vite.config.ts`
- `hub/internal/atlas/handlers.go`
- `hub/internal/orbita/handlers.go`
- `hub/internal/libra/handlers.go`
- `hub/api/openapi-internal.yaml`

### Data

UI não é autoridade: rascunhos, revisão concorrente, histórico e ações persistem no domínio. Sessão autenticada isolada por pessoa; preferência de filtro não contém segredo ou corpo sensível.

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

APIs operacionais dependem das changes 02/03/04/06/07/08. P-09 nomeia administradores. É possível construir navegação e listas cedo, mas nenhuma tela com mock será aceita como integração concluída.
