# Arquitetura e Comunicação — Delta de Especificação

## ADDED Requirements

### Requirement: ARQ-01 — Fronteiras
O sistema SHALL manter cinco aplicações de negócio — Atlas (controle), Órbita (execução), Cometa (integração), Pulsar (entrega) e Libra (financeiro) — e um gateway, Portal, cada um com autoridade exclusiva sobre as responsabilidades listadas para si e SHALL NOT assumir responsabilidades explicitamente atribuídas a outro componente. Os nomes são identificadores de trabalho e SHALL sempre aparecer junto da função. Órbita SHALL conter a composição e o roteamento; Cometa SHALL conter o scheduler durável de polling e os adaptadores. O sistema NÃO SHALL criar, por padrão, um microserviço separado para cada adaptador ou temporizador; isolamento em deployments próprios MAY ocorrer por risco, capacidade ou dependência nativa, preservando a autoridade de dados do domínio.

#### Scenario: Portal não orquestra nem precifica
- GIVEN um pedido chega ao Portal e requer decisão de orquestração ou de preço
- WHEN o Portal processa TLS, autenticação de borda, quotas técnicas, rotas e limites de payload
- THEN o Portal SHALL delegar a orquestração a Órbita e a precificação a Atlas, SHALL NOT executar essas funções nem atuar como banco de protocolo

#### Scenario: Órbita delega execução externa a Cometa sem manter segredos
- GIVEN um passo do grafo de Órbita exige uma chamada a um provedor externo
- WHEN Órbita processa admissão, protocolo, idempotência e roteamento desse passo
- THEN Órbita SHALL delegar a operação/tentativa externa a Cometa, SHALL NOT manter segredos de provedores nem realizar entrega de webhook ou ledger

#### Scenario: Libra não infere consumo apenas de logs nem fatura todo evento técnico
- GIVEN um evento técnico é gerado durante a execução de um protocolo
- WHEN Libra realiza medição econômica e apuração
- THEN Libra SHALL basear a medição em fatos e reservas estritas definidos pelo contrato, SHALL NOT inferir consumo somente a partir de logs nem faturar automaticamente todo evento técnico

#### Scenario: Isolamento de adaptador ou temporizador apenas por critério explícito
- GIVEN um adaptador de Cometa ou um temporizador candidato a deployment próprio
- WHEN a arquitetura avalia se ele deve ser isolado em deployment separado
- THEN o sistema SHALL manter o padrão de não separar microserviço por adaptador/temporizador, permitindo isolamento apenas quando justificado por risco, capacidade ou dependência nativa, sem perder a autoridade de dados do domínio

### Requirement: ARQ-02 — Topologia de referência
O sistema SHALL seguir a topologia de referência em células: Portal roteia para Órbita; Órbita usa HTTPS direto para SYNC até Cometa e intenção durável via SNS/SQS da célula para ASYNC; a fila também SHALL alimentar Pulsar (webhook) e Libra (medição e ledger); Órbita e Cometa SHALL gravar cada um somente seu próprio estado no PostgreSQL core compartilhado da célula, sem acesso cruzado a schemas; Libra SHALL gravar em PostgreSQL financeiro distinto. Atlas SHALL distribuir projeções versionadas ao Portal e às demais aplicações, e o tráfego já estabelecido NÃO SHALL consultar Atlas/Crossplane a cada pedido.

#### Scenario: SYNC persiste o fato externo via resposta HTTPS sem depender da fila
- GIVEN Órbita despacha uma operação SYNC para Cometa via HTTPS direto
- WHEN Cometa executa e persiste o fato externo
- THEN Órbita SHALL receber o fato já persistido na própria resposta HTTPS, SHALL consolidar/persistir o estado final e responder à conexão do cliente sem esperar que esse fato atravesse a fila

#### Scenario: ASYNC usa fila e mantém os mesmos invariantes com deduplicação de eventos posteriores
- GIVEN Órbita despacha uma operação ASYNC para Cometa
- WHEN a intenção durável é publicada via SNS/SQS da célula e eventos posteriores repetem a observação do mesmo fato
- THEN o sistema SHALL processar a intenção com os mesmos invariantes do modo SYNC e SHALL deduplicar os eventos repetidos, sem regressão de estado

