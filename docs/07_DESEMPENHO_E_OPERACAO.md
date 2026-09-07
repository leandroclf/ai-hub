# 07 · Desempenho, alta disponibilidade e operação

## OPE-01 · Modelo de capacidade — proprietário: Arquitetura / SRE

Dimensionar por requisições por segundo, duração, fan-out, quantidade de polls, tamanho de conteúdo, concorrência por provedor e prazo de recuperação. “Milhares de conexões” isoladamente não especifica capacidade. Deve existir perfil medido por classe de produto/serviço e por tenant, incluindo máximos, não só médias.

Relações de planejamento: operações externas por segundo ≈ pedidos/s × média de passos executados; consultas de status/s ≈ operações pendentes com polling ÷ intervalo médio; concorrência de chamadas ≈ taxa de chamadas × latência média, em regime estável. Rajadas, caudas e variância exigem margem adicional. Em composição, o caminho crítico limita a duração, mesmo com alto paralelismo.

Exemplo sintético: 1.000 pedidos/s e três passos geram até 3.000 operações/s; se 1.000 operações/s ficam pendentes por 60 s, são aproximadamente 60.000 operações pendentes. Consultar todas a cada 5 s adiciona cerca de 12.000 polls/s. Configurar polling sem calcular esse custo e a quota do provedor pode inviabilizar a oferta.

Perfis iniciais propostos para qualificação: 1.000 admissões/s sustentadas por 60 minutos; rajada de 2.000/s por 5 minutos; 2.000 GETs de resultado/s; teste de longa duração de 8 horas em 50% da carga nominal. Testar mix de 70% serviço simples, 20% agregação de três passos e 10% composição de cinco passos, com perfil de latência e proporção assíncrona declarados. Não usar esses números como SLA ou escala real antes de P-02.

Arquivos grandes têm ensaio separado: carga mista, fluxo contínuo e limite máximo por adaptador. Throughput de bytes e uso de memória precisam ser medidos junto da taxa de pedidos. Entrada sintética deve incluir limites, payloads inválidos e respostas grandes; não só respostas mínimas em memória.

## OPE-02 · Estratégia de baixa latência — proprietário: Engenharia

Persistência é parte do caminho de aceite e não pode ser removida para melhorar benchmark. Reduzir overhead por pools limitados e reutilizados, conexões HTTP persistentes, projeções locais de configuração, índices dos caminhos críticos, transações curtas, publicação/consumo em lotes controlados e processamento paralelo limitado. Não executar todos os passos de um produto em série sem dependência funcional.

Consultas finais usam o banco autoritativo ou objeto do hub, sem provedor. Relatórios pesados usam projeção/replica analítica e não disputam com admissão/consulta. Cache opcional serve configuração imutável ou resultado por versão; nunca é ledger, outbox ou única fonte de correlação. Arquivos usam upload direto e streaming; transformações com materialização têm limite próprio.

Concorrência máxima deve existir por pod, serviço, provedor/conta e tenant, incluindo orçamento de conexões ao PostgreSQL. Soma de pools no máximo de réplicas deve caber no orçamento de banco; autoscaling não autoriza exaurir conexões. Quotas externas são globais por conta e não multiplicadas pelo número de pods. Clientes de alto volume não podem esgotar todos os slots; aplicar distribuição justa e orçamento separado de submissão/polling/entrega.

## OPE-03 · SLOs e semântica de garantia — proprietário: Produto / SRE

| Indicador | Meta inicial proposta | Como medir e delimitar |
| --- | --- | --- |
| Disponibilidade de admissão e GET local | Proposta elevada para 99,99% mensal por API, célula e região; perfil crítico exige aprovação específica | Requisições válidas elegíveis servidas corretamente; 5xx/timeouts e 429 por falta de capacidade do Hub contam como falha; manutenção também conta |
| Latência de aceite ASYNC | p95 ≤ 250 ms; p99 ≤ 500 ms | Chegada à borda até resposta após commit; sem transferência de arquivo grande, dentro da região e perfil qualificado |
| Consulta de resultado local | p95 ≤ 100 ms; p99 ≤ 250 ms | Resposta com metadados/JSON até 64 KiB; download de objeto tem medição própria |
| Final ASYNC | p95 ≤ 2 s; p99 ≤ 5 s | Final durável observado no Cometa até resultado durável na Órbita; não é orçamento para o SYNC direto |
| Final SYNC | Deadline próprio da oferta; overhead medido separadamente do provedor | Mesma conexão, commits e projeção incluídos; filas de comando/final ausentes do percurso obrigatório |
| Primeira tentativa de webhook | p95 ≤ 5 s | Resultado final persistido até início da tentativa; não mede confirmação do destino |
| Atualidade financeira pós-paga | p95 ≤ 60 s | Fato econômico confirmado até apuração visível; fechamento exige completude, não percentil |
| Polling | 99% das consultas iniciadas até 5 s após next_run_at | Medir atrasos por quota externa separadamente, sem ocultá-los do prazo ao cliente |
| Recuperação de falha de zona | RTO de referência geral proposto ≤ 5 min; não adotado automaticamente no perfil crítico | Exercício de falha, failover e retomada por componente; RPO deve ser demonstrado pela configuração |
| Desastre regional | RTO de referência geral proposto ≤ 4 h; RPO pendente de P-08; insuficiente para vários usos críticos | A referência AWS regional não demonstra RPO zero regional; bloquear essa promessa até qualificação específica |
| Obrigações sem rastreio | Zero no conjunto reconciliado | Todo protocolo aceito tem estado, intenção/resultado, pendências e relações verificáveis |

Erros de autenticação/validação e bloqueio contratual genuíno não entram como indisponibilidade; devem ter métricas próprias. Timeout de execução causado por provedor entra no SLI de resultado do produto e na visão do cliente, mesmo que não seja falha da API de admissão. Não excluir falhas de provedor do relatório ponta a ponta para inflar disponibilidade percebida.

