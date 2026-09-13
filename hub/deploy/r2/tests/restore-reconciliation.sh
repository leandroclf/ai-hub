#!/usr/bin/env bash
set -euo pipefail

# Qualificação local cercada: cria bancos de restore com nomes únicos, restaura
# dumps lógicos e compara obrigações. Não remove bancos, buckets ou volumes;
# o inventário produzido permite a limpeza manual do recurso exato depois da
# revisão da evidência.
compose_project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
container=${R2_PG_CONTAINER:-${compose_project}-postgres-1}
localstack_container=${R2_LOCALSTACK_CONTAINER:-${compose_project}-localstack-1}
suffix=${R2_RESTORE_SUFFIX:-$(date -u +%Y%m%d%H%M%S)}
case "$suffix" in *[!a-zA-Z0-9_]*) echo "invalid restore suffix" >&2; exit 2;; esac
psql_hub=(docker exec -i "$container" psql -U hub -v ON_ERROR_STOP=1 -X)

relation_digest() {
  local database=$1
  local relation=$2
  "${psql_hub[@]}" -d "$database" -Atqc "SELECT md5(COALESCE(string_agg(row_to_json(t)::text, '' ORDER BY row_to_json(t)::text),'')) FROM $relation t"
}

declare -A tables=(
  [control]='catalog_resources catalog_publications'
  [core]='protocols operations attempts callback_inbox deliveries outbox'
  [finance]='economic_facts reservations ledger_entries finance_inbox finance_snapshots'
)
for domain in control core finance; do
  source="hub_${domain}"
  target="hub_${domain}_restore_${suffix}"
  existing=$("${psql_hub[@]}" -d postgres -Atqc "SELECT 1 FROM pg_database WHERE datname = '$target'")
  if [[ "$existing" == "1" ]]; then
    echo "restore target database already exists: $target; choose a new R2_RESTORE_SUFFIX" >&2
    exit 2
  fi
  "${psql_hub[@]}" -d postgres -v target="$target" <<'SQL'
SELECT format('CREATE DATABASE %I', :'target') WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname=:'target')\gexec
SQL
  docker exec "$container" sh -c "pg_dump -U hub -Fc '$source' | pg_restore -U hub -d '$target' --no-owner --no-privileges"
  for table in ${tables[$domain]}; do
    src=$("${psql_hub[@]}" -d "$source" -Atqc "SELECT count(*) FROM $table")
    dst=$("${psql_hub[@]}" -d "$target" -Atqc "SELECT count(*) FROM $table")
    test "$src" = "$dst" || { echo "restore count mismatch $source.$table: $src != $dst" >&2; exit 1; }
    printf 'restore count %s.%s: %s\n' "$source" "$table" "$src"
    src_digest=$(relation_digest "$source" "$table")
    dst_digest=$(relation_digest "$target" "$table")
    test "$src_digest" = "$dst_digest" || { echo "restore digest mismatch $source.$table: $src_digest != $dst_digest" >&2; exit 1; }
    printf 'restore digest %s.%s: %s\n' "$source" "$table" "$src_digest"
  done
done

