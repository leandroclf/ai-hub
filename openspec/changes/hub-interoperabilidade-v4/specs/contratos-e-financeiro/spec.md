# Contratos, Consumo e Financeiro — Delta de Especificação

## ADDED Requirements

### Requirement: FIN-01 — Contrato de aquisição
Cada conta de provedor SHALL estar vinculada a um contrato de compra versionado contendo contraparte, identificador externo, vigência, ambiente, serviços/operações, moeda, unidades, tabela de preços, descontos, franquias, mínimos, teto, quota, concorrência, prazo de resposta, condições de armazenamento e formato de extrato/fatura. O contrato SHALL especificar os fatos tarifáveis (submit aceito, execução final, consulta de status, download de resultado, cancelamento, compensação ou outro medidor claramente definido) e SHALL declarar se timeout, resultado negativo, erro técnico, resposta tardia e retransmissão são cobrados. Ausência de tabela de preços ou interpretação ambígua de um fato tarifável SHALL bloquear a publicação comercial e NÃO SHALL ser tratada como custo zero. Contrato de aquisição vencido SHALL bloquear novas operações conforme política definida, mas operações já aceitas e consultas necessárias à conclusão SHALL seguir regra de continuidade previamente acordada; credencial bloqueada SHALL ser tratada como impedimento técnico, distinto da condição comercial de contrato vencido. A plataforma SHALL registrar qual versão de aquisição se aplica a cada operação e a regra temporal para seus polls/fetches posteriores.

#### Scenario: Conta de provedor vinculada a contrato de compra completo
- GIVEN uma nova conta de provedor está sendo cadastrada
- WHEN o contrato de compra é vinculado a essa conta
- THEN o sistema SHALL exigir contraparte, identificador externo, vigência, ambiente, serviços/operações, moeda, unidades, tabela de preços, descontos, franquias, mínimos, teto, quota, concorrência, prazo de resposta, condições de armazenamento e formato de extrato/fatura antes de habilitar a conta

#### Scenario: Tabela ausente ou ambígua bloqueia publicação, não é custo zero
- GIVEN um contrato de aquisição sem tabela de preços para um fato tarifável, ou com interpretação ambígua desse fato
- WHEN alguém tenta publicar esse contrato comercialmente
- THEN o sistema SHALL bloquear a publicação comercial e NÃO SHALL tratar a ausência ou ambiguidade como custo zero

#### Scenario: Contrato vencido bloqueia novas operações mas respeita continuidade acordada
- GIVEN um contrato de aquisição vencido com operações já aceitas que ainda exigem consulta para conclusão
- WHEN o sistema avalia se pode iniciar uma nova operação ou continuar uma já aceita
- THEN o sistema SHALL bloquear novas operações conforme a política vigente, SHALL permitir a continuidade das operações já aceitas conforme regra previamente acordada, e SHALL distinguir esse bloqueio comercial de um bloqueio por credencial tecnicamente impedida

#### Scenario: Versão de aquisição registrada por operação orienta polls posteriores
- GIVEN uma operação executada sob uma versão específica do contrato de aquisição
- WHEN consultas de status (polls/fetches) subsequentes são realizadas para essa operação
- THEN o sistema SHALL registrar qual versão de aquisição se aplica à operação e SHALL aplicar a regra temporal correspondente a essas consultas posteriores

### Requirement: FIN-02 — Contrato de venda e planos
O contrato do cliente SHALL informar tenant, aplicações, ofertas/versões autorizadas, plano, vigência, moeda, periodicidade, fuso de fechamento, unidade, marco de cobrança, preços, descontos, franquias, excedentes, mínimos, limites e forma de integração financeira, e o plano SHALL declarar o custo de resultado parcial, sem resultado, falha, cancelamento, revisão e reexecução. O sistema SHALL suportar os modelos preço unitário por unidade, assinatura com franquia e excedente, faixas progressivas marginais, preço por faixa de volume total, pacote de produto e híbrido explicitamente decomposto, e a configuração SHALL diferenciar "progressivo marginal" (preço aplicado apenas à parcela na faixa) de "volume total" (faixa atingida aplicada a todo o volume elegível), definindo os ajustes correspondentes no fechamento. A franquia SHALL declarar unidade, período, escopo, compartilhamento entre aplicações, ordem de consumo, reset, rollover e devolução por estorno, com default de sem rollover e sem devolução automática fora do período salvo exceção contratual; quota de tráfego NÃO SHALL ser tratada como franquia financeira.

#### Scenario: Contrato de venda com todos os campos exigidos
- GIVEN um novo contrato de cliente está sendo configurado
- WHEN o plano é vinculado ao tenant
- THEN o sistema SHALL exigir tenant, aplicações, ofertas/versões autorizadas, plano, vigência, moeda, periodicidade, fuso de fechamento, unidade, marco de cobrança, preços, descontos, franquias, excedentes, mínimos, limites e forma de integração financeira, além da declaração de custo para resultado parcial, sem resultado, falha, cancelamento, revisão e reexecução

#### Scenario: Configuração diferencia faixa progressiva marginal de faixa por volume total
- GIVEN um plano de venda baseado em faixas de consumo
- WHEN a configuração define o modelo de precificação por faixa
- THEN o sistema SHALL diferenciar explicitamente "progressivo marginal" (preço aplicado somente à parcela dentro de cada faixa) de "volume total" (faixa atingida aplicada a todo o volume elegível), e SHALL definir os ajustes de fechamento correspondentes ao modelo escolhido

