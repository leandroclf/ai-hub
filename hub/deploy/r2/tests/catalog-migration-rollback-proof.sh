#!/usr/bin/env bash
set -euo pipefail

# Prova cercada de R2-CAT-04.1. Os bancos temporários são criados pelo runner,
# usados para a migração aditiva do catálogo e removidos no cleanup; o volume
# oficial e as obrigações do Compose não são tocados.
compose_project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
compose_file=${R2_COMPOSE_FILE:-hub/deploy/r2/compose.yaml}
container=${R2_PG_CONTAINER:-${compose_project}-postgres-1}
suffix=${R2_CATALOG_MIGRATION_SUFFIX:-$(date -u +%Y%m%d%H%M%S)}
case "$suffix" in *[!a-zA-Z0-9_]*) echo "invalid catalog migration suffix" >&2; exit 2;; esac

control_db="hub_control_catalog_${suffix}"
core_db="hub_core_catalog_${suffix}"
finance_db="hub_finance_catalog_${suffix}"
evidence=${R2_CATALOG_MIGRATION_EVIDENCE:-hub/evidence/r2/execution/catalog-migration-rollback-latest.log}
psql_hub=(docker exec -i "$container" psql -U hub -v ON_ERROR_STOP=1 -X)

created_databases=()
cleanup() {
  local database
  for database in "${created_databases[@]}"; do
    "${psql_hub[@]}" -d postgres -v database="$database" <<'SQL' >/dev/null 2>&1 || true
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE datname = :'database' AND pid <> pg_backend_pid();
SELECT format('DROP DATABASE IF EXISTS %I', :'database') \gexec
SQL
  done
}
trap cleanup EXIT

for database in "$control_db" "$core_db" "$finance_db"; do
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

run_migrations() {
  COMPOSE_PROJECT_NAME="$compose_project" docker compose -f "$compose_file" run --rm \
    -e MIGRATION_DB_SUFFIX="_catalog_${suffix}" migrate
}

run_migrations >/dev/null

# A variante histórica é inserida somente depois da instalação limpa. A
# migration 0020 deve convertê-la em inventário DRAFT sem alterar sua origem,
# identidade ou a tabela legada consultada por consumidores antigos.
"${psql_hub[@]}" -d "$control_db" <<'SQL'
INSERT INTO services(code,version,description,schema_input,schema_output,modes,client_sla_seconds,retry_ttl_seconds,published)
VALUES
  ('legacy-migration-alpha',1,'Serviço legado alpha','{"cpf":{"type":"string"}}','{"status":{"type":"string"}}',ARRAY['SYNC'],30,5,true),
  ('legacy-migration-beta',2,'Serviço legado beta','{"documento":{"type":"string"}}','{"status":{"type":"string"}}',ARRAY['ASYNC'],45,10,false);
SQL
docker exec -i "$container" psql -U hub -d "$control_db" -v ON_ERROR_STOP=1 -X < hub/migrations/control/0020_catalog_versions.sql >/dev/null

legacy_count=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT count(*) FROM services WHERE code LIKE 'legacy-migration-%'")
catalog_count=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT count(*) FROM catalog_resources WHERE kind='services' AND id LIKE 'legacy-migration-%' AND state='DRAFT' AND data->>'provenance'='legacy-service'")
[[ "$legacy_count" == "2" && "$catalog_count" == "2" ]] || {
  echo "legacy backfill mismatch: services=$legacy_count catalog_drafts=$catalog_count" >&2
  exit 1
}

legacy_ids=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT string_agg(code || ':' || version, ',' ORDER BY code,version) FROM services WHERE code LIKE 'legacy-migration-%'")
catalog_ids=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT string_agg(id || ':' || version, ',' ORDER BY id,version) FROM catalog_resources WHERE kind='services' AND id LIKE 'legacy-migration-%'")
[[ "$legacy_ids" == "$catalog_ids" ]] || {
  echo "legacy identity mismatch: services=$legacy_ids catalog=$catalog_ids" >&2
  exit 1
}

# Simula um protocolo aceito com snapshot v1 antes do recuo. O recuo suspende
# somente v2 e preserva a versão anterior e a obrigação em voo; não existe
# DELETE nem reescrita de conteúdo publicado.
"${psql_hub[@]}" -d "$control_db" <<'SQL'
INSERT INTO catalog_resources(kind,id,version,tenant_id,name,state,data,content_hash,author)
VALUES
 ('services','migration-rollback-service',1,'acme','Serviço v1','PUBLISHED','{"schema":"v1","provenance":"test"}','hash-v1','migration-proof'),
 ('services','migration-rollback-service',2,'acme','Serviço v2','PUBLISHED','{"schema":"v2","provenance":"test"}','hash-v2','migration-proof');
