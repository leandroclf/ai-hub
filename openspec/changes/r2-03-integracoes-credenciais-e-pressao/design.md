# Design: Adaptadores reais, credenciais, polling, callbacks e pressão

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/integracoes-credenciais-e-pressao/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Executar contratos homologados do provedor com credencial correta e capacidade adaptativa compartilhada, preservando cada observação e entrega.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Executar contratos homologados do provedor com credencial correta e capacidade adaptativa compartilhada, preservando cada observação e entrega. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Adapter e credencial efetivos

Decision: Módulo Go homologado por capacidade; configuração declarativa limitada; resolver segredo do binding real, não provider_account genérica.

Rationale: Une metadados publicados à chamada executada.

Trade-offs: Especialidades SOAP/SFTP/gRPC demandam adapter próprio quando contratadas; nenhuma oferta é ativada só por URL.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Agenda e recebimento duráveis

Decision: Polling/callback são observadores de uma mesma operação; persistir recibo autenticado e só então ACK; agenda usa claim/fencing e política versionada.

Rationale: Evita perda/duplicidade sob réplicas e corrida callback/poll.

Trade-offs: Armazenar evidências custa espaço; expurgo depende de custódia/contestações.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: AIMD com autoridade por domínio

Decision: Aplicar feedback, créditos/leases agregados, reservas de reconciliação e fairness; parâmetros de exemplo são apenas para ensaio.

Rationale: Pressão responde a degradação e crescimento do parceiro, sem multiplicar quota por pod.

Trade-offs: T-R2-04 precisa ensaio entre células/partição; coordenação tem custo, mas não é substituída por Redis autoritativo.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 4: Entrega por identidade versionada

Decision: Destination_id/version/key_id e tenant fixam entrega; bytes vêm da representação final; HMAC com timestamp e claims por tentativa.

Rationale: Corrige lookup de chave por URL e reentrega ambígua.

Trade-offs: Rotação precisa política de verificação da chave histórica; não muda destino retroativamente.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/internal/cometa | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/providerauth/client.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/pulsar | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/atlasclient/client.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/migrations/control | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/migrations/core | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Carregar adapter/perfil e binding/version elegíveis do snapshot/projeção.
2. Obter crédito de capacidade e posse válida, autenticar e enviar tentativa.
3. Conservar retorno ou agendar observações autenticadas por polling/callback; aplicar TTL/prazos.
4. Entregar fato ao domínio; Pulsar agenda/assina corpo final por destino/versão.

## Error Flows

- R2-INT-01: Protocolo especializado — capacidade fica não disponível; suporte futuro exige adapter e ensaio específico, sem simulação silenciosa via REST.
- R2-INT-02: Rotação e cache — cache antigo não é reutilizado indevidamente; operação já aceita conserva conta, snapshot e pagador.
- R2-INT-03: Cofre indisponível — operação afetada aguarda/recusa dentro da política sem credencial inválida; outras contas elegíveis continuam.
- R2-INT-04: Falha de recibo — ACK 2xx não é emitido; o parceiro pode retransmitir conforme contrato.
- R2-INT-05: Timeout após envio — somente consulta/reconciliação segura é permitida antes de autorizar nova execução.
- R2-INT-06: Ruído de cliente — B mantém SLO do perfil e reservas; A recebe contenção/recusa explícita em vez de consumir capacidade de B.
- R2-INT-07: Duplicata no receptor — event_id/result_version identificam o mesmo final e assinatura/timestamp são verificáveis.
- R2-INT-08: Coorte ainda aberta — painel mostra elegíveis, abertos, cumpridos, vencidos e exclusões, sem contar abertos como sucesso.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

Perfil de adapter versionado, binding/secret_version, capacity_domain/lease, retry_until, agenda/recibo de polling, inbox de callback, delivery e key_id versionados; segredos resolvidos no cofre, tokens somente L1 privado.

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

Executar cenários `R2-INT-01` a `R2-INT-08` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Introduzir adapters por capacidade e manter provider-sim só no perfil de ensaio. Migrar bindings explícitos e invalidar cache por conta antigo; em operações existentes, fixar a identidade da conta sem trocar o pagador durante rotação.

## Rollback Strategy

Rollback de adapter apenas para versão compatível com operações em voo; suspender novos envios quando incompatível. Não voltar a token compartilhado por conta nem simular mTLS. Manter recebimento e reconciliação.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Integrações e Engenharia de Core responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-05 contém garantias reais dos parceiros; P-07 define TTL/SLA; P-11 os envelopes máximos. Sem confirmação de idempotência não habilitar retry de efeito ambíguo. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
