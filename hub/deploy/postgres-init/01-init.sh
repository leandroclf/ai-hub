#!/bin/bash
# Cria as tres bases logicas do hub (DAD-01: hub_control, hub_core,
# hub_finance) e aplica as migrations iniciais de cada uma. Em
# local/dev estas bases compartilham o mesmo servidor (permitido por
# DAD-01); em ppd/prd seriam instancias/clusters separados por celula.
set -e

for db in hub_control hub_core hub_finance; do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d postgres <<-EOSQL
    CREATE DATABASE ${db};
EOSQL
done

for file in /migrations/control/*.sql; do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_control -f "$file"
done
for file in /migrations/core/*.sql; do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_core -f "$file"
done
for file in /migrations/finance/*.sql; do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_finance -f "$file"
done
