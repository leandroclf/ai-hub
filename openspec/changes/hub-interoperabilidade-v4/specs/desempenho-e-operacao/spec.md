# Desempenho, alta disponibilidade e operação

## ADDED Requirements

### Requirement: OPE-01 — Modelo de capacidade
O sistema SHALL dimensionar capacidade por requisições por segundo, duração, fan-out, quantidade de polls, tamanho de conteúdo, concorrência por provedor e prazo de recuperação, mantendo perfil medido por classe de produto/serviço e por tenant que inclua máximos e não apenas médias. O sistema SHALL calcular relações de planejamento (operações externas/s, consultas de status/s, concorrência de chamadas) considerando rajadas, caudas e variância, reconhecendo que o caminho crítico limita a duração mesmo com paralelismo alto. O sistema SHALL medir separadamente arquivos grandes quanto a carga mista, fluxo contínuo, limite máximo por adaptador, throughput de bytes e uso de memória, com entrada sintética incluindo limites, payloads inválidos e respostas grandes.

#### Scenario: Cálculo de operações pendentes e custo de polling
- GIVEN um perfil de 1.000 pedidos/s com três passos por pedido
- WHEN até 1.000 operações/s ficam pendentes por 60 segundos e são consultadas a cada 5 segundos
- THEN o sistema SHALL considerar aproximadamente 60.000 operações pendentes e cerca de 12.000 polls/s no dimensionamento de capacidade e na quota do provedor, sem habilitar polling que ignore esse custo

#### Scenario: Qualificação por perfil misto de carga
- GIVEN os perfis iniciais propostos (1.000 admissões/s por 60 minutos, rajada de 2.000/s por 5 minutos, 2.000 GETs/s, teste de 8 horas a 50% da carga nominal)
- WHEN a qualificação de capacidade é executada com mix de 70% serviço simples, 20% agregação de três passos e 10% composição de cinco passos
- THEN o sistema SHALL registrar perfil de latência e proporção assíncrona observados, sem tratar esses números como SLA ou escala real antes de P-02

#### Scenario: Ensaio isolado de arquivos grandes
- GIVEN um adaptador que processa uploads de arquivos grandes
- WHEN o ensaio de capacidade é executado
- THEN o sistema SHALL medir throughput de bytes e uso de memória junto da taxa de pedidos, com entradas sintéticas que incluam limites, payloads inválidos e respostas grandes, sem herdar o perfil de capacidade de payloads mínimos em memória

### Requirement: OPE-02 — Estratégia de baixa latência
O sistema SHALL manter a persistência como parte do caminho de aceite, sem removê-la para melhorar benchmark, e SHALL reduzir overhead por meio de pools limitados e reutilizados, conexões HTTP persistentes, projeções locais de configuração, índices nos caminhos críticos, transações curtas, publicação/consumo em lotes controlados e processamento paralelo limitado quando não houver dependência funcional entre passos. O sistema SHALL impor concorrência máxima por pod, serviço, provedor/conta e tenant, incluindo orçamento de conexões ao PostgreSQL, de modo que a soma de pools no máximo de réplicas caiba no orçamento de banco e que autoscaling não exaura conexões nem quotas externas globais por conta.

#### Scenario: Consultas finais usam fonte autoritativa sem provedor
- GIVEN uma consulta de resultado final de um produto
- WHEN a consulta é atendida
- THEN o sistema SHALL responder a partir do banco autoritativo ou objeto do hub, sem consultar o provedor, e relatórios pesados SHALL usar projeção/réplica analítica sem disputar recursos com admissão/consulta

#### Scenario: Cache opcional não vira fonte única de correlação
- GIVEN um cache configurado para configuração imutável ou resultado versionado
- WHEN o cache está indisponível ou é invalidado
- THEN o sistema SHALL continuar operando corretamente, pois o cache MUST NOT ser a única fonte de correlação, ledger ou outbox

#### Scenario: Orçamento justo de conexões sob autoscaling
- GIVEN um cliente de alto volume concorrendo com outros tenants pelo pool de conexões ao PostgreSQL
- WHEN o autoscaling aumenta o número de réplicas
- THEN o sistema SHALL aplicar distribuição justa e orçamento separado de submissão/polling/entrega, impedindo que esse cliente esgote todos os slots e garantindo que a soma de pools não ultrapasse o orçamento de banco

### Requirement: OPE-03 — SLOs e semântica de garantia
O sistema SHALL medir e reportar os indicadores de serviço definidos (disponibilidade de admissão/GET local, latência de aceite ASYNC, consulta de resultado local, final ASYNC, final SYNC, primeira tentativa de webhook, atualidade financeira pós-paga, polling, recuperação de falha de zona, desastre regional e obrigações sem rastreio) segundo os critérios de medição e delimitação especificados, sem excluir falhas de provedor do relatório ponta a ponta para inflar disponibilidade percebida.

