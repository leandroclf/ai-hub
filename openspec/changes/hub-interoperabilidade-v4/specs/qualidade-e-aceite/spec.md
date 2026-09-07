# Qualidade, Rastreabilidade e Aceite — Delta de Especificação

## ADDED Requirements

### Requirement: QUA-01 — Estratégia de engenharia
O processo de qualidade e promoção SHALL exigir que cada requisito identificado em CAT, ARQ, COM, DAD, EXE, FIN, CFG, SEG e OPE — incluindo os próprios requisitos de qualidade e decisão — tenha cenário positivo, cenários negativos relevantes e evidência verificável mapeados em uma matriz de rastreabilidade única. A matriz NÃO SHALL ser interpretada como declaração de testes executados: o status inicial de todo cenário mapeado SHALL ser NÃO EXECUTADO até que implementação e evidência real existam. O processo SHALL cobrir, no mínimo, os níveis de revisão de domínio e contratos, testes de invariantes transacionais, integração PostgreSQL/broker/objetos, contrato de cada provedor/cliente, ponta a ponta, concorrência/falhas, segurança, carga e longa duração, recuperação e reconciliação financeira. Cobertura percentual de código, quando existir implementação, SHALL ser tratada como sinal auxiliar, NÃO SHALL substituir a cobertura de riscos e requisitos. O processo SHALL manter uma coleção sintética de referência contendo, no mínimo, serviço síncrono, assíncrono só callback, assíncrono só polling, ambos, operação sem idempotência, callback antecipado, resultado grande, produto agregado, produto composto, provedor que cobra status, cliente com franquia e contrato com saldo estrito. O simulador local SHALL reproduzir comportamento determinístico e falhas injetadas; a homologação SHALL usar também sandbox real para capacidades e limites não simuláveis pelo simulador.

#### Scenario: Matriz de rastreabilidade nasce com status NÃO EXECUTADO
- GIVEN um requisito recém-mapeado de qualquer capítulo (CAT, ARQ, COM, DAD, EXE, FIN, CFG, SEG, OPE ou QUA)
- WHEN esse requisito é incluído na matriz de rastreabilidade com seus cenários positivo e negativos
- THEN o processo SHALL registrar o status inicial de cada cenário como NÃO EXECUTADO, sem declarar teste executado apenas por existir na matriz

#### Scenario: Cobertura de código alta não substitui cobertura de risco
- GIVEN uma implementação com cobertura percentual de código elevada, mas sem cenário negativo mapeado para um invariante de risco relevante
- WHEN o gate de qualidade avalia a suficiência da cobertura
- THEN o processo NÃO SHALL aceitar a cobertura percentual de código como substituto da cobertura de riscos e requisitos, exigindo o cenário negativo ausente

#### Scenario: Coleção sintética de referência contém os casos obrigatórios
- GIVEN a coleção sintética de referência mantida para a estratégia de engenharia
- WHEN essa coleção é auditada quanto à sua composição
- THEN o processo SHALL exigir a presença de serviço síncrono, assíncrono só callback, assíncrono só polling, ambos, operação sem idempotência, callback antecipado, resultado grande, produto agregado, produto composto, provedor que cobra status, cliente com franquia e contrato com saldo estrito

#### Scenario: Homologação usa sandbox real para capacidades não simuláveis
- GIVEN uma capacidade ou limite de provedor que o simulador local não reproduz fielmente (por exemplo, quota real, latência de rede externa ou comportamento proprietário)
- WHEN a fatia correspondente avança para homologação
- THEN o processo SHALL exigir o uso de sandbox real do provedor para essa capacidade, além do simulador local determinístico usado nos níveis anteriores

