# AI Hub — interoperabilidade de serviços

Implementação em Go sob [`hub/`](hub/), com PostgreSQL, filas SQS/SNS, armazenamento S3 e contratos OpenAPI/AsyncAPI. O ambiente local usa Compose/LocalStack; configurações e evidências não certificam produção ou contratos comerciais reais.

## Começar pela documentação

- [Implementação e comandos locais](hub/README.md)
- [Auditoria da fatia inicial de setembro](IMPLEMENTATION_AUDIT.md): registro histórico; comparar com o código e migrações atuais antes de usar seus limites como estado atual.
- [Contratos](hub/api/) e [migrações](hub/migrations/)
- [Instruções operacionais](AGENTS.md): manter um único ecossistema Compose, preservar volumes e não iniciar outro ambiente sem inventário.

O manifest [`hub/go.mod`](hub/go.mod) declara Go 1.24. Verificação de código: `cd hub` e `go test ./...`; testes de integração e configuração do ambiente seguem o guia do pacote. Não executar comandos de produção por inferência.

## Fronteira com Product Platform

Este serviço trata de interoperabilidade, protocolos e entrega a provedores; não é um gateway de modelos nem executor de engenharia. A autoridade de construção de produtos é [product-platform](https://github.com/leandroclf/product-platform). Não importar runtime, memória ou ledger do hub para o harness por semelhança de nomes. Reutilização exige contrato explícito, semântica de tenant, idempotência, recuperação e consumidores comprovados. Veja o [estudo de consolidação](https://github.com/leandroclf/product-platform/blob/main/docs/REPOSITORY-CONSOLIDATION.md).
