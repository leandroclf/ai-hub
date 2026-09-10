# Proposal: Conclusão integrada da baseline
## Change ID
r4-04-conclusao-integrada-da-baseline
## Status
Em execução — gates integrados parciais; conclusão da baseline pendente.
## Why
As specs têm 189 requisitos e 696 cenários (416 v4+205 R2+75 R3), mas matriz registra 280, omitindo cenários v4. FINAL_REPORT diz OpenSpec não executado; CHECKPOINT informa strict 17/17 e novas mudanças. Título do merge afirma conclusão, incompatível com gates ainda abertos.
Nenhum arquivo de hub/admin-ui mudou entre os snapshots. Permanecem a escolha incorreta de delivery_id, SLA/reconcile sem rota efetiva e comandos financeiros incompatíveis. Domínios financeiro, DAG, capacidade, objetos e kind também não receberam fechamento funcional neste delta.
go test -race passa com 37 testes pass e 18 skip nesta revisão. OpenSpec strict 17/17 passa sem config.yaml; não é bloqueio técnico atual. Não há comprovação nova de DB/Compose/kind/browser/HA. AGENTS atualizado exige um ecossistema Compose por vez, contrariando prompts históricos que sugeriam laboratórios paralelos.
## Context
Brownfield do commit b9d0f90ce02aa0c27cad546745153d160ff5867f. A baseline tem 189 requisitos/696 cenários.
## Problem
As lacunas observadas impedem contrato/custódia/qualificação integral.
## Goals
- Inventário integral e evidência coerente com o conteúdo
- Entrega vertical do console e baseline remanescente
- Qualificação reproduzível com um ecossistema local
## Non-Goals
Reescrever stack, aprovar contrato comercial ou reduzir semântica de SLA.
## Users / Actors Impacted
Clientes, provedores e operadores nominais. Responsável funcional: Engenharia, Frontend e Qualidade.
## Scope
### In scope
Requisitos deste change e fechamento das obrigações herdadas vinculadas.
### Out of scope
Push/merge/deploy remoto e exclusão de dados existentes sem autorização.
## Product Requirements Summary
- R4-QUA-01: A engenharia SHALL gerar inventário diretamente de todas as specs, preservar os 696 cenários herdados e acrescentar R4 sem omissões. Relatório/checkpoint/matrizes devem ser consistentes e vinculados a commit+hash de conteúdo/digests. PASS exige resultado verificável; histórico e evidência do working tree devem ter proveniência distinguível.
- R4-QUA-02: A próxima implementação SHALL concluir o backlog remanescente v4/R2/R3 por fatias verticais, com contratos backend/UI, persistência, oráculos e browser. Cada jornada só termina quando seu efeito real e autorização são demonstrados; componentes auxiliares e documentação isolados não satisfazem a tarefa. Os requisitos herdados não são substituídos pelos novos R4.
- R4-QUA-03: A qualificação SHALL preparar ferramentas fixadas, executar gates integrados sem skip obrigatório e cumprir um único ecossistema Compose ativo do Hub por vez. Inventariar containers/projeto, preservar volumes e terceiros, trocar somente alvos identificados e registrar limpeza. Ausência real de runtime é bloqueio delimitado, não PASS nem motivo para abandonar implementação independente.
## Business Rules
Custódia antes de ACK, UUIDv7 após aceite, identidade por recurso, resultado terminal único, snapshots e finanças exatas permanecem.
## Affected Capabilities
r4-04-conclusao-integrada-da-baseline
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
