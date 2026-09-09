# Delta for 05-financeiro-e-entregas

## ADDED Requirements

### Requirement: R3-FIN-01 — Reserva estrita e franquia antes do efeito
O Hub SHALL autorizar saldo/franquia antes do efeito, reservar exposição máxima ou obter autorização incremental prévia e conciliar captura/liberação com valor efetivo exato. UNKNOWN mantém retenção até evidência positiva; expiração do cliente não prova ausência de custo.

#### Scenario: R3-FIN-01-S01 — tarifa efetiva 2 e reserva inicial 1
- GIVEN tarifa efetiva 2 e reserva inicial 1
- WHEN autorizar/concluir
- THEN exposição não autorizada é impedida e captura é exata

#### Scenario: R3-FIN-01-S02 — franquia DENY esgotada e admissões concorrentes
- GIVEN franquia DENY esgotada e admissões concorrentes
- WHEN tentar executar
- THEN nenhum efeito excede autorização e recusa é idempotente

#### Scenario: R3-FIN-01-S03 — UNKNOWN seguido de prova de execução/ausência
- GIVEN UNKNOWN seguido de prova de execução/ausência
- WHEN reconciliar
- THEN captura ou liberação ocorre uma vez sem hold inexplicado permanente

### Requirement: R3-FIN-02 — Incidência completa e fechamento operacional
O Hub SHALL medir unidades de compra/venda do snapshot, incluindo SUBMIT/STATUS/FETCH contratados, com attempt_id/evidence_id e deduplicação por unidade. Watermarks derivam de completude demonstrável; disputa, conciliação e exportação devem funcionar sem editar banco manualmente.

#### Scenario: R3-FIN-02-S01 — contrato cobra cada STATUS e receita só no sucesso
- GIVEN contrato cobra cada STATUS e receita só no sucesso
- WHEN executar três polls e um final
- THEN três unidades de compra e uma receita sem cobrança por redelivery

#### Scenario: R3-FIN-02-S02 — fato faltante ou disputa aberta
- GIVEN fato faltante ou disputa aberta
- WHEN fechar período
- THEN bloqueio identifica obrigação e resolução auditada habilita fechamento

#### Scenario: R3-FIN-02-S03 — duplicata ou fato após período fechado
- GIVEN duplicata ou fato após período fechado
- WHEN reprocessar
- THEN não duplica lançamento e fato tardio tem ajuste/quarentena recuperável

### Requirement: R3-FIN-03 — Destino de webhook congelado por aplicação
O Hub SHALL congelar destinos autorizados e contrato de entrega, conservar representação final imutável e distinguir reentrega de execução. Ausência de destino tem disposição explícita. Tentativas, leases e comando manual respeitam budgets auditáveis.

#### Scenario: R3-FIN-03-S01 — destino v1 no aceite e v2 antes do final
- GIVEN destino v1 no aceite e v2 antes do final
- WHEN entregar
- THEN destino segue snapshot e bytes coincidem com GET

#### Scenario: R3-FIN-03-S02 — aplicações A/B do mesmo tenant com destinos diferentes
- GIVEN aplicações A/B do mesmo tenant com destinos diferentes
- WHEN finalizar ambas
- THEN não ocorre fan-out cruzado

#### Scenario: R3-FIN-03-S03 — queda após envio e redelivery manual
- GIVEN queda após envio e redelivery manual
- WHEN retomar
- THEN identidade/bytes são preservados e nenhum SUBMIT é provocado
