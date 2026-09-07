# Auditoria de autenticação de provedores — 07/09/2026

## Resultado

Execução local realizada com Docker Compose, PostgreSQL, Redis, LocalStack e
`provider-sim`. O comando `go test -tags e2e ./...` passou, assim como `npm run build`,
`node screenshots/validate.mjs` e a captura Playwright das telas e do Swagger.

## Cenários executados

| Cenário | Conta | Política | Resultado |
|---|---|---|---|
| Basic | `prov-basic-1` | `SHARED_HUB` | HTTP 200 / `SUCCEEDED` |
| OAuth inicial | `prov-oauth-1` | `SHARED_HUB` | HTTP 200 / `SUCCEEDED` |
| OAuth cache hit | `prov-oauth-1` | `SHARED_HUB` | HTTP 200 / `SUCCEEDED` |
| mTLS + OAuth | `prov-mtls-oauth-1` | `SHARED_HUB` | HTTP 200 / `SUCCEEDED` |
| OAuth dedicado | `prov-oauth-1` | `TENANT_DEDICATED` | HTTP 200 / `SUCCEEDED` |

O registro detalhado está em `../auth_scenarios_output.txt`. Os perfis persistidos no
PostgreSQL estão em `../db/provider_auth_profiles.txt`. O Redis foi consultado sem
imprimir valores sensíveis: `../db/redis_token_cache.txt` registra as chaves e seus
TTLs, e a execução confirmou `DBSIZE=2`.

## Controles observados

- A API do Atlas normaliza `NONE`/TTL padrão e rejeita tipo desconhecido, URL inválida,
  OAuth incompleto, Basic sem usuário/referência ou mTLS sem referência de certificado.
- O Cometa obtém tokens por Client Credentials, grava-os em Redis com margem de
  expiração e reutiliza a chave por conta de provedor.
- A seleção `SHARED_HUB` versus `TENANT_DEDICATED` ocorre no Atlas sem fallback
  silencioso para outro tenant.
- Segredos não são persistidos em claro; somente referências são usadas na massa local.
- Os logs da rodada estão em `../logs/auth_integration_final.log`; a janela estável sem
  FATAL/ERROR/WARN está em `../logs/auth_clean_window.log`.

## Limitações para homologação/produção

O cenário mTLS é uma simulação controlada: o provedor local valida a referência por
header, não um certificado apresentado em handshake TLS. Para produção, é necessário
HTTPS com certificado cliente, cadeia de confiança, rotação e armazenamento seguro.
As referências `vault://` também precisam ser ligadas a Secrets Manager/KMS ou cofre
equivalente. Inbound OAuth/OIDC do cliente ainda não está implementado; a entrada local
continua baseada em `X-Tenant-Id`.

## Compatibilidade dos testes legados

Os testes E2E originais da fatia inicial referenciam deliberadamente os fixtures
`consulta-cadastral`/`prov-sync-1`. Após a substituição solicitada, uma execução sem
cache (`go test -count=1 -tags e2e ./test/e2e/...`) não é um teste válido do catálogo
HivePlace e falha por ausência desses fixtures. Os testes unitários, o build, o Compose,
a validação das telas e a verificação de contagem do catálogo importado passaram. Uma
nova suíte E2E específica do contrato externo HivePlace deve ser criada antes de chamar
essas APIs reais; ela não foi executada para não disparar operações externas com as
credenciais do environment.
