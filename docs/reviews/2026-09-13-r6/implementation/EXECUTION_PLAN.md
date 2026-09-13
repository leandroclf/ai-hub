# Plano de execução — R6 (estado vivo)

Ordem herdada de `02-PLANO-DE-IMPLEMENTACAO.md`. Esta tabela é atualizada a cada sessão; não é um plano fixo.

| Ordem | Fatia | Estado | Nota |
|---|---|---|---|
| 0 | Fixar candidato e laboratório | CONCLUÍDO | HEAD confirmado `a540b40`; `hub-local` (Compose já ativo) reaproveitado; nenhum recurso novo criado. |
| 1a | R6-SEG-01 — mecanismo (policies + helpers Go + prova real) | CONCLUÍDO | Ver `CHECKPOINT.md`. Testado contra Postgres real, não simulado. |
| 1b | R6-SEG-01 — adoção no caminho real (call sites + corte de credencial) | CONCLUÍDO | Todos os 5 serviços (orbita/cometa/atlas/pulsar/libra, 132 call sites) convertidos. Corte de credencial `hub`→`hub_runtime` aplicado em compose/main.go. 178/178 testes reais verdes sob `hub_runtime`. Pendente não bloqueante: teste de carga de pool de conexão única (S02) e runbook/métricas (S02/D). |
| 1c | R6-SEG-02 (custódia de callback) | EM ANDAMENTO | Rotação de chave (S01) implementada e provada real. Falta S02 (retry sob 3 falhas transitórias) e S03 (retenção/alertas formais, parcial). |
| 1d | R6-FIN-01 (replay financeiro) | NÃO INICIADO | `hub/internal/libra/consumers.go`, `store.go`, `handlers.go`. |
| 2 | R6-EXE-01…04 | NÃO INICIADO | `hub/internal/orbita/{product_plan,intents,product_store}.go`, `hub/internal/cometa/executor.go`. |
| 3 | R6-FIN-02, R6-OPE-04 | NÃO INICIADO | |
| 4 | R6-UX-01/02 | NÃO INICIADO | `hub/admin-ui/`. |
| 5 | R6-OPE-02/03 | NÃO INICIADO | Requer Docker/kind; laboratório disponível localmente (ver CHECKPOINT). |
| 6 | R6-OPE-01, R6-QUA-01 | NÃO INICIADO | |

## Próximo item executável
R6-SEG-02-S02: cenário de 3 falhas transitórias no apply de um callback já aceito (`internal/cometa/custody.go` `conserveObservation`/`ConserveRecoveryObligation` + `ReconcileCallbackInboxBatch`'s `processing_attempts`) — restaurar dependência e reprocessar sem nova emissão do provedor. Depois seguir para a fatia 2 (R6-EXE-01…04).
