# Risk matrix

| Risco | Prioridade | Mitigação | Prova | Responsável |
|---|---|---|---|---|
| Reserva exata existe, mas captura não ajusta seu valor ao valor efetivo. DENY por franquia ocorre na incidência após execução externa. | P0 | R3-FIN-01 | R3-FIN-01-S01/S02/S03 | Financeiro e Core |
| Fatos não transportam todas as unidades/identidades ATTEMPT; STATUS/FETCH carecem de incidência integrada. SetWatermark tem chamada em teste sem fluxo produtor de completude; ciclo de disputa/fechamento não está completo. | P1 | R3-FIN-02 | R3-FIN-02-S01/S02/S03 | Financeiro e Core |
| Pulsar escolhe versões ACTIVE atuais por tenant no consumo, em vez de destinos/aplicação congelados na admissão. | P1 | R3-FIN-03 | R3-FIN-03-S01/S02/S03 | Financeiro e Core |

P0 bloqueia qualificação de uso; P1 bloqueia conclusão do escopo. Probabilidade quantitativa não foi medida.
