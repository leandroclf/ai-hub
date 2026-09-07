# Persistência e Dados — Delta de Especificação

## ADDED Requirements

### Requirement: DAD-01 — Bancos e propriedade
O sistema SHALL persistir em PostgreSQL com três funções lógicas segregadas — hub_control para Atlas, hub_core para Órbita/Cometa/Pulsar e hub_finance para Libra —, cada célula de execução possuindo seu próprio core e financeiro, com controle independente distribuindo projeções versionadas. Um contrato com saldo estrito SHALL ter uma única autoridade/célula financeira ativa, sem fragmentação do mesmo saldo entre células sem protocolo específico. Cada domínio SHALL ter usuário de aplicação, usuário de migração e permissões próprios, e o sistema NÃO SHALL permitir escritas, joins operacionais ou chaves estrangeiras entre domínios, ainda que compartilhem servidor; referências externas SHALL ser apenas IDs e projeções. Em ppd/prd, core e financeiro SHALL ter instâncias/clusters separados por célula, e hub_control SHALL ter persistência independente para que consultas administrativas e publicação não esgotem o banco de execução. O core SHALL usar primário com redundância multi-AZ, sem escrita ativa-ativa presumida. A autoridade de escrita SHALL atender aceite, transições, saldo estrito e leitura que exija atualidade; réplicas SHALL servir relatórios e histórico com atraso explícito, e a ausência de um registro em cópia atrasada NÃO SHALL ser tratada como prova de inexistência. O pool de conexões SHALL ser limitado por aplicação e dimensionado considerando o máximo de réplicas Kubernetes.

#### Scenario: Isolamento entre domínios impede join, escrita cruzada e chave estrangeira
- GIVEN dois domínios lógicos distintos (por exemplo hub_core de Órbita e hub_finance de Libra), mesmo compartilhando o mesmo servidor físico em ambiente local/dev/hom
- WHEN uma operação tenta realizar join operacional, escrita cruzada ou criar chave estrangeira entre os dois domínios
- THEN o sistema NÃO SHALL permitir a operação, exigindo que a referência entre domínios seja feita apenas por ID e projeção

#### Scenario: Contrato de saldo estrito mantém uma única autoridade financeira ativa
- GIVEN um contrato com saldo estrito vinculado a uma célula financeira ativa
- WHEN uma tentativa de operação busca tratar outra célula como autoridade financeira concorrente para o mesmo saldo, sem protocolo específico de transição
- THEN o sistema NÃO SHALL fragmentar o saldo entre células, mantendo uma única autoridade/célula financeira ativa para aquele contrato

#### Scenario: Ausência em réplica atrasada não é tratada como inexistência do registro
- GIVEN uma réplica de leitura com atraso de replicação explícito em relação à autoridade de escrita
- WHEN uma consulta de relatório ou histórico não encontra um registro recém-gravado nessa réplica
- THEN o sistema NÃO SHALL interpretar essa ausência como prova de inexistência do dado, e uma leitura que exija atualidade SHALL ser direcionada à autoridade de escrita

### Requirement: DAD-02 — Modelo lógico mínimo
O sistema SHALL manter identidade própria e vínculos verificáveis, com regras de integridade observáveis, para cada entidade lógica de domínio: Cliente/aplicação, Serviço/produto versão, Conta/vínculo de provedor, Vínculo de credencial, Placement/capacidade e Contrato/plano versão (Atlas); Protocolo, Passo e Resultado (Órbita); Operação externa, Tentativa, Recibo externo e Agenda de polling (Cometa); Entrega (Pulsar); Uso/fato econômico, Lançamento, Reserva/consumo de franquia e Fatura/conciliação (Libra); Objeto no domínio proprietário; Outbox/inbox/timer, Intenção de despacho e Auditoria em cada domínio pertinente. Essas são entidades lógicas e regras de integridade, não DDL; modelo físico, índices finais e migrations SHALL exigir projeto e testes posteriores. Campos de negócio frequentemente consultados SHALL ser tipados; JSONB SHALL ser reservado a payloads canônicos/versionados e extensões, e o sistema NÃO SHALL esconder tenant, estado, prazo ou chave idempotente somente dentro de um campo JSON.

