# Decisões, recursos e bloqueios

| item | estado | responsável | impacto | trabalho independente |
|---|---|---|---|---|
| D-01…D-07 | baseline técnica adotada para laboratório/homologação; ratificação de produção pendente | responsáveis funcionais correspondentes | não impede implementação local; impede ativação comercial/operacional definitiva | defaults registrados em `05-DECISOES-E-MIGRACAO.md`, fixtures e harness |
| T-R2-01 | regra conservadora adotada operacionalmente; decisão normativa/técnica final pendente | Core/Dados/SRE | não impede ensaios; impede declarar semântica final do prazo | contraexemplos, relógios separados, fencing e prova de commit |
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

## Baseline executável adotada

Para não paralisar o planejamento, foram escolhidos defaults reproduzíveis para local/homologação: 200 RPS sintéticos por célula, SLA padrão de 30 segundos, reserva de finalização de 5 segundos, retry por unidade econômica efetiva, dados sintéticos classificados, Kubernetes/EKS como alvo de produção, S3 versionado, polling+callback, idempotência externa obrigatória quando houver reenvio e catálogo versionado com planos por unidade/franquia/faixa. Os valores e critérios completos estão em [`05-DECISOES-E-MIGRACAO.md`](../05-DECISOES-E-MIGRACAO.md).

Essas decisões habilitam código e qualificação local. Não autorizam ativação comercial, contratação de provedor, provisionamento pago ou promoção para produção sem ratificação nominal e evidência específica.
