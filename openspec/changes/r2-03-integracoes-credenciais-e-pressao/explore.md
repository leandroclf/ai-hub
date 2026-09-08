# Explore: Adaptadores reais, credenciais, polling, callbacks e pressão

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-03 | P0 | Callback confirma recebimento sem custódia comprovada | [hub/internal/cometa/handlers.go:73](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/handlers.go#L73) |
| F-10 | P1 | TTL de retry configurado não governa execução | [hub/internal/atlas/store.go:40](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/store.go#L40) |
| F-12 | P0 | Credencial dedicada resolvida não é a usada na chamada | [hub/internal/cometa/executor.go:83](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L83) |
| F-13 | P0 | Autenticação do simulador não equivale a integração segura real | [hub/internal/providerauth/client.go:57](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/providerauth/client.go#L57) |
| F-15 | P1 | Polling sem autenticação, lease e orçamento configurável | [hub/internal/cometa/poller.go:48](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/poller.go#L48) |
| F-16 | P1 | Não há amortecimento adaptativo nem isolamento de carga | [hub/internal/cometa/worker.go:34](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/worker.go#L34) |
| F-26 | P1 | Webhook usa URL como identidade do segredo | [hub/internal/pulsar/store.go:64](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/store.go#L64) |
| F-27 | P1 | Webhook sem claim, recibos completos e política por cliente | [hub/internal/pulsar/store.go:105](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/store.go#L105) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Executar contratos homologados do provedor com credencial correta e capacidade adaptativa compartilhada, preservando cada observação e entrega.

Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## Riscos e perguntas técnicas

P-05 contém garantias reais dos parceiros; P-07 define TTL/SLA; P-11 os envelopes máximos. Sem confirmação de idempotência não habilitar retry de efeito ambíguo.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
