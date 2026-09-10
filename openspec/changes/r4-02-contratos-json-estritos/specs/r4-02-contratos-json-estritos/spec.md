# Delta for r4-02-contratos-json-estritos

## ADDED Requirements

### Requirement: R4-CTR-01 — Documento JSON único e tipos sem coerção
O Hub SHALL aceitar exatamente um documento JSON completo e validar tipo JSON sem coerção por representação Go. null só é válido quando permitido explicitamente; strings numéricas não são números. Entrada deve ser totalmente consumida antes de retornar/persistir o payload, com mesma regra com/sem mapping.

#### Scenario: R4-CTR-01-S01 — objeto com inteiro 9007199254740993 e string legítima
- GIVEN objeto com inteiro 9007199254740993 e string legítima
- WHEN transformar com/sem mapping
- THEN precisão e tipo permanecem corretos

#### Scenario: R4-CTR-01-S02 — null para string ou "123" para integer
- GIVEN null para string ou "123" para integer
- WHEN validar
- THEN retorna erro de tipo com caminho sem aceitar coerção

#### Scenario: R4-CTR-01-S03 — segundo documento ou lixo após objeto válido
- GIVEN segundo documento ou lixo após objeto válido
- WHEN transformar
- THEN entrada inteira é recusada antes do efeito e nenhum body inválido é devolvido

### Requirement: R4-CTR-02 — Dialeto de schema publicado e semântica numérica
O Hub SHALL declarar dialeto/subconjunto e compilar/validar schemas na publicação, recusando keywords não suportadas em qualquer nível. Publicação e runtime usam a mesma semântica; equivalência numérica e tipo integer devem seguir o dialeto documentado com precisão exata. Limitar profundidade/custo sem ignorar regras.

#### Scenario: R4-CTR-02-S01 — minimum/maximum/enum suportados ou explicitamente não suportados
- GIVEN minimum/maximum/enum suportados ou explicitamente não suportados
- WHEN publicar e validar
- THEN ou a regra é aplicada integralmente ou a publicação é recusada

#### Scenario: R4-CTR-02-S02 — valores 1, 1.0 e 1e0 no dialeto JSON Schema escolhido
- GIVEN valores 1, 1.0 e 1e0 no dialeto JSON Schema escolhido
- WHEN validar integer e enum numérico
- THEN semântica matemática é coerente sem conversão imprecisa

#### Scenario: R4-CTR-02-S03 — keyword desconhecida aninhada ou schema excessivamente profundo
- GIVEN keyword desconhecida aninhada ou schema excessivamente profundo
- WHEN publicar
- THEN erro aponta localização/limite e não habilita contrato parcialmente validado
