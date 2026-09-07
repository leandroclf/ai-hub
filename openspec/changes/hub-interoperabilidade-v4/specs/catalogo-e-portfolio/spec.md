# Catálogo e Portfólio — Delta de Especificação

## ADDED Requirements

### Requirement: CAT-01 — Catálogo de serviços
Cada versão de serviço canônico SHALL declarar código estável, descrição funcional, classificação dos dados, schema de entrada e de resultado, validações, unidades, condições de resultado válido/sem resultado, efeitos externos, modos de atendimento do hub, deadline de negócio, retenção, elegibilidade comercial, SLA e política de resultado tardio. O sistema SHALL distinguir erro técnico de uma resposta negativa válida, de modo que "não encontrado" possa ser tratado como sucesso funcional. A publicação de uma versão SHALL exigir ao menos um vínculo de provedor homologado ou classificação explícita como serviço interno. Versões publicadas SHALL ser imutáveis. Uma versão descontinuada SHALL permanecer interpretável para protocolos históricos, e novas admissões SHALL cessar na data programada.

#### Scenario: Publicação de versão com metadados completos e vínculo homologado
- GIVEN uma nova versão de serviço canônico com schema de entrada/resultado, validações, unidades, SLA, deadline de negócio e retenção declarados
- WHEN a versão possui ao menos um vínculo de provedor homologado ou está classificada explicitamente como serviço interno
- THEN a versão pode ser publicada e passa a ser imutável

#### Scenario: Entrada incompatível é recusada antes da execução
- GIVEN uma versão publicada com schema de entrada definido
- WHEN um pedido chega com entrada incompatível com o schema declarado
- THEN o sistema SHALL recusar a entrada antes de iniciar a execução, sem consumir efeitos externos

#### Scenario: Consulta histórica de versão descontinuada preserva semântica original
- GIVEN uma versão de serviço descontinuada com protocolos históricos associados
- WHEN uma consulta histórica é realizada sobre esses protocolos
- THEN o sistema SHALL preservar a semântica original da versão na resposta, e novas admissões para essa versão SHALL permanecer cessadas a partir da data programada

### Requirement: CAT-02 — Portfólio e disponibilidade
O sistema SHALL representar cada oferta em um dos estados RASCUNHO, EM_VALIDACAO, PUBLICADA, SUSPENSA ou DESCONTINUADA, tratando disponibilidade do catálogo, elegibilidade do contrato e saúde dos provedores como dimensões distintas. Suspender novas vendas de uma oferta NÃO SHALL cancelar pedidos já aceitos. O portal SHALL explicar indisponibilidade sem expor segredos, preços de aquisição ou dados de outros tenants. O cliente SHALL visualizar apenas ofertas contratadas e versões autorizadas, e a decisão de admissão SHALL considerar datas de início/fim, região, canal, quotas e permissão para dados sensíveis.

#### Scenario: Cliente visualiza apenas ofertas contratadas e autorizadas
- GIVEN um cliente com um conjunto específico de ofertas contratadas e versões autorizadas
- WHEN o cliente consulta o portfólio disponível
- THEN o sistema SHALL exibir somente as ofertas contratadas e as versões autorizadas para esse cliente

#### Scenario: Oferta publicada mas não contratada é negada mesmo com token válido
- GIVEN uma oferta no estado PUBLICADA no catálogo geral
- WHEN um cliente com token de autenticação válido, mas sem contrato para essa oferta, tenta acessá-la
- THEN o sistema SHALL negar o acesso à oferta

#### Scenario: Suspensão de novas vendas não afeta pedidos já aceitos
- GIVEN uma oferta com pedidos já aceitos em andamento
- WHEN a oferta é movida para o estado SUSPENSA para novas vendas
- THEN os pedidos já aceitos SHALL continuar sendo processados normalmente, sem cancelamento automático

