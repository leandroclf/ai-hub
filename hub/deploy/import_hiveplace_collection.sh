#!/bin/sh
set -eu

# Importa uma collection Postman como catálogo de produtos/endpoints.
# O environment é usado somente para ler baseUrl; segredos são convertidos
# em referências e nunca são enviados ao Atlas nem persistidos.
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
COLLECTION=${1:-"$SCRIPT_DIR/../../postman/collections/Hive.postman_collection.json"}
ENVIRONMENT=${2:-"$SCRIPT_DIR/../../postman/environments/Hive.postman_env.json"}
COMPOSE_FILE=${COMPOSE_FILE:-"$SCRIPT_DIR/docker-compose.yml"}
ATLAS_URL=${ATLAS_URL:-"http://localhost:8081"}

test -r "$COLLECTION" || { echo "collection não encontrada: $COLLECTION" >&2; exit 1; }
test -r "$ENVIRONMENT" || { echo "environment não encontrado: $ENVIRONMENT" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq é obrigatório" >&2; exit 1; }

BASE_URL=$(jq -r '.values[] | select(.key == "baseUrl" and .enabled == true) | .value' "$ENVIRONMENT" | head -n 1)
test -n "$BASE_URL" && test "$BASE_URL" != "null" || { echo "baseUrl habilitada não encontrada" >&2; exit 1; }

echo "aplicando migration do catálogo de endpoints..."
docker compose -f "$COMPOSE_FILE" exec -T postgres psql -v ON_ERROR_STOP=1 -U hub -d hub_control -f /migrations/control/0003_provider_api_catalog.sql >/dev/null

WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT

jq -r '.item | to_entries[] | [.key, .value.name] | @tsv' "$COLLECTION" > "$WORK_DIR/folders.tsv"
jq -r '
  def requests($index; $folder):
    .[]? | if type != "object" then empty elif has("request") then
      [$index, $folder, .name, .request.method, (.request.url.raw // ""),
       (.request.auth.type // "none"),
       ((.request.header // []) | map(if type == "object" then (.key | tostring) else tostring end) | join(",")),
       ((.request.url.query // []) | map(if type == "object" then (.key | tostring) else tostring end) | join(","))] | @tsv
    elif has("item") then .item | requests($index; $folder) else empty end;
  .item | to_entries[] | requests(.key; .value.name)
' "$COLLECTION" > "$WORK_DIR/endpoints.tsv"

# O catálogo anterior é exclusivamente local/de teste. Históricos em hub_core e
# hub_finance não são removidos; apenas configurações do plano de controle são trocadas.
docker compose -f "$COMPOSE_FILE" exec -T postgres psql -v ON_ERROR_STOP=1 -U hub -d hub_control <<'SQL'
BEGIN;
DELETE FROM provider_api_catalog;
DELETE FROM credential_bindings;
DELETE FROM contracts;
DELETE FROM services;
DELETE FROM provider_accounts;
COMMIT;
SQL

echo "cadastrando conta HivePlace HML sem transportar segredos..."
PROVIDER_PAYLOAD=$(jq -n --arg base "$BASE_URL" '{provider_account_id:"provider-hiveplace-hml",provider_id:"HivePlace",environment:"HML",base_url:$base,provider_mode:"sync",auth_type:"OAUTH_CLIENT_CREDENTIALS",oauth_token_url:($base + "/security/oauth2/token"),oauth_client_id:"postman-Hive-HML-userName",oauth_client_secret_ref:"postman-Hive-HML-userPass",token_ttl_seconds:300}')
curl -sf -X POST "$ATLAS_URL/v1/provider-accounts" -H 'Content-Type: application/json' -d "$PROVIDER_PAYLOAD" >/dev/null

TAB=$(printf '\t')
while IFS="$TAB" read -r index folder; do
  code=$(printf '%s' "$folder" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g; s/--*/-/g; s/^-//; s/-$//')
  code=$(printf 'hiveplace-%02d-%s' "$((index + 1))" "$code")
  PRODUCT_PAYLOAD=$(jq -n --arg code "$code" --arg folder "$folder" '{code:$code,version:1,description:("HivePlace API — " + $folder),modes:["SYNC","ASYNC","AUTO"],client_sla_seconds:30,retry_ttl_seconds:10,published:true}')
  curl -sf -X POST "$ATLAS_URL/v1/services" -H 'Content-Type: application/json' -d "$PRODUCT_PAYLOAD" >/dev/null
  printf '%s\t%s\n' "$index" "$code" >> "$WORK_DIR/product_map.tsv"
done < "$WORK_DIR/folders.tsv"

awk -F '\t' 'NR==FNR {map[$1]=$2; next} {
  auth=toupper($6); if (auth == "APIKEY") auth="API_KEY"; else if (auth != "BASIC" && auth != "BEARER") auth="NONE";
  metadata=sprintf("{\"header_names\":\"%s\",\"query_names\":\"%s\"}", $7, $8);
  printf "hiveplace-%s-%05d\t%s\t1\tprovider-hiveplace-hml\tHive\t%s\t%s\t%s\t%s\t%s\t%s\n", $1, NR, map[$1], $2, $3, $4, $5, auth, metadata;
}' "$WORK_DIR/product_map.tsv" "$WORK_DIR/endpoints.tsv" > "$WORK_DIR/catalog.tsv"

docker compose -f "$COMPOSE_FILE" exec -T postgres psql -v ON_ERROR_STOP=1 -U hub -d hub_control <<'SQL'
INSERT INTO credential_bindings (binding_id, credential_mode, provider_account_id, secret_ref, settlement_party, state)
VALUES ('bind-shared-provider-hiveplace-hml', 'SHARED_HUB', 'provider-hiveplace-hml', 'postman:Hive:HML Tec:credentials', 'HUB', 'ATIVO');
INSERT INTO contracts (tenant_id, plan, unit_price, strict_balance, client_sla_seconds, credential_mode_required)
VALUES ('hiveplace-sandbox', 'unit', 1.0000, false, 30, 'SHARED_HUB');
SQL
docker compose -f "$COMPOSE_FILE" exec -T postgres psql -v ON_ERROR_STOP=1 -U hub -d hub_control -c '\copy provider_api_catalog (api_id, product_code, product_version, provider_account_id, source_collection, source_folder, request_name, http_method, path_template, auth_type, request_metadata) FROM STDIN WITH (FORMAT text)' < "$WORK_DIR/catalog.tsv"

echo "importação concluída: $(wc -l < "$WORK_DIR/folders.tsv") produtos e $(wc -l < "$WORK_DIR/endpoints.tsv") endpoints; provider=HivePlace"
