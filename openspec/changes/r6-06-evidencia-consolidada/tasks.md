# Tasks: Evidência consolidada da implementação

## Execução incremental
Responsável: Engenharia e Qualidade. Dependências entre changes e ambiente estão no plano da revisão.

## R6-QUA-01
- [ ] R6-QUA-01-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-QUA-01-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md, docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv, docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-QUA-01-B — Implementar a regra no caminho real
  - Objetivo: Preservar v4/R2/R3/R4/R5 e incorporar deltas R6, com semântica distinta para achado, requisito, cenário e teste. Registrar SHA/diff, toolchain, imagem, fixture, comando, esperado/observado, logs saneados e digests. Revalidar só o que mudança material afeta, mas completar cenários obrigatórios sem skip. CLI OpenSpec fixada, sem @latest em evidência reproduzível. Decisões externas não aprovadas permanecem pendentes sem impedir implementação de fixture técnica.
  - Componentes: docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md, docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv, docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md.
  - Depende de: R6-QUA-01-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-QUA-01-C — Qualificar falhas e concorrência
  - Objetivo: A engenharia SHALL manter inventário integral extraído das specs e resultado individual com evidência por cenário. Contagens SHALL ser calculadas dos logs, com subtestes/pacotes diferenciados. Um resultado histórico sem vínculo verificável com artefato SHALL permanecer histórico e não virar PASS do candidato.
  - Componentes: docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md, docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv, docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md.
  - Depende de: R6-QUA-01-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-QUA-01-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md, docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv, docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md.
  - Depende de: R6-QUA-01-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] QUA-G1 — Validar OpenSpec strict e compatibilidade com todos os requisitos anteriores vinculados.
- [ ] QUA-G2 — Conferir tarefas contra evidências e registrar riscos remanescentes sem autodeclarar aprovação produtiva.
