# Explore: Console administrativo completo e orientado às jornadas

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-20 | P1 | Console limitado a quatro formulários e histórico efêmero | [hub/admin-ui/src/App.tsx:7](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L7) |
| F-21 | P1 | Faltam jornadas administrativas de produto, operação e financeiro | [hub/admin-ui/src/App.tsx:9](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L9) |
| F-22 | P2 | Formulários expõem detalhes técnicos e estados pouco guiados | [hub/admin-ui/src/pages/ServicesPage.tsx:81](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/ServicesPage.tsx#L81) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Permitir que operadores configurem, publiquem, diagnostiquem e conciliem o Hub por jornadas completas, com dados do servidor e permissões efetivas.

Dependências: r2-01-identidade-e-isolamento

## Riscos e perguntas técnicas

APIs operacionais dependem das changes 02/03/04/06/07/08. P-09 nomeia administradores. É possível construir navegação e listas cedo, mas nenhuma tela com mock será aceita como integração concluída.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