#### Scenario: Unicidade de idempotência do Protocolo é preservada sob concorrência
- GIVEN um Protocolo identificado por tenant, chave idempotente e versão para controle de concorrência
- WHEN dois pedidos concorrentes chegam com a mesma combinação de tenant/protocolo/chave idempotente
- THEN o sistema SHALL garantir unicidade dessa combinação e SHALL usar a versão do registro para arbitrar a atualização concorrente de estado

#### Scenario: Resultado é imutável por versão e só o ponteiro atual muda com auditoria
- GIVEN um Resultado publicado para uma versão específica de protocolo, com estado final, schema, checksum e datas registrados
- WHEN uma revisão posterior altera qual resultado é considerado o atual
- THEN o sistema SHALL preservar o registro de versão anterior como imutável e SHALL atualizar somente o apontador para o resultado atual, com evento de auditoria associado

#### Scenario: Campos de negócio consultados com frequência não ficam escondidos apenas em JSON
- GIVEN uma entidade lógica cujo tenant, estado, prazo ou chave idempotente são consultados com frequência por índices e regras de negócio
- WHEN o modelo físico é projetado para essa entidade
- THEN esses campos SHALL ser tipados fora do JSONB, que SHALL ficar reservado a payloads canônicos, versionados ou de extensão

### Requirement: DAD-03 — Transações e referências
Na admissão, o sistema SHALL persistir atomicamente protocolo, idempotência, pedido ou referência já validada, versões aplicáveis e intenção de trabalho. Na conclusão, o sistema SHALL preparar e validar a representação do cliente e suas referências antes de persistir atomicamente estado, versão final, decisão de prazo, histórico e evento final; o cliente NÃO SHALL observar um candidato final não confirmado. O sistema NÃO SHALL manter transação aberta enquanto aguarda HTTP, provedor, arquivo ou saldo remoto. Toda atualização concorrente de estado SHALL verificar a versão do registro; lease e token de posse SHALL impedir um worker antigo de consolidar localmente, mas NÃO SHALL impedir que um provedor execute uma chamada já recebida. Valores monetários e quantidades fracionárias SHALL usar decimal exato; timestamps SHALL ser UTC com fuso contratual preservado para fechamento. Histórico volumoso MAY ser particionado por tempo após medição, preservando a unicidade de idempotência em estrutura não sujeita a expurgo prematuro.

#### Scenario: Admissão persiste protocolo, idempotência, pedido e intenção em uma única transação
- GIVEN um novo pedido validado com chave idempotente, versões aplicáveis e intenção de trabalho definida
- WHEN a admissão é confirmada
- THEN o sistema SHALL persistir atomicamente o protocolo, a idempotência, o pedido/referência validada, as versões aplicáveis e a intenção de trabalho, sem estado intermediário parcialmente visível

#### Scenario: Sistema não mantém transação aberta esperando resposta externa
- GIVEN uma operação em andamento que depende de uma chamada HTTP a um provedor, de um arquivo ou de confirmação de saldo remoto
- WHEN a resposta externa demora além do esperado
- THEN o sistema NÃO SHALL manter a transação de banco aberta durante essa espera, liberando a conexão e tratando a resposta tardia por mecanismo assíncrono

#### Scenario: Atualização concorrente de estado é arbitrada por versão, e lease não neutraliza chamada já enviada ao provedor
- GIVEN um registro de protocolo sendo disputado por um worker antigo que perdeu a posse e um worker atual com lease válido
- WHEN ambos tentam consolidar o estado
- THEN o sistema SHALL verificar a versão do registro para impedir que o worker antigo consolide localmente, mas o lease NÃO SHALL impedir que o provedor execute uma chamada que já havia recebido antes da perda de posse

