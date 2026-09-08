# Design: Ambientes completos, disponibilidade, escala e observabilidade

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/operacao-elastica-e-observavel/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Subir a plataforma completa, escalar com isolamento e manter operações elegíveis durante falhas de dependências não autoritativas.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Subir a plataforma completa, escalar com isolamento e manter operações elegíveis durante falhas de dependências não autoritativas. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Papéis de runtime isolados

Decision: Separar admissão/leitura, APIs diretas/reserva e workers nos mesmos domínios; probes/pools/scalers por capacidade.

Rationale: Impede backlog ou writer afetar toda a superfície elegível.

Trade-offs: Mais Deployments/configurações; templates compartilhados evitam divergência e não criam novas autoridades.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Laboratório e overlays completos

Decision: Compose completo, kind laboratorial, overlays dev/hom/ppd/prd, identidade e telemetria provisionadas.

Rationale: Transforma manifests ilustrativos em instalação reproduzível.

Trade-offs: kind não comprova HA física; promoção cloud exige P-01 e ensaios representativos.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: Escala em camadas

Decision: HPA/KEDA → nós Karpenter → células por política Atlas/Crossplane; headroom e limite de conexões governam o conjunto.

Rationale: Atende crescimento sem chamados rotineiros dentro do envelope aprovado.

Trade-offs: Quotas físicas/budget não são infinitos; SQL não escala por replicação de pod; controlar dono de recurso IaC.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 4: Telemetria não autoritativa

Decision: Prometheus/Loki/Alloy/Grafana/OTel/Tempo com buffers limitados e SLIs de custódia/pressão/contrato.

Rationale: Aferição útil sem colocar observabilidade no caminho de bloqueio do negócio.

Trade-offs: Logs/traces amostrados não substituem ledger/auditoria; cardinalidade precisa orçamento.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/deploy | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/cmd | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/platform | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/atlasclient | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/queue | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/objectstore | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Subir dependências/CRDs/configuração e migrations por ordem documentada.
2. Iniciar papéis com probes próprias, projeções válidas e identidades por ambiente.
3. Escalar por sinais de capacidade e antecipar nós/células sem duplicar autoridade.
4. Observar, drenar/manter e recuperar com SLIs e provas por perfil.

## Error Flows

- R2-OPE-01: Cache desligado — serviços não falham por dependência do cache opcional.
- R2-OPE-02: Produção sem perfil — gate bloqueia ativação e mostra pendências sem impedir desenvolvimento local.
- R2-OPE-03: Provedor fora — não reinicia todos os pods; circuito/quotas contêm apenas a capacidade afetada.
- R2-OPE-04: Quota esgotada — estado registra impedimento e novas admissões são limitadas; não rouba reserva de clientes existentes.
- R2-OPE-05: Falha parcial de provisionamento — não atribui tráfego até recursos e canário estarem qualificados; retry de reconcile não duplica dono.
- R2-OPE-06: Todas cópias autoritativas indisponíveis — Hub informa indisponibilidade sem 202 fictício; preserva repetição pela mesma chave quando voltar.
- R2-OPE-07: Backend de observabilidade fora — buffer/export limitado não bloqueia thread de negócio; perda de telemetria é sinalizada e não altera ledger/auditoria duráveis.
- R2-OPE-08: Atualização de dependência — registra suporte/licença, lock/digest, scan e ensaios de contrato; referência de versão não vira aprovação automática.
- R2-OPE-09: Uso crítico — ativação é bloqueada para esse perfil, com contingência e pendências explicitadas.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

Placement e epoch por tenant/contrato/célula; estado de provisionamento e qualificação; métricas/alertas provisionados; parâmetros por ambiente com segredo externo; orçamento de pools e reservas por célula.

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

Executar cenários `R2-OPE-01` a `R2-OPE-09` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Criar overlays e papéis de runtime preservando cinco aplicações lógicas; implantar controllers antes de seus recursos; migrar filas por célula sem duas autoridades. Executar prova local/kind antes de promoção.

## Rollback Strategy

GitOps reverte imagem/config compatíveis; drenar workloads e conservar filas/DB/objetos; Terraform/Crossplane não podem destruir recursos por rollback. Perda de autoridade cerca a célula antiga antes de failover.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Plataforma, SRE e Engenharia responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-01/P-08/P-10/P-11 bloqueiam ativação cloud/crítica, não a entrega do laboratório completo. Distinguir kind como laboratório de Kubernetes gerenciado remoto. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