#### Scenario: Disponibilidade de admissão e GET local
- GIVEN requisições válidas elegíveis de admissão e GET local
- WHEN o serviço mede disponibilidade mensal por API, célula e região
- THEN o sistema SHALL contar 5xx, timeouts e 429 por falta de capacidade do Hub, além de janelas de manutenção, como falha, perseguindo a meta inicial proposta de 99,99% mensal, sem promover perfil crítico sem aprovação específica

#### Scenario: Latência de aceite ASYNC e consulta de resultado local
- GIVEN uma requisição ASYNC dentro da região e do perfil qualificado, sem transferência de arquivo grande
- WHEN a latência de aceite é medida da chegada à borda até a resposta após commit
- THEN o sistema SHALL manter p95 ≤ 250 ms e p99 ≤ 500 ms para o aceite, e p95 ≤ 100 ms e p99 ≤ 250 ms para consulta de resultado local com metadados/JSON até 64 KiB, tratando download de objeto com medição própria

#### Scenario: Final ASYNC, primeira tentativa de webhook e atualidade financeira
- GIVEN um resultado final ASYNC observado no Cometa, um resultado final persistido aguardando entrega, e um fato econômico pós-pago confirmado
- WHEN cada marco é medido até seu respectivo desfecho (resultado durável na Órbita; início da tentativa de webhook; apuração financeira visível)
- THEN o sistema SHALL manter p95 ≤ 2 s / p99 ≤ 5 s para o final ASYNC, p95 ≤ 5 s para a primeira tentativa de webhook (sem medir confirmação do destino), e p95 ≤ 60 s para atualidade financeira pós-paga, exigindo completude (não percentil) para fechamento

#### Scenario: RTO de zona e de região não substituem qualificação crítica
- GIVEN um exercício de falha de zona ou de desastre regional
- WHEN o RTO de referência geral (≤ 5 min para zona; ≤ 4 h para região) é avaliado
- THEN o sistema SHALL demonstrar RPO por configuração e SHALL NOT adotar automaticamente esses RTOs de referência no perfil crítico, bloqueando qualquer promessa de RPO zero regional até qualificação específica sob P-08

### Requirement: OPE-04 — Ambientes e alta disponibilidade
O sistema SHALL manter os ambientes local, dev, hom, ppd e prd com as topologias, isolamento de dados e finalidades especificadas, promovendo o mesmo digest de imagem entre ambientes com apenas configuração e segredos variando. O sistema SHALL manter réplicas mínimas de duas em dev/hom e três distribuídas em três zonas em ppd/prd, com capacidade de continuar operando dentro da carga crítica após perder uma zona, e SHALL manter requests/limits, probes e configurações explícitas em todos os ambientes.

#### Scenario: Dependências gerenciadas com HA real antes do gate ppd
- GIVEN o ambiente ppd em preparação para qualificação de carga
- WHEN o desenho físico de PostgreSQL, filas, objetos, Prometheus, Loki, Tempo e Grafana é avaliado
- THEN o sistema SHALL exigir failover gerenciado, backups testados, réplicas independentes e topologia HA compatível para cada dependência, sem aceitar "um container de cada" como evidência de alta disponibilidade

#### Scenario: Kind e emulação local não substituem serviços gerenciados
- GIVEN um ambiente local usando Docker Compose e kind como laboratório Kubernetes
- WHEN alguém propõe usar essa emulação como prova de limites, IAM, durabilidade ou failover de serviços gerenciados
- THEN o sistema SHALL rejeitar essa prova como insuficiente, exigindo que qualquer intenção de infraestrutura integralmente autogerenciada no Kubernetes seja alterada explicitamente em P-01 antes da implementação

#### Scenario: Isolamento por célula complementa e não substitui separação de ambiente
- GIVEN aplicações distribuídas em múltiplos pods em ppd/prd
- WHEN o isolamento de carga é avaliado
- THEN o sistema SHALL aplicar as células e classes de isolamento de OPE-08 como complemento, pois separar somente pods MUST NOT ser considerado suficiente para isolar banco, filas ou contas de provedor