### Requirement: QUA-02 — Casos críticos de falha e concorrência
O processo de qualidade SHALL exigir, para cada caso crítico de falha e concorrência identificado, uma injeção de falha reproduzível e a verificação do invariante esperado correspondente como critério de aceite, cobrindo no mínimo commit sem resposta ao cliente, banco gravado com broker fora, publicação sem marca local, consumidor nas fronteiras de ack, timeout externo ambíguo, callback antecipado, callback/polling simultâneos, finais divergentes, expiração de lease com chamada em voo, quota de provedor compartilhada sob escala horizontal, cobrança dupla em polling/fetch pagos, saldo disputado, objeto gravado sem commit, indisponibilidade de webhook, expurgo de resultado e restauração com efeitos externos já realizados. O processo SHALL exigir também, além desses casos, testes de alteração de preço em voo, vencimento de contrato, evento antigo, incompatibilidade de schema, revogação, certificado vencido, troca de tenant, SSRF, SQL/RLS, XML malicioso quando SOAP for usado e consumo excessivo de recursos por arquivos. Todo caso desta seção permanece NÃO EXECUTADO até implementação e evidência real.

#### Scenario: Commit sem resposta ao cliente preserva idempotência
- GIVEN uma operação que perde a conexão com o cliente logo após a admissão durável ser confirmada
- WHEN o cliente reenvia a mesma chave de idempotência
- THEN o processo SHALL exigir que o sistema recupere o mesmo protocolo original, sem criar nova operação

#### Scenario: Banco gravado com broker fora preserva o outbox
- GIVEN uma gravação bem-sucedida no banco seguida de interrupção da publicação no broker
- WHEN a publicação é restaurada posteriormente
- THEN o processo SHALL exigir que o outbox tenha preservado o evento e que a publicação ocorra depois, sem perda do pedido

#### Scenario: Timeout externo ambíguo resulta em UNKNOWN sem failover indevido
- GIVEN um provedor que processa a operação mas cuja resposta se perde antes de chegar ao Hub
- WHEN o timeout é detectado
- THEN o processo SHALL exigir que o sistema marque o estado como UNKNOWN, realize consulta segura ao provedor e NÃO dispare failover automático indevido

#### Scenario: Callback e polling simultâneos produzem uma única conclusão
- GIVEN callback e polling entregando o mesmo resultado final em paralelo
- WHEN ambos os canais chegam próximos no tempo
- THEN o processo SHALL exigir uma única conclusão, uma entrega por destino e uma única unidade de receita, sem duplicação

#### Scenario: Quota de provedor compartilhada respeita limite global sob escala horizontal
- GIVEN um aumento no número de pods e tenants compartilhando a mesma quota de provedor
- WHEN a carga é distribuída entre as réplicas
- THEN o processo SHALL exigir que o limite global da quota seja respeitado, sem que o HPA multiplique a quota por réplica

#### Scenario: Saldo disputado não excede o limite estrito do contrato
- GIVEN pedidos simultâneos concorrendo pela última unidade de saldo de um contrato estrito
- WHEN as requisições chegam ao mesmo tempo
- THEN o processo SHALL exigir que nenhum efeito ocorra sem reserva prévia e que o consumo total não exceda o limite estrito

#### Scenario: Webhook indisponível não gera reexecução indevida
- GIVEN uma janela de indisponibilidade do endpoint de webhook seguida de recuperação
- WHEN as tentativas de entrega se esgotam durante a janela
- THEN o processo SHALL exigir estado EXHAUSTED ou retentativa rastreada, com consulta GET continuando a funcionar localmente e sem reexecução do serviço original

#### Scenario: Restauração de backup com efeitos externos já realizados bloqueia saída até reconciliar
- GIVEN uma restauração de backup anterior a operações que já produziram efeitos externos reais
- WHEN o sistema restaurado é colocado em operação
- THEN o processo SHALL exigir que a saída fique bloqueada até a reconciliação, sem repetir custos ou efeitos externos às cegas

#### Scenario: Casos adicionais de segurança e contrato são cobertos além da tabela principal
- GIVEN os casos adicionais de alteração de preço em voo, vencimento de contrato, evento antigo, incompatibilidade de schema, revogação, certificado vencido, troca de tenant, SSRF, SQL/RLS, XML malicioso em SOAP e consumo excessivo de recursos por arquivos
- WHEN a suíte de qualidade planeja sua cobertura
- THEN o processo SHALL exigir que cada um desses casos tenha injeção de falha e invariante esperado definidos, com status NÃO EXECUTADO até evidência real

