#!/usr/bin/env bash
# R6-SEG-01: prova negativa e positiva de isolamento por tenant, com a
# transação COMMITADA antes de cada asserção (o script anterior fazia
# ROLLBACK das fixtures antes de contar linhas dentro da MESMA transação:
# a contagem impressa nunca provava nada sobre uma sessão futura). Este
# script também prova recusa de escrita cruzada e a policy audited_scope
# (acesso cruzado só com motivo não vazio, somente leitura).
set -euo pipefail
container=${R2_PG_CONTAINER:-ai_hub_r3qual-postgres-1}
run_db() {
  local db=$1
  shift
  docker exec -i "$container" psql "host=127.0.0.1 port=5432 dbname=$db user=hub_runtime password=r2-runtime-fixture" -v ON_ERROR_STOP=1 -X "$@"
}

# --- webhook_destinations (hub_core) ---
run_db hub_core <<'SQL'
BEGIN;
SELECT set_config('app.tenant_id','r2-rls-a',true);
INSERT INTO webhook_destinations(tenant_id,url,hmac_secret) VALUES ('r2-rls-a','https://r2-rls-a.invalid/hook','fixture-a') ON CONFLICT DO NOTHING;
COMMIT;
BEGIN;
SELECT set_config('app.tenant_id','r2-rls-b',true);
INSERT INTO webhook_destinations(tenant_id,url,hmac_secret) VALUES ('r2-rls-b','https://r2-rls-b.invalid/hook','fixture-b') ON CONFLICT DO NOTHING;
COMMIT;
SQL
own=$(run_db hub_core -Atqc "BEGIN; SELECT set_config('app.tenant_id','r2-rls-a',true); SELECT count(*) FROM webhook_destinations WHERE tenant_id='r2-rls-a'; COMMIT;" | tail -1)
test "$own" = 1
cross=$(run_db hub_core -Atqc "BEGIN; SELECT set_config('app.tenant_id','r2-rls-a',true); SELECT count(*) FROM webhook_destinations WHERE tenant_id='r2-rls-b'; COMMIT;" | tail -1)
test "$cross" = 0
write_rows=$(run_db hub_core -Atqc "BEGIN; SELECT set_config('app.tenant_id','r2-rls-a',true); UPDATE webhook_destinations SET url='https://hijack.invalid' WHERE tenant_id='r2-rls-b' RETURNING 1; COMMIT;" | grep -c '^1$' || true)
test "$write_rows" = 0
no_reason=$(run_db hub_core -Atqc "BEGIN; SELECT count(*) FROM webhook_destinations WHERE tenant_id IN ('r2-rls-a','r2-rls-b'); COMMIT;" | tail -1)
test "$no_reason" = 0
audited=$(run_db hub_core -Atqc "BEGIN; SELECT set_config('app.access_reason','r2-audit-proof',true); SELECT count(*) FROM webhook_destinations WHERE tenant_id IN ('r2-rls-a','r2-rls-b'); COMMIT;" | tail -1)
test "$audited" = 2
audited_write=$(run_db hub_core -Atqc "BEGIN; SELECT set_config('app.access_reason','r2-audit-proof',true); UPDATE webhook_destinations SET url='https://hijack2.invalid' WHERE tenant_id='r2-rls-b' RETURNING 1; COMMIT;" | grep -c '^1$' || true)
test "$audited_write" = 0
echo "hub_core (webhook_destinations): PASS"

# --- catalog_resources (hub_control) ---
run_db hub_control <<'SQL'
BEGIN;
SELECT set_config('app.tenant_id','r2-rls-a',true);
INSERT INTO catalog_resources(kind,id,version,tenant_id,name,data,content_hash,author)
VALUES ('clients','rls-a',1,'r2-rls-a','RLS A','{}','fixture-a','runtime-proof') ON CONFLICT DO NOTHING;
COMMIT;
BEGIN;
SELECT set_config('app.tenant_id','r2-rls-b',true);
INSERT INTO catalog_resources(kind,id,version,tenant_id,name,data,content_hash,author)
VALUES ('clients','rls-b',1,'r2-rls-b','RLS B','{}','fixture-b','runtime-proof') ON CONFLICT DO NOTHING;
COMMIT;
SQL
own=$(run_db hub_control -Atqc "BEGIN; SELECT set_config('app.tenant_id','r2-rls-a',true); SELECT count(*) FROM catalog_resources WHERE tenant_id='r2-rls-a'; COMMIT;" | tail -1)
test "$own" -ge 1
cross=$(run_db hub_control -Atqc "BEGIN; SELECT set_config('app.tenant_id','r2-rls-a',true); SELECT count(*) FROM catalog_resources WHERE tenant_id='r2-rls-b'; COMMIT;" | tail -1)
test "$cross" = 0
echo "hub_control (catalog_resources): PASS"

# --- credit_limits (hub_finance) ---
run_db hub_finance <<'SQL'
BEGIN;
SELECT set_config('app.tenant_id','r2-rls-a',true);
INSERT INTO credit_limits(tenant_id,limit_amount,currency) VALUES ('r2-rls-a',10,'BRL') ON CONFLICT DO NOTHING;
COMMIT;
BEGIN;
SELECT set_config('app.tenant_id','r2-rls-b',true);
INSERT INTO credit_limits(tenant_id,limit_amount,currency) VALUES ('r2-rls-b',20,'BRL') ON CONFLICT DO NOTHING;
COMMIT;
SQL
own=$(run_db hub_finance -Atqc "BEGIN; SELECT set_config('app.tenant_id','r2-rls-a',true); SELECT count(*) FROM credit_limits WHERE tenant_id='r2-rls-a'; COMMIT;" | tail -1)
test "$own" -ge 1
cross=$(run_db hub_finance -Atqc "BEGIN; SELECT set_config('app.tenant_id','r2-rls-a',true); SELECT count(*) FROM credit_limits WHERE tenant_id='r2-rls-b'; COMMIT;" | tail -1)
test "$cross" = 0
echo "hub_finance (credit_limits): PASS"

for db in hub_core hub_control hub_finance; do
  role_flags=$(docker exec -i "$container" psql "dbname=$db user=hub" -Atqc "SELECT rolsuper::text||':'||rolcreaterole::text||':'||rolbypassrls::text FROM pg_roles WHERE rolname='hub_runtime'")
  test "$role_flags" = 'false:false:false'
done
echo 'RLS runtime proof (control/core/finance, non-owner role, committed reads, cross-tenant write denial, audited-scope): PASS'
