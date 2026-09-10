#!/bin/sh
set -eu
# Cada migration e seu checksum são confirmados na mesma transação.
# Advisory lock serializa migradores; adulteração de migration aplicada falha.
suffix=${MIGRATION_DB_SUFFIX:-}
case "$suffix" in *[!a-zA-Z0-9_]*) echo "invalid database suffix" >&2; exit 2;; esac
# 0002 existiu em uma variante R3 conhecida que já continha API_KEY. Essa é
# a única reconciliação automática permitida: o schema é verificado antes de
# alinhar o ledger, e a decisão fica auditada. Hashes desconhecidos continuam
# sendo erro fatal.
provider_auth_legacy_sha256=37241e604376471efd7de394e9394673feaec0a32476a517c38363890dad1541
provider_auth_known_variant_sha256=f4174fbf3e0157b42c8d8fbed0d92405337a9a3bd00647af2f9a52f4c4390fbb
for domain in control core finance; do
  db="hub_${domain}${suffix}"
  psql -v ON_ERROR_STOP=1 -d "$db" -c 'CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT clock_timestamp())'
  for file in /migrations/"$domain"/*.sql; do
    version=$(basename "$file")
    checksum=$(sha256sum "$file" | cut -d ' ' -f 1)
    provider_auth_migration=false
    if [ "$domain/$version" = "control/0002_provider_auth.sql" ]; then provider_auth_migration=true; fi
    psql -v ON_ERROR_STOP=1 -v provider_auth_migration="$provider_auth_migration" -d "$db" <<SQL
BEGIN;
SELECT pg_advisory_xact_lock(72820309);
SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = '$version') AS applied \gset
\if :applied
  SELECT checksum = '$checksum' AS intact FROM schema_migrations WHERE version = '$version' \gset
  \if :intact
    \echo migration $domain/$version already applied
  \else
    \if :provider_auth_migration
      SELECT checksum = '$provider_auth_known_variant_sha256' AS known_variant FROM schema_migrations WHERE version = '$version' \gset
      \if :known_variant
        SELECT EXISTS (
          SELECT 1 FROM information_schema.columns
          WHERE table_name='provider_accounts' AND column_name='api_key_header'
        ) AND EXISTS (
          SELECT 1 FROM pg_constraint c JOIN pg_class t ON t.oid=c.conrelid
          WHERE t.relname='provider_accounts' AND pg_get_constraintdef(c.oid) LIKE '%API_KEY%'
        ) AS known_schema \gset
        \if :known_schema
          CREATE TABLE IF NOT EXISTS schema_migration_reconciliations (
            version text PRIMARY KEY, previous_checksum text NOT NULL,
            reconciled_checksum text NOT NULL, reason text NOT NULL,
            reconciled_at timestamptz NOT NULL DEFAULT clock_timestamp()
          );
          INSERT INTO schema_migration_reconciliations(version,previous_checksum,reconciled_checksum,reason)
          VALUES ('$version','$provider_auth_known_variant_sha256','$provider_auth_legacy_sha256','known R3 provider-auth variant verified')
          ON CONFLICT (version) DO NOTHING;
          UPDATE schema_migrations SET checksum='$provider_auth_legacy_sha256' WHERE version='$version';
          \echo reconciled known provider-auth variant $domain/$version
        \else
          DO \$\$ BEGIN RAISE EXCEPTION 'known provider-auth checksum without expected schema'; END \$\$;
        \endif
      \else
        \echo checksum mismatch $domain/$version
        DO \$\$ BEGIN
          RAISE EXCEPTION 'checksum mismatch %/%', '$domain', '$version';
        END \$\$;
      \endif
    \else
    \echo checksum mismatch $domain/$version
    DO \$\$ BEGIN
      RAISE EXCEPTION 'checksum mismatch %/%', '$domain', '$version';
    END \$\$;
    \endif
  \endif
\else
  \i $file
  INSERT INTO schema_migrations(version,checksum) VALUES ('$version','$checksum');
\endif
COMMIT;
SQL
  done
done