#### Scenario: Cada domínio grava somente seu próprio estado no core compartilhado
- GIVEN Órbita e Cometa compartilham o mesmo PostgreSQL core da célula
- WHEN cada aplicação persiste dados de sua responsabilidade
- THEN Órbita SHALL gravar apenas seu estado de execução e Cometa apenas seu estado de integração, SHALL NOT haver acesso cruzado a schemas entre os dois domínios

#### Scenario: Tráfego estabelecido não consulta Atlas/Crossplane a cada pedido
- GIVEN Atlas distribuiu uma projeção versionada de catálogo/configuração a uma célula
- WHEN pedidos subsequentes chegam a essa célula dentro da mesma versão
- THEN o sistema SHALL atender esses pedidos com a projeção local já distribuída, SHALL NOT consultar Atlas ou Crossplane a cada pedido individual

### Requirement: ARQ-03 — Escolha tecnológica
A arquitetura de referência v4 SHALL adotar, por camada: Go para aplicações; PostgreSQL para banco transacional (banco de controle e bancos core/finance por célula); SNS + SQS Standard para o barramento, com fila por consumidor e domínio de isolamento, sem presumir ordem; S3 com chaves/versões imutáveis para arquivos; Secrets Manager + KMS para segredos; Kubernetes/EKS multi-AZ para execução remota; Kong para gateway, sem deslocar regras de negócio para plugins; OpenTelemetry/Prometheus/Grafana/Loki/Alloy/Tempo para observabilidade; cache L1 local obrigatório com Redis L2 opcional e bypass; HPA/KEDA, Karpenter e Crossplane para expansão de capacidade no plano de plataforma. AWS permanece referência da v1, não infraestrutura já contratada; conta, região e orçamento SHALL ser aprovados em P-01.

#### Scenario: PostgreSQL é a escolha para o banco transacional por célula
- GIVEN a necessidade de um banco transacional para controle, core e finance de uma célula
- WHEN a arquitetura seleciona a tecnologia dessa camada
- THEN o sistema SHALL usar PostgreSQL pelo critério de ACID, unicidade, integridade, RLS e independência financeira

#### Scenario: SNS + SQS Standard é a escolha para o barramento sem presumir ordem
- GIVEN a necessidade de distribuir fatos a consumidores independentes e entregar comandos ao menos uma vez
- WHEN a arquitetura seleciona a tecnologia de barramento
- THEN o sistema SHALL usar SNS + SQS Standard, com fila por consumidor e domínio de isolamento, SHALL NOT presumir ordem global das mensagens

#### Scenario: Cache funciona e cumpre o SLO qualificado sem Redis
- GIVEN a camada de cache é composta por L1 local e Redis L2 opcional com bypass
- WHEN o Redis está indisponível ou não foi provisionado
- THEN o sistema SHALL continuar funcionando e cumprindo o SLO qualificado apenas com o cache L1 local

#### Scenario: Operar sem serviços AWS exige nova decisão de arquitetura
- GIVEN uma proposta de operar sem os serviços AWS de referência (broker, armazenamento)
- WHEN a arquitetura avalia uma alternativa não aprovada
- THEN o sistema SHALL exigir uma nova decisão que selecione broker e armazenamento equivalentes e requalifique a semântica; compatibilidade de API SHALL NOT ser tratada como prova de equivalência operacional, e a arquitetura NÃO SHALL adotar a alternativa sem essa nova decisão

### Requirement: ARQ-04 — Linguagem e tecnologia por aplicação
As cinco APIs/workers do domínio (Atlas API, Órbita, Cometa, Pulsar, Libra) SHALL ser implementadas em Go, na mesma toolchain do núcleo; a interface administrativa de Atlas SHALL ser implementada em TypeScript + React e NÃO SHALL processar pedidos de clientes; o Portal SHALL usar Kong Gateway como produto pronto de API management, SHALL NOT ser reimplementado em Go. Bancos, brokers e agentes de observabilidade são produtos operados e NÃO SHALL ser reescritos na linguagem do hub. O sistema NÃO SHALL declarar Go universalmente mais rápido ou mais barato que Java/Rust sem comparação de carga, hardware, limites e equipe equivalentes. A regra de concorrência é ausência de espera bloqueante evitável no caminho crítico, não promessa de zero bloqueios.

