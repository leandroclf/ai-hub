# Explore: Medição, saldo estrito, ledger e fechamento

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-23 | P0 | Custo e receita não usam contratos econômicos congelados | [hub/internal/libra/consumers.go:17](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/consumers.go#L17) |
| F-24 | P0 | Saldo estrito não contabiliza consumo já capturado | [hub/internal/libra/store.go:73](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/store.go#L73) |
| F-25 | P1 | Ledger, precisão monetária e fechamento não estão implementados | [hub/internal/libra/store.go:27](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/store.go#L27) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Explicar cada receita, custo, reserva, ajuste e pagamento a partir de contrato e evidência, sem redelivery alterar valor.

Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados, r2-04-catalogo-produtos-e-contratos

## Riscos e perguntas técnicas

P-03 e P-06 aprovam tarifas, incidência, arredondamento e ERP. Corrigir saldo fail-open e congelar contratos não depende da definição de preço comercial real.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
