# Design: Console administrativo operável

## Context
Brownfield a540b40007fe6b8ed523e17afe00e96ff8f8ad50; responsabilidade Frontend e Produto. O delta adiciona critérios de fechamento, não substitui regras anteriores.

## R6-UX-01 — decisão e justificativa
Componente comum de seleção remota com debounce, cancelamento, cursor e resolução pontual do valor selecionado. Passar consulta por editor e não refazer todos os lookups por tecla. Diferenciar vazio de erro e impedir resposta antiga de outro tenant. Contratos de elegibilidade vêm da API, não de filtros locais parciais.

### Contrato obrigatório
O console SHALL permitir localizar e selecionar qualquer referência elegível por ID/versão, inclusive em etapas e rotas, sem carregar todo o catálogo nem ocultar a referência atualmente selecionada. Paginação e busca SHALL preservar tenant, filtros e estado de edição.

### Superfícies afetadas
[hub/admin-ui/src/pages/CatalogPage.tsx:25](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L25), [hub/admin-ui/src/pages/CatalogPage.tsx:59](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L59), [hub/admin-ui/src/pages/CatalogPage.tsx:76](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L76)

### Provas de fechamento
- R6-UX-01-S01: mais de 50 serviços/contas/bindings elegíveis → editar produto e rota escolhendo item de página posterior → seleção salva ID e versão corretos e reaparece ao reabrir.
- R6-UX-01-S02: item já selecionado fora da primeira página → abrir edição e pesquisar outro termo → referência atual continua identificável e não é substituída silenciosamente.
- R6-UX-01-S03: tenant muda com busca em voo → resposta antiga chega depois → nenhuma opção ou rascunho de outro tenant é aplicado.

## R6-UX-02 — decisão e justificativa
Propagar validade do editor ao formulário pai sem apagar texto inválido. Guards ou schemas gerados para DTOs de catálogo, finanças, protocolos e operações; erro de contrato não altera draft. Journal local da intenção sem tokens/segredos, escopado por usuário/tenant/ambiente, reconciliado com autoridade antes de gerar outra chave. Testes browser de timeout após commit, duplo clique, sessão expirada, ETag e permissões.

### Contrato obrigatório
O console SHALL validar contratos de sucesso e erro por operação e bloquear mutações enquanto qualquer editor apresenta entrada inválida. Uma intenção de mutação com resultado desconhecido SHALL conservar identidade, payload e versão até reconciliação, inclusive após nova tentativa ou recarga.

### Superfícies afetadas
[hub/admin-ui/src/api/admin.ts:18](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/api/admin.ts#L18), [hub/admin-ui/src/api/admin.ts:32](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/api/admin.ts#L32), [hub/admin-ui/src/pages/CatalogPage.tsx:62](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L62), [hub/admin-ui/src/pages/CatalogPage.tsx:41](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L41)

### Provas de fechamento
- R6-UX-02-S01: mapping editado para JSON inválido → clicar salvar/publicar → nenhuma mutação enviada e texto preservado com erro acessível.
- R6-UX-02-S02: servidor confirma mutação mas duas respostas se perdem → recarregar e tentar novamente → mesma intenção é reconciliada, um efeito e nenhuma chave nova prematura.
- R6-UX-02-S03: detalhe ou retorno de mutação fora do contrato → carregar editor e confirmar ação → erro explícito sem corrupção do rascunho nem sucesso aparente.

## Persistência e atomicidade
A autoridade do domínio conserva transição, inbox/outbox e recibos em transação local. Cache é derivado; queda de Redis não altera a decisão durável. Para efeitos externos, lease sozinho não é prova de ausência de efeito: usar fencing reconhecido, idempotência do provedor ou reconciliação por chave. Estado UNKNOWN conserva obrigação.
## Comunicação e paralelismo
HTTP/JSON nos contratos existentes; SNS/SQS para fatos/obrigações; HTTP direto preservado para SYNC. Workers paralelos limitados por tenant/rota/produto, sem goroutine ilimitada. Timeout de transporte não encerra automaticamente o estado econômico.
## Segurança e privacidade
Identidade nominal/MFA para administração, autorização por recurso, logs sem payload sensível ou segredo. Testes com identidades A/B e controles positivos. Não registrar credenciais em evidências.
## Observabilidade
Métricas de aceites, finais, idade de obrigação, retries, violações de prazo, drift, quarentenas e ações administrativas. Labels limitadas por agregação; UUID/protocolo em logs/traces, não em séries ilimitadas. Cada erro de custódia tem alerta e disposição.
## Estratégia de migração e rollback
Aplicar migrações novas, manter checksums antigos e ensaiar upgrade populado. Rollback binário só se schema/dados forem compatíveis; caso contrário bloquear downgrade e aplicar forward fix. Não apagar recibos, históricos, saldos ou volumes para passar teste.
## Alternativas
Reescrita da stack acrescentaria risco sem resolver os invariantes. Mocks isolados são úteis para contrato puro, insuficientes para concorrência/DB/restore. Adoção de componente adicional requer ADR com benefício medido.
## Estratégia de validação
Reproduzir caso negativo antes da correção; confirmar chamada real, sucesso, falha, concorrência, crash e idempotência conforme cenário. DB/broker/objetos/OIDC reais em laboratório quando necessários. Browser para jornadas humanas. Medir carga/HA apenas com perfil de capacidade identificado.
