# Proposal: Arquivos, propriedade de dados, retenção e recuperação

## Change ID

`r2-07-objetos-dados-e-retencao`

## Status

Draft — pacote completo para revisão/implementação; nenhuma tarefa executada nesta entrega. Responsável funcional proposto: **Dados, Segurança e Plataforma**.

## Context

O repositório no commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` implementa uma fatia inicial. Esta change acrescenta critérios verificáveis para fechar lacunas sem reescrever nem declarar atendida a baseline v4. Leia [explore.md](explore.md).

## Why

Manter conteúdo volumoso fora do caminho de mensagens, preservando autorização, integridade, vínculo durável e recuperação reconciliada.

## Problem

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-29 | P1 | Arquivos grandes não participam dos fluxos reais | [hub/internal/objectstore/objectstore.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/objectstore/objectstore.go#L43) |
| F-30 | P1 | Persistência sem isolamento de papéis, expurgo e auditoria durável | [hub/deploy/postgres-init/01-init.sh:8](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/postgres-init/01-init.sh#L8) |

## Goals

- Upload direto e referência imutável por tenant
- Custódia de resultado volumoso
- Retenção por classe e obrigações abertas
- Propriedade, RLS e continuidade de leitura
- Restauração reconciliada sem duplicar efeito

## Non-Goals

- Reescrever todo o Hub, trocar linguagens ou criar um microserviço por função.
- Resolver preços, credenciais, contas cloud ou criticidade por suposição.
- Apagar histórico, alterar contratos publicados retroativamente ou atestar disponibilidade sem ensaio.

## Users / Actors Impacted

Clientes e aplicações consumidoras, operadores/desenvolvedores com papéis explícitos, provedores homologados e a equipe Dados, Segurança e Plataforma.

## Scope

### In scope

- R2-DAD-01 — Upload direto e referência imutável por tenant (baseline: DAD-05, COM-02, SEG-02).
- R2-DAD-02 — Custódia de resultado volumoso (baseline: DAD-05, DAD-04, DAD-09, OPE-11).
- R2-DAD-03 — Retenção por classe e obrigações abertas (baseline: DAD-06, DAD-08, OPE-11).
- R2-DAD-04 — Propriedade, RLS e continuidade de leitura (baseline: DAD-01, DAD-07, DAD-10, SEG-03).
- R2-DAD-05 — Restauração reconciliada sem duplicar efeito (baseline: DAD-07, OPE-11, OPE-15).

### Out of scope

Aplicar infraestrutura remota, publicar release ou executar operação financeira real nesta entrega documental. Capacidades não homologadas permanecem indisponíveis até suas provas.

## What Changes

Nova capability incremental `objetos-dados-e-retencao` com 5 requisitos R2 e 15 cenários. Adiciona contrato de comportamento e tarefas para os arquivos abaixo; não remove os requisitos anteriores.

## Product Requirements Summary

Atender os cenários normativos em [spec](specs/objetos-dados-e-retencao/spec.md), mantendo a rastreabilidade com COM-02, DAD-01, DAD-04, DAD-05, DAD-06, DAD-07, DAD-08, DAD-09, DAD-10, OPE-11, OPE-15, SEG-02, SEG-03. A forma detalhada do console, APIs, dados e operação está nos documentos compartilhados da revisão.

## Business Rules

- Nenhuma mudança de versão/prazo/credencial pode reinterpretar silenciosamente protocolo já aceito.
- Estado do cliente, observação externa, entrega e fato financeiro têm autoridades e relógios distintos.
- Sucesso/ACK exige a custódia definida na spec; conhecimento incompleto não se transforma em resultado presumido.

## Affected Capabilities

- `objetos-dados-e-retencao` (ADDED).
- Baseline afetada por rastreabilidade: COM-02, DAD-01, DAD-04, DAD-05, DAD-06, DAD-07, DAD-08, DAD-09, DAD-10, OPE-11, OPE-15, SEG-02, SEG-03. Esses IDs não são removidos nem renomeados.

## Expected Impact

### Code

- `hub/internal/objectstore`
- `hub/internal/orbita`
- `hub/migrations`
- `hub/deploy/postgres-init/01-init.sh`
- `hub/deploy/terraform/main.tf`

### Data

FileRef por tenant/objeto/versão/hash, sessões de upload, estados de validação, referências/pins de retenção, tombstones, trilha de expurgo, papéis/RLS e inventário de restauração.

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

P-04 define classes/regiões/retenção; P-08 define domínio de recuperação. Nenhum período legal ou direito de custódia é presumido pela revisão.