### Requirement: DAD-04 — Resultado como fonte de consulta
O hub SHALL conservar resultado normalizado, schema aplicado, versão, origem, horários e evidências necessárias, e o endpoint público NÃO SHALL usar um GET do cliente como gatilho para consultar o provedor, valendo isso para os estados PENDENTE, FINAL, EXPIRADO, conteúdo expurgado e indisponibilidade externa; o polling do provedor SHALL permanecer um processo independente de Cometa. Um resultado final SHALL ser servido do PostgreSQL ou de objeto pertencente ao hub, e o sistema NÃO SHALL devolver URL de provedor como única custódia da resposta; se o provedor retorna link temporário, Cometa SHALL capturar o conteúdo permitido pelo contrato antes de considerá-lo disponível no hub. O sistema NÃO SHALL prometer custódia quando o contrato do provedor proíbe armazenar os dados, exigindo decisão explícita antes de publicação nesse caso. Resultados expurgados SHALL retornar informação de indisponibilidade conforme retenção, sem recriar o pedido; "atualizar dados" SHALL ser tratado como nova execução com novo protocolo e avaliação de custo, não como consulta do resultado antigo. Cache opcional SHALL usar chave tenant/protocolo/versão, e cache miss SHALL ler a persistência do hub.

#### Scenario: GET do cliente nunca aciona consulta ao provedor, em qualquer estado do resultado
- GIVEN um pedido cujo resultado está em estado PENDENTE, FINAL, EXPIRADO ou expurgado
- WHEN o cliente realiza um GET no endpoint público
- THEN o sistema NÃO SHALL usar essa chamada como gatilho para consultar o provedor, servindo a resposta a partir do que já está conservado no hub ou informando indisponibilidade

#### Scenario: Link temporário de provedor deve ser capturado pelo hub antes de ser considerado disponível
- GIVEN um provedor que retorna um link temporário como forma de entrega do resultado
- WHEN Cometa recebe esse link
- THEN o sistema SHALL capturar o conteúdo permitido pelo contrato antes de considerar o resultado disponível no hub, e NÃO SHALL devolver a URL do provedor como única custódia da resposta

#### Scenario: Contrato que proíbe armazenamento bloqueia publicação sem decisão explícita
- GIVEN um contrato de provedor que proíbe o armazenamento dos dados retornados
- WHEN o serviço correspondente é avaliado para publicação no catálogo do hub
- THEN o sistema NÃO SHALL prometer custódia para essa oferta, e o serviço NÃO SHALL ser publicado sem uma decisão explícita anterior sobre como atender esse caso

#### Scenario: Resultado expurgado não recria o pedido e "atualizar dados" gera novo protocolo
- GIVEN um resultado já expurgado conforme a retenção aplicável
- WHEN o cliente consulta esse resultado ou solicita "atualizar dados"
- THEN o sistema SHALL retornar informação de indisponibilidade sem recriar o pedido original, e um pedido de atualização SHALL abrir uma nova execução com novo protocolo e nova avaliação de custo, nunca uma reconsulta do resultado antigo

### Requirement: DAD-05 — Arquivos e atomicidade com objetos
Entradas e saídas volumosas SHALL ficar no S3, e metadados e pequenos resultados tipados SHALL ficar no PostgreSQL. Objetos SHALL transitar pelos estados RESERVADO, RECEBIDO, EM_VALIDACAO, DISPONIVEL, REJEITADO ou EXPURGADO; reservar SHALL exigir proprietário, tamanho máximo, tipo aceito, finalidade e prazo, e concluir upload SHALL verificar tamanho, checksum, tipo real e controles de conteúdo aplicáveis, de modo que somente um objeto DISPONIVEL possa integrar um pedido executável. Como objeto e PostgreSQL não têm transação conjunta, o sistema SHALL primeiro confirmar o conteúdo íntegro no armazenamento e só então confirmar a referência e a conclusão no banco; falha entre essas etapas SHALL gerar um órfão para coleta, nunca um resultado final com referência inexistente. A reconciliação SHALL detectar órfãos, referências quebradas e objetos presos em validação. Saída obrigatória indisponível SHALL impedir a publicação de sucesso completo. Links externos arbitrários NÃO SHALL ser uma forma alternativa de upload, e download SHALL ser autorizado a cada emissão de URL, já que a chave não constitui segredo de autorização.