#### Scenario: Franquia sem rollover e sem devolução automática por default
- GIVEN uma franquia contratual sem exceção explícita registrada
- WHEN o período de apuração da franquia se encerra
- THEN o sistema SHALL aplicar o default de sem rollover de saldo não utilizado e sem devolução automática fora do período, permitindo exceção apenas quando prevista contratualmente

#### Scenario: Quota de tráfego não é confundida com franquia financeira
- GIVEN uma quota técnica de tráfego configurada para um tenant
- WHEN o sistema apura consumo para fins de franquia financeira
- THEN o sistema NÃO SHALL tratar essa quota de tráfego como franquia financeira nem usá-la como substituto da franquia declarada no plano

### Requirement: FIN-03 — Vigência e snapshot
O sistema SHALL fixar no protocolo o contrato/plano de venda e a versão da oferta, SHALL fixar na operação externa o contrato/conta de compra, e SHALL persistir o preço/regra efetivamente aplicados, o marco de incidência, o instante de ocorrência e o instante de recebimento do fato, com o fuso contratual definindo a competência e relógios/registros internos em UTC. Uma alteração comum de tabela SHALL valer apenas para novas admissões e NÃO SHALL reprifica protocolos em voo, exceto quando o próprio contrato registrar exceção temporal explícita (por exemplo, provedor que cobra polls pela tabela vigente no instante de cada consulta), caso em que cada consulta SHALL referenciar a tabela aplicável naquele instante. Contratos sobrepostos SHALL exigir prioridade explícita e NÃO SHALL deixar duas regras elegíveis sem resolução. Aditivo retroativo SHALL exigir autorização, escopo e lançamentos de ajuste que preservem o cálculo original, e revogação de segurança MAY interromper operações em voo, devendo gerar evidência, sem reescrever o contrato nem o custo já incorrido.

#### Scenario: Snapshot fixa contrato, versão da oferta e instantes do fato
- GIVEN um protocolo é admitido e uma operação externa é despachada
- WHEN o sistema registra o fato tarifável correspondente
- THEN o sistema SHALL persistir o contrato/plano de venda e a versão da oferta fixados no protocolo, o contrato/conta de compra fixados na operação externa, o preço/regra efetivamente aplicados, o marco de incidência, o instante de ocorrência e o instante de recebimento do fato, com competência definida pelo fuso contratual e registros internos em UTC

#### Scenario: Alteração de tabela não reprifica protocolos em voo
- GIVEN uma tabela de preços é alterada durante a vigência de protocolos já admitidos
- WHEN a nova tabela entra em vigor
- THEN o sistema SHALL aplicar a nova tabela apenas a novas admissões e NÃO SHALL reprifica os protocolos já em voo sob a tabela anterior

#### Scenario: Exceção temporal para polls cobrados pela tabela vigente no instante da consulta
- GIVEN um contrato de aquisição declara que o provedor cobra cada poll pela tabela vigente no instante da consulta
- WHEN uma consulta de status é realizada durante a vigência do protocolo
- THEN o sistema SHALL tratar essa cobrança como exceção temporal registrada na versão fixada e SHALL fazer cada consulta referenciar a tabela aplicável naquele instante específico

#### Scenario: Contratos sobrepostos exigem prioridade explícita
- GIVEN dois contratos elegíveis se sobrepõem para a mesma operação
- WHEN o sistema precisa decidir qual contrato aplicar
- THEN o sistema SHALL exigir prioridade explícita entre os contratos e NÃO SHALL deixar as duas regras elegíveis sem resolução

#### Scenario: Aditivo retroativo preserva cálculo original com lançamento de ajuste
- GIVEN um aditivo retroativo é aprovado para um contrato já vigente
- WHEN o aditivo é aplicado
- THEN o sistema SHALL exigir autorização e escopo definidos, SHALL gerar lançamentos de ajuste vinculados, e SHALL preservar o cálculo original já registrado, sem sobrescrevê-lo

#### Scenario: Revogação de segurança gera evidência sem reescrever contrato ou custo incorrido
- GIVEN uma revogação de segurança é acionada com operações em voo
- WHEN a revogação interrompe essas operações
- THEN o sistema SHALL gerar evidência da interrupção e NÃO SHALL reescrever o contrato nem o custo já incorrido pelas operações interrompidas

### Requirement: FIN-04 — Medição e chaves econômicas
O sistema SHALL distinguir telemetria técnica (chamadas, latência, falhas) de medição econômica (identificação da unidade tarifável), com Libra recebendo fatos duráveis de Cometa e Órbita, reconhecendo a chave semântica e aplicando o contrato correspondente; logs técnicos NÃO SHALL ser tratados como fonte autoritativa de faturamento. O sistema SHALL aplicar as chaves econômicas mínimas por natureza: receita por produto executado (cliente/contrato/protocolo/medidor de produto, reconhecendo uma unidade elegível apesar de múltiplos passos), receita por componente (cliente/contrato/protocolo/passo/medidor, somente se a política do produto cobrar componentes), custo por operação (provedor/conta/contrato/operation_id/medidor, preservando identidade entre retransmissões), custo por consulta de status (operação/interação física tarifável/medidor, com evidência própria por consulta cobrada), ajuste/estorno (lançamento original/tipo/identidade do ajuste autorizado, sem duplicar por reenvio) e assinatura/mínimo (contrato/período/parcela recorrente, uma parcela por competência). Quando o provedor cobra cada chamada de transporte, inclusive retransmissão, a chave de custo SHALL incluir tentativa e evidência contratual, pois "idempotente" NÃO SHALL ser interpretado como "gratuito". Quando o custo é incerto, o sistema SHALL registrar provisão/pendência identificável e NÃO SHALL registrar zero definitivo nem débito confirmado sem regra ou evidência.

