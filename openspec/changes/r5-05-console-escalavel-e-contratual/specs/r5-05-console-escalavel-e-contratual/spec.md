# Delta for r5-05-console-escalavel-e-contratual

## ADDED Requirements

### Requirement: R5-UX-01 — Console preserva intenção e escala com catálogo
O console SHALL permitir selecionar referências por busca/paginação de servidor sem teto funcional de 1000, validar contratos de resposta por jornada e preservar identidade de intenção após resposta incerta. Configuração operacional suportada deve ter fluxo autenticado, versionado e auditável; o usuário deve distinguir pendente, confirmado e recusado.

#### Scenario: R5-UX-01-S01 — catálogo possui mais de 1000 itens e referência fora da primeira página
- GIVEN catálogo possui mais de 1000 itens e referência fora da primeira página
- WHEN pesquisar, selecionar, salvar e reabrir
- THEN referência correta disponível sem carregar tudo

#### Scenario: R5-UX-01-S02 — comando persiste e duas respostas se perdem; usuário recarrega
- GIVEN comando persiste e duas respostas se perdem; usuário recarrega
- WHEN retomar ou consultar intenção
- THEN mesma intenção/recibo sem duplicar mutação financeira

#### Scenario: R5-UX-01-S03 — API responde objeto de formato errado ou usuário sem escopo
- GIVEN API responde objeto de formato errado ou usuário sem escopo
- WHEN usar jornada
- THEN erro explícito e nenhuma confirmação falsa; backend recusa acesso
