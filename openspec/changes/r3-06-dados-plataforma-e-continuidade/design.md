# Design: Dados, plataforma e continuidade
## Context
a4a876a9f8e875db882f7ca45cf7dece24d57aee; fontes em explore.md. Responsável: Plataforma, Dados e SRE.
## Goals and Constraints
Implementar comportamentos completos, preservar v4/R2 e impedir perda/duplicação de efeito.
## Proposed Architecture
Preservar autoridade PostgreSQL, object storage para bytes e filas para transporte. Bancos core/control/finance possuem donos e papéis mínimos; testar RLS com credencial runtime sem BYPASSRLS/ownership e contexto transacional para não vazar por pool. Migrações expand/contract; leitores finais podem usar réplica/materialização somente com consistência e autorização documentadas.
Compose local e kind completo são dois perfis separados. Em kind, dependências devem ter DNS/volumes/health no cluster; não ponte para IP de Docker Compose. Remotos podem usar AWS gerenciada com IaC equivalente. Emulador não substitui teste de semântica AWS real antes de produção.
HPA escala pods, autoscaler de nós escala nós; limites/placement/capacidade de banco também precisam controlador. Probes expressam capacidades; liveness não reinicia por falha transitória de dependência. Shutdown para admissão, drena HTTP/workers e conserva leases.
Prometheus/Loki/Grafana/Tempo/Alloy devem produzir e consultar telemetria real de domínio. Separar overhead Hub, espera de capacidade, rede, provedor, persistência e entrega. Testar Redis fora, broker fora em SYNC, failover/restore e isolamento de tenant sob carga.
## Technical Decisions
Preservar Go, PostgreSQL e a stack atual porque a lacuna principal é integração e prova, não linguagem.
DTOs tipados e validação de runtime nas fronteiras; transações para invariantes locais; outbox/inbox para handoff.
## Alternatives Considered
Trocar broker/linguagem não corrige custódia, contratos ou autorização. Estado somente em Redis/memória não cumpre aceite durável. Duplicar regras no frontend cria divergência.
## Affected Components
| Requisito | Componentes | Motivo |
|---|---|---|
| R3-OPE-01 | hub/internal/objectstore/catalog.go, hub/internal/objectstore/retention.go, hub/internal/cometa/executor.go, hub/internal/orbita/admission.go | Objetos conectados ao fluxo e retenção |
| R3-OPE-02 | hub/deploy/r2/kind/render-runtime.py, hub/deploy/r2/k8s/base/workloads.yaml, hub/deploy/r2/k8s/overlays/prd/kustomization.yaml | Kind completo e ambientes reprodutíveis |
| R3-OPE-03 | hub/deploy/r2/k8s/base/workloads.yaml, hub/internal/platform/httpserver | Escala e continuidade com envelope |
| R3-OPE-04 | hub/migrations/core, hub/migrations/control, hub/internal/platform/pg, hub/internal/orbita/admission.go | Autoridade durável, isolamento e restore |
| R3-OPE-05 | hub/internal/platform/httpserver/telemetry.go, hub/admin-ui/src/pages/OperationsPage.tsx, hub/deploy/r2 | SLA bilateral e telemetria verificável |

## Main Flows
Aplicar cada fluxo GIVEN/WHEN/THEN de specs/06-dados-plataforma-e-continuidade/spec.md com armazenamento e chamadas reais.
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
