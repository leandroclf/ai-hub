# Riscos: Console administrativo completo e orientado às jornadas

Estado de todos os riscos: **aberto; mitigação especificada, não implementada/qualificada nesta revisão**. P0 bloqueia exposição/ativação do fluxo afetado; P1 bloqueia completude da capacidade; P2 qualifica usabilidade/governança. Classificação está vinculada ao código e ao uso proposto, sem afirmar exploração em ambiente remoto.

| ID/prioridade | Risco e impacto | Mitigação/prova exigida | Responsável | Evidência |
|---|---|---|---|---|
| F-20 / P1 | Operador não encontra nem administra a população real de clientes, ofertas e provedores; refresh elimina o histórico visual. | Listagens do servidor, filtros paginados, URLs retomáveis e detail/edit auditáveis; refresh e troca de operador não dependem de dados da sessão anterior. | Produto, Frontend e UX | [hub/admin-ui/src/App.tsx:7](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L7) |
| F-21 / P1 | Evoluir só a aparência dos quatro formulários não atende o escopo operacional do Hub. | Executar as jornadas ADM-01 a ADM-12 documentadas, com APIs autoritativas, papéis, estados de erro, observabilidade e trilha de ações. | Produto, Frontend e UX | [hub/admin-ui/src/App.tsx:9](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L9) |
| F-22 / P2 | Operação manual fica propensa a erro e depende de conhecimento de endpoints/IDs. Acessibilidade não foi qualificada. | Controles de escolha e ajuda de negócio, validação coerente servidor/UI, foco e mensagens acessíveis, confirmação de impacto e estados draft/published distintos. | Produto, Frontend e UX | [hub/admin-ui/src/pages/ServicesPage.tsx:81](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/ServicesPage.tsx#L81) |

Risco de integração: dependências r2-01-identidade-e-isolamento. Ver também riscos transversais de identidade, custódia e rollback no design.
