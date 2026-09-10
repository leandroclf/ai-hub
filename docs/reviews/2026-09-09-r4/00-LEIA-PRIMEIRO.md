# R4 — Revisão incremental e conclusão da baseline
Repositório: https://github.com/leandroclf/ai-hub
Snapshot: b9d0f90ce02aa0c27cad546745153d160ff5867f; merge da PR #2 em 2026-09-09T10:08:45Z.
Referência anterior: a4a876a9f8e875db882f7ca45cf7dece24d57aee.

**R4 é a rodada de revisão, não uma substituição da arquitetura v4.**
Extraia o ZIP na raiz do repositório. Ele adiciona docs/reviews/2026-09-09-r4/
e quatro changes openspec/changes/r4-*. Não sobrescreva v4/R2/R3.

A R4 registra 12 achados/requisitos adicionais, 36 cenários novos e preserva 189 requisitos/696 cenários herdados.
A união tem 201 requisitos e 732 cenários. Os IDs V4:<requisito>:S<ordem> são identificadores de auditoria;
não alteram os títulos nem reescrevem as specs originais.

Leia o relatório, o backlog de conclusão, as matrizes e o prompt desta pasta.
Os 25 itens R3 continuam rastreados individualmente, distinguindo correção pontual de qualificação integral.
Não se deve repetir o mesmo trabalho corrigido nem abandonar obrigações antigas ao criar nova rodada.

O AGENTS atual exige **um único ecossistema Compose do Hub ativo por vez**.
Essa regra deve prevalecer sobre sugestões históricas de laboratórios paralelos.
Preserve volumes/dados e componentes de terceiros; não execute prune amplo.

Esta entrega contém documentação e instrumentos de auditoria. Nenhum código de produção foi alterado.
