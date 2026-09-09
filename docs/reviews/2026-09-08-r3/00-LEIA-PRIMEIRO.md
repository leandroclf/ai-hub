# R3 — Revisão da implementação v4 + R2
Snapshot: a4a876a9f8e875db882f7ca45cf7dece24d57aee. Repositório: https://github.com/leandroclf/ai-hub. Revisão: 2026-09-08.
Entrega de especificação e revisão; nenhum código de produção foi alterado.

Extraia o ZIP na raiz do repositório: ele adiciona docs/reviews/2026-09-08-r3/ e sete pastas openspec/changes/r3-*/.
Não coloque o pacote dentro de hub/ e não substitua as pastas v4/R2. O prompt completo está nesta pasta como PROMPT_IMPLEMENTACAO_INTEGRAL_R3.md.

Leia 01-RELATORIO.md, 02-ARQUITETURA-E-INVARIANTES.md, 03-CONSOLE-E-CONTRATOS.md,
04-QUALIFICACAO-E-OPERACAO.md, 05-DECISOES-E-MIGRACAO.md e as matrizes CSV.
A matriz 06 contém os 164 requisitos herdados; a R3 acrescenta 25 requisitos de integração,
75 cenários e sete changes. Nenhuma obrigação herdada é removida.

Os status são de auditoria: mecanismo observado não equivale a requisito qualificado.
Os testes integrados não executados continuam pendentes. Não há certificação de produção, HA ou segurança vital.
Cada P0 é um gate antes de homologação/produção, inclusive quando o risco exige prova integrada ainda ausente.
Os cenários novos não substituem os cenários v4/R2: o agente deve executar a união.

As provas negativas de contrato estão em evidence/. São testes externos por overlay, deliberadamente falhando
contra este snapshot para demonstrar defeitos; não foram adicionados ao código do repositório.
