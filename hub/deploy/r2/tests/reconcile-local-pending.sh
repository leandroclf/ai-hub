#!/usr/bin/env bash
set -euo pipefail

# Fecha somente obrigações da fixture local cujo oráculo sintético prova que o
# protocol_id nunca foi registrado. Uma resposta 200 do oráculo é bloqueio:
# nesse caso a obrigação exige reconciliação normal com o provider_request_id.
project=${R2_COMPOSE_PROJECT:-ai_hub_r3qual}
postgres=${R2_POSTGRES_CONTAINER:-${project}-postgres-1}
provider_url=${R2_PROVIDER_URL:-http://127.0.0.1:18090}
confirm=${R2_LOCAL_RECONCILIATION_CONFIRM:-}
evidence_file=${R2_RECONCILIATION_EVIDENCE_FILE:-hub/evidence/r2/execution/capacity-reconciliation-latest.log}

if [[ "$confirm" != "I_UNDERSTAND_LOCAL_FIXTURE" ]]; then
  echo "reconciliação local não executada: defina R2_LOCAL_RECONCILIATION_CONFIRM=I_UNDERSTAND_LOCAL_FIXTURE" >&2
  exit 2
fi

mkdir -p "$(dirname "$evidence_file")"
exec > >(tee "$evidence_file")

# A capacidade normalmente referencia uma operação durável. Durante uma
# interrupção entre a confirmação da outbox e a projeção de operações, a
# operação pode não estar mais disponível, mas o evento operation.observed
# continua sendo a fonte de identidade do protocolo. Esse fallback só aceita
# eventos UNKNOWN sem provider_request_id; qualquer correlação externa mantém
# a obrigação protegida.
query="SELECT p.domain_id,
              p.permit_id,
              COALESCE(o.command->>'protocol_id', observed.payload->>'protocol_id')
       FROM capacity_permits p
       LEFT JOIN operations o ON o.operation_id=p.permit_id::uuid
       LEFT JOIN LATERAL (
         SELECT payload
         FROM outbox
         WHERE aggregate_id=p.permit_id::text
           AND event_type='operation.observed'
         ORDER BY created_at DESC
         LIMIT 1
       ) observed ON true
       WHERE p.pending_external
         AND p.domain_id LIKE 'r4-%'
         AND COALESCE(o.provider_request_id, observed.payload->>'provider_request_id','')=''
         AND COALESCE(o.command->>'protocol_id', observed.payload->>'protocol_id','')<>''
       ORDER BY p.domain_id,p.created_at"

rows=$(docker exec "$postgres" psql -U hub -d hub_core -At -F $'\t' -v ON_ERROR_STOP=1 -c "$query")
closed=0
protected=0
while IFS=$'\t' read -r domain permit protocol; do
  [[ -z "$domain" ]] && continue
  [[ "$domain" =~ ^r4-[a-z0-9-]+$ ]] || { echo "invalid domain: $domain" >&2; exit 1; }
  [[ "$permit" =~ ^[0-9a-f-]{36}$ && "$protocol" =~ ^[0-9a-f-]{36}$ ]] || { echo "invalid reconciliation identity" >&2; exit 1; }
  status=$(curl -sS -o /dev/null -w '%{http_code}' "$provider_url/__qualification/protocols/$protocol")
  case "$status" in
    404)
      evidence="local-oracle-absence:$protocol"
      docker exec "$postgres" psql -U hub -d hub_core -v ON_ERROR_STOP=1 -c "UPDATE capacity_permits SET pending_external=false,completed_at=clock_timestamp(),evidence_ref='$evidence' WHERE domain_id='$domain' AND permit_id='$permit' AND pending_external" >/dev/null
      echo "CLOSED domain=$domain permit=$permit protocol=$protocol evidence=$evidence"
      closed=$((closed+1))
      ;;
    200)
      echo "PROTECTED provider-effect-present domain=$domain permit=$permit protocol=$protocol" >&2
      protected=$((protected+1))
      ;;
    *)
      echo "provider oracle unavailable for protocol=$protocol status=$status" >&2
      exit 1
      ;;
  esac
done <<< "$rows"

printf 'summary closed=%d protected=%d\n' "$closed" "$protected"
if (( protected > 0 )); then
  exit 1
fi
