# Riscos e mitigação

| Risco | Prioridade | Mitigação | Aceite |
|---|---|---|---|
| RLS no caminho real e prova com dados existentes | P0 | O Hub SHALL aplicar escopo autenticado por transação e por item de trabalho com roles runtime sem bypass; tabelas sem tenant direto devem ter autorização por relação ou autoridade operacional específica. Testes de isolamento SHALL afirmar existência e acesso próprio, invisibilidade alheia e recusa de escrita cruzada antes de remover fixtures. | R6-SEG-01-S01…S03 |
| Callback aceito sobrevive à rotação e à indisponibilidade | P0 | O Hub SHALL preservar a atestação de autenticação obtida no ingresso e a disposição recuperável de cada callback aceito. Rotação e falha transitória não podem invalidar custódia legítima; esgotamento de retry deve manter obrigação em quarentena reprocessável, distinta de rejeição de contrato. | R6-SEG-02-S01…S03 |
