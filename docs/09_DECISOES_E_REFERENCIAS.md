# 09 · Decisões, pendências e referências

## DEC-01 · Decisões arquiteturais de referência — proprietário: Arquitetura

Estas decisões orientam a especificação v4; não representam aprovação comercial ou infraestrutura existente.

| Decisão | Escolha e consequência | Alternativa e motivo de não adotar agora |
| --- | --- | --- |
| ADR-01 — limites de domínio | Cinco aplicações, roteamento/composição na Órbita; polling no Cometa | Serviço separado por função aumenta custo distribuído sem necessidade demonstrada |
| ADR-02 — comunicação | REST/HTTPS nos contratos interativos e execução SYNC direta; SNS/SQS para ASYNC e fatos | gRPC em tudo e request/reply por fila sem benefício medido |
| ADR-03 — persistência | PostgreSQL de controle e core/finance por célula; isolamento físico conforme classe | Um cluster global de dados ainda permite interferência entre tenants apesar de pods separados |
| ADR-04 — resultado | Hub conserva versão final e atende GET local | Proxy de resultado no provedor viola a custódia e a independência solicitadas |
| ADR-05 — execução | Máquina de estados durável, grafo limitado, deadline rígido e TTL sem renovação | Workflow arbitrário/BPMN não é requisito; retorno tardio não pode reabrir protocolo expirado |
| ADR-06 — integridade | Outbox/inbox, chaves semânticas, reconciliação e UNKNOWN | Promessa end-to-end de exactly-once sem suporte externo é infundada |
| ADR-07 — financeiro | Fatos, receita e custo independentes, ledger gerencial, integração ERP | Cobrar todo evento técnico ou inferir custo pelo provedor vencedor perde correção |
| ADR-08 — plataforma | AWS/EKS como referência; Docker e kind no laboratório | Pilha integral autogerenciada exige decisão e operação distintas |
| ADR-09 — configuração | Versões publicadas, capacidades homologadas, sem scripts arbitrários | Endpoint cadastrado sozinho não implementa semântica ou segurança |
| ADR-10 — observabilidade | Prometheus, Loki/Alloy, Grafana e OpenTelemetry/Tempo | Logs sem trace e métricas não dão visão completa; auditoria não pode depender só de logs |

Conta, região, condições de dados, orçamento, licenciamento e modo gerenciado/autogerenciado são pendências de implantação. A especificação define referência e limites para que Engenharia não escolha silenciosamente uma tecnologia por ambiente.

## DEC-02 · Registro de decisões pendentes — proprietário: liderança técnica

| ID | Informação necessária e responsável final | Condição de bloqueio / critério de resolução |
| --- | --- | --- |
| P-01 | Plataforma: conta, região, Kubernetes gerenciado ou infraestrutura integralmente própria, orçamento, registro OCI e licenças | Antes de IaC remota e ppd; aprovação registrada da topologia e custo |
| P-02 | Produto: catálogo, volumes, fan-out, perfis legados, mix real e classes de isolamento | Antes de assumir SLO/capacidade; perfis de carga e limites aprovados |
| P-03 | Comercial: contratos de compra/venda, tarifas, planos, marcos, faixas, parcialidade e riscos de failover | Antes de habilitar oferta comercial; matriz econômica sem lacunas |
| P-04 | Responsável pelos dados: retenção, finalidade, região permitida e direitos de custódia dos resultados | Antes de dados reais/publicação; política por classe e contrato aprovada com Segurança |
| P-05 | Integrações: idempotência, SYNC/polling/callback, credenciais compartilhadas/dedicadas, recuperação de resposta, domínios de capacidade e equivalência | Antes de habilitar vínculo/failover; evidências de homologação |
| P-06 | Financeiro: ERP, layout/API, plano de contas gerencial, arredondamento, competência, ajustes e liquidação | Antes de fechamento/exportação; exemplos reconciliados e contrato de integração aprovado |
| P-07 | Produto: SLA bilateral, TTL em segundos, enforcement do provedor, reserva final e compromisso de entrega | Antes de vender SLA; contrato sem garantias incompatíveis com provedores/DR |
| P-08 | SRE: domínio de falhas, RPO zero contratado, eventual confirmação entre regiões, RTO e capacidade por célula | Antes de prometer ausência de perda regional/prd; desenho e prova de recuperação, impacto de latência e custo |
| P-09 | Liderança técnica: responsáveis, perfis administrativos dos desenvolvedores e aprovação da v4 | Antes do gate G0 da fatia; registro de decisão e responsáveis |
| P-10 | Produto: perfil de criticidade com responsável do sistema consumidor, máxima interrupção em segundos e contingência | Antes de ofertar uso com impacto em vida/segurança; riscos, SLO/RPO/RTO e provas aprovados |
| P-11 | Plataforma: envelopes de escala, quotas preautorizadas, reserva, horizonte de provisionamento e política de células | Antes do gate ppd de expansão automática; cenário sem chamados demonstrado com limites/custos |

