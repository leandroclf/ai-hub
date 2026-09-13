# CHECKPOINT — R5 (retomada exata)

## Estado no fechamento desta sessão
- SHA base: `f87ce33034ae29c9431b1910dcc6a633b545e330` (main, working tree sujo, sem commits novos criados nesta sessão — nenhum push/merge realizado, conforme instrução).
- `git status --porcelain` no fechamento: os mesmos 21 arquivos modificados do início da sessão (já em progresso antes desta rodada), mais os 6 changes `openspec/changes/r5-*` e 5 migrações não rastreadas, **mais** as edições desta sessão sobre esses mesmos arquivos (não há arquivos novos além dos já listados no `git status` inicial). `docs/reviews/2026-09-12-r5/implementation/` é novo (criado nesta sessão).
- Toolchain local: Go `1.24.0` (linux/amd64) — próximo do Go `1.24.13` usado na auditoria original; **diverge** do `go 1.26`/base Nginx atual do `hub/deploy/Dockerfile` (gap OPE-02/qualificação de imagem, não fechado nesta sessão).

## Laboratório ativo (recursos exatos abertos por esta sessão)
Ecossistema Compose: **hub-local** (`hub/deploy/docker-compose.yml`), subido parcialmente — apenas os serviços de dados, sem as imagens de aplicação:
- `hub-local-postgres-1` (postgres:16-alpine, porta 5432, init aplica `hub/migrations/{control,core,finance}/*.sql` via `hub/deploy/postgres-init/01-init.sh`).
- `hub-local-redis-1` (redis:7.4-alpine, porta 6379).
- `hub-local-localstack-1` — **recriado manualmente** fora do compose (`docker run -d --name hub-local-localstack-1 -p 4566:4566 -e SERVICES=sqs,sns,s3,secretsmanager ...`) porque o `docker-compose.yml` original só habilita `sqs,sns,s3` e os testes de vault precisam de `secretsmanager`. Fixture criada: secret `r2/provider/fixture` = `r2-synthetic-provider-password`.
- Verificado antes de subir: `docker --context default compose ls` e `docker ps -a` — nenhum outro ecossistema `hub-local` ativo. O único outro projeto Docker Compose rodando na máquina é `hive-ia-agents` (repositório diferente), **preservado intocado**.
- Nota de acesso: o contexto Docker padrão (`desktop-linux`) não respondia (`dial unix .../docker.sock`); o contexto `default` (socket em `/var/run/docker.sock`) funciona e foi o usado. Scripts do repo que chamam `docker`/`docker compose` sem `--context` precisam de `DOCKER_CONTEXT=default` no ambiente para funcionar nesta máquina.

## Para retomar
```bash
cd hub/deploy && docker --context default compose ps   # confirma postgres/redis vivos
# se localstack não tiver secretsmanager, recriar:
docker --context default compose stop localstack && docker --context default compose rm -f localstack
docker --context default run -d --name hub-local-localstack-1 -p 4566:4566 \
  -e SERVICES=sqs,sns,s3,secretsmanager -e DEFAULT_REGION=us-east-1 localstack/localstack:3.8
docker --context default exec hub-local-localstack-1 awslocal secretsmanager create-secret \
  --name r2/provider/fixture --secret-string r2-synthetic-provider-password

cd ../../hub
env R2_FINANCE_DSN="postgres://hub:hub@localhost:5432/hub_finance?sslmode=disable" \
    R2_CORE_TEST_DSN="postgres://hub:hub@localhost:5432/hub_core?sslmode=disable" \
    R2_CONTROL_TEST_DSN="postgres://hub:hub@localhost:5432/hub_control?sslmode=disable" \
    ATLAS_TEST_DSN="postgres://hub:hub@localhost:5432/hub_control?sslmode=disable" \
    R2_S3_TEST_ENDPOINT="http://localhost:4566" R2_TEST_SECRET_REF="r2/provider/fixture" \
    SECRETS_ENDPOINT="http://localhost:4566" ENVIRONMENT="local" \
    AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 \
    go test ./... -race
```

## Próximo passo concreto
Continuar pela ordem do `EXECUTION_PLAN.md`: fechar R5-SEG-01/SEG-03 com prova end-to-end dos 3 cenários cada (hoje só S01 de SEG-02 tem teste de regressão novo; SEG-01/SEG-03 têm implementação real mas sem teste de cenário adversarial dedicado escrito nesta sessão), depois EXE-02..05, DAD-01/02, OPE-01/03, UX-01, QUA-01 — nessa ordem, conforme `docs/reviews/2026-09-12-r5/02-PLANO-DE-IMPLEMENTACAO.md`.
