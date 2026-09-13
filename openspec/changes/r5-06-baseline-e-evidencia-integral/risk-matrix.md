# Riscos

| Risco observado | Prioridade | Mitigação | Dono |
|---|---|---|---|
| Repositório contém R4 original (12 requisitos/36 cenários), não o quinto change da edição regenerada (quatro requisitos adicionais). Baseline real é 201/732; matriz reconhece 231 linhas associadas e 501 sem qualificação. Gerador associa resultado por scenario_id, sem exigir digest do código/spec da execução na linha de origem; vínculo textual não prova atualidade nem PASS. R4 adicional precisa ponte explícita, sem copiar 744 como contagem remota. | P1 | R5-QUA-01 | Engenharia e Qualidade |
