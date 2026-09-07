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

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_control -f /migrations/control/0001_init.sql
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_control -f /migrations/control/0002_provider_auth.sql
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_control -f /migrations/control/0003_provider_api_catalog.sql
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_core -f /migrations/core/0001_init.sql
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -d hub_finance -f /migrations/finance/0001_init.sql
