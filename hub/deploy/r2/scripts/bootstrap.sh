#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
# Autoriza apenas endereços das três fixtures desta stack; nenhuma credencial real.
docker compose -f "$root/compose.yaml" up -d postgres localstack identity
# Aplicação de migration explícita antes das APIs; nova versão não reexecuta DDL já confirmado.
docker compose -f "$root/compose.yaml" run --rm migrate
docker compose -f "$root/compose.yaml" up -d --build provider-sim webhook-sink
export R2_PROVIDER_CIDR R2_SINK_CIDR R2_IDENTITY_CIDR
R2_PROVIDER_CIDR=$(docker inspect ai-hub-r2-provider-sim-1 --format '{{(index .NetworkSettings.Networks "ai-hub-r2_default").IPAddress}}')/32
R2_SINK_CIDR=$(docker inspect ai-hub-r2-webhook-sink-1 --format '{{(index .NetworkSettings.Networks "ai-hub-r2_default").IPAddress}}')/32
R2_IDENTITY_CIDR=$(docker inspect ai-hub-r2-identity-1 --format '{{(index .NetworkSettings.Networks "ai-hub-r2_default").IPAddress}}')/32
docker compose -f "$root/compose.yaml" up -d --build
printf '%s\n' 'UI http://localhost:13000 | IdP http://localhost:18085 | Grafana http://localhost:13001'
