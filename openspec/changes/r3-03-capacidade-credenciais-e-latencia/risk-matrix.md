# Risk matrix

| Risco | Prioridade | Mitigação | Prova | Responsável |
|---|---|---|---|---|
| Capacity possui código/testes, mas Execute/requestPoll não adquirem concessão nem realimentam o controlador. Relatório R2 admite a desconexão. | P1 | R3-INT-01 | R3-INT-01-S01/S02/S03 | Integrações e Plataforma |
| Bearer resolve segredo antes do L1, mantém mutex do TokenCache durante Redis/OAuth e grava bearer token no Redis, contrariando a restrição R2. | P1 | R3-INT-02 | R3-INT-02-S01/S02/S03 | Integrações e Plataforma |
| NewClient cria Transport por submit/poll/OAuth. MaxConnsPerHost em transports distintos não limita consumo agregado e impede reaproveitamento eficiente. | P1 | R3-INT-03 | R3-INT-03-S01/S02/S03 | Integrações e Plataforma |

P0 bloqueia qualificação de uso; P1 bloqueia conclusão do escopo. Probabilidade quantitativa não foi medida.
