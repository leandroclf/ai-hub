#!/bin/sh
# Popula dados sinteticos minimos para o ensaio local da fatia inicial
# (QUA-04): um servico, uma conta de provedor sincrona simulada, uma
# credencial SHARED_HUB, um contrato e um destino de webhook.
set -e

echo "aguardando Atlas, Pulsar e Cometa ficarem prontos..."
until curl -sf "$ATLAS_URL/healthz/ready" >/dev/null; do sleep 1; done
until curl -sf "$PULSAR_URL/healthz/ready" >/dev/null; do sleep 1; done
until curl -sf "$COMETA_URL/healthz/ready" >/dev/null; do sleep 1; done

echo "publicando servico demo (produto sincrono, com fallback assincrono)..."
curl -sf -X POST "$ATLAS_URL/v1/services" -H 'Content-Type: application/json' -d '{
  "code": "consulta-cadastral", "version": 1, "description": "servico sincrono de demonstracao",
  "modes": ["SYNC", "ASYNC", "AUTO"], "client_sla_seconds": 30, "retry_ttl_seconds": 15, "published": true
}'

echo "publicando produto assincrono (provedor so aceita ASYNC/AUTO)..."
curl -sf -X POST "$ATLAS_URL/v1/services" -H 'Content-Type: application/json' -d '{
  "code": "protocolo-assincrono", "version": 1, "description": "produto de provedor nativamente assincrono (polling/callback)",
  "modes": ["ASYNC", "AUTO"], "client_sla_seconds": 45, "retry_ttl_seconds": 20, "published": true
}'

echo "cadastrando conta de provedor sincrona (SYNC direto)..."
curl -sf -X POST "$ATLAS_URL/v1/provider-accounts" -H 'Content-Type: application/json' -d '{
  "provider_account_id": "prov-sync-1", "provider_id": "provider-sim", "environment": "local",
  "base_url": "http://provider-sim:8090", "provider_mode": "sync"
}'

echo "cadastrando conta de provedor assincrona (polling)..."
curl -sf -X POST "$ATLAS_URL/v1/provider-accounts" -H 'Content-Type: application/json' -d '{
  "provider_account_id": "prov-poll-1", "provider_id": "provider-sim", "environment": "local",
  "base_url": "http://provider-sim:8090", "provider_mode": "async_poll"
}'

echo "cadastrando conta de provedor assincrona (callback)..."
curl -sf -X POST "$ATLAS_URL/v1/provider-accounts" -H 'Content-Type: application/json' -d '{
  "provider_account_id": "prov-callback-1", "provider_id": "provider-sim", "environment": "local",
  "base_url": "http://provider-sim:8090", "provider_mode": "async_callback"
}'

echo "cadastrando segunda conta de provedor assincrona (polling, provedor B)..."
curl -sf -X POST "$ATLAS_URL/v1/provider-accounts" -H 'Content-Type: application/json' -d '{
  "provider_account_id": "prov-poll-2", "provider_id": "provider-sim-b", "environment": "local",
  "base_url": "http://provider-sim:8090", "provider_mode": "async_poll"
}'

echo "cadastrando credencial compartilhada (SHARED_HUB)..."
for acc in prov-sync-1 prov-poll-1 prov-callback-1 prov-poll-2; do
  curl -sf -X POST "$ATLAS_URL/v1/credential-bindings" -H 'Content-Type: application/json' -d "{
    \"binding_id\": \"bind-shared-$acc\", \"credential_mode\": \"SHARED_HUB\",
    \"provider_account_id\": \"$acc\", \"secret_ref\": \"vault://shared/$acc\",
    \"settlement_party\": \"HUB\", \"state\": \"ATIVO\"
  }"
done

echo "cadastrando contrato do tenant demo (acme)..."
curl -sf -X POST "$ATLAS_URL/v1/contracts" -H 'Content-Type: application/json' -d '{
  "tenant_id": "acme", "plan": "unit", "unit_price": 1.00, "strict_balance": false,
  "client_sla_seconds": 30, "credential_mode_required": "SHARED_HUB"
}'

echo "cadastrando contrato do tenant demo com saldo estrito (acme-strict)..."
curl -sf -X POST "$ATLAS_URL/v1/contracts" -H 'Content-Type: application/json' -d '{
  "tenant_id": "acme-strict", "plan": "unit", "unit_price": 1.00, "strict_balance": true,
  "client_sla_seconds": 30, "credential_mode_required": "SHARED_HUB"
}'

echo "cadastrando contrato do tenant com credencial dedicada (acme-dedicated)..."
curl -sf -X POST "$ATLAS_URL/v1/contracts" -H 'Content-Type: application/json' -d '{
  "tenant_id": "acme-dedicated", "plan": "unit", "unit_price": 1.00, "strict_balance": false,
  "client_sla_seconds": 30, "credential_mode_required": "TENANT_DEDICATED"
}'

echo "cadastrando credencial dedicada (TENANT_DEDICATED) do tenant acme-dedicated..."
curl -sf -X POST "$ATLAS_URL/v1/credential-bindings" -H 'Content-Type: application/json' -d '{
  "binding_id": "bind-dedicated-acme", "credential_mode": "TENANT_DEDICATED", "tenant_id": "acme-dedicated",
  "provider_account_id": "prov-sync-1", "secret_ref": "vault://dedicated/acme", "settlement_party": "HUB", "state": "ATIVO"
}'

echo "cadastrando destino de webhook do tenant acme (webhook-sink)..."
curl -sf -X POST "$PULSAR_URL/internal/destinations" -H 'Content-Type: application/json' -d '{
  "tenant_id": "acme", "url": "http://webhook-sink:8091/webhook", "hmac_secret": "s3gr3d0-local"
}'

echo "seed concluido."
