# Matriz de risco

| Risco | Prioridade | Mitigação | Responsável | Evidência exigida |
|---|---|---|---|---|
| As specs têm 189 requisitos e 696 cenários (416 v4+205 R2+75 R3), mas matriz registra 280, omitindo cenários v4. FINAL_REPORT diz OpenSpec não executado; CHECKPOINT informa strict 17/17 e novas mudanças. Título do merge afirma conclusão, incompatível com gates ainda abertos. | P1 | R4-QUA-01 | Engenharia, Frontend e Qualidade | R4-QUA-01-S01/S02/S03 |
| Nenhum arquivo de hub/admin-ui mudou entre os snapshots. Permanecem a escolha incorreta de delivery_id, SLA/reconcile sem rota efetiva e comandos financeiros incompatíveis. Domínios financeiro, DAG, capacidade, objetos e kind também não receberam fechamento funcional neste delta. | P1 | R4-QUA-02 | Engenharia, Frontend e Qualidade | R4-QUA-02-S01/S02/S03 |
| go test -race passa com 37 testes pass e 18 skip nesta revisão. OpenSpec strict 17/17 passa sem config.yaml; não é bloqueio técnico atual. Não há comprovação nova de DB/Compose/kind/browser/HA. AGENTS atualizado exige um ecossistema Compose por vez, contrariando prompts históricos que sugeriam laboratórios paralelos. | P1 | R4-QUA-03 | Engenharia, Frontend e Qualidade | R4-QUA-03-S01/S02/S03 |

Probabilidade quantitativa não medida. Evidência estática não é incidente observado.
