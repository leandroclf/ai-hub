# Design: Hub de Interoperabilidade "Constelação" v4.0

## Context

O Hub intermedia clientes (tenants) e provedores externos, oferecendo um catálogo de serviços/produtos versionados, com admissão durável, execução em modo SYNC/ASYNC/AUTO, custódia de resultado, contratos comerciais rastreáveis e operação configurável. A v4 substitui, em relação à v3, o transporte obrigatório por fila no caminho SYNC por um despacho direto (comando/resposta HTTPS), mantendo a mesma máquina de estados, os mesmos invariantes de idempotência/deadline e a mesma custódia de resultado. Este documento descreve **como** a arquitetura de referência atende aos requirements das nove capabilities desta mudança, sem introduzir decisões não presentes nos capítulos-fonte (`docs/01_*.md` a `docs/09_*.md`).

Este design é ele próprio um artefato **documental**: não há código, manifesto ou ambiente de implantação associado a esta mudança. Ele orienta a implementação futura das tarefas descritas em `tasks.md`.

## Goals and Constraints

### Goals
- Descrever a arquitetura de referência (5 aplicações de negócio + gateway) e como cada invariante obrigatório (`docs/00_LEIA_PRIMEIRO.md`) é sustentado por essa arquitetura.
- Explicar a diferença de transporte entre SYNC (despacho direto) e ASYNC/AUTO (fila durável), preservando a mesma máquina de estados (ADR-19).
- Descrever o modelo de dados por domínio (hub_control / hub_core / hub_finance) e a estratégia de isolamento por célula.
- Descrever a estratégia de segurança de credenciais (SHARED_HUB / TENANT_DEDICATED), isolamento de tenant e prevenção de SSRF.
- Descrever a estratégia de observabilidade, escala automática e continuidade seletiva por dependência.

### Constraints
- Nenhuma tecnologia fora da já decidida em ARQ-03/04/05/06 pode ser introduzida sem uma nova decisão registrada (ex.: não adotar gRPC, Kafka, Vault, Temporal, mesh/Envoy como padrão).
- Redis nunca pode ser a única cópia de protocolo, resultado, idempotência, saldo, outbox, agenda, autorização ou quota global (DAD-10, invariante 20).
- O deadline do cliente (`client_sla_seconds`) é absoluto e não pode ser renovado por retry, migração ou reinício (EXE-11, invariante 11).
- Nenhuma decisão pendente (P-01 a P-11) pode ser tratada como resolvida por este design.

## Proposed Architecture

Cinco aplicações de negócio Go + um gateway, organizadas em **células** com capacidade reservada (ARQ-01, ARQ-02, OPE-08):

| Componente | Responsabilidade central |
|---|---|
| Portal (Kong) | TLS, autenticação de borda, quotas técnicas, rotas, limites de payload |
| Atlas | Catálogo, contratos, planos, vínculos de credenciais, placement, capacidade desejada, configuração publicada |
| Órbita | Admissão, protocolo, idempotência, grafo de composição/agregação, roteamento, estado final, resultado e consulta |
| Cometa | Operações/tentativas externas, adaptação por provedor, quotas de conta, callbacks, polling, evidência |
| Pulsar | Obrigações de webhook, agenda de entrega, assinatura, tentativas, recibos |
| Libra | Medição econômica, reservas estritas, apuração, ledger, conciliação, integração ERP |

Cada célula contém um conjunto próprio de deployments, filas, bancos `core`/`finance`, pools e orçamentos (OPE-08). Atlas é plano de controle fora do caminho por pedido: distribui projeções versionadas (catálogo, contratos, mapa tenant→célula, credenciais) e o tráfego estabelecido não o consulta a cada requisição (ARQ-02, DAD-11).

### Transporte SYNC vs. ASYNC/AUTO (ADR-19, COM-01, COM-06, EXE-14/15)

- **SYNC**: Órbita despacha um comando `DIRECT` idempotente via HTTPS/mTLS interno diretamente ao Cometa, após registrar durávelmente protocolo, idempotência e intenção. Cometa executa a chamada ao provedor, persiste a observação e retorna o fato à Órbita pela mesma chamada HTTPS. Órbita consolida, decide o prazo e responde ao cliente na mesma conexão — sem fila obrigatória no comando nem no final.
- **ASYNC/AUTO**: o mesmo conteúdo lógico de comando vai à fila (SQS dedicada à célula/classe) após commit da intenção `QUEUED`. Fatos posteriores (callback, polling, resultado) sempre usam SNS/SQS com outbox/inbox transacional, independentemente do modo de admissão.
- Em ambos os casos, o mecanismo de posse única de EXE-15 impede que uma intenção `DIRECT` seja simultaneamente tratada como comando `QUEUED`, e vice-versa.

