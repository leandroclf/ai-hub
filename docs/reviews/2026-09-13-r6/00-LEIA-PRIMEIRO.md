# AI Hub — R6
Revisão de 13/09/2026, snapshot a540b40007fe6b8ed523e17afe00e96ff8f8ad50, commit de 13/09/2026 12:36:24 UTC.

Extraia AI_HUB_OPENSPEC_R6.zip **na raiz do repositório**. Serão adicionados docs/reviews/2026-09-13-r6/ e seis changes openspec/changes/r6-*.
Preserve v4/R2/R3/R4/R5; não substitua a R5 nem arquive suas pendências como resolvidas. Esta rodada complementa a implementação atual.

Leia 01-RELATORIO-REVISAO.md e 02-PLANO-DE-IMPLEMENTACAO.md. Envie PROMPT_IMPLEMENTACAO_INTEGRAL_R6.md ao agente no checkout atualizado.

Baseline: 218 requisitos / 783 cenários / 27 changes. R6 adiciona 15 requisitos e 45 cenários em seis changes. União: 233 requisitos / 828 cenários / 33 changes. Há 72 tarefas nos changes e 25 verificações de continuidade no plano, total 97 itens de trabalho.

As prioridades são técnicas, não evidência de incidente em produção. Somente dois achados novos têm probe reproduzido aqui; os demais têm fonte estática e cenários de qualificação explícitos. Relatos históricos de integração são preservados, sem promoção automática para este SHA.