### Requirement: OPE-05 — Probes, escala e proteção
O sistema SHALL implementar probes de startup, readiness e liveness com propósitos distintos — readiness avaliando capacidade de aceitar responsabilidade com verificações autenticadas e prazo curto, liveness identificando apenas processo travado sem reiniciar por indisponibilidade de provedor — e SHALL separar health de diagnóstico detalhado sem expor credenciais ou topologia interna. O sistema SHALL aplicar HPA para APIs, KEDA para workers de fila (um controlador de escala por Deployment, sem HPA paralelo concorrente), distribuição por zona/nó, PodDisruptionBudget, rolling update com capacidade adicional, encerramento gracioso, backpressure, circuit breaker e bulkheads.

#### Scenario: Readiness por capacidade evita que falha de envio derrube leitura saudável
- GIVEN um endpoint que atende capacidades com dependências diferentes (por exemplo, envio de callback e consulta local)
- WHEN a dependência de envio de callback falha
- THEN o sistema SHALL manter a consulta local saudável respondendo normalmente, pois o desenho de rotas/workers MUST impedir que a falha de uma capacidade derrube outra através de uma readiness booleana compartilhada

#### Scenario: Escala de workers por backlog e idade da obrigação
- GIVEN workers ASYNC/polling gerenciados por KEDA
- WHEN o backlog ou a idade da obrigação cresce
- THEN o sistema SHALL escalar workers com base nessa métrica externa e na quantidade de timers vencidos, respeitando mínimo/máximo, estabilização de redução e teto imposto pela capacidade de banco/provedor, sem permitir HPA paralelo concorrente no mesmo Deployment

#### Scenario: Drenagem preserva UNKNOWN em chamada externa incerta
- GIVEN um worker em processo de encerramento gracioso durante rolling update
- WHEN há uma chamada externa em andamento com resultado incerto no momento da drenagem
- THEN o sistema SHALL parar de reclamar novo trabalho, terminar ou devolver posse com segurança, e preservar o estado UNKNOWN em vez de declarar sucesso ou falha sem confirmação

### Requirement: OPE-06 — Observabilidade e runbooks
O sistema SHALL expor métricas de baixa cardinalidade (admissões, recusas por causa, latências por etapa, erros canônicos, idade de outbox/backlog/timers, workers ativos, saturação de pools, operações UNKNOWN, callbacks órfãos/conflitantes, entregas esgotadas, resultados não disponíveis, uso não apurado e divergências de fechamento) com dimensões limitadas a ambiente, componente, serviço e provedor, mantendo protocol_id, event_id e dados pessoais fora de labels de métricas. O sistema SHALL manter os alertas iniciais propostos com responsável e ação definidos, e SHALL manter runbooks obrigatórios para cada cenário de indisponibilidade listado, cada um com gatilho, diagnóstico seguro, ação autorizada, limite de automação, rollback, evidência e critério de encerramento.

#### Scenario: Alerta de outbox envelhecida preserva pendência
- GIVEN a idade da outbox mais antiga excedendo 30 segundos por 5 minutos
- WHEN o alerta dispara
- THEN o sistema SHALL notificar SRE para verificar publicador/broker, preservando os eventos pendentes sem apagá-los

#### Scenario: Divergência financeira bloqueia fechamento incompleto
- GIVEN uma apuração financeira atrasada por mais de 5 minutos ou uma divergência não classificada
- WHEN o alerta correspondente dispara
- THEN o sistema SHALL acionar Financeiro/Engenharia para verificar fatos, dedup e contrato, bloqueando o fechamento incompleto até resolução

#### Scenario: Auditoria de negócio não depende de amostragem de trace
- GIVEN traces coletados pelo Tempo com amostragem controlada
- WHEN uma auditoria durável de negócio precisa de evidência
- THEN o sistema SHALL fornecer essa evidência a partir de registros de auditoria duráveis, sem depender de traces amostrados que possam ter descartado o evento relevante

### Requirement: OPE-07 — Amortecimento e controle adaptativo do provedor
O sistema SHALL manter, no Cometa, um controlador de feedback adaptativo (AIMD) por domínio de controle (provedor + conta + grupo de endpoints/capacidade + região quando a quota for regional) que ajuste taxa de envio efetiva, concorrência HTTP simultânea e quantidade de operações assíncronas pendentes, respeitando piso operacional, limite inicial, teto técnico de segurança e teto contratual, sem jamais ultrapassar o teto contratado por tentativa. O sistema SHALL interpretar os sinais observados (429/Retry-After, 503, timeout, crescimento de p95/pendência, janela saudável, 400/422, 401/403, métricas ausentes) conforme a tabela de ação definida, e SHALL coordenar decisões entre réplicas do Cometa por meio de um proprietário ativo/versionado por domínio, com lease e epoch persistidos e fencing.

