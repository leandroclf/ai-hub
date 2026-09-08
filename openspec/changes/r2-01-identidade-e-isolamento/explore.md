# Explore: Identidade, isolamento de tenants e administração segura

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-01 | P0 | Identidade do cliente controlada pelo próprio chamador | [hub/internal/orbita/handlers.go:95](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L95) |
| F-02 | P0 | APIs internas e administrativas sem identidade de workload | [hub/internal/orbita/handlers.go:53](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L53) |
| F-41 | P1 | Destinos configuráveis não possuem defesa SSRF | [hub/internal/atlas/handlers.go:100](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/handlers.go#L100) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Fechar a cadeia identidade → autorização por recurso → auditoria, da borda até os domínios internos.

Dependências: Sem pré-requisito de implementação para iniciar; integrações específicas descritas abaixo.

## Riscos e perguntas técnicas

P-09 define responsáveis e perfis nominais; P-04 define dados permitidos. Essas pendências não impedem eliminar confiança em headers arbitrários.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