#### Scenario: Somente objeto no estado DISPONIVEL integra pedido executável
- GIVEN um objeto que passou pelas verificações de tamanho, checksum, tipo real e controles de conteúdo aplicáveis
- WHEN o objeto atinge o estado DISPONIVEL
- THEN o sistema SHALL permitir sua associação a um pedido executável; objetos em RESERVADO, RECEBIDO, EM_VALIDACAO, REJEITADO ou EXPURGADO NÃO SHALL ser aceitos para essa finalidade

#### Scenario: Falha entre confirmação do objeto e gravação da referência gera órfão, não referência quebrada
- GIVEN um upload cujo conteúdo já foi confirmado íntegro no S3
- WHEN ocorre uma falha (crash) antes que a referência e a conclusão sejam confirmadas no PostgreSQL
- THEN o sistema SHALL tratar o objeto como órfão sujeito a coleta pela reconciliação, e NÃO SHALL apresentar um resultado final com referência a um objeto inexistente

#### Scenario: Saída obrigatória indisponível impede publicação de sucesso completo
- GIVEN um pedido cujo contrato declara uma saída em objeto como obrigatória
- WHEN essa saída não está disponível no momento da conclusão
- THEN o sistema NÃO SHALL publicar o pedido como sucesso completo enquanto a saída obrigatória não estiver íntegra e disponível

### Requirement: DAD-06 — Retenção e expurgo
Todo dado SHALL ter classe, finalidade, retenção, proprietário e política de backup definidos, sendo os prazos parâmetros sujeitos a aprovação, não orientação jurídica. O sistema NÃO SHALL publicar um serviço sem prazo de retenção explícito e compatível com a obrigação ao cliente, e NÃO SHALL usar TTL para eliminar dado enquanto houver obrigação pendente (protocolos ativos, operações incertas ou resultado necessário a entrega/reenvio dentro da janela contratual). Chaves de idempotência, inbox e correlação SHALL ser retidas por pelo menos 35 dias e todo o período ativo, nunca menos que o necessário para replay, retry e callbacks aceitos. O expurgo SHALL abranger objetos, índices, caches e projeções; backups SHALL seguir seu próprio ciclo e acesso restrito. Uma solicitação de exclusão que conflite com preservação contratual SHALL ser encaminhada ao responsável pelos dados, com decisão auditável; um tombstone SHALL preservar apenas metadados mínimos autorizados, e o sistema NÃO SHALL conservar o payload sob outro nome como forma de contornar o expurgo.

#### Scenario: Serviço não é publicado sem prazo de retenção explícito
- GIVEN uma nova oferta em processo de publicação
- WHEN o prazo de retenção de protocolos, metadados e resultados não está definido explicitamente ou é incompatível com a obrigação contratual ao cliente
- THEN o sistema NÃO SHALL permitir a publicação dessa oferta

#### Scenario: Resultado necessário a entrega pendente não é removido antes do fim da janela contratual
- GIVEN um resultado ainda necessário para uma entrega ou reenvio dentro da janela contratual de entrega
- WHEN o prazo normal de retenção desse resultado já teria expirado
- THEN o sistema NÃO SHALL remover o resultado até o término da janela contratual de entrega/reenvio, pois a entrega NÃO SHALL depender de um resultado já removido

#### Scenario: Exclusão em conflito com preservação é decidida pelo responsável pelos dados, sem contornar via tombstone
- GIVEN uma solicitação de exclusão de dados que conflita com uma preservação autorizada por disputa ou obrigação contratual
- WHEN a solicitação é avaliada
- THEN o sistema SHALL encaminhar o caso ao responsável pelos dados para decisão auditável, e um eventual tombstone NÃO SHALL conservar o payload original sob outro nome ou identificador

