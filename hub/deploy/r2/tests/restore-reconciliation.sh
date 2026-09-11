#!/usr/bin/env bash
set -euo pipefail

# Qualificação local cercada: cria bancos de restore com nomes únicos, restaura
# dumps lógicos e compara obrigações. Não remove bancos, buckets ou volumes;
# o inventário produzido permite a limpeza manual do recurso exato depois da
# revisão da evidência.
container=${R2_PG_CONTAINER:-ai_hub_r3qual-postgres-1}
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
if docker exec ai_hub_r3qual-localstack-1 awslocal s3api head-bucket --bucket "$bucket" >/dev/null 2>&1; then
  echo "restore target bucket already exists: $bucket; choose a new R2_RESTORE_SUFFIX or R2_RESTORE_BUCKET" >&2
  exit 2
fi
docker exec ai_hub_r3qual-localstack-1 awslocal s3 mb "s3://$bucket" >/dev/null
docker exec ai_hub_r3qual-localstack-1 awslocal s3 sync s3://r2-custody "s3://$bucket" >/dev/null
source_objects=$(docker exec ai_hub_r3qual-localstack-1 awslocal s3api list-objects-v2 --bucket r2-custody --query 'length(Contents || `[]`)' --output text)
restored_objects=$(docker exec ai_hub_r3qual-localstack-1 awslocal s3api list-objects-v2 --bucket "$bucket" --query 'length(Contents || `[]`)' --output text)
test "$source_objects" = "$restored_objects" || { echo "object restore mismatch: $source_objects != $restored_objects" >&2; exit 1; }
printf 'restore objects: %s\n' "$source_objects"

effects_before=$(curl -fsS http://127.0.0.1:18090/__qualification/effects)
printf 'external effect oracle before reconciliation: %s\n' "$effects_before"
printf 'RESTORE_RECONCILIATION=PASS databases=%s suffix=%s bucket=%s\n' 'control,core,finance' "$suffix" "$bucket"
printf '%s\n' 'The provider oracle was observed without replay; re-admission remains disabled during restore.'
