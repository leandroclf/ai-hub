#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
compose=(docker compose -p "$project" -f "$root/compose.yaml")
# Autoriza apenas endereços das três fixtures desta stack; nenhuma credencial real.
"${compose[@]}" up -d postgres localstack identity
# O import do realm é somente inicial; esta etapa torna usuários, OTP, claims e
# mapper de subject determinísticos também quando o volume do Keycloak já existe.
"${compose[@]}" run --rm identity-reconcile
# Aplicação de migration explícita antes das APIs; nova versão não reexecuta DDL já confirmado.
"${compose[@]}" run --rm migrate
"${compose[@]}" up -d --build provider-sim webhook-sink
# Fixture de cofre exclusivamente local, idempotente e sem valor comercial.
localstack_container="${project}-localstack-1"
if ! docker exec "$localstack_container" awslocal secretsmanager describe-secret --secret-id r2/provider/fixture >/dev/null 2>&1; then
  docker exec "$localstack_container" awslocal secretsmanager create-secret \
    --name r2/provider/fixture --secret-string r2-synthetic-provider-password >/dev/null
fi
export R2_PROVIDER_CIDR R2_SINK_CIDR R2_IDENTITY_CIDR
network="${project}_default"
R2_PROVIDER_CIDR=$(docker inspect "${project}-provider-sim-1" --format "{{(index .NetworkSettings.Networks \"$network\").IPAddress}}")/32
R2_SINK_CIDR=$(docker inspect "${project}-webhook-sink-1" --format "{{(index .NetworkSettings.Networks \"$network\").IPAddress}}")/32
R2_IDENTITY_CIDR=$(docker inspect "${project}-identity-1" --format "{{(index .NetworkSettings.Networks \"$network\").IPAddress}}")/32
"${compose[@]}" up -d --build
printf '%s\n' 'UI http://localhost:13000 | IdP http://localhost:18085 | Grafana http://localhost:13001'
