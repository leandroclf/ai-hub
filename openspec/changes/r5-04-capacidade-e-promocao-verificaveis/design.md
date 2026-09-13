# Design: Capacidade, ambientes e promoção verificáveis
## Context
f87ce33034ae29c9431b1910dcc6a633b545e330; auditoria incremental, sem apagar implementação qualificada.
## Goals and Constraints
Fechar fronteiras documentadas. Modelo de falhas, compatibilidade e decisão comercial não podem ser ocultados.
## Proposed Architecture
### R5-OPE-01 — Capacidade sem bypass e recuperação de permits
Gate de publicação para CapacityDomain obrigatório; exceção synthetic local limitada por ambiente. Registrar settlement como obrigação reconciliável/outbox, não apenas log. Agregados/indexação/retenção de permits fechados sem apagar evidência. Ensaiar múltiplas réplicas e sinais 429/timeout/latência, recuperação aditiva e mínimo de observação externa.
### R5-OPE-02 — Promoção exige evidência vinculada ao artefato
Manifesto obrigatório para promoção, hashes verificados de logs/imagens e referência ao sistema autorizado de aprovação; shell local pode gerar proposta, não autoridade de aprovação. Separar gate local de CI protegido. Pins por digest e matriz toolchain. Não inventar que tag+digest estão incorretos: verificar registry em CI e construir imagem alvo.
### R5-OPE-03 — Ambientes elásticos com dados duráveis e isolamento completo
Preservar kind independente e um único Compose conforme AGENTS. Criar IaC/valores completos de destino, sem provisionar recursos pagos nesta missão. PV/backup de banco/objetos/broker, disruption budgets, topology spread, HPA/KEDA e node autoscaler com quota. Métricas bilaterais de SLA, alertas exercitados, drenagem HTTP/worker e isolamento de tenant sob carga.
## Technical Decisions
Preservar Go/PostgreSQL/React-TypeScript/SNS-SQS/S3. Justificativas e limites por componente no documento 03.
Transações protegem invariantes locais; transporte e oráculo externo fecham o fluxo distribuído. O nome da tecnologia não garante entrega.
## Alternatives Considered
Evitar reescrita de stack e correção só em helper. Desabilitar validação/auth/checksum ou diminuir assertion não corrige requisito.
Resolução manual permanente de obrigação não satisfaz autonomia operacional; intervenção fica para casos de incerteza irredutível com evidência.
## Affected Components
| Requisito | Arquivos | Responsável |
|---|---|---|
| R5-OPE-01 | hub/internal/cometa/executor.go, hub/internal/cometa/capacity.go, hub/internal/cometa/capacity.go, hub/internal/cometa/executor.go | Plataforma e Operações |
| R5-OPE-02 | hub/deploy/r2/tests/promotion-gate.sh, hub/deploy/r2/tests/validate-qualification-evidence.py, hub/deploy/Dockerfile, hub/deploy/r2/Dockerfile.ui | Plataforma e Operações |
| R5-OPE-03 | hub/deploy/r2/kind/render-independent-dependencies.py, hub/deploy/r2/kind/bootstrap-independent.sh, hub/deploy/r2/k8s/overlays/prd/kustomization.yaml | Plataforma e Operações |

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
