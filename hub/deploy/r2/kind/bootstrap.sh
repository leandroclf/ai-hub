#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
tool_dir=${R2_TOOL_DIR:-/tmp/ai-hub-r2-tools}
compose_project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
compose=(docker compose -p "$compose_project" -f "$root/compose.yaml")
mkdir -p "$tool_dir"
kind_bin=${KIND_BIN:-$tool_dir/kind}
kubeconfig="$tool_dir/kubeconfig"
if [[ ! -x "$kind_bin" ]]; then
  curl -fsSL https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64 -o "$kind_bin"
  curl -fsSL https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64.sha256sum -o "$tool_dir/kind.sha256sum"
  python3 - "$kind_bin" "$tool_dir/kind.sha256sum" <<'PY'
import hashlib,pathlib,sys
assert hashlib.sha256(pathlib.Path(sys.argv[1]).read_bytes()).hexdigest()==pathlib.Path(sys.argv[2]).read_text().split()[0]
PY
  chmod +x "$kind_bin"
fi
if ! "$kind_bin" get clusters | rg -qx 'ai-hub-r2'; then
  "$kind_bin" create cluster --name ai-hub-r2 --image kindest/node:v1.32.2 --config "$root/kind/cluster.yaml" --kubeconfig "$kubeconfig" --wait 90s
else
  # Reexporta o contexto para execuções idempotentes após limpeza do diretório
  # temporário de ferramentas. O cluster existente continua sendo a fonte de
  # verdade e não é recriado nem tem volumes alterados.
  "$kind_bin" export kubeconfig --name ai-hub-r2 --kubeconfig "$kubeconfig"
fi
for node in $("$kind_bin" get nodes --name ai-hub-r2); do
  compose_network="${compose_project}_default"
  if ! docker inspect "$node" --format '{{json .NetworkSettings.Networks}}' | jq -e --arg network "$compose_network" 'has($network)' >/dev/null; then docker network connect "$compose_network" "$node"; fi
done
# Os endpoints das dependências são containers Compose fora do CIDR de pods.
# Sem masqueradeAll, o retorno desses containers não conhece as redes 10.244/16
# e os workloads ficam em CrashLoop mesmo com o endpoint acessível no nó.
if kubectl --kubeconfig "$kubeconfig" -n kube-system get cm kube-proxy -o jsonpath='{.data.config\.conf}' | rg -q 'masqueradeAll: false'; then
  kubectl --kubeconfig "$kubeconfig" -n kube-system get cm kube-proxy -o json \
    | python3 -c 'import json,sys; o=json.load(sys.stdin); o["data"]["config.conf"]=o["data"]["config.conf"].replace("masqueradeAll: false","masqueradeAll: true",1); print(json.dumps(o))' \
    | kubectl --kubeconfig "$kubeconfig" apply -f -
  kubectl --kubeconfig "$kubeconfig" -n kube-system rollout restart daemonset/kube-proxy
  kubectl --kubeconfig "$kubeconfig" -n kube-system rollout status daemonset/kube-proxy --timeout=60s
fi
curl -fsSL https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.7.2/components.yaml -o "$tool_dir/metrics-server.yaml"
curl -fsSL https://github.com/kedacore/keda/releases/download/v2.17.2/keda-2.17.2.yaml -o "$tool_dir/keda.yaml"
kubectl --kubeconfig "$kubeconfig" apply -f "$tool_dir/metrics-server.yaml"
# Certificados kubelet autoassinados apenas neste laboratório kind.
kubectl --kubeconfig "$kubeconfig" -n kube-system patch deployment metrics-server --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp","--secure-port=10250","--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname","--kubelet-use-node-status-port","--metric-resolution=15s","--kubelet-insecure-tls"]}]}}}}'
kubectl --kubeconfig "$kubeconfig" apply --server-side -f "$tool_dir/keda.yaml"
for service in atlas orbita cometa pulsar libra; do "$kind_bin" load docker-image "ai-hub-r2-$service:r2" --name ai-hub-r2; done
for domain in control core finance; do
  "${compose[@]}" exec -T postgres psql -U hub -d postgres -v ON_ERROR_STOP=1 <<SQL
SELECT 'CREATE DATABASE hub_${domain}_kind' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='hub_${domain}_kind')\gexec
SQL
done
"${compose[@]}" run --rm -e MIGRATION_DB_SUFFIX=_kind migrate
python3 "$root/kind/render-runtime.py" > "$tool_dir/runtime.yaml"
kubectl --kubeconfig "$kubeconfig" apply -f "$tool_dir/runtime.yaml"
kubectl --kubeconfig "$kubeconfig" apply -k "$root/k8s/overlays/local-kind"
kubectl --kubeconfig "$kubeconfig" -n ai-hub-local-kind wait --for=condition=available deployment --all --timeout=60s
