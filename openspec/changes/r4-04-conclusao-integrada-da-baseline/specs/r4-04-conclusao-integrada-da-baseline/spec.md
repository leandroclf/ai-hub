# Delta for r4-04-conclusao-integrada-da-baseline

## ADDED Requirements

### Requirement: R4-QUA-01 — Inventário integral e evidência coerente com o conteúdo
A engenharia SHALL gerar inventário diretamente de todas as specs, preservar os 696 cenários herdados e acrescentar R4 sem omissões. Relatório/checkpoint/matrizes devem ser consistentes e vinculados a commit+hash de conteúdo/digests. PASS exige resultado verificável; histórico e evidência do working tree devem ter proveniência distinguível.

#### Scenario: R4-QUA-01-S01 — specs atuais e pacote R4
- GIVEN specs atuais e pacote R4
- WHEN gerar e verificar matrizes
- THEN todos os requisitos/cenários aparecem uma vez por identidade sem omissão v4

#### Scenario: R4-QUA-01-S02 — código muda após teste ou relatório contradiz checkpoint
- GIVEN código muda após teste ou relatório contradiz checkpoint
- WHEN executar gate de evidência
- THEN inconsistência invalida conclusão até reexecução/reconciliação

#### Scenario: R4-QUA-01-S03 — correção pontual com requisito integral pendente
- GIVEN correção pontual com requisito integral pendente
- WHEN atualizar status
- THEN progresso é registrado sem encerrar requisito nem repetir genericamente todos os achados como abertos

### Requirement: R4-QUA-02 — Entrega vertical do console e baseline remanescente
A próxima implementação SHALL concluir o backlog remanescente v4/R2/R3 por fatias verticais, com contratos backend/UI, persistência, oráculos e browser. Cada jornada só termina quando seu efeito real e autorização são demonstrados; componentes auxiliares e documentação isolados não satisfazem a tarefa. Os requisitos herdados não são substituídos pelos novos R4.

#### Scenario: R4-QUA-02-S01 — operador faz catálogo→execução→consulta→entrega→financeiro
- GIVEN operador faz catálogo→execução→consulta→entrega→financeiro
- WHEN usar navegador e API reais
- THEN jornada persiste e corresponde ao efeito e ao contrato por tenant

#### Scenario: R4-QUA-02-S02 — duas entregas do mesmo protocolo, erro financeiro e token consumidor
- GIVEN duas entregas do mesmo protocolo, erro financeiro e token consumidor
- WHEN abrir detalhe/submeter/repetir/acessar admin
- THEN IDs corretos, intenção idempotente e autorização/erros efetivos são provados

#### Scenario: R4-QUA-02-S03 — helper ou menu implementado sem conexão real
- GIVEN helper ou menu implementado sem conexão real
- WHEN tentar encerrar requisito
- THEN gate rejeita encerramento e mantém tarefa com ponto exato de integração pendente

### Requirement: R4-QUA-03 — Qualificação reproduzível com um ecossistema local
A qualificação SHALL preparar ferramentas fixadas, executar gates integrados sem skip obrigatório e cumprir um único ecossistema Compose ativo do Hub por vez. Inventariar containers/projeto, preservar volumes e terceiros, trocar somente alvos identificados e registrar limpeza. Ausência real de runtime é bloqueio delimitado, não PASS nem motivo para abandonar implementação independente.

#### Scenario: R4-QUA-03-S01 — ambiente com ecossistema Hub existente
- GIVEN ambiente com ecossistema Hub existente
- WHEN preparar qualificação
- THEN inventário precede troca e somente um projeto Hub fica ativo sem apagar dados

#### Scenario: R4-QUA-03-S02 — dependência integrada ausente ou teste t.Skip
- GIVEN dependência integrada ausente ou teste t.Skip
- WHEN executar gate obrigatório
- THEN gate falha com diagnóstico e preparação é tentada dentro da autorização

#### Scenario: R4-QUA-03-S03 — Compose e kind preparados com imagens do conteúdo atual
- GIVEN Compose e kind preparados com imagens do conteúdo atual
- WHEN executar jornadas, sinais e recuperação
- THEN resultados reais distinguem render/build de runtime e registram SHA/hash/digests
