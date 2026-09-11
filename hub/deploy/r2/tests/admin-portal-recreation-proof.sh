#!/usr/bin/env bash
set -euo pipefail

# R2-ADM-12/4.1: a fatia do portal não possui migration de dados própria.
# Esta prova cerca o rollout da SPA com o estado de catálogo que ela consulta:
# a imagem é recriada, as rotas continuam renderizando, a API segue protegida
# sem sessão e as contagens dos quatro cadastros não mudam.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
PROJECT="${R2_COMPOSE_PROJECT:-ai_hub_r3qual}"
COMPOSE=(docker compose -p "$PROJECT" -f "$ROOT_DIR/hub/deploy/r2/compose.yaml")
POSTGRES="${PROJECT}-postgres-1"
EVIDENCE_FILE="${R2_ADMIN_PORTAL_RECREATION_EVIDENCE_FILE:-$ROOT_DIR/hub/evidence/r2/execution/admin-portal-recreation-latest.log}"
mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

catalog_snapshot() {
  docker exec "$POSTGRES" psql -U hub -d hub_control -At -F '|' -c \
    "SELECT kind, count(*) FROM catalog_resources WHERE kind IN ('applications','clients','products','services') GROUP BY kind ORDER BY kind;"
}

assert_route() {
  local route="$1"
  local status
  status="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:13000/${route}")"
  test "$status" = 200
  status="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:13000/api/atlas/admin/v1/${route}?tenant_id=acme")"
  test "$status" = 401
  echo "route=${route} page_http=200 api_without_session=401"
}

echo "ADMIN_PORTAL_RECREATION_PROOF=START"
before_container="$(docker inspect "${PROJECT}-admin-ui-1" --format '{{.Id}}')"
before_catalog="$(catalog_snapshot)"
echo "before_container=${before_container}"
echo "before_catalog=$(tr '\n' ';' <<<"$before_catalog")"
for route in clients applications services products; do assert_route "$route"; done

"${COMPOSE[@]}" up -d --no-deps --force-recreate admin-ui >/dev/null

ready=0
for _ in $(seq 1 45); do
  if curl -fsS --max-time 3 http://127.0.0.1:13000/services >/dev/null; then
    ready=1
    break
  fi
  sleep 1
done
test "$ready" = 1

after_container="$(docker inspect "${PROJECT}-admin-ui-1" --format '{{.Id}}')"
after_catalog="$(catalog_snapshot)"
echo "after_container=${after_container}"
echo "after_catalog=$(tr '\n' ';' <<<"$after_catalog")"
for route in clients applications services products; do assert_route "$route"; done

if [[ "$before_container" == "$after_container" || "$before_catalog" != "$after_catalog" ]]; then
  echo "ADMIN_PORTAL_RECREATION_PROOF=FAIL state_or_container_not_preserved"
  exit 1
fi

echo "ADMIN_PORTAL_RECREATION_PROOF=PASS container_changed=true catalog_preserved=true routes_protected=true data_migration=not_applicable"