### Requirement: QUA-03 — Critérios de aceite e evidência
O processo de qualidade SHALL exigir que um caso de teste só seja considerado aprovado quando registrar pré-condições, massa de dados, versão de contrato/configuração, instantes, resultado observado, protocolo de teste e comparação com o esperado. Evidência de desempenho SHALL registrar hardware/cluster, réplicas, requests/limits, versões, latência do simulador/provedor, taxa, tamanho, fan-out, duração, erros e percentis; um relatório sem essas condições NÃO SHALL qualificar um SLO. Para ausência de perda, o processo SHALL exigir a contabilização do conjunto de entradas aceitas confrontado com protocolos, obrigações, resultados e pendências, e NÃO SHALL aceitar "fila vazia" como prova de completude. Para faturamento, o processo SHALL exigir recomputo de exemplos sintéticos por calculadora independente, verificação de chaves econômicas, balanceamento por moeda, arredondamento e conciliação de extrato. Para consulta local, o processo SHALL exigir instrumentação de contagem de chamadas ao provedor antes/depois e zero novas chamadas atribuíveis ao GET, além da comparação de hash e contrato da representação final entre GET e webhook, incluindo falha de SLA, payload legado e reentrega. O processo SHALL tratar como bloqueadores de release: perda de pedido aceito, acesso entre tenants, cobrança duplicada, resultado final sobrescrito sem revisão, retry inseguro de operação incerta, fechamento sem explicação de divergências, segredo exposto e protocolo sem possibilidade de recuperação; a avaliação de outros defeitos SHALL depender de impacto, e o processo NÃO SHALL reclassificar bloqueadores para cumprir prazo.

#### Scenario: Relatório de desempenho incompleto não qualifica um SLO
- GIVEN um relatório de desempenho que omite réplicas, requests/limits ou percentis de latência
- WHEN esse relatório é apresentado como evidência de cumprimento de SLO
- THEN o processo NÃO SHALL qualificar o SLO com base nesse relatório, exigindo o registro completo de hardware/cluster, réplicas, requests/limits, versões, latência, taxa, tamanho, fan-out, duração, erros e percentis

#### Scenario: "Fila vazia" não prova ausência de perda
- GIVEN uma fila de processamento observada vazia ao final de um ciclo de teste
- WHEN essa observação é usada como evidência de ausência de perda de pedidos
- THEN o processo NÃO SHALL aceitar "fila vazia" isoladamente como prova, exigindo a contabilização do conjunto de entradas aceitas confrontado com protocolos, obrigações, resultados e pendências

#### Scenario: Faturamento é validado por calculadora independente antes do aceite
- GIVEN um conjunto de exemplos sintéticos de faturamento produzidos pelo sistema
- WHEN a evidência de aceite financeiro é preparada
- THEN o processo SHALL exigir o recomputo desses exemplos por calculadora independente, verificando chaves econômicas, balanceamento por moeda, arredondamento e conciliação de extrato

#### Scenario: Consulta local (GET) não gera nova chamada ao provedor
- GIVEN uma contagem instrumentada de chamadas ao provedor antes e depois de uma sequência de consultas GET
- WHEN as consultas GET são executadas sobre um resultado já obtido
- THEN o processo SHALL exigir zero novas chamadas ao provedor atribuíveis ao GET, comparando também hash e contrato da representação final entre GET e webhook

#### Scenario: Bloqueador de release não é reclassificado para cumprir prazo
- GIVEN um defeito identificado como perda de pedido aceito, acesso entre tenants, cobrança duplicada ou outro item da lista de bloqueadores
- WHEN a data prevista de release se aproxima
- THEN o processo NÃO SHALL reclassificar esse defeito para uma severidade menor apenas para cumprir o prazo, mantendo-o como bloqueador até resolução

