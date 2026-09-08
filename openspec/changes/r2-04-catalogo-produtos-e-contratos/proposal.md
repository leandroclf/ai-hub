# Proposal: Catálogo operável, produtos compostos e contratos por cliente

## Change ID

`r2-04-catalogo-produtos-e-contratos`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Produto, Atlas e Engenharia de Core**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Separar serviço, produto, oferta, contrato de integração e adapter; publicar versões validadas que governam admissão e execução paralela.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-17 | P1 | Catálogo publicado pode ser sobrescrito e não governa admissão | [hub/internal/atlas/store.go:48](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/store.go#L48) |
| F-18 | P1 | Produtos, DAG e contratos legados ainda não existem | [hub/internal/orbita/handlers.go:23](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L23) |
| F-19 | P1 | Importação não ativa adaptadores e substitui configurações locais | [hub/deploy/import_hiveplace_collection.sh:38](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/import_hiveplace_collection.sh#L38) |

## Goals

- Catálogo versionado com publicação governada
- Ofertas por cliente e roteamento autorizado
- Agregação paralela limitada
- Composição por DAG e compensações
- Perfis técnicos de entrada e saída por cliente
- Snapshot e projeções independentes de consulta por pedido
- Importação segura e inventário executável distinto
- Consulta administrativa persistente e onboarding

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Produto, Atlas e Engenharia de Core.

## Scope

### In scope

- R2-CAT-01 — Catálogo versionado com publicação governada (baseline: CAT-01, CAT-02, CFG-02, CFG-04).
- R2-CAT-02 — Ofertas por cliente e roteamento autorizado (baseline: CAT-02, CAT-06, CAT-10, CAT-11, FIN-02).
- R2-CAT-03 — Agregação paralela limitada (baseline: CAT-03, CAT-05, CAT-08, EXE-13).
- R2-CAT-04 — Composição por DAG e compensações (baseline: CAT-04, CAT-05, CAT-08, EXE-03, EXE-13).
- R2-CAT-05 — Perfis técnicos de entrada e saída por cliente (baseline: CAT-09, COM-02, COM-04, COM-05).
- R2-CAT-06 — Snapshot e projeções independentes de consulta por pedido (baseline: FIN-03, CFG-02, DAD-08, DAD-10, ARQ-02).
- R2-CAT-07 — Importação segura e inventário executável distinto (baseline: CAT-01, CAT-02, CFG-02, COM-02, QUA-01).
- R2-CAT-08 — Consulta administrativa persistente e onboarding (baseline: CFG-01, CFG-06, CAT-11, DAD-11).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `catalogo-produtos-e-contratos-versionados` com 8 requisitos R2 e 24 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/catalogo-produtos-e-contratos-versionados/spec.md), mantendo a rastreabilidade com ARQ-02, CAT-01, CAT-02, CAT-03, CAT-04, CAT-05, CAT-06, CAT-08, CAT-09, CAT-10, CAT-11, CFG-01, CFG-02, CFG-04, CFG-06, COM-02, COM-04, COM-05, DAD-08, DAD-10, DAD-11, EXE-03, EXE-13, FIN-02, FIN-03, QUA-01. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `catalogo-produtos-e-contratos-versionados` (ADDED).
- Baseline afetada por rastreabilidade: ARQ-02, CAT-01, CAT-02, CAT-03, CAT-04, CAT-05, CAT-06, CAT-08, CAT-09, CAT-10, CAT-11, CFG-01, CFG-02, CFG-04, CFG-06, COM-02, COM-04, COM-05, DAD-08, DAD-10, DAD-11, EXE-03, EXE-13, FIN-02, FIN-03, QUA-01. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/internal/atlas`
- `hub/internal/atlasclient`
- `hub/internal/orbita`
- `hub/migrations/control`
- `hub/deploy/import_hiveplace_collection.sh`
- `hub/api/openapi.yaml`

### Data

Clientes/aplicações, serviço/produto/oferta, DAG versionado, perfil técnico por cliente, contratos com vigência, snapshots, publicação e projeções, lote de importação em staging.

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

P-02 define portfólio e perfis legados reais; P-03 regras comerciais; P-05 equivalência. Artefatos usam fixtures sintéticas, sem declarar os 232 endpoints homologados.
