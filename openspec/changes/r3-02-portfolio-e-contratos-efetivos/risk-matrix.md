# Risk matrix

| Risco | Prioridade | Mitigação | Prova | Responsável |
|---|---|---|---|---|
| Provas executadas: TransformJSON altera 9007199254740993 para 9007199254740992 e aceita INVALID contra enum [OK]. Usa float64 e validação parcial. | P0 | R3-CAT-01 | R3-CAT-01-S01/S02/S03 | Produto e Core |
| Admissão usa tempos/modos do target sem composição efetiva de overrides. service_version recebido não participa da resolução. Finalizer usa FinalBody fixo sem OutputMapping. | P1 | R3-CAT-02 | R3-CAT-02-S01/S02/S03 | Produto e Core |
| Catálogo contém steps, mas admissão produz um comando com StepID igual ao protocolo; não há executor integrado do DAG/agregação nesse caminho. | P1 | R3-CAT-03 | R3-CAT-03-S01/S02/S03 | Produto e Core |
| ResolveOffer recusa tenant com mais de 100 ofertas antes de filtrar aplicação/serviço. Caminho Offer consulta Atlas em cada requisição. | P1 | R3-CAT-04 | R3-CAT-04-S01/S02/S03 | Produto e Core |

P0 bloqueia qualificação de uso; P1 bloqueia conclusão do escopo. Probabilidade quantitativa não foi medida.
