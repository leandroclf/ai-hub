# Fontes, método e limites

## Fontes primárias do projeto

Snapshot a39d394b0d87185ed4cc3861c12ec45f2c302d9e; commit de main com assunto “feat(catalogo): importa APIs HivePlace da collection Postman”. As evidências do relatório usam links imutáveis de arquivo e linha nesse SHA. O texto do usuário e a baseline docs/00 a 09 e OpenSpec v4 compõem o alvo, não prova de implementação.

Metodologia aplicada: AGENTS.md; docs/openspec-docs/01-prompt-final-completo.md; 03-template-openspec-change.md; 04-checklist-avaliador.md; 05-referencias-metodologia.md. O pedido de pacote completo permite concluir proposal/specs/design/tasks sem pausa de aprovação entre cada artefato. Esta entrega não implementa aplicação.

Inspeção orientada por caminho real de execução: admissão → contrato/credencial → despacho → persistência/tentativa → provedor → final → consulta/webhook → custo/receita. Revisão adicional de frontend, banco, estado, recuperação, deploy e observabilidade. Buscas negativas foram cruzadas com árvore e código, não apenas README. Relatórios e screenshots preexistentes foram tratados como evidência histórica do projeto, não ensaio repetido nesta rodada.

## Referências técnicas verificadas durante a revisão

Consulta em 07/09/2026; conclusão documental em 08/09/2026. Não se declara compatibilidade de todas as versões remotas apenas por essa consulta.

| Fonte | Uso delimitado |
|---|---|
| [OpenSpec — CLI](https://github.com/Fission-AI/OpenSpec/blob/main/docs/cli.md) | Comando validate/strict e distinção entre validação e archive; CLI 1.12.0 instalado para conferir este pacote |
| [AWS — Transactional Outbox](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html) | Estado e intenção na mesma transação; deduplicação ainda necessária |
| [AWS — SQS at-least-once](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/standard-queues-at-least-once-delivery.html) | Possibilidade de reentrega em filas Standard; aplicação controla ACK/efeito |
| [PostgreSQL — RLS](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) | Políticas por linha e exceções de papéis privilegiados |
| [Kubernetes — HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) | Escala horizontal baseada em métricas; não dimensiona banco ou quota de provedor |
| [KEDA — Scaling Deployments](https://keda.sh/docs/2.20/concepts/scaling-deployments/) | Integração com HPA e escala de workloads por eventos |
| [Go — FAQ/goroutines](https://go.dev/doc/faq#goroutines) | Concorrência suportada pelo runtime; não comprova latência deste projeto |
| [W3C — WCAG 2.2](https://www.w3.org/TR/WCAG22/) | Referência para gate AA proposto das jornadas administrativas |

Os algoritmos, divisão de trabalho, APIs propostas e decisões de produto são recomendações desta revisão fundamentadas no código e requisitos do usuário. Não são citações atribuídas a essas fontes. Não houve revisão jurídica, certificação de segurança funcional, auditoria de uma implantação remota nem prova de RPO regional.