### Requirement: QUA-04 — Gates e definição de pronto
O processo de promoção SHALL organizar a evolução de cada fatia em gates sequenciais G0 a G4 e SHALL bloquear a promoção ao gate seguinte enquanto os critérios do gate atual não estiverem cumpridos. G0 (pronto para implementar) SHALL exigir domínio, estados, matrizes de atendimento, contratos econômicos, modelo de dados e critérios de aceite revisados, pendências que afetam a fatia resolvidas e responsáveis nominais definidos, sem exigir conta de produção para uma fatia sintética e sem inventar condições comerciais. G1 (local/dev) SHALL exigir implementação da fatia, revisão de código futura, testes de contrato/integração, Compose e kind, probes e telemetria demonstrados, com emulação identificada como tal. G2 (hom) SHALL exigir cliente/provedor sandbox homologados, catálogo publicável, TTL/deadline e rejeição tardia demonstrados, contratos legados/GET-webhook equivalentes, polling/callback e apuração aprovados, e acesso administrativo e testes negativos entre tenants aprovados. G3 (ppd) SHALL exigir infraestrutura representativa por célula, carga/longa duração, ruído entre tenants, adaptação de pressão, escala para cima/baixo, perda de zona e controlador, restore, aferição bilateral de SLA, dashboards e runbooks ensaiados, usando o mesmo artefato destinado à produção e exigindo prova correspondente ao RPO assumido, incluindo escala automática de nós/células, falta de Redis, writer/cofre indisponíveis, modo SYNC sem broker e manutenção sob carga. G4 (prd) SHALL exigir SLOs aprovados, contratos e retenção válidos, observabilidade e plantão ativos, segregação e credenciais verificadas, plano de rollback/restore aprovado e nenhuma falha bloqueadora; oferta crítica SHALL exigir também perfil OPE-15, limites de interrupção e cenários QUA-06 aprovados. O processo NÃO SHALL declarar a plataforma pronta em G4 com endpoints de negócio retornando "não implementado". A fatia inicial recomendada é o serviço síncrono de provedor disponível em SYNC direto e ASYNC, com UUIDv7, resultado local, credencial compartilhada/dedicada e regra simples de receita/custo, seguida por polling/callback concorrentes e depois produto composto e planos avançados; essa sequência é de entrega, NÃO SHALL ser interpretada como exclusão de requisitos da v4, e não há prazo de calendário presumido. Ausência de Redis e falhas nas fronteiras de escrita SHALL ser qualificadas desde a primeira fatia.

#### Scenario: G0 não exige conta de produção para fatia sintética
- GIVEN uma fatia sintética com domínio, estados, contratos econômicos e critérios de aceite revisados
- WHEN a fatia é avaliada para promoção ao gate G0
- THEN o processo SHALL permitir a promoção sem exigir conta de produção, mas NÃO SHALL aceitar condições comerciais inventadas em seu lugar

#### Scenario: G2 bloqueia promoção sem testes negativos entre tenants aprovados
- GIVEN uma fatia em homologação cujo catálogo, TTL/deadline e apuração já foram demonstrados, mas cujos testes negativos entre tenants ainda não foram aprovados
- WHEN a promoção ao gate G2 é solicitada
- THEN o processo SHALL bloquear a promoção até que o acesso administrativo e os testes negativos entre tenants sejam aprovados

#### Scenario: G3 exige o mesmo artefato de produção e prova correspondente ao RPO assumido
- GIVEN uma fatia em ppd que assume um RPO específico nos dashboards e runbooks ensaiados
- WHEN a promoção ao gate G3 é avaliada
- THEN o processo SHALL exigir que o artefato testado seja o mesmo destinado à produção e SHALL exigir prova correspondente ao RPO assumido, não apenas a suposição documental

#### Scenario: G4 não declara produção pronta com endpoints "não implementado"
- GIVEN uma fatia com SLOs, observabilidade e plano de rollback aprovados, mas com algum endpoint de negócio ainda retornando "não implementado"
- WHEN a promoção ao gate G4 é avaliada
- THEN o processo NÃO SHALL declarar a plataforma pronta para produção enquanto esse endpoint retornar "não implementado"

