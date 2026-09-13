# Explore: Console administrativo operável

Snapshot a540b40007fe6b8ed523e17afe00e96ff8f8ad50. Leitura direta do delta R5→R6.

## F-R6-13 · P1 · ANALISE_ESTATICA
Lookup agora busca no servidor com limite 50, resolvendo o antigo teto de 1000 para campos com pesquisa. StepsEditor e RoutesEditor recebem apenas essa primeira lista e não expõem pesquisa/cursor próprios; serviços, contas e bindings fora dos primeiros 50 ficam inalcançáveis nesses editores. Referência selecionada fora da página também não é carregada pontualmente.

Fontes: [hub/admin-ui/src/pages/CatalogPage.tsx:25](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L25), [hub/admin-ui/src/pages/CatalogPage.tsx:59](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L59), [hub/admin-ui/src/pages/CatalogPage.tsx:76](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L76)

Vínculo anterior: R5-UX-01.

## F-R6-14 · P1 · ANALISE_ESTATICA
Guards runtime foram introduzidos, mas são opcionais e usados sobretudo nas listagens; detalhe, save e command ainda fazem cast sem guard. mappingErrors é local ao StepsEditor e não participa do bloqueio save do pai: JSON de mapping inválido pode deixar salvo o valor anterior. command mantém chave por duas tentativas, mas uma nova ação/reload cria outra identidade.

Fontes: [hub/admin-ui/src/api/admin.ts:18](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/api/admin.ts#L18), [hub/admin-ui/src/api/admin.ts:32](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/api/admin.ts#L32), [hub/admin-ui/src/pages/CatalogPage.tsx:62](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L62), [hub/admin-ui/src/pages/CatalogPage.tsx:41](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L41)

Vínculo anterior: R5-UX-01.