## Technical Decisions

### Decision 1: Despacho direto para SYNC, fila apenas para ASYNC/AUTO e fatos
Decision: Órbita→Cometa em SYNC usa HTTPS interno idempotente com deadline propagado (COM-01); ASYNC usa SQS dedicada.
Rationale: Elimina a latência de fila obrigatória do caminho síncrono (ADR-19), mantendo os mesmos invariantes de durabilidade, idempotência e deadline via registros prévios ao efeito (DAD-09).
Trade-offs: Aumenta o acoplamento temporal entre Órbita e Cometa durante a chamada síncrona; exige orçamento de conexão reservado (OPE-09) e commits mais rápidos no caminho crítico.
Consequences: A publicação de SYNC exige capacidade reservada de chamada e commits (CAT-11); um provedor lento ou banco em failover produz erro/timeout contratual, nunca uma conversão silenciosa para 202 (EXE-14).

### Decision 2: Persistência por domínio e por célula, sem cluster global
Decision: PostgreSQL com três funções lógicas (`hub_control`, `hub_core`, `hub_finance`), cada célula com seu próprio core/finance (DAD-01, ADR-03).
Rationale: Limita contenção e falhas por tenant/carga; um contrato de saldo estrito tem uma única autoridade financeira ativa.
Trade-offs: Exige automação de placement/expansão (Crossplane) em vez de um único banco compartilhado; aumenta a complexidade operacional de migração entre células (DAD-11).
Consequences: Não há junção/FK entre domínios, mesmo no mesmo servidor; referências cruzadas são só IDs e projeções.

### Decision 3: UUIDv7 universal com idempotência por chave, não por identidade
Decision: `protocol_id` sempre UUIDv7 (RFC 9562), com `Idempotency-Key` obrigatória e resolução por hash semântico (EXE-01, EXE-16, ADR-20).
Rationale: Permite recuperação segura de qualquer aceite após perda de resposta HTTP, sem duplicar efeito nem exigir consulta a diretório global por UUID.
Trade-offs: Exige que o hash de idempotência exclua campos voláteis (timestamps de trace, URLs temporárias); qualquer inclusão indevida quebra a deduplicação.
Consequences: UUID nunca é usado como prova de autorização; GET sempre valida tenant antes de servir qualquer conteúdo.

### Decision 4: Credencial explícita, sem fallback implícito
Decision: Cada vínculo declara `SHARED_HUB` ou `TENANT_DEDICATED`, conta externa, `settlement_party` e política de rotação (CFG-05, SEG-05, ADR-21, FIN-11).
Rationale: Evita que uma credencial dedicada ausente/revogada seja substituída silenciosamente por uma compartilhada ou de outro tenant.
Trade-offs: Uma credencial dedicada indisponível bloqueia apenas aquele vínculo (não outros saudáveis), exigindo desenho de rotas independentes por vínculo.
Consequences: Rotação, cofre indisponível e revogação emergencial têm procedimentos próprios (SEG-05) que não trocam de conta/pagador automaticamente.

### Decision 5: Redis sempre opcional (L1 local + L2 opcional)
Decision: Cache L1 em memória para configuração imutável/resultado versionado; Redis L2 opcional com bypass automático (DAD-10, ADR-22).
Rationale: Nenhuma garantia crítica (idempotência, saldo, outbox, quota) pode depender de uma cópia única em cache.
Trade-offs: A origem (PostgreSQL) precisa suportar o pico contratado com cache frio e falha de zona — não é permitido subdimensionar o núcleo "porque há cache".
Consequences: Todo ensaio de aceite inclui o cenário com Redis totalmente desligado (QUA-06).

## Alternatives Considered

### Alternative 1: gRPC/Protobuf e request/reply sobre broker no caminho SYNC
Description: Usar gRPC para toda comunicação interna e um padrão de request/reply sobre SNS/SQS para o caminho síncrono.
Why not chosen: Nenhum benefício medido sobre REST/HTTPS direto (ADR-02); reintroduziria a espera obrigatória em fila que a v4 explicitamente remove do caminho SYNC (COM-01).

