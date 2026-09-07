# 05 · Contratos, consumo e financeiro

## FIN-01 · Contrato de aquisição — proprietário: Atlas / Comercial

Cada conta de provedor deve estar vinculada a contrato de compra versionado com contraparte, identificador externo, vigência, ambiente, serviços/operações, moeda, unidades, tabela de preços, descontos, franquias, mínimos, teto, quota, concorrência, prazo de resposta, condições de armazenamento e formato de extrato/fatura.

Especificar os fatos tarifáveis: submit aceito, execução final, consulta de status, download de resultado, cancelamento, compensação ou outro medidor claramente definido. Declarar se timeout, resultado negativo, erro técnico, resposta tardia e retransmissão são cobrados. Ausência de tabela ou interpretação ambígua bloqueia publicação comercial, não é tratada como custo zero.

Contrato vencido bloqueia novas operações conforme política. Operações já aceitas e consultas necessárias à conclusão precisam de regra de continuidade previamente acordada. Credencial bloqueada é impedimento técnico; contrato vencido é condição comercial distinta. A plataforma registra qual versão de aquisição se aplica a cada operação e a regra temporal para seus polls/fetches posteriores.

## FIN-02 · Contrato de venda e planos — proprietário: Atlas / Comercial

Contrato do cliente deve informar tenant, aplicações, ofertas/versões autorizadas, plano, vigência, moeda, periodicidade, fuso de fechamento, unidade, marco de cobrança, preços, descontos, franquias, excedentes, mínimos, limites e forma de integração financeira. O plano declara custo de resultado parcial, sem resultado, falha, cancelamento, revisão e reexecução.

Modelos suportados pela especificação: preço unitário por unidade; assinatura com franquia e excedente; faixas progressivas marginais; preço por faixa de volume total; pacote de produto; híbrido explicitamente decomposto. “Progressivo marginal” aplica preço apenas à parcela na faixa; “volume total” aplica a faixa atingida a todo o volume elegível. A configuração deve diferenciá-los e definir ajustes no fechamento.

Franquia declara unidade, período, escopo (contrato/produto/serviço), compartilhamento entre aplicações, ordem de consumo, reset, rollover permitido ou não e devolução por estorno. Default proposto: sem rollover e sem devolução automática fora do período; exceções são contratuais. Quota de tráfego não é franquia financeira.

## FIN-03 · Vigência e snapshot — proprietário: Atlas / Libra

Fixar no protocolo contrato/plano de venda e versão da oferta. Fixar na operação externa contrato/conta de compra. Persistir preço/regra efetivamente aplicados, marco de incidência, instante de ocorrência e instante de recebimento do fato. Fuso contratual define competência; relógios e registros internos são UTC.

Alteração comum vale para novas admissões; não reprifica protocolos em voo. Se o provedor cobra polls pela tabela vigente no instante de cada consulta, isso deve constar como exceção temporal na versão fixada, e cada consulta referencia a tabela aplicável naquele instante. Contratos sobrepostos exigem prioridade explícita e não podem deixar duas regras elegíveis sem resolução.

Aditivo retroativo requer autorização, escopo e lançamentos de ajuste, preservando cálculo original. Revogação de segurança pode interromper operações em voo e deve gerar evidência; não reescreve contrato nem custo já incorrido.

## FIN-04 · Medição e chaves econômicas — proprietário: Libra

Telemetria técnica contabiliza chamadas, latência e falhas; medição econômica identifica unidade tarifável. Libra recebe fatos duráveis de Cometa e Órbita, reconhece a chave semântica e aplica contrato. Logs não são a fonte autoritativa de faturamento.

| Natureza | Chave econômica mínima | Regra |
| --- | --- | --- |
| Receita por produto executado | cliente/contrato/protocolo/medidor de produto | Uma unidade elegível apesar de múltiplos passos |
| Receita por componente | cliente/contrato/protocolo/passo/medidor | Somente se a política do produto cobra componentes |
| Custo por operação | provedor/conta/contrato/operation_id/medidor | Preserva identidade entre retransmissões |
| Custo por consulta de status | operação/interação física tarifável/medidor | Cada consulta cobrada tem evidência própria |
| Ajuste/estorno | lançamento original/tipo/identidade do ajuste autorizado | Não duplica por reenvio do evento de ajuste |
| Assinatura/mínimo | contrato/período/parcela recorrente | Uma parcela por competência, independentemente de pedidos |

