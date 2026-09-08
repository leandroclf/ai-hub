#!/bin/sh
set -eu
# Fixture isolada. Runtime usa roles por domínio; migração usa administrador local.
for db in hub_control hub_core hub_finance hub_identity; do
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE $db"
done
