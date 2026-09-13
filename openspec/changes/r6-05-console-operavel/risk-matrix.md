# Riscos e mitigação

| Risco | Prioridade | Mitigação | Aceite |
|---|---|---|---|
| Seletores completos para produtos e rotas | P1 | O console SHALL permitir localizar e selecionar qualquer referência elegível por ID/versão, inclusive em etapas e rotas, sem carregar todo o catálogo nem ocultar a referência atualmente selecionada. Paginação e busca SHALL preservar tenant, filtros e estado de edição. | R6-UX-01-S01…S03 |
| Validação e intenção preservadas em toda mutação | P1 | O console SHALL validar contratos de sucesso e erro por operação e bloquear mutações enquanto qualquer editor apresenta entrada inválida. Uma intenção de mutação com resultado desconhecido SHALL conservar identidade, payload e versão até reconciliação, inclusive após nova tentativa ou recarga. | R6-UX-02-S01…S03 |
