# Explore: Evidência consolidada da implementação

Snapshot a540b40007fe6b8ed523e17afe00e96ff8f8ad50. Leitura direta do delta R5→R6.

## F-R6-15 · P1 · ANALISE_ESTATICA
O relatório R5 é explicitamente parcial. EVIDENCE_INDEX diz que os logs brutos ficaram no scroll, não foram persistidos. O commit menciona 246 testes, enquanto FINAL_REPORT/EVIDENCE_INDEX mencionam 245. SCENARIO_RESULTS contém apenas subconjunto e IDs de gate ad hoc. As matrizes originais da auditoria não foram regeneradas. Isto limita verificabilidade, não prova que testes históricos falharam.

Fontes: [docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md:29](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md#L29), [docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv:11](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv#L11), [docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md:48](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md#L48)

Vínculo anterior: R5-QUA-01.
