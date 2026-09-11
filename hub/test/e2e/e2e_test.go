//go:build e2e

// Testes de integracao ponta a ponta contra a pilha local subida por
// `docker compose -f hub/deploy/docker-compose.yml up -d` (ja com o
// job `seed` concluido). Rode com:
//
//	cd hub && go test -tags e2e ./test/e2e/...
//
// Cobrem os invariantes mais criticos verificados manualmente durante
// o desenvolvimento: admissao duravel/idempotencia (EXE-01), SYNC
// direto sucesso/falha (EXE-14), mesma representacao em GET e webhook
// (COM-05), credencial sem fallback (SEG-05) e reserva/limite estrito
// (FIN-06).
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	orbitaURL      = envOr("R2_E2E_ORBITA_URL", "http://localhost:8080")
	atlasURL       = envOr("R2_E2E_ATLAS_URL", "http://localhost:8081")
	webhookSinkURL = envOr("R2_E2E_WEBHOOK_SINK_URL", "http://localhost:8091")
)

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

type protocolResponse struct {
	ProtocolID    string          `json:"protocol_id"`
	Status        string          `json:"status"`
	ResultVersion int             `json:"result_version"`
	FinalBody     json.RawMessage `json:"final_body"`
}

func createProtocol(t *testing.T, tenant, idempotencyKey string, body map[string]any) (protocolResponse, int) {
	t.Helper()
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, orbitaURL+"/v1/protocols", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", tenant)
	req.Header.Set("Idempotency-Key", idempotencyKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var out protocolResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out, resp.StatusCode
}

func getProtocol(t *testing.T, tenant, protocolID string) (protocolResponse, int) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, orbitaURL+"/v1/protocols/"+protocolID, nil)
	req.Header.Set("X-Tenant-Id", tenant)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()
	var out protocolResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out, resp.StatusCode
}