### Requirement: CAT-03 — Agregação
O sistema SHALL permitir que um produto de agregação reúna resultados independentes em uma única resposta, sendo que cada seção da resposta SHALL identificar serviço, versão, estado, horário de observação e origem permitida pelo contrato. A regra de merge SHALL definir precedência, deduplicação, unidades e tratamento de valores conflitantes, e o sistema NÃO SHALL escolher arbitrariamente o primeiro retorno recebido. Quando uma parte declarada obrigatória do produto falhar, o sistema NÃO SHALL anunciar o produto como sucesso completo. Respostas opcionais ausentes SHALL ser explicitadas na resposta, e não omitidas silenciosamente.

#### Scenario: Agregação bem-sucedida com seções obrigatória e opcionais presentes
- GIVEN um produto de agregação cujo contrato declara situação cadastral obrigatória, endereço opcional e geolocalização opcional
- WHEN as três seções retornam com sucesso
- THEN a resposta consolidada SHALL identificar, para cada seção, serviço, versão, estado, horário de observação e origem, aplicando a regra de merge definida para precedência, deduplicação e unidades

#### Scenario: Falha da parte obrigatória impede anúncio de sucesso completo
- GIVEN um produto de agregação com situação cadastral declarada obrigatória
- WHEN a seção de situação cadastral falha
- THEN o sistema NÃO SHALL anunciar o produto como sucesso completo, ainda que outras seções opcionais tenham retornado

#### Scenario: Resposta opcional ausente é declarada explicitamente
- GIVEN um produto de agregação com geolocalização declarada opcional
- WHEN a seção de geolocalização não retorna resultado
- THEN a resposta SHALL indicar explicitamente a ausência dessa seção, em vez de omitir o campo silenciosamente

### Requirement: CAT-04 — Composição
O sistema SHALL permitir que um produto de composição encadeie serviços em um grafo acíclico, no qual um passo utiliza saídas tipadas de passos anteriores, descrevendo entradas/saídas por passo, transformações permitidas, condição de execução, dependências, obrigatoriedade, deadline, custo e política de falha. Um produto MAY combinar agregação e composição. Nenhum adaptador SHALL decidir uma regra de produto sem configuração publicada. Na v4, os passos SHALL referenciar apenas serviços, ficando fora do escopo a composição recursiva de produtos e workflows cíclicos. O sistema SHALL impor um limite inicial de até 20 passos e cinco passos simultâneos por protocolo, como tetos configuráveis após qualificação. O sistema SHALL impedir a publicação de um produto de composição quando houver ciclos, dependências inexistentes, tipos incompatíveis entre passos ou ausência de política de falha.

#### Scenario: Composição publicada com grafo acíclico válido
- GIVEN um produto de composição descrevendo os passos normalizar endereço, geocodificar e consultar cobertura, com validação cadastral em paralelo, dentro do limite de 20 passos e cinco passos simultâneos
- WHEN todas as entradas/saídas tipadas, dependências, deadline, custo e política de falha estão configurados e publicados
- THEN o produto pode ser publicado e executado conforme o grafo definido

#### Scenario: Geocodificação com múltiplas opções segue política do contrato
- GIVEN um passo de geocodificação que retorna múltiplas opções de resultado
- WHEN o contrato do produto define a política para esse caso
- THEN o sistema SHALL aplicar a seleção configurada, retornar a ambiguidade ao contratante ou tratar como falha funcional, conforme definido no contrato, sem decisão arbitrária do adaptador

#### Scenario: Publicação recusada por ciclo ou ausência de política de falha
- GIVEN um produto de composição cujo grafo contém um ciclo, uma dependência inexistente, tipos incompatíveis entre passos ou ausência de política de falha
- WHEN a publicação do produto é solicitada
- THEN o sistema SHALL impedir a publicação até que a inconsistência seja corrigida