#### Scenario: Atlas API usa Go pela uniformidade de toolchain, não pela velocidade de CRUD
- GIVEN a necessidade de implementar validação, publicação e projeções em Atlas API
- WHEN a linguagem de implementação é escolhida
- THEN o sistema SHALL usar Go pela mesma toolchain do núcleo e redução de manutenção, SHALL NOT justificar essa escolha apenas por velocidade de operações CRUD

#### Scenario: Espera de rede suportada libera a thread, mas bloqueio evitável no caminho crítico é proibido
- GIVEN um passo em Go realiza uma espera de rede suportada durante uma chamada externa
- WHEN a goroutine aguarda a resposta
- THEN o sistema SHALL permitir que a goroutine estacione e libere a thread para outro trabalho, e SHALL NOT introduzir espera bloqueante evitável no caminho crítico, ainda que CPU, código nativo ou certas chamadas de sistema possam ocupar threads

#### Scenario: Interface de Atlas em TypeScript/React não processa pedidos de clientes
- GIVEN a interface administrativa web de Atlas é construída em TypeScript + React
- WHEN clientes externos enviam pedidos ao hub
- THEN esses pedidos NÃO SHALL ser processados por essa interface, e a validação de negócio SHALL permanecer no backend, independente da tipagem do frontend

#### Scenario: Comparação de desempenho entre Go e outras linguagens exige equivalência de condições
- GIVEN uma afirmação de que Go é mais rápido ou mais barato que Java/Rust para um componente do hub
- WHEN essa afirmação é usada para justificar uma decisão de arquitetura
- THEN o sistema NÃO SHALL aceitar essa comparação sem carga, hardware, limites e equipe equivalentes entre as alternativas

### Requirement: ARQ-05 — Razões das dependências e revisão de alternativas
Cada dependência de plataforma (PostgreSQL gerenciado, SNS + SQS Standard, S3, Secrets Manager + KMS, EKS/Kubernetes, Kong, OpenTelemetry + Tempo, Prometheus, Loki + Alloy, Grafana, Metrics Server + HPA/KEDA, cache local) SHALL ter razão de escolha, limite conhecido e alternativa de revisão registrados, e SHALL ter versão suportada fixada, plano de atualização, licença e proprietário operacional definidos antes de produção. O sistema NÃO SHALL ser obrigado a acrescentar mesh ou Envoy para o controle adaptativo do Cometa; OPE-07 especifica essa política, e referências de outros produtos fundamentam o mecanismo sem implicar adoção automática.

#### Scenario: PostgreSQL compartilhado entre células é separado por célula
- GIVEN um servidor PostgreSQL potencialmente compartilhado entre múltiplas células
- WHEN a capacidade e o isolamento de I/O são avaliados
- THEN o sistema SHALL separar o banco por célula, pois compartilhar servidor NÃO SHALL ser tratado como isolamento de I/O suficiente

#### Scenario: Dependência de produção exige versão fixada, plano de atualização, licença e proprietário
- GIVEN uma dependência de plataforma candidata a uso em produção
- WHEN essa dependência é qualificada para o ambiente produtivo
- THEN o sistema SHALL exigir versão suportada fixada, plano de atualização, licença, proprietário operacional e alternativa registrada antes da liberação

#### Scenario: Mesh ou Envoy não são adotados automaticamente para o controle adaptativo do Cometa
- GIVEN referências de mercado sugerem uso de service mesh ou Envoy para controle adaptativo
- WHEN a arquitetura avalia o mecanismo de controle do Cometa
- THEN o sistema NÃO SHALL adotar mesh/Envoy automaticamente por essa referência; a política aplicável é a especificada em OPE-07

#### Scenario: Traces do OpenTelemetry/Tempo não servem como fonte contábil ou única evidência de SLA
- GIVEN traces distribuídos coletados via OpenTelemetry com backend Tempo
- WHEN uma disputa de SLA ou apuração financeira precisa de evidência
- THEN o sistema NÃO SHALL usar exclusivamente traces amostrados como fonte contábil ou única evidência de cumprimento de SLA

### Requirement: ARQ-06 — Tecnologias para expansão e continuidade
O sistema SHALL usar HPA para réplicas das APIs, KEDA como único proprietário do HPA de cada Deployment de workers/timers (métricas de backlog/idade, com mínimo aquecido), Karpenter com NodePools para nós EKS, e Crossplane em plano de gestão separado para reconciliar recursos de novas células. Crossplane NÃO SHALL decidir sozinho quando expandir nem migrar dados de negócio; Atlas SHALL registrar demanda/placement, e a política de capacidade SHALL disparar os recursos desejados. O sistema SHALL proteger bancos, chaves e buckets contra exclusão automática; redução de pods/nós NÃO SHALL apagar dados. Depois do repasse formal de propriedade, um recurso NÃO SHALL ter dois reconciliadores concorrentes.

