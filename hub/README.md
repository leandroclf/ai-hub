# Hub de Interoperabilidade "Constelação" — implementação de referência (fatia inicial)

Esta pasta contém a **implementação real** (não apenas especificação) da fatia inicial
recomendada em `openspec/changes/hub-interoperabilidade-v4/tasks.md` (grupos 1–8), a partir
das specs em `openspec/changes/hub-interoperabilidade-v4/specs/*/spec.md` e do
`design.md` da mesma mudança.

Este é um ambiente **local/dev** (Docker Compose + LocalStack): não há conta AWS real,
cluster Kubernetes real, nem dados/contratos comerciais reais. Consulte
`IMPLEMENTATION_AUDIT.md` na raiz do repositório para o relatório completo do que foi
implementado, testado, simplificado ou deixado como placeholder, requisito por requisito.

## Estrutura

```
hub/
  cmd/<servico>/main.go       — um binário por aplicação (atlas, orbita, cometa, pulsar, libra, provider-sim, webhook-sink)
  internal/<dominio>/         — lógica de negócio de cada aplicação
  internal/platform/          — pacotes compartilhados (config, http, postgres, logging, ids)
  internal/queue/             — cliente SQS/SNS (aponta para LocalStack em local/dev)
  internal/objectstore/       — cliente S3 (aponta para LocalStack em local/dev)
  migrations/{control,core,finance}/  — SQL das três bases lógicas (DAD-01)
  api/                        — OpenAPI (público e interno) e AsyncAPI (eventos/comandos)
  deploy/
    docker-compose.yml        — stack local completa
    Dockerfile                — build genérico por serviço (build arg SERVICE)
    postgres-init/            — cria as 3 bases e aplica as migrations no boot do container
    seed/                     — dados sintéticos mínimos para o ensaio local
    kong/kong.yml              — configuração declarativa do gateway
    k8s/                      — manifests Kubernetes de referência (não aplicados a nenhum cluster)
    terraform/                — IaC de referência para AWS (não aplicado a nenhuma conta)
  admin-ui/                   — interface administrativa (TypeScript + React) do Atlas
  test/e2e/                   — testes de integração ponta a ponta contra a stack local
```

## Como rodar localmente

```bash
cd hub/deploy
docker compose up -d --build
```

Isso sobe: PostgreSQL (com as 3 bases já migradas), LocalStack (SQS/SNS/S3), Kong, os
seis serviços Go e um job `seed` que popula um catálogo, contas de provedor simuladas
(síncrona, polling e callback), credenciais `SHARED_HUB`, dois contratos de tenant (um
comum e um com saldo estrito) e um destino de webhook apontando para o `webhook-sink`
(utilitário de teste, não faz parte do domínio do hub).

Serviços e portas:

| Serviço | Porta | Papel |
|---|---|---|
| orbita | 8080 | admissão, protocolo, consulta |
| atlas | 8081 | catálogo, contratos, credenciais |
| cometa | 8082 | execução externa, polling, credencial |
| pulsar | 8083 | entrega de webhook |
| libra | 8084 | financeiro |
| provider-sim | 8090 | provedor externo simulado (apenas ensaio) |
| webhook-sink | 8091 | destino de webhook de teste (apenas ensaio) |
| kong | 8000 / 8001 | gateway (proxy / admin) |
| postgres | 5432 | três bases: hub_control, hub_core, hub_finance |
| localstack | 4566 | SQS, SNS, S3 emulados |

## Identidade local

As APIs de negócio exigem `Authorization: Bearer` emitido pelo realm OIDC do
ambiente. `X-Tenant-Id`, `X-Application-Id` e `X-Cell-Id` são removidos pelo
middleware e nunca ampliam identidade. Para o laboratório R2, a reconciliação
do realm cria os usuários sintéticos descritos em
`deploy/r2/scripts/reconcile_identity.py`; a senha e o OTP devem ser fornecidos
por variáveis do ambiente de execução e não são versionados.

O portal usa Authorization Code + PKCE, mantém somente o access token em
memória e limpa o estado ao sair ou ao expirar a sessão.

## Exemplo manual (SYNC direto)

```bash
curl -X POST http://localhost:8080/v1/protocols \
  -H 'Authorization: Bearer <access-token-oidc>' \
  -H 'Content-Type: application/json' -H 'Idempotency-Key: exemplo-001' \
  -d '{"mode":"SYNC","provider_account_id":"prov-sync-1","service_code":"consulta-cadastral","service_version":1,"input":{"cpf":"12345678900"}}'
```

## Testes automatizados

```bash
cd hub
go test ./...                       # testes unitários (não exigem a stack no ar)
go test -tags e2e ./test/e2e/...    # testes de integração — exigem a stack + seed no ar
```

## Limitações desta referência (ver auditoria completa)

- O laboratório local usa Keycloak e tokens OIDC sintéticos; homologação de
  provedores comerciais, PKI/mTLS de produção, AWS regional e políticas
  comerciais continuam pendentes de ambiente e aprovação próprios.
- Composição/agregação de múltiplos passos (CAT-03/04), planos comerciais avançados
  (faixas, franquia) e execução integrada de produtos compostos ainda exigem
  qualificação adicional; o executor DAG possui contrato e testes unitários,
  mas não é usado como despachante de produto no fluxo público.
- O laboratório resolve referências de segredo por fixture/LocalStack; a
  integração regional com Secrets Manager/KMS real ainda não foi homologada.
- Manifests Kubernetes e Terraform em `deploy/k8s` e `deploy/terraform` são
  **referência não aplicada** — não há cluster nem conta AWS disponível (P-01).