#### Scenario: Redução multiplicativa sob 429 e Retry-After
- GIVEN uma resposta 429 com cabeçalho Retry-After de um provedor
- WHEN o controlador processa esse sinal
- THEN o sistema SHALL reduzir rapidamente a taxa de envio, respeitar o cooldown e a orientação externa, e o Retry-After SHALL prevalecer quando mais restritivo que o cooldown padrão de 30 s

#### Scenario: Aumento gradual em janela saudável
- GIVEN uma janela de observação de 30 segundos com no mínimo 50 amostras saudáveis e demanda real com folga de prazo
- WHEN três janelas saudáveis consecutivas são confirmadas
- THEN o sistema SHALL aumentar a taxa/concorrência em passos de até 5%, dentro dos tetos definidos, observando e estabilizando antes de novo aumento

#### Scenario: Falha funcional não reduz capacidade global
- GIVEN uma resposta 400/422 de validação ou falha funcional de um provedor
- WHEN o controlador processa esse sinal
- THEN o sistema SHALL NOT reduzir a capacidade global automaticamente, direcionando a investigação para contrato/entrada em vez de amortecimento

#### Scenario: Coordenação distribuída em mudança de líder
- GIVEN uma mudança de líder do controlador de um domínio de capacidade compartilhado entre réplicas do Cometa
- WHEN a posse ou validade da concessão está em dúvida
- THEN o sistema SHALL bloquear novas concessões ou usar apenas concessão ainda comprovadamente válida, começando de forma conservadora e liberando slot de chamada incerta apenas pela regra homologada, nunca por expiração cega de lease

### Requirement: OPE-08 — Isolamento de concorrência e células
O sistema SHALL organizar a capacidade em células (conjunto de deployments do Portal/Órbita/Cometa/Pulsar/Libra, filas, bancos core/finance, pools e orçamentos próprios) selecionadas por tenant_id via projeção autenticada, sem prometer impacto absolutamente zero entre cargas que compartilham CPU, banco, rede, broker ou provedor. O sistema SHALL impedir apropriação da capacidade reservada e manter os SLOs dos tenants saudáveis durante sobrecarga isolada, aplicando as classes Compartilhada protegida, Célula dedicada e Dedicada estrita com as proteções e limites contratuais especificados para cada uma.

#### Scenario: Classe Compartilhada protegida impõe limites por tenant
- GIVEN tenants operando na classe Compartilhada protegida
- WHEN a carga de um tenant cresce
- THEN o sistema SHALL aplicar filas lógicas/agendas por tenant e classe, distribuição justa ponderada, limites de CPU/tarefas/bytes/conexões e capacidade mínima reservada, reconhecendo que recursos físicos comuns dentro da célula ainda existem

#### Scenario: Célula dedicada isola deployments e bancos
- GIVEN um tenant provisionado em Célula dedicada
- WHEN outro tenant em célula distinta sofre sobrecarga
- THEN o sistema SHALL manter deployments, filas e bancos próprios, nodes/pools e capacidade de rede reservados para o tenant dedicado, mesmo reconhecendo que serviços de nuvem/provedor compartilhados ainda são dependências comuns

#### Scenario: Dedicada estrita não elimina falha externa global
- GIVEN um tenant contratado na classe Dedicada estrita com isolamento físico/cluster/conta
- WHEN uma falha de provedor global ou de infraestrutura comum ocorre
- THEN o sistema SHALL comunicar essa falha como risco explícito, sem prometer disponibilidade absoluta apesar do isolamento físico

#### Scenario: Teste de ruído entre tenants
- GIVEN o tenant A enviando 10 vezes sua quota por 15 minutos com payload/CPU máximos e provedor degradado, enquanto o tenant B opera na capacidade contratada
- WHEN o teste de ruído é executado
- THEN o sistema SHALL manter a disponibilidade e os percentis contratados de B sem consumo da reserva por A, recusando o excesso não aceito de A com 429/503 explícito, e qualquer reprovação SHALL levar a aumentar isolamento ou reduzir oferta, nunca a declarar "sem impacto" sem evidência

### Requirement: OPE-09 — Orçamento de latência e paralelismo
O sistema SHALL decompor a latência controlável do hub em borda/autorização, validação/tradução, commit de aceite, espera de fila/agenda, aquisição de slot, transporte, persistência externa observada, consolidação/projeção final e primeira entrega, medindo a latência do provedor e da rede externa separadamente mas integrando-a à experiência e ao SLA do cliente. O sistema SHALL manter a distribuição inicial de orçamento de aceite (borda/autorização 40 ms, validação/tradução 60 ms, persistência 100 ms, margem 50 ms) como alocação de planejamento, e SHALL usar, no fluxo SYNC da v4, comando/resposta HTTPS diretos entre Órbita e Cometa após registros duráveis, sem espera por SQS no comando ou no final.