#### Scenario: KEDA é o único proprietário do HPA de um Deployment de worker
- GIVEN um Deployment de worker escalado por backlog/idade de obrigações
- WHEN HPA e KEDA são configurados para esse Deployment
- THEN o sistema SHALL usar KEDA como único proprietário do HPA daquele Deployment, mantendo um mínimo aquecido quando a métrica de backlog estiver ausente

#### Scenario: Crossplane reconcilia recursos mas não decide quando expandir
- GIVEN uma nova célula precisa de bancos, filas, IAM, rede e demais dependências homologadas
- WHEN Crossplane reconcilia os recursos desejados dessa célula
- THEN o sistema SHALL usar Crossplane apenas para reconciliar o estado desejado; a decisão de quando expandir SHALL vir do registro de demanda/placement em Atlas e da política de capacidade, não do próprio Crossplane

#### Scenario: Redução de pods ou nós não apaga dados protegidos
- GIVEN o Karpenter ou o HPA/KEDA reduzem pods ou nós por baixa demanda
- WHEN a capacidade é reconciliada para baixo
- THEN o sistema SHALL proteger bancos, chaves e buckets contra exclusão automática, e a redução de pods/nós NÃO SHALL apagar dados de negócio

#### Scenario: Perda do plano de gestão impede novas expansões mas não desliga células prontas
- GIVEN o plano de gestão do Crossplane fica indisponível
- WHEN novas expansões ou mudanças de capacidade são solicitadas
- THEN o sistema SHALL bloquear apenas novas expansões e mudanças, SHALL NOT desligar células já provisionadas e em operação

### Requirement: COM-01 — Comunicação interna definida
O sistema SHALL usar, para cada rota interna definida, o transporte especificado: HTTPS REST/JSON canônico de Portal para Órbita/Atlas; HTTPS interno idempotente com mTLS e deadline propagado de Órbita para Cometa em SYNC; comando durável em SQS dedicada da célula/classe de Órbita para Cometa em ASYNC; resposta HTTPS direta de Cometa para Órbita em SYNC, com consulta interna por operation_id para recuperação; fatos via SNS com filas SQS independentes de Cometa para Órbita e Libra; fato final via SNS/SQS de Órbita para Pulsar e Libra; evento de publicação e consulta HTTPS da versão imutável de Atlas para as demais aplicações; HTTPS autenticado de Pulsar para Órbita, sem acesso direto ao banco da Órbita; HTTPS idempotente de Órbita para Libra apenas para reserva de saldo/franquia; conexões autenticadas e criptografadas com identidade por componente entre aplicações e bancos/objetos. O sistema NÃO SHALL adotar gRPC como padrão inicial nem request/reply sobre broker para o caminho SYNC; retorno HTTP interno perdido NÃO SHALL autorizar nova operação.

#### Scenario: SYNC entre Órbita e Cometa dispensa broker no percurso obrigatório
- GIVEN Órbita despacha uma execução direta para Cometa em modo SYNC
- WHEN Cometa responde via HTTPS interno idempotente com mTLS e deadline propagado
- THEN o sistema SHALL tratar essa resposta como o retorno válido do caminho SYNC, sem exigir que um broker participe do percurso obrigatório

#### Scenario: Retorno HTTP interno perdido não autoriza nova operação
- GIVEN a resposta HTTPS interna entre Órbita e Cometa se perde por timeout de rede
- WHEN Órbita avalia como recuperar o estado da operação
- THEN o sistema NÃO SHALL autorizar uma nova operação a partir dessa perda; a recuperação SHALL usar a identidade durável (operation_id) via consulta interna

#### Scenario: gRPC e request/reply sobre broker não são adotados como padrão do caminho SYNC
- GIVEN uma proposta de usar gRPC ou request/reply sobre broker para o caminho SYNC
- WHEN essa proposta é avaliada contra a arquitetura de referência
- THEN o sistema NÃO SHALL adotar essas opções como padrão inicial para SYNC, mantendo a mesma máquina de estados, persistência, deadline, controle de pressão e idempotência do transporte direto

