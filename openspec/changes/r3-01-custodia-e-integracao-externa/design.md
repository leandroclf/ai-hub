# Design: Custódia e integração externa
## Context
a4a876a9f8e875db882f7ca45cf7dece24d57aee; fontes em explore.md. Responsável: Core e Integrações.
## Goals and Constraints
Implementar comportamentos completos, preservar v4/R2 e impedir perda/duplicação de efeito.
## Proposed Architecture
Orbita possui protocolo, intenção, deadline e representação; Cometa possui tentativas, correlação e recibos externos; o broker transporta fatos/comandos. A chave de efeito deve ser estável por operação/etapa. Um grant não substitui uma tentativa durável. Persistir intenção antes de egress; persistir recibo/fato atomicamente; confirmar origem só depois. Registrar UNKNOWN separadamente do final de atendimento. O recuperador observa antes de reenviar.
O callback requer rota própria com política da conta (assinatura/mTLS/token homologado), proteção replay e inbox órfã. Não exigir implicitamente token de workload Hub de um provedor externo.
Persistir first_transient_failure_at uma única vez; retry_deadline não renova; TTL zero impede retry. provider_deadline e client_deadline são distintos. O horizonte de reconciliação não reabre protocolo expirado.
SYNC deve usar HTTP direto idempotente Orbita→Cometa e retornar final/erro terminal na mesma requisição, com UUIDv7 recuperável após aceite. Não inserir fila obrigatória nesse caminho. ASYNC utiliza intenção/outbox; AUTO tem espera limitada explícita. Erro antes do aceite não pode fingir protocolo persistido.
Topologia deve ter filas/assinaturas obrigatórias verificadas antes do relay e um reconciliador de obrigações, inclusive para perda de handoff em retenção/restore.
## Technical Decisions
Preservar Go, PostgreSQL e a stack atual porque a lacuna principal é integração e prova, não linguagem.
DTOs tipados e validação de runtime nas fronteiras; transações para invariantes locais; outbox/inbox para handoff.
## Alternatives Considered
Trocar broker/linguagem não corrige custódia, contratos ou autorização. Estado somente em Redis/memória não cumpre aceite durável. Duplicar regras no frontend cria divergência.
## Affected Components
| Requisito | Componentes | Motivo |
|---|---|---|
| R3-EXE-01 | hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, hub/internal/atlasclient/client.go | Adapters executáveis e autenticação completa |
| R3-EXE-02 | hub/internal/cometa/handlers.go, hub/internal/cometa/executor.go, hub/cmd/cometa/main.go | Callback com confirmação de custódia |
| R3-EXE-03 | hub/internal/cometa/custody.go, hub/internal/cometa/executor.go, hub/internal/orbita/intents.go | Recuperação de efeito incerto e fencing |
| R3-EXE-04 | hub/internal/orbita/handlers.go, hub/internal/cometa/polling_custody.go, hub/internal/orbita/finalize.go, docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md | Relógios de retry, polling e prazo final |
| R3-EXE-05 | hub/internal/queue/bootstrap.go, hub/internal/queue/queue.go, hub/internal/libra/consumers.go, hub/internal/libra/store.go | Topologia de mensagens e quarentena recuperável |

## Main Flows
Aplicar cada fluxo GIVEN/WHEN/THEN de specs/01-custodia-e-integracao-externa/spec.md com armazenamento e chamadas reais.
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