#### Scenario: Percentis ponta a ponta medidos no mesmo conjunto de requisições
- GIVEN as metas de p95 de aceite (250 ms) e GET local (100 ms) definidas em OPE-03
- WHEN os percentis ponta a ponta são calculados
- THEN o sistema SHALL medi-los sobre o mesmo conjunto de requisições, sem tratar a soma de orçamentos por etapa como soma estatística de percentis independentes

#### Scenario: SYNC remove fila obrigatória sem remover outbox de fatos
- GIVEN uma requisição SYNC na v4
- WHEN a resposta é processada entre Órbita e Cometa
- THEN o sistema SHALL usar comando/resposta HTTPS diretos após registros duráveis, preservando outbox de fatos, deadline e o estado comum, sem remover commits, custo de autorização ou transformação do benchmark de latência

#### Scenario: Reserva de finalização considera prazo e capacidade mínimos
- GIVEN uma operação que exige captura de objeto obrigatório, transformação customizada e persistência antes da transição terminal
- WHEN o Cometa avalia se agenda a chamada de finalização
- THEN o sistema SHALL agendar a chamada somente se houver prazo/capacidade mínimos para terminar, aferindo separadamente o tempo em fila interna e a espera por amortecimento externo, sem retirar silenciosamente esse tempo do SLA do cliente

### Requirement: OPE-10 — Aferição de SLA em ambas as relações
O sistema SHALL manter dois contratos de aferição independentes — cliente→hub e hub→provedor — além de SLOs internos por etapa, sem tratar violação do provedor como isenção automática do hub nem toda falha HTTP como quebra de prazo. O sistema SHALL registrar os marcos mínimos definidos (recebimento na borda, aceite durável, modo DIRECT/QUEUED, envio externo, aceite externo, callback/status, conteúdo pronto, conclusão durável, primeira tentativa e ack de webhook) usando relógios sincronizados, e o Atlas SHALL oferecer consulta paginada com os filtros e detalhamentos especificados, respeitando a visibilidade autorizada de cada perfil.

#### Scenario: Taxa de quebra usa população elegível da coorte
- GIVEN pedidos em diferentes estados (cumpridos, violados, abertos no prazo, abertos vencidos, expirados)
- WHEN a taxa de quebra de SLA é calculada
- THEN o sistema SHALL usar a população elegível da coorte contratual, sem presumir sucesso para pedidos ainda no prazo e sem omitir retornos tardios, rejeições, exclusões ou backlog do relatório

#### Scenario: Evidência de SLA durável e independente de amostragem
- GIVEN uma contestação de SLA por um cliente
- WHEN a evidência é consultada
- THEN o sistema SHALL fornecer evidência em armazenamento durável com idempotência por protocolo/operação/obrigação/versão, sem depender de tracing amostrado, e a exportação SHALL incluir versão, competência e relatório de lacunas/lag

#### Scenario: Alerta preventivo de orçamento não é violação confirmada
- GIVEN o consumo de orçamento de erro atingindo 80% do limite
- WHEN o alerta preventivo dispara
- THEN o sistema SHALL apresentá-lo como aviso configurável distinto de uma violação já confirmada, sem bloquear o encerramento por prazo ou o controle adaptativo, que não esperam a atualização da visão operacional (meta de até 60 s)

### Requirement: OPE-11 — Preservação de informação e limites da garantia
O sistema SHALL definir a garantia de preservação sobre entradas confirmadas como aceitas (requisição pública após commit, callback externo após recibo durável, mensagem consumida após commit de efeito/intenção), respondendo explicitamente a tráfego recusado antes do aceite em vez de tratá-lo como obrigação invisível. O sistema SHALL garantir RPO zero dos fatos já confirmados para crash de processo, duplicação, indisponibilidade transitória de broker e falha de uma zona dentro da topologia homologada, com retomada e reconciliação sem perda silenciosa, evidenciando persistência replicada compatível.

#### Scenario: Reconciliação identifica lacuna como incidente
- GIVEN contagens de aceites, protocolos, intenções, operações, finais, entregas e uso financeiro
- WHEN a reconciliação periódica é executada
- THEN o sistema SHALL tratar qualquer lacuna encontrada como incidente, não apenas como métrica, e um retorno rejeitado por SLA SHALL manter evidência permitida sem expurgo motivado somente pela rejeição

#### Scenario: RPO zero regional exige persistência síncrona confirmada
- GIVEN uma oferta que reivindica RPO zero regional
- WHEN a arquitetura de referência regional v4 é avaliada
- THEN o sistema SHALL bloquear essa venda sob P-08 até que exista persistência síncrona/confirmada em regiões independentes, arbitragem contra split-brain e validação de objetos, pois backup assíncrono não satisfaz esse requisito isoladamente