#### Scenario: Pulsar consulta Órbita por HTTPS autenticado sem acesso direto ao banco
- GIVEN Pulsar precisa obter o resultado de um protocolo por versão para reentrega de webhook
- WHEN essa consulta é necessária
- THEN Pulsar SHALL usar HTTPS autenticado contra Órbita, SHALL NOT acessar diretamente o banco de dados da Órbita

### Requirement: COM-02 — Integração de clientes e provedores
O sistema SHALL disponibilizar REST/JSON por HTTPS como canal padrão v4 para cliente-hub e hub-provedor, mediante OpenAPI e schemas homologados; webhook HTTPS com representação contratada para entrega assíncrona e recepção de callback; polling HTTPS mediante endpoint e mapa de estados homologados; SOAP/XML como perfil legado homologado por cliente; upload/download de objetos via URL temporária autorizada. SFTP com chave, gRPC/Protobuf e integração com broker externo do parceiro SHALL permanecer como evolução futura mediante demanda, contrato homologado ou ADR explícito, sem aceite automático na v4. O catálogo declarativo SHALL permitir configurar apenas capacidades já implementadas; NÃO SHALL criar suporte universal a protocolo, criptografia ou transformação desconhecidos. Conexão direta do cliente ao banco e consultas arbitrárias ao banco do provedor ficam fora do escopo.

#### Scenario: REST/JSON por HTTPS é disponibilizado mediante OpenAPI e schemas homologados
- GIVEN um novo cliente ou provedor a integrar via canal REST/JSON
- WHEN o OpenAPI e os schemas dessa integração são homologados
- THEN o sistema SHALL disponibilizar esse canal como padrão v4 para comunicação cliente-hub e hub-provedor

#### Scenario: SFTP e gRPC/Protobuf não são aceitos automaticamente na v4
- GIVEN uma solicitação para usar SFTP com chave ou gRPC/Protobuf como canal de integração
- WHEN essa solicitação é avaliada
- THEN o sistema SHALL tratar esse canal como evolução futura mediante demanda e contrato homologado (layout, manifesto, checksum, ack para SFTP; schema e prazos para gRPC), SHALL NOT aceitá-lo automaticamente na v4

#### Scenario: Conexão direta do cliente ao banco é recusada independentemente do catálogo
- GIVEN um cliente ou parceiro solicita conexão direta ao banco do hub ou consulta arbitrária ao banco de um provedor
- WHEN essa solicitação é avaliada contra o catálogo declarativo de canais
- THEN o sistema SHALL recusar essa conexão, pois está fora do escopo independentemente da configuração do catálogo

#### Scenario: Integração via broker externo do parceiro exige ADR explícito
- GIVEN um parceiro propõe integração adicional via seu próprio broker de mensagens
- WHEN essa integração é avaliada
- THEN o sistema SHALL exigir um ADR explícito definindo ACL, entrega, correlação e retenção antes de adotar esse canal, que permanece fora do canal público padrão

### Requirement: COM-03 — Mensagens e idempotência
Toda mensagem SHALL carregar o envelope obrigatório: event_id, tipo, versão do schema, produtor, tenant_id, protocol_id, data de ocorrência UTC, data de registro, causation_id, versão do agregado e versões de configuração, incluindo step_id, operation_id e attempt_id quando pertinentes. A publicação SHALL usar outbox na mesma transação do estado local; o consumo SHALL gravar inbox junto do efeito local ou da intenção durável de trabalho, com ack ao broker somente após o commit. O consumidor SHALL usar uma chave semântica para reconhecer fatos econômicos equivalentes publicados com event_ids distintos. O sistema NÃO SHALL presumir ordem global nem execução externa exatamente uma vez; eventos antigos NÃO SHALL regredir estado, e o broker NÃO SHALL ser tratado como banco de resultado ou arquivo financeiro.

#### Scenario: Publicação usa outbox na mesma transação do estado local
- GIVEN um evento a ser publicado após uma mudança de estado local
- WHEN a transação que altera o estado é commitada
- THEN o sistema SHALL gravar o evento no outbox dentro dessa mesma transação, contendo o envelope obrigatório completo

#### Scenario: Ack ao broker ocorre somente após o commit do inbox
- GIVEN uma mensagem é consumida do broker
- WHEN o efeito local ou a intenção durável de trabalho correspondente é processado
- THEN o sistema SHALL gravar o inbox junto desse efeito e SHALL enviar o ack ao broker somente depois do commit dessa gravação