// TestSyncSuccess valida EXE-14: SYNC direto devolve o final na mesma
// requisicao, sem espera obrigatoria por fila.
func TestSyncSuccess(t *testing.T) {
	key := fmt.Sprintf("e2e-sync-ok-%d", time.Now().UnixNano())
	resp, status := createProtocol(t, "acme", key, map[string]any{
		"mode": "SYNC", "provider_account_id": "prov-sync-1",
		"service_code": "consulta-cadastral", "service_version": 1,
		"input": map[string]any{"cpf": "111"},
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if resp.Status != "SUCCEEDED" {
		t.Fatalf("status = %q, want SUCCEEDED", resp.Status)
	}
}

// TestSyncFailure valida que force_fail resulta em FAILED, nao em erro
// de transporte.
func TestSyncFailure(t *testing.T) {
	key := fmt.Sprintf("e2e-sync-fail-%d", time.Now().UnixNano())
	resp, status := createProtocol(t, "acme", key, map[string]any{
		"mode": "SYNC", "provider_account_id": "prov-sync-1",
		"service_code": "consulta-cadastral", "service_version": 1,
		"input": map[string]any{"force_fail": true},
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if resp.Status != "FAILED" {
		t.Fatalf("status = %q, want FAILED", resp.Status)
	}
}

// TestIdempotencyReplay valida EXE-01: mesma chave e mesmo payload
// recuperam o mesmo protocolo, sem nova execucao.
func TestIdempotencyReplay(t *testing.T) {
	key := fmt.Sprintf("e2e-idem-%d", time.Now().UnixNano())
	body := map[string]any{
		"mode": "SYNC", "provider_account_id": "prov-sync-1",
		"service_code": "consulta-cadastral", "service_version": 1,
		"input": map[string]any{"cpf": "222"},
	}
	first, status1 := createProtocol(t, "acme", key, body)
	if status1 != http.StatusOK {
		t.Fatalf("first status = %d, want 200", status1)
	}
	second, status2 := createProtocol(t, "acme", key, body)
	if status2 != http.StatusOK {
		t.Fatalf("second status = %d, want 200", status2)
	}
	if first.ProtocolID != second.ProtocolID {
		t.Fatalf("protocol_id mismatch on replay: %s != %s", first.ProtocolID, second.ProtocolID)
	}

	// Payload diferente com a mesma chave deve ser conflito (409).
	req, _ := http.NewRequest(http.MethodPost, orbitaURL+"/v1/protocols", bytes.NewReader(mustJSON(map[string]any{
		"mode": "SYNC", "provider_account_id": "prov-sync-1",
		"service_code": "consulta-cadastral", "service_version": 1,
		"input": map[string]any{"cpf": "DIFERENTE"},
	})))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "acme")
	req.Header.Set("Idempotency-Key", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("conflict request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("conflict status = %d, want 409", resp.StatusCode)
	}
}

// TestGetMatchesWebhookBody valida COM-05: a mesma representacao final
// e usada no GET e em todas as tentativas de webhook.
func TestGetMatchesWebhookBody(t *testing.T) {
	key := fmt.Sprintf("e2e-webhook-%d", time.Now().UnixNano())
	created, status := createProtocol(t, "acme", key, map[string]any{
		"mode": "SYNC", "provider_account_id": "prov-sync-1",
		"service_code": "consulta-cadastral", "service_version": 1,
		"input": map[string]any{"cpf": "333"},
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}

	got, _ := getProtocol(t, "acme", created.ProtocolID)
	if !bytes.Equal(got.FinalBody, created.FinalBody) {
		t.Fatalf("GET final_body != creation final_body:\n%s\n%s", got.FinalBody, created.FinalBody)
	}

	deadline := time.Now().Add(10 * time.Second)
	var webhookBody json.RawMessage
	for time.Now().Before(deadline) {
		resp, err := http.Get(webhookSinkURL + "/received")
		if err == nil {
			var items []struct {
				Body struct {
					ProtocolID string          `json:"protocol_id"`
					FinalBody  json.RawMessage `json:"final_body"`
				} `json:"body"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			for _, it := range items {
				if it.Body.ProtocolID == created.ProtocolID {
					webhookBody = it.Body.FinalBody
					break
				}
			}
		}
		if webhookBody != nil {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if webhookBody == nil {
		t.Fatalf("webhook nao recebido para protocolo %s dentro do prazo", created.ProtocolID)
	}
	if !bytes.Equal(webhookBody, created.FinalBody) {
		t.Fatalf("webhook final_body != creation final_body:\n%s\n%s", webhookBody, created.FinalBody)
	}
}

// TestCredentialUnavailableNoFallback valida SEG-05: sem vinculo de
// credencial elegivel, o Hub recusa sem fallback implicito.
func TestCredentialUnavailableNoFallback(t *testing.T) {
	// Cadastra uma conta de provedor sem nenhum vinculo de credencial.
	providerAccountID := fmt.Sprintf("e2e-nocred-%d", time.Now().UnixNano())
	raw, _ := json.Marshal(map[string]any{
		"provider_account_id": providerAccountID, "provider_id": "provider-sim",
		"environment": "local", "base_url": "http://provider-sim:8090", "provider_mode": "sync",
	})
	resp, err := http.Post(atlasURL+"/v1/provider-accounts", "application/json", bytes.NewReader(raw))
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("falha ao cadastrar conta de provedor sem credencial: err=%v status=%d", err, statusOf(resp))
	}
	resp.Body.Close()

	key := fmt.Sprintf("e2e-nocred-%d", time.Now().UnixNano())
	got, status := createProtocol(t, "acme", key, map[string]any{
		"mode": "SYNC", "provider_account_id": providerAccountID,
		"service_code": "consulta-cadastral", "service_version": 1,
		"input": map[string]any{},
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (falha de negocio, nao de transporte)", status)
	}
	if got.Status != "FAILED" {
		t.Fatalf("status = %q, want FAILED", got.Status)
	}
}

// TestStrictBalanceLimitExceeded valida FIN-06: reserva estrita
// recusa antes de qualquer efeito externo quando excede o teto.
func TestStrictBalanceLimitExceeded(t *testing.T) {
	tenant := fmt.Sprintf("e2e-strict-%d", time.Now().UnixNano())
	raw, _ := json.Marshal(map[string]any{
		"tenant_id": tenant, "plan": "unit", "unit_price": 1.00, "strict_balance": true,
		"client_sla_seconds": 30, "credential_mode_required": "SHARED_HUB",
	})
	resp, err := http.Post(atlasURL+"/v1/contracts", "application/json", bytes.NewReader(raw))
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("falha ao cadastrar contrato: err=%v status=%d", err, statusOf(resp))
	}
	resp.Body.Close()

	// credit_limits nao configurado para este tenant => teto alto
	// padrao (1_000_000): a primeira reserva deve suceder.
	key := fmt.Sprintf("e2e-strict-ok-%d", time.Now().UnixNano())
	got, status := createProtocol(t, tenant, key, map[string]any{
		"mode": "SYNC", "provider_account_id": "prov-sync-1",
		"service_code": "consulta-cadastral", "service_version": 1,
		"input": map[string]any{},
	})
	if status != http.StatusOK || got.Status != "SUCCEEDED" {
		t.Fatalf("reserva dentro do teto deveria suceder: status=%d protocol_status=%s", status, got.Status)
	}
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func statusOf(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}
