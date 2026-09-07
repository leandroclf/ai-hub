#!/bin/bash
# Gera trafego real contra a stack local para produzir evidencia de
# testes (banco de dados, filas, S3, logs) documentada em
# hub/evidence/EVIDENCE.md. Nao e um teste automatizado (ver
# hub/test/e2e para isso); e um script de geracao de evidencia.
set -uo pipefail

ORBITA=http://localhost:8080
ATLAS=http://localhost:8081

req() {
  local desc="$1"; shift
  echo "### $desc"
  curl -s -w '\nHTTP %{http_code}\n' "$@"
  echo
}

req "1. SYNC sucesso (consulta-cadastral, prov-sync-1, tenant acme)" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-sync-ok-001' \
  -d '{"mode":"SYNC","provider_account_id":"prov-sync-1","service_code":"consulta-cadastral","service_version":1,"input":{"cpf":"11111111111"}}'

req "2. SYNC falha forcada" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-sync-fail-001' \
  -d '{"mode":"SYNC","provider_account_id":"prov-sync-1","service_code":"consulta-cadastral","service_version":1,"input":{"force_fail":true}}'

req "3. ASYNC polling, produto assincrono, provedor A (prov-poll-1)" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-async-poll-a-001' \
  -d '{"mode":"ASYNC","provider_account_id":"prov-poll-1","service_code":"protocolo-assincrono","service_version":1,"input":{"delay_ms":2500}}'

req "4. ASYNC polling, produto assincrono, provedor B (prov-poll-2)" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-async-poll-b-001' \
  -d '{"mode":"ASYNC","provider_account_id":"prov-poll-2","service_code":"protocolo-assincrono","service_version":1,"input":{"delay_ms":2500}}'

req "5. ASYNC callback" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-async-callback-001' \
  -d '{"mode":"ASYNC","provider_account_id":"prov-callback-1","service_code":"protocolo-assincrono","service_version":1,"input":{"delay_ms":2500}}'

req "6. AUTO" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-auto-001' \
  -d '{"mode":"AUTO","provider_account_id":"prov-poll-1","service_code":"protocolo-assincrono","service_version":1,"input":{"delay_ms":1500}}'

req "7. Saldo estrito - reserva dentro do limite (tenant acme-strict)" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme-strict' -H 'Idempotency-Key: evid-strict-ok-001' \
  -d '{"mode":"SYNC","provider_account_id":"prov-sync-1","service_code":"consulta-cadastral","service_version":1,"input":{"cpf":"22222222222"}}'

req "8. Credencial dedicada (tenant acme-dedicated)" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme-dedicated' -H 'Idempotency-Key: evid-dedicated-001' \
  -d '{"mode":"SYNC","provider_account_id":"prov-sync-1","service_code":"consulta-cadastral","service_version":1,"input":{"cpf":"33333333333"}}'

echo "### 9. Cadastrando conta de provedor sem credencial (para o cenario de recusa)"
curl -s -X POST "$ATLAS/v1/provider-accounts" -H 'Content-Type: application/json' -d '{
  "provider_account_id": "prov-nocred-evid", "provider_id": "provider-sim", "environment": "local",
  "base_url": "http://provider-sim:8090", "provider_mode": "sync"
}'
echo
req "9b. SYNC com credencial indisponivel (SEG-05, sem fallback)" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-nocred-001' \
  -d '{"mode":"SYNC","provider_account_id":"prov-nocred-evid","service_code":"consulta-cadastral","service_version":1,"input":{}}'

req "10. Idempotencia - repeticao da chave do item 1 (mesmo payload)" \
  -X POST "$ORBITA/v1/protocols" -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: acme' -H 'Idempotency-Key: evid-sync-ok-001' \
  -d '{"mode":"SYNC","provider_account_id":"prov-sync-1","service_code":"consulta-cadastral","service_version":1,"input":{"cpf":"11111111111"}}'

echo "Aguardando 6s para as operacoes assincronas (itens 3-6) resolverem..."
sleep 6
echo "trafego gerado. Consulte hub/evidence/db, hub/evidence/queues e hub/evidence/logs para a evidencia capturada a partir daqui."