### Alternative 2: Cluster de dados único (não segmentado por célula)
Description: Um único cluster PostgreSQL para todos os tenants, com isolamento apenas por schema/linha.
Why not chosen: Permite interferência entre tenants mesmo com pods separados; contraria ADR-03 e o requisito de isolamento de capacidade física (OPE-08).

### Alternative 3: Workflow engine arbitrário (ex. Temporal) para composição
Description: Delegar orquestração de passos/composição a um motor de workflow genérico.
Why not chosen: O grafo tipado e limitado (até 20 passos, 5 simultâneos) já satisfaz o escopo da v4 sem exigir uma nova dependência (ARQ-04, CAT-04).

### Alternative 4: Jornal distribuído alternativo como autoridade de admissão
Description: Substituir PostgreSQL por um "journal" próprio (fila, disco local, cache) como autoridade de aceite.
Why not chosen: Duplicaria o problema de consenso sem prova de qualificação; PostgreSQL gerenciado multi-AZ permanece a autoridade definida (DAD-09).

## Affected Components

| Component | Change | Reason |
|---|---|---|
| Portal (Kong) | Nenhuma mudança de código do hub; configuração de rotas/auth/quotas | Gateway é produto pronto (ARQ-04); regras de negócio não migram para plugins |
| Atlas | Novo: publicação de perfis de célula, credenciais, placement, capacidade | CFG-06, DAD-11 |
| Órbita | Novo: despacho DIRECT/QUEUED com posse única, temporizador de deadline serializável | EXE-14/15, COM-06 |
| Cometa | Novo: controlador adaptativo (AIMD), scheduler de polling, resolução de credencial | OPE-07, EXE-05, SEG-05 |
| Pulsar | Novo: entrega de webhook com mesma representação de GET | COM-05, EXE-08 |
| Libra | Novo: reserva estrita, ledger, conciliação por três modelos de settlement | FIN-06/07/08/11 |

## Main Flows

### Flow 1: Admissão e execução SYNC direta (EXE-14)
1. Portal autentica e roteia o pedido; Órbita valida contrato/entrada/capacidade.
2. Órbita confirma protocolo/intenção `DIRECT` no core (UUIDv7 + Idempotency-Key), reserva financeiro se estrito (FIN-06).
3. Órbita envia comando `DIRECT` idempotente ao Cometa (COM-06), com deadline propagado.
4. Cometa resolve credencial (SEG-05), chama o provedor, persiste observação e retorna o fato à Órbita pela mesma chamada HTTPS.
5. Órbita captura/consolida o que falta, valida a projeção final, decide o prazo (EXE-11) e confirma o final; responde na conexão original.
6. Publicação de fatos (Libra, Pulsar) e webhook seguem por outbox, fora do caminho obrigatório de resposta (DAD-09).

### Flow 2: Admissão ASYNC com callback e polling concorrentes (EXE-05/06)
1. Órbita aceita e responde 202 após commit da intenção `QUEUED`.
2. Cometa consome o comando, executa a chamada, ativa polling conforme o modo configurado (`CALLBACK_ONLY`, `POLLING_ONLY`, `CALLBACK_WITH_POLLING_FALLBACK`).
3. Callback e/ou polling reportam observações ao mesmo consolidado da operação externa; uma única transição é aplicada, evidências duplicadas são preservadas sem reabrir o caso.
4. Órbita consolida o final, publica fato via SNS; Pulsar entrega o webhook com a mesma representação que o GET já serviria.

### Flow 3: Agregação/composição de produto (CAT-03/04/05)
1. Órbita agenda em paralelo os passos `READY` sem dependências pendentes, respeitando limites de protocolo/tenant/célula (EXE-13).
2. Cada passo é uma operação independente em Cometa, com seu próprio orçamento dentro do deadline absoluto do protocolo pai.
3. Consolidação aplica a política publicada (todos obrigatórios / parcialidade aceita / quórum), sem rollback universal de efeitos irreversíveis (CAT-05).

## Error Flows

### Error Flow 1: Timeout de rede entre Órbita e Cometa em SYNC (COM-06, EXE-09)
1. Órbita não recebe resposta do comando `DIRECT` dentro do orçamento.
2. Estado permanece `UNKNOWN`/incerto; nenhuma nova operação é enviada ao provedor sem consultar o estado existente em Cometa.
3. Se o deadline do cliente já expirou, Órbita finaliza `EXPIRED`/`SLA_EXCEEDED` de forma durável; se ainda há orçamento, retry seguro segue a janela de EXE-10.

