#!/usr/bin/env bash
set -euo pipefail

# Gate reutilizável por CI/CD. Desenvolvimento pode prosseguir sem a
# qualificação de produção; prd exige perfil, isolamento e aprovações
# explícitas. Nenhum valor de segredo é aceito por este contrato.
environment=${R2_PROMOTION_ENV:-dev}
profile=${R2_QUALIFICATION_PROFILE:-}
approvals=" ${R2_PROMOTION_APPROVALS:-} "
isolation=${R2_ENVIRONMENT_ISOLATION_PROOF:-}
criticality=${R2_CRITICALITY_PROFILE:-standard}
critical_approved=${R2_CRITICAL_PROFILE_APPROVED:-}
regional_profile=${R2_REGIONAL_PROFILE:-}
regional_custody_confirmation=${R2_REGIONAL_CUSTODY_CONFIRMATION:-}
regional_custody_proof=${R2_REGIONAL_CUSTODY_PROOF:-}
regional_fencing_proof=${R2_REGIONAL_FENCING_PROOF:-}
evidence_manifest=${R2_QUALIFICATION_EVIDENCE_MANIFEST:-}
evidence_validator="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/validate-qualification-evidence.py"

has_approval(){ [[ "$approvals" == *" $1 "* || "$approvals" == *" $1,"* || "$approvals" == *",$1 "* || "$approvals" == *",$1,"* ]]; }

if [[ -n "$evidence_manifest" ]]; then
  python3 "$evidence_validator" --manifest "$evidence_manifest" --root "$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../" && pwd)" --require-pass || {
    echo "PROMOTION_GATE=BLOCK environment=$environment missing=qualification_evidence"
    exit 1
  }
fi

case "$environment" in
  local|dev|hom|ppd)
    echo "PROMOTION_GATE=ALLOW environment=$environment reason=non-production-profile"
    exit 0
    ;;
  prd)
    missing=()
    [[ -n "$profile" ]] || missing+=(qualification_profile)
    [[ "$isolation" == PASS ]] || missing+=(environment_isolation_proof)
    has_approval P-01 || missing+=(P-01)
    has_approval P-08 || missing+=(P-08)
    has_approval P-10 || missing+=(P-10)
    if [[ "$criticality" == critical && "$critical_approved" != PASS ]]; then
      missing+=(critical_profile_approval)
    fi
    if [[ "$regional_profile" == RPO_ZERO ]]; then
      [[ "$regional_custody_confirmation" == PASS ]] || missing+=(regional_custody_confirmation)
      [[ "$regional_custody_proof" == PASS ]] || missing+=(regional_custody_proof)
      [[ "$regional_fencing_proof" == PASS ]] || missing+=(regional_fencing_proof)
    fi
    if ((${#missing[@]} > 0)); then
      echo "PROMOTION_GATE=BLOCK environment=prd missing=$(IFS=,; echo "${missing[*]}")"
      exit 1
    fi
    echo "PROMOTION_GATE=ALLOW environment=prd profile=$profile approvals=P-01,P-08,P-10 regional_profile=${regional_profile:-none}"
    ;;
  *)
    echo "PROMOTION_GATE=BLOCK environment=$environment missing=unsupported_environment"
    exit 1
    ;;
esac