### Requirement: CAT-05 — Consolidação e efeitos
O sistema SHALL suportar como políticas de consolidação: conclusão de todos os passos obrigatórios, parcialidade explicitamente aceita, ou quórum definido por quantidade e qualidade. Um quórum atingido NÃO SHALL cancelar automaticamente trabalho externo em voo. A regra de consolidação SHALL informar se o produto aguarda passos opcionais até o deadline ou finaliza e trata retornos posteriores como evidências tardias. Uma falha parcial NÃO SHALL gerar rollback universal. Cada serviço SHALL declarar seu efeito como somente leitura, reversível com compensação ou irreversível, sendo que uma compensação SHALL ser tratada como uma nova operação rastreada, com possíveis custos, falhas e prazo próprios. Um produto com efeito irreversível SHALL apresentar ao contratante a possibilidade de resultado parcial, e o encerramento do produto NÃO SHALL apagar operações ainda incertas.

#### Scenario: Consolidação por quórum sem cancelar trabalho em voo
- GIVEN um produto cuja política de consolidação é quórum definido por quantidade e qualidade
- WHEN o quórum é atingido enquanto ainda há chamadas externas em andamento
- THEN o sistema NÃO SHALL cancelar automaticamente essas chamadas em voo, e a regra SHALL indicar se aguarda os retornos até o deadline ou os trata como evidências tardias após finalizar

#### Scenario: Falha parcial não gera rollback universal
- GIVEN um produto com múltiplos passos, alguns já concluídos com efeito reversível ou irreversível
- WHEN um passo não obrigatório falha
- THEN o sistema NÃO SHALL reverter automaticamente todos os passos já concluídos, e a compensação, quando aplicável, SHALL ser registrada como nova operação rastreada

#### Scenario: Produto com efeito irreversível oferece opção de resultado parcial
- GIVEN um produto contendo um serviço com efeito irreversível
- WHEN a execução não pode ser concluída integralmente
- THEN o sistema SHALL apresentar ao contratante a possibilidade de resultado parcial, e o encerramento do protocolo NÃO SHALL apagar operações ainda incertas

### Requirement: CAT-06 — Seleção e equivalência de provedores
O sistema SHALL selecionar um provedor primeiro por elegibilidade — contrato vigente, conta autorizada, serviço/versão homologados, região, finalidade e dados permitidos, qualidade mínima, capacidade, quota e política de custo — e somente depois ordenar os elegíveis por prioridade, peso, custo ou latência conforme regra versionada. O sistema SHALL persistir a regra aplicada, os candidatos elegíveis, o provedor escolhido e a razão da escolha, de modo que a decisão possa ser explicada posteriormente. A homologação de equivalência entre provedores SHALL comparar significado dos campos, cobertura, atualização, precisão, evidências, efeitos e critérios de sucesso, preservando o hub seu contrato canônico e identificando limitações. Failover SHALL ocorrer somente com prova de segurança da nova operação; timeout após envio SHALL ser tratado como incerteza, não como autorização para failover. Hedging (duplicação especulativa de chamadas) SHALL permanecer desabilitado na versão inicial.

#### Scenario: Seleção de provedor elegível ordenado por regra versionada com decisão persistida
- GIVEN múltiplos provedores com contrato vigente, conta autorizada, serviço/versão homologados e dados/região compatíveis para uma operação
- WHEN o sistema aplica a regra de ordenação versionada (prioridade, peso, custo ou latência)
- THEN o sistema SHALL selecionar o provedor elegível de maior ordem e SHALL persistir a regra aplicada, os candidatos elegíveis, o escolhido e a razão da escolha

#### Scenario: Failover recusado sem prova de segurança após timeout
- GIVEN uma chamada enviada a um provedor que expira em timeout sem resposta confirmada
- WHEN o sistema avalia acionar failover para outro provedor
- THEN o sistema NÃO SHALL tratar o timeout como autorização automática para failover, exigindo prova de segurança da nova operação antes de prosseguir

