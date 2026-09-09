#!/usr/bin/env bash
set -euo pipefail
container=${R2_PG_CONTAINER:-ai_hub_r3qual-postgres-1}
run_db() {
  local db=$1
  shift
  docker exec -i "$container" psql "host=127.0.0.1 port=5432 dbname=$db user=hub_runtime password=r2-runtime-fixture" -v ON_ERROR_STOP=1 -X "$@"
}

run_db hub_core <<'SQL'
BEGIN;
SET LOCAL app.tenant_id = 'r2-rls-a';
INSERT INTO webhook_destinations(tenant_id,url,hmac_secret) VALUES ('r2-rls-a','https://r2-rls-a.invalid/hook','fixture-a');
SET LOCAL app.tenant_id = 'r2-rls-b';
INSERT INTO webhook_destinations(tenant_id,url,hmac_secret) VALUES ('r2-rls-b','https://r2-rls-b.invalid/hook','fixture-b');
SELECT count(*) AS tenant_b_visible FROM webhook_destinations WHERE tenant_id IN ('r2-rls-a','r2-rls-b');
ROLLBACK;
SQL
visible=$(run_db hub_core -Atqc "SET app.tenant_id='r2-rls-a'; SELECT count(*) FROM webhook_destinations WHERE tenant_id='r2-rls-b';")
test "$visible" = 0

run_db hub_control <<'SQL'
BEGIN;
SET LOCAL app.tenant_id = 'r2-rls-a';
INSERT INTO catalog_resources(kind,id,version,tenant_id,name,data,content_hash,author)
VALUES ('clients','rls-a',1,'r2-rls-a','RLS A','{}','fixture-a','runtime-proof');
SET LOCAL app.tenant_id = 'r2-rls-b';
INSERT INTO catalog_resources(kind,id,version,tenant_id,name,data,content_hash,author)
VALUES ('clients','rls-b',1,'r2-rls-b','RLS B','{}','fixture-b','runtime-proof');
SELECT count(*) AS tenant_b_visible FROM catalog_resources WHERE tenant_id IN ('r2-rls-a','r2-rls-b');
ROLLBACK;
SQL
visible=$(run_db hub_control -Atqc "SET app.tenant_id='r2-rls-a'; SELECT count(*) FROM catalog_resources WHERE tenant_id='r2-rls-b';")
test "$visible" = 0

run_db hub_finance <<'SQL'
BEGIN;
SET LOCAL app.tenant_id = 'r2-rls-a';
INSERT INTO credit_limits(tenant_id,limit_amount,currency) VALUES ('r2-rls-a',10,'BRL');
SET LOCAL app.tenant_id = 'r2-rls-b';
INSERT INTO credit_limits(tenant_id,limit_amount,currency) VALUES ('r2-rls-b',20,'BRL');
SELECT count(*) AS tenant_b_visible FROM credit_limits WHERE tenant_id IN ('r2-rls-a','r2-rls-b');
ROLLBACK;
SQL
visible=$(run_db hub_finance -Atqc "SET app.tenant_id='r2-rls-a'; SELECT count(*) FROM credit_limits WHERE tenant_id='r2-rls-b';")
test "$visible" = 0

role_flags=$(docker exec -i "$container" psql "dbname=hub_core user=hub" -Atqc "SELECT rolsuper::text||':'||rolcreaterole::text||':'||rolbypassrls::text FROM pg_roles WHERE rolname='hub_runtime'")
test "$role_flags" = 'false:false:false'
echo 'RLS runtime negative cross-tenant proof (control/core/finance, non-owner role): PASS'
