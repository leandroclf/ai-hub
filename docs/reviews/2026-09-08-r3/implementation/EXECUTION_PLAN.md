# Execução R3 — plano operacional

## Estado inicial

- HEAD inicial: `a4a876a9f8e875db882f7ca45cf7dece24d57aee`.
- Branch: `codex/r3-implementacao-integral`.
- Alterações preservadas: remoção local de `IMPLEMENTATION_AUDIT.md` e pacote R3 não rastreado recebido.
- Baseline executada: `cd hub && go test ./...` passou; a qualificação integrada ainda contém testes condicionais não executados.
- `openspec/config.yaml` não existe neste checkout; o CLI/configuração efetiva ainda precisa ser identificado.

## Sequência e checkpoints

1. Contratos e harness: matriz, cenários, precisão/schema e comandos reprodutíveis.
2. Custódia e autorização: replay, callback, fencing, topologia e isolamento.
3. Integração real: adapters, autenticação, polling/callback, capacidade e pools.
4. Catálogo/execução: política efetiva, DAG, snapshot e representação congelada.
5. Financeiro/entregas/console: reserva, incidência, destinos e jornadas reais.
6. Dados/operação: objetos, Compose/kind, telemetria, restore e rollback.
7. Qualificação integral: v4 + R2 + R3, sem skip obrigatório, com evidência do SHA atual.

Cada etapa exige código conectado, teste real e evidência atualizada. Decisões D-01…D-07 e T-R2-01 permanecem gates externos até prova ou decisão nominal.