### Requirement: DAD-07 — Isolamento, recuperação e auditoria
O sistema SHALL aplicar tenant_id nas entidades de cliente e nas constraints adequadas, usando RLS como defesa adicional; contas de aplicação NÃO SHALL ter BYPASSRLS, e o papel de migração NÃO SHALL ser o mesmo papel de runtime. O contexto de tenant em conexões reutilizadas SHALL ser transacional e limpo a cada uso. A política de backup SHALL incluir backups automáticos diários e PITR de 35 dias para PostgreSQL, com criptografia e restauração isolada ensaiada mensalmente, além de proteção de versões de objetos e cópia para recuperação regional conforme retenção aprovada; backups de core, financeiro e objetos SHALL formar um ponto de recuperação reconciliável, e o sistema NÃO SHALL prometer snapshot atômico entre bancos distintos e S3. Ao restaurar, o sistema SHALL bloquear a saída de chamadas, recuperar os dados, confrontar outboxes e fatos, consultar operações incertas de forma segura e reconciliar o ledger antes de retomar efeitos; um restore NÃO SHALL autorizar a reexecução de todos os pedidos do período.

#### Scenario: Conta de aplicação não possui BYPASSRLS e papel de migração é distinto do runtime
- GIVEN uma conexão de aplicação em runtime e uma conexão de migração de schema
- WHEN os privilégios de cada papel são avaliados
- THEN a conta de aplicação NÃO SHALL possuir BYPASSRLS, e o papel de migração NÃO SHALL ser o mesmo usado pelo runtime, mantendo RLS como defesa efetiva além do tenant_id aplicado nas entidades e constraints

#### Scenario: Contexto de tenant em conexão reutilizada é limpo a cada uso do pool
- GIVEN uma conexão de banco reutilizada pelo pool entre requisições de tenants diferentes
- WHEN uma nova requisição assume essa conexão
- THEN o sistema SHALL definir o contexto de tenant de forma transacional para essa requisição e SHALL limpá-lo ao final, impedindo vazamento de contexto entre tenants

#### Scenario: Restauração bloqueia efeitos externos e reconcilia antes de retomar, sem reexecutar todo o período
- GIVEN uma restauração de backup em andamento envolvendo core, financeiro e objetos
- WHEN os dados são recuperados
- THEN o sistema SHALL bloquear a saída de chamadas, confrontar outboxes e fatos, consultar operações incertas de forma segura e reconciliar o ledger antes de retomar efeitos, e o restore NÃO SHALL autorizar a reexecução automática de todos os pedidos do período restaurado

### Requirement: DAD-08 — Persistência adicional de contrato, prazo e pressão
O sistema SHALL conservar e indexar, por protocolo/operação, campos ampliados de contrato, prazo e pressão: no Protocolo, accepted_at, client_deadline_at, versões de contrato/retry/output e terminal_reason, com deadline absoluto NÃO SHALL renovado por mensagem, migração ou retry; no Passo/operação, first_transient_at, retry_until, SLA e deadline do provedor e budgets, distinguindo tempo total, tempo de tentativa e tempo de fila; na Observação externa, timestamps do provedor SHALL ser conservados como evidência separada, e NÃO SHALL substituir o relógio autoritativo do hub; na Representação do cliente, uma única representação final SHALL ser usada por GET e webhook, sem mudar em replay; no Domínio adaptativo, checkpoint durável e lease exclusivo SHALL ser mantidos, com sinais em memória reconstruíveis; na Permissão operacional, consulta entre tenants SHALL passar por política explícita, não por flag manipulável pelo cliente. O sistema SHALL indexar deadline de protocolos abertos, timers vencidos, fatos de SLA por período/contrato e relações tenant/célula, sem varrer a tabela de protocolo a cada tela de SLA. Uma tentativa tardia SHALL continuar vinculada à operação original e ao eventual custo, mas sua evidência NÃO SHALL entrar na resposta de sucesso de um protocolo já expirado.