### Error Flow 2: Retorno tardio após expiração por SLA (EXE-12)
1. Cometa/Órbita recebem um final (callback ou observação de polling) depois do encerramento do protocolo por SLA.
2. O recibo é persistido como `REJECTED_LATE_SLA`, com evidência e correlação; o erro final do protocolo não é substituído, nenhum sucesso é reenviado.
3. Financeiro pode registrar custo comprovado ou abrir contestação, sem gerar receita do medidor de sucesso (FIN-10).

### Error Flow 3: Autoridade financeira (Libra) indisponível (FIN-06, OPE-13)
1. Órbita tenta obter reserva idempotente para contrato de saldo estrito.
2. Sem confirmação de Libra dentro do prazo de admissão financeira, o passo externo não é liberado (negação exige falha conhecida, não apenas timeout).
3. Contratos pós-pagos sem teto estrito podem continuar processando dentro do limite de risco aprovado.

## API / Contract Design

Especificação formal (OpenAPI para REST, AsyncAPI para eventos) é responsabilidade de uma tarefa futura de `tasks.md`; este design fixa apenas as regras de contrato já normativas:

- GET de protocolo é o endpoint unificado de estado/resultado; mesma representação (schema, media type, hash) usada no corpo do webhook (COM-05).
- Toda resposta de criação inclui `protocol_id` (UUIDv7) e endereço de consulta, em headers documentados e no corpo canônico (EXE-16).
- Erros de negócio (falha, EXPIRED, parcialidade) usam o mesmo contrato do cliente, nunca um wrapper adicional só no webhook.
- Mudanças de contrato seguem versionamento aditivo (COM-04): remoção de campo, mudança de unidade/identidade/cobrança exige nova versão.

## Data Model and Persistence

Modelo lógico mínimo (DAD-02), sem DDL física: Cliente/aplicação, Serviço/produto versão, Conta/vínculo de provedor, Vínculo de credencial, Placement/capacidade, Contrato/plano versão (Atlas); Protocolo, Passo, Resultado (Órbita); Operação externa, Tentativa, Recibo externo, Agenda de polling (Cometa); Entrega (Pulsar); Uso/fato econômico, Lançamento, Reserva/consumo de franquia, Fatura/conciliação (Libra); Objeto (S3, referenciado por domínio); Outbox/inbox/timer e Auditoria (cada domínio).

Índices previstos (DAD-03): tenant/protocolo; tenant/chave idempotente; tenant/estado/data; operação/conta/provider_request_id; agenda/next_run_at/estado; outbox pendente/data; entrega/estado/próxima tentativa; contrato/período/unidade; chave econômica.

Retenção por classe de dado (DAD-06): protocolos/resultados propostos em 90 dias online; idempotência/inbox/correlação em pelo menos 35 dias; replay de eventos em 30 dias; logs/traces/métricas em 30/7/90 dias propostos — todos sujeitos a aprovação (P-04), não orientação jurídica definitiva.

## Authentication and Authorization

- Clientes de máquina: OAuth2 client credentials + JWT homologado (SEG-01).
- Usuários administrativos: OIDC + MFA, sessões e privilégios com prazo.
- Comunicação interna: identidade de workload + mTLS.
- Perfil administrativo `hub_protocol_reader`: atribuído a identidades humanas individuais, nunca a conta compartilhada, com leitura entre tenants auditada e mascarada por padrão (SEG-04).
- RLS como defesa adicional no PostgreSQL; contas de aplicação sem `BYPASSRLS` (DAD-07).

## Security and Privacy

- Prevenção de SSRF em destinos configuráveis (webhooks de cliente, callbacks de provedor): bloqueio de metadados/redes internas, revalidação de DNS, controle de egress (SEG-02).
- Segredos apenas em cofre (Secrets Manager + KMS), nunca em eventos, configuração em claro, logs ou arquivos versionados.
- Dado pessoal nunca vira label de métrica Prometheus (OPE-06, SEG-02).
- Resolução de credencial determinística, sem fallback implícito entre `SHARED_HUB` e `TENANT_DEDICATED` (SEG-05).
- Isolamento estrito de tenant: nenhum acesso cross-tenant por manipulação de UUID/header/body; resposta idêntica para "inexistente" e "de outro tenant" (SEG-04).

## Observability

### Logs
Logs estruturados via Alloy → Loki; payload sensível não indexado indiscriminadamente (SEG-02, OPE-06).

