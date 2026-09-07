# 01 · Domínio, serviços e portfólio

## Vocabulário e relações

**Tenant/cliente** é a unidade de autorização e faturamento. Uma organização pode ter contas e aplicações consumidoras subordinadas, com limites próprios. **Provedor** é a organização fornecedora; **conta de provedor** reúne vínculo contratual, ambiente, credenciais e quotas. Conta compartilhada entre clientes exige medição segregada e respeito ao limite global.

**Serviço canônico** define a capacidade oferecida pelo hub: entradas, saídas, semântica, validações, qualidade e efeitos possíveis. **Vínculo de integração** mapeia uma versão desse serviço para uma operação suportada por uma conta do provedor. Dois provedores com campos parecidos não são necessariamente equivalentes.

**Produto** é uma oferta versionada, com identidade comercial e contrato de resultado. Pode conter um serviço, agregar resultados independentes ou compor serviços dependentes. **Plano** define regras de cobrança, franquias e limites. **Contrato do cliente** vincula cliente, plano e ofertas autorizadas; **contrato do provedor** estabelece condições de aquisição. O preço de venda não deriva implicitamente do custo de compra.

## CAT-01 · Catálogo de serviços — proprietário: Atlas / Produto

Cada versão deve declarar código estável, descrição funcional, classificação dos dados, schema de entrada e resultado, validações, unidades, condições de resultado válido/sem resultado, efeitos externos, modos de atendimento do hub, deadline de negócio, retenção, elegibilidade comercial, SLA e política de resultado tardio. O catálogo deve distinguir erro técnico de uma resposta negativa válida: “não encontrado” pode ser sucesso funcional.

A publicação exige ao menos um vínculo de provedor homologado ou classificação explícita como serviço interno. Versões publicadas são imutáveis. Versão descontinuada continua interpretável para protocolos históricos; novas admissões cessam na data programada. Aceite: entrada incompatível é recusada antes de execução e consulta histórica preserva a semântica original.

## CAT-02 · Portfólio e disponibilidade — proprietário: Atlas / Produto

A oferta tem estados RASCUNHO, EM_VALIDACAO, PUBLICADA, SUSPENSA e DESCONTINUADA. Disponibilidade do catálogo, elegibilidade do contrato e saúde dos provedores são dimensões diferentes. Suspender novas vendas não cancela pedidos aceitos. O portal deve explicar indisponibilidade sem expor segredos, preços de aquisição ou outros tenants.

O cliente vê apenas ofertas contratadas e versões autorizadas. Datas de início/fim, região, canal, quotas e permissão para dados sensíveis participam da decisão de admissão. Aceite: uma oferta publicada mas não contratada é negada, mesmo com token válido.

## CAT-03 · Agregação — proprietário: Órbita / Produto

Agregação reúne resultados independentes em uma resposta de produto. Cada seção deve identificar serviço, versão, estado, horário de observação e origem permitida pelo contrato. A regra de merge deve definir precedência, deduplicação, unidades e tratamento de valores conflitantes; não escolher arbitrariamente o primeiro retorno.

Exemplo ilustrativo: produto de análise cadastral reúne situação cadastral, endereço e geolocalização em paralelo. O contrato declara situação cadastral obrigatória, endereço opcional e geolocalização opcional. Se a parte obrigatória falhar, o produto não será anunciado como sucesso completo. Respostas opcionais ausentes são explícitas, não campos silenciosamente omitidos.

## CAT-04 · Composição — proprietário: Órbita / Produto

Composição encadeia serviços: um passo utiliza saídas tipadas de passos anteriores. O produto deve descrever grafo acíclico, entradas/saídas por passo, transformações permitidas, condição de execução, dependências, obrigatoriedade, deadline, custo e política de falha. Produto pode combinar agregação e composição.

Exemplo: normalizar endereço → geocodificar → consultar cobertura; validação cadastral pode ocorrer paralelamente. Se geocodificação retornar múltiplas opções, o contrato deve determinar seleção, retorno de ambiguidade ou falha funcional. Nenhum adaptador decide uma regra de produto sem configuração publicada.

