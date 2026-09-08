#!/bin/sh
set -eu
# Cada migration e seu checksum são confirmados na mesma transação.
# Advisory lock serializa migradores; adulteração de migration aplicada falha.
suffix=${MIGRATION_DB_SUFFIX:-}
case "$suffix" in *[!a-zA-Z0-9_]*) echo "invalid database suffix" >&2; exit 2;; esac
for domain in control core finance; do
  db="hub_${domain}${suffix}"
  psql -v ON_ERROR_STOP=1 -d "$db" -c 'CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT clock_timestamp())'
  for file in /migrations/"$domain"/*.sql; do
    version=$(basename "$file")
    checksum=$(sha256sum "$file" | cut -d ' ' -f 1)
    psql -v ON_ERROR_STOP=1 -d "$db" <<SQL
BEGIN;
SELECT pg_advisory_xact_lock(72820309);
SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = '$version') AS applied \gset
\if :applied
  SELECT checksum = '$checksum' AS intact FROM schema_migrations WHERE version = '$version' \gset
  \if :intact
    \echo migration $domain/$version already applied
  \else
    \echo checksum mismatch $domain/$version
    \quit 3
  \endif
\else
  \i $file
  INSERT INTO schema_migrations(version,checksum) VALUES ('$version','$checksum');
\endif
COMMIT;
SQL
  done
done
