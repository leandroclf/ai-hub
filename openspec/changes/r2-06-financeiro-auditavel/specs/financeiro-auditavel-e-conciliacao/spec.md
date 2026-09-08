# Delta for financeiro-auditavel-e-conciliacao

## ADDED Requirements

### Requirement: R2-FIN-01 — Compra, venda e incidência por snapshot

Cada unidade econômica SHALL usar contrato de compra/venda, vigência, moeda, plano/medidor e responsabilidade de liquidação congelados no aceite ou marco explicitamente contratado. Receita e custo SHALL ser independentes. Alterar contrato atual NÃO SHALL recalcular evento passado; CLIENT_DIRECT NÃO SHALL gerar automaticamente contas a pagar do Hub.

Baseline relacionada: FIN-01, FIN-02, FIN-03, FIN-05, FIN-11.

#### Scenario: R2-FIN-01-S01 — Preço muda antes de consumir

- GIVEN pedido aceito em tarifa v1 e evento ainda está em fila
- WHEN tarifa v2 é publicada antes da apuração
- THEN apuração usa v1 e conserva prova do snapshot

#### Scenario: R2-FIN-01-S02 — Cliente paga diretamente

- GIVEN binding usa TENANT_DEDICATED com settlement_party=CLIENT_DIRECT
- WHEN operação gera consumo
- THEN consumo é registrado, mas passivo do Hub só existe se contrato de compra o prevê

#### Scenario: R2-FIN-01-S03 — Compra por submit

- GIVEN contrato prevê custo no aceite externo e venda por sucesso
- WHEN protocolo depois falha/expira
- THEN custo e ausência de receita seguem marcos distintos, sem preço constante de simulador

### Requirement: R2-FIN-02 — Medição exata e deduplicação econômica

Medição SHALL identificar a unidade faturável por contrato, natureza, medidor e identidade da obrigação pertinente (protocolo/passo/operação/tentativa), além do event_id. Valores e quantidades monetárias SHALL usar representação decimal exata fim a fim com regra de arredondamento por moeda/contrato. Eventos técnicos NÃO SHALL ser cobrados sem incidência definida.

Baseline relacionada: FIN-04, FIN-05, FIN-09.

#### Scenario: R2-FIN-02-S01 — Dois passos custosos

- GIVEN produto chama duas operações faturáveis do mesmo provedor
- WHEN fatos chegam em ordem inversa e duplicados
- THEN ambos custos elegíveis aparecem uma vez, sem colidir pela mesma chave de protocolo

#### Scenario: R2-FIN-02-S02 — Polling tarifado

- GIVEN contrato cobra status/fetch e três tentativas elegíveis ocorreram
- WHEN Libra apura
- THEN unidades e valores correspondem às três evidências, sem cobrar retransmissão da mensagem

#### Scenario: R2-FIN-02-S03 — Precisão

- GIVEN fixture tem 0,1 + 0,2 e taxas com quatro casas
- WHEN motor apura e exporta
- THEN resultado decimal é exato e arredondamento é único no marco contratado

### Requirement: R2-FIN-03 — Saldo estrito e retenção de incerteza

Oferta estrita SHALL exigir autoridade financeira única e limite/saldo explícito. Antes de efeito externo, reserva atômica SHALL considerar saldo liquidado, reservas e holds incertos, com identidade/moeda/valor consistentes. Ausência ou falha da autoridade NÃO SHALL conceder limite fictício. EXPIRED do cliente NÃO SHALL liberar hold de efeito externo ainda UNKNOWN.

Baseline relacionada: FIN-06, FIN-10, DAD-03, DAD-11.

#### Scenario: R2-FIN-03-S01 — Limite cumulativo

- GIVEN fixture tem saldo 2,00 e cada execução consome 1,00
- WHEN três execuções sequenciais são solicitadas após capturas
- THEN apenas duas executam; terceira é negada antes de efeito externo

#### Scenario: R2-FIN-03-S02 — Concorrência pelo último saldo

- GIVEN duas solicitações disputam 1,00 restante
- WHEN reservas ocorrem simultaneamente
- THEN apenas uma é concedida; moeda/tenant/valor fazem parte da validação

#### Scenario: R2-FIN-03-S03 — Limite não configurado

