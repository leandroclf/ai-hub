# Delta for r6-03-reconciliacao-financeira

## ADDED Requirements

### Requirement: R6-FIN-01 — Replay financeiro distingue aplicação de nova quarentena
O Hub SHALL produzir disposição tipada para replay: aplicado, ainda em quarentena, conflito ou falha transitória. Uma obrigação não aplicada SHALL continuar visível e recuperável. Todo payload custodiado SHALL declarar formato/versão e ter vínculo auditável com bytes de origem, inclusive sem identidade de domínio válida.

#### Scenario: R6-FIN-01-S01 — quarentena cujo payload continua inválido
- GIVEN quarentena cujo payload continua inválido
- WHEN operador solicita replay
- THEN permanece pendente com tentativa registrada, nunca REPLAYED por simples nil

#### Scenario: R6-FIN-01-S02 — fato tardio após período fechado
- GIVEN fato tardio após período fechado
- WHEN consultar e reconciliar pelo endpoint
- THEN formato é interpretado e ajuste vinculado ao período sem reemitir SUBMIT

#### Scenario: R6-FIN-01-S03 — crash após aplicação e antes da confirmação do replay
- GIVEN crash após aplicação e antes da confirmação do replay
- WHEN repetir comando com mesma identidade
- THEN um efeito econômico, estado consistente e auditoria por ator

### Requirement: R6-FIN-02 — Completude e liquidação verificáveis no fluxo real
O Hub SHALL fechar período somente com prova de completude de todos os produtores e obrigações da coorte. Reserva, captura efetiva, liberação e ajustes SHALL conservar identidade e valores exatos sob reordenação, duplicação e resultados tardios.

#### Scenario: R6-FIN-02-S01 — fatos fora de ordem e um produtor atrasado
- GIVEN fatos fora de ordem e um produtor atrasado
- WHEN solicitar fechamento
- THEN período bloqueia até cobertura durável de todos os fatos elegíveis

#### Scenario: R6-FIN-02-S02 — reserva 10 e consumo efetivo 7, com duplicata
- GIVEN reserva 10 e consumo efetivo 7, com duplicata
- WHEN processar no consumer e consultar saldo/journal
- THEN captura 7, liberação 3 e nenhum lançamento duplicado

#### Scenario: R6-FIN-02-S03 — fato econômico chega depois de exportação
- GIVEN fato econômico chega depois de exportação
- WHEN reconciliar e aprovar ajuste
- THEN exportação original permanece íntegra e ajuste é consultável e rastreável
