# Evaluator checklist

- [ ] Requisito SHALL e cenários observáveis, sem teste que espelha apenas a implementação.
- [ ] Todos os achados têm fonte atual e rastreio para tarefa.
- [ ] DTO, schema, erros, autorização e versão definidos.
- [ ] Caminho real usa o mecanismo, não apenas helper/teste isolado.
- [ ] Custódia, idempotência, concorrência e recuperação demonstradas.
- [ ] Identidade tenant/aplicação/célula e acesso administrativo testados negativamente.
- [ ] Evidência no SHA candidato, sem segredos e sem skips obrigatórios.
- [ ] Migração/rollback não removem obrigações históricas.
- [ ] Requisitos herdados vinculados foram requalificados.
- [ ] Riscos/decisões pendentes não foram considerados resolvidos por inferência.
- [ ] OpenSpec strict passou; não arquivar baseline antes da qualificação.
