#!/usr/bin/env bash
set -euo pipefail

# R2-OPE-03-S01: falha somente a telemetria opcional e verifica que os
# workloads elegíveis permanecem prontos. O runner toca apenas o Compose
# oficial e restaura Alloy no caminho de saída.
project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
compose_file=${R2_COMPOSE_FILE:-hub/deploy/r2/compose.yaml}
postgres="${project}-postgres-1"
evidence_file=${R2_OPTIONAL_DEPENDENCY_EVIDENCE_FILE:-hub/evidence/r2/execution/optional-dependency-runtime-latest.log}
mkdir -p "$(dirname "$evidence_file")"
exec > >(tee "$evidence_file")

fixture_ip(){
  docker inspect "${project}-$1-1" --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'
}

compose(){
  R2_PROVIDER_CIDR="$(fixture_ip provider-sim)/32" \
  R2_SINK_CIDR="$(fixture_ip webhook-sink)/32" \
  R2_IDENTITY_CIDR="$(fixture_ip identity)/32" \
  docker compose --project-name "$project" --file "$compose_file" "$@"
}

container_id(){ docker inspect "${project}-$1-1" --format '{{.Id}}'; }
restart_count(){ docker inspect "${project}-$1-1" --format '{{.RestartCount}}'; }

baseline_services=(atlas pulsar libra provider-sim webhook-sink)
declare -A baseline_ids baseline_restarts
for service in "${baseline_services[@]}"; do
  baseline_ids[$service]="$(container_id "$service")"
  baseline_restarts[$service]="$(restart_count "$service")"
done

cleanup(){
  set +e
  compose up -d alloy >/dev/null 2>&1
  compose up -d --no-deps orbita cometa >/dev/null 2>&1
}
trap cleanup EXIT

echo "OPTIONAL_DEPENDENCY_PROOF=START"
compose stop alloy
alloy_status="$(docker inspect "${project}-alloy-1" --format '{{.State.Status}}')"
alloy_probe="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 2 http://127.0.0.1:12345/ 2>/dev/null || true)"
compose up -d --force-recreate --no-deps orbita cometa

ready=0
live=0
deadline=$((SECONDS+45))
while ((SECONDS<deadline)); do
  if curl -fsS http://127.0.0.1:18080/healthz/ready >/dev/null 2>&1 && curl -fsS http://127.0.0.1:18082/healthz/ready >/dev/null 2>&1; then ready=1; break; fi
  sleep 1
done
if curl -fsS http://127.0.0.1:18080/healthz/live >/dev/null 2>&1 && curl -fsS http://127.0.0.1:18082/healthz/live >/dev/null 2>&1; then live=1; fi

unchanged=1
for service in "${baseline_services[@]}"; do
  current_id="$(container_id "$service")"
  current_restart="$(restart_count "$service")"
  echo "service=${service} id_unchanged=$([[ "$current_id" == "${baseline_ids[$service]}" ]] && echo true || echo false) restart_count_before=${baseline_restarts[$service]} restart_count_after=${current_restart}"
  [[ "$current_id" == "${baseline_ids[$service]}" && "$current_restart" == "${baseline_restarts[$service]}" ]] || unchanged=0
done

echo "alloy_status=${alloy_status} alloy_probe_http=${alloy_probe} ready=${ready} live=${live} unaffected_workloads=${unchanged}"
if [[ "$alloy_status" == exited && "$ready" == 1 && "$live" == 1 && "$unchanged" == 1 ]]; then
  echo "OPTIONAL_DEPENDENCY_PROOF=PASS dependency=telemetry workload_restart_cascade=false"
else
  echo "OPTIONAL_DEPENDENCY_PROOF=FAIL dependency=telemetry"
  exit 1
fi