#### Scenario: Chave semântica reconhece fatos econômicos equivalentes com event_ids distintos
- GIVEN o produtor republica o mesmo fato com um novo event_id
- WHEN o consumidor processa essa mensagem
- THEN o sistema SHALL usar uma chave semântica para reconhecer a equivalência com o fato já processado, evitando duplicar o efeito econômico

#### Scenario: Evento antigo não regride estado e gera pendência de reconciliação
- GIVEN um evento antigo chega para um agregado que já avançou de versão
- WHEN esse evento é processado
- THEN o sistema NÃO SHALL regredir o estado do agregado, e a lacuna relevante SHALL ficar pendente para reconciliação

### Requirement: COM-04 — Evolução de contratos
APIs SHALL ser especificadas em OpenAPI; comandos/eventos SHALL ser especificados em AsyncAPI com schemas versionados. Mudanças aditivas que não alterem significado SHALL manter compatibilidade; remover campo, alterar unidade, identidade, enum de interpretação incompatível ou cobrança SHALL exigir nova versão. Um consumidor que desconheça a versão obrigatória SHALL colocar a mensagem em quarentena e alertar, NÃO SHALL descartá-la nem tentar adivinhá-la. O ciclo de publicação SHALL exigir teste produtor/consumidor, janela de coexistência, migração e retirada anunciada. Campos desconhecidos opcionais MAY ser ignorados; campos críticos ausentes SHALL rejeitar a mensagem. Contratos NÃO SHALL conter payload de produção ou segredo em exemplos.

#### Scenario: Mudança aditiva sem alteração de significado mantém compatibilidade
- GIVEN um novo campo opcional é adicionado a um schema publicado sem alterar o significado dos campos existentes
- WHEN essa mudança é publicada
- THEN o sistema SHALL manter a compatibilidade do contrato sem exigir nova versão major

#### Scenario: Remoção de campo ou mudança de unidade exige nova versão
- GIVEN uma mudança que remove um campo, altera uma unidade, identidade ou o enum de interpretação de forma incompatível
- WHEN essa mudança é publicada
- THEN o sistema SHALL exigir uma nova versão do contrato, com janela de coexistência e retirada anunciada da versão anterior

#### Scenario: Consumidor desconhecendo a versão obrigatória coloca a mensagem em quarentena
- GIVEN uma mensagem chega anunciando uma versão de schema que o consumidor não reconhece como obrigatória
- WHEN o consumidor processa essa mensagem
- THEN o sistema SHALL colocar a mensagem em quarentena e alertar, NÃO SHALL descartá-la nem tentar adivinhar seu conteúdo

#### Scenario: Campo crítico ausente rejeita a mensagem
- GIVEN uma mensagem chega sem um campo declarado crítico pelo schema
- WHEN essa mensagem é validada
- THEN o sistema SHALL rejeitar a mensagem, ao passo que campos desconhecidos opcionais MAY ser ignorados sem rejeição

### Requirement: COM-05 — Mesmo contrato em GET e webhook
Para um tenant, protocolo, versão de representação e versão final, o GET de protocolo e o webhook SHALL usar o mesmo schema, media type, campos, tipos, semântica de estados/erros e conteúdo de negócio; a projeção final SHALL ser materializada uma única vez pela Órbita, com hash e versão, e Pulsar NÃO SHALL transformá-la novamente. Somente cabeçalhos de transporte (assinatura, timestamp da tentativa, autenticação, identificador da tentativa) MAY diferir entre os dois canais. O contrato SHALL cobrir pendência, sucesso, parcialidade permitida, falha e EXPIRED; GET pendente SHALL retornar 200 com a variante pendente do mesmo contrato, sem ser, por si só, um evento final a notificar. Referências de arquivo no corpo SHALL ser identificadores/URLs estáveis autenticadas do hub; URLs pré-assinadas de curta duração SHALL ser obtidas em operação própria, NÃO SHALL ser injetadas dinamicamente no corpo final congelado.

