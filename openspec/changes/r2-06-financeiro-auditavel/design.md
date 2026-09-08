# Design: Medição, saldo estrito, ledger e fechamento

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/financeiro-auditavel-e-conciliacao/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Explicar cada receita, custo, reserva, ajuste e pagamento a partir de contrato e evidência, sem redelivery alterar valor.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Explicar cada receita, custo, reserva, ajuste e pagamento a partir de contrato e evidência, sem redelivery alterar valor. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Decimal e unidade econômica

Decision: Valores exatos, chave semântica por obrigação e incidência de compra/venda versionada.

Rationale: Corrige placeholder e colisão de custos de múltiplos passos.

Trade-offs: Escolha da representação decimal deve ser coerente em Go/JSON/Postgres/exportação; não conversão intermediária por float.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Conta única e holds

Decision: Saldo/limites e reservas na autoridade Libra; capturas debitam e UNKNOWN mantém hold até evidência.

Rationale: Evita limite cumulativo fictício e liberação antecipada.

Trade-offs: Recusa estrita quando autoridade indisponível é necessária; ledger não pode ser cache eventual.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: Journal balanceado

Decision: Lotes de partidas por moeda, origem e ajustes compensatórios; fechamento depende de completude e watermark.

Rationale: Permite explicar, conciliar e exportar cada valor.

Trade-offs: Regra comercial e ERP seguem P-03/P-06; ensaiar com fixtures sem inventar aprovação de tarifa.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/internal/libra | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/libraclient/client.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/atlas/store.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/migrations/finance | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/orbita/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Receber fato custodiado com identidade e snapshot econômico.
2. Validar unidade elegível, deduplicar e aplicar reserva/hold/captura conforme marco.
3. Confirmar apuração/journal balanceado e inbox atomicamente.
4. Reconciliar período completo, fechar e exportar com identidade/hash/recibo.

## Error Flows

- R2-FIN-01: Compra por submit — custo e ausência de receita seguem marcos distintos, sem preço constante de simulador.
- R2-FIN-02: Precisão — resultado decimal é exato e arredondamento é único no marco contratado.
- R2-FIN-03: Timeout com efeito incerto — hold permanece até evidência de cancelamento/ausência/custo; duplicata não libera nem captura duas vezes.
- R2-FIN-04: Franquia final concorrente — uma usa franquia e a outra segue regra de excedente ou recusa explicitamente contratada.
- R2-FIN-05: Estorno autorizado — partidas compensatórias preservam original, razão e correlação.
- R2-FIN-06: Contestação de SLA — valor, evidência, disputa e ajuste ficam separados da resposta final imutável do cliente.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

Contratos de compra/venda e incidência versionados, unidades econômicas, contas por moeda, holds incertos, journal balanceado, lotes de fechamento/exportação e recibos de conciliação.

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

Executar cenários `R2-FIN-01` a `R2-FIN-06` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Migrar decimal sem converter por float; carregar fatos antigos em lote de abertura com proveniência LEGACY_UNVERIFIED onde snapshot faltar. Não inventar tarifas históricas. Bloquear fechamento definitivo até reconciliação e aprovação.

## Rollback Strategy

Não apagar lançamentos; estorno compensatório com origem e razão. Recuo de software preserva formato novo de ledger e bloqueia apuração incompatível. Liberar reserva somente por fato seguro.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Financeiro, Comercial e Engenharia Libra responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados, r2-04-catalogo-produtos-e-contratos Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-03 e P-06 aprovam tarifas, incidência, arredondamento e ERP. Corrigir saldo fail-open e congelar contratos não depende da definição de preço comercial real. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
