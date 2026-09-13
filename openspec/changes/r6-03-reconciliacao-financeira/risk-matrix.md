# Riscos e mitigação

| Risco | Prioridade | Mitigação | Aceite |
|---|---|---|---|
| Replay financeiro distingue aplicação de nova quarentena | P0 | O Hub SHALL produzir disposição tipada para replay: aplicado, ainda em quarentena, conflito ou falha transitória. Uma obrigação não aplicada SHALL continuar visível e recuperável. Todo payload custodiado SHALL declarar formato/versão e ter vínculo auditável com bytes de origem, inclusive sem identidade de domínio válida. | R6-FIN-01-S01…S03 |
| Completude e liquidação verificáveis no fluxo real | P0 | O Hub SHALL fechar período somente com prova de completude de todos os produtores e obrigações da coorte. Reserva, captura efetiva, liberação e ajustes SHALL conservar identidade e valores exatos sob reordenação, duplicação e resultados tardios. | R6-FIN-02-S01…S03 |
