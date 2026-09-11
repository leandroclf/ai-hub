#!/usr/bin/env bash
set -euo pipefail

# Valida a política declarada de escala e placement. É uma prova estrutural:
# backlog real, provisionamento cloud e expansão de célula continuam exigindo
# um ambiente representativo próprio.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
KUBECTL_BIN="${KUBECTL_BIN:-kubectl}"
EVIDENCE_FILE="${R2_SCALING_POLICY_EVIDENCE_FILE:-${ROOT_DIR}/hub/evidence/r2/execution/scaling-policy-latest.log}"
TMP_DIR="$(mktemp -d /tmp/ai-hub-scaling-policy.XXXXXX)"
trap 'find "$TMP_DIR" -depth -delete' EXIT
mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

render(){ "$KUBECTL_BIN" kustomize "$ROOT_DIR/hub/deploy/r2/k8s/overlays/$1"; }
ppd_file="$TMP_DIR/ppd.yaml"
local_file="$TMP_DIR/local-kind.yaml"
render ppd >"$ppd_file"
render local-kind >"$local_file"

count(){ rg -c -- "$1" "$2" || true; }
fail(){ echo "SCALING_POLICY_PROOF=FAIL reason=$*"; exit 1; }

echo "SCALING_POLICY_PROOF=START"
echo "kubectl=$($KUBECTL_BIN version --client -o json | tr '\n' ' ' | sed 's/ $//')"
[[ "$(count '^kind: HorizontalPodAutoscaler$' "$ppd_file")" == 4 ]] || fail ppd_hpa_count
[[ "$(count '^kind: ScaledObject$' "$ppd_file")" == 1 ]] || fail ppd_scaledobject_count
[[ "$(count '^kind: PodDisruptionBudget$' "$ppd_file")" == 5 ]] || fail ppd_pdb_count
[[ "$(count 'minReplicas: 3' "$ppd_file")" == 4 ]] || fail ppd_hpa_floor
[[ "$(count 'maxReplicas: 6' "$ppd_file")" == 4 ]] || fail ppd_hpa_ceiling
[[ "$(count 'minReplicaCount: 3' "$ppd_file")" == 1 ]] || fail ppd_keda_floor
[[ "$(count 'maxReplicaCount: 6' "$ppd_file")" == 1 ]] || fail ppd_keda_ceiling
rg -q "query: SELECT count\(\*\) FROM deliveries WHERE state IN \('PENDING','RETRY_SCHEDULED'\)" "$ppd_file" || fail backlog_query_missing
[[ "$(count 'topologyKey: kubernetes.io/hostname' "$ppd_file")" == 5 ]] || fail topology_spread_missing
[[ "$(count 'maxUnavailable: 1' "$ppd_file")" == 5 ]] || fail pdb_budget_missing
[[ "$(count 'minReplicas: 2' "$local_file")" == 4 ]] || fail local_kind_floor
[[ "$(count 'minReplicaCount: 2' "$local_file")" == 1 ]] || fail local_kind_keda_floor
ppd_digest="$(sha256sum "$ppd_file" | awk '{print $1}')"
local_digest="$(sha256sum "$local_file" | awk '{print $1}')"
echo "ppd PASS hpa=4 keda=1 pdb=5 floor=3 ceiling=6 digest=${ppd_digest}"
echo "local-kind PASS hpa=4 keda=1 floor=2 digest=${local_digest}"
echo "SCALING_POLICY_PROOF=PASS backlog_signal=postgres_pending_deliveries placement=topology_spread pbd=maxUnavailable-1"