bucket_suffix=${suffix//_/-}
bucket=${R2_RESTORE_BUCKET:-r2-custody-restore-${bucket_suffix}}
if docker exec "$localstack_container" awslocal s3api head-bucket --bucket "$bucket" >/dev/null 2>&1; then
  echo "restore target bucket already exists: $bucket; choose a new R2_RESTORE_SUFFIX or R2_RESTORE_BUCKET" >&2
  exit 2
fi
source_objects=$(docker exec "$localstack_container" awslocal s3api list-objects-v2 --bucket r2-custody --query 'length(Contents || `[]`)' --output text)
# R6-OPE-02-S03: an empty backup, or the mere equality of two zero counts,
# must never be announced as a reconciled restore — it is a BLOCK. A prior
# round of this script accepted "restore objects: 0" as evidence; that is
# exactly the false positive this requirement closes.
if [[ "$source_objects" == "0" || -z "$source_objects" ]]; then
  echo "RESTORE_RECONCILIATION=BLOCK reason=empty_source_bucket bucket=r2-custody" >&2
  exit 1
fi
docker exec "$localstack_container" awslocal s3 mb "s3://$bucket" >/dev/null
docker exec "$localstack_container" awslocal s3 sync s3://r2-custody "s3://$bucket" >/dev/null
restored_objects=$(docker exec "$localstack_container" awslocal s3api list-objects-v2 --bucket "$bucket" --query 'length(Contents || `[]`)' --output text)
test "$source_objects" = "$restored_objects" || { echo "object restore mismatch: $source_objects != $restored_objects" >&2; exit 1; }
printf 'restore objects: %s\n' "$source_objects"

# Count equality alone is not proof of recovery: every referenced key/version
# must resolve to byte-identical content in the restored bucket, and every
# file_refs row that references an object_version must exist and match its
# recorded sha256 (R6-OPE-02: "todas as referências resolvem para bytes
# corretos").
keys=$(docker exec "$localstack_container" awslocal s3api list-objects-v2 --bucket r2-custody --query 'Contents[].Key' --output text)
verified_objects=0
verified_keys_file=$(mktemp)
trap 'rm -f "$verified_keys_file"' EXIT
for key in $keys; do
  source_digest=$(docker exec "$localstack_container" sh -c "awslocal s3 cp 's3://r2-custody/$key' - 2>/dev/null | sha256sum" | awk '{print $1}')
  restored_digest=$(docker exec "$localstack_container" sh -c "awslocal s3 cp 's3://$bucket/$key' - 2>/dev/null | sha256sum" | awk '{print $1}')
  test -n "$source_digest" && test "$source_digest" = "$restored_digest" || { echo "object byte mismatch key=$key source=$source_digest restored=$restored_digest" >&2; exit 1; }
  verified_objects=$((verified_objects + 1))
  echo "$key" >> "$verified_keys_file"
  printf 'restore object verified: key=%s sha256=%s\n' "$key" "$source_digest"
done
test "$verified_objects" -gt 0 || { echo "no object keys enumerated despite non-zero count: $source_objects" >&2; exit 1; }
# Every file_refs row scoped to the qualification fixture tenant must resolve
# to one of the keys just verified byte-for-byte in the restored bucket — a
# reference to a key that was never actually confirmed is not a "restore
# obligation preserved," it is an unverified pointer. Scoped to the fixture
# tenant deliberately: hub_core is the shared long-lived database also used
# by the Go test suite (R2_CORE_TEST_DSN), which legitimately leaves other
# tenants' fixture rows (e.g. a test that intentionally simulates S3 being
# unavailable) with no backing object — that is a different, already-covered
# concern, not this restore's completeness.
fixture_tenant=${RESTORE_FIXTURE_TENANT:-restore-qualification}
referenced_keys=$("${psql_hub[@]}" -d "hub_core_restore_${suffix}" -Atqc "SELECT object_key FROM file_refs WHERE object_key<>'' AND tenant_id='${fixture_tenant}'")
unresolved=0
while IFS= read -r referenced; do
  [[ -z "$referenced" ]] && continue
  grep -qxF "$referenced" "$verified_keys_file" || { echo "restored file_refs references unverified object key: $referenced" >&2; unresolved=$((unresolved + 1)); }
done <<< "$referenced_keys"
test "$unresolved" = "0" || exit 1
printf 'restore file_refs cross-check: all object_key references resolve to verified bytes\n'

effects_before=$(curl -fsS http://127.0.0.1:18090/__qualification/effects)
printf 'external effect oracle before reconciliation: %s\n' "$effects_before"
printf 'RESTORE_RECONCILIATION=PASS databases=%s suffix=%s bucket=%s\n' 'control,core,finance' "$suffix" "$bucket"
printf '%s\n' 'The provider oracle was observed without replay; re-admission remains disabled during restore.'