A promessa é custódia durável, retomada, repetição segura e conclusão contratual ou pendência explícita dentro dos limites de retenção/recuperação. Não é sucesso de todo provedor nem entrega confirmada a todo cliente em qualquer condição. A v3 retira a meta regional de 15 min como padrão implícito diante do requisito de preservação. Contrato exigindo RPO zero regional requer persistência confirmada em domínios regionais independentes, fencing e qualificação próprios; a referência regional atual não pode anunciar esse compromisso. Pedidos aceitos jamais devem ser abandonados silenciosamente por rotina normal de expurgo.

## OPE-04 · Ambientes e alta disponibilidade — proprietário: Plataforma/SRE

| Ambiente | Execução | Dados e dependências | Finalidade e isolamento |
| --- | --- | --- | --- |
| local | Docker Compose; kind como laboratório Kubernetes | PostgreSQL, emulação AWS e observabilidade em containers; dados sintéticos | Verificação funcional local; não comprova HA multi-AZ |
| dev | Kubernetes de desenvolvimento | Dependências isoladas, contas/filas/buckets próprios | Integração contínua e contratos, sem dados reais por padrão |
| hom | Kubernetes de homologação | Serviços reais em sandbox quando disponíveis; configuração representativa | Validação funcional, contratual e financeira |
| ppd | Kubernetes com topologia de produção | Bancos separados, múltiplas zonas e observabilidade durável | Qualificação de carga, falhas, restore e promoção |
| prd | Kubernetes gerenciado multi-AZ | Core/finance separados, serviços AWS gerenciados, identidade e backups | Operação real com SLO, plantão e controle de mudança |

Configurações iniciais: duas réplicas mínimas de aplicações em dev/hom; três em ppd/prd distribuídas em três zonas, com capacidade para continuar dentro da carga crítica após perder uma. Local usa uma ou mais réplicas para ensaio. Todos os ambientes têm requests/limits, probes e configurações explícitas, mesmo quando HA não é exigida. Células e classes de isolamento de OPE-08 complementam essa distribuição; separar somente pods não isola banco, filas ou contas de provedor. Promover o mesmo digest de imagem entre ambientes; somente configuração e segredos variam.

PostgreSQL usa failover gerenciado e backups testados. Filas e objetos usam redundância do serviço com políticas verificadas. Prometheus remoto precisa réplicas independentes e armazenamento histórico durável quando a retenção exigir; Loki e Tempo usam armazenamento de objetos e topologia HA compatível; Grafana usa banco/configuração duráveis e mais de uma instância em produção. O desenho físico e a capacidade dessas dependências são entregáveis de Plataforma antes do gate ppd; um container de cada não satisfaz HA.

Kind é ferramenta de laboratório. “Tudo em containers localmente” não exige hospedar S3/SQS/RDS reais em pods na nuvem. Emulação local não prova limites, IAM, durabilidade e failover dos serviços gerenciados. Se a intenção for infraestrutura integralmente autogerenciada no Kubernetes, P-01 deve alterar explicitamente a referência antes da implementação.

## OPE-05 · Probes, escala e proteção — proprietário: Plataforma/SRE / Engenharia

Startup confirma inicialização e carga mínima de configuração; protege inicializações lentas. Readiness avalia capacidade de aceitar a responsabilidade daquela instância, com verificações autenticadas e prazo curto. Liveness identifica processo travado/incapaz de progredir internamente; indisponibilidade de provedor não deve reiniciar todos os pods. Separar health de diagnóstico detalhado; não expor credenciais ou topologia interna.

Readiness da Órbita depende de persistir novos pedidos ou responder consultas corretamente; do Cometa, de registrar trabalho durável e operar seus workers; do Pulsar, de registrar tentativas; de Libra, de persistir uso e reservas. Quando o endpoint atende capacidades com dependências diferentes, desenho de rotas/workers deve impedir que falha de envio derrube uma consulta local saudável. Redundância não dispensa tempos limite por dependência. A separação é coerente com as [probes Kubernetes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/).