Se o provedor cobra cada chamada de transporte, inclusive retransmissão, a chave de custo inclui tentativa e evidência contratual. “Idempotente” não significa “gratuito”. Se o custo é incerto, registrar provisão/pendência identificável; não registrar zero definitivo nem débito confirmado sem regra/evidência.

## FIN-05 · Matriz de incidência — proprietário: Financeiro / Produto

| Situação | Receita do cliente | Custo do provedor |
| --- | --- | --- |
| Pedido recusado antes do aceite | Sem consumo do serviço | Sem operação externa |
| Pedido aceito ainda pendente | Conforme marco contratado; não presumir final | Pode haver submit tarifável |
| Sucesso válido | Medidor previsto no plano | Operações/consultas tarifáveis comprovadas |
| “Não encontrado” válido | Conforme semântica do serviço/plano | Conforme aquisição, mesmo sem dados |
| Falha técnica conhecida | Política explícita; default proposto sem receita por sucesso | Pode haver custo de tentativa ou aceitação |
| Resultado parcial de produto | Preço/regra de parcialidade publicada | Custos dos passos efetivamente tarifáveis |
| Webhook falhou ou foi reentregue | Não altera receita por execução por si só | Não reexecuta provedor |
| GET de resultado armazenado | Sem nova receita no plano padrão v4 | Nenhuma nova chamada externa |
| Callback e polling reportam mesmo final | Uma unidade de receita | Operação única; polling em voo pode ter custo real |
| Failover autorizado após incerteza | Conforme uma entrega de produto ou regra explícita | A e B podem cobrar; contabilizar ambos quando comprovados |
| Cancelamento/compensação | Regra de estorno explícita | Operação original pode continuar cobrada e compensação ter custo |
| Resultado corrigido | Ajuste rastreável quando necessário; não reabrir EXPIRED por SLA | Não assumir cancelamento do custo original |
| Final tardio após EXPIRED/SLA_EXCEEDED | Sem receita do medidor de sucesso; crédito/estorno de aceite conforme contrato | Custo real ou contestação, sem eliminação automática |

Contrato omitindo uma linha aplicável não pode ser habilitado para faturamento automático. Defaults propostos precisam aparecer no contrato publicado; não operar com regras invisíveis no runtime.

## FIN-06 · Reserva de saldo e limites — proprietário: Libra / Órbita

Separar: rate limit técnico; quota operacional; franquia de apuração; e saldo/teto financeiro estrito. Somente os dois últimos exigem semântica econômica. Franquia usada apenas para cálculo posterior pode ser contabilizada por eventos. Promessa de limite estrito requer reserva atômica na autoridade financeira antes de qualquer efeito externo.

Para contrato estrito, Órbita registra pedido e intenção; obtém de Libra uma reserva idempotente ligada ao protocolo. O protocolo pode estar ACCEPTED aguardando autorização, mas nenhum passo externo é liberado antes da reserva confirmada. Libra indisponível mantém estado financeiro pendente até prazo de admissão financeira; negar definitivamente exige falha conhecida, não timeout incerto.

Reserva deve cobrir o máximo autorizável do produto, incluindo passos opcionais e consultas cobradas dentro do orçamento. Se custo máximo não puder ser limitado, não oferecer esse contrato como teto estrito. Captura aplica consumo efetivo; libera excedente comprovado. Expiração de reserva não pode liberar saldo de operação externa ainda em voo; reconciliar antes. Consulta idempotente resolve resposta de reserva perdida. Contratos pós-pagos sem teto estrito podem continuar processando durante atraso de apuração, dentro do limite de risco aprovado.

## FIN-07 · Apuração e ledger — proprietário: Libra / Financeiro

Manter fatos de uso imutáveis, cálculo versionado e ledger gerencial de partidas balanceadas por moeda. Receita a receber e obrigação a pagar são lançamentos distintos, não uma única diferença líquida. Conta de passagem, provisão, desconto e estorno têm natureza definida no plano de contas gerencial aprovado por Financeiro. Esse ledger suporta integração contábil, mas não substitui automaticamente a contabilidade oficial.

