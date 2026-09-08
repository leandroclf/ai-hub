# Design: Qualificação por requisito e evidência reproduzível

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/qualificacao-integrada-r2/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Converter o alvo de 98 requisitos em entregas demonstráveis e impedir que teste superficial, manifesto de referência ou screenshot se torne prova de produção.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Converter o alvo de 98 requisitos em entregas demonstráveis e impedir que teste superficial, manifesto de referência ou screenshot se torne prova de produção. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Prova por cenário e SHA

Decision: Rastreabilidade inclui alvo, finding, tarefa, teste, ambiente e resultado; conservar falhas e limites.

Rationale: Corrige declaração de pronto baseada no nome do teste ou em manifests.

Trade-offs: Gera trabalho de evidência, mas evita regressão silenciosa em garantias críticas.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Oráculos independentes

Decision: Fixture de provedor devolve principal/marcador; banco/ledger e receiver de webhook verificam invariantes completas.

Rationale: Teste não apenas repete a implementação defeituosa.

Trade-offs: E2E depende de bootstrap reproduzível; collection privada e banco histórico não são fixtures.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: OpenSpec incremental

Decision: R2 usa capabilities ADDED próprias, ligadas aos 98 IDs; não arquivar v4 como implementada.

Rationale: Preserva alvo e histórico sem depender de MODIFIED contra base canônica ausente.

Trade-offs: T-R2-05 reconcilia futuras specs canônicas; validade estrutural não certifica entrega de software.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| openspec/changes/hub-interoperabilidade-v4 | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| IMPLEMENTATION_AUDIT.md | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/test/e2e/e2e_test.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/atlas/handlers_test.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/evidence | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| docs/openspec-docs | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Registrar snapshot, contratos, mudanças e fixtures sintéticas.
2. Executar cenários e falhas com oráculos de conteúdo/identidade/custódia.
3. Vincular prova a requisito/tarefa e reabrir diferenças.
4. Validar gates de perfil e reconciliar baseline sem fechar itens pendentes.

## Error Flows

- R2-QUA-01: Decisão ainda pendente — pendência permanece aberta; não presume avanço por commit ou lembrete.
- R2-QUA-02: Regressão financeira — falha se terceira chamada externa ocorrer.
- R2-QUA-03: P0 reaberto — ativação do fluxo afetado é bloqueada até corrigir ou manter explicitamente indisponível.
- R2-QUA-04: Regressão posterior — status reabre com referência ao novo SHA, sem apagar prova histórica nem reduzir requisito.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

Matriz requisito → gap → change → cenário → tarefa → execução/artefato/SHA; fixtures sintéticas; evidências sanitizadas; decisões com estado e responsável.

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

Executar cenários `R2-QUA-01` a `R2-QUA-04` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Manter v4 como baseline normativo com lacunas abertas. R2 acrescenta deltas, não declara a v4 implementada. Reconciliação futura da árvore canônica deve preservar IDs e separar aceitação de especificação de pronto operacional.

## Rollback Strategy

Reabrir status quando regressão invalidar prova; manter histórico de evidências e supersessão. Não apagar falha de teste ou reduzir requisito para obter verde.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Qualidade, Liderança técnica e SRE responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: Sem pré-requisito de implementação para iniciar; integrações específicas descritas abaixo. Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-01 a P-11 permanecem com responsáveis de função, sem nomes inventados. Falta de decisão limita ativação específica; não é justificativa para declarar funcionalidade incompleta como pronta. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
