# Design: Capacidade, credenciais e latência
## Context
a4a876a9f8e875db882f7ca45cf7dece24d57aee; fontes em explore.md. Responsável: Integrações e Plataforma.
## Goals and Constraints
Implementar comportamentos completos, preservar v4/R2 e impedir perda/duplicação de efeito.
## Proposed Architecture
Cometa aplica controlador de concorrência antes de todo I/O, inclusive STATUS/FETCH. Agrupamento de capacidade reflete limite real da conta/provedor, não a réplica. Algoritmo deve ter mínimo, teto contratual, janela, EWMA, passo de aumento, fator de redução, cooldown, jitter, Retry-After e métricas; publicar valores no perfil de homologação e medir estabilidade. Não importar valores arbitrários como SLA comercial.
Concessões duráveis com epoch/lease e retenção de incerteza evitam multiplicar capacidade em scale-out. Filas por classe com justiça ponderada impedem um tenant de consumir todos os workers.
Cache de tokens em memória por binding/version tem single-flight por chave; expiração/revogação governadas. Redis só metadados não sensíveis e jamais autoridade. Reutilizar transporte por origem/identidade TLS mantendo prevenção SSRF/DNS rebind/redirect em cada conexão. Não usar pool comum para certificados incompatíveis.
## Technical Decisions
Preservar Go, PostgreSQL e a stack atual porque a lacuna principal é integração e prova, não linguagem.
DTOs tipados e validação de runtime nas fronteiras; transações para invariantes locais; outbox/inbox para handoff.
## Alternatives Considered
Trocar broker/linguagem não corrige custódia, contratos ou autorização. Estado somente em Redis/memória não cumpre aceite durável. Duplicar regras no frontend cria divergência.
## Affected Components
| Requisito | Componentes | Motivo |
|---|---|---|
| R3-INT-01 | hub/internal/cometa/capacity.go, hub/internal/cometa/executor.go, hub/internal/cometa/poller.go | Controle adaptativo conectado a todo I/O |
| R3-INT-02 | hub/internal/providerauth/client.go | Cache de autenticação isolado e sem tokens no Redis |
| R3-INT-03 | hub/internal/platform/egress, hub/internal/cometa/executor.go, hub/internal/cometa/poller.go, hub/internal/providerauth/client.go | Pools HTTP e budgets de concorrência |

## Main Flows
Aplicar cada fluxo GIVEN/WHEN/THEN de specs/03-capacidade-credenciais-e-latencia/spec.md com armazenamento e chamadas reais.
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
