# Design: Produtos, prazos e compensação duráveis
## Context
f87ce33034ae29c9431b1910dcc6a633b545e330; auditoria incremental, sem apagar implementação qualificada.
## Goals and Constraints
Fechar fronteiras documentadas. Modelo de falhas, compatibilidade e decisão comercial não podem ser ocultados.
## Proposed Architecture
### R5-EXE-01 — Horizonte de retry e SLA verificados antes do despacho
Unificar semântica em modelo/contrato/UI, sem renomear prazo para mascarar divergência. Claim e transição de expiração transacionais; fencing imediatamente antes de egress com budget restante. Compatibilidade de comandos sem RetryTTLSeconds. T-R2-01 continua investigação normativa: não trocar confirmação durável no prazo por checagem pré-commit.
### R5-EXE-02 — Limite de execução de produto atômico e retomável
Usar lock do plano ou contador de slots com constraint/epoch e transação única intent+step. Separar vaga do workflow e permit de transporte. RowsAffected deve validar posse/transição. Desativar/remover helper não usado ou mantê-lo qualificado explicitamente; sua correção não substitui teste de dois publishers no PostgreSQL.
### R5-EXE-03 — Contrato, provedor e incidência próprios por etapa
Resolver ofertas de etapas na publicação/admissão e persistir sub-snapshots completos, não reutilizar conta do produto por conveniência. Hash canônico do plano e filhos. Compra por operação/attempt e venda por produto conforme contrato. Qualificar adapter externo distinto do simulador com dois oráculos independentes.
### R5-EXE-04 — Compensação tem ordem causal e prazo próprio
Modelar saga de compensação vinculada ao plano, com dependências reversas e estados próprios. Não depender do contexto HTTP nem do status terminal do cliente. Definir destino financeiro de custo/estorno de compensação por contrato; ausência de regra real é gate comercial delimitado, fixtures permitem engenharia.
### R5-EXE-05 — Topologia de mensagens pronta antes de publicar obrigações
Topologia gerenciada por bootstrap único/IaC idempotente com manifesto versionado e gate do publisher. Readiness HTTP e readiness de publicação separados. Não confundir sucesso Publish com consumidor inscrito. Validar ARN/queue policy/DLQ/redrive e namespace/célula. Ensaio em namespace vazio sem destruir dados existentes.
## Technical Decisions
Preservar Go/PostgreSQL/React-TypeScript/SNS-SQS/S3. Justificativas e limites por componente no documento 03.
Transações protegem invariantes locais; transporte e oráculo externo fecham o fluxo distribuído. O nome da tecnologia não garante entrega.
## Alternatives Considered
Evitar reescrita de stack e correção só em helper. Desabilitar validação/auth/checksum ou diminuir assertion não corrige requisito.
Resolução manual permanente de obrigação não satisfaz autonomia operacional; intervenção fica para casos de incerteza irredutível com evidência.
## Affected Components
| Requisito | Arquivos | Responsável |
|---|---|---|
| R5-EXE-01 | hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/cometa/polling_custody.go, hub/admin-ui/src/pages/CatalogPage.tsx | Core e Integrações |
| R5-EXE-02 | hub/internal/orbita/intents.go, hub/internal/orbita/product_store.go, hub/internal/atlas/executor.go | Core e Integrações |
| R5-EXE-03 | hub/internal/orbita/product_plan.go, hub/internal/orbita/product_plan.go, hub/internal/atlas/offers.go | Core e Integrações |
| R5-EXE-04 | hub/internal/orbita/product_store.go, hub/internal/orbita/product_plan.go, hub/internal/orbita/intents.go | Core e Integrações |
| R5-EXE-05 | hub/cmd/orbita/main.go, hub/cmd/cometa/main.go, hub/cmd/libra/main.go, hub/internal/queue/bootstrap.go | Core e Integrações |

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