#### Scenario: GET e webhook entregam o mesmo corpo final, diferindo apenas em cabeçalhos de transporte
- GIVEN um protocolo concluído cuja projeção final foi materializada uma única vez pela Órbita, com hash e versão
- WHEN o cliente consulta esse protocolo via GET e recebe a mesma informação via webhook
- THEN ambos SHALL apresentar o mesmo schema, media type, campos, tipos e conteúdo de negócio, SHALL diferir somente em cabeçalhos de transporte como assinatura, timestamp e identificador da tentativa

#### Scenario: GET pendente retorna 200 com a variante pendente sem ser evento final
- GIVEN um protocolo ainda em execução, sem resultado final
- WHEN o cliente consulta esse protocolo via GET
- THEN o sistema SHALL retornar 200 com a variante pendente do mesmo contrato, e essa resposta NÃO SHALL ser tratada, por si só, como um evento final a notificar

#### Scenario: Protocolo EXPIRED usa representação de erro prevista e não envia resultado tardio
- GIVEN um protocolo atinge EXPIRED por SLA vencido ou retry esgotado
- WHEN o resultado final é servido via GET ou webhook
- THEN o sistema SHALL usar a representação de erro prevista pelo contrato do cliente para esse estado, NÃO SHALL enviar o resultado tardio caso ele chegue posteriormente

#### Scenario: URL pré-assinada de arquivo não é injetada no corpo final congelado
- GIVEN o corpo final de um protocolo contém uma referência de arquivo
- WHEN o cliente precisa obter acesso temporário a esse arquivo
- THEN o sistema SHALL expor apenas o identificador/URL estável autenticada no corpo final, e a URL pré-assinada de curta duração SHALL ser obtida por operação própria separada, NÃO SHALL ser injetada dinamicamente no corpo final já congelado

### Requirement: COM-06 — Contrato do despacho direto
A chamada interna de execução entre Órbita e Cometa SHALL incluir tenant_id, protocol_id, step_id, command_id, operation_id quando já alocado, dispatch_mode, epoch de posse, versão de entrada/catálogo, referência de vínculo/conta e prazos absolutos, e NÃO SHALL incluir segredo em claro. Cometa SHALL validar escopo/versão, posse e orçamento antes de emitir efeito externo. Somente uma resposta indicando fato final externo durável SHALL permitir que Órbita finalize sucesso no mesmo HTTP público, após suas validações e commit; código 200 do adaptador sem fato conservado NÃO SHALL ser tratado como final válido. Timeout de rede entre Órbita e Cometa NÃO SHALL ser tratado como prova de que o comando deixou de executar; o sistema NÃO SHALL realizar retries de negócio automáticos no gateway, mesh ou biblioteca HTTP.

#### Scenario: Fato final externo durável permite finalizar sucesso no mesmo HTTP público
- GIVEN Cometa retorna uma resposta direta indicando fato final externo durável e conservado
- WHEN Órbita valida essa resposta e realiza o commit
- THEN Órbita SHALL finalizar sucesso na mesma conexão HTTP pública que originou o pedido do cliente

#### Scenario: HTTP 200 sem fato conservado não é tratado como final válido
- GIVEN o adaptador de Cometa retorna código HTTP 200 sem um fato externo conservado
- WHEN Órbita avalia essa resposta
- THEN o sistema NÃO SHALL tratar esse 200 como resultado final válido, exigindo a conservação do fato antes de qualquer finalização de sucesso

#### Scenario: Timeout de rede não autoriza retry automático de negócio fora de Órbita/Cometa
- GIVEN ocorre um timeout de rede na chamada direta entre Órbita e Cometa
- WHEN o sistema decide como reagir a essa falha de comunicação
- THEN o sistema NÃO SHALL tratar o timeout como prova de que o comando deixou de executar, e NÃO SHALL realizar retries de negócio automáticos no gateway, mesh ou biblioteca HTTP; apenas Órbita/Cometa SHALL aplicar a política com a identidade e a segurança do efeito conhecidas

#### Scenario: Intenção ASYNC marcada DIRECT não é tratada como comando QUEUED
- GIVEN uma intenção durável marcada como DIRECT é publicada após o commit
- WHEN o publicador processa o mecanismo de despacho especificado em EXE-15
- THEN o sistema SHALL impedir que essa intenção DIRECT seja tratada como um comando QUEUED

## Notas de origem

Este delta deriva integralmente do capítulo `docs/02_ARQUITETURA_E_COMUNICACAO.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos ARQ-01 a ARQ-06 e COM-01 a COM-06 conforme texto normativo de origem.
