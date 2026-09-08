# Design: Console administrativo completo e orientado às jornadas

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/console-administrativo-operavel/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Permitir que operadores configurem, publiquem, diagnostiquem e conciliem o Hub por jornadas completas, com dados do servidor e permissões efetivas.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Permitir que operadores configurem, publiquem, diagnostiquem e conciliem o Hub por jornadas completas, com dados do servidor e permissões efetivas. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Manter React/TypeScript SPA

Decision: Evoluir quatro páginas para rotas/jornadas, componentes de formulário/detalhe/tabela e cliente tipado validado.

Rationale: Reutiliza stack existente e foca gaps funcionais.

Trade-offs: Roteamento, consulta/invalidação e acessibilidade exigem trabalho; novo framework não entregaria APIs faltantes.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Fonte de verdade no servidor

Decision: Listas paginadas, ETag e comandos idempotentes; estado local só edição/visualização transitória.

Rationale: Refresh, concorrência e vários operadores preservam integridade.

Trade-offs: Depende das APIs de domínio; mocks servem desenvolvimento, não aceite integrado.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: Interfaces por papel e tarefa

Decision: Navegação por jornadas com escopo, ambiente, validação, diff e justificativas de ações com efeitos.

Rationale: Evita formulários de IDs livres e superusuário implícito.

Trade-offs: Menus não são segurança; servidor aplica cada permissão e teste negativo verifica bypass.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 4: Entrega incremental por jornada

Decision: Liberar módulos disponíveis por capacidade/permissão, cada um com provas reais de ponta a ponta.

Rationale: Permite avanço antes de todo backend ficar pronto sem declarar módulo mockado concluído.

Trade-offs: Tratar indisponível/pendente como estado explícito; não esconder backlog de APIs.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/admin-ui/src | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/admin-ui/package.json | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/admin-ui/vite.config.ts | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/atlas/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/orbita/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/libra/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/api/openapi-internal.yaml | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Entrar por identidade, recuperar rota/escopo e consultar lista do servidor.
2. Abrir detalhe/version e efetuar ação permitida com validação/conflito.
3. Acompanhar publicação/execução ou obrigação por timeline com dados reais.
4. Consultar SLA/entrega/financeiro conforme papel e conservar trilha de ações.

## Error Flows

- R2-ADM-01: Logout — não reaparecem dados, filtros sensíveis ou permissões do anterior.
- R2-ADM-02: Conflito de edição — UI apresenta conflito e diff/recarregamento sem sobrescrita silenciosa.
- R2-ADM-03: Suspensão — confirma impacto e razão; timeline histórica continua acessível ao papel autorizado.
- R2-ADM-04: Sem adapter — estado deixa claro que não está disponível para consumo.
- R2-ADM-05: Simulação sem efeitos — nenhuma chamada faturável real é feita; relatório identifica fixture e versões.
- R2-ADM-06: Pressão reduzida — vê limite efetivo, teto, latência, erro, backlog e motivo da redução, sem tratá-lo como limite fixo eterno.
- R2-ADM-07: SYNC incompatível — relatório aponta incompatibilidade e bloqueia publicação.
- R2-ADM-08: Filtro entre tenants — API e UI não revelam protocolos alheios.
- R2-ADM-09: Leitor administrativo — servidor nega; possuir leitura global não concede operação.
- R2-ADM-10: Consulta limitada — arquivo contém somente registros permitidos e exportação é auditada.
- R2-ADM-11: Estorno — lançamento compensatório aparece no histórico; original não desaparece.
- R2-ADM-12: Jornada integrada — estado persiste no servidor; evidência inclui papel, SHA, requests e screenshot sanitizado, não só renderização.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

UI não é autoridade: rascunhos, revisão concorrente, histórico e ações persistem no domínio. Sessão autenticada isolada por pessoa; preferência de filtro não contém segredo ou corpo sensível.

Campos, chaves, consultas/índices e autoridades estão na tabela do modelo lógico compartilhado. Persistência do domínio confirma estado e intenção/inbox na mesma transação quando exigido; nenhuma chamada de rede externa fica dentro de transação aberta. Escritas/joins entre autoridades são proibidos; projeções e IDs conservam contexto.

## Authentication and Authorization

Aplicar R2-SEG-01 a R2-SEG-05 a cada superfície envolvida. Escopo de tenant/aplicação é autenticado; privilégio administrativo individual não se transforma em acesso público global. Segredos e tokens não aparecem em retorno, evento ou logs. Papéis de aplicação/migração são distintos.

## Security and Privacy

Validar URLs e conexão efetiva, limites de tamanho/complexidade e dados permitidos. Auditoria durável de ações sensíveis com mascaramento e classe de retenção. Projeção ou cache inválido não amplia autorização. Usuário e workload só acessam recursos próprios da concessão/célula.

## Observability

### Logs

Eventos estruturados com correlação e estado/causa; não corpo sensível/segredo. Registrar falhas sem perder a obrigação de negócio.

### Metrics

Instrumentar as decisões/tempos relevantes aos cenários: contagem de recusas/erros, idade de obrigação, conflitos/duplicatas, execução/final/entrega ou mudança administrativa pertinente. Usar rota normalizada e dimensões de cardinalidade orçada, sem UUID como label.

### Tracing

Propagar contexto de trace e causation/event/command correlacionados; protocolo permanece identidade de negócio. Traces não são ledger nem prova única de custódia.

### Alerts

Alertar violação de invariante, obrigação envelhecida e dependência bloqueadora com runbook; acionar ensaio para provar detecção. Detalhamento em [operação](../../../docs/reviews/2026-09-07-r2/06-operacao-escala-e-observabilidade.md).

## Testing Strategy

Executar cenários `R2-ADM-01` a `R2-ADM-12` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Preservar acessos aos quatro cadastros com rotas novas e listagens reais; manter UI antiga restrita até migração, sem permitir escrita anônima. Introduzir por jornadas disponíveis e feature flags por permissão.

## Rollback Strategy

Reverter versão de UI compatível com APIs expandidas; nenhuma reversão altera resultado/contrato já publicado. Limpar estado sensível ao encerrar sessão ou trocar de tenant.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Produto, Frontend e UX responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: r2-01-identidade-e-isolamento Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

APIs operacionais dependem das changes 02/03/04/06/07/08. P-09 nomeia administradores. É possível construir navegação e listas cedo, mas nenhuma tela com mock será aceita como integração concluída. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