#### Scenario: Deadline absoluto do protocolo não é renovado por retry, mensagem ou migração
- GIVEN um protocolo com client_deadline_at definido no momento do aceite
- WHEN ocorrem retries, novas mensagens ou uma migração de célula durante a execução
- THEN o sistema NÃO SHALL renovar ou estender o deadline absoluto original em razão desses eventos

#### Scenario: Timestamp do provedor é evidência separada e não substitui o relógio autoritativo do hub
- GIVEN uma observação externa cujo horário reportado pelo provedor diverge do horário de recebimento e persistência no hub
- WHEN a decisão de tempestividade é avaliada
- THEN o sistema SHALL registrar o timestamp do provedor como evidência separada, sem que ele substitua o relógio autoritativo do hub para decisões de prazo

#### Scenario: Tentativa tardia permanece vinculada à operação, mas não reabre resposta de sucesso já expirada
- GIVEN um protocolo já expirado cuja operação externa recebe uma tentativa/observação tardia
- WHEN essa evidência tardia chega ao hub
- THEN o sistema SHALL vincular a tentativa tardia e seu eventual custo à operação original, mas NÃO SHALL incluir essa evidência na resposta de sucesso do protocolo já expirado

#### Scenario: Representação final do cliente não muda em replay de GET ou webhook
- GIVEN uma representação final do cliente já materializada para um protocolo concluído
- WHEN o mesmo resultado é servido novamente por GET ou reenviado por webhook
- THEN o sistema SHALL usar exatamente a mesma representação final previamente materializada, sem gerar uma projeção diferente no replay

### Requirement: DAD-09 — Dependência mínima de escrita, sem custódia fictícia
Minimizar dependência SHALL significar eliminar consultas repetidas e operações duráveis desnecessárias, sem eliminar a evidência indispensável de execução; configuração, contratos e mapas já publicados SHALL ser tratados como projeções locais válidas, e o sistema NÃO SHALL consultar catálogo, cofre, relatórios, broker e financeiro pós-pago em série a cada chamada. As fronteiras síncronas do caminho SYNC — Aceite Órbita, Preparação Cometa, Observação Cometa e Final Órbita — SHALL registrar seus dados necessários antes de prosseguir e NÃO SHALL ser adiadas para depois da resposta ao cliente; publicação e apuração pós-paga MAY ser assíncronas, dentro de prazo, capacidade durável e limite de risco. Durante falha do writer, o sistema SHALL tentar recuperação de conexão/failover com orçamento curto e idempotência, sem manter transação aberta em espera externa; um commit com resposta perdida SHALL ser tratado como estado incerto, consultando a mesma autoridade recuperada, nunca assumindo rollback, e sem cópia autoritativa apta a confirmar, novas execuções NÃO SHALL ser enviadas ao provedor. O sistema NÃO SHALL trocar a escrita de aceite para Redis, disco de pod, memória, um banco independente ou uma fila avulsa para emitir aceite fictício; PostgreSQL gerenciado multi-AZ permanece a autoridade definida, e um journal distribuído alternativo NÃO SHALL assumir a admissão a menos que seja projetado como autoridade única de idempotência, ordem, resultados e recuperação.

#### Scenario: Fronteiras síncronas obrigatórias registram seus dados antes da resposta ao cliente
- GIVEN um pedido percorrendo Aceite Órbita, Preparação Cometa, Observação Cometa e Final Órbita no caminho SYNC
- WHEN cada uma dessas fronteiras é atravessada
- THEN o sistema SHALL registrar de forma durável os dados exigidos por cada fronteira antes de prosseguir, e a publicação/apuração pós-paga MAY ocorrer de forma assíncrona depois, dentro do prazo e do limite de risco definidos

