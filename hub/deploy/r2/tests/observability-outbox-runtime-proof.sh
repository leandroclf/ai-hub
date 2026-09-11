#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
PROJECT="${R2_COMPOSE_PROJECT:-ai_hub_r3qual}"
COMPOSE=(docker compose -p "$PROJECT" -f "$ROOT_DIR/hub/deploy/r2/compose.yaml")
POSTGRES="${PROJECT}-postgres-1"
PROMETHEUS_URL="${R2_PROMETHEUS_URL:-http://127.0.0.1:19090}"
EVIDENCE_FILE="${R2_OBSERVABILITY_OUTBOX_EVIDENCE_FILE:-$ROOT_DIR/hub/evidence/r2/execution/observability-outbox-runtime-latest.log}"
mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

ROW_ID="observability-outbox-proof"
cleanup() {
  set +e
  docker exec "$POSTGRES" psql -U hub_runtime -d hub_core -v ON_ERROR_STOP=1 -q -c "DELETE FROM outbox WHERE aggregate_type='protocol' AND aggregate_id='${ROW_ID}';" >/dev/null 2>&1
  "${COMPOSE[@]}" up -d localstack >/dev/null 2>&1
}
trap cleanup EXIT

echo "OBSERVABILITY_OUTBOX_RUNTIME=START"
"${COMPOSE[@]}" stop localstack >/dev/null
docker exec "$POSTGRES" psql -U hub_runtime -d hub_core -v ON_ERROR_STOP=1 -q -c "INSERT INTO outbox(aggregate_type,aggregate_id,event_type,payload,published,created_at) VALUES('protocol','${ROW_ID}','observability.proof','{}',FALSE,clock_timestamp()-interval '60 seconds');"

firing="false"
for _ in $(seq 1 12); do
  metric="$(curl -fsS --max-time 5 http://127.0.0.1:18080/metrics | awk -F' ' '/hub_obligation_age_seconds\{component="orbita",kind="outbox"\}/{print $2; exit}')"
  query="$(curl -fsS --max-time 5 -G --data-urlencode 'query=ALERTS{alertname="HubOutboxDelayed",alertstate="firing"}' "$PROMETHEUS_URL/api/v1/query")"
  firing="$(python3 -c 'import json,sys; print("true" if json.load(sys.stdin).get("data",{}).get("result") else "false")' <<<"$query")"
  echo "sample metric_age_seconds=${metric:-missing} alert_firing=${firing}"
  if [[ "$firing" == true && "${metric:-0}" =~ ^[0-9]+([.][0-9]+)?$ ]] && awk "BEGIN {exit !($metric > 30)}"; then
    echo "OBSERVABILITY_OUTBOX_RUNTIME=PASS outbox_age_metric=true alert_firing=true broker=unavailable"
    exit 0
  fi
  sleep 5
done
echo "OBSERVABILITY_OUTBOX_RUNTIME=FAIL outbox_age_metric=${metric:-missing} alert_firing=${firing}"
exit 1
