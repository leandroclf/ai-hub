# Explore — Console escalável e contratual

Snapshot: f87ce33034ae29c9431b1910dcc6a633b545e330. Fontes lidas; hipóteses de concorrência são estáticas até ensaio.

## F-R5-14 (P1)
Delivery ID, SLA/reconcile, DTO financeiro, destinos, OIDC e multimodalidade foram melhorados e têm smokes. Persistem lookup que carrega até 1000 e falha acima disso, validação runtime apenas isRecord e nova chave a cada nova chamada da ação após falha/reload. Capacidade é consulta somente leitura; onboarding de capacidade/qualificação segue seed direto no laboratório. UI TTL ainda descreve aceite (rastreado em R5-EXE-01).

[hub/admin-ui/src/pages/CatalogPage.tsx:25](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/CatalogPage.tsx#L25), [hub/admin-ui/src/api/admin.ts:26](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/api/admin.ts#L26), [hub/admin-ui/src/api/admin.ts:29](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/api/admin.ts#L29), [hub/admin-ui/src/pages/CapacityDomainsPage.tsx:7](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/CapacityDomainsPage.tsx#L7)
