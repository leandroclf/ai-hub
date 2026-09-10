# Design: Upgrade, cache e crescimento
## Context
b9d0f90ce02aa0c27cad546745153d160ff5867f. Fontes em explore.md. Responsável: Plataforma, Dados e Integrações.
## Goals and Constraints
Corrigir comportamento observado e preservar baseline v4/R2/R3. AGENTS governa ambiente.
## Proposed Architecture
Cache quente parte da identidade/versão autorizada conhecida, com validade/revogação explícitas. Buscar cofre só quando necessário; cache expirado/revogado nunca é usado para disponibilidade. Coordenação por chave precisa cancelamento e limite de memória, inclusive mapa de locks durante rotação. Não liberar lock com waiter ativo ao fazer eviction.
Resolver oferta por índice seletivo e vigência; paginação é de listagem, não licença para materializar portfólio no caminho crítico. Snapshot/projeção governados continuam necessários para independência do control plane.
Migrações existentes com checksum são imutáveis. A migração nova deve carregar API_KEY; restaurar histórico exige identificar bancos já criados com a variante alterada. Desenhar reconciliação explícita de hashes conhecidos com validação de schema e auditoria; jamais atualizar todos os checksums ou apagar volumes. Testar upgrade do SHA R2 e instalação limpa, depois rollback compatível.
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
| R4-OPE-01 | hub/internal/providerauth/client.go, hub/internal/providerauth/client.go | L1 utilizável na falha de cofre e coordenação limitada |
| R4-OPE-02 | hub/migrations/control/0002_provider_auth.sql, hub/migrations/control/0004_provider_api_key.sql, hub/deploy/r2/scripts/migrate.sh | Upgrade com migração histórica imutável |
| R4-OPE-03 | hub/internal/atlas/offers.go, hub/internal/atlas/catalog.go, hub/internal/atlasclient/client.go | Consulta de oferta seletiva sem materializar o portfólio |

## Main Flows
Executar os cenários de specs/r4-03-upgrade-cache-e-crescimento/spec.md no caminho real.
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