#### Scenario: Operação ambígua preserva UNKNOWN sem inventar resultado
- GIVEN uma operação cujo desfecho externo não pôde ser confirmado
- WHEN o sistema precisa reportar o estado ao cliente
- THEN o sistema SHALL preservar o estado UNKNOWN/reconciliação e informar o encerramento contratual quando cabível, sem inventar resultado ou apagar custo

### Requirement: OPE-12 — Escala automática de ponta a ponta
O sistema SHALL acionar capacidade adicional por automação (HPA para APIs, KEDA para workers, Karpenter para nós, política Atlas/Plataforma com Crossplane para células/core/finance) dentro dos envelopes aprovados, sem exigir chamado manual para criar instância, fila ou banco, mantendo a arquitetura lógica estável enquanto o número e o tamanho dos recursos aumentam. O sistema SHALL prever o tempo de expansão de modo que demand_headroom cubra o crescimento esperado durante o p99 medido de provisionamento/qualificação, mais a margem N−1 e de rajada, e SHALL manter uma autoridade de orçamento do Cometa por domínio de capacidade com concessões limitadas, persistidas, epoch e validade.

#### Scenario: Expansão antecipa saturação de banco/célula
- GIVEN reservas de capacidade excedendo 60% do envelope N−1 ou previsão atingindo 80% dentro do horizonte de provisionamento
- WHEN a política Atlas/Plataforma avalia o gatilho de expansão
- THEN o sistema SHALL iniciar a expansão de célula/banco antes que o limite seja atingido, reconhecendo que provisionar uma célula pode levar dezenas de minutos e que não se deve esperar o banco atingir seu limite para começar

#### Scenario: Controller de expansão indisponível não retira capacidade ativa
- GIVEN o controller de expansão automática indisponível
- WHEN a demanda continua crescendo
- THEN o sistema SHALL manter a capacidade ativa e os mínimos configurados, acionando alarme de horizonte de folga em vez de revogar recursos de produção

#### Scenario: Falha da autoridade de orçamento admite apenas concessões válidas
- GIVEN uma conta compartilhada entre células com falha na autoridade de orçamento do Cometa
- WHEN novas submissões daquele domínio chegam
- THEN o sistema SHALL admitir somente concessões ainda comprovadamente válidas, e novas submissões SHALL aguardar ou expirar segundo o contrato até a autoridade ser restaurada

### Requirement: OPE-13 — Matriz de dependências e fallback seguro
O sistema SHALL operar segundo os estados NORMAL, DEGRADADO_COM_CONTINUIDADE, ADMISSAO_RESTRITA e RECUPERANDO por capacidade/célula/vínculo, sem usar uma flag global para derrubar todos os serviços por falha isolada, e cada fallback SHALL ter pré-condição, prazo e limite de consistência definidos — nunca "continuar a qualquer custo".

#### Scenario: Indisponibilidade do Redis usa bypass sem informação crítica exclusiva
- GIVEN o Redis indisponível
- WHEN requisições SYNC, ASYNC, GET e de controle chegam
- THEN o sistema SHALL continuar atendendo essas requisições pelo caminho sem L2, com bypass protegido por timeout/breaker, reaquecendo gradualmente e comprovando capacidade de origem com cache frio na recuperação, sem que nenhuma informação crítica exclusiva dependa do Redis

#### Scenario: Atlas indisponível mantém rotas conhecidas e bloqueia novo onboarding
- GIVEN o Atlas ou banco de controle indisponível
- WHEN pedidos usam contratos/configurações já válidos e rotas conhecidas
- THEN o sistema SHALL continuar atendendo esses pedidos, mas SHALL colocar novas publicações/onboarding em espera, sem ignorar autorização revogada ou desatualizada, retomando de forma incremental com projeções, versão e validade de segurança verificadas

#### Scenario: PostgreSQL core writer indisponível preserva finais confirmados sem novos efeitos sem posse
- GIVEN o PostgreSQL core writer indisponível
- WHEN um final já confirmado precisa ser servido a partir de cópia de leitura, ou uma execução já enviada pode ainda produzir efeito externo
- THEN o sistema SHALL servir o final confirmado e permitir que execuções em voo produzam efeito, mas SHALL NOT iniciar novos efeitos sem intenção/posse duráveis, SHALL NOT dar ack de callback sem recibo, e SHALL NOT declarar final perdido em memória, retornando 503/erro de transporte quando a comprovação faltar, com recuperação por chave e reconciliação de UNKNOWN