Os números de metas e retenção marcados como propostos permitem projetar e ensaiar; não substituem estas decisões. Algumas pendências bloqueiam apenas a produção ou a oferta específica, não todo trabalho de engenharia sintética.

## DEC-03 · Decisões herdadas da v3 e preservadas — proprietário: Arquitetura / Produto

| Decisão | Motivo e consequência verificável |
| --- | --- |
| ADR-11 — TTL e SLA | Janela em segundos limita retries; deadline absoluto limita experiência do cliente mesmo sem erro técnico. Não se renova com reinício/failover |
| ADR-12 — retorno tardio | EXPIRED por SLA não vira sucesso; recibo permanece como evidência e eventual custo. HTTP 2xx de recebimento não é aceitação de negócio |
| ADR-13 — contrato do cliente | Entrada/saída podem variar por cliente; projeção final única e persistida garante equivalência entre GET e webhook |
| ADR-14 — pressão adaptativa | Cometa ajusta taxa, concorrência e pendência externa com feedback, teto contratual e coordenação global por domínio |
| ADR-15 — células | Cargas ruidosas exigem isolamento de filas, pools, dados e capacidade, além de pods; classe dedicada para requisitos mais fortes |
| ADR-16 — administração | Perfil individual de desenvolvedor permite leitura entre tenants sem conceder esse poder a clientes ou usar conta compartilhada |
| ADR-17 — linguagem por componente | Go nas APIs/workers, TypeScript/React na interface; ferramentas de infraestrutura não serão reescritas. Trade-offs detalhados em ARQ-04/05 |
| ADR-18 — preservação | Objetivo de ausência de perda é ligado ao modelo de falhas testado; compromisso regional de RPO zero bloqueado até desenho adequado |

## Compatibilidade com as revisões anteriores

A v4 mantém as correções introduzidas na v3 sobre a v2: consulta de resultado pendente passa a usar representação unificada (200); retorno tardio depois de encerramento por SLA não pode publicar revisão de sucesso; bancos e filas são segmentados por célula; a antiga meta regional de RPO de 15 min não é mais padrão implícito. São mudanças de especificação, não alegação de migração de sistema implementado.

Os capítulos existentes mantêm seus IDs para rastreabilidade; novos requisitos têm IDs próprios. Todos os arquivos e casos de aceite foram revistos quanto ao impacto, incluindo financeiro e acesso administrativo. O relatório registra a cobertura e as limitações. Não há adição de código de aplicação ou deploy nesta revisão.

## Fontes primárias e alcance

As fontes fundamentam propriedades técnicas; limites, nomes, regras comerciais e SLOs são decisões propostas deste projeto, não afirmações atribuídas a fornecedores. Referências verificadas na elaboração das revisões estão assinaladas como consultadas; as fontes novas da v4 foram conferidas em 6 de setembro de 2026; referências complementares devem ser revalidadas na implementação antes de fixar versões.

