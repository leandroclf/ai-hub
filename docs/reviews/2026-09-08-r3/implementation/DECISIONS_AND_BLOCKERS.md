# Decisões, recursos e bloqueios

| item | estado | responsável | impacto | trabalho independente |
|---|---|---|---|---|
| D-01…D-07 | não aprovadas neste checkout | Negócio/Engenharia/Operações | impede afirmar envelope comercial/operacional | fixtures sintéticas, contratos e harness |
| T-R2-01 | aberto | Core/Dados/SRE | impede declarar semântica final de prazo | contraexemplos, relógios separados e fencing |
| PostgreSQL/S3/OIDC/kind/browser | disponibilidade a verificar | Engenharia/Plataforma | integração não pode ser marcada como PASS sem runtime | testes unitários, contratos e validações locais |

Nenhum bloqueio externo autoriza reduzir requisito, usar fixture como efeito real ou marcar cenário obrigatório como skip.

## Atualização técnica de 2026-09-09

| item | estado | evidência | consequência |
|---|---|---|---|
| RLS runtime | prova negativa PASS com `hub_runtime`; migrações 0033/0034 aplicadas | `rls-runtime-proof.sh` | falta migrar os serviços para o DSN não proprietário e repetir a suíte completa |
| restore | prova cercada PASS para dumps, contagens, objetos e não-replay observado | `restore-reconciliation.sh`, sufixo `autonomous2` | falta restore populado reconciliando efeitos externos e financeiro |
| executor DAG | componente de execução bounded/compensável PASS em unit/race | `hub/internal/atlas/executor_test.go` | falta conexão ao DAG persistido e jornada administrativa |
| 42 achados | reavaliação individual documentada, nenhum encerrado | `REAVALIACAO-42-INDIVIDUAL.md` | encerramento exige evidência específica por achado |

D-01…D-07 e T-R2-01 continuam decisões que exigem responsáveis externos. O trabalho técnico independente permanece válido e reproduzível com fixtures sintéticas, sem apresentar essas fixtures como contratos comerciais aprovados.
