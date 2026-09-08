# Explore: Custódia, execução única e resultado final correto

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-04 | P0 | 202 após falha de envio de comando, sem retomada de despacho | [hub/internal/orbita/handlers.go:178](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L178) |
| F-05 | P0 | Consumidores removem mensagens mesmo quando o efeito falha | [hub/internal/cometa/worker.go:34](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/worker.go#L34) |
| F-06 | P0 | Resultado retornado não é o resultado do provedor | [hub/internal/cometa/executor.go:120](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L120) |
| F-07 | P0 | Concorrência pode disparar mais de uma operação externa | [hub/internal/cometa/store.go:32](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/store.go#L32) |
| F-08 | P0 | Estado externo e fato são gravados em transações separadas | [hub/internal/cometa/executor.go:192](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L192) |
| F-09 | P0 | Corrida temporal permite sucesso depois do deadline | [hub/internal/orbita/store.go:204](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/store.go#L204) |
| F-11 | P1 | SYNC, AUTO e UUID em erros não cumprem todo o contrato | [hub/internal/orbita/handlers.go:168](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L168) |
| F-28 | P1 | Resultado final não é materializado como representação imutável completa | [hub/internal/orbita/finalize.go:49](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/finalize.go#L49) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Uma obrigação recuperável por aceite, um único dono de despacho e uma representação final proveniente do provedor e durável antes da resposta.

Dependências: r2-01-identidade-e-isolamento

## Riscos e perguntas técnicas

P-07 fecha SLAs e elegibilidade temporal comercial. A proposta conserva deadline rígido da v4; não inventa tolerância para resultado tardio.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