### Metrics
Prometheus com baixa cardinalidade: admissões, recusas por causa, latências por etapa, idade de outbox/backlog, operações UNKNOWN, entregas esgotadas (OPE-06). `protocol_id`/`event_id`/dados pessoais nunca em labels.

### Tracing
OpenTelemetry propaga contexto; Tempo armazena traces com amostragem controlada — não é fonte contábil nem única evidência de SLA (ARQ-05).

### Alerts
Conforme tabela de OPE-06: outbox antiga > 30s, DLQ com item ou callback órfão > 5 min, polling atrasado > 30s, resultado sem entrega > 30s, apuração atrasada > 5 min, operação UNKNOWN além do prazo homologado.

## Testing Strategy
- Unit tests: por domínio (Órbita, Cometa, Pulsar, Libra), cobrindo invariantes transacionais.
- Integration tests: PostgreSQL/broker/objetos, por célula.
- Contract tests: por provedor/cliente (COM-02/04).
- E2E tests: fluxos SYNC e ASYNC completos, incluindo agregação/composição.
- Performance tests: perfis de carga de OPE-01 (1.000 admissões/s sustentadas, rajada 2.000/s, 2.000 GETs/s, 8h de longa duração a 50%).
- Security tests: SSRF, RLS/troca de tenant, credencial cruzada, XML malicioso em SOAP (SEG-02/03, QUA-02).
- Manual validation: nenhuma prevista além da qualificação formal descrita em `qualidade-e-aceite` (gates G0–G4).

Todos os cenários desta mudança e da matriz-fonte permanecem **NÃO EXECUTADO** — nenhum teste foi rodado nesta revisão documental (QUA-01/06).

## Migration Strategy
Não aplicável nesta mudança (não há sistema implementado anteriormente). Migrações futuras de dados (ex. realocação de tenant entre células) seguem o workflow de DAD-11: preparar destino, drenar/reconciliar, barreira de admissão, cópia incremental, fencing da origem, ativação do destino, conferência antes de retirar a origem.

## Rollback Plan
Não aplicável a artefatos de especificação. Publicação de configuração (Atlas) segue o ciclo de CFG-02: rascunho → validação → simulação → aprovação → publicação versionada → observação → rollback por nova publicação (nunca por exclusão de histórico).

## Compatibility
Compatibilidade com a v3 preservada conforme DEC-03/DEC-04: contratos de entrada/saída por cliente, mesma representação em GET/webhook, deadline rígido, TTL sem renovação, rejeição de retorno tardio, isolamento de tenant e administrador individual auditado permanecem sem alteração de comportamento. A única mudança de transporte é a remoção da fila obrigatória no caminho SYNC (ADR-19) e a adoção universal de UUIDv7 (ADR-20).

## Operational Considerations
Ambientes local/dev/hom/ppd/prd conforme OPE-04, com o mesmo digest de imagem promovido entre ambientes. Gates G0–G4 (QUA-04) definem os critérios de promoção. Escala automática por HPA (APIs)/KEDA (workers)/Karpenter (nós)/Crossplane (células) dentro de envelopes homologados (OPE-12), sem chamado manual para crescimento rotineiro.

## Remaining Risks

| Risk | Impact | Mitigation | Owner/Decision |
|---|---|---|---|
| Duplicação de efeito entre despacho direto e fila (GAP-02) | Cobrança/execução duplicada no provedor | Intenção, posse e identidade comuns entre DIRECT e QUEUED (EXE-15) | Engenharia — qualificação pendente |
| "Banco opcional" confirmar aceite sem prova de custódia (GAP-05) | Aceite fictício sem durabilidade real | Fronteiras duráveis explícitas (DAD-09) e continuidade restrita (OPE-13) | Arquitetura — limite assumido, não flexibilizável |
| Perda regional / split-brain (GAP-12) | Perda de dados aceitos ou duas autoridades ativas simultâneas | Topologia regional adicional só com autoridade e cópias adequadas (OPE-11/15) | Arquitetura — bloqueado por P-08 |
| Migração de célula perder idempotência (GAP-09) | Duas autoridades de escrita ativas para o mesmo tenant | Workflow de placement/epoch com corte exclusivo (DAD-11) | Dados — qualificação pendente |
| Confundir especificação completa com disponibilidade comprovada (GAP-15) | Declarar prontidão de produção sem evidência | Gates G0–G4, matriz NÃO EXECUTADO (QUA-03/04/06) | Liderança técnica — gate por gate |

## Open Questions
Ver seção "Open Questions" do `proposal.md` (P-01 a P-11) — nenhuma é resolvida por este design.