Limites iniciais propostos: até 20 passos e cinco passos simultâneos por protocolo; valores são tetos configuráveis após qualificação. Na v4, passos referenciam serviços; composição recursiva de produtos e workflows cíclicos ficam fora do escopo. Aceite: ciclos, dependências inexistentes, tipos incompatíveis ou ausência de política de falha impedem publicação.

## CAT-05 · Consolidação e efeitos — proprietário: Órbita / Produto

Políticas permitidas: todos os passos obrigatórios concluídos; parcialidade explicitamente aceita; ou quórum definido por quantidade e qualidade. Quórum não cancela automaticamente trabalho externo em voo. A regra informa se aguarda opcionais até deadline ou finaliza e trata os retornos posteriores como evidências tardias.

Uma falha parcial não gera rollback universal. Cada serviço declara efeito somente leitura, reversível com compensação ou irreversível. Compensação é uma nova operação rastreada, com possíveis custos, falhas e prazo. Produto com efeito irreversível deve apresentar ao contratante a possibilidade de resultado parcial. O encerramento não apaga operações ainda incertas.

## CAT-06 · Seleção e equivalência de provedores — proprietário: Órbita / Integrações

Selecionar primeiro por elegibilidade: contrato vigente, conta autorizada, serviço/versão homologados, região, finalidade e dados permitidos, qualidade mínima, capacidade, quota e política de custo. Ordenar depois por prioridade, peso, custo ou latência conforme regra versionada. Persistir regra, candidatos elegíveis, escolhido e razão; a decisão deve poder ser explicada posteriormente.

A homologação de equivalência deve comparar significado dos campos, cobertura, atualização, precisão, evidências, efeitos e critérios de sucesso. O hub preserva seu contrato canônico e identifica limitações. Failover só ocorre com prova de segurança da nova operação; timeout após envio é incerteza, não autorização. Hedging (duplicação especulativa de chamadas) fica desabilitado na versão inicial, pois pode gerar custos e efeitos duplicados.

## CAT-07 · Política comercial do produto — proprietário: Atlas / Comercial

Cada produto deve optar por preço de pacote, soma dos serviços ou modelo híbrido com parcelas explicitamente identificadas. A configuração não pode permitir cobrança simultânea de pacote e componentes por omissão. Franquia, medidor de sucesso, parcialidade e consequências de cancelamento pertencem ao contrato do produto.

A receita pode ser uma unidade de produto enquanto o custo contém várias operações de provedores. Sem contrato válido para todos os passos necessários, a oferta não é liberada. Aceite: um pacote com três passos gera a quantidade de receita configurada e discrimina os três custos elegíveis, sem triplicar o preço de venda.

## CAT-08 · Rastreabilidade da composição — proprietário: Órbita / QA

Um protocolo pai contém passos, dependências e resultados individuais; cada interação externa referencia seu passo. Passo reutilizado em dois ramos só pode executar uma vez se a versão do produto declarar compartilhamento e equivalência de entrada. Não deduplicar globalmente pedidos distintos por coincidência de payload: atualidade, autorização e cobrança podem diferir.

O resultado consolidado registra quais passos e versões contribuíram, transformações aplicadas, partes omitidas e motivo. Aceite: a resposta final e a apuração financeira podem ser reconstruídas a partir de evidências persistidas, dentro da retenção aprovada.

## CAT-09 · Contratos técnicos por cliente — proprietário: Atlas / Produto

Um mesmo serviço ou produto pode possuir contratos técnicos distintos por cliente e aplicação: entrada, saída, erros, nomes/tipos de campos, obrigatoriedade, enumerações, datas, unidades, encoding, media type e referências legadas. Contrato técnico não é tabela de preços. A identidade de domínio continua canônica; o modelo legado é traduzido nas fronteiras.

Cada vínculo cliente/oferta fixa customer_contract_id, versão de entrada, versão de saída, transformação de entrada, transformação de saída e versão da política de erros. A entrada do cliente é validada antes da tradução e o resultado canônico também é validado. Saída customizada é materializada e validada antes da conclusão do protocolo. Não basta renomear campos se unidade, significado ou nulabilidade diferem.

