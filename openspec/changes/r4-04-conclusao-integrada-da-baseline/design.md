# Design: Conclusão integrada da baseline
## Context
b9d0f90ce02aa0c27cad546745153d160ff5867f. Fontes em explore.md. Responsável: Engenharia, Frontend e Qualidade.
## Goals and Constraints
Corrigir comportamento observado e preservar baseline v4/R2/R3. AGENTS governa ambiente.
## Proposed Architecture
Uma única fonte de inventário percorre specs sem reescrever títulos. IDs v4 de auditoria derivam de requisito+ordem e guardam caminho/linha/hash do texto; se specs mudarem, recalcular e reconciliar sem importar status cegamente. Resultados de execução ficam separados do catálogo normativo.
Qualificação contém commit, hash de working tree quando houver, digests de imagens, versões e comandos, contagens pass/fail/skip. Relatório deriva desses registros; mensagem de merge não concede status. Regressões pontuais não equivalem a cenário integral.
Uma execução Compose estável por vez, conforme AGENTS. Inspecionar projeto ativo, preservar volumes, encerrar somente componentes identificados do Hub, verificar término e só então subir substituto. Kind não pode criar projeto Compose auxiliar paralelo; dependências ficam no cluster no perfil qualificado.
Gate inclui browser real e efeito no backend; TypeScript não valida runtime. Reutilizar tasks v4/R2/R3 e fechar por fatias: catálogo→admissão→provedor→resultado→entrega→financeiro→console. CI falha em skips obrigatórios; restrição externa delimita apenas o gate afetado.
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
| R4-QUA-01 | docs/reviews/2026-09-08-r3/implementation/SCENARIO_RESULTS.csv, docs/reviews/2026-09-08-r3/implementation/FINAL_REPORT.md, docs/reviews/2026-09-08-r3/implementation/CHECKPOINT.md | Inventário integral e evidência coerente com o conteúdo |
| R4-QUA-02 | hub/admin-ui/src/pages/OperationsPage.tsx, hub/admin-ui/src/pages/FinancePage.tsx, hub/internal/orbita/admin.go | Entrega vertical do console e baseline remanescente |
| R4-QUA-03 | AGENTS.md, hub/internal/cometa/custody_test.go, hub/deploy/r2/kind/render-runtime.py | Qualificação reproduzível com um ecossistema local |

## Main Flows
Executar os cenários de specs/r4-04-conclusao-integrada-da-baseline/spec.md no caminho real.
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