#### Scenario: Banco financeiro indisponível restringe reservas estritas
- GIVEN o banco financeiro/Libra indisponível
- WHEN pedidos pós-pagos com fatos duráveis e GETs sem reserva estrita chegam
- THEN o sistema SHALL continuar atendendo esses pedidos, mas novas reservas estritas SHALL aguardar ou falhar no prazo sem saldo inventado, drenando fatos, deduplicando e conciliando antes do fechamento

#### Scenario: SNS/SQS indisponível recusa ASYNC sem slot durável
- GIVEN SNS/SQS indisponível
- WHEN uma requisição SYNC chega enquanto core/outbox têm capacidade, ou uma requisição ASYNC chega
- THEN o sistema SHALL continuar atendendo SYNC direto e GET/registro de fatos/recibos duráveis, mas SHALL aceitar ASYNC somente com slot durável e política qualificada que comporte o atraso, retornando 503 antes de efeito caso contrário, com outbox conservando a intenção e drain com limites

#### Scenario: S3 indisponível não finaliza sucesso com referência não confirmada
- GIVEN o S3/armazenamento de objetos indisponível
- WHEN uma operação com objeto obrigatório precisa ser concluída
- THEN o sistema SHALL atender operações de payload pequeno sem objeto obrigatório e resultados locais já íntegros, mas SHALL fazer upload/captura obrigatória aguardar ou falhar no prazo, sem finalizar sucesso com referência indisponível não confirmada, recuperando por object_id/checksum sem buscar no provedor por GET

#### Scenario: Cofre/KMS indisponível bloqueia apenas o binding afetado
- GIVEN o Cofre/KMS com instância fria ou chave inválida
- WHEN vínculos com material já obtido, utilizável e autorizado dentro da validade continuam operando
- THEN o sistema SHALL manter esses vínculos operando, mas SHALL bloquear apenas o binding afetado sem usar outra conta, aplicando pré-aquecimento, refresh antecipado e sobreposição de validade, pois ciphertext sem chave disponível não é fallback funcional

#### Scenario: IdP indisponível não emite nova identidade por conta própria
- GIVEN o IdP/distribuição de autorização indisponível
- WHEN tokens/identidades já válidos e prova de autorização dentro da política existem
- THEN o sistema SHALL continuar aceitando esses tokens, mas SHALL NOT emitir nova identidade por conta própria, negando escopo sem validade comprovada e mantendo cache de chaves públicas com rotação/validade, sem adiar revogação indefinidamente

#### Scenario: Controlador de quota externa em partição nega novos tokens
- GIVEN o controlador de quota externa em partição de rede
- WHEN concessões globais previamente autorizadas ainda são válidas
- THEN o sistema SHALL continuar honrando essas concessões, mas nenhum pod SHALL inventar novos tokens/slots ou ampliar teto durante a partição, aplicando eleição/fencing com checkpoint e reinício conservador na recuperação

#### Scenario: Provedor específico indisponível isola apenas o vínculo afetado
- GIVEN um provedor específico indisponível
- WHEN outros vínculos/células permanecem saudáveis
- THEN o sistema SHALL continuar atendendo outros vínculos e GET de qualquer final local autorizado, aplicando retry com TTL, circuit breaker e redução de pressão apenas ao provedor afetado, sem tratar UNKNOWN como failover seguro

#### Scenario: Webhook do cliente indisponível preserva obrigação sem reexecutar produto
- GIVEN o endpoint de webhook do cliente indisponível
- WHEN o Pulsar tenta entregar o resultado
- THEN o sistema SHALL continuar o processamento e a consulta local e outras entregas, conservando a obrigação com retry até EXHAUSTED, sem reexecutar o produto, e reenviando a mesma representação/event_id após a recuperação

#### Scenario: Observabilidade indisponível não bloqueia auditoria de negócio
- GIVEN Prometheus/Loki/Grafana/Tempo indisponíveis
- WHEN o processamento de negócio continua
- THEN o sistema SHALL continuar processando com limites locais, usando buffers de telemetria limitados e contabilizando descarte técnico sem bloquear por log, pois auditoria/financeiro SHALL NOT depender de logs amostrados

#### Scenario: Kubernetes API/Crossplane/Karpenter indisponíveis mantêm workloads saudáveis
- GIVEN a API do Kubernetes, Crossplane ou Karpenter indisponíveis
- WHEN workloads e células já saudáveis continuam operando
- THEN o sistema SHALL manter esses workloads em operação, aceitando que mudanças, nova expansão e reposição fiquem limitadas, pois o runtime SHALL NOT consultar a API do Kubernetes por pedido

