# Design: Callbacks autenticados e recuperáveis
## Context
b9d0f90ce02aa0c27cad546745153d160ff5867f. Fontes em explore.md. Responsável: Integrações e Core.
## Goals and Constraints
Corrigir comportamento observado e preservar baseline v4/R2/R3. AGENTS governa ambiente.
## Proposed Architecture
O ingresso autentica a conta antes de admitir body de órfão. Cometa mantém política de autenticação e escopo do recibo, valida correlação/schema e conserva conteúdo original mais observação normalizada. Capability aleatória é suplemento com rotação/expiração, não identidade da conta. Evitar segredo em query string; caso legado exija, aplicar redaction comprovada em toda infraestrutura e ADR de risco.
Inbox registra origem autenticada, tenant, aplicação quando resolvida, conta, célula, evento do provedor, hash do body, versão de política, disposition, timestamps e retenção. Antes da correlação completa, escopo mínimo autenticado continua obrigatório. Identidade de deduplicação não pode ser apropriada por token inválido. Payload inválido não deve virar observação terminal.
Worker de reconciliação usa lote/claim/epoch/lease, limita tentativas/custo por escopo, fecha cursor antes de chamadas aninhadas e roda independentemente do HTTP. Aceite conhecido só aguarda sua custódia. Eventos inválidos recebem disposição própria; um poison item não bloqueia toda fila.
Uma função de domínio comum valida SUBMIT/poll/callback contra o mesmo snapshot e correlação. Recibo original não equivale a final de atendimento: Orbita continua dona do final/SLA. A operação externa pode permanecer reconciliável após cliente EXPIRED.
## Technical Decisions
Manter stack atual; concentrar correção na fronteira do domínio e fluxo de custódia. Go para backend,
PostgreSQL para invariantes locais, React/TypeScript para UI, SNS/SQS para transporte e S3 para bytes.
Motivos/limites por componente estão no documento de arquitetura R4.
## Alternatives Considered
Correção apenas em helper não fecha fluxo. Desabilitar auth/checksum/schema/skip não é solução.
Reescrita de linguagem não resolve integração nem elimina necessidade de teste adversarial.
## Affected Components
| Requisito | Fontes | Motivo |
|---|---|---|
| R4-CBK-01 | hub/cmd/cometa/main.go, hub/internal/cometa/handlers.go, hub/internal/cometa/executor.go, hub/internal/platform/httpserver/httpserver.go | Rota de callback compatível com a autenticação do provedor |
| R4-CBK-02 | hub/internal/cometa/handlers.go, hub/internal/cometa/custody.go, hub/migrations/core/0032_callback_inbox.sql | Admissão e deduplicação segura da inbox órfã |
| R4-CBK-03 | hub/internal/cometa/custody.go, hub/internal/cometa/handlers.go | Recuperador autônomo limitado e independente do HTTP |
| R4-CBK-04 | hub/internal/cometa/executor.go, hub/internal/cometa/executor.go, hub/internal/cometa/custody.go | Uma validação de resultado para SUBMIT, polling e callback |

## Main Flows
Executar os cenários de specs/r4-01-callbacks-autenticados-e-recuperaveis/spec.md no caminho real.
## Error Flows
Conservar obrigação após aceite; não fabricar sucesso. Divergência de autenticação/schema recebe disposição
segura, sem sobrescrever identidade legítima. Não bloquear outras operações por item inválido.
## API / Contract Design
Documentar path/método/media type/schema/versão/status/idempotência/autorização antes de ligar consumidor.
Preservar escopo tenant/application/cell; GET e webhook seguem representação congelada.
## Data Model and Persistence
Propriedade de domínio, índices/constraints e validade explícitos. Recibo bruto protegido separado de observação normalizada.
Alteração de migração exige plano que preserve checksums conhecidos e dados; rollback não apaga obrigação.
## Authentication and Authorization
Identidade da conta no ingresso externo; workload nas APIs internas; admin nominal com MFA/escopo.
Não usar headers fornecidos pelo consumidor como identidade autoritativa.
## Security and Privacy
Segredos fora de logs/URLs observadas; cache conforme versão/revogação; payload/custo/retention limitados.
## Observability
Métricas de pendência/idade/claim/conflito/cache/latência sem cardinalidade irrestrita.
Logs correlacionados e saneados; alertas devem ser demonstrados em séries reais.
## Testing Strategy
Unitários adversariais, contratos, DB durável, oráculos externos, browser, concorrência/falhas,
upgrade/rollback e carga do cenário. Teste não executado permanece não executado.
## Migration Strategy
Inventariar schema e versão atuais; expandir, backfill validado, habilitar canário e reconciliar.
Preservar caminho de leitura/recuperação de obrigações antigas.
## Rollback Plan
Voltar ativação/imagem compatível, drenar workers com fencing e manter schema expandido.
Não apagar dados ou reexecutar efeito externo para limpar estado.
## Compatibility
Capabilities R4 são ADDED porque baseline canônica openspec/specs ainda não existe.
Não removem nem relaxam obrigações herdadas.
## Operational Considerations
Um ecossistema Compose, ferramentas fixadas e evidência do conteúdo atual.
Sem runtime real, preparar gate e concluir trabalho independente sem declarar qualificação.
## Remaining Risks
Ver matriz. P0 impede homologação; decisões comerciais/SLA normativo não podem ser aprovadas pelo agente.
## Open Questions
Registrar somente conflitos concretos e seus responsáveis; não usar lista genérica como motivo para parar.