#### Scenario: Oferta crítica em G4 exige perfil OPE-15 e cenários QUA-06 aprovados
- GIVEN uma oferta classificada como crítica avançando para o gate G4
- WHEN os critérios de G4 são avaliados para essa oferta
- THEN o processo SHALL exigir adicionalmente o perfil OPE-15, os limites de interrupção definidos e os cenários QUA-06 aprovados, além dos critérios gerais de G4

### Requirement: QUA-05 — Qualificação dos incrementos v3
O processo de qualidade SHALL exigir, para os incrementos v3, procedimentos de qualificação com critério de aceite específico para cada caso obrigatório, cobrindo no mínimo TTL contínuo através de reinícios e troca de worker, TTL zero e erro permanente, provedor lento sem falhas HTTP, fronteira temporal do deadline, recebimento cedo com consolidação tardia, retorno tardio com impacto financeiro, callback/poll concorrentes cruzando o deadline, pressão variável do provedor simulado, múltiplas réplicas do Cometa, ausência de métrica de controle, acúmulo assíncrono, cliente ruidoso, contrato legado por cliente, equivalência entre GET e webhook, administração entre tenants e preservação de fatos confirmados após falha. Os cenários SHALL incluir a matriz dos contratos, passos paralelos/sequenciais, tempos precisos e histórico da decisão adaptativa, e SHALL comparar o controlador adaptativo com um limite fixo conservador no mesmo provedor simulado quanto a throughput útil, erros, latência, oscilação, tempo de recuperação e consumo faturável; mais requisições enviadas NÃO SHALL ser considerada melhoria se o final útil e o SLA piorarem. Todos os testes desta seção permanecem NÃO EXECUTADOS até implementação e evidência; a revisão documental garante cobertura descrita e coerência, NÃO SHALL ser tomada como prova de isolamento, arbitragem de commit ou RPO. Casos de deadline, acesso entre tenants, duplicação econômica e perda de fato confirmado SHALL bloquear a promoção.

#### Scenario: TTL contínuo mantém a mesma janela através de reinício e troca de worker
- GIVEN um primeiro erro transiente seguido de retry, reinício de pod e troca de worker durante a mesma operação
- WHEN o processo tenta renovar a janela de TTL
- THEN o critério de aceite SHALL exigir o mesmo first_transient_at/retry_until e um fim previsível em segundos, sem reiniciar a contagem a cada troca

#### Scenario: TTL zero e erro permanente não geram repetição indevida
- GIVEN um serviço configurado sem retry e outro que retorna erro de contrato permanente
- WHEN a operação com TTL zero é executada em ambos os serviços
- THEN o critério de aceite SHALL exigir ausência de repetição indevida e resposta final compatível com o previsto pelo cliente

#### Scenario: Provedor lento sem falhas HTTP resulta em expiração explícita, não em sucesso tardio
- GIVEN um provedor cujos submit e polls retornam sempre 200, mas cujo resultado final chega depois do deadline
- WHEN o deadline é ultrapassado
- THEN o critério de aceite SHALL exigir estado EXPIRED/SLA_EXCEEDED e a rejeição do resultado final tardio

#### Scenario: Fronteira temporal do deadline produz um único final
- GIVEN callback, poll e finalização ocorrendo em D−ε, D e D+ε, com atraso deliberado na transação e no scheduler
- WHEN os eventos concorrentes cruzam a fronteira do deadline
- THEN o critério de aceite SHALL exigir um único final, tratando a igualdade com D como atraso, sem sucesso sem elegibilidade durável comprovada

#### Scenario: Retorno tardio com impacto financeiro não gera receita de sucesso
- GIVEN um provedor cobrado por aceite que conclui a operação somente após o estado EXPIRED
- WHEN esse retorno tardio chega ao Hub
- THEN o critério de aceite SHALL exigir zero receita de sucesso, preservação do custo/contestação e ausência de reabertura do protocolo

