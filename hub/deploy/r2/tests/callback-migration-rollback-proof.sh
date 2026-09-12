#!/usr/bin/env bash
set -euo pipefail

# Prova cercada de R4-CBK-01.6. O upgrade é aditivo e o rollback operacional é
# exercitado restaurando um snapshot pré-0044 em outro banco temporário. O
# volume oficial do Compose e as bases de negócio existentes não são tocados.
compose_project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
container=${R2_PG_CONTAINER:-${compose_project}-postgres-1}
suffix=${R2_CALLBACK_MIGRATION_SUFFIX:-$(date -u +%Y%m%d%H%M%S)}
case "$suffix" in *[!a-zA-Z0-9_]*) echo "invalid callback migration suffix" >&2; exit 2;; esac

upgrade_db="hub_core_callback_${suffix}"
rollback_db="hub_core_callback_rollback_${suffix}"
evidence=${R2_CALLBACK_MIGRATION_EVIDENCE:-hub/evidence/r2/execution/callback-migration-rollback-latest.log}
dump_file=$(mktemp)
psql_hub=(docker exec -i "$container" psql -U hub -v ON_ERROR_STOP=1 -X)
created_databases=()

cleanup() {
  local database
  rm -f "$dump_file"
  for database in "${created_databases[@]}"; do
    "${psql_hub[@]}" -d postgres -v database="$database" <<'SQL' >/dev/null 2>&1 || true
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE datname = :'database' AND pid <> pg_backend_pid();
SELECT format('DROP DATABASE IF EXISTS %I', :'database') \gexec
SQL
  done
}
trap cleanup EXIT

for database in "$upgrade_db" "$rollback_db"; do
  existing=$("${psql_hub[@]}" -d postgres -Atqc "SELECT 1 FROM pg_database WHERE datname='${database}'")
  if [[ "$existing" == "1" ]]; then
    echo "temporary database already exists: $database; choose another suffix" >&2
    exit 2
  fi
  "${psql_hub[@]}" -d postgres -v database="$database" <<'SQL'
SELECT format('CREATE DATABASE %I', :'database') \gexec
SQL
  created_databases+=("$database")
done

# Recria o schema anterior à migration de autenticação por conta. As demais
# migrations continuam sendo aplicadas na ordem original, inclusive o plano
# de produto com versão numérica repetida.
for file in hub/migrations/core/*.sql; do
  [[ "$(basename "$file")" == "0044_callback_account_auth.sql" ]] && continue
  "${psql_hub[@]}" -d "$upgrade_db" < "$file" >/dev/null
done

pre_upgrade_columns=$("${psql_hub[@]}" -d "$upgrade_db" -Atqc "SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='callback_inbox' AND column_name IN ('provider_account_id','callback_auth_version','callback_timestamp','callback_signature')")
[[ "$pre_upgrade_columns" == "0" ]] || {
  echo "schema pre-0044 já contém colunas novas: $pre_upgrade_columns" >&2
  exit 1
}

"${psql_hub[@]}" -d "$upgrade_db" <<'SQL'
INSERT INTO callback_inbox(inbox_id,operation_id,token_hash,body_sha256,body,disposition)
VALUES ('00000000-0000-4000-8000-000000000101',
        '00000000-0000-4000-8000-000000000102',
        'legacy-token-hash', 'legacy-body-hash', '{"status":"SUCCEEDED"}'::bytea,
        'RECEIVED');
SQL

# Snapshot de rollback antes do upgrade: os dados legados devem sobreviver à
# restauração e não podem carregar as colunas específicas da versão nova.
docker exec "$container" pg_dump -U hub --no-owner --no-privileges --format=plain "$upgrade_db" > "$dump_file"

"${psql_hub[@]}" -d "$upgrade_db" < hub/migrations/core/0044_callback_account_auth.sql >/dev/null
"${psql_hub[@]}" -d "$upgrade_db" < hub/migrations/core/0044_callback_account_auth.sql >/dev/null

post_upgrade_columns=$("${psql_hub[@]}" -d "$upgrade_db" -Atqc "SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='callback_inbox' AND column_name IN ('provider_account_id','callback_auth_version','callback_timestamp','callback_signature')")
legacy_after_upgrade=$("${psql_hub[@]}" -d "$upgrade_db" -Atqc "SELECT count(*) FROM callback_inbox WHERE inbox_id='00000000-0000-4000-8000-000000000101' AND token_hash='legacy-token-hash' AND disposition='RECEIVED'")
[[ "$post_upgrade_columns" == "4" && "$legacy_after_upgrade" == "1" ]] || {
  echo "upgrade preservou estado inesperado: columns=$post_upgrade_columns legacy=$legacy_after_upgrade" >&2
  exit 1
}

"${psql_hub[@]}" -d "$rollback_db" < "$dump_file" >/dev/null
rollback_columns=$("${psql_hub[@]}" -d "$rollback_db" -Atqc "SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='callback_inbox' AND column_name IN ('provider_account_id','callback_auth_version','callback_timestamp','callback_signature')")
legacy_after_rollback=$("${psql_hub[@]}" -d "$rollback_db" -Atqc "SELECT count(*) FROM callback_inbox WHERE inbox_id='00000000-0000-4000-8000-000000000101' AND token_hash='legacy-token-hash' AND disposition='RECEIVED'")
[[ "$rollback_columns" == "0" && "$legacy_after_rollback" == "1" ]] || {
  echo "rollback não restaurou compatibilidade legada: columns=$rollback_columns legacy=$legacy_after_rollback" >&2
  exit 1
}

mkdir -p "$(dirname "$evidence")"
{
  echo "AI Hub R2 — prova de upgrade/rollback da autenticação de callback por conta"
  echo "date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "environment=Compose ${compose_project} / PostgreSQL 16 / bancos temporários cercados"
  echo "migration=0044_callback_account_auth.sql"
  echo "pre_upgrade_new_columns=${pre_upgrade_columns}"
  echo "post_upgrade_new_columns=${post_upgrade_columns} legacy_row_preserved=${legacy_after_upgrade}"
  echo "reapply=idempotent"
  echo "rollback_new_columns=${rollback_columns} legacy_row_restored=${legacy_after_rollback}"
  echo "CALLBACK_MIGRATION_ROLLBACK_PROOF=PASS"
  echo "cleanup=temporary databases removed; official Compose volume preserved"
} > "$evidence"
cat "$evidence"