HPA v4 para APIs com requests declarados e métricas de CPU mais sinais de concorrência quando qualificados; Metrics Server fornece métricas de recursos. Workers usam backlog/idade da obrigação por métrica externa, com um controlador de escala por Deployment; referência: [KEDA](https://keda.sh/docs/latest/concepts/scaling-deployments/) para filas, gerando seu HPA, sem HPA paralelo concorrente. Polling considera quantidade de timers vencidos; não apenas tamanho da fila. Escala tem mínimo/máximo, estabilização de redução e teto imposto pela capacidade de banco/provedor. [HPA Kubernetes](https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/) exige fonte de métricas apropriada; Prometheus sozinho não habilita HPA por CPU.

Distribuição por zona/nó, PodDisruptionBudget, rolling update com capacidade adicional, encerramento gracioso e drenagem de workers são obrigatórios. PDB protege interrupções voluntárias dentro de limites, não garante sobrevivência a qualquer pane. Na drenagem, parar de reclamar trabalho, terminar ou devolver posse com segurança e preservar UNKNOWN quando houver chamada externa incerta.

Backpressure impede novas admissões acima do orçamento durável e mantém consultas/recuperação prioritárias. Circuit breaker, bulkheads e limites por conta reduzem propagação de falha; abrir circuito não cancela trabalho em voo nem autoriza failover inseguro. Deadlines e filas têm políticas por classe; polling não deve consumir toda a quota necessária a novas submissões, nem o inverso.

## OPE-06 · Observabilidade e runbooks — proprietário: SRE / Operações

Métricas de baixa cardinalidade: admissões, recusas por causa, latências por etapa, erros canônicos, idade de outbox/backlog/timers, workers ativos, saturação de pools, operações UNKNOWN, callbacks órfãos/conflitantes, entregas esgotadas, resultados não disponíveis, uso não apurado e divergências de fechamento. Dimensões permitidas: ambiente, componente, serviço e provedor com cardinalidade controlada. protocol_id, event_id e dados pessoais pertencem a logs/traces protegidos, não a labels de métricas.

Prometheus coleta réplicas individualmente; Grafana apresenta operação, integração, experiência do cliente e financeiro. Alloy envia logs estruturados ao Loki. OpenTelemetry propaga trace context; [Tempo](https://grafana.com/docs/tempo/latest/introduction/) mantém traces, com amostragem controlada e preservação de erros quando possível. Auditoria durável de negócio não depende de amostragem. Dashboards financeiros respeitam RBAC e podem usar projeções próprias em vez de métricas de alta cardinalidade.

| Alerta inicial proposto | Responsável e ação |
| --- | --- |
| Outbox mais antiga > 30 s por 5 min | SRE verifica publicador/broker; preserva eventos, não apaga pendência |
| DLQ com qualquer item ou callback órfão > 5 min | Integrações diagnostica identidade/schema e corrige com replay controlado |
| Polling atrasado > 30 s ou risco de deadline | Integrações verifica quotas, backlog e plano; não reduz intervalo indiscriminadamente |
| Resultado final sem obrigação de entrega > 30 s | Operações reconcilia evento e Pulsar; não reexecuta produto |
| Apuração atrasada > 5 min ou divergência não classificada | Financeiro/Engenharia verifica fatos, dedup e contrato; bloqueia fechamento incompleto |
| Elevação de erro/latência com consumo acelerado do orçamento de erro | SRE interrompe rollout e aplica plano de mitigação |
| Operação UNKNOWN além do prazo homologado | Integrações abre reconciliação; sem troca automática insegura |

Runbooks obrigatórios: provedor indisponível; banco indisponível; broker indisponível; DLQ; callback inválido/órfão; atraso de polling; entrega esgotada; divergência financeira; perda de zona; restore regional; rotação de segredo e incidente de dados. Cada um deve ter gatilho, diagnóstico seguro, ação autorizada, limite de automação, rollback, evidência e critério de encerramento. Exemplos: recuperar broker exige drenar outbox com limite; recuperar banco exige confirmar operações incertas; recuperar webhook exige reenviar a entrega existente. Toda intervenção financeira relevante exige aprovação segregada.

## OPE-07 · Amortecimento e controle adaptativo do provedor — proprietário: Cometa / SRE

Objetivo: manter a maior taxa útil sustentável observada, respeitando o contrato e o prazo dos clientes; não perseguir throughput bruto que gera erro, filas externas e cobranças sem resultado. A referência é um controlador de feedback com aumento gradual e redução multiplicativa (AIMD), latência suavizada, orçamento de erro e sondagens de recuperação. Não é aprendizado de máquina nem previsão infalível da capacidade máxima.

**Domínio de controle:** provedor + conta + grupo de endpoints/capacidade + região quando a quota for regional. Se a quota é global, o domínio também é global. Endpoints que compartilham limite pertencem ao mesmo domínio; separar por tenant não multiplica quota. Contas realmente independentes podem possuir controladores separados. Mapear limites sobrepostos de provedor, conta e operação, aplicando o mais restritivo a cada chamada.

**Controles independentes:** taxa de envio efetiva (operações por segundo com burst limitado), concorrência HTTP simultânea e quantidade de operações assíncronas pendentes no provedor. Uma resposta de submit rápida não prova que a fila assíncrona está saudável; medir idade do pendente, latência de conclusão e taxa de finalizações. Retries, submit, status, fetch, cancelamento e sondagens consomem o orçamento da capacidade que utilizam. Reservar parcelas mínimas para acompanhar/concluir operações existentes, evitando que submissões novas impeçam sua conclusão.

| Sinal observado | Interpretação e ação do controle |
| --- | --- |
| 429 / Retry-After | Pressão ou quota excedida: reduzir rapidamente taxa, respeitar cooldown e orientação externa |
| 503, indisponibilidade ou erro de capacidade homologado | Reduzir taxa e concorrência; falha persistente abre circuito com poucas sondagens |
| Timeout após envio | Considerar saturação como hipótese e reduzir pressão; manter UNKNOWN e não criar retry inseguro |
| Crescimento sustentado de p95/RTT e pendência assíncrona | Reduzir admissão externa antes de acumular fila; verificar também rede e gargalo no próprio hub |
| Janela saudável, demanda real e folga de prazo | Aumentar taxa/concorrência em passos pequenos dentro dos tetos, observar e estabilizar |
| 400/422 de validação ou falha funcional | Não reduzir capacidade global automaticamente; investigar contrato/entrada |
| 401/403 / certificado inválido | Suspender vínculo/credencial afetados e alertar; mais retries não curam autenticação |
| Métricas ausentes ou antigas | Não ampliar limite; entrar em patamar conservador ou parar novas submissões conforme validade do orçamento |

Parâmetros iniciais **para ensaio**, não universalmente corretos: janela de controle 5 s, janela de observação 30 s, mínimo de 50 amostras para aumento, redução pela metade na saturação confirmada, aumento de até 5% após três janelas saudáveis e cooldown de 30 s. Definir por vínculo erro alvo, baseline de latência e margem tolerada. Baselines separam submit, status, fetch, tipo de operação e classes relevantes de tamanho; não interpretar mudança do mix para operações naturalmente lentas como saturação por si só. Sob baixo volume, usar intervalo maior/amostragem conservadora; não supor capacidade ilimitada por ausência de erros. Retry-After prevalece se mais restritivo. Histerese, suavização e jitter evitam oscilações e retomada simultânea de todos os clientes.

Manter piso operacional, limite inicial, teto técnico de segurança e teto contratual. O limite efetivo varia continuamente abaixo do teto; isso resolve a ineficiência de um rate limit fixo sem permitir ultrapassar contrato. Provedor que escala pode receber aumentos graduais após recuperação; se o limite técnico vigente impedir explorar nova capacidade, revisão controlada pode aumentá-lo. Não descobrir teto contratado por tentativa de violá-lo. Amortecimento não cancela chamadas que já chegaram ao provedor.

**Coordenação distribuída:** há um proprietário ativo/versionado da decisão por domínio de capacidade, com lease e epoch persistidos. Réplicas do Cometa aplicam concessões de tokens/slots contabilizadas globalmente, com validade e fencing; aumento de pods não multiplica o orçamento. Em dúvida sobre posse/validade, bloquear novas concessões ou usar apenas concessão ainda comprovadamente válida. Mudança de líder começa conservadora; liberar slot de chamada incerta apenas pela regra homologada, não por expiração cega de lease. O controlador é módulo do Cometa; domínios podem ser particionados entre instâncias.

A política de retry possui orçamento próprio dentro da taxa total para evitar amplificação. A fila não pode crescer ilimitadamente: deadline e espaço durável entram na admissão. Decisões do controlador registram limites anteriores/novos, amostra, causa, contrato e hora; valores agregados vão ao Prometheus e o histórico auditável fica no hub. Interfaces de consulta mostram limite efetivo, teto, utilização, motivo de redução, circuito e previsão estimada de espera.

SLA fornece o objetivo; o controlador usa sinais locais recentes e não espera a atualização de um dashboard para proteger o provedor. AWS discute a amplificação por retries e o uso de backoff/jitter; Envoy documenta adaptação de concorrência por sinal de latência. O algoritmo concreto aqui descrito é uma proposta do hub, não cópia do algoritmo do Envoy. [AWS — timeouts e retries](https://d1.awsstatic.com/builderslibrary/pdfs/timeouts-retries-and-backoff-with-jitter.pdf), [Envoy — adaptive concurrency](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/adaptive_concurrency_filter).

## OPE-08 · Isolamento de concorrência e células — proprietário: Arquitetura / SRE

Não é tecnicamente defensável prometer impacto absolutamente zero entre cargas que compartilham CPU, banco, rede, broker ou provedor. O requisito verificável é impedir apropriação da capacidade reservada e manter os SLOs dos tenants saudáveis durante sobrecarga isolada dentro do modelo qualificado. Disponibilidade do provedor compartilhado e falha de infraestrutura comum são dependências explícitas; não desaparecem com goroutines.

Célula é conjunto de deployments do Portal/Órbita/Cometa/Pulsar/Libra, filas, bancos core/finance, pools e orçamentos próprios. Continua havendo cinco aplicações lógicas, não uma nova aplicação por célula. Atlas é plano de controle fora do caminho por pedido. tenant_id determina uma célula ativa via projeção autenticada; consulta e execução usam a autoridade de DAD-11. Expansão provisiona novas células preventivamente; eventual migração automatizada exige drenagem, transferência de estado/idempotência e fencing, sem duas células executando o mesmo protocolo/contrato estrito.

| Classe | Proteção exigida | Limite que deve constar do contrato |
| --- | --- | --- |
| Compartilhada protegida | Filas lógicas/agendas por tenant e classe, distribuição justa ponderada, limites de CPU/tarefas/bytes/conexões, capacidade mínima reservada, teto e admissão | Ainda há recursos físicos comuns dentro da célula; qualificar ruído e limites |
| Célula dedicada | Deployments, filas e bancos próprios; nodes/pools e capacidade de rede reservados; controles de observabilidade/consulta separados | Serviços de nuvem/provedor ainda podem ser dependências comuns; necessidade de cluster/conta dedicados é contratual |
| Dedicada estrita | Além da célula, isolamento físico/cluster/conta conforme risco e capacidade contratada de provedor segregada | Não elimina falha externa/global; não prometer disponibilidade absoluta |

Particionar e limitar tráfego por tenant e por serviço/produto. Uma única fila FIFO compartilhada não pode impor bloqueio de cabeça de fila a todos. Em SQS, adotar filas separadas por célula/classe de isolamento e distribuição justa no scheduler; fila por tenant é requerida para classes que precisam dessa separação e deve respeitar quotas/custo. Produtos volumosos/CPU-intensivos usam pools próprios. Consultas GET, timers de deadline e recepção de callbacks têm capacidade reservada independente de submissões novas.

Bancos têm pools e orçamento de consultas por workload; relatórios administrativos/SLA usam réplicas/projeções em recursos separados, com lag informado. Telemetria possui limites para que inundação de logs de um cliente não derrube o runtime. Prepaid/limite financeiro estrito permanece no caminho daquele contrato e recebe autoridade/capacidade isoladas; processamento pós-pago dos demais não depende dessa reserva.

Teste proposto de ruído: tenant A envia 10 vezes sua quota por 15 min, com payload/CPU máximos e provedor degradado; B permanece na capacidade contratada. B deve manter disponibilidade e percentis contratados, sem consumo da reserva por A; meta adicional de variação de p95 ≤10% sobre baseline equivalente, a qualificar. Excesso não aceito de A recebe 429/503 explícito. Reprovar leva a aumentar isolamento ou reduzir oferta, não a prometer “sem impacto” sem evidência. [AWS — isolamento de workloads](https://d1.awsstatic.com/builderslibrary/pdfs/workload-isolation-using-shuffle-sharding.pdf).

## OPE-09 · Orçamento de latência e paralelismo — proprietário: Arquitetura / Engenharia

A latência controlável do hub é decomposta em borda/autorização, validação/tradução, commit de aceite, espera de fila/agenda, aquisição de slot, transporte, persistência externa observada, consolidação/projeção final e primeira entrega. Latência do provedor e rede externa são medidas separadamente, mas integram a experiência e o SLA do cliente. Não declarar baixa latência ponta a ponta apenas porque o gateway é rápido.

Metas p95 de aceite 250 ms e GET local 100 ms da OPE-03 permanecem propostas para o perfil informado. Distribuição inicial de orçamento de aceite: borda/autorização 40 ms, validação/tradução 60 ms, persistência 100 ms e margem de 50 ms. É alocação de planejamento, não soma estatística de percentis independentes. Percentis ponta a ponta são medidos no mesmo conjunto de requisições. Perfis legados e arquivos grandes têm ensaio específico e não podem herdar meta por suposição.

Manter versão de configuração em memória, conexões HTTP/2 persistentes quando compatíveis, pools delimitados, proximidade regional dos componentes e transações curtas. Paralelizar passos independentes, upload/validação por buffers e consumidores financeiros/entrega após final. Na v4, SYNC usa comando/resposta HTTPS diretos entre Órbita e Cometa após registros duráveis, sem espera por SQS no comando ou no final. Essa revisão remove a latência de fila obrigatória da v3, preservando outbox de fatos, deadline e o estado comum. Não remover commits, custo de autorização ou transformação do benchmark. ASYNC/AUTO mantêm trabalho em filas e respectivas métricas.

Reserva de finalização cobre captura de objeto obrigatório, transformação customizada, persistência e transição terminal. A política só agenda chamada se houver prazo/capacidade mínimos para terminar. Aferir separadamente tempo em fila interna e espera por amortecimento externo: ambos consomem deadline e não podem ser retirados silenciosamente do SLA do cliente.

## OPE-10 · Aferição de SLA em ambas as relações — proprietário: Produto / SRE / Financeiro

Manter dois contratos de aferição independentes: cliente→hub e hub→provedor. Registrar ainda SLOs internos por etapa. Violação do provedor não isenta automaticamente o hub; um provedor pontual pode coexistir com atraso de fila/consolidação do hub. Contrato define obrigação, marco, limite, exclusões aprovadas, calendário, compensação e fonte de evidência. Não tratar toda falha HTTP como quebra de prazo nem todo HTTP 200 como SLA cumprido.

Marcos mínimos: recebimento na borda, aceite durável, modo DIRECT/QUEUED, enfileiramento quando aplicável, aquisição de slot, envio externo, aceite externo observado, callback/status recebido e persistido, conteúdo pronto, projeção do cliente pronta, conclusão durável, primeira tentativa e ack do webhook. Calcular durações por relógios sincronizados, mantendo fontes e incerteza quando houver. O deadline é absoluto no hub; hora reportada pelo provedor não retroage o aceite de sucesso.

O Atlas deve oferecer consulta paginada por tenant, produto/serviço, provedor/conta, contrato/versão, período, célula, estado e motivo de quebra; detalhar linha do tempo, deadline, excedente em segundos, componente/parte atribuída, evidências e situação da contestação. Cliente vê apenas sua relação/execuções e campos comerciais autorizados. Perfil administrativo pode cruzar tenants sob SEG-04. Provedor só recebe visão autorizada de sua conta/contrato, se esse canal existir.

Painéis mostram quantidade elegível, cumpridos, violados, abertos ainda no prazo, abertos vencidos, expirados, indisponibilidade, p50/p95/p99 e faixa de tempo excedida. Taxa de quebra usa população elegível da coorte contratual; pedidos ainda no prazo não são sucessos presumidos. Retornos tardios, rejeições, exclusões e backlog devem aparecer, evitando viés de contar somente os concluídos rápidos.

Evidência de SLA fica em armazenamento durável, com idempotência por protocolo/operação/obrigação/versão, sem depender de tracing amostrado. Exportação por período tem versão, competência e relatório de lacunas/lag; crédito ou contestação referencia essa evidência. Meta proposta de atualização da visão operacional: até 60 s após fato persistido; encerramento por prazo e controle adaptativo não esperam essa projeção. Alertas preventivos ao atingir, por exemplo, 80% do orçamento devem ser configuráveis e não confundidos com violação já confirmada.

## OPE-11 · Preservação de informação e limites da garantia — proprietário: Arquitetura / Dados / SRE

Definir a garantia sobre entradas **confirmadas como aceitas**: requisição pública após commit; callback externo após recibo durável; mensagem consumida após commit de efeito/intenção. Tráfego recusado antes do aceite não vira obrigação invisível: há resposta explícita e registro seguro quando possível. O hub não consegue preservar bytes que nunca chegaram ou que um remetente perdeu sem repetir.

Para crash de processo, duplicação, indisponibilidade transitória de broker e falha de uma zona dentro da topologia homologada, o objetivo obrigatório é RPO zero dos fatos já confirmados, retomada e reconciliação sem perda silenciosa. Evidenciar persistência replicada compatível com esse compromisso. Publicação por outbox, inbox transacional, correlação, agendas duráveis, versões, checksums, custódia de objetos e dedup econômica se complementam; nenhum isoladamente prova a garantia.

Corpo de pedido, recibo e resultado volumoso devem estar confirmados no domínio de armazenamento exigido antes do respectivo ack/final. SQS não é o único arquivo de obrigações; expiração ou DLQ não elimina estado de trabalho. Reconciliar contagem de aceites contra protocolos, intenções, operações, finais, entregas e uso financeiro; qualquer lacuna é incidente, não somente métrica. Retorno rejeitado por SLA mantém evidência permitida, sem expurgo motivado apenas pela rejeição.

“Sem nenhuma perda” não pode abranger destruição de todas as cópias, expurgo autorizado, desastre fora do domínio contratado ou falha de um parceiro sem idempotência/retry. RPO zero regional depende de persistência síncrona/confirmada em regiões independentes para todos os dados necessários, arbitragem contra split-brain e validação de objetos. Isso adiciona latência e pode recusar novos aceites quando não há quorum/cópias exigidas. A referência regional v4 não demonstra esse perfil; P-08 bloqueia a venda dessa garantia até o desenho/ensaio correspondente. Backup assíncrono não satisfaz sozinho esse requisito.

O hub garante, no modelo qualificado, obrigação rastreável e efeitos locais idempotentes; não promete que todo provedor concluirá com sucesso nem efeito externo exatamente uma vez sem suporte. Em operação ambígua, preserva UNKNOWN/reconciliação e informa o encerramento contratual ao cliente quando cabível, sem inventar resultado ou apagar custo.

## OPE-12 · Escala automática de ponta a ponta — proprietário: Plataforma / Atlas / SRE

Objetivo operacional: crescimento rotineiro de clientes/provedores dentro dos envelopes aprovados aciona capacidade por automação, sem chamado para criar instância, fila ou banco. A arquitetura lógica permanece estável; o número e o tamanho dos recursos aumentam. Configuração inicial, aprovação de orçamento/quotas e homologação de novas classes são responsabilidades de engenharia/plataforma, não intervenções manuais por cliente. Não confundir expansão automática com ausência de limites físicos ou comerciais.

| Camada | Sinal e automação | Proteção e responsabilidade |
| --- | --- | --- |
| APIs SYNC/GET/admissão | HPA: concorrência ativa, uso de recursos e sinais qualificados; capacidade mínima permanente | Engenharia define limites por instância; Plataforma garante nós/pools e fonte de métricas; sem scale-to-zero em rotas críticas |
| Workers ASYNC/polling/Pulsar/Libra | KEDA: backlog, idade, timers devidos e atraso de fatos; um controlador por Deployment | Não escalar consumo acima de banco, domínio de provedor ou capacidade de entrega |
| Nós | Karpenter: pods sem capacidade e demanda prevista coberta por reserva aquecida | Pools/zonas/tipos homologados, quotas de conta/IP/rede, N−1 e orçamento; capacidade crítica não depende apenas de Spot |
| Células/core/finance | Política Atlas/Plataforma: capacidade reservada, previsão e saturação antecipada; Crossplane cria perfil homologado | Banco/filas novos prontos antes de admitir tenants; pools e capacidade de escrita são gates próprios |
| Objetos e histórico | Quotas, crescimento/retention, particionamento e políticas automáticas aprovadas | Crescimento de storage não remove limite de throughput, custo ou prazo de custódia |
| Conta de provedor | Cometa: pressão adaptativa e orçamento hierárquico | Escala do Hub não cria capacidade do parceiro nem autoriza ultrapassar contrato |
| Configuração/observabilidade | Projeções incrementais, paginação, réplicas e retenção por volume | Milhares de vínculos não viram consulta completa por requisição nem labels sem controle |

A política prevê o tempo de expansão: demand_headroom deve cobrir crescimento esperado durante o p99 medido de provisionamento/qualificação, além da margem N−1 e de rajada. Se provisionar célula demora dezenas de minutos, não esperar o banco atingir seu limite para começar. Exemplos para ensaio: iniciar expansão quando reservas excederem 60% do envelope N−1 ou quando a previsão atingir 80% dentro do horizonte de provisionamento. Os números são pontos iniciais, sujeitos ao mix e custo; nunca limites comerciais presumidos.

Reserva e health de capacidade são distintas de uso momentâneo. HPA pode ajustar pods rapidamente, enquanto bancos/células são preparados antes. Redução tem histerese, período mínimo, drenagem e proteção de obrigações; não destruir banco/objeto porque a fila está vazia. Reconciliar drift de infraestrutura sem revogar recursos de produção não previstos para exclusão. Um controller de expansão indisponível não retira capacidade ativa: acionar alarme de horizonte de folga e manter mínimos.

Quotas de nuvem, endereços, conexões, filas, certificados, IAM e limite de orçamento são inventariados automaticamente. Pedidos de aumento podem ser automatizados onde a plataforma permitir, mas aprovação externa não é garantida. Reservar essas quotas antecipadamente para o crescimento comercial contratado; impedimentos são excepcionais e visíveis antes de afetar clientes ativos. O teste de escala deve provar o fluxo sem intervenção humana e registrar o que acontece quando uma quota é negada.

Conta compartilhada entre células mantém uma autoridade de orçamento do Cometa por domínio de capacidade, com concessões limitadas, persistidas, epoch e validade. Distribuir lotes de tokens/slots evita round trip global por chamada, mas não autoriza conceder mais que o limite total. As reservas em voo desconhecidas não são automaticamente devolvidas. Falha da autoridade admite somente concessões ainda comprovadamente válidas; depois, novas submissões daquele domínio aguardam/expiram segundo contrato. Isso é dependência real da conta compartilhada e recebe redundância/capacidade próprias; credenciais dedicadas não eliminam um teto global comum do mesmo provedor.

## OPE-13 · Matriz de dependências e fallback seguro — proprietário: Arquitetura / SRE

Fallback tem pré-condição, prazo e limite de consistência. Não é “continuar a qualquer custo”. Os estados operacionais são NORMAL, DEGRADADO_COM_CONTINUIDADE, ADMISSAO_RESTRITA e RECUPERANDO por capacidade/célula/vínculo. Não usar uma flag global para derrubar todos os serviços por falha isolada. SLO mede a experiência efetiva, inclusive erros de restrição técnica.

| Dependência indisponível | O que pode continuar | Restrição e retorno seguro | Recuperação/limite |
| --- | --- | --- | --- |
| Redis | SYNC, ASYNC, GET e controle pelo caminho sem L2 | Bypass com timeout/breaker; nenhuma informação crítica exclusiva | Reaquecer gradualmente; capacidade de origem com cache frio comprovada |
| Atlas / banco de controle | Pedidos de contratos/configurações já válidos e rotas conhecidas | Novas publicações/onboarding aguardam; autorização revogada/velha não é ignorada | Projeções, versão e validade de segurança verificadas; retomada incremental |
| PostgreSQL core writer | Final já confirmado, autorizado e adequado em cópia de leitura; execuções já enviadas podem ainda produzir efeito externo | Não iniciar novos efeitos sem intenção/posse duráveis; sem ack de callback sem recibo; não declarar final perdido em memória | Failover/reconexão limitada; recuperar por chave e reconciliar UNKNOWN; falta de final comprovado retorna 503/erro de transporte |
| Banco financeiro / Libra | Pós-pago com fatos duráveis e exposição autorizada; GET e clientes sem reserva estrita | Novas reservas estritas não aprovadas aguardam ou falham no prazo; sem saldo inventado | Drenar fatos, dedup e conciliar antes de fechamento; teto por bytes/idade/valor de exposição |
| SNS/SQS | SYNC direto enquanto core/outbox possuem capacidade; GET e registro de fatos/recibos duráveis | ASYNC só é aceito com slot durável e política qualificada que comporte atraso; caso contrário 503 antes de efeito | Outbox conserva intenção; drain com limites; TTL/SLA seguem correndo; assinatura 202 não prova conclusão |
| S3 / objetos | Operações de payload pequeno sem objeto obrigatório e resultados locais já íntegros | Upload/captura obrigatória aguarda ou falha no prazo; não finalizar sucesso com referência indisponível não confirmada | Recuperar por object_id/checksum; não buscar no provedor por GET; download indisponível tem SLI próprio |
| Cofre / KMS | Vínculos com material já obtido, utilizável e autorizado dentro da validade | Instância fria ou chave inválida bloqueia apenas binding afetado; não usar outra conta | Pré-aquecimento, refresh antecipado, sobreposição e validade; ciphertext sem chave disponível não é fallback funcional |
| IdP / distribuição de autorização | Tokens/identidades já válidos e prova de autorização dentro da política | Não emitir nova identidade por conta própria; escopo sem validade comprovada é negado | Cache de chaves públicas com rotação/validade; sessões novas dependem de emissão; revogação não é adiada indefinidamente |
| Controlador de quota externa | Concessões globais previamente autorizadas ainda válidas | Nenhum pod inventa novos tokens/slots ou amplia teto durante partição | Eleição/fencing com checkpoint; reconciliação de chamadas em voo; reinício conservador |
| Provedor específico | Outros vínculos/células saudáveis; GET de qualquer final local autorizado | Retry TTL, circuit breaker, redução de pressão; UNKNOWN não é failover seguro | Sondagens graduais; polling/callback e reconciliação; retorno tardio rejeitado |
| Endpoint de webhook do cliente | Processamento e consulta local; outras entregas | Conservar obrigação, retry e EXHAUSTED; não reexecutar produto | Reenvio da mesma representação/event_id após recuperação |
| Prometheus/Loki/Grafana/Tempo | Processamento com limites locais e evidência de negócio durável | Buffers de telemetria limitados; descarte técnico contabilizado, sem bloquear por log | Reconstruir métricas possíveis; controlar queda de telemetria; auditoria/financeiro não dependem de logs amostrados |
| Kubernetes API / Crossplane / Karpenter | Workloads e células já saudáveis | Mudanças, nova expansão e reposição podem ficar limitadas | Capacidade aquecida, controladores redundantes e alarme de folga; runtime não consulta API Kubernetes por pedido |

Disponibilidade parcial deve ser roteada por função. Órbita expõe deployments/rotas distintos de admissão-execução e leitura; continuam a mesma aplicação/domínio, com pools e readiness específicos. Cometa separa recepção durável de callback e workers de provedor quando necessário. Uma readiness booleana de um pod compartilhado não expressa todas essas capacidades: se a separação não existir, não alegar continuidade seletiva que o balanceador não consegue oferecer. Liveness não depende de Redis, cofre, provedor ou broker externos.

Fallback de leitura não torna novas escritas disponíveis. Se o retorno do provedor chega quando não há custódia autoritativa, tentar persistência dentro do orçamento e usar capacidades externas de reentrega/status; sem essas capacidades, a resposta pode ficar desconhecida. Essa limitação deve bloquear ofertas críticas incompatíveis, não ser omitida como “garantia de mensagem”. Circuitos, buffers e retries são limitados por bytes, tempo, idade e risco; esgotamento recusa antes de novo aceite, preservando obrigações já confirmadas.

## OPE-14 · Manutenção sem indisponibilidade evitável — proprietário: Plataforma / SRE / Engenharia

Toda dependência tem plano de atualização, compatibilidade de versão, teste em ppd, sequência de substituição, condição de aborto e rollback/roll-forward. Manutenção faz parte do cálculo de disponibilidade, sem exclusão automática. Antes da janela, verificar saúde das cópias, capacidade N−1, backups/restauração, backlog, validade de credenciais e prazo dos protocolos em voo. Não atualizar simultaneamente todas as zonas/cópias ou componentes correlacionados.

| Alvo | Procedimento obrigatório | Evidência de aceite |
| --- | --- | --- |
| APIs e workers | Canary/rolling com surge e reserva; retirar prontidão, drenar conexões/tarefas e encerrar com prazo maior que a drenagem homologada | Chamadas SYNC existentes concluídas ou explicitamente recuperáveis; nenhuma perda de intenção/efeito; client retry idempotente ensaiado |
| Redis | Ativar bypass, atualizar/reiniciar e retornar por aquecimento gradual | Mesmos SLOs no perfil qualificado, sem avalanche de conexões no core |
| PostgreSQL | Atualização/failover suportados, migração de schema expand/contract, reconexão e recuperação de transação incerta | Tempo real de indisponibilidade de escrita, perda zero no domínio qualificado, ausência de efeito duplicado |
| Broker / objetos | Mudança compatível e políticas preservadas; teste de indisponibilidade e drain de outbox | Obrigações/finais íntegros; replay não altera resultado/custo |
| Credencial / certificado | Sobreposição e pré-aquecimento; novos usos na versão válida, operações em voo correlacionadas | Sem vazamento ou troca de conta; revogação e cold start ensaiados |
| Nós / zona | Drenagem controlada, distribuição, N−1 e interrupção limitada de consolidação | Capacidade crítica continua; um PDB não substitui capacidade ou protege toda falha involuntária |
| Plano de gestão e observabilidade | Atualizar réplicas em sequência sem desligar processamento | Runtime conserva configurações válidas; alertas externos/sintéticos detectam lacunas |

RDS Multi-AZ com failover de instância pode levar tipicamente 60–120 segundos, e mais em certas condições; a documentação também exige restabelecer conexões. Isso não é uma garantia de tempo máximo nem uma propriedade de todo tipo de cluster PostgreSQL. Um perfil SYNC que tolere menos que a interrupção medida não pode ser qualificado apenas com a declaração “RDS multi-AZ”. [AWS — failover de instância Multi-AZ](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.Failover.html).

A seleção física de implantação PostgreSQL deve demonstrar o RTO exigido no perfil, incluindo refresh de endpoint, conexões, transações e cache frio. Proxy/pool pode reduzir reconexão, mas não autoriza confirmar commit sem writer nem migra uma transação em voo magicamente. Se a manutenção exceder o orçamento síncrono, o evento é falha de disponibilidade mensurada; erro explícito com protocolo é recuperação responsável, não “zero downtime” comprovado.

Migração irreversível de dados exige mecanismo de roll-forward/restauração e verificação; não presumir que voltar a imagem reverte schema. Rollouts e expansão automática devem respeitar as mesmas políticas de drenagem. Automação pode interromper atualização ao detectar risco, preservando tráfego existente; intervenção extraordinária de incidente continua possível, mas não é o mecanismo ordinário de crescimento.

## OPE-15 · Criticidade, tolerância a falhas e continuidade operacional — proprietário: Produto / Arquitetura / SRE

Indisponibilidade, atraso excessivo, resposta incorreta, uso de resultado antigo como atual, repetição de efeito externo e vazamento entre clientes são riscos críticos do Hub. Serviço que possa participar de operação com impacto em vida/segurança exige perfil de criticidade formal antes de contratação/produção. A análise deve envolver o responsável do sistema consumidor: o Hub não conhece sozinho o tempo máximo tolerável ou a consequência de uma resposta ausente/incerta.

| Dimensão do perfil | Definição exigida e critério |
| --- | --- |
| Função e consequência | Operação consumidora, efeito externo e consequência de ausência, atraso, duplicação ou dado desatualizado; responsável nominal pela aceitação de risco |
| Prazo e continuidade | Máximo tolerável de interrupção em segundos, SLA final por modalidade, política de erro/UNKNOWN e recuperação do consumidor |
| Falhas cobertas | Processo, nó, zona, writer, broker, cofre, identidade, provedor, rede/partição, região e falhas correlacionadas; não apenas pod reiniciado |
| Custódia | RPO exigido por pedido, objeto, resultado, recibo, idempotência e fato econômico; confirmação em cópias adequadas antes do ack |
| Capacidade | Carga crítica contratada com N−1, cold cache, queda do plano de controle e pico de outros tenants; reservas segregadas |
| Operação | Detecção, mitigação automática, escalonamento, plantão, exercícios e comunicação contratual de falha; uso de contingência testado |
| Evidência | Ensaio observado de cada cenário no perfil real; risco residual, limite de oferta e aprovação antes do gate prd |

A proposta de 99,99% equivale a cerca de 4 min 19 s de indisponibilidade em 30 dias; percentual mensal não estabelece quanto dura uma falha individual nem prova adequação a esse uso. O perfil crítico pode exigir metas mais fortes e RTO muito menor. P-10 resolve esses limites; não herdar os RTOs gerais de cinco minutos/quatro horas como se fossem adequados. Disponibilidade de SYNC mede final válido dentro do orçamento na conexão, além da disponibilidade isolada de POST/GET; 202 rápido não conta como sucesso de SYNC.

Arquitetura regional: células multi-AZ, réplicas aquecidas, bancos/filas/objetos com durabilidade qualificada, leitura independente, recuperação automática e isolamento reduzem falhas evitáveis. Não depender de restaurar backup ou criar a primeira réplica depois do incidente para sustentar o caminho crítico. Aplicações de classe crítica usam capacidade reservada e domínios externos compatíveis; provedor único sem mecanismo de recuperação/reentrega de resultado é risco explícito de ponta a ponta.

Se o perfil exigir sobreviver à perda de região com RPO zero e RTO muito curto, a referência regional é insuficiente. A decisão deverá detalhar autoridade transacional tolerante à perda regional, confirmação das cópias/objetos em regiões independentes, capacidade ativa/aquecida, quorum/fencing e roteamento qualificado. Durante partição, apenas o lado com autoridade válida pode aceitar efeitos; servir ambos com escritas independentes ameaça idempotência e saldo. Duplicar EKS ou replicar banco assincronamente não satisfaz esse requisito. A troca da topologia de dados seria uma decisão estrutural motivada por novo domínio de falhas, não uma etapa ordinária de crescimento de clientes.

O contrato de contingência do consumidor deve definir uso permitido do último resultado confirmado, idade máxima, resposta de indisponibilidade e recuperação com a mesma chave; o Hub não fabrica dado novo ou sucesso para encobrir falha. Sem persistência/autoridade suficiente, restringe novas operações e informa a situação. Isso pode representar indisponibilidade, mas evita afirmar processamento ou estado inexistente. Segurança de efeito e integridade não são abandonadas para melhorar o indicador.

Gate bloqueador: não ativar oferta crítica cuja interrupção máxima, RPO, dependência externa ou comportamento de contingência não tenham evidência compatível. A revisão é uma especificação de engenharia, não certificação de segurança funcional. DEC-05 separa risco mitigado no desenho e risco ainda não comprovado por ensaio.
