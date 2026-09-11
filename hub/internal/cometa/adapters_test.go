package cometa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-hub/hub/internal/contracts/files"
	"ai-hub/hub/internal/dispatch"
)

func TestAdapterRegistryFailsClosedForUnknownRuntimeAdapter(t *testing.T) {
	r := NewAdapterRegistry("provider-sim")
	if !r.Supports("provider-sim") {
		t.Fatal("adapter instalado foi recusado")
	}
	if r.Supports("provider-not-installed") {
		t.Fatal("adapter não instalado foi aceito")
	}
}

func TestAdapterRegistryReadsExplicitRuntimeAllowlist(t *testing.T) {
	t.Setenv("COMETA_ADAPTERS", "provider-acme-v1, provider-sim")
	r := AdapterRegistryFromEnv()
	if r.Supports("provider-acme-v1") || !r.Supports("provider-sim") || r.Supports("synthetic-provider") {
		t.Fatal("allowlist explícita não foi aplicada")
	}
}

func TestAdapterRegistryDoesNotTurnCatalogIDIntoTransport(t *testing.T) {
	if NewAdapterRegistry("provider-acme-v1").Supports("provider-acme-v1") {
		t.Fatal("ID de catálogo sem implementação compilada foi habilitado")
	}
}

func TestAdapterRegistrySupportsVersionedRESTAdapter(t *testing.T) {
	if !NewAdapterRegistry("rest-json-v1").Supports("rest-json-v1") {
		t.Fatal("adapter REST versionado homologado não foi habilitado")
	}
}

func TestVersionedRESTAdapterUsesIndependentHTTPContract(t *testing.T) {
	t.Parallel()
	var received struct {
		IdempotencyKey string `json:"idempotency_key"`
		Mode           string `json:"mode"`
		Input          struct {
			Marker string `json:"marker"`
		} `json:"input"`
		FileRefs []struct {
			ID      string `json:"file_id"`
			SHA256  string `json:"sha256"`
			Version string `json:"version"`
		} `json:"file_refs"`
	}
	var method, path string
	var decodeErr error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		body, err := io.ReadAll(r.Body)
		decodeErr = err
		if decodeErr == nil {
			decodeErr = json.Unmarshal(body, &received)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"provider_request_id":"external-42","status":"SUCCEEDED"}`)
	}))
	defer server.Close()

	registry := NewAdapterRegistry("rest-json-v1")
	adapter, err := registry.Get("rest-json-v1")
	if err != nil {
		t.Fatal(err)
	}
	request, err := adapter.BuildSubmitRequest(context.Background(), server.URL, dispatch.Command{
		CommandID:   "command-42",
		ProtocolID:  "protocol-42",
		RequestBody: map[string]string{"marker": "from-client"},
		FileRefs: []files.Reference{{ID: "file-42", Version: "v3", SHA256: "sha-42", Size: 7,
			ContentType: "application/json", Purpose: "INPUT", State: "READY"}},
	}, "sync", "")
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.DecodeResult(response.StatusCode, body)
	if err != nil {
		t.Fatal(err)
	}
	if result.ProviderRequestID != "external-42" || result.Status != "SUCCEEDED" {
		t.Fatalf("resposta REST não foi normalizada: %+v", result)
	}
	if method != http.MethodPost || path != "/v1/operations" || decodeErr != nil {
		t.Fatalf("endpoint/payload REST inesperado: method=%s path=%s err=%v", method, path, decodeErr)
	}
	if received.IdempotencyKey != "protocol-42" || received.Mode != "sync" || received.Input.Marker != "from-client" || len(received.FileRefs) != 1 || received.FileRefs[0].ID != "file-42" || received.FileRefs[0].Version != "v3" || received.FileRefs[0].SHA256 != "sha-42" {
		t.Fatalf("contrato REST perdeu identidade/modo/entrada: %+v", received)
	}
}

func TestVersionedRESTAdapterBuildsEscapedStatusRequest(t *testing.T) {
	adapter, err := NewAdapterRegistry("rest-json-v1").Get("rest-json-v1")
	if err != nil {
		t.Fatal(err)
	}
	request, err := adapter.BuildStatusRequest(context.Background(), "https://provider.example/api/", "external/request 42")
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodGet || request.URL.String() != "https://provider.example/api/v1/operations/external%2Frequest%2042" {
		t.Fatalf("requisição de status não escapou a identidade externa: %s %s", request.Method, request.URL.String())
	}
}

func TestVersionedRESTAdapterRejectsInvalidResponses(t *testing.T) {
	adapter, err := NewAdapterRegistry("rest-json-v1").Get("rest-json-v1")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{name: "campo desconhecido", status: http.StatusOK, body: `{"provider_request_id":"id","status":"SUCCEEDED","secret":"leak"}`},
		{name: "documentos concatenados", status: http.StatusOK, body: `{"provider_request_id":"id","status":"SUCCEEDED"}{}`},
		{name: "status HTTP inesperado", status: http.StatusBadGateway, body: `{"provider_request_id":"id","status":"SUCCEEDED"}`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := adapter.DecodeResult(testCase.status, []byte(testCase.body)); err == nil {
				t.Fatal("resposta inválida foi aceita")
			}
		})
	}
}
