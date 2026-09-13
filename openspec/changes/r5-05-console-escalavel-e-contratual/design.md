# Design: Console escalável e contratual
## Context
f87ce33034ae29c9431b1910dcc6a633b545e330; auditoria incremental, sem apagar implementação qualificada.
## Goals and Constraints
Fechar fronteiras documentadas. Modelo de falhas, compatibilidade e decisão comercial não podem ser ocultados.
## Proposed Architecture
### R5-UX-01 — Console preserva intenção e escala com catálogo
Busca incremental server-side, paginação e cancelamento de requisições antigas. Schemas runtime por DTO e IDs textuais. Identidade de intenção persistida com escopo/payload hash, expiração e mecanismo de consultar recibo; não guardar segredos ou payload sensível indiscriminadamente no browser. Jornada de capacidade/qualificação somente dentro do processo autorizado, com segregação de aprovação. E2E por efeito real e acessibilidade, não só visita a 16 rotas.
## Technical Decisions
Preservar Go/PostgreSQL/React-TypeScript/SNS-SQS/S3. Justificativas e limites por componente no documento 03.
Transações protegem invariantes locais; transporte e oráculo externo fecham o fluxo distribuído. O nome da tecnologia não garante entrega.
## Alternatives Considered
Evitar reescrita de stack e correção só em helper. Desabilitar validação/auth/checksum ou diminuir assertion não corrige requisito.
Resolução manual permanente de obrigação não satisfaz autonomia operacional; intervenção fica para casos de incerteza irredutível com evidência.
## Affected Components
| Requisito | Arquivos | Responsável |
|---|---|---|
| R5-UX-01 | hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/pages/CapacityDomainsPage.tsx | Frontend e Produto |

## Main Flows
Executar cenários S01/S02/S03 pelo ingress/worker/autoridade efetivos, incluindo o consumidor da mudança.
## Error Flows
Falha antes de custódia: recusar sem prometer aceite. Depois de custódia: preservar obrigação/reconciliação.
UNKNOWN não autoriza SUBMIT novo. Final do cliente e estado externo não se confundem.
## API / Contract Design
Especificar path/método/schema/status, idempotência e autorização. Tenant/aplicação/célula vêm de identidade autoritativa.
GET e webhook usam representação congelada. Falhas de transporte, rejeição contratual e resposta inválida têm disposições distintas.
## Data Model and Persistence
Domínio dono, constraints de unicidade/escopo, índices, leases e estado durável. Backfill com inventário de registros existentes.
Não apagar inbox/outbox/ledger por conveniência; payload protegido por retenção e criptografia aplicáveis.
## Authentication and Authorization
Conta externa verificada antes de custódia. Workload para comunicação interna; operador nominal, MFA e escopo global explícito.
## Security and Privacy
Não copiar payloads/segredos reais para evidência. Fixture sintética identificada; redaction testada.
## Observability
Métricas de latência/idade/backlog/disposição por dimensões limitadas, logs correlacionados, traces e alertas exercitados.
Não usar UUID por requisição como label de série. Cenário de alerta deve provar emissão e consulta real.
## Testing Strategy
Unitários adversariais; integração PostgreSQL/objetos/broker; contrato com oráculo externo; browser nas jornadas afetadas.
Concorrência, crash/retomada, expiração e upgrade/rollback são gates do comportamento; skip obrigatório não aprova.
## Migration Strategy
Inventário→expand→backfill verificado→canário→reconciliação→ativação. Testar banco limpo e volumes de R2/R3/R4.
## Rollback Plan
Reverter ativação/imagem compatível; manter schema expandido e obrigações legíveis. Drenar/fencear workers antes de trocar.
Nunca replay cego, apagar volumes ou editar checksum para verde.
## Compatibility
R5 é aditiva: baseline v4/R2/R3/R4 permanece. Ponte dos quatro requisitos R4 ausentes está no documento 06.
## Operational Considerations
Um ecossistema Compose por vez; kind independente já existe e deve ser preservado. Perfis remotos precisam gates próprios.
## Remaining Risks
Probabilidade não quantificada. P0 impede liberação do fluxo afetado até prova; requisito comercial externo não se resolve por código.
## Open Questions
Registrar responsável e próximo passo do gate realmente dependente; continuar tarefas independentes.