#### Scenario: Hedging permanece desabilitado
- GIVEN uma operação elegível para múltiplos provedores equivalentes
- WHEN o sistema decide como enviar a chamada
- THEN o sistema NÃO SHALL duplicar especulativamente a chamada entre provedores (hedging), pois essa capacidade permanece desabilitada na versão inicial

### Requirement: CAT-07 — Política comercial do produto
Cada produto SHALL adotar exatamente uma entre as opções de preço de pacote, soma dos serviços, ou modelo híbrido com parcelas explicitamente identificadas. A configuração NÃO SHALL permitir cobrança simultânea de pacote e componentes por omissão. Franquia, medidor de sucesso, parcialidade e consequências de cancelamento SHALL pertencer ao contrato do produto. O sistema NÃO SHALL liberar uma oferta sem contrato válido para todos os passos necessários do produto.

#### Scenario: Pacote com três passos gera receita única e discrimina custos sem triplicar preço
- GIVEN um produto de pacote com três passos, cada um com custo elegível de provedor
- WHEN o produto é executado com sucesso
- THEN o sistema SHALL gerar a quantidade de receita configurada para o pacote e SHALL discriminar os três custos elegíveis, sem triplicar o preço de venda ao cliente

#### Scenario: Configuração impede cobrança simultânea de pacote e componentes
- GIVEN um produto configurado com preço de pacote
- WHEN a configuração comercial é validada
- THEN o sistema NÃO SHALL permitir, por omissão, que o mesmo produto cobre simultaneamente o preço de pacote e os componentes individuais

#### Scenario: Oferta não liberada sem contrato válido para todos os passos
- GIVEN um produto composto cujos passos exigem contratos de provedor distintos
- WHEN falta contrato válido para pelo menos um dos passos necessários
- THEN o sistema NÃO SHALL liberar a oferta para venda

### Requirement: CAT-08 — Rastreabilidade da composição
Um protocolo pai SHALL conter seus passos, dependências e resultados individuais, e cada interação externa SHALL referenciar o passo correspondente. Um passo reutilizado em dois ramos SHALL executar apenas uma vez somente se a versão do produto declarar compartilhamento e equivalência de entrada; o sistema NÃO SHALL deduplicar globalmente pedidos distintos por coincidência de payload, pois atualidade, autorização e cobrança podem diferir. O resultado consolidado SHALL registrar quais passos e versões contribuíram, as transformações aplicadas, as partes omitidas e o motivo da omissão.

#### Scenario: Resposta final e apuração financeira reconstruídas a partir de evidências persistidas
- GIVEN um protocolo pai concluído com múltiplos passos, transformações e omissões parciais
- WHEN uma reconstrução da resposta final e da apuração financeira é solicitada dentro do período de retenção aprovado
- THEN o sistema SHALL permitir reconstruir a resposta e a apuração a partir das evidências persistidas, incluindo passos, versões, transformações e motivos de omissão

#### Scenario: Passo compartilhado executa uma única vez quando declarado pelo produto
- GIVEN um passo referenciado por dois ramos do mesmo produto, com a versão do produto declarando compartilhamento e equivalência de entrada
- WHEN os dois ramos requerem esse passo
- THEN o sistema SHALL executar o passo uma única vez e reutilizar o resultado nos dois ramos

#### Scenario: Pedidos distintos com payload coincidente não são deduplicados globalmente
- GIVEN dois pedidos distintos cujo payload de entrada coincide, mas que diferem em atualidade, autorização ou cobrança
- WHEN o sistema processa esses pedidos
- THEN o sistema NÃO SHALL deduplicar globalmente os pedidos apenas por coincidência de payload