#### Scenario: Libra aplica contrato a partir de fatos duráveis, não de logs técnicos
- GIVEN Cometa ou Órbita emitem um fato durável de execução
- WHEN Libra processa esse fato para fins de medição econômica
- THEN o sistema SHALL basear o cálculo no fato durável e na chave semântica reconhecida, aplicando o contrato correspondente, e NÃO SHALL usar logs técnicos como fonte autoritativa de faturamento

#### Scenario: Receita por produto é uma unidade elegível apesar de múltiplos passos
- GIVEN um produto é executado através de múltiplos passos internos
- WHEN o sistema apura a receita desse produto
- THEN o sistema SHALL reconhecer uma única unidade elegível pela chave cliente/contrato/protocolo/medidor de produto, independentemente do número de passos internos executados

#### Scenario: Custo por operação preserva identidade entre retransmissões
- GIVEN uma operação externa é retransmitida ao provedor por falha de transporte
- WHEN o provedor cobra cada chamada de transporte, incluindo a retransmissão
- THEN o sistema SHALL incluir a tentativa e a evidência contratual na chave de custo por operação (provedor/conta/contrato/operation_id/medidor), preservando a identidade entre as tentativas, sem tratar a idempotência da operação como isenção de custo

#### Scenario: Custo incerto é registrado como provisão, nunca como zero definitivo
- GIVEN o custo de uma operação ainda não pode ser confirmado com regra ou evidência suficiente
- WHEN o sistema apura essa operação
- THEN o sistema SHALL registrar uma provisão ou pendência identificável e NÃO SHALL registrar zero definitivo nem débito confirmado sem regra ou evidência

### Requirement: FIN-05 — Matriz de incidência
O sistema SHALL aplicar a matriz de incidência publicada no contrato para cada situação de receita do cliente e custo do provedor, e um contrato que omitir uma linha aplicável dessa matriz NÃO SHALL ser habilitado para faturamento automático; os defaults propostos SHALL aparecer explicitamente no contrato publicado, e o sistema NÃO SHALL operar com regras invisíveis em runtime.

#### Scenario: Pedido recusado antes do aceite não gera receita nem custo
- GIVEN um pedido é recusado antes do aceite
- WHEN o sistema apura receita e custo dessa tentativa
- THEN o sistema SHALL registrar ausência de consumo do serviço para o cliente e ausência de operação externa para o provedor

#### Scenario: Pedido aceito ainda pendente segue o marco contratado sem presumir final
- GIVEN um pedido foi aceito e permanece pendente de conclusão
- WHEN o sistema apura receita e custo dessa situação
- THEN o sistema SHALL reconhecer receita conforme o marco contratado, sem presumir resultado final, e SHALL reconhecer que pode haver submit tarifável do lado do custo

#### Scenario: Sucesso válido gera receita pelo medidor previsto e custo comprovado
- GIVEN uma execução termina em sucesso válido
- WHEN o sistema apura receita e custo
- THEN o sistema SHALL reconhecer receita pelo medidor previsto no plano e custo pelas operações/consultas tarifáveis comprovadas

#### Scenario: "Não encontrado" válido segue a semântica do serviço mesmo sem dados
- GIVEN o provedor retorna um "não encontrado" válido segundo a semântica do serviço
- WHEN o sistema apura receita e custo
- THEN o sistema SHALL reconhecer receita conforme a semântica do serviço/plano e custo conforme a aquisição contratada, mesmo na ausência de dados retornados

#### Scenario: Falha técnica conhecida não gera receita de sucesso por default, mas pode gerar custo
- GIVEN ocorre uma falha técnica conhecida durante a execução
- WHEN o sistema apura receita e custo, na ausência de política explícita distinta
- THEN o sistema SHALL aplicar o default proposto de sem receita por sucesso, e SHALL reconhecer que pode haver custo de tentativa ou de aceitação do lado do provedor

#### Scenario: Resultado parcial de produto usa preço de parcialidade publicado
- GIVEN um produto retorna resultado parcial
- WHEN o sistema apura receita e custo
- THEN o sistema SHALL aplicar o preço/regra de parcialidade publicada como receita e SHALL reconhecer como custo apenas os passos efetivamente tarifáveis executados

#### Scenario: Falha ou reentrega de webhook não altera receita nem reexecuta o provedor
- GIVEN um webhook falha ou é reentregue ao cliente
- WHEN o sistema apura receita e custo dessa execução
- THEN o sistema NÃO SHALL alterar a receita já reconhecida pela execução apenas por essa falha/reentrega, e NÃO SHALL reexecutar o provedor por causa dela

