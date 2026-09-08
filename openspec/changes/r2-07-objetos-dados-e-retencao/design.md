# Design: Arquivos, propriedade de dados, retenção e recuperação

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/objetos-dados-e-retencao/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Manter conteúdo volumoso fora do caminho de mensagens, preservando autorização, integridade, vínculo durável e recuperação reconciliada.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Manter conteúdo volumoso fora do caminho de mensagens, preservando autorização, integridade, vínculo durável e recuperação reconciliada. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: FileRef imutável

Decision: S3 armazena bytes grandes, catálogo por tenant/versão/hash faz custódia e autorização; presigned URL separada.

Rationale: Reduz payload e mantém igualdade do resultado público no tempo.

Trade-offs: Sem transação única S3+PG; confirmação/reconciliação e pins são necessários.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Retenção orientada a obrigações

Decision: Classes e holds protegem resultado/recibo/idempotência até conclusão e contestação; expurgo deixa evidência mínima permitida.

Rationale: Evita apagar único resultado não entregue ou reexecutar evento antigo.

Trade-offs: P-04 aprova prazos/dados/regiões; expurgo irreversível exige dry-run e teste.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: Restore cercado

Decision: Recuperar dados/objetos e reconciliar antes de liberar egress; papéis e RLS por domínio.

Rationale: Backup antigo não autoriza nova chamada ao parceiro.

Trade-offs: RTO/RPO dependem de topologia e ensaio; leitura por réplica requer atualidade/prova do final.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/internal/objectstore | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/orbita | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/migrations | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/deploy/postgres-init/01-init.sh | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/deploy/terraform/main.tf | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Autorizar sessão de upload e transferir bytes diretamente ao storage.
2. Validar tamanho/tipo/checksum/proprietário e criar referência imutável.
3. Consolidar resultado somente com custódia comprovada e proteger referências por obrigação.
4. Expurgar apenas elegíveis e ensaiar restore cercado com reconciliação.

## Error Flows

- R2-DAD-01: Multipart incompleto — referência não fica READY; sessão pode ser retomada/expirada conforme política.
- R2-DAD-02: Objeto sem commit do protocolo — liga objeto somente à obrigação correta ou marca órfão para política segura, sem reexecutar provedor.
- R2-DAD-03: Mensagem antiga — tombstone/dedup impede novo efeito e produz diagnóstico.
- R2-DAD-04: Isolamento de domínio — nega escrita fora da autoridade mesmo que aplicação tenha bug.
- R2-DAD-05: Falha regional — perfil não é qualificado nem ativado até prova de custódia e fencing compatíveis.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

FileRef por tenant/objeto/versão/hash, sessões de upload, estados de validação, referências/pins de retenção, tombstones, trilha de expurgo, papéis/RLS e inventário de restauração.

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

Executar cenários `R2-DAD-01` a `R2-DAD-05` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Criar catálogo de objetos e permissões sem reescrever finais históricos; materializar referência somente depois de validar existência e dono. Backfill separado e retomável; nunca criar file_ref com objeto não confirmado.

## Rollback Strategy

Manter objetos e pins ao recuar; expurgo não tem rollback lógico, portanto executar só após dry-run, janela aprovada e teste de restauração. Proteções contra exclusão permanecem ativas.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Dados, Segurança e Plataforma responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-04 define classes/regiões/retenção; P-08 define domínio de recuperação. Nenhum período legal ou direito de custódia é presumido pela revisão. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