### Requirement: CAT-09 — Contratos técnicos por cliente
O sistema SHALL permitir que um mesmo serviço ou produto possua contratos técnicos distintos por cliente e aplicação, cobrindo entrada, saída, erros, nomes/tipos de campos, obrigatoriedade, enumerações, datas, unidades, encoding, media type e referências legadas, mantendo a identidade de domínio canônica e traduzindo o modelo legado apenas nas fronteiras. Cada vínculo cliente/oferta SHALL fixar customer_contract_id, versão de entrada, versão de saída, transformação de entrada, transformação de saída e versão da política de erros. A entrada do cliente SHALL ser validada antes da tradução, o resultado canônico SHALL também ser validado, e a saída customizada SHALL ser materializada e validada antes da conclusão do protocolo. As transformações SHALL ser determinísticas, sem rede, sem código arbitrário e com limites de CPU, memória, profundidade e tamanho; defaults SHALL ser permitidos apenas quando o contrato define seu significado, sem preencher documento, valor ou status fictício para satisfazer schema. O cliente SHALL receber exclusivamente sua projeção, sem poder selecionar tenant, contrato de outro cliente ou transformação por nome arbitrário no pedido. Uma alteração de configuração NÃO SHALL mudar protocolo em voo, GET histórico ou reenvio de webhook. Erro de transformação SHALL impedir sucesso publicável, SHALL conservar o resultado canônico como evidência e SHALL usar a representação de falha pré-homologada do cliente, dentro do deadline.

#### Scenario: Contrato técnico do cliente valida entrada, traduz e valida saída antes da conclusão
- GIVEN um vínculo cliente/oferta com customer_contract_id, versões de entrada/saída e transformações fixadas
- WHEN o cliente envia um pedido em seu formato customizado
- THEN o sistema SHALL validar a entrada do cliente antes da tradução, validar o resultado canônico, materializar e validar a saída customizada, e entregar exclusivamente a projeção desse cliente

#### Scenario: Erro de transformação impede sucesso e preserva evidência canônica
- GIVEN um pedido cujo resultado canônico foi produzido com sucesso, mas cuja transformação de saída falha
- WHEN o protocolo tenta concluir
- THEN o sistema NÃO SHALL publicar sucesso, SHALL conservar o resultado canônico como evidência e SHALL retornar a representação de falha pré-homologada do cliente dentro do deadline

#### Scenario: Cliente não pode selecionar tenant ou transformação arbitrária no pedido
- GIVEN um pedido de um cliente autenticado com contrato técnico próprio
- WHEN o pedido tenta especificar um tenant, contrato de outro cliente ou nome de transformação arbitrário
- THEN o sistema SHALL recusar essa seleção e aplicar somente a configuração publicada e versionada para esse cliente

#### Scenario: Alteração de configuração não afeta protocolo em voo nem histórico
- GIVEN uma nova versão de transformação publicada para um cliente
- WHEN existe um protocolo em voo, um GET histórico ou um reenvio de webhook referente à configuração anterior
- THEN o sistema NÃO SHALL aplicar a nova configuração a esses casos, preservando o comportamento vigente no momento original

### Requirement: CAT-10 — Elegibilidade de SLA e capacidade
Cada oferta e vínculo de cliente SHALL declarar prazo de resultado do hub, TTL de retry, política de término, SLA de provedor, margem interna de finalização e classe de isolamento. O sistema NÃO SHALL vender um prazo menor que o caminho crítico qualificado sem provedor e capacidade compatíveis, e o tempo em fila, reserva financeira, transformação e captura de arquivos SHALL integrar o prazo do hub. Um produto com passos paralelos SHALL possuir deadline absoluto no protocolo pai, com orçamento por passo que não o ultrapasse; um produto sequencial SHALL reservar tempo para sucessores e consolidação. Parcialidade MAY ser publicada antes do deadline apenas se contratada; o sistema NÃO SHALL converter automaticamente uma quebra de SLA em sucesso parcial. Em timeout, o corpo final SHALL informar a quebra no vocabulário do cliente e SHALL manter evidências de passos em voo. Uma classe compartilhada NÃO SHALL ser anunciada como isolamento físico total.

