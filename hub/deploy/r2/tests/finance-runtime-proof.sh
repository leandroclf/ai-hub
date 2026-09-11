#!/usr/bin/env bash
set -euo pipefail

# Oráculo somente leitura para a qualificação financeira do workload autorizado.
# Execute depois de authorized-load.mjs (ou de uma jornada real equivalente),
# para que a janela contenha fatos produzidos pelo runtime e não apenas fixtures.

project="${COMPOSE_PROJECT_NAME:-ai_hub_r3qual}"
compose=(docker compose -p "$project" -f "${COMPOSE_FILE:-hub/deploy/r2/compose.yaml}")
window="${WINDOW_SECONDS:-900}"

db() {
  local database="$1"
  shift
  "${compose[@]}" exec -T postgres psql -U hub -d "$database" -Atqc "$1"
}

core_window="clock_timestamp()-make_interval(secs=>$window)"
finance_window="clock_timestamp()-make_interval(secs=>$window)"
core_fixture="payload->>'tenant_id' IN ('acme','acme-strict','acme-dedicated')"
finance_fixture="tenant_id IN ('acme','acme-strict','acme-dedicated')"

outbox_total="$(db hub_core "SELECT count(*) FROM outbox WHERE created_at>$core_window AND $core_fixture")"
outbox_incidence_missing_attempt="$(db hub_core "SELECT count(*) FROM outbox WHERE created_at>$core_window AND $core_fixture AND payload->>'economic_kind' IN ('SUBMITTED','STATUS','FETCH') AND COALESCE(payload->>'attempt_id','')=''")"
outbox_invalid_incidence="$(db hub_core "SELECT count(*) FROM outbox WHERE created_at>$core_window AND $core_fixture AND COALESCE(payload->>'economic_kind','') NOT IN ('','SUBMITTED','STATUS','FETCH','SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED')")"
finance_facts="$(db hub_finance "SELECT count(*) FROM economic_facts WHERE occurred_at>$finance_window AND $finance_fixture")"
finance_cost="$(db hub_finance "SELECT count(*) FROM economic_facts WHERE occurred_at>$finance_window AND $finance_fixture AND kind='COST'")"
finance_revenue="$(db hub_finance "SELECT count(*) FROM economic_facts WHERE occurred_at>$finance_window AND $finance_fixture AND kind='REVENUE'")"
finance_inbox="$(db hub_finance "SELECT count(*) FROM finance_inbox WHERE recorded_at>$finance_window AND $finance_fixture")"
unbalanced="$(db hub_finance "SELECT count(*) FROM (SELECT batch_id,currency FROM ledger_entries GROUP BY batch_id,currency HAVING SUM(CASE direction WHEN 'DEBIT' THEN amount ELSE -amount END)<>0) batches")"
invalid_quarantine="$(db hub_finance "SELECT count(*) FROM finance_quarantine WHERE created_at>$finance_window AND reason='INVALID_ECONOMIC_INCIDENCE'")"

for value in "$outbox_total" "$outbox_incidence_missing_attempt" "$outbox_invalid_incidence" "$finance_facts" "$finance_cost" "$finance_revenue" "$finance_inbox" "$unbalanced" "$invalid_quarantine"; do
  [[ "$value" =~ ^[0-9]+$ ]] || { echo "FINANCE_RUNTIME_PROOF=FAIL invalid_numeric=$value"; exit 1; }
done

if (( outbox_total < 1 || outbox_incidence_missing_attempt != 0 || outbox_invalid_incidence != 0 || finance_facts < 1 || finance_cost < 1 || finance_revenue < 1 || finance_inbox < 1 || unbalanced != 0 || invalid_quarantine != 0 )); then
  printf 'FINANCE_RUNTIME_PROOF=FAIL window_seconds=%s outbox=%s missing_attempt=%s invalid_incidence=%s facts=%s cost=%s revenue=%s inbox=%s unbalanced=%s invalid_quarantine=%s\n' \
    "$window" "$outbox_total" "$outbox_incidence_missing_attempt" "$outbox_invalid_incidence" "$finance_facts" "$finance_cost" "$finance_revenue" "$finance_inbox" "$unbalanced" "$invalid_quarantine"
  exit 1
fi

printf 'FINANCE_RUNTIME_PROOF=PASS window_seconds=%s outbox=%s missing_attempt=%s invalid_incidence=%s facts=%s cost=%s revenue=%s inbox=%s unbalanced=%s invalid_quarantine=%s\n' \
  "$window" "$outbox_total" "$outbox_incidence_missing_attempt" "$outbox_invalid_incidence" "$finance_facts" "$finance_cost" "$finance_revenue" "$finance_inbox" "$unbalanced" "$invalid_quarantine"