#### Scenario: GET de resultado armazenado não gera nova receita nem nova chamada externa
- GIVEN o cliente solicita um GET de um resultado já armazenado
- WHEN o sistema processa essa requisição sob o plano padrão v4
- THEN o sistema NÃO SHALL reconhecer nova receita para esse GET e NÃO SHALL realizar nenhuma nova chamada externa ao provedor

#### Scenario: Callback e polling do mesmo final geram uma única unidade de receita
- GIVEN um callback e um polling reportam o mesmo resultado final de uma operação
- WHEN o sistema apura receita e custo
- THEN o sistema SHALL reconhecer uma única unidade de receita e uma operação única, admitindo que o polling em voo pode ter gerado custo real antes da chegada do callback

#### Scenario: Failover autorizado após incerteza contabiliza ambos os custos comprovados
- GIVEN um failover autorizado ocorre após incerteza, com o provedor A e o provedor B ambos comprovadamente acionados
- WHEN o sistema apura receita e custo dessa entrega
- THEN o sistema SHALL reconhecer receita conforme uma entrega de produto ou regra explícita, e SHALL contabilizar os custos de A e de B quando ambos forem comprovados, sem ocultar o custo do provedor que não venceu

#### Scenario: Cancelamento ou compensação segue regra de estorno explícita sem eliminar custo original
- GIVEN uma operação é cancelada ou compensada
- WHEN o sistema apura receita e custo
- THEN o sistema SHALL aplicar a regra de estorno explícita à receita, e SHALL reconhecer que a operação original pode continuar cobrada e que a compensação pode ter custo próprio

#### Scenario: Resultado corrigido gera ajuste rastreável sem assumir cancelamento do custo original
- GIVEN um resultado já apurado é corrigido posteriormente
- WHEN o sistema apura receita e custo dessa correção
- THEN o sistema SHALL gerar ajuste rastreável quando necessário, NÃO SHALL reabrir um protocolo EXPIRED por SLA por causa dessa correção, e NÃO SHALL assumir cancelamento automático do custo original

#### Scenario: Final tardio após EXPIRED/SLA_EXCEEDED não gera receita de sucesso
- GIVEN um protocolo expira por SLA e o provedor entrega um final tardio depois disso
- WHEN o sistema apura receita e custo
- THEN o sistema NÃO SHALL reconhecer receita pelo medidor de sucesso, SHALL aplicar crédito/estorno de aceite conforme o contrato quando cabível, e SHALL registrar o custo real ou abrir contestação, sem eliminação automática desse custo

#### Scenario: Contrato omitindo linha aplicável não é habilitado para faturamento automático
- GIVEN um contrato de venda ou compra não cobre uma linha aplicável da matriz de incidência
- WHEN esse contrato é avaliado para faturamento automático
- THEN o sistema NÃO SHALL habilitar o faturamento automático desse contrato até que a linha aplicável seja explicitada, e os defaults propostos SHALL constar do contrato publicado, sem regras invisíveis em runtime

### Requirement: FIN-06 — Reserva de saldo e limites
O sistema SHALL separar rate limit técnico, quota operacional, franquia de apuração e saldo/teto financeiro estrito, reconhecendo que somente os dois últimos exigem semântica econômica; uma franquia usada apenas para cálculo posterior MAY ser contabilizada por eventos, mas a promessa de limite estrito SHALL exigir reserva atômica na autoridade financeira antes de qualquer efeito externo. Para contrato de teto estrito, Órbita SHALL registrar pedido e intenção e SHALL obter de Libra uma reserva idempotente ligada ao protocolo, podendo o protocolo permanecer ACCEPTED aguardando autorização, mas nenhum passo externo SHALL ser liberado antes da reserva confirmada; Libra indisponível SHALL manter o estado financeiro pendente até o prazo de admissão financeira, e uma negação definitiva SHALL exigir falha conhecida, não timeout incerto. A reserva SHALL cobrir o máximo autorizável do produto, incluindo passos opcionais e consultas cobradas dentro do orçamento, e se o custo máximo não puder ser limitado o sistema NÃO SHALL oferecer esse contrato como teto estrito; a captura SHALL aplicar o consumo efetivo e SHALL liberar o excedente comprovado, a expiração de reserva NÃO SHALL liberar saldo de operação externa ainda em voo antes de reconciliar, e uma consulta idempotente SHALL resolver resposta de reserva perdida. Contratos pós-pagos sem teto estrito MAY continuar processando durante atraso de apuração, dentro do limite de risco aprovado.

#### Scenario: Reserva atômica obrigatória antes de qualquer efeito externo em teto estrito
- GIVEN um contrato de teto estrito recebe um novo pedido
- WHEN Órbita registra a intenção e solicita a reserva a Libra
- THEN o sistema SHALL obter uma reserva idempotente confirmada ligada ao protocolo antes de liberar qualquer passo externo, mesmo que o protocolo permaneça ACCEPTED aguardando essa confirmação

#### Scenario: Libra indisponível mantém pendência até o prazo, negação exige falha conhecida
- GIVEN Libra está indisponível no momento em que uma reserva de teto estrito é solicitada
- WHEN o sistema aguarda a resposta dentro do prazo de admissão financeira
- THEN o sistema SHALL manter o estado financeiro como pendente até esse prazo, e SHALL negar definitivamente a reserva apenas mediante falha conhecida, NÃO SHALL negar por timeout incerto

