#!/usr/bin/env bash
set -euo pipefail

# Valida a renderização determinística dos overlays sem aplicar recursos no
# cluster. A prova é deliberadamente estrutural: não promove por si só os
# cenários de máquina limpa, recriação de nós ou dependências gerenciadas.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
KUBECTL_BIN="${KUBECTL_BIN:-kubectl}"
EVIDENCE_FILE="${R2_KUSTOMIZE_EVIDENCE_FILE:-${ROOT_DIR}/hub/evidence/r2/execution/kustomize-overlay-latest.log}"
TMP_DIR="$(mktemp -d /tmp/ai-hub-kustomize-overlay.XXXXXX)"
trap 'find "$TMP_DIR" -depth -delete' EXIT

mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

overlays=(local-kind dev hom ppd prd)
expected_images=(
  "ai-hub-r2-atlas:r2"
  "ai-hub-r2-orbita:r2"
  "ai-hub-r2-cometa:r2"
  "ai-hub-r2-pulsar:r2"
  "ai-hub-r2-libra:r2"
)

echo "KUSTOMIZE_OVERLAY_PROOF=START"
echo "root=${ROOT_DIR}"
kubectl_version="$($KUBECTL_BIN version --client -o json | tr '\n' ' ')"
echo "kubectl=${kubectl_version% }"

for environment in "${overlays[@]}"; do
  output_file="${TMP_DIR}/${environment}.yaml"
  overlay_dir="${ROOT_DIR}/hub/deploy/r2/k8s/overlays/${environment}"
  "$KUBECTL_BIN" kustomize "$overlay_dir" >"$output_file"

  namespace="$(awk '/^  namespace:/{print $2; exit}' "$output_file")"
  expected_namespace="ai-hub-${environment}"
  [[ "$namespace" == "$expected_namespace" ]] || {
    echo "overlay=${environment} FAIL namespace=${namespace} expected=${expected_namespace}"
    exit 1
  }

  resource_count="$(rg -c '^kind:' "$output_file" || true)"
  image_count="$(rg -c 'image:' "$output_file" || true)"
  [[ "$resource_count" == 20 && "$image_count" == 5 ]] || {
    echo "overlay=${environment} FAIL resources=${resource_count} images=${image_count}"
    exit 1
  }

  while IFS= read -r image; do
    case " ${expected_images[*]} " in
      *" ${image#*: } "*) ;;
      *)
        echo "overlay=${environment} FAIL unexpected_image=${image}"
        exit 1
        ;;
    esac
  done < <(rg 'image:' "$output_file" | sed 's/.*image: //')

  for kind in Service Deployment PodDisruptionBudget HorizontalPodAutoscaler ScaledObject; do
    kind_count="$(rg -c "^kind: ${kind}$" "$output_file" || true)"
    expected_kind_count=5
    [[ "$kind" == HorizontalPodAutoscaler ]] && expected_kind_count=4
    [[ "$kind" == ScaledObject ]] && expected_kind_count=1
    [[ "$kind_count" == "$expected_kind_count" ]] || {
      echo "overlay=${environment} FAIL kind=${kind} count=${kind_count} expected=${expected_kind_count}"
      exit 1
    }
  done

  if [[ "$environment" == prd ]]; then
    if rg -n 'ENVIRONMENT: local|127\.0\.0\.1|localhost|ai-hub-local-kind' "$output_file"; then
      echo "overlay=${environment} FAIL local_mode_detected"
      exit 1
    fi
  fi

  digest="$(sha256sum "$output_file" | awk '{print $1}')"
  echo "overlay=${environment} PASS namespace=${namespace} resources=${resource_count} images=${image_count} sha256=${digest}"
done

echo "KUSTOMIZE_OVERLAY_PROOF=PASS overlays=${#overlays[@]} production_local_mode=absent"
echo "evidence=${EVIDENCE_FILE#${ROOT_DIR}/}"
