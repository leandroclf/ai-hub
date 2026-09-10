# Design: Contratos JSON estritos
## Context
b9d0f90ce02aa0c27cad546745153d160ff5867f. Fontes em explore.md. Responsável: Core e Catálogo.
## Goals and Constraints
Corrigir comportamento observado e preservar baseline v4/R2/R3. AGENTS governa ambiente.
## Proposed Architecture
Preservar Go e RawMessage para bytes, mas não inferir tipo JSON por sucesso de Unmarshal em string/json.Number. Exigir documento completo: segundo Decode deve encontrar EOF, e retorno sem mapping não pode devolver bytes que não foram integralmente validados.
Adotar um dialeto explicitamente suportado, com compilação na publicação e runtime compartilhado. Uma biblioteca madura compatível pode reduzir risco frente ao validador manual; escolha depende de suporte, limite de recursos e testes contra dialeto. Não adicionar dependência sem comparação objetiva.
Se um subconjunto for mantido, recusar keyword não suportada recursivamente; não aceitar silenciosamente minimum/pattern/format/refs. Resolver refs apenas de origem permitida sem acesso arbitrário à rede. Semântica integer/enum deve ser matemática com precisão exata no dialeto escolhido; diferenciar representação lexical de valor. Profundidade/custo/bytes têm limites testados.
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
| R4-CTR-01 | hub/internal/atlas/offers.go, hub/internal/atlas/offers.go | Documento JSON único e tipos sem coerção |
| R4-CTR-02 | hub/internal/atlas/offers.go, hub/internal/atlas/offers.go, hub/internal/atlas/catalog.go | Dialeto de schema publicado e semântica numérica |

## Main Flows
Executar os cenários de specs/r4-02-contratos-json-estritos/spec.md no caminho real.
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
