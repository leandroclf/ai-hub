#!/usr/bin/env bash
set -euo pipefail

# Exercita o worker Pulsar real do Compose com uma entrega sintética. O registro
# é removido ao final; o domínio de capacidade e seu histórico permanecem para
# auditoria local da execução.
project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
compose=(docker compose -p "$project" -f "$root/compose.yaml")
psql() { "${compose[@]}" exec -T postgres psql -U hub -d hub_core -Atqc "$1"; }

delivery_id=$(psql "SELECT gen_random_uuid()")
destination_id=$(psql "SELECT gen_random_uuid()")
protocol_id=$(psql "SELECT gen_random_uuid()")
event_id=$(psql "SELECT gen_random_uuid()")
body='{"runtime":"webhook-capacity"}'
body_hash=$(printf '%s' "$body" | sha256sum | awk '{print $1}')
secret_version=$("${compose[@]}" exec -T localstack awslocal secretsmanager get-secret-value --secret-id r2/provider/fixture --query VersionId --output text)

cleanup() {
  psql "DELETE FROM capacity_feedback WHERE permit_id IN (SELECT wa.id::text FROM webhook_attempts wa WHERE wa.delivery_id='$delivery_id')" >/dev/null || true
  psql "DELETE FROM capacity_permits WHERE evidence_ref IN ('webhook-http-receipt:$delivery_id','webhook-transport-unconfirmed:$delivery_id')" >/dev/null || true
  psql "DELETE FROM webhook_attempts WHERE delivery_id='$delivery_id'" >/dev/null || true
  psql "DELETE FROM deliveries WHERE delivery_id='$delivery_id'" >/dev/null || true
  psql "DELETE FROM webhook_destination_versions WHERE id='$destination_id'" >/dev/null || true
}
trap cleanup EXIT

psql "INSERT INTO webhook_destination_versions(id,version,tenant_id,cell_id,application_id,url,secret_ref,secret_version,max_attempts,timeout_seconds,state,created_by) VALUES('$destination_id',1,'acme','r2-cell-a','runtime','http://webhook-sink:8091/webhook','r2/provider/fixture','$secret_version',1,2,'ACTIVE','webhook-capacity-runtime')"
psql "INSERT INTO deliveries(delivery_id,protocol_id,event_id,destination_url,state,next_attempt_at,tenant_id,cell_id,destination_id,destination_version,representation,body_sha256) VALUES('$delivery_id','$protocol_id','$event_id','http://webhook-sink:8091/webhook','PENDING',clock_timestamp(),'acme','r2-cell-a','$destination_id',1,'$body','$body_hash')"

for _ in $(seq 1 20); do
  state=$(psql "SELECT state FROM deliveries WHERE delivery_id='$delivery_id'")
  if [[ "$state" == "DELIVERED" ]]; then
    break
  fi
  sleep 1
done

if [[ "$state" != "DELIVERED" ]]; then
  echo "delivery_state=$state" >&2
  psql "SELECT 'attempt=' || COALESCE(state,'') || '|status=' || COALESCE(http_status::text,'') || '|error=' || COALESCE(error_code,'') FROM webhook_attempts WHERE delivery_id='$delivery_id'" >&2 || true
  exit 1
fi

permit=$(psql "SELECT count(*) || '|open=' || count(*) FILTER (WHERE transport_open) || '|pending=' || count(*) FILTER (WHERE pending_external) FROM capacity_permits WHERE domain_id='r4-webhook' AND evidence_ref='webhook-http-receipt:$delivery_id'")
if [[ "$permit" != "1|open=0|pending=0" ]]; then
  echo "capacity_permit=$permit" >&2
  exit 1
fi
printf 'PASS delivery=%s capacity=%s\n' "$delivery_id" "$permit"