| Fonte | Uso |
| --- | --- |
| [AWS — Transactional Outbox](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html) | Consultada: integridade da escrita entre estado e intenção de mensagem; duplicatas ainda exigem tratamento |
| [AWS — SQS at-least-once delivery](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/standard-queues-at-least-once-delivery.html) | Consultada: filas Standard podem repetir entrega |
| [PostgreSQL — Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) | Consultada: políticas por linha e exceções de papéis privilegiados |
| [Kubernetes — Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/) | Consultada: diferenças entre startup, readiness e liveness |
| [Kubernetes — HPA](https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/) | Consultada: fontes de métricas, requests e comportamento de escala |
| [OWASP — SSRF Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html) | Consultada: controles para destinos externos configuráveis |
| [RFC 9110 — HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) | Complementar: semântica HTTP e aceite 202 |
| [RFC 9562 — UUID](https://www.rfc-editor.org/rfc/rfc9562.html#section-5.7) | Consultada v4: UUIDv7 em ambas as modalidades; identidade não é autorização |
| [W3C — Trace Context](https://www.w3.org/TR/trace-context/) | Complementar: propagação de contexto técnico |
| [Grafana — Alloy Kubernetes logs](https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.kubernetes/) | Consultada: coleta de logs Kubernetes |
| [KEDA — Scaling Deployments](https://keda.sh/docs/latest/concepts/scaling-deployments/) | Consultada: escala por eventos e integração com HPA |
| [Grafana — Tempo](https://grafana.com/docs/tempo/latest/introduction/) | Consultada: backend de tracing distribuído |

Referências específicas da v3:

- [Go — FAQ e goroutines](https://go.dev/doc/faq#goroutines): concorrência do runtime e limites; não prova benchmark do hub.
- [React — TypeScript](https://react.dev/learn/typescript): suporte à linguagem para a interface administrativa.
- [Kong Gateway](https://developer.konghq.com/gateway/): responsabilidades do gateway como produto.
- [Envoy — Adaptive Concurrency](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/adaptive_concurrency_filter): exemplo primário de feedback de latência; não implica adotar o filtro nem igualar seus algoritmos ao AIMD proposto.
- [AWS Builders' Library — Timeouts, retries e jitter](https://d1.awsstatic.com/builderslibrary/pdfs/timeouts-retries-and-backoff-with-jitter.pdf): amplificação de carga por retries e recuperação controlada.
- [AWS Builders' Library — Workload isolation](https://d1.awsstatic.com/builderslibrary/pdfs/workload-isolation-using-shuffle-sharding.pdf): isolamento e contenção de impactos; a classe/célula definida aqui é decisão própria do hub.

Não há afirmação de versões atuais de imagens, limites máximos de S3, preço de nuvem ou conformidade regulatória garantida. Essas propriedades variam e deverão ser verificadas no projeto de implementação.

## DEC-04 · Decisões e compatibilidade introduzidas na v4 — proprietário: Arquitetura

| Decisão | Motivo, efeito e compatibilidade |
| --- | --- |
| ADR-19 — execução SYNC direta | Retira fila obrigatória no comando/final do caminho síncrono da v3. Conserva estado, intenção, deadline e fatos. ASYNC/AUTO mantêm fila; não há fallback silencioso de SYNC para 202 |
| ADR-20 — UUIDv7 universal | Protocolo persistido e consulta futura em todas as modalidades; HTTP perdido é recuperado por idempotência; recusa anterior ao aceite não vira protocolo fictício |
| ADR-21 — credencial explícita | SHARED_HUB e TENANT_DEDICATED com vínculo/conta/tenant, rotação e proibição de fallback implícito. Pagador é regra contratual separada |
| ADR-22 — dependências mínimas | L1/projeções e Redis opcional; quatro fronteiras lógicas duráveis no SYNC simples, sem broker/Atlas/cofre por pedido quando material válido já existe. Não adotar escrita alternativa sem autoridade |
| ADR-23 — expansão completa | HPA/KEDA para pods, Karpenter para nós, Atlas/Plataforma + Crossplane para células e recursos. Recursos crescem automaticamente dentro de perfis/quotas, sem alterar fronteiras de negócio |
| ADR-24 — continuidade seletiva | Rotas/deployments de leitura e admissão com readiness e pools próprios; fallbacks por operação e validade. Banco, saldo estrito e autenticação não se tornam opcionais por estar em incidente |
| ADR-25 — criticidade e manutenção | Manutenção conta no SLI; RTO e perda de região exigem prova específica. Percentual mensal e réplica multi-AZ não bastam para usos com impacto em vida/segurança |

A v4 altera a escolha de transporte e retira o default de oito segundos para todos os síncronos. Preserva integralmente os requisitos de TTL em segundos, deadline não renovável, recusa tardia, aferição bilateral, custódia de resultado, agregação/composição, pressão adaptativa, contratos legados, equivalência GET/webhook, compra/venda e acesso administrativo individual. Os IDs herdados permanecem; novos IDs ampliam a matriz.

Novas fontes primárias verificadas: [Karpenter — conceitos](https://karpenter.sh/docs/concepts/), [Crossplane — plano de controle](https://docs.crossplane.io/latest/whats-crossplane/), [AWS — cache Go de segredos](https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_cache-go.html), [AWS — failover Multi-AZ de instância RDS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.Failover.html). Estas fontes sustentam capacidades e limites dos produtos; os perfis de capacidade, algoritmos, regras e metas do Hub são decisões desta especificação. Links não fixam versões de implantação.

## DEC-05 · Registro de gaps, riscos e mitigação — proprietário: Arquitetura / liderança técnica

“Resolvido no desenho” significa regra definida e contradição documental removida; não significa mitigação implementada ou testada. Todos os itens exigem evidência antes da liberação do escopo indicado. Limite arquitetural não pode ser encerrado apenas por revisão de texto.

| Gap / severidade | Risco identificado | Mitigação e requisitos | Responsável final / condição de encerramento |
| --- | --- | --- | --- |
| GAP-01 / Alta | Fila obrigatória compromete contrato SYNC | Transporte direto com commits; COM-01/06, EXE-14 | Engenharia: final na mesma conexão com broker fora; resolvido no desenho, qualificação pendente |
| GAP-02 / Crítica | Caminhos direto e fila duplicam efeitos | Intenção, posse, identidade e recuperação comuns; EXE-15 | Engenharia: crash/duplicata/UNKNOWN sem reexecução indevida; qualificação pendente |
| GAP-03 / Crítica | Credencial de outro cliente ou conta compartilhada usada indevidamente | Resolução explícita e sem fallback implícito; CFG-05, SEG-05 | Segurança: testes cruzados, rotação e logs sem segredo; resolvido no desenho |
| GAP-04 / Alta | Redis vira ponto único ou causa avalanche na origem | Cache dispensável, timeout/bypass e origem dimensionada; DAD-10 | SRE: pico N−1 com Redis desligado e frio; qualificação pendente |
| GAP-05 / Crítica | “DB opcional” confirma pedido/final sem prova de custódia | Fronteiras duráveis e continuidade restrita; DAD-09, OPE-13 | Arquitetura: limite assumido; nenhuma aprovação pode chamar escrita sem autoridade de durável |
| GAP-06 / Alta | HPA aumenta pods, mas nó/banco/célula exigem chamado | Automação de camadas e ativação com capacidade; CFG-06, OPE-12 | Plataforma: expansão completa sem intervenção no envelope aprovado; P-11 |
| GAP-07 / Alta | Células/credenciais novas multiplicam quota externa | Orçamento global, concessões e pressão hierárquica; OPE-07/12 | Integrações: múltiplas células, dono perdido e slots incertos; qualificação pendente |
| GAP-08 / Alta | Cofre/Atlas/identidade fora causam indisponibilidade ou autorização velha | Pré-aquecimento e validade de material/projeções; CFG-02, SEG-05, OPE-13 | Segurança: máxima defasagem aprovada e ensaios quente/frio; risco residual explícito |
| GAP-09 / Crítica | Migração perde idempotência ou cria duas autoridades | Placement/epoch, cópia e drenagem com corte exclusivo; DAD-11 | Dados: interrupção em cada fase sem split-brain; limite de corte dentro do SLA |
| GAP-10 / Crítica | Final no limite do SLA é aceito por relógio ou commit incorreto | Arbitragem terminal com prova temporal; EXE-11/14 | Engenharia: D−ε/D/D+ε e commit lento; garantia ainda não demonstrada |
| GAP-11 / Crítica | Failover de banco excede tolerância do consumidor | Perfil de criticidade e qualificação física; OPE-14/15 | Produto: P-10 e SRE comprovam máximo de interrupção; oferta crítica bloqueada se incompatível |
| GAP-12 / Crítica | Perda regional destrói dados aceitos ou cria split-brain | Topologia regional adicional somente com autoridade e cópias adequadas; OPE-11/15 | Arquitetura: P-08; referência atual não qualifica RPO zero regional |
| GAP-13 / Crítica | Provedor executa, resposta some e não oferece reentrega/status/idempotência | UNKNOWN, contrato de recuperação e restrição de oferta; EXE-09, OPE-13/15 | Integrações: P-05; não vender garantia externa que o parceiro não sustenta |
| GAP-14 / Alta | Conta dedicada altera cobrança/pagador por inferência | Vínculo econômico explícito e fatos completos; FIN-11 | Financeiro: três modelos reconciliados, incluindo polls e retorno tardio |
| GAP-15 / Crítica | Texto completo é confundido com disponibilidade comprovada | Gates, cenários, matriz e estado NÃO EXECUTADO; QUA-03/04/06 | Liderança técnica: evidências da fatia/perfil, riscos residuais aceitos e responsáveis nominais |

Avaliação de compatibilidade: crescimento por quantidade de clientes/provedores é compatível com células e automação; isolamento continua limitado pelo domínio de recursos contratado. SYNC direto é compatível com durabilidade, mas acrescenta commits necessários ao prazo da conexão. Cache opcional é compatível com custódia, pois não guarda a única cópia; ausência simultânea de todas as autoridades de escrita é incompatível com novo aceite durável irrestrito. RPO zero regional e latência mínima em partição não podem ser prometidos ao mesmo tempo sem explicitar quorum, autoridade e restrições de disponibilidade.

O plano de mitigação prioriza GAP-02/03/05/09/10/11/12/13/15 como bloqueadores do escopo afetado. Resolver P-01 a P-11 com responsáveis é parte do gate correspondente; não interrompe a elaboração da engenharia documental nem autoriza hipóteses comerciais ocultas. Nenhum código, manifesto, provisionamento ou teste de aplicação foi produzido nesta revisão.
