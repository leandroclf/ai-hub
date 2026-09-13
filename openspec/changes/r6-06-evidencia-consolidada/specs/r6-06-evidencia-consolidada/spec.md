# Delta for r6-06-evidencia-consolidada

## ADDED Requirements

### Requirement: R6-QUA-01 — Evidência completa e vinculada ao código revisado
A engenharia SHALL manter inventário integral extraído das specs e resultado individual com evidência por cenário. Contagens SHALL ser calculadas dos logs, com subtestes/pacotes diferenciados. Um resultado histórico sem vínculo verificável com artefato SHALL permanecer histórico e não virar PASS do candidato.

#### Scenario: R6-QUA-01-S01 — specs vigentes e resultados parciais antigos
- GIVEN specs vigentes e resultados parciais antigos
- WHEN gerar inventário e matriz
- THEN cada cenário consta exatamente uma vez e lacunas ficam NOT_RUN

#### Scenario: R6-QUA-01-S02 — teste altera subcasos e SHA muda
- GIVEN teste altera subcasos e SHA muda
- WHEN regenerar evidências e relatório
- THEN contagens derivadas e aprovação ligada ao candidato correto

#### Scenario: R6-QUA-01-S03 — dependência obrigatória indisponível
- GIVEN dependência obrigatória indisponível
- WHEN executar gate integrado
- THEN FAIL/BLOCK com causa, nunca skip contado como aprovação
