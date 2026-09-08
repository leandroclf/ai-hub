# Explore: Catálogo operável, produtos compostos e contratos por cliente

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-17 | P1 | Catálogo publicado pode ser sobrescrito e não governa admissão | [hub/internal/atlas/store.go:48](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/store.go#L48) |
| F-18 | P1 | Produtos, DAG e contratos legados ainda não existem | [hub/internal/orbita/handlers.go:23](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L23) |
| F-19 | P1 | Importação não ativa adaptadores e substitui configurações locais | [hub/deploy/import_hiveplace_collection.sh:38](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/import_hiveplace_collection.sh#L38) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Separar serviço, produto, oferta, contrato de integração e adapter; publicar versões validadas que governam admissão e execução paralela.

Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## Riscos e perguntas técnicas

P-02 define portfólio e perfis legados reais; P-03 regras comerciais; P-05 equivalência. Artefatos usam fixtures sintéticas, sem declarar os 232 endpoints homologados.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
