# Delta for r5-06-baseline-e-evidencia-integral

## ADDED Requirements

### Requirement: R5-QUA-01 — Baseline reconciliada e resultados com proveniência
A engenharia SHALL inventariar specs reais e reconciliar revisões não incorporadas com rastreio explícito, sem perder requisito nem duplicar identidade. Resultado deve carregar proveniência de cenário/conteúdo/artefato e oráculo, distinguindo PASS, FAIL, NOT_RUN e bloqueio externo. Associação a arquivo não é conformidade; mudança material invalida a prova afetada.

#### Scenario: R5-QUA-01-S01 — baseline remota 201/732 e quatro obrigações R4 ausentes
- GIVEN baseline remota 201/732 e quatro obrigações R4 ausentes
- WHEN gerar próxima rodada
- THEN ponte explícita para requisitos R5 sem remover baseline

#### Scenario: R5-QUA-01-S02 — cenário ou código muda após execução
- GIVEN cenário ou código muda após execução
- WHEN regenerar matriz
- THEN resultado anterior não é promovido automaticamente para PASS

#### Scenario: R5-QUA-01-S03 — log histórico, smoke exploratório e teste integrado atual
- GIVEN log histórico, smoke exploratório e teste integrado atual
- WHEN relatar conclusão
- THEN escopos distintos, contagens verificáveis e P0 não encerrado por documentação