#### Scenario: Reserva cobre o máximo autorizável ou o contrato não é oferecido como teto estrito
- GIVEN um produto tem passos opcionais e consultas cobertas por orçamento
- WHEN o sistema calcula o valor da reserva de teto estrito
- THEN o sistema SHALL dimensionar a reserva para cobrir o máximo autorizável do produto, incluindo passos opcionais e consultas cobradas, e NÃO SHALL oferecer esse contrato como teto estrito se o custo máximo não puder ser limitado

#### Scenario: Expiração de reserva não libera saldo de operação em voo sem reconciliar
- GIVEN uma reserva expira enquanto a operação externa correspondente ainda está em voo
- WHEN o sistema processa essa expiração
- THEN o sistema NÃO SHALL liberar o saldo reservado antes de reconciliar o estado real da operação externa

#### Scenario: Pós-pago sem teto estrito continua processando dentro do limite de risco aprovado
- GIVEN um contrato pós-pago sem teto estrito enfrenta atraso na apuração
- WHEN novos pedidos chegam durante esse atraso
- THEN o sistema MAY continuar processando esses pedidos, desde que dentro do limite de risco previamente aprovado

### Requirement: FIN-07 — Apuração e ledger
O sistema SHALL manter fatos de uso imutáveis, cálculo versionado e ledger gerencial de partidas balanceadas por moeda, registrando receita a receber e obrigação a pagar como lançamentos distintos, nunca como uma única diferença líquida; conta de passagem, provisão, desconto e estorno SHALL ter natureza definida no plano de contas gerencial aprovado por Financeiro, e esse ledger SHALL suportar integração contábil sem substituir automaticamente a contabilidade oficial. O sistema NÃO SHALL usar ponto flutuante monetário, SHALL declarar regra de arredondamento com precisão da tarifa, escala da quantidade, momento do arredondamento e moeda, e totais exibidos e exportados SHALL reproduzir o mesmo cálculo; o sistema NÃO SHALL somar moedas distintas sem câmbio com fonte, data, taxa e regra contratual explícitas, e tributos SHALL ser tratados como campos e integrações de responsabilidade definida, sem motor fiscal presumido. Rascunhos MAY ser recalculados com versões, mas um lançamento publicado SHALL ser imutável, e correção SHALL ocorrer por lançamento inverso ou ajuste vinculado; o sistema SHALL garantir que débitos e créditos do lote balanceiem por moeda, que toda linha tenha chave econômica/origem, e que nenhum fechamento deixe fato elegível silenciosamente não apurado.

#### Scenario: Receita a receber e obrigação a pagar são lançamentos distintos
- GIVEN uma operação gera simultaneamente receita do cliente e custo do provedor
- WHEN o sistema registra esses lançamentos no ledger
- THEN o sistema SHALL registrar a receita a receber e a obrigação a pagar como lançamentos distintos, NÃO SHALL colapsá-los em uma única diferença líquida

#### Scenario: Arredondamento declarado e totais reproduzem o mesmo cálculo sem ponto flutuante monetário
- GIVEN um cálculo monetário precisa ser exibido e exportado
- WHEN o sistema aplica a regra de arredondamento declarada (precisão da tarifa, escala da quantidade, momento do arredondamento, moeda)
- THEN o sistema NÃO SHALL usar ponto flutuante monetário, e os totais exibidos e exportados SHALL reproduzir exatamente o mesmo cálculo

#### Scenario: Moedas distintas não são somadas sem câmbio explícito
- GIVEN valores em moedas diferentes precisam ser consolidados
- WHEN o sistema processa essa consolidação
- THEN o sistema NÃO SHALL somar moedas distintas sem uma conversão de câmbio com fonte, data, taxa e regra contratual explícitas

#### Scenario: Lançamento publicado é imutável, correção usa lançamento inverso ou ajuste vinculado
- GIVEN um lançamento já foi publicado no ledger
- WHEN uma correção nesse lançamento é necessária
- THEN o sistema NÃO SHALL alterar o lançamento publicado diretamente, e SHALL registrar a correção como lançamento inverso ou ajuste vinculado, mantendo rascunhos recalculáveis com versões apenas antes da publicação

#### Scenario: Fechamento não deixa fato elegível silenciosamente não apurado
- GIVEN um período de apuração está sendo fechado
- WHEN o sistema verifica o invariante de balanceamento do lote
- THEN o sistema SHALL garantir que débitos e créditos balanceiem por moeda, que toda linha tenha chave econômica/origem, e NÃO SHALL permitir que algum fato elegível permaneça silenciosamente não apurado

