#!/usr/bin/env bash
set -euo pipefail

# R2-FIN-06/4.1: executa a migration financeira em banco descartável com
# registro legado, confirmando precisão NUMERIC, proveniência conservadora e
# replay idempotente. O banco oficial nunca é usado nem alterado.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
POSTGRES="${R2_PG_CONTAINER:-ai_hub_r3qual-postgres-1}"
SUFFIX="$(date +%s)"
DATABASE="hub_finance_r2proof_${SUFFIX}"
EVIDENCE_FILE="${R2_FINANCE_MIGRATION_EVIDENCE_FILE:-$ROOT_DIR/hub/evidence/r2/execution/finance-migration-rollback-latest.log}"
mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

psql() {
  docker exec -i "$POSTGRES" psql -U hub -v ON_ERROR_STOP=1 "$@"
}

cleanup() {
  set +e
  docker exec "$POSTGRES" psql -U hub -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS \"$DATABASE\";" >/dev/null 2>&1
}
trap cleanup EXIT

echo "FINANCE_MIGRATION_ROLLBACK_PROOF=START database=${DATABASE}"
docker exec "$POSTGRES" psql -U hub -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$DATABASE\";" >/dev/null
psql -d "$DATABASE" < "$ROOT_DIR/hub/migrations/finance/0001_init.sql"
psql -d "$DATABASE" -c "INSERT INTO economic_facts(tenant_id,protocol_id,kind,meter,amount,currency) VALUES('legacy-acme','00000000-0000-4000-8000-000000000001','COST','SUBMITTED',9876543210.1234,'BRL');"
psql -d "$DATABASE" < "$ROOT_DIR/hub/migrations/finance/0030_auditable_finance.sql"

legacy="$(psql -At -F '|' -d "$DATABASE" -c "SELECT amount::text,provenance,COALESCE(snapshot::text,'') FROM economic_facts WHERE protocol_id='00000000-0000-4000-8000-000000000001';")"
types="$(psql -At -F '|' -d "$DATABASE" -c "SELECT table_name,numeric_precision,numeric_scale FROM information_schema.columns WHERE table_name IN ('economic_facts','reservations','credit_limits','ledger_entries') AND column_name IN ('amount','limit_amount') ORDER BY table_name;")"
psql -d "$DATABASE" < "$ROOT_DIR/hub/migrations/finance/0030_auditable_finance.sql" >/dev/null
replayed="$(psql -At -F '|' -d "$DATABASE" -c "SELECT count(*),min(amount::text),max(provenance) FROM economic_facts;")"

echo "legacy_record=${legacy}"
echo "numeric_columns=$(tr '\n' ';' <<<"$types")"
echo "replayed_summary=${replayed}"
test "$legacy" = '9876543210.12340000|LEGACY_UNVERIFIED|'
test "$(grep -c '^economic_facts|30|8$\|^reservations|30|8$\|^credit_limits|30|8$\|^ledger_entries|30|8$' <<<"$types")" = 4
test "$replayed" = '1|9876543210.12340000|LEGACY_UNVERIFIED'
echo "FINANCE_MIGRATION_ROLLBACK_PROOF=PASS exact_numeric=true conservative_provenance=true replay_idempotent=true official_database_untouched=true"
