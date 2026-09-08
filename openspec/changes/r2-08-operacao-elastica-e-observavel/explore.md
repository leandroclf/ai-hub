# Explore: Ambientes completos, disponibilidade, escala e observabilidade

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-14 | P1 | Redis e broker bloqueiam inclusive reinício do caminho SYNC | [hub/cmd/cometa/main.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/cometa/main.go#L43) |
| F-31 | P1 | Docker local não inclui toda a plataforma e não garante persistência na recriação | [hub/deploy/docker-compose.yml:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L3) |
| F-32 | P1 | Kubernetes é referência incompleta, sem os cinco ambientes | [hub/deploy/k8s/README.md:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/README.md#L3) |
| F-33 | P1 | Autoscaling de células e reconciliação de IaC não fecham o circuito | [hub/migrations/control/0001_init.sql:62](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/control/0001_init.sql#L62) |
| F-34 | P1 | Clientes AWS e bootstrap são fixos ao ambiente local | [hub/internal/queue/queue.go:58](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/queue/queue.go#L58) |
| F-35 | P1 | Probes não acompanham capacidades e não há drenagem graciosa | [hub/internal/platform/httpserver/httpserver.go:86](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/httpserver.go#L86) |
| F-36 | P1 | Observabilidade não mede SLA, pressão nem custódia | [hub/internal/platform/httpserver/metrics.go:16](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/metrics.go#L16) |
| F-42 | P1 | Recepção e pools não limitam custo por tenant | [hub/internal/orbita/handlers.go:114](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L114) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Subir a plataforma completa, escalar com isolamento e manter operações elegíveis durante falhas de dependências não autoritativas.

Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## Riscos e perguntas técnicas

P-01/P-08/P-10/P-11 bloqueiam ativação cloud/crítica, não a entrega do laboratório completo. Distinguir kind como laboratório de Kubernetes gerenciado remoto.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