São capacidades iniciais: REST com JSON e projeções declarativas tipadas; XML/SOAP exige perfil de adaptação homologado conforme COM-02. Transformações são determinísticas, sem rede, sem código arbitrário e com limites de CPU, memória, profundidade e tamanho. Defaults só são permitidos quando o contrato define seu significado; não preencher documento, valor ou status fictício para satisfazer schema.

O cliente recebe exclusivamente sua projeção; não pode selecionar tenant, contrato de outro cliente ou transformação por nome arbitrário no pedido. Toda configuração é publicada e versionada. Alteração não muda protocolo em voo, GET histórico ou reenvio de webhook. Erro de transformação impede sucesso publicável, conserva resultado canônico como evidência e usa a representação de falha pré-homologada do cliente, dentro do deadline. Consultas e webhooks seguem COM-05.

## CAT-10 · Elegibilidade de SLA e capacidade — proprietário: Produto / Arquitetura

Cada oferta e vínculo de cliente devem declarar prazo de resultado do hub, TTL de retry, política de término, SLA de provedor, margem interna de finalização e classe de isolamento. Não vender um prazo menor que o caminho crítico qualificado sem provedor e capacidade compatíveis. Tempo em fila, reserva financeira, transformação e captura de arquivos integra o prazo do hub.

Produto com passos paralelos possui deadline absoluto no protocolo pai; cada passo recebe orçamento que não o ultrapassa. Produto sequencial reserva tempo para sucessores e consolidação. Parcialidade pode ser publicada antes do deadline apenas se contratada; não converter automaticamente uma quebra de SLA em sucesso parcial. Em timeout, o corpo final informa a quebra no vocabulário do cliente e mantém evidências de passos em voo.

Quotas de capacidade contratada e faixas de compartilhamento são parte da oferta. Uma classe compartilhada não deve ser anunciada como isolamento físico total. Cliente exigindo proteção contra exaustão de recursos compartilhados precisa da classe dedicada descrita em OPE-08, inclusive para conta/capacidade de provedor quando necessário.

## CAT-11 · Crescimento, modalidade e elegibilidade de credencial — proprietário: Produto / Atlas

Adicionar cliente, aplicação e provedor deve ser operação do plano de controle, sem alteração de código ou chamado de infraestrutura quando utilizar capacidade, protocolo e perfil já homologados. O catálogo publica o envelope suportado: clientes ativos, vínculos, taxa e concorrência por classe, objetos, quantidade de passos, capacidade externa e orçamento. O crescimento do número de clientes/provedores amplia células e recursos por automação, mantendo os mesmos componentes e contratos internos. Não se promete capacidade física infinita nem suporte declarativo a um protocolo ainda não implementado.

Cada oferta declara separadamente a modalidade do provedor e as modalidades disponíveis ao cliente. A publicação de SYNC exige provedor síncrono, resposta final previsível no orçamento de conexão, transformações e persistência qualificadas e capacidade reservada; produto composto exige isso em todo caminho obrigatório. ASYNC pode envolver provedores síncronos ou assíncronos. AUTO é uma terceira escolha explícita, nunca uma conversão silenciosa de SYNC.

Para cada cliente/oferta/provedor, o contrato resolve SHARED_HUB ou TENANT_DEDICATED, conta externa e vínculo de credencial. A disponibilidade desse vínculo integra a elegibilidade antes de enviar. Não expor ao cliente segredos do Hub ou credenciais de outros clientes. Contrato pode autorizar mais de um provedor equivalente, mas não permite trocar de conta/credencial ignorando efeito já enviado ou obrigação econômica.

Aceite: integrar um novo cliente a serviço existente por configuração, validar sua credencial/modalidade/plano, obter capacidade por automação e executar as duas modalidades previstas. Se faltar capacidade homologada, o estado de ativação deve explicar a restrição; não publicar oferta ficticiamente pronta nem afetar clientes já ativos.
