# Explore — Upgrade, cache e crescimento

Snapshot b9d0f90ce02aa0c27cad546745153d160ff5867f; revisão incremental v4/R2/R3.

## F-R4-07
Tokens deixaram Redis e lock passou a ser por chave. Contudo, Resolve é chamado antes do L1; prova com L1 válido e cofre indisponível falha. locks cresce sem remoção por binding/versão e mutex não respeita cancelamento durante espera.

[hub/internal/providerauth/client.go:74](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/providerauth/client.go#L74), [hub/internal/providerauth/client.go:144](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/providerauth/client.go#L144)

## F-R4-08
Estado atual (ensaio de 11/09/2026): o runner foi executado em instalação
limpa e replay idempotente; checksum desconhecido falhou antes de prosseguir e
a variante histórica conhecida de `0002_provider_auth.sql` foi reconciliada
somente com schema API_KEY/header comprovado. Rollback de versão publicada ainda
não foi ensaiado.

[hub/migrations/control/0002_provider_auth.sql:3](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/migrations/control/0002_provider_auth.sql#L3), [hub/migrations/control/0004_provider_api_key.sql:2](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/migrations/control/0004_provider_api_key.sql#L2), [hub/deploy/r2/scripts/migrate.sh:22](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/deploy/r2/scripts/migrate.sh#L22)

## F-R4-09
Estado atual (commit 8aae6fb): a resolução deixou de agregar páginas do
catálogo. `listEligibleOffers` filtra tenant, aplicação, alvo, versão e vigência
no PostgreSQL, separa o tenant local do catálogo global com `UNION ALL` para
preservar planos indexáveis e usa `LIMIT 2` para distinguir zero, uma oferta ou
ambiguidade. A qualificação com 1.500 registros irrelevantes e dois candidatos
confirmou o índice e o limite; o cliente consumidor também reutiliza projeção
válida durante indisponibilidade e recusa projeção vencida. O cenário de
crescimento produtivo com medição de memória/latência e provedor comercial
permanece não qualificado.

[hub/internal/atlas/offers.go:33](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L33), [hub/internal/atlas/catalog.go:419](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/catalog.go#L419), [hub/internal/atlasclient/client.go:22](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlasclient/client.go#L22)

Preservar correções anteriores. Nenhum achado estático é relatado como incidente de produção.
