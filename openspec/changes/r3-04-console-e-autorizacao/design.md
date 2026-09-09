# Design: Console e autorização
## Context
a4a876a9f8e875db882f7ca45cf7dece24d57aee; fontes em explore.md. Responsável: Frontend e Segurança.
## Goals and Constraints
Implementar comportamentos completos, preservar v4/R2 e impedir perda/duplicação de efeito.
## Proposed Architecture
React/TypeScript continuam como stack. Modelar DTOs por endpoint com validação em runtime, erros de campo e estado de intenção. O backend é autoridade de autorização, ator e idempotência. Não derivar permissão de visibilidade de menu.
Cada página deve completar listar/paginar/pesquisar/detalhar/criar versão/validar/publicar/suspender conforme domínio. Datas transitam como instantes ISO com offset; o editor exibe fuso e preserva o instante. Dinheiro permanece string decimal.
As ações de protocolo, entrega e financeiro têm semânticas distintas e IDs distintos. O operador global possui identidade nominal, leitura autorizada, MFA e auditoria. Contas consumidoras não ganham admin por compartilhar protocols:read.
## Technical Decisions
Preservar Go, PostgreSQL e a stack atual porque a lacuna principal é integração e prova, não linguagem.
DTOs tipados e validação de runtime nas fronteiras; transações para invariantes locais; outbox/inbox para handoff.
## Alternatives Considered
Trocar broker/linguagem não corrige custódia, contratos ou autorização. Estado somente em Redis/memória não cumpre aceite durável. Duplicar regras no frontend cria divergência.
## Affected Components
| Requisito | Componentes | Motivo |
|---|---|---|
| R3-ADM-01 | hub/internal/orbita/admin.go, hub/internal/platform/auth/auth.go | Fronteira administrativa e aplicação |
| R3-ADM-02 | hub/admin-ui/src/pages/OperationsPage.tsx, hub/internal/pulsar/handlers.go, hub/internal/orbita/admin.go | Ações operacionais com API efetiva |
| R3-ADM-03 | hub/admin-ui/src/pages/FinancePage.tsx, hub/internal/libra/handlers.go | Comandos financeiros compatíveis |
| R3-ADM-04 | hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src | Formulários completos e tempo local |

## Main Flows
Aplicar cada fluxo GIVEN/WHEN/THEN de specs/04-console-e-autorizacao/spec.md com armazenamento e chamadas reais.
## Error Flows
Erro antes de custódia não gera ACK/aceite falso; erro após custódia deixa obrigação recuperável. UNKNOWN não autoriza reenvio cego. Erro de autorização não revela dado.
## API / Contract Design
Versionar DTO e perfil, preservar UUIDv7 e escopo tenant/application/cell. Definir método, path, schema, status, idempotência, paginação e erro no contrato antes de ligar a UI. Ver docs/reviews/2026-09-08-r3/03-CONSOLE-E-CONTRATOS.md.
## Data Model and Persistence
Não compartilhar escrita entre donos de domínio. Campos de snapshot/hash/epoch/evidence precisam de índices e constraints; backfill incremental validado. Detalhamento específico acima.
## Authentication and Authorization
Identidade autenticada em cada fronteira, autorização por recurso, papéis mínimos e trilha nominal. Headers fornecidos pelo consumidor não definem tenant.
## Security and Privacy
Limitar payload/custo, sanear logs, segredos somente por referência e cache autorizado. Evidência usa dados sintéticos.
## Observability
Logs estruturados com motivo/identidade não sensível; métricas limitadas; trace de fluxo. Alertas de obrigação, falha de handoff e saturação apontam runbook. Provar emissão/consulta real.
## Testing Strategy
Unitário para regras; integração em PostgreSQL/broker/objetos; contrato na fronteira; E2E para jornada; concorrência/falhas para custódia; carga para budgets; segurança negativa para isolamento. Não substituir integração por mocks.
## Migration Strategy
Expandir schema/DTO, migrar dados com contagem/hash, habilitar por canário, reconciliar obrigações e só então retirar caminho antigo.
## Rollback Plan
Desabilitar nova admissão da capacidade, conservar leitura/recuperação, drenar workers com fencing e voltar imagem compatível com schema expandido. Nunca apagar outbox/ledger/recibos para rollback.
## Compatibility
Nenhum requisito v4/R2 é removido. Alteração incompatível exige versão/migração e decisão explícita; não reinterpretar prazo de commit.
## Operational Considerations
Credenciais/quotas comerciais podem impedir homologação externa, mas não bloqueiam fixtures e restante do trabalho independente.
## Remaining Risks
Ver risk-matrix.md. Aprovação comercial não é inferida de código funcionando.
## Open Questions
Registrar conflitos concretos no arquivo de decisões; prosseguir com itens independentes.
