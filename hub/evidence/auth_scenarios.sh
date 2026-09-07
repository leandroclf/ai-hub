#!/bin/bash
# Ensaios de autenticacao outbound; nunca imprime token ou segredo.
set -euo pipefail
OUT="${OUT:-auth_scenarios_output.txt}"
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
COMPOSE=(docker compose -f "$SCRIPT_DIR/../deploy/docker-compose.yml")
RUN_ID="$(date +%s)"
: > "$OUT"

case_run() {
  local label="$1" tenant="$2" provider="$3" key="$4"
  printf '### %s tenant=%s provider=%s\n' "$label" "$tenant" "$provider" | tee -a "$OUT"
  curl -fsS -w '\nHTTP %{http_code}\n' -X POST http://localhost:8080/v1/protocols \
    -H 'Content-Type: application/json' -H "X-Tenant-Id: $tenant" -H "Idempotency-Key: $key" \
    -d "{\"mode\":\"SYNC\",\"provider_account_id\":\"$provider\",\"service_code\":\"consulta-cadastral\",\"service_version\":1,\"input\":{\"scenario\":\"$label\"}}" | tee -a "$OUT"
  printf '\n' >> "$OUT"
}

case_run BASIC acme prov-basic-1 "auth-basic-evidence-$RUN_ID"
case_run OAUTH_FIRST acme prov-oauth-1 "auth-oauth-first-evidence-$RUN_ID"
case_run OAUTH_CACHE_HIT acme prov-oauth-1 "auth-oauth-cache-evidence-$RUN_ID"
case_run MTLS_OAUTH acme prov-mtls-oauth-1 "auth-mtls-oauth-evidence-$RUN_ID"
case_run DEDICATED_OAUTH acme-dedicated-oauth prov-oauth-1 "auth-dedicated-oauth-evidence-$RUN_ID"

printf '### redis token metadata (valores nunca exibidos)\n' | tee -a "$OUT"
keys=$("${COMPOSE[@]}" exec -T redis redis-cli --scan --pattern 'hub:provider-token:*')
for key in $keys; do
	printf '%s ttl=' "$key" | tee -a "$OUT"
	"${COMPOSE[@]}" exec -T redis redis-cli TTL "$key" | tee -a "$OUT"
done