#### Scenario: Pressão variável ajusta o limite efetivo sem ultrapassar o teto
- GIVEN um provedor simulado que dobra a capacidade e depois cai para metade, com injeção de 429/timeouts
- WHEN o controlador adaptativo reage a essas variações
- THEN o critério de aceite SHALL exigir que o limite efetivo caia e se recupere gradualmente, sem ultrapassar o teto nem amplificar retries

#### Scenario: Múltiplas réplicas do Cometa respeitam a quota agregada com fencing
- GIVEN uma escala do Cometa com perda do controlador e disputa de posse antiga
- WHEN novas réplicas assumem a coordenação
- THEN o critério de aceite SHALL exigir que a quota agregada seja respeitada, com fencing efetivo, sem multiplicação do orçamento por réplica

#### Scenario: Cliente ruidoso não consome a reserva de outro tenant
- GIVEN o tenant A excedendo sua quota enquanto o tenant B mantém carga dentro do contratado, com saturação de CPU, banco, logs e provedor atribuída a A
- WHEN essa saturação ocorre simultaneamente
- THEN o critério de aceite SHALL exigir que B mantenha seu SLO dentro do domínio qualificado e que A não tome a reserva de B

#### Scenario: Administração entre tenants nega cliente e audita administrador
- GIVEN um cliente tentando acessar outro tenant, um desenvolvedor com e sem perfil administrativo global, e uma concessão revogada
- WHEN cada uma dessas tentativas de acesso ocorre
- THEN o critério de aceite SHALL exigir que o cliente seja sempre negado, que o administrador autorizado seja aceito e auditado, e que o acesso revogado seja negado

#### Scenario: Preservação de fatos confirmados após falha correlacionada
- GIVEN um crash após commits, falha de broker e de zona, seguidos de restore e reconciliação
- WHEN a recuperação é concluída
- THEN o critério de aceite SHALL exigir que nenhum fato confirmado dentro do modelo seja perdido, com lacunas identificadas explicitamente e o compromisso de RPO documentado

### Requirement: QUA-06 — Qualificação integrada da v4
O processo de qualidade SHALL exigir, para a qualificação integrada da v4, procedimentos com resultado esperado e evidência específicos cobrindo no mínimo SYNC nativo com broker cortado, comparação SYNC versus AUTO, UUID em todo aceite, fronteira DIRECT/QUEUED, dupla observação da mesma operação, prazo SYNC e TTL, paralelismo SYNC, Redis totalmente desligado, Redis intermitente, writer indisponível antes do aceite, commit incerto sem custódia do retorno externo, réplica atrasada, cofre quente/frio, credencial dedicada ausente, rotação de segredo em voo, modelos de conta e pagador, escala sem chamado humano, quota negada com controlador de infraestrutura fora do ar, conta global em várias células, realocação interrompida, manutenção sob carga, Redis versus autorização, falha correlacionada crítica e perda regional/partição. A massa de teste SHALL cruzar modalidades, classes de criticidade, conta compartilhada/dedicada, contrato legado, agregação/composição e saldo estrito/pós-pago. O baseline de performance SHALL incluir o núcleo durável completo, autenticação, transformação e provedor simulado controlável, seguido de sandbox real. Desligar o Redis SHALL ser tratado como requisito de aceite, não apenas como opção de benchmark; provar a ausência de chamada à fila obrigatória de SYNC NÃO SHALL ser interpretado como desligar a outbox de fatos. A análise de falhas SHALL relacionar cada risco à ação preventiva, detecção, fallback, recuperação, dependência e responsável. Os ensaios SHALL registrar números antes/depois de latência final SYNC, aceite ASYNC, GET, throughput útil, erros, backlog, instante do efeito e dos commits, uso de conexões/memória, conta/credencial sem segredo exposto e completude econômica; quando não houver dado suficiente, o resultado SHALL ser NÃO QUALIFICADO, e não aprovação por ausência de erro observado. Cobertura documental NÃO SHALL ser tomada como medida de confiabilidade; todos os casos desta seção e as linhas da matriz permanecem NÃO EXECUTADOS. Falta de teste crítico SHALL bloquear apenas a oferta/perfil afetado, e a produção dessa oferta NÃO SHALL ser liberada por aceite de uma fatia menos exigente.