### Requirement: FIN-08 — Fechamento, conciliação e pagamento
O ciclo de venda SHALL seguir os estados ABERTO → PRE_APURADO → EM_CONCILIACAO → APROVADO → EXPORTADO, com o recebimento tratado como situação separada (PENDENTE/PARCIAL/LIQUIDADO/ESTORNADO); o ciclo de compra SHALL comparar o consumo apurado com o extrato/fatura do provedor e gerar obrigação aprovada para pagamento, e o sistema NÃO SHALL marcar PAGO ou RECEBIDO apenas ao exportar o arquivo. A conciliação SHALL cruzar contrato, conta, período, operação externa, medidor, quantidade e valor; na ausência de provider_request_id o sistema SHALL exigir referência alternativa homologada, e agregação sem granularidade suficiente SHALL abrir divergência, NÃO SHALL ser tratada como casamento aproximado por dados pessoais. Diferenças por duplicação, ausência, tarifa, moeda, período e quantidade SHALL ter classificação, responsável, prazo e decisão registrados. A integração financeira v4 SHALL oferecer exportação e importação de confirmações com identidade de lote, versão, checksum, moeda, contraparte e chave por linha, e reexportação do mesmo lote SHALL preservar essa identidade para dedup no ERP; pagamento parcial e recebimento parcial SHALL atualizar saldo, sem editar o valor original. A conciliação bancária/liquidação SHALL permanecer no sistema financeiro autorizado, com o hub apenas acompanhando a confirmação, e transferência bancária automática e emissão fiscal NÃO SHALL ser autorizadas por esta especificação. Fato tardio após o fechamento NÃO SHALL alterar a fatura já emitida silenciosamente, devendo abrir ajuste em nova competência ou documento de correção aprovado, e ajuste manual, aprovação e confirmação de pagamento SHALL exigir segregação de funções.

#### Scenario: Ciclo de venda segue os estados definidos sem marcar PAGO/RECEBIDO na exportação
- GIVEN um período de apuração de venda avança pelos estados do ciclo
- WHEN o lote é exportado ao final do estado APROVADO
- THEN o sistema SHALL transitar ABERTO → PRE_APURADO → EM_CONCILIACAO → APROVADO → EXPORTADO, e NÃO SHALL marcar a situação de recebimento como LIQUIDADO apenas por ter exportado o arquivo

#### Scenario: Agregação sem granularidade suficiente abre divergência, não casamento aproximado
- GIVEN a conciliação não possui provider_request_id nem referência alternativa homologada suficiente para casar uma operação
- WHEN o sistema tenta conciliar essa linha
- THEN o sistema SHALL abrir uma divergência explícita e NÃO SHALL realizar casamento aproximado usando dados pessoais como substituto de granularidade

#### Scenario: Fato tardio após fechamento não altera fatura emitida silenciosamente
- GIVEN um fato de uso chega depois do fechamento de uma competência já faturada
- WHEN o sistema processa esse fato tardio
- THEN o sistema NÃO SHALL alterar a fatura já emitida silenciosamente, e SHALL abrir um ajuste em nova competência ou um documento de correção aprovado

#### Scenario: Transferência bancária automática e emissão fiscal não estão autorizadas
- GIVEN uma obrigação aprovada para pagamento ou uma fatura aprovada existe no ledger
- WHEN o sistema conclui o ciclo de fechamento dessa obrigação ou fatura
- THEN o sistema NÃO SHALL executar transferência bancária automática nem emissão fiscal, deixando a liquidação bancária no sistema financeiro autorizado e apenas acompanhando sua confirmação

#### Scenario: Ajuste manual, aprovação e confirmação de pagamento exigem segregação de funções
- GIVEN um ajuste manual no ledger precisa ser aprovado e uma confirmação de pagamento precisa ser registrada
- WHEN essas ações são executadas
- THEN o sistema SHALL exigir segregação de funções entre quem propõe o ajuste, quem aprova e quem confirma o pagamento, sem permitir que a mesma função concentre essas etapas

### Requirement: FIN-09 — Exemplos de referência para QA
Os valores a seguir SHALL ser tratados como sintéticos, não como condições negociadas, e SHALL servir de referência de comportamento verificável para QA sobre pacote, franquia, faixas marginais versus volume total, incerteza/failover e correção.

#### Scenario: Produto em pacote reconhece receita e custo compostos sem alteração por GET/callback
- GIVEN um produto em pacote tem preço de R$ 1,00, custo do componente A de R$ 0,20, custo do componente B de R$ 0,30, e duas consultas de B a R$ 0,01 cada
- WHEN o sistema apura essa execução, incluindo GETs do cliente e um callback repetido
- THEN o sistema SHALL reconhecer receita de R$ 1,00, custo de R$ 0,52 e diferença bruta de R$ 0,48 (antes de infraestrutura, tributos e outros custos), e os GETs do cliente e o callback repetido NÃO SHALL alterar esses valores

#### Scenario: Franquia mensal com excedente calcula parcelas distintas
- GIVEN uma mensalidade de R$ 50,00 inclui 100 sucessos com excedente de R$ 0,80 por unidade além da franquia
- WHEN o cliente atinge 103 sucessos elegíveis no período, sem que o plano declare consumo de franquia por falhas
- THEN o sistema SHALL reconhecer receita total de R$ 52,40, tratando a receita recorrente de R$ 50,00 e o excedente de R$ 2,40 como parcelas distintas

#### Scenario: Faixa progressiva marginal aplica preço apenas à parcela dentro de cada faixa
- GIVEN um plano de faixas progressivas marginais cobra R$ 1,00 pelas 100 primeiras unidades e R$ 0,80 pelas unidades seguintes
- WHEN o cliente consome 120 unidades no período
- THEN o sistema SHALL calcular a receita como R$ 116,00 (100 unidades a R$ 1,00 mais 20 unidades a R$ 0,80), sem chamar esse modelo apenas de "faixa" indistintamente do modelo de volume total

