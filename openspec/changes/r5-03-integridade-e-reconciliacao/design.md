# Design: Integridade de dados e reconciliação
## Context
f87ce33034ae29c9431b1910dcc6a633b545e330; auditoria incremental, sem apagar implementação qualificada.
## Goals and Constraints
Fechar fronteiras documentadas. Modelo de falhas, compatibilidade e decisão comercial não podem ser ocultados.
## Proposed Architecture
### R5-DAD-01 — Precisão do resultado preservada em todas as fronteiras
RawMessage para payloads opacos e Decoder.UseNumber ou tipos decimais exatos para transformação efetiva. Auditar cada Unmarshal em any, inclusive dispatch.Result e consolidação. Testes de round-trip nas fronteiras reais, não só TransformJSON. API frontend deve preservar IDs como strings com migração compatível.
### R5-DAD-02 — Restore populado e reconciliação antes de retomada
Povoar cenários e medir antes/depois. Backups por domínio exigem protocolo de consistência/reconciliação, não só dumps em sequência. Preservar version IDs ou mapear novas versões com integridade verificada. Reconciliar no destino restaurado e ligar workers após barreira. Não deletar volumes originais nem simular garantia regional com kind.
### R5-DAD-03 — Liquidação por valor efetivo e completude demonstrável
Separar valor reservado, capturado e liberado; custódia e idempotência por unidade econômica. Protocolo de watermark por partição/produtor com prova de drain e relógio/corte explícitos. Ligar disputa e ajuste ao fato tardio. Testar franquia/strict balance antes do efeito, múltiplas etapas, compensação e credencial CLIENT_DIRECT.
### R5-DAD-04 — Quarentena financeira conserva mensagem recuperável
Armazenar envelope bruto protegido ou referência objeto com hash/versão/escopo, junto ao recibo de quarentena. Identificar também falhas de parsing de envelope SNS/SQS na borda antes de ProcessEnvelope. Implementar replay de domínio sem reenviar SUBMIT. Migração aditiva; registros antigos só-hash exigem diagnóstico de recuperação a partir de fonte externa, nunca reconstrução inventada.
## Technical Decisions
Preservar Go/PostgreSQL/React-TypeScript/SNS-SQS/S3. Justificativas e limites por componente no documento 03.
Transações protegem invariantes locais; transporte e oráculo externo fecham o fluxo distribuído. O nome da tecnologia não garante entrega.
## Alternatives Considered
Evitar reescrita de stack e correção só em helper. Desabilitar validação/auth/checksum ou diminuir assertion não corrige requisito.
Resolução manual permanente de obrigação não satisfaz autonomia operacional; intervenção fica para casos de incerteza irredutível com evidência.
## Affected Components
| Requisito | Arquivos | Responsável |
|---|---|---|
| R5-DAD-01 | hub/internal/orbita/factconsumer.go, hub/internal/orbita/factconsumer.go, hub/internal/orbita/product_store.go, hub/admin-ui/src/pages/FinancePage.tsx | Dados e Financeiro |
| R5-DAD-02 | hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh | Dados e Financeiro |
| R5-DAD-03 | hub/internal/libra/store.go, hub/internal/libra/store.go, hub/internal/libra/settlement.go, hub/internal/libra/store.go | Dados e Financeiro |
| R5-DAD-04 | hub/internal/libra/store.go, hub/internal/libra/consumers.go | Dados e Financeiro |

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
