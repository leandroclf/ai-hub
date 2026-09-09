# Design: Qualificação integral
## Context
a4a876a9f8e875db882f7ca45cf7dece24d57aee; fontes em explore.md. Responsável: Engenharia e Qualidade.
## Goals and Constraints
Implementar comportamentos completos, preservar v4/R2 e impedir perda/duplicação de efeito.
## Proposed Architecture
Manter v4/R2 como baseline histórica até encerramento conforme OpenSpec. Novas capabilities R3 usam ADDED porque openspec/specs ainda não possui baseline canônica; isso não remove requisitos anteriores. Matriz da união é gate obrigatório.
Construir harness determinístico com provider-sim e webhook-sink como oráculos externos de contagem/bytes, não apenas mocks internos. Subir identidade limpa sem reutilizar volume corrompido silenciosamente; isolar execuções e preservar dados do usuário.
CI deve separar unitário de integração obrigatória. Configuração ausente falha o gate integrado. Evidência contém SHA, digest de imagens, comando, ambiente, versões, timestamps, contagem pass/fail/skip e hash de logs saneados. Browser testa autorização negativa e efeitos reais da UI. Não usar screenshots como única prova de persistência/financeiro.
## Technical Decisions
Preservar Go, PostgreSQL e a stack atual porque a lacuna principal é integração e prova, não linguagem.
DTOs tipados e validação de runtime nas fronteiras; transações para invariantes locais; outbox/inbox para handoff.
## Alternatives Considered
Trocar broker/linguagem não corrige custódia, contratos ou autorização. Estado somente em Redis/memória não cumpre aceite durável. Duplicar regras no frontend cria divergência.
## Affected Components
| Requisito | Componentes | Motivo |
|---|---|---|
| R3-QUA-01 | hub/internal/atlas/catalog_test.go, hub/internal/cometa/custody_test.go, docs/reviews/2026-09-07-r2/implementation/FINAL_REPORT.md | Qualificação integral do SHA sem skips ocultos |

## Main Flows
Aplicar cada fluxo GIVEN/WHEN/THEN de specs/07-qualificacao-integral/spec.md com armazenamento e chamadas reais.
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
