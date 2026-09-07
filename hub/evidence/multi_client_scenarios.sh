#!/bin/bash
# Ensaio reproduzível de consumo por vários clientes e provedores.
# Executa contra a stack local já iniciada e grava a saída literal para auditoria.
set -euo pipefail

ORBITA="${ORBITA:-http://localhost:8080}"
ATLAS="${ATLAS:-http://localhost:8081}"
OUT="${OUT:-multi_client_scenarios_output.txt}"
: > "$OUT"

register_service() {
  local code="$1" description="$2" modes="$3" sla="$4"
  curl -fsS -X POST "$ATLAS/v1/services" -H 'Content-Type: application/json' \
    -d "{\"code\":\"$code\",\"version\":1,\"description\":\"$description\",\"modes\":$modes,\"client_sla_seconds\":$sla,\"retry_ttl_seconds\":10,\"published\":true}" \
    | tee -a "$OUT"
  printf '\n' >> "$OUT"
  curl -fsS "$ATLAS/v1/services/$code/1" | tee -a "$OUT"
  printf '\n' >> "$OUT"
}

consume() {
  local tenant="$1" provider="$2" mode="$3" service="$4" key="$5" input="$6"
  printf '### tenant=%s provider=%s mode=%s service=%s\n' "$tenant" "$provider" "$mode" "$service" | tee -a "$OUT"
  curl -fsS -w '\nHTTP %{http_code}\n' -X POST "$ORBITA/v1/protocols" \
    -H 'Content-Type: application/json' -H "X-Tenant-Id: $tenant" -H "Idempotency-Key: $key" \
    -d "{\"mode\":\"$mode\",\"provider_account_id\":\"$provider\",\"service_code\":\"$service\",\"service_version\":1,\"input\":$input}" \
    | tee -a "$OUT"
  printf '\n' >> "$OUT"
}

register_service "consulta-cadastral-publica" "API publica validada de consulta cadastral" '["SYNC","ASYNC","AUTO"]' 30
register_service "protocolo-publico" "API publica validada de protocolo assíncrono" '["ASYNC","AUTO"]' 45

for tenant in cliente-alpha cliente-beta cliente-gamma; do
  curl -fsS -X POST "$ATLAS/v1/contracts" -H 'Content-Type: application/json' \
    -d "{\"tenant_id\":\"$tenant\",\"plan\":\"unit\",\"unit_price\":1.25,\"strict_balance\":false,\"client_sla_seconds\":30,\"credential_mode_required\":\"SHARED_HUB\"}" >> "$OUT"
  printf '\n' >> "$OUT"
done

consume cliente-alpha prov-sync-1 SYNC consulta-cadastral-publica multi-alpha-sync '{"cpf":"11111111111"}'
consume cliente-alpha prov-poll-1 ASYNC protocolo-publico multi-alpha-poll '{"delay_ms":500}'
consume cliente-beta prov-sync-1 SYNC consulta-cadastral-publica multi-beta-sync '{"cpf":"22222222222"}'
consume cliente-beta prov-poll-2 ASYNC protocolo-publico multi-beta-poll '{"delay_ms":500}'
consume cliente-gamma prov-callback-1 ASYNC protocolo-publico multi-gamma-callback '{"delay_ms":500}'
consume cliente-gamma prov-sync-1 SYNC consulta-cadastral-publica multi-gamma-fail '{"force_fail":true}'

sleep 3
printf '### metricas\n' | tee -a "$OUT"
curl -fsS "$ORBITA/metrics" | tee -a "$OUT"
printf '\ncenarios concluidos\n' | tee -a "$OUT"
