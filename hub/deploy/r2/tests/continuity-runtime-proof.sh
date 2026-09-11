#!/usr/bin/env bash
set -euo pipefail

# Qualifica o envelope local do laboratório Kind sem confundir a prova com HA
# físico ou com um cluster independente. A exclusão controlada de um pod é
# intencional e limitada aos deployments do namespace indicado.
kubeconfig=${R2_KUBECONFIG:-/tmp/ai-hub-r2-tools/kubeconfig}
namespace=${R2_KIND_NAMESPACE:-ai-hub-local-kind}
kubectl_bin=${KUBECTL_BIN:-kubectl}
delete_pods=${R2_CONTINUITY_DELETE_PODS:-1}
require_owned_dependencies=${R2_REQUIRE_KIND_OWNED_DEPS:-0}

kubectl_cmd=("$kubectl_bin" --kubeconfig "$kubeconfig")
fail() {
  echo "CONTINUITY_RUNTIME_PROOF=FAIL reason=$*" >&2
  exit 1
}

wait_available() {
  local workload=$1
  local deadline=$(( $(date +%s) + 120 ))
  while (( $(date +%s) < deadline )); do
    local available desired
    available=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true)
    desired=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.spec.replicas}' 2>/dev/null || true)
    if [[ "${available:-0}" == "$desired" && "${available:-0}" -ge 2 ]]; then
      return 0
    fi
    sleep 2
  done
  fail "$workload não atingiu réplicas disponíveis: disponível=${available:-0} desejado=${desired:-0}"
}

command -v "$kubectl_bin" >/dev/null || fail "kubectl ausente"
[[ -r "$kubeconfig" ]] || fail "kubeconfig ausente: $kubeconfig"
"${kubectl_cmd[@]}" cluster-info >/dev/null || fail "cluster indisponível"

ready_nodes=$("${kubectl_cmd[@]}" get nodes --no-headers | awk '$2 == "Ready" {count++} END {print count+0}')
[[ "$ready_nodes" -ge 3 ]] || fail "nós Ready insuficientes: $ready_nodes"

workloads=(atlas orbita cometa pulsar libra)
for workload in "${workloads[@]}"; do
  available=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.status.availableReplicas}')
  desired=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.spec.replicas}')
  [[ "${available:-0}" == "$desired" && "${available:-0}" -ge 2 ]] || fail "$workload disponível=${available:-0} desejado=${desired:-0}"
  nodes=$("${kubectl_cmd[@]}" -n "$namespace" get pods -l "app.kubernetes.io/name=$workload" -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' | sort -u | awk 'NF {count++} END {print count+0}')
  [[ "$nodes" -ge 2 ]] || fail "$workload não distribuído em dois nós: $nodes"
  "${kubectl_cmd[@]}" -n "$namespace" get pdb "$workload" >/dev/null || fail "PDB ausente: $workload"
done

for scaler in atlas orbita cometa libra; do
  min=$("${kubectl_cmd[@]}" -n "$namespace" get hpa "$scaler" -o jsonpath='{.spec.minReplicas}')
  max=$("${kubectl_cmd[@]}" -n "$namespace" get hpa "$scaler" -o jsonpath='{.spec.maxReplicas}')
  [[ "${min:-0}" -ge 1 && "${max:-0}" -ge "$min" ]] || fail "HPA inválido: $scaler"
done
"${kubectl_cmd[@]}" -n "$namespace" get scaledobject pulsar >/dev/null || fail "ScaledObject ausente: pulsar"

dependency_mode=cluster-owned
for dependency in postgres localstack identity alloy provider-sim webhook-sink; do
  addresses=$("${kubectl_cmd[@]}" -n "$namespace" get endpoints "$dependency" -o jsonpath='{range .subsets[*].addresses[*]}{.ip}{"\n"}{end}')
  [[ -n "$addresses" ]] || fail "endpoint ausente: $dependency"
  if printf '%s\n' "$addresses" | rg -qv '^10\.244\.'; then
    dependency_mode=compose-linked
  fi
done
if [[ "$require_owned_dependencies" == "1" && "$dependency_mode" != "cluster-owned" ]]; then
  fail "dependências ainda não estão materializadas no cluster (modo=$dependency_mode)"
fi

if [[ "$delete_pods" == "1" ]]; then
  for workload in cometa pulsar; do
    pod=$("${kubectl_cmd[@]}" -n "$namespace" get pods -l "app.kubernetes.io/name=$workload" -o jsonpath='{range .items[?(@.status.phase=="Running")]}{.metadata.name}{"\n"}{end}' | head -n 1)
    [[ -n "$pod" ]] || fail "pod Running ausente antes da recuperação: $workload"
    started=$(date +%s%3N)
    "${kubectl_cmd[@]}" -n "$namespace" delete pod "$pod" --wait=false >/dev/null
    wait_available "$workload"
    available=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.status.availableReplicas}')
    desired=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.spec.replicas}')
    if [[ "$available" != "$desired" ]]; then
      wait_available "$workload"
      available=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.status.availableReplicas}')
      desired=$("${kubectl_cmd[@]}" -n "$namespace" get deployment "$workload" -o jsonpath='{.spec.replicas}')
    fi
    [[ "$available" == "$desired" ]] || fail "$workload não recuperou $available/$desired"
    ended=$(date +%s%3N)
    printf 'pod_recovery workload=%s rto_ms=%s available=%s/%s\n' "$workload" "$((ended-started))" "$available" "$desired"
  done
fi

printf 'CONTINUITY_RUNTIME_PROOF=PASS nodes_ready=%s workloads=%s dependency_mode=%s rpo=restore-digest-gate\n' "$ready_nodes" "${#workloads[@]}" "$dependency_mode"
