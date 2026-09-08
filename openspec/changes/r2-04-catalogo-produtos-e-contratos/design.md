# Design: Catálogo operável, produtos compostos e contratos por cliente

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/catalogo-produtos-e-contratos-versionados/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Separar serviço, produto, oferta, contrato de integração e adapter; publicar versões validadas que governam admissão e execução paralela.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Separar serviço, produto, oferta, contrato de integração e adapter; publicar versões validadas que governam admissão e execução paralela. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Versões publicadas e snapshots

Decision: Atlas publica versões imutáveis; contratos/ofertas e DAG/perfis são congelados no aceite; segurança de revogação conserva prevalência definida.

Rationale: Preço ou contrato novo não modifica pedido antigo.

Trade-offs: Mais modelos e migrations; exige indexar vigência e evitar vinculação ambígua.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: DAG limitado na Órbita

Decision: Manter agregação/composição em autoridade existente, com steps duráveis e paralelismo limitado.

Rationale: Atende produto composto sem microserviço por etapa.

Trade-offs: Não é motor BPMN arbitrário; limitar fan-out/transformações e registrar compensação incompleta.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: Importação em staging

Decision: Lote sanitizado com diff, identidade estável e homologação antes de ativar oferta.

Rationale: Remove dependência de delete global e contagem enganosa de endpoints.

Trade-offs: Exige resolver diferenças do catálogo existente e coleção externa; não adivinhar campos/segredos ausentes.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 4: Projeções válidas

Decision: Evento de publicação indica versão/hash; workloads mantêm snapshot/projeção limitada e recuperável.

Rationale: Reduz consulta de controle por pedido e permite continuidade seletiva.

Trade-offs: Validade/revogação e bootstrap devem ser ensaiados; material expirado não vira autorização.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/internal/atlas | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/atlasclient | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/orbita | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/migrations/control | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/deploy/import_hiveplace_collection.sh | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/api/openapi.yaml | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Criar/importar rascunhos e definir serviço/produto/perfil/contratos/oferta.
2. Validar grafo, capacidade, credencial, schemas e vigência; simular sem efeitos.
3. Publicar versão/hash e distribuir projeção para execução.
4. Admitir com snapshot; executar passos independentes em paralelo e dependentes após pré-condições.

## Error Flows

- R2-CAT-01: Edição concorrente — servidor detecta versão desatualizada e oferece diff/recarregamento sem sobrescrever trabalho alheio.
- R2-CAT-02: Sem combinação de SLA viável — recebe bloqueios específicos e oferta não é ativada.
- R2-CAT-03: Fan-out abusivo — configuração/pedido é recusado antes de exceder recursos de outros clientes.
- R2-CAT-04: Compensação incompleta — protocolo informa estado contratado e gera obrigação de reconciliação, sem declarar rollback integral fictício.
- R2-CAT-05: Upgrade de schema — mudança incompatível exige nova versão e migração; protocolos v1 conservam seu resultado.
- R2-CAT-06: Revogação de segurança — política de revogação de segurança prevalece para novos acessos ao segredo; preserva histórico sem reutilizar credencial proibida.
- R2-CAT-07: Credenciais na collection — valores secretos não aparecem no lote, logs ou UI; somente referências e metadados permitidos são conservados.
- R2-CAT-08: Capacidade insuficiente — registra provisionamento automático e só ativa após qualificação; não aponta tráfego para célula incompleta.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

Clientes/aplicações, serviço/produto/oferta, DAG versionado, perfil técnico por cliente, contratos com vigência, snapshots, publicação e projeções, lote de importação em staging.

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

Executar cenários `R2-CAT-01` a `R2-CAT-08` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Transformar catálogo atual em rascunhos de origem rastreável; manter IDs e consultas históricas. Campos herdados por tenant viram defaults explícitos, nunca autorização a qualquer serviço. Reconciliar importação com diff sem limpeza global.

## Rollback Strategy

Reverter ponteiro da publicação para versão anterior; pedidos em voo permanecem no snapshot antigo. Não editar versão publicada nem apagar contratos históricos para desfazer catálogo.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Produto, Atlas e Engenharia de Core responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-02 define portfólio e perfis legados reais; P-03 regras comerciais; P-05 equivalência. Artefatos usam fixtures sintéticas, sem declarar os 232 endpoints homologados. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
