#!/usr/bin/env bash
set -euo pipefail

# Sobe o perfil local-kind com as dependências materializadas no cluster.
# Diferentemente do bootstrap compatível, nenhum Service de dependência recebe
# Endpoints de containers Compose. Os dados são sintéticos e os volumes são
# efêmeros de laboratório; restore/backup continua sendo qualificado pelo gate
# separado.
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
tool_dir=${R2_TOOL_DIR:-/tmp/ai-hub-r2-tools}
namespace=${R2_KIND_NAMESPACE:-ai-hub-local-kind}
kind_bin=${KIND_BIN:-$tool_dir/kind}
kubeconfig=${R2_KUBECONFIG:-$tool_dir/kubeconfig}
kubectl_cmd=(kubectl --kubeconfig "$kubeconfig")
mkdir -p "$tool_dir"

if [[ ! -x "$kind_bin" ]]; then
  curl -fsSL https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64 -o "$kind_bin"
  curl -fsSL https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64.sha256sum -o "$tool_dir/kind.sha256sum"
  python3 - "$kind_bin" "$tool_dir/kind.sha256sum" <<'PY'
import hashlib, pathlib, sys
assert hashlib.sha256(pathlib.Path(sys.argv[1]).read_bytes()).hexdigest() == pathlib.Path(sys.argv[2]).read_text().split()[0]
PY
  chmod +x "$kind_bin"
fi
if ! "$kind_bin" get clusters | rg -qx 'ai-hub-r2'; then
  "$kind_bin" create cluster --name ai-hub-r2 --image kindest/node:v1.32.2 --config "$root/kind/cluster.yaml" --kubeconfig "$kubeconfig" --wait 90s
else
  "$kind_bin" export kubeconfig --name ai-hub-r2 --kubeconfig "$kubeconfig"
fi

# O perfil independente deve preparar os operadores que seus manifests
# declaram. Sem estes CRDs, um cluster novo falha somente no apply do
# ScaledObject e deixa um laboratório parcialmente materializado. Os
# manifestos e versões são fixos no diretório de ferramentas para permitir
# reexecuções idempotentes; métricas-server usa TLS inseguro apenas no kind.
metrics_server_manifest="$tool_dir/metrics-server.yaml"
keda_manifest="$tool_dir/keda.yaml"
if [[ ! -s "$metrics_server_manifest" ]]; then
  curl -fsSL https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.7.2/components.yaml -o "$metrics_server_manifest"
fi
if [[ ! -s "$keda_manifest" ]]; then
  curl -fsSL https://github.com/kedacore/keda/releases/download/v2.17.2/keda-2.17.2.yaml -o "$keda_manifest"
fi
"${kubectl_cmd[@]}" apply -f "$metrics_server_manifest"
"${kubectl_cmd[@]}" -n kube-system patch deployment metrics-server --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp","--secure-port=10250","--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname","--kubelet-use-node-status-port","--metric-resolution=15s","--kubelet-insecure-tls"]}]}}}}'
"${kubectl_cmd[@]}" apply --server-side -f "$keda_manifest"
"${kubectl_cmd[@]}" wait --for=condition=Established crd/scaledobjects.keda.sh --timeout=180s

apply_configmap_file() {
  local name=$1 key=$2 file=$3
  "${kubectl_cmd[@]}" -n "$namespace" create configmap "$name" --from-file="$key=$file" --dry-run=client -o yaml | "${kubectl_cmd[@]}" apply -f -
}
apply_configmap_dir() {
  local name=$1 directory=$2
  "${kubectl_cmd[@]}" -n "$namespace" create configmap "$name" --from-file="$directory" --dry-run=client -o yaml | "${kubectl_cmd[@]}" apply -f -
}

"${kubectl_cmd[@]}" create namespace "$namespace" --dry-run=client -o yaml | "${kubectl_cmd[@]}" apply -f -
apply_configmap_file identity-realm realm.json "$root/identity/realm.json"
apply_configmap_file identity-reconcile reconcile_identity.py "$root/scripts/reconcile_identity.py"
apply_configmap_file migrate-script migrate.sh "$root/scripts/migrate.sh"
apply_configmap_dir control-migrations "$root/../../migrations/control"
apply_configmap_dir core-migrations "$root/../../migrations/core"
apply_configmap_dir finance-migrations "$root/../../migrations/finance"
apply_configmap_file alloy-config config.alloy "$root/observability/config-independent.alloy"
apply_configmap_file loki-config config.yaml "$root/observability/loki.yaml"
apply_configmap_file tempo-config tempo.yaml "$root/observability/tempo.yaml"
apply_configmap_file prometheus-config prometheus.yml "$root/observability/prometheus.yml"
apply_configmap_file kong-config kong.yml "$root/kong.yml"

for image in \
  ai-hub-r2-atlas:r2 ai-hub-r2-orbita:r2 ai-hub-r2-cometa:r2 ai-hub-r2-pulsar:r2 ai-hub-r2-libra:r2 \
  ai-hub-r2-provider-sim:r2 ai-hub-r2-webhook-sink:r2 ai-hub-r2-admin-ui:r2 \
  ai-hub-kind-postgres:16 ai-hub-kind-localstack:3.8 ai-hub-kind-keycloak:26.7.3 \
  ai-hub-kind-loki:3.3.2 ai-hub-kind-tempo:2.6.1 ai-hub-kind-alloy:v1.5.1 \
  ai-hub-kind-prometheus:v3.1.0 ai-hub-kind-grafana:11.4.0 ai-hub-kind-kong:3.8 ai-hub-kind-python:3.12-alpine; do
  "$kind_bin" load docker-image "$image" --name ai-hub-r2
done

"${kubectl_cmd[@]}" -n "$namespace" delete job hub-migrate identity-reconcile --ignore-not-found >/dev/null
python3 "$root/kind/render-independent-dependencies.py" | "${kubectl_cmd[@]}" apply -f -
"${kubectl_cmd[@]}" -n "$namespace" rollout status deployment/postgres --timeout=180s
"${kubectl_cmd[@]}" -n "$namespace" wait --for=condition=complete job/hub-migrate --timeout=300s
"${kubectl_cmd[@]}" -n "$namespace" rollout status deployment/identity --timeout=300s
"${kubectl_cmd[@]}" -n "$namespace" wait --for=condition=complete job/identity-reconcile --timeout=300s

R2_KIND_DEPENDENCIES=cluster R2_KIND_NAMESPACE="$namespace" python3 "$root/kind/render-runtime.py" | "${kubectl_cmd[@]}" apply -f -
"${kubectl_cmd[@]}" apply -k "$root/k8s/overlays/local-kind"
for workload in atlas orbita cometa pulsar libra; do
  "${kubectl_cmd[@]}" -n "$namespace" rollout status "deployment/$workload" --timeout=180s
done
"${kubectl_cmd[@]}" -n "$namespace" rollout status deployment/admin-ui --timeout=180s
printf 'KIND_INDEPENDENT_BOOTSTRAP=PASS namespace=%s dependencies=cluster-owned workloads=5 ui=admin-ui gateway=kong\n' "$namespace"
printf '%s\n' 'Port-forwardes: kubectl --kubeconfig '"$kubeconfig"' -n '"$namespace"' port-forward svc/admin-ui 13000:8080 svc/identity 18085:8080'