#### Scenario: Commit com resposta perdida é tratado como incerto, consultando a mesma autoridade recuperada
- GIVEN uma escrita cujo commit foi aplicado no PostgreSQL, mas cuja confirmação de resposta se perdeu por falha de rede ou do writer
- WHEN o sistema se recupera dessa falha
- THEN o sistema SHALL consultar a chave da operação na mesma autoridade recuperada em vez de assumir rollback, e NÃO SHALL reenviar a operação ao provedor sem uma cópia autoritativa apta a confirmar o estado

#### Scenario: Sistema não emite aceite fictício por escrita alternativa durante falha do writer
- GIVEN uma falha do writer PostgreSQL durante a admissão de um pedido
- WHEN a aplicação avalia como responder ao cliente
- THEN o sistema NÃO SHALL trocar a escrita de aceite para Redis, disco local, memória, outro banco independente ou fila avulsa para simular um aceite; a admissão SHALL aguardar recuperação da autoridade definida ou retornar indisponibilidade

### Requirement: DAD-10 — Cache dispensável e continuidade de leitura
O fluxo de leitura SHALL seguir L1 em memória (limitado por bytes/entradas/idade), L2 Redis opcional quando saudável, e persistência do hub como fonte final, sem que nenhuma consulta pública acione o provedor; a chave de cache SHALL incluir tenant, protocolo, versão da representação e contrato técnico, e a idempotência de criação permanece autoritativa no banco independentemente do cache. Redis SHALL usar timeout pequeno, circuit breaker, bypass automático e recuperação gradual, com orçamento de ensaio de até 5 ms para lookup L2 saudável e até 10 ms de tentativa degradada antes do bypass. Durante indisponibilidade de escrita, um GET MAY servir um final imutável já confirmado de réplica ou snapshot local íntegro, dentro de retenção e autorização válidas; para versão corrente, é necessária projeção atual do ponteiro ou garantia publicada de não revisão, e cache antigo de final revisável NÃO SHALL ser tratado como prova de atualidade. O sistema NÃO SHALL retornar 404 definitivo, pendência supostamente atual ou sucesso novo apenas porque uma réplica atrasada desconhece o registro; uma consulta sem dado comprovadamente adequado SHALL retornar indisponibilidade transitória com correlação, sem fabricar resultado. Autorização SHALL ser verificada antes de entregar bytes, inclusive em fallback, e a disponibilidade de cache NÃO SHALL permitir acesso entre tenants nem ressuscitar conteúdo apagado.

#### Scenario: Redis indisponível aciona bypass automático sem comprometer a corretude da leitura
- GIVEN um L2 Redis que ultrapassa o orçamento de latência ou fica indisponível
- WHEN uma leitura de resultado imutável é solicitada
- THEN o circuit breaker SHALL acionar bypass automático para a persistência do hub, preservando a resposta correta ao custo de latência adicional, sem que nenhuma consulta seja encaminhada ao provedor

#### Scenario: Réplica atrasada não gera 404 definitivo nem sucesso fabricado
- GIVEN uma réplica de leitura atrasada que ainda não recebeu um registro recém-confirmado na autoridade de escrita
- WHEN um GET consulta esse registro através da réplica
- THEN o sistema NÃO SHALL retornar 404 definitivo, pendência supostamente atual, ou sucesso novo; SHALL retornar indisponibilidade transitória com correlação quando conhecida, sem fabricar um resultado

#### Scenario: Cache antigo de final revisável não prova atualidade para versão corrente
- GIVEN uma classe de resultado final que admite revisão, com uma entrada de cache antiga referente ao ponteiro atual
- WHEN um GET solicita a versão corrente desse resultado
- THEN o sistema NÃO SHALL tratar o cache antigo como prova de atualidade, exigindo projeção atual do ponteiro ou garantia publicada de que aquela classe de final não admite revisão