- GIVEN contrato é estrito e não tem limite aprovado
- WHEN pedido solicita reserva
- THEN é recusado sem usar default alto; consulta de final existente continua

#### Scenario: R2-FIN-03-S04 — Timeout com efeito incerto

- GIVEN protocolo expira e provedor pode cobrar
- WHEN reserva é reconciliada
- THEN hold permanece até evidência de cancelamento/ausência/custo; duplicata não libera nem captura duas vezes

### Requirement: R2-FIN-04 — Planos, franquias e política de produto

Motor comercial SHALL suportar políticas versionadas de preço unitário, pacote, soma/híbrido, franquia e faixas marginais ou por volume explicitamente distinguidas. Consumo de franquia estrita SHALL ser atômico. Parcialidade, retries e compensações SHALL ter incidência definida antes da ativação.

Baseline relacionada: FIN-02, FIN-05, FIN-09, CAT-07.

#### Scenario: R2-FIN-04-S01 — Faixa marginal

- GIVEN fixture aprovada define primeiras duas unidades a 1,00 e excedentes a 0,50
- WHEN quatro unidades elegíveis são apuradas
- THEN total é 3,00 na regra marginal; regra por volume não é inferida

#### Scenario: R2-FIN-04-S02 — Pacote versus soma

- GIVEN produto A+B tem contrato de pacote 5,00 e custos independentes
- WHEN ambos passos concluem
- THEN receita é 5,00 uma vez, custos conforme compra, sem somar venda dos componentes indevidamente

#### Scenario: R2-FIN-04-S03 — Franquia final concorrente

- GIVEN resta uma unidade de franquia estrita
- WHEN duas operações concorrentes consomem
- THEN uma usa franquia e a outra segue regra de excedente ou recusa explicitamente contratada

### Requirement: R2-FIN-05 — Ledger imutável e ajustes compensatórios

Apuração SHALL gerar journal de partidas balanceadas por moeda com chaves idempotentes e origem rastreável. Lançamentos confirmados NÃO SHALL ser editados/apagados para corrigir valores; ajuste SHALL referenciar origem, razão e autorizador. Fato elegível sem journal SHALL impedir conclusão contábil do lote.

Baseline relacionada: FIN-07, FIN-10.

#### Scenario: R2-FIN-05-S01 — Partidas balanceadas

- GIVEN fatos elegíveis foram deduplicados
- WHEN apuração confirma lote
- THEN débitos e créditos se equilibram por moeda e cada lançamento aponta contrato/unidade/evidência

#### Scenario: R2-FIN-05-S02 — Reprocessamento

- GIVEN mesmo fato e lote são apresentados novamente
- WHEN motor tenta apurar
- THEN não há novos lançamentos nem alteração do valor original

#### Scenario: R2-FIN-05-S03 — Estorno autorizado

- GIVEN cobrança incorreta precisa correção
- WHEN usuário autorizado aprova ajuste
- THEN partidas compensatórias preservam original, razão e correlação

### Requirement: R2-FIN-06 — Fechamento, reconciliação e integração financeira

Fechamento SHALL demonstrar completude de fatos, pendências, watermarks e ajustes do período, reconciliar compra e venda separadamente e produzir exportação versionada/idempotente com checksum e recibo. Percentil de atualidade NÃO SHALL provar completude. Formato ERP e execução de pagamentos SHALL depender do contrato aprovado, sem inferir autorização de transferência bancária.

Baseline relacionada: FIN-08, FIN-10, FIN-11.

#### Scenario: R2-FIN-06-S01 — Fato atrasado

- GIVEN watermark ou UNKNOWN financeiro indica pendência do período
- WHEN operador solicita fechamento definitivo
- THEN lote fica bloqueado ou em exceção formal auditada, nunca silenciosamente completo

#### Scenario: R2-FIN-06-S02 — Exportação repetida

- GIVEN lote fechado já foi exportado e recibo se perdeu
- WHEN integração reenvia mesma identidade
- THEN destino pode deduplicar; Hub mantém estado e checksum sem novo débito/pagamento

#### Scenario: R2-FIN-06-S03 — Contestação de SLA

- GIVEN custo tardio está sendo contestado
- WHEN financeiro revisa conciliação
- THEN valor, evidência, disputa e ajuste ficam separados da resposta final imutável do cliente