#### Scenario: SYNC nativo conclui na mesma conexão com broker cortado
- GIVEN um provedor síncrono, core saudável e broker cortado, com capacidade de outbox suficiente
- WHEN a operação SYNC é executada
- THEN o resultado esperado SHALL exigir final na mesma conexão, traço sem espera de fila, e todos os commits e fatos recuperáveis apesar do broker cortado

#### Scenario: UUIDv7 persiste em todo protocolo mesmo com resposta de criação perdida
- GIVEN uma operação em modo SYNC, ASYNC ou AUTO cuja resposta de criação se perde antes de chegar ao cliente, em cenário de sucesso ou falha
- WHEN o cliente reenvia com a mesma chave de idempotência
- THEN o resultado esperado SHALL exigir UUIDv7 persistido em cada protocolo, com a mesma chave recuperando o mesmo UUID, sem que o conhecimento do ID conceda acesso

#### Scenario: Redis totalmente desligado mantém o SLO qualificado sem dependência oculta
- GIVEN um pico de carga contratado, cache L1 frio e perda de uma zona, com Redis totalmente desligado
- WHEN operações válidas são processadas nessas condições
- THEN o resultado esperado SHALL exigir que essas operações mantenham o SLO qualificado, sem dependência oculta de lock, idempotência ou quota no Redis

#### Scenario: Writer indisponível antes do aceite impede falso aceite
- GIVEN a autoridade transacional do writer cortada e uma nova operação de criação enviada
- WHEN essa criação é processada sem writer disponível
- THEN o resultado esperado SHALL exigir que nenhum efeito externo ocorra e que nenhum falso aceite seja emitido, mantendo leitura comprovada independente quando possível

#### Scenario: Cofre quente/frio limita uso do material ao período de validade
- GIVEN a perda do Secrets Manager/KMS, com e sem material de credencial previamente obtido
- WHEN o workload tenta usar esse material durante a indisponibilidade do cofre
- THEN o resultado esperado SHALL exigir uso apenas dentro da validade, bloqueio do binding afetado em cold start, e ausência de troca de conta como contorno

#### Scenario: Conta e pagador permanecem corretos e segregados sob polls pagos e final tardio
- GIVEN três modelos de FIN-11 operando com polls pagos e um final tardio
- WHEN o fechamento financeiro é apurado
- THEN o resultado esperado SHALL exigir custo, receita e conta a pagar corretos e segregados por modelo, sem débito presumido apenas pelo tipo da chave usada

#### Scenario: Redis versus autorização nega leitura mesmo em fallback
- GIVEN um cache Redis contendo resultado de outro tenant ou associado a um acesso já revogado
- WHEN uma leitura em modo de fallback é tentada usando esse cache
- THEN o resultado esperado SHALL exigir que a leitura seja negada mesmo em fallback, sem vazamento por bypass do cache

#### Scenario: Perda regional/partição exige RPO/RTO comprovados dentro do perfil contratado
- GIVEN um exercício de perda regional ou partição de rede dentro do perfil de continuidade contratado
- WHEN esse exercício é conduzido
- THEN o resultado esperado SHALL exigir autoridade única preservada e RPO/RTO comprovados para o perfil contratado, permanecendo bloqueado qualquer perfil regional ainda não qualificado

#### Scenario: Dado insuficiente resulta em NÃO QUALIFICADO, não em aprovação por ausência de erro
- GIVEN um ensaio de qualificação integrada cujos números antes/depois de latência, throughput, backlog ou completude econômica não foram registrados de forma suficiente
- WHEN o resultado desse ensaio é apurado
- THEN o processo SHALL registrar o resultado como NÃO QUALIFICADO, e NÃO SHALL aprovar o caso apenas por ausência de erro observado

## Notas de origem

Este delta deriva integralmente do capítulo `docs/08_QUALIDADE_E_ACEITE.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos QUA-01, QUA-02, QUA-03, QUA-04, QUA-05 e QUA-06 conforme texto normativo de origem.
