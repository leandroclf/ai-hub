# Status dos achados R3 e históricos

## Regra de encerramento

`ABERTO` é o estado inicial. Um achado só muda após teste reproduzível no SHA atual, evidência referenciada e revalidação dos cenários herdados. Logs históricos não são reutilizados como prova nova.

## R3

Os 25 achados F-R3-01…F-R3-25 permanecem `ABERTO` até a execução correspondente. O primeiro defeito reproduzido foi F-R3-06; a implementação adicionou regressão positiva para precisão numérica e enum, mas a qualificação integrada e o restante do requisito ainda estão pendentes.

## Histórico

Os 42 achados R2 permanecem `REVALIDACAO_PENDENTE`. A execução anterior tem valor histórico e não qualifica este SHA.
