# Design: Financeiro e entregas
## Context
a4a876a9f8e875db882f7ca45cf7dece24d57aee; fontes em explore.md. Responsável: Financeiro e Core.
## Goals and Constraints
Implementar comportamentos completos, preservar v4/R2 e impedir perda/duplicação de efeito.
## Proposed Architecture
Libra é dono do ledger, reservas, franquias, fatos, disputas e períodos. Persistir valores decimais exatos e moeda; evento econômico referencia snapshot, unidade, tentativa e evidência. Uma duplicata de transporte não é novo consumo.
Reserva estrita considera exposição já capturada e retida; franquia estrita precisa reserva prévia. Capture precisa valor efetivo, não apenas troca de status. Expiração gera hold quando há incerteza externa, e reconciliação positiva o resolve.
Watermark prova completude de produtores/obrigações até um corte; relógio sozinho não prova completude. Fechamento é bloqueado por lacuna/disputa, ajustes são compensatórios, exportação é determinística e idempotente.
Pulsar conserva bytes e snapshot de destino por aplicação. Redelivery mantém identidade/representação; troca de destino de protocolo antigo requer política e auditoria explícitas, não seleção dinâmica silenciosa.
## Technical Decisions
Preservar Go, PostgreSQL e a stack atual porque a lacuna principal é integração e prova, não linguagem.
DTOs tipados e validação de runtime nas fronteiras; transações para invariantes locais; outbox/inbox para handoff.
## Alternatives Considered
Trocar broker/linguagem não corrige custódia, contratos ou autorização. Estado somente em Redis/memória não cumpre aceite durável. Duplicar regras no frontend cria divergência.
## Affected Components
| Requisito | Componentes | Motivo |
|---|---|---|
| R3-FIN-01 | hub/internal/libra/store.go, hub/internal/orbita/handlers.go, hub/internal/contracts/economics/publication.go | Reserva estrita e franquia antes do efeito |
| R3-FIN-02 | hub/internal/cometa/custody.go, hub/internal/cometa/polling_custody.go, hub/internal/libra/store.go, hub/internal/libra/handlers.go | Incidência completa e fechamento operacional |
| R3-FIN-03 | hub/internal/pulsar/custody.go, hub/internal/orbita/finalize.go, hub/internal/pulsar/handlers.go | Destino de webhook congelado por aplicação |

## Main Flows
Aplicar cada fluxo GIVEN/WHEN/THEN de specs/05-financeiro-e-entregas/spec.md com armazenamento e chamadas reais.
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
