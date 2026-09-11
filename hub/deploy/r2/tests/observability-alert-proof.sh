#!/usr/bin/env bash
set -euo pipefail

# Prova estrutural dos alertas locais: regra carregável, runbook existente,
# dashboard presente e ausência de identificadores de negócio como labels.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
ALERTS_FILE="${ROOT_DIR}/hub/deploy/r2/observability/alerts.yml"
PROM_FILE="${ROOT_DIR}/hub/deploy/r2/observability/prometheus.yml"
RUNBOOK_FILE="${ROOT_DIR}/hub/deploy/r2/RUNBOOKS.md"
DASHBOARD_FILE="${ROOT_DIR}/hub/deploy/r2/observability/dashboards/hub.json"
EVIDENCE_FILE="${R2_OBSERVABILITY_ALERT_EVIDENCE_FILE:-${ROOT_DIR}/hub/evidence/r2/execution/observability-alert-latest.log}"

mkdir -p "$(dirname "$EVIDENCE_FILE")"
exec > >(tee "$EVIDENCE_FILE")

echo "OBSERVABILITY_ALERT_PROOF=START"
test -s "$ALERTS_FILE" && test -s "$PROM_FILE" && test -s "$RUNBOOK_FILE" && test -s "$DASHBOARD_FILE"
grep -Fq 'rule_files: [/etc/prometheus/alerts.yml]' "$PROM_FILE"

while IFS= read -r url; do
  path="${url%%#*}"
  anchor="${url#*#}"
  test -f "${ROOT_DIR}/${path}"
  grep -Fq "## \`${anchor}\`" "${ROOT_DIR}/${path}"
  echo "runbook PASS path=${path} anchor=${anchor}"
done < <(sed -n "s/^[[:space:]]*runbook_url: *'\([^']*\)'.*/\1/p" "$ALERTS_FILE")

if grep -Eq 'protocol_id|tenant_id|operation_id|attempt_id' "$ALERTS_FILE"; then
  echo "OBSERVABILITY_ALERT_PROOF=FAIL business_identifier_label_detected"
  exit 1
fi

alert_count="$(grep -c '^      - alert:' "$ALERTS_FILE")"
test "$alert_count" -ge 3
echo "OBSERVABILITY_ALERT_PROOF=PASS alerts=${alert_count} runbooks=resolved labels=bounded dashboard=present"
echo "evidence=${EVIDENCE_FILE#${ROOT_DIR}/}"
