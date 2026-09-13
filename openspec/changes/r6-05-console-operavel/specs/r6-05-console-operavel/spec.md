# Delta for r6-05-console-operavel

## ADDED Requirements

### Requirement: R6-UX-01 — Seletores completos para produtos e rotas
O console SHALL permitir localizar e selecionar qualquer referência elegível por ID/versão, inclusive em etapas e rotas, sem carregar todo o catálogo nem ocultar a referência atualmente selecionada. Paginação e busca SHALL preservar tenant, filtros e estado de edição.

#### Scenario: R6-UX-01-S01 — mais de 50 serviços/contas/bindings elegíveis
- GIVEN mais de 50 serviços/contas/bindings elegíveis
- WHEN editar produto e rota escolhendo item de página posterior
- THEN seleção salva ID e versão corretos e reaparece ao reabrir

#### Scenario: R6-UX-01-S02 — item já selecionado fora da primeira página
- GIVEN item já selecionado fora da primeira página
- WHEN abrir edição e pesquisar outro termo
- THEN referência atual continua identificável e não é substituída silenciosamente

#### Scenario: R6-UX-01-S03 — tenant muda com busca em voo
- GIVEN tenant muda com busca em voo
- WHEN resposta antiga chega depois
- THEN nenhuma opção ou rascunho de outro tenant é aplicado

### Requirement: R6-UX-02 — Validação e intenção preservadas em toda mutação
O console SHALL validar contratos de sucesso e erro por operação e bloquear mutações enquanto qualquer editor apresenta entrada inválida. Uma intenção de mutação com resultado desconhecido SHALL conservar identidade, payload e versão até reconciliação, inclusive após nova tentativa ou recarga.

#### Scenario: R6-UX-02-S01 — mapping editado para JSON inválido
- GIVEN mapping editado para JSON inválido
- WHEN clicar salvar/publicar
- THEN nenhuma mutação enviada e texto preservado com erro acessível

#### Scenario: R6-UX-02-S02 — servidor confirma mutação mas duas respostas se perdem
- GIVEN servidor confirma mutação mas duas respostas se perdem
- WHEN recarregar e tentar novamente
- THEN mesma intenção é reconciliada, um efeito e nenhuma chave nova prematura

#### Scenario: R6-UX-02-S03 — detalhe ou retorno de mutação fora do contrato
- GIVEN detalhe ou retorno de mutação fora do contrato
- WHEN carregar editor e confirmar ação
- THEN erro explícito sem corrupção do rascunho nem sucesso aparente