#### Scenario: Oferta com prazo compatível com caminho crítico qualificado
- GIVEN uma oferta cujo caminho crítico qualificado (incluindo fila, reserva financeira, transformação e captura de arquivos) exige um tempo mínimo X
- WHEN a oferta é configurada com prazo de resultado do hub
- THEN o sistema NÃO SHALL permitir a venda de um prazo menor que X sem provedor e capacidade compatíveis

#### Scenario: Timeout não é convertido automaticamente em sucesso parcial
- GIVEN um produto sem contratação de parcialidade explícita
- WHEN o deadline é atingido antes da conclusão de todos os passos obrigatórios
- THEN o sistema NÃO SHALL converter automaticamente essa quebra de SLA em sucesso parcial, e o corpo final SHALL informar a quebra no vocabulário do cliente mantendo evidências dos passos em voo

#### Scenario: Classe compartilhada não é anunciada como isolamento físico total
- GIVEN uma oferta configurada na classe de isolamento compartilhada
- WHEN a oferta é apresentada ao cliente
- THEN o sistema NÃO SHALL anunciar essa classe como isolamento físico total; um cliente que exigir proteção contra exaustão de recursos compartilhados SHALL ser direcionado à classe dedicada

### Requirement: CAT-11 — Crescimento, modalidade e elegibilidade de credencial
Adicionar cliente, aplicação e provedor SHALL ser uma operação do plano de controle, sem alteração de código ou chamado de infraestrutura, quando utilizar capacidade, protocolo e perfil já homologados. O catálogo SHALL publicar o envelope suportado: clientes ativos, vínculos, taxa e concorrência por classe, objetos, quantidade de passos, capacidade externa e orçamento. O sistema NÃO SHALL prometer capacidade física infinita nem suporte declarativo a um protocolo ainda não implementado. Cada oferta SHALL declarar separadamente a modalidade do provedor e as modalidades disponíveis ao cliente. A publicação de SYNC SHALL exigir provedor síncrono, resposta final previsível no orçamento de conexão, transformações e persistência qualificadas e capacidade reservada, inclusive em todo caminho obrigatório de um produto composto. ASYNC MAY envolver provedores síncronos ou assíncronos. AUTO SHALL ser uma terceira escolha explícita, nunca uma conversão silenciosa de SYNC. Para cada cliente/oferta/provedor, o contrato SHALL resolver SHARED_HUB ou TENANT_DEDICATED, conta externa e vínculo de credencial, e a disponibilidade desse vínculo SHALL integrar a elegibilidade antes do envio. O sistema NÃO SHALL expor ao cliente segredos do hub ou credenciais de outros clientes, nem permitir trocar de conta/credencial ignorando efeito já enviado ou obrigação econômica.

#### Scenario: Integração de novo cliente por configuração, sem alteração de código
- GIVEN um novo cliente a ser integrado a um serviço existente usando capacidade, protocolo e perfil já homologados
- WHEN o cliente é cadastrado no plano de controle com credencial, modalidade e plano validados
- THEN o sistema SHALL obter capacidade por automação, sem exigir alteração de código ou chamado de infraestrutura, e SHALL executar as duas modalidades previstas para esse cliente

#### Scenario: Ativação recusada explicitamente por falta de capacidade homologada
- GIVEN um novo cliente ou provedor cuja ativação requer capacidade ainda não homologada
- WHEN a ativação é solicitada
- THEN o sistema NÃO SHALL publicar a oferta como ficticiamente pronta nem afetar clientes já ativos, e o estado de ativação SHALL explicar a restrição de capacidade

#### Scenario: AUTO nunca é conversão silenciosa de SYNC
- GIVEN uma oferta cujo cliente selecionou a modalidade AUTO
- WHEN o sistema decide como atender o pedido
- THEN o sistema SHALL tratar AUTO como uma escolha explícita e distinta, sem convertê-la silenciosamente em SYNC

## Notas de origem

Este delta deriva integralmente do capítulo `docs/01_DOMINIO_E_PORTFOLIO.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos CAT-01 a CAT-11 conforme texto normativo de origem.
