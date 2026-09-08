# Proposal: Medição, saldo estrito, ledger e fechamento

## Change ID

`r2-06-financeiro-auditavel`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Financeiro, Comercial e Engenharia Libra**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Explicar cada receita, custo, reserva, ajuste e pagamento a partir de contrato e evidência, sem redelivery alterar valor.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-23 | P0 | Custo e receita não usam contratos econômicos congelados | [hub/internal/libra/consumers.go:17](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/consumers.go#L17) |
| F-24 | P0 | Saldo estrito não contabiliza consumo já capturado | [hub/internal/libra/store.go:73](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/store.go#L73) |
| F-25 | P1 | Ledger, precisão monetária e fechamento não estão implementados | [hub/internal/libra/store.go:27](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/store.go#L27) |

## Goals

- Compra, venda e incidência por snapshot
- Medição exata e deduplicação econômica
- Saldo estrito e retenção de incerteza
- Planos, franquias e política de produto
- Ledger imutável e ajustes compensatórios
- Fechamento, reconciliação e integração financeira

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Financeiro, Comercial e Engenharia Libra.

## Scope

### In scope

- R2-FIN-01 — Compra, venda e incidência por snapshot (baseline: FIN-01, FIN-02, FIN-03, FIN-05, FIN-11).
- R2-FIN-02 — Medição exata e deduplicação econômica (baseline: FIN-04, FIN-05, FIN-09).
- R2-FIN-03 — Saldo estrito e retenção de incerteza (baseline: FIN-06, FIN-10, DAD-03, DAD-11).
- R2-FIN-04 — Planos, franquias e política de produto (baseline: FIN-02, FIN-05, FIN-09, CAT-07).
- R2-FIN-05 — Ledger imutável e ajustes compensatórios (baseline: FIN-07, FIN-10).
- R2-FIN-06 — Fechamento, reconciliação e integração financeira (baseline: FIN-08, FIN-10, FIN-11).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `financeiro-auditavel-e-conciliacao` com 6 requisitos R2 e 19 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/financeiro-auditavel-e-conciliacao/spec.md), mantendo a rastreabilidade com CAT-07, DAD-03, DAD-11, FIN-01, FIN-02, FIN-03, FIN-04, FIN-05, FIN-06, FIN-07, FIN-08, FIN-09, FIN-10, FIN-11. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `financeiro-auditavel-e-conciliacao` (ADDED).
- Baseline afetada por rastreabilidade: CAT-07, DAD-03, DAD-11, FIN-01, FIN-02, FIN-03, FIN-04, FIN-05, FIN-06, FIN-07, FIN-08, FIN-09, FIN-10, FIN-11. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/internal/libra`
- `hub/internal/libraclient/client.go`
- `hub/internal/atlas/store.go`
- `hub/migrations/finance`
- `hub/internal/orbita/handlers.go`

### Data

Contratos de compra/venda e incidência versionados, unidades econômicas, contas por moeda, holds incertos, journal balanceado, lotes de fechamento/exportação e recibos de conciliação.

### APIs / Contracts

Contrato público versionado e APIs administrativas/internas conforme [contratos](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md). Métodos novos do console estão em [jornadas](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md); não são endpoints existentes presumidos.

### Integrations

Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados, r2-04-catalogo-produtos-e-contratos Contratos compartilhados são acordados antes da implementação consumidora; mocks não contam como integração pronta.

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

P-03 e P-06 aprovam tarifas, incidência, arredondamento e ERP. Corrigir saldo fail-open e congelar contratos não depende da definição de preço comercial real.
