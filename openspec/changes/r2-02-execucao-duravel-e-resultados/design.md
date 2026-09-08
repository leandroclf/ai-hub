# Design: Custódia, execução única e resultado final correto

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/execucao-duravel-e-resultados/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Uma obrigação recuperável por aceite, um único dono de despacho e uma representação final proveniente do provedor e durável antes da resposta.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Uma obrigação recuperável por aceite, um único dono de despacho e uma representação final proveniente do provedor e durável antes da resposta. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Intenção e efeito local atômicos

Decision: Persistir protocolo+intent; operação/tentativa antes de envio; observação+resultado+outbox no commit da operação; inbox+efeito antes do ACK.

Rationale: Fecha janelas de perda em F-04/F-05/F-08 sem trocar broker.

Trade-offs: Mais escritas necessárias; batchear relay, indexar agenda e manter transações curtas; rede do provedor fora de transação.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Fencing e reconciliação

Decision: CAS/epoch/lease duráveis em cada dono de operação; resultado de conflito determina se pode enviar, não ON CONFLICT cego.

Rationale: Redelivery e recovery não duplicam efeito.

Trade-offs: Parceiro sem idempotência/consulta pode ficar UNKNOWN acionável; não existe garantia end-to-end exactly-once inventada.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: Materialização por perfil

Decision: Guardar bytes completos/hash/versão do final e servir em POST/GET/webhook; payload sempre vem de observação validada do provedor.

Rationale: Corrige eco/resultado ausente e evita drift por reserialização após upgrade.

Trade-offs: Conserva armazenamento por retenção; perfil v1 permanece explicitamente suportado.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 4: Elegibilidade temporal conservadora

Decision: Manter EXE-11 rígido e exigir prova de confirmação durável antes do limite; resolver T-R2-01 antes de considerar deadline pronto.

Rationale: UPDATE condicional por estado/versão sozinho não prova cumprimento temporal.

Trade-offs: Mecanismo físico deve ser qualificado com commit lento e incerteza de relógio; se não houver prova, não emitir sucesso.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/internal/orbita | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/cometa/executor.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/cometa/store.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/outbox | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/queue | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/dispatch/contract.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/migrations/core/0001_init.sql | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Validar contrato e conservar aceite/snapshot/intenção; reservar saldo se estrito.
2. Reivindicar operação/tentativa com fencing e executar efeito fora da transação.
3. Validar e conservar observação/resultado/outbox do Cometa.
4. Arbitrar prazo/final na Órbita e materializar resposta; eventos seguintes usam inbox.

## Error Flows

- R2-EXE-01: Commit incerto — o Hub resolve pela autoridade de idempotência, sem afirmar não execução nem criar outro efeito.
- R2-EXE-02: Erro após aceite — o erro expõe UUID e consulta quando o aceite é conhecido; retry usa a mesma chave.
- R2-EXE-03: Fencing antigo — sua ação é rejeitada; lease expirada não prova ausência de efeito já enviado.
- R2-EXE-04: Commit falha — Cometa não declara fato durável; mantém obrigação recuperável e sem novo submit automático.
- R2-EXE-05: Schema desconhecido — original sanitizado e metadados ficam em quarentena; alerta e replay autorizado preservam a identidade de origem.
- R2-EXE-06: Resultado tardio custoso — GET e webhook conservam o erro final; Libra considera apenas o custo elegível e a contestação.
- R2-EXE-07: Cliente desconecta — custódia/reconciliação continua com contexto interno limitado; retry e GET recuperam a mesma identidade.
- R2-EXE-08: Resultado pendente ou expirado — recebe 200 com estado contratado local; consulta não faz polling no provedor nem mascara atraso como sucesso.
- R2-EXE-09: Callback antes da associação — recibo é reconciliado sem perda ou segunda operação.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

Comando durável, dispatch lease/epoch, tentativa pré-envio, observação externa, resultado por versão, inbox/outbox, relógio/decisão de deadline e obrigação de reconciliação. Tabelas novas são propostas, não existentes.

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

Executar cenários `R2-EXE-01` a `R2-EXE-09` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Adicionar estruturas por migration nova, sem editar 0001 aplicado. Backfill com proveniência dos resultados legados; não reconstruir resposta do provedor a partir da entrada. Pedidos sem prova viram LEGACY_UNVERIFIED e seguem tratamento assistido.

## Rollback Strategy

Expand/contract com leitura das duas versões durante transição; parar novos despachos antes de recuar worker. Preservar intents, inbox, epochs e bytes finais. Não reapresentar operação externa para recuperar migração.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Engenharia de Core e Dados responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: r2-01-identidade-e-isolamento Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-07 fecha SLAs e elegibilidade temporal comercial. A proposta conserva deadline rígido da v4; não inventa tolerância para resultado tardio. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