#### Scenario: Faixa por volume total aplica a faixa atingida a todo o volume elegível
- GIVEN um plano de faixa por volume total aplica R$ 0,80 a toda unidade quando o volume atinge 101 unidades
- WHEN o cliente consome 120 unidades no período
- THEN o sistema SHALL calcular a receita como R$ 96,00 (120 unidades a R$ 0,80), reconhecendo que esse resultado difere do modelo progressivo marginal para o mesmo volume

#### Scenario: Incerteza com failover autorizado contabiliza o custo dos dois provedores comprovados
- GIVEN os provedores A e B cobram R$ 0,30 e R$ 0,40 respectivamente pela mesma necessidade lógica, sob autorização de risco, com ambas as operações comprovadas
- WHEN o sistema apura essa execução com uma única receita de produto de R$ 1,00
- THEN o sistema SHALL reconhecer custo total de R$ 0,70 (R$ 0,30 de A mais R$ 0,40 de B), e NÃO SHALL registrar apenas o provedor vencedor, o que ocultaria R$ 0,30 de custo

#### Scenario: Correção de cobrança indevida gera estorno vinculado com lote balanceado
- GIVEN uma cobrança original de R$ 1,00 é identificada como indevida
- WHEN o sistema registra o estorno de R$ 1,00 vinculado a essa cobrança
- THEN o sistema SHALL manter valor líquido zero, preservar no histórico ambas as entradas (cobrança original e estorno) e manter o lote balanceado

### Requirement: FIN-10 — Quebra de SLA, custo tardio e créditos
O contrato SHALL diferenciar conclusão elegível, rejeição por SLA, fim de janela de retry, efeito externo tardio, crédito ao cliente e contestação de compra; um protocolo EXPIRED por SLA NÃO SHALL gerar receita do medidor de sucesso mesmo que o provedor conclua depois, e um plano que cobra aceite SHALL tratar essa cobrança como incidência distinta, definindo estorno/crédito quando o hub não cumprir o prazo, sem cobrar sucesso apenas por converter tardiamente o estado. O custo do provedor NÃO SHALL ser eliminado automaticamente pela rejeição no hub, devendo a apuração registrar custo comprovado, provisão incerta ou crédito esperado conforme o contrato de aquisição; contestação SHALL vincular a operação, a versão de SLA, os marcos observados, o valor contestado, as evidências e a decisão da contraparte, e crédito prometido mas não confirmado NÃO SHALL virar redução definitiva de obrigação sem regra financeira aprovada. O sistema SHALL separar créditos devidos pelo hub a clientes dos créditos devidos por provedores ao hub, SHALL basear divergência de SLA em fatos duráveis (não em captura isolada de dashboard), e o controle adaptativo MAY registrar polls/retries adicionais para apuração correta sem que a mudança de limite altere preço ou unidade contratada; diagnósticos de desenvolvedores NÃO SHALL alterar ledger, aprovar ajuste ou confirmar liquidação.

#### Scenario: EXPIRED por SLA não gera receita de sucesso mesmo com conclusão tardia do provedor
- GIVEN um cliente paga R$ 1,00 por sucesso concluído até 60 segundos, o provedor custa R$ 0,30 por operação aceita, e o provedor finaliza apenas aos 80 segundos
- WHEN o protocolo expira aos 60 segundos por SLA
- THEN o sistema SHALL reconhecer receita zero pelo medidor de sucesso e SHALL registrar o custo confirmado de R$ 0,30, abrindo contestação se houver crédito de compra por atraso previsto no contrato de aquisição, sem presumir margem zero ou apagar a operação

#### Scenario: Custo do provedor não é eliminado automaticamente pela rejeição no hub
- GIVEN uma operação é rejeitada no hub por quebra de SLA
- WHEN o sistema apura o custo dessa operação junto ao provedor
- THEN o sistema SHALL registrar custo comprovado, provisão incerta ou crédito esperado conforme o contrato de aquisição, NÃO SHALL eliminar automaticamente esse custo apenas pela rejeição no hub

#### Scenario: Crédito prometido não confirmado não reduz obrigação sem regra aprovada
- GIVEN um crédito de compra é prometido pelo provedor por atraso, mas ainda não confirmado
- WHEN o sistema apura a obrigação a pagar relacionada
- THEN o sistema NÃO SHALL reduzir definitivamente a obrigação com base no crédito prometido sem uma regra financeira aprovada, e SHALL manter separados os créditos devidos pelo hub a clientes dos créditos devidos por provedores ao hub

#### Scenario: Diagnósticos de desenvolvedores não alteram ledger, ajuste ou liquidação
- GIVEN uma ferramenta de diagnóstico é usada por desenvolvedores para investigar uma divergência de SLA
- WHEN essa ferramenta acessa dados de apuração
- THEN o sistema NÃO SHALL permitir que esse diagnóstico altere o ledger, aprove ajuste ou confirme liquidação, mantendo a divergência de SLA baseada em fatos duráveis, não em captura isolada de dashboard

#### Scenario: Controle adaptativo registra retries extras sem alterar preço ou unidade contratada
- GIVEN o controle adaptativo do sistema aumenta a frequência de polls/retries para apuração correta de uma operação
- WHEN essa mudança de limite técnico ocorre
- THEN o sistema SHALL registrar os polls/retries adicionais para fins de apuração, e NÃO SHALL alterar o preço ou a unidade contratada em razão dessa mudança de limite

