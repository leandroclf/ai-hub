#!/usr/bin/env bash
set -euo pipefail

# Prova estrutural de overlays e do gate de promoção. Não tenta simular
# credenciais nem declara isolamento de runtime remoto sem um cluster/IdP
# correspondente.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
KUBECTL_BIN="${KUBECTL_BIN:-kubectl}"
GATE="${ROOT_DIR}/hub/deploy/r2/tests/promotion-gate.sh"
EVIDENCE_FILE="${R2_ENVIRONMENT_BOUNDARY_EVIDENCE_FILE:-${ROOT_DIR}/hub/evidence/r2/execution/environment-boundary-latest.log}"
TMP_DIR="$(mktemp -d /tmp/ai-hub-environment-boundary.XXXXXX)"
trap 'find "$TMP_DIR" -depth -delete' EXIT

mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

overlays=(local-kind dev hom ppd prd)
declare -A namespaces
echo "ENVIRONMENT_BOUNDARY_PROOF=START"
kubectl_version="$($KUBECTL_BIN version --client -o json | tr '\n' ' ')"
echo "kubectl=${kubectl_version% }"

for environment in "${overlays[@]}"; do
  output_file="${TMP_DIR}/${environment}.yaml"
  overlay_dir="${ROOT_DIR}/hub/deploy/r2/k8s/overlays/${environment}"
  "$KUBECTL_BIN" kustomize "$overlay_dir" >"$output_file"
  namespace_lines="$(rg '^  namespace:' "$output_file" | sed 's/.*namespace: //' | sort -u)"
  expected_namespace="ai-hub-${environment}"
  [[ "$namespace_lines" == "$expected_namespace" ]] || {
    echo "overlay=${environment} FAIL namespaces=$(tr '\n' ',' <<<"$namespace_lines") expected=${expected_namespace}"
    exit 1
  }
  namespaces["$expected_namespace"]="$environment"
  if rg -n '^(data|stringData):|secretString:|client-secret:|password:' "$output_file"; then
    echo "overlay=${environment} FAIL materialized_secret_value"
    exit 1
  fi
  for service in atlas cometa libra orbita pulsar; do
    rg -q "secretName: ${service}-workload" "$output_file" || {
      echo "overlay=${environment} FAIL missing_workload_secret=${service}-workload"
      exit 1
    }
  done
  rg -q 'name: hub-runtime' "$output_file" || {
    echo "overlay=${environment} FAIL missing_runtime_secret_reference"
    exit 1
  }
  digest="$(sha256sum "$output_file" | awk '{print $1}')"
  echo "overlay=${environment} PASS namespace=${expected_namespace} digest=${digest}"
done

[[ "${#namespaces[@]}" == "5" ]] || { echo "FAIL namespaces_not_unique"; exit 1; }

if R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE= R2_PROMOTION_APPROVALS= R2_ENVIRONMENT_ISOLATION_PROOF= "$GATE"; then
  echo "promotion_prd_missing_profile FAIL gate_did_not_block"
  exit 1
else
  echo "promotion_prd_missing_profile PASS blocked_as_expected"
fi

dev_gate="$(R2_PROMOTION_ENV=dev "$GATE")"
echo "promotion_dev PASS ${dev_gate}"

qualified_gate="$(R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE=synthetic-qualification-fixture R2_PROMOTION_APPROVALS='P-01,P-08,P-10' R2_ENVIRONMENT_ISOLATION_PROOF=PASS "$GATE")"
echo "promotion_prd_qualified_fixture PASS ${qualified_gate}"

if R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE=regional-rpo-zero R2_PROMOTION_APPROVALS='P-01,P-08,P-10' R2_ENVIRONMENT_ISOLATION_PROOF=PASS R2_REGIONAL_PROFILE=RPO_ZERO R2_REGIONAL_CUSTODY_CONFIRMATION=PASS "$GATE"; then
  echo "promotion_prd_regional_confirmation_only FAIL gate_did_not_block"
  exit 1
else
  echo "promotion_prd_regional_confirmation_only PASS blocked_as_expected"
fi

regional_gate="$(R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE=regional-rpo-zero R2_PROMOTION_APPROVALS='P-01,P-08,P-10' R2_ENVIRONMENT_ISOLATION_PROOF=PASS R2_REGIONAL_PROFILE=RPO_ZERO R2_REGIONAL_CUSTODY_CONFIRMATION=PASS R2_REGIONAL_CUSTODY_PROOF=PASS R2_REGIONAL_FENCING_PROOF=PASS "$GATE")"
echo "promotion_prd_regional_proofs PASS ${regional_gate}"
echo "ENVIRONMENT_BOUNDARY_PROOF=PASS overlays=${#overlays[@]} unique_namespaces=${#namespaces[@]} production_gate=blocked_without_profile"
echo "evidence=${EVIDENCE_FILE#${ROOT_DIR}/}"
