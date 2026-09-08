# Explore: Arquivos, propriedade de dados, retenção e recuperação

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-29 | P1 | Arquivos grandes não participam dos fluxos reais | [hub/internal/objectstore/objectstore.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/objectstore/objectstore.go#L43) |
| F-30 | P1 | Persistência sem isolamento de papéis, expurgo e auditoria durável | [hub/deploy/postgres-init/01-init.sh:8](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/postgres-init/01-init.sh#L8) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Manter conteúdo volumoso fora do caminho de mensagens, preservando autorização, integridade, vínculo durável e recuperação reconciliada.

Dependências: r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados

## Riscos e perguntas técnicas

P-04 define classes/regiões/retenção; P-08 define domínio de recuperação. Nenhum período legal ou direito de custódia é presumido pela revisão.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