Não usar ponto flutuante monetário. Regra de arredondamento declara precisão da tarifa, escala da quantidade, momento do arredondamento (linha ou total) e moeda. Totais exibidos e exportados devem reproduzir o mesmo cálculo. Não somar moedas distintas; câmbio só com fonte, data, taxa e regra contratual explícitas. Tributos são campos e integrações de responsabilidade definida, sem motor fiscal presumido.

Rascunhos podem ser recalculados com versões; lançamento publicado é imutável. Correção é lançamento inverso ou ajuste vinculado. Invariante: débitos e créditos do lote balanceiam por moeda; toda linha tem chave econômica/origem; nenhum fechamento deixa fato elegível silenciosamente não apurado.

## FIN-08 · Fechamento, conciliação e pagamento — proprietário: Libra / Financeiro

Ciclo de venda: ABERTO → PRE_APURADO → EM_CONCILIACAO → APROVADO → EXPORTADO; recebimento é situação separada (PENDENTE/PARCIAL/LIQUIDADO/ESTORNADO). Ciclo de compra compara consumo apurado com extrato/fatura do provedor e gera obrigação aprovada para pagamento. Não marcar PAGO ou RECEBIDO ao exportar arquivo.

Conciliação cruza contrato, conta, período, operação externa, medidor, quantidade e valor. Ausência de provider_request_id exige referência alternativa homologada; agregação sem granularidade suficiente abre divergência, não casamento aproximado por dados pessoais. Diferenças por duplicação, ausência, tarifa, moeda, período e quantidade têm classificação, responsável, prazo e decisão.

Integração financeira v4 deve oferecer exportação e importação de confirmações com identidade de lote, versão, checksum, moeda, contraparte e chave por linha; formato final/API depende de P-06. Reexportação do mesmo lote preserva identidade para dedup no ERP. Pagamento parcial e recebimento parcial atualizam saldo, não editam valor original. Conciliação bancária/liquidação fica no sistema financeiro autorizado; o hub acompanha confirmação. Transferência bancária automática e emissão fiscal não estão autorizadas por esta especificação.

Fato tardio após fechamento não altera fatura emitida silenciosamente: abre ajuste em nova competência ou documento de correção aprovado. Reservas, provisões, fatos pendentes e divergências acompanham relatório de fechamento. Ajuste manual, aprovação e confirmação de pagamento exigem segregação de funções.

## FIN-09 · Exemplos de referência para QA — proprietário: Financeiro / QA

Valores abaixo são sintéticos e não representam condições negociadas.

- **Produto em pacote:** preço R$ 1,00; custo A R$ 0,20 e B R$ 0,30; duas consultas B a R$ 0,01. Receita R$ 1,00; custo R$ 0,52; diferença bruta R$ 0,48, antes de infraestrutura, tributos e outros custos. GETs do cliente e callback repetido não mudam esses valores.
- **Franquia:** mensalidade R$ 50,00 inclui 100 sucessos, excedente R$ 0,80. Com 103 sucessos elegíveis: R$ 52,40. Falhas só consomem franquia se o plano declarar. Receita recorrente e excedente são parcelas distintas.
- **Faixas marginais:** 100 primeiras unidades a R$ 1,00, seguintes a R$ 0,80; 120 unidades = R$ 116,00. **Faixa por volume total:** atingindo 101 unidades, todas a R$ 0,80; 120 = R$ 96,00. O portal não pode chamar ambos apenas de “faixa”.
- **Incerteza e failover excepcional:** A e B cobram R$ 0,30 e R$ 0,40 pela mesma necessidade lógica, com autorização de risco e ambas as operações comprovadas. Uma receita de produto a R$ 1,00, custo R$ 0,70. Registrar só o provedor vencedor ocultaria R$ 0,30 de custo.
- **Correção:** cobrança original R$ 1,00 indevida; estorno de R$ 1,00 vinculado. Valor líquido zero, histórico com ambas as entradas e lote balanceado.

## FIN-10 · Quebra de SLA, custo tardio e créditos — proprietário: Financeiro / Comercial