INSERT INTO catalog_publications(kind,resource_id,resource_version,revision,content_hash,actor,reason,validation)
VALUES
 ('services','migration-rollback-service',1,1,'hash-v1','migration-proof','publicação v1','{}'),
 ('services','migration-rollback-service',2,1,'hash-v2','migration-proof','publicação v2','{}');
UPDATE catalog_resources SET state='SUSPENDED',revision=revision+1,updated_at=clock_timestamp()
WHERE kind='services' AND id='migration-rollback-service' AND version=2 AND state='PUBLISHED';
INSERT INTO catalog_publications(kind,resource_id,resource_version,revision,content_hash,actor,reason,validation)
SELECT kind,id,version,revision,content_hash,'migration-proof','recuo controlado para v1','{"action":"rollback-pointer"}'
FROM catalog_resources
WHERE kind='services' AND id='migration-rollback-service' AND version=2;
SQL

"${psql_hub[@]}" -d "$core_db" <<'SQL'
INSERT INTO protocols(protocol_id,tenant_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,result_version,client_deadline_at)
VALUES ('00000000-0000-4000-8000-000000000001','acme','catalog-migration-in-flight','migration-proof','{"service":"migration-rollback-service","service_version":1}','ASYNC','QUEUED','00000000-0000-4000-8000-000000000002','ACCEPTED',0,clock_timestamp()+interval '30 minutes');
SQL

v1_state=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT state FROM catalog_resources WHERE kind='services' AND id='migration-rollback-service' AND version=1")
v2_state=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT state FROM catalog_resources WHERE kind='services' AND id='migration-rollback-service' AND version=2")
publication_count=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT count(*) FROM catalog_publications WHERE resource_id='migration-rollback-service'")
in_flight=$("${psql_hub[@]}" -d "$core_db" -Atqc "SELECT count(*) FROM protocols WHERE idempotency_key='catalog-migration-in-flight' AND status='ACCEPTED' AND request_body->>'service_version'='1'")
[[ "$v1_state" == "PUBLISHED" && "$v2_state" == "SUSPENDED" && "$publication_count" == "3" && "$in_flight" == "1" ]] || {
  echo "rollback preservation mismatch: v1=$v1_state v2=$v2_state publications=$publication_count in_flight=$in_flight" >&2
  exit 1
}

# O replay do migrador deve ser idempotente e manter os dados de catálogo e a
# obrigação em voo; a saída não é usada como oráculo único, as consultas acima
# continuam sendo o critério de aprovação.
run_migrations >/dev/null
replay_catalog_count=$("${psql_hub[@]}" -d "$control_db" -Atqc "SELECT count(*) FROM catalog_resources WHERE kind='services' AND id LIKE 'legacy-migration-%' AND state='DRAFT'")
replay_in_flight=$("${psql_hub[@]}" -d "$core_db" -Atqc "SELECT count(*) FROM protocols WHERE idempotency_key='catalog-migration-in-flight' AND status='ACCEPTED'")
[[ "$replay_catalog_count" == "2" && "$replay_in_flight" == "1" ]] || {
  echo "migration replay changed preserved state: catalog=$replay_catalog_count in_flight=$replay_in_flight" >&2
  exit 1
}

mkdir -p "$(dirname "$evidence")"
{
  echo "AI Hub R2 — prova de migração e recuo do catálogo"
  echo "date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "environment=Compose ${compose_project} / PostgreSQL 16 / bancos temporários cercados"
  echo "legacy_services=${legacy_count} catalog_drafts=${catalog_count} identities=${legacy_ids}"
  echo "rollback=v1:${v1_state} v2:${v2_state} publication_history=${publication_count}"
  echo "in_flight_protocol_snapshot_v1=${in_flight}"
  echo "replay_catalog_drafts=${replay_catalog_count} replay_in_flight=${replay_in_flight}"
  echo "CATALOG_MIGRATION_ROLLBACK_PROOF=PASS"
  echo "cleanup=temporary databases removed; official Compose volume preserved"
} > "$evidence"
cat "$evidence"
