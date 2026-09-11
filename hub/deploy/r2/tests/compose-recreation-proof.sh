#!/usr/bin/env bash
set -euo pipefail

# Recria somente um serviço do Compose oficial, sem remover volumes, e
# confronta a custódia persistente antes/depois. A prova não reinicia o
# ecossistema inteiro nem declara HA regional.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
PROJECT="${R2_COMPOSE_PROJECT:-ai_hub_r3qual}"
COMPOSE_FILE="${R2_COMPOSE_FILE:-${ROOT_DIR}/hub/deploy/r2/compose.yaml}"
SERVICE="${R2_RECREATE_SERVICE:-atlas}"
ATLAS_URL="${R2_ATLAS_URL:-http://127.0.0.1:18081}"
EVIDENCE_FILE="${R2_COMPOSE_RECREATION_EVIDENCE_FILE:-${ROOT_DIR}/hub/evidence/r2/execution/compose-recreation-latest.log}"

mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

compose=(docker compose --project-name "$PROJECT" --file "$COMPOSE_FILE")
fail() {
  echo "COMPOSE_RECREATION_PROOF=FAIL reason=$*" >&2
  exit 1
}

container_id() {
  "${compose[@]}" ps -q "$SERVICE"
}

postgres_id="$("${compose[@]}" ps -q postgres)"
[[ -n "$postgres_id" ]] || fail "container postgres ausente no projeto oficial"
before_id="$(container_id)"
[[ -n "$before_id" ]] || fail "serviço ${SERVICE} não está em execução"

catalog_snapshot() {
  docker exec "$postgres_id" psql -U hub -d hub_control -Atc \
    "select kind||'='||count(*) from catalog_resources group by kind order by kind;"
}

protocol_count() {
  docker exec "$postgres_id" psql -U hub -d hub_core -Atc \
    "select count(*) from protocols;"
}

wait_ready() {
  local deadline=$(( $(date +%s) + 60 ))
  while (( $(date +%s) < deadline )); do
    if curl -fsS "$ATLAS_URL/healthz/ready" >/dev/null; then
      return 0
    fi
    sleep 1
  done
  fail "Atlas não voltou a ficar pronto em 60 segundos"
}

before_catalog="$(catalog_snapshot)"
before_protocols="$(protocol_count)"
before_created="$(docker inspect -f '{{.Created}}' "$before_id")"

echo "COMPOSE_RECREATION_PROOF=START"
echo "project=${PROJECT} service=${SERVICE} before_container=${before_id}"
echo "before_created=${before_created} before_protocols=${before_protocols}"
echo "before_catalog=$(tr '\n' ',' <<<"$before_catalog" | sed 's/,$//')"
echo "action=up --detach --no-deps --force-recreate ${SERVICE} volumes=preserved"
"${compose[@]}" up -d --no-deps --force-recreate "$SERVICE"
wait_ready

after_id="$(container_id)"
[[ -n "$after_id" && "$after_id" != "$before_id" ]] || fail "container não foi recriado: before=${before_id} after=${after_id}"
after_catalog="$(catalog_snapshot)"
after_protocols="$(protocol_count)"
after_created="$(docker inspect -f '{{.Created}}' "$after_id")"

[[ "$after_catalog" == "$before_catalog" ]] || fail "catálogo mudou após recriação"
[[ "$after_protocols" == "$before_protocols" ]] || fail "protocolos mudaram após recriação"

echo "after_container=${after_id} after_created=${after_created} after_protocols=${after_protocols}"
echo "after_catalog=$(tr '\n' ',' <<<"$after_catalog" | sed 's/,$//')"
echo "COMPOSE_RECREATION_PROOF=PASS service=${SERVICE} container_changed=true catalog_preserved=true protocols_preserved=true"