#### Scenario: Autorização insuficiente suspende a leitura mesmo em fallback de cache
- GIVEN um fallback de cache/réplica sendo usado para servir um GET durante indisponibilidade de escrita
- WHEN não é possível garantir a atualidade mínima exigida de autorização ou retenção para esse conteúdo
- THEN o sistema SHALL suspender a leitura afetada em vez de entregar os bytes, e a disponibilidade de cache NÃO SHALL permitir acesso entre tenants nem ressuscitar conteúdo já expurgado

### Requirement: DAD-11 — Dados, placement e crescimento de células
Novos tenants SHALL ser alocados em células prontas com capacidade reservável, e o crescimento do número de clientes SHALL ser atendido por novas células do mesmo perfil, sem criar tabela, schema, fila física ou deployment exclusivo para cada cliente compartilhado. O mapa tenant→célula SHALL ter versão/epoch e origem autenticada, distribuído pelo Atlas aos gateways e runtimes; o cliente NÃO SHALL escolher célula por header, e para uma chamada, tenant autorizado e mapa SHALL determinar a autoridade de protocolo sem consulta a um diretório global por UUID no caminho habitual. Na referência inicial, um tenant SHALL possuir uma célula ativa de admissão e uma autoridade financeira estrita; fragmentar um único saldo/protocolo entre várias células ativas NÃO SHALL ser comportamento implícito, e workloads acima do maior perfil homologado SHALL exigir novo perfil qualificado antes de oferta comercial. A realocação, quando necessária, SHALL ser um workflow automatizado que prepara destino, drena ou reconcilia operações incertas, estabelece barreira de admissão com espera limitada, completa cópia incremental, revoga/fenceia o epoch de escrita da origem, ativa o destino e publica novo mapa antes de retirar a origem; se não couber no orçamento homologado, o sistema SHALL manter a autoridade antiga e abortar a migração antes do corte, já que uma operação incerta pode impedir a migração mas NÃO SHALL autorizar duas autoridades ativas simultâneas.

#### Scenario: Novo tenant é alocado em célula pronta sem infraestrutura exclusiva
- GIVEN uma célula existente do perfil compatível com capacidade reservável disponível
- WHEN um novo tenant compartilhado é integrado ao hub
- THEN o sistema SHALL alocá-lo nessa célula existente, sem criar tabela, schema, fila física ou deployment exclusivo para esse cliente

#### Scenario: Cliente não seleciona célula por header, e o caminho habitual não consulta diretório global
- GIVEN uma chamada de um tenant autorizado, cujo mapa tenant→célula já está publicado com versão/epoch e origem autenticada
- WHEN a chamada tenta indicar uma célula específica via header
- THEN o sistema NÃO SHALL aceitar essa escolha; a autoridade de protocolo SHALL ser determinada pelo tenant autorizado combinado ao mapa vigente, sem consulta a um diretório global por UUID no caminho habitual

#### Scenario: Migração é abortada e autoridade antiga é mantida quando o orçamento homologado não é atendido
- GIVEN uma realocação de célula em andamento com operações externas incertas ainda não reconciliadas ou uma cópia incremental que excede o orçamento homologado
- WHEN o corte para o destino é avaliado
- THEN o sistema SHALL manter a autoridade antiga e abortar a migração antes do corte, e a existência de uma operação incerta NÃO SHALL ser usada para justificar duas autoridades ativas simultâneas

#### Scenario: Fragmentação de saldo entre múltiplas células ativas não é comportamento implícito
- GIVEN um tenant com uma única autoridade financeira estrita ativa em sua célula de admissão
- WHEN o crescimento de carga desse tenant é avaliado
- THEN o sistema NÃO SHALL fragmentar implicitamente o saldo/protocolo entre múltiplas células ativas; o atendimento SHALL ocorrer pelos recursos/limites homologados da célula atual, por aumento de capacidade, por realocação planejada, ou por qualificação de um novo perfil antes de oferta comercial

## Notas de origem

Este delta deriva integralmente do capítulo `docs/03_PERSISTENCIA_E_DADOS.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos DAD-01 a DAD-11 conforme texto normativo de origem.
