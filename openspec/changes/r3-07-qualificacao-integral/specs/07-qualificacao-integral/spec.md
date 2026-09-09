# Delta for 07-qualificacao-integral

## ADDED Requirements

### Requirement: R3-QUA-01 — Qualificação integral do SHA sem skips ocultos
A engenharia SHALL qualificar o SHA entregue com integração obrigatória sem skips, fixture OIDC reproduzível, navegador real e oráculos externos, cobrindo toda baseline v4/R2/R3. Evidência histórica não qualifica outro SHA; tarefas só fecham com provas correspondentes e bloqueios materiais permanecem explícitos.

#### Scenario: R3-QUA-01-S01 — checkout limpo e imagens do SHA
- GIVEN checkout limpo e imagens do SHA
- WHEN rodar gate completo
- THEN bancos/broker/objetos/identidade/browser/kind têm evidência com zero skip obrigatório

#### Scenario: R3-QUA-01-S02 — dependência de fixture ausente
- GIVEN dependência de fixture ausente
- WHEN executar CI integrado
- THEN gate falha claramente em vez de falso verde

#### Scenario: R3-QUA-01-S03 — um achado corrigido e baseline pendente
- GIVEN um achado corrigido e baseline pendente
- WHEN atualizar matriz
- THEN somente itens comprovados fecham e pendências continuam visíveis