### Requirement: FIN-11 — Conta, credencial e responsabilidade econômica
O modo da credencial NÃO SHALL determinar por si só quem paga o provedor; cada vínculo contratual SHALL explicitar settlement_party (HUB ou CLIENT_DIRECT), titular da conta externa, contraparte faturada, contrato de compra/acesso e permissão de medição, e o sistema NÃO SHALL inferir essas condições pelo nome do segredo. O sistema SHALL aplicar a regra de apuração por modelo de acesso: SHARED_HUB comprado pelo Hub (conta global com segregação por tenant/operação/medidor; custo de aquisição ao Hub e venda independente por plano do cliente), TENANT_DEDICATED comprado pelo Hub (conta/vínculo dedicado com contrato do tenant e capacidade externa aplicável; custo e venda rastreados, sem presumir que a chave dedicada elimina a obrigação de compra) e TENANT_DEDICATED com compra direta do cliente (evidência de uso autorizada e identificação da contraparte; sem gerar conta a pagar pelo Hub, cobrando somente os medidores de Hub contratados, com custo externo possivelmente apenas informativo). Todo fato de consumo SHALL conservar tenant_id, provider_account_id, credential_binding_id/version, referência da versão de segredo utilizada, operation_id/attempt_id, contrato/medidor e responsável econômico, e NÃO SHALL incluir o valor do segredo; rotação de segredo na mesma conta NÃO SHALL iniciar novo medidor, nova quota global ou novo ciclo de franquia, e quando contas dedicadas compartilham teto/franquia comercial do provedor o agrupamento contratual SHALL prevalecer. Em falha de Libra, contrato pós-pago MAY continuar processando se fatos e orçamento de risco forem preservados no core, aguardando pagamento/apuração a recuperação, enquanto limite estrito continua exigindo reserva confirmada; credencial nova, troca de pagador ou ampliação de exposição NÃO SHALL ser adotadas como fallback automático durante incidente.

#### Scenario: Settlement_party é explícito por vínculo, não inferido pelo nome do segredo
- GIVEN um vínculo contratual entre tenant, conta externa e contrato de compra/acesso está sendo configurado
- WHEN o sistema determina quem paga o provedor
- THEN o sistema SHALL exigir que settlement_party (HUB ou CLIENT_DIRECT), titular da conta, contraparte faturada e permissão de medição estejam explicitados no vínculo, e NÃO SHALL inferir essas condições a partir do nome do segredo ou credencial usada

#### Scenario: SHARED_HUB comprado pelo Hub segrega consumo e mantém venda independente
- GIVEN uma conta de provedor é do modelo SHARED_HUB, comprada pelo Hub
- WHEN múltiplos tenants consomem essa conta global
- THEN o sistema SHALL segregar o consumo por tenant/operação/medidor, SHALL atribuir o custo de aquisição ao Hub, e SHALL apurar a venda a cada cliente de forma independente conforme seu plano

#### Scenario: TENANT_DEDICATED comprado pelo Hub não elimina obrigação de compra
- GIVEN uma conta dedicada a um tenant é comprada pelo Hub
- WHEN o sistema apura custo e venda dessa conta
- THEN o sistema SHALL rastrear custo e venda separadamente, e NÃO SHALL presumir que a existência de uma chave dedicada elimina a obrigação de compra do Hub perante o provedor

#### Scenario: TENANT_DEDICATED com compra direta do cliente não gera conta a pagar pelo Hub
- GIVEN uma conta dedicada tem compra direta pelo cliente junto ao provedor, com evidência de uso autorizada e contraparte identificada
- WHEN o sistema apura essa conta
- THEN o sistema NÃO SHALL gerar conta a pagar pelo Hub para esse custo externo, SHALL cobrar somente os medidores de Hub efetivamente contratados, e o custo externo MAY ser tratado apenas como informativo

#### Scenario: Rotação de segredo não reinicia medidor, quota ou franquia
- GIVEN um segredo de credencial é rotacionado na mesma conta de provedor
- WHEN o sistema processa operações após essa rotação
- THEN o sistema NÃO SHALL iniciar novo medidor, nova quota global ou novo ciclo de franquia apenas em razão da rotação do segredo

#### Scenario: Fato de consumo conserva identificadores completos sem expor valor do segredo
- GIVEN uma operação de consumo é registrada como fato durável
- WHEN o sistema persiste esse fato
- THEN o sistema SHALL conservar tenant_id, provider_account_id, credential_binding_id/version, referência da versão de segredo, operation_id/attempt_id, contrato/medidor e responsável econômico, e NÃO SHALL incluir o valor do segredo nesse registro

#### Scenario: Falha de Libra não autoriza fallback automático de credencial ou pagador
- GIVEN Libra está indisponível durante um incidente
- WHEN um contrato pós-pago sem teto estrito continua recebendo pedidos com fatos e orçamento de risco preservados no core
- THEN o sistema MAY continuar processando esses pedidos aguardando recuperação para pagamento/apuração, SHALL continuar exigindo reserva confirmada para contratos de limite estrito, e NÃO SHALL adotar credencial nova, troca de pagador ou ampliação de exposição como fallback automático durante o incidente

## Notas de origem

Este delta deriva integralmente do capítulo `docs/05_CONTRATOS_CONSUMO_E_FINANCEIRO.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos FIN-01 a FIN-11 conforme texto normativo de origem.
