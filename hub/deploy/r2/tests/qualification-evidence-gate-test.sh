#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)"
VALIDATOR="${ROOT_DIR}/hub/deploy/r2/tests/validate-qualification-evidence.py"
FIXTURE_DIR="${ROOT_DIR}/hub/deploy/r2/tests/fixtures/qualification-evidence"

python3 "$VALIDATOR" --manifest "$FIXTURE_DIR/compatible.json" --root "$ROOT_DIR"
R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE=synthetic-fixture R2_PROMOTION_APPROVALS='P-01,P-08,P-10' R2_ENVIRONMENT_ISOLATION_PROOF=PASS R2_QUALIFICATION_EVIDENCE_MANIFEST="$FIXTURE_DIR/compatible.json" "${ROOT_DIR}/hub/deploy/r2/tests/promotion-gate.sh"

if python3 "$VALIDATOR" --manifest "$FIXTURE_DIR/incompatible.json" --root "$ROOT_DIR"; then
  echo "QUALIFICATION_EVIDENCE_TEST=FAIL incompatible_evidence_was_accepted"
  exit 1
fi

if R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE=synthetic-fixture R2_PROMOTION_APPROVALS='P-01,P-08,P-10' R2_ENVIRONMENT_ISOLATION_PROOF=PASS R2_QUALIFICATION_EVIDENCE_MANIFEST="$FIXTURE_DIR/incompatible.json" "${ROOT_DIR}/hub/deploy/r2/tests/promotion-gate.sh"; then
  echo "QUALIFICATION_EVIDENCE_TEST=FAIL promotion_gate_accepted_incompatible_evidence"
  exit 1
fi

echo "QUALIFICATION_EVIDENCE_TEST=PASS incompatible_evidence_blocked"