### Requirement: OPE-14 — Manutenção sem indisponibilidade evitável
O sistema SHALL manter, para toda dependência, plano de atualização, compatibilidade de versão, teste em ppd, sequência de substituição, condição de aborto e rollback/roll-forward, incluindo manutenção no cálculo de disponibilidade sem exclusão automática. O sistema SHALL verificar, antes da janela de manutenção, a saúde das cópias, capacidade N−1, backups/restauração, backlog, validade de credenciais e prazo dos protocolos em voo, sem atualizar simultaneamente todas as zonas/cópias ou componentes correlacionados.

#### Scenario: Canary/rolling de APIs e workers preserva intenção
- GIVEN uma atualização de APIs e workers
- WHEN o procedimento canary/rolling com surge e reserva é executado
- THEN o sistema SHALL retirar prontidão, drenar conexões/tarefas e encerrar com prazo maior que a drenagem homologada, garantindo que chamadas SYNC existentes sejam concluídas ou explicitamente recuperáveis, sem perda de intenção/efeito

#### Scenario: Manutenção de PostgreSQL demonstra RTO real, não apenas "multi-AZ"
- GIVEN uma manutenção de PostgreSQL com failover suportado
- WHEN o tempo real de indisponibilidade de escrita é medido
- THEN o sistema SHALL demonstrar o RTO exigido no perfil qualificado (incluindo refresh de endpoint, conexões, transações e cache frio), sem qualificar um perfil SYNC apenas pela declaração "RDS multi-AZ", pois o failover de instância pode levar tipicamente 60–120 segundos ou mais

#### Scenario: Manutenção que excede orçamento síncrono é falha de disponibilidade
- GIVEN uma janela de manutenção que ultrapassa o orçamento síncrono definido no SLO
- WHEN o evento é registrado
- THEN o sistema SHALL classificá-lo como falha de disponibilidade mensurada, tratando erro explícito com protocolo como recuperação responsável e não como "zero downtime" comprovado

### Requirement: OPE-15 — Criticidade, tolerância a falhas e continuidade operacional
O sistema SHALL exigir perfil de criticidade formal, envolvendo o responsável do sistema consumidor, antes de contratação/produção para qualquer serviço que possa participar de operação com impacto em vida/segurança, cobrindo função e consequência, prazo e continuidade, falhas cobertas, custódia, capacidade, operação e evidência conforme especificado. O sistema SHALL bloquear a ativação de oferta crítica cuja interrupção máxima, RPO, dependência externa ou comportamento de contingência não tenham evidência compatível, sem herdar os RTOs gerais de cinco minutos/quatro horas como adequados para o perfil crítico.

#### Scenario: 99,99% mensal não prova adequação a impacto crítico
- GIVEN a meta proposta de disponibilidade de 99,99% mensal (equivalente a cerca de 4 min 19 s de indisponibilidade em 30 dias)
- WHEN um perfil crítico avalia essa meta
- THEN o sistema SHALL reconhecer que o percentual mensal não estabelece quanto dura uma falha individual, exigindo metas mais fortes e RTO muito menor quando P-10 assim determinar, e SHALL medir disponibilidade de SYNC pelo final válido dentro do orçamento na conexão, não por um 202 rápido

#### Scenario: Perda de região com RPO zero exige autoridade transacional qualificada
- GIVEN um perfil crítico que exige sobreviver à perda de região com RPO zero e RTO muito curto
- WHEN a arquitetura regional de referência é avaliada
- THEN o sistema SHALL declarar essa referência insuficiente, exigindo autoridade transacional tolerante à perda regional, confirmação de cópias/objetos em regiões independentes, capacidade ativa/aquecida, quorum/fencing e roteamento qualificado, sem aceitar duplicação de EKS ou replicação assíncrona de banco como substituto

#### Scenario: Durante partição regional apenas um lado aceita efeitos
- GIVEN uma partição de rede entre regiões durante operação crítica
- WHEN ambos os lados poderiam potencialmente aceitar escritas
- THEN o sistema SHALL permitir que apenas o lado com autoridade válida aceite efeitos, pois servir ambos com escritas independentes ameaça idempotência e saldo

#### Scenario: Contrato de contingência do consumidor define uso de resultado antigo
- GIVEN uma indisponibilidade que impede confirmar novo processamento
- WHEN o contrato de contingência do consumidor é aplicado
- THEN o sistema SHALL restringir novas operações e informar a situação, usando apenas o último resultado confirmado dentro da idade máxima definida, sem fabricar dado novo ou sucesso para encobrir a falha

## Notas de origem

Requisitos extraídos e adaptados de `docs/07_DESEMPENHO_E_OPERACAO.md`, capítulo 07 · Desempenho, alta disponibilidade e operação, da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação).