O contrato diferencia conclusão elegível, rejeição por SLA, fim de janela de retry, efeito externo tardio, crédito ao cliente e contestação de compra. Protocolo EXPIRED por SLA não gera receita do medidor de sucesso, mesmo que o provedor conclua depois. Plano que cobra aceite é outra incidência e precisa definir estorno/crédito quando o hub não cumpre prazo; não cobrar sucesso por converter tarde o estado.

Custo do provedor não é eliminado automaticamente pela rejeição no hub. A apuração registra custo comprovado, provisão incerta ou crédito esperado conforme contrato de aquisição. Contestação deve vincular a operação, versão de SLA, marcos observados, valor contestado, evidências e decisão da contraparte. Crédito prometido mas não confirmado não vira redução definitiva de obrigação sem regra financeira aprovada.

Separar créditos devidos pelo hub a clientes e créditos devidos por provedores ao hub. Divergência de SLA usa fatos duráveis, não captura isolada de um dashboard. O controle adaptativo registra polls/retries adicionais para apuração correta, mas a mudança de limite não altera preço ou unidade contratada. Diagnósticos de desenvolvedores não podem alterar ledger, aprovar ajuste ou confirmar liquidação.

Exemplo sintético: cliente paga R$1,00 por sucesso até 60 s; provedor custa R$0,30 por operação aceita e finaliza aos 80 s. Protocolo expira aos 60 s, receita por sucesso é zero e custo confirmado pode ser R$0,30. Se houver crédito de compra por atraso, abrir contestação; não presumir margem zero ou apagar a operação.

## FIN-11 · Conta, credencial e responsabilidade econômica — proprietário: Comercial / Financeiro / Libra

Modo da credencial não determina quem paga o provedor. Cada vínculo contratual explicita settlement_party (HUB ou CLIENT_DIRECT), titular da conta externa, contraparte faturada, contrato de compra/acesso e permissão de medição. Uma chave dedicada pode ser de uma conta cujo custo ainda pertence ao Hub; uma integração em que o cliente paga diretamente pode ter apenas tarifa de intermediação do Hub. Não inferir essas condições pelo nome do segredo.

| Modelo de acesso | Controle de consumo | Regra de apuração |
| --- | --- | --- |
| SHARED_HUB, compra pelo Hub | Conta global e segregação por tenant/operação/medidor | Custo de aquisição ao Hub; venda independente por plano do cliente |
| TENANT_DEDICATED, compra pelo Hub | Conta/vínculo dedicado, contrato do tenant e capacidade externa aplicável | Custo e venda rastreados; não presumir que a chave dedicada elimina obrigação de compra |
| TENANT_DEDICATED, compra direta do cliente | Evidência de uso autorizada e identificação da contraparte | Não gerar conta a pagar pelo Hub; cobrar somente os medidores de Hub contratados; custo externo pode ser apenas informativo |

Fato de consumo conserva tenant_id, provider_account_id, credential_binding_id/version, referência da versão de segredo utilizada, operation_id/attempt_id, contrato/medidor e responsável econômico. Nunca inclui valor do segredo. Rotação de segredo na mesma conta não inicia novo medidor, nova quota global ou novo ciclo de franquia. Se contas dedicadas compartilham teto/franquia comercial do provedor, o agrupamento contratual prevalece.

Polls, fetches, callbacks com custo, retries, compensações e evidências tardias seguem o contrato fixado da operação. A apuração não pode agregar clientes dedicados à conta compartilhada por falta de campo de associação. Importação de extrato usa conta/contrato/período e referências homologadas; operação sem correspondência vira divergência explícita. Acesso ao extrato não autoriza revelar preços/consumo de outro cliente.

Em falha de Libra, pós-pago pode continuar se fatos e orçamento de risco forem preservados no core; pagamento/apuração aguardam recuperação. Limite estrito continua exigindo reserva confirmada na autoridade financeira. Credencial nova, troca de pagador ou ampliação de exposição não são fallback automático durante incidente. O teste de aceite inclui as três linhas acima, rotação, polling pago e rejeição tardia por SLA, sem duplicação ou transferência de dívida entre partes.
