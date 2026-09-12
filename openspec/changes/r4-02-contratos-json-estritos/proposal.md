# Proposal: Contratos JSON estritos
## Change ID
r4-02-contratos-json-estritos
## Status
Implementada — cenários R4-CTR-01/02 e rollout/rollback local qualificados; a requalificação dos cenários herdados permanece pendente.
## Why
Provas novas executadas: null foi aceito como string; "123" como integer; e dois documentos concatenados retornaram sem erro. Decoder lê apenas o primeiro documento e, sem mapping, devolve input original. A preservação do inteiro grande anterior foi corrigida e não deve regredir.
Prova executada: minimum:0 é ignorado e -1 aceito. A validação de publicação verifica só type object/properties. O comentário de recusa a construções não qualificadas não corresponde a whitelist efetiva. integer rejeita representação com .eE e enum compara json.Number lexicalmente, divergindo da semântica usual de JSON Schema.
## Context
Brownfield do commit b9d0f90ce02aa0c27cad546745153d160ff5867f. A baseline tem 189 requisitos/696 cenários.
## Problem
As lacunas observadas impedem contrato/custódia/qualificação integral.
## Goals
- Documento JSON único e tipos sem coerção
- Dialeto de schema publicado e semântica numérica
## Non-Goals
Reescrever stack, aprovar contrato comercial ou reduzir semântica de SLA.
## Users / Actors Impacted
Clientes, provedores e operadores nominais. Responsável funcional: Core e Catálogo.
## Scope
### In scope
Requisitos deste change e fechamento das obrigações herdadas vinculadas.
### Out of scope
Push/merge/deploy remoto e exclusão de dados existentes sem autorização.
## Product Requirements Summary
- R4-CTR-01: O Hub SHALL aceitar exatamente um documento JSON completo e validar tipo JSON sem coerção por representação Go. null só é válido quando permitido explicitamente; strings numéricas não são números. Entrada deve ser totalmente consumida antes de retornar/persistir o payload, com mesma regra com/sem mapping.
- R4-CTR-02: O Hub SHALL declarar dialeto/subconjunto e compilar/validar schemas na publicação, recusando keywords não suportadas em qualquer nível. Publicação e runtime usam a mesma semântica; equivalência numérica e tipo integer devem seguir o dialeto documentado com precisão exata. Limitar profundidade/custo sem ignorar regras.
## Business Rules
Custódia antes de ACK, UUIDv7 após aceite, identidade por recurso, resultado terminal único, snapshots e finanças exatas permanecem.
## Affected Capabilities
r4-02-contratos-json-estritos
## Expected Impact
### Code
Fontes e responsabilidades em design/explore.
### Data
Migrações aditivas, recibos e evidência sem alteração cega de histórico.
### APIs / Contracts
Validação/autorização efetivas e versionamento compatível.
### Integrations
Provedor, callback, cofre e contratos existentes com fixtures externas.
### Operations
Um Compose do Hub ativo por vez; scripts, gates e runbooks verificáveis.
### Security / Privacy
Menor privilégio, escopo de custódia e evidência sem segredos.
## Risks and Mitigations
Ver risk-matrix.md.
## Success Criteria
Cenários passam com evidência atual e requisitos herdados correspondentes requalificados.
## Assumptions
Dados comerciais ausentes não impedem fixtures sintéticas; não equivalem a aprovação.
## Open Questions
D-01…D-07, P-01…P-11 e T-R2-01 continuam sujeitos ao estado real do registro; não presumir aprovação.
