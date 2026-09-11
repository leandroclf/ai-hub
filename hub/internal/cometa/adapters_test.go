package cometa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/contracts/files"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/providerauth"
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
	}, atlas.AdapterContract{}, "sync", "")
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
	request, err := adapter.BuildStatusRequest(context.Background(), "https://provider.example/api/", atlas.AdapterContract{}, "external/request 42")
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodGet || request.URL.String() != "https://provider.example/api/v1/operations/external%2Frequest%2042" {
		t.Fatalf("requisição de status não escapou a identidade externa: %s %s", request.Method, request.URL.String())
	}
}

type adapterTestSecretResolver struct{}

func (adapterTestSecretResolver) Resolve(_ context.Context, ref, version string) (providerauth.Secret, error) {
	if ref != "rest-secret" || version != "v7" {
		return providerauth.Secret{}, io.ErrUnexpectedEOF
	}
	return providerauth.Secret{Value: "rest-api-key-value", Version: "v7"}, nil
}

func TestVersionedRESTAdapterUsesAPIKeyAndAsyncMarker(t *testing.T) {
	t.Parallel()
	var submitKey, statusKey, submitMode string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			submitKey = r.Header.Get("X-API-Key")
			var payload restJSONSubmitRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("payload REST assíncrono inválido: %v", err)
			}
			submitMode = payload.Mode
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"provider_request_id":"external-async-42","status":"PENDING"}`)
			return
		}
		statusKey = r.Header.Get("X-API-Key")
		if r.Method != http.MethodGet || r.URL.Path != "/consulta/external-async-42" {
			t.Fatalf("consulta REST assíncrona inesperada: %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"provider_request_id":"external-async-42","status":"SUCCEEDED","detail":"real-rest-marker"}`)
	}))
	defer server.Close()

	adapter, err := NewAdapterRegistry("rest-json-v1").Get("rest-json-v1")
	if err != nil {
		t.Fatal(err)
	}
	command := dispatch.Command{CommandID: "command-async-42", ProtocolID: "protocol-async-42", RequestBody: map[string]string{"marker": "async"}}
	contract := atlas.AdapterContract{SubmitPath: "/analise", StatusPath: "/consulta/{id}"}
	restRequest, err := adapter.BuildSubmitRequest(context.Background(), server.URL, command, contract, "async_poll", "")
	if err != nil {
		t.Fatal(err)
	}
	if restRequest.URL.Path != "/analise" {
		t.Fatalf("path de submit REST não respeitou contrato declarativo: %s", restRequest.URL.Path)
	}
	cache := providerauth.NewTokenCache("127.0.0.1:1")
	defer cache.Close()
	cache.Resolver = adapterTestSecretResolver{}
	if err := cache.Apply(context.Background(), http.DefaultClient, "account", providerauth.Config{AuthType: providerauth.APIKey, APIKeyHeader: "X-API-Key", SecretRef: "rest-secret", SecretVersion: "v7"}, restRequest); err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(restRequest)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := adapter.DecodeResult(response.StatusCode, body)
	if err != nil || accepted.Status != "PENDING" {
		t.Fatalf("aceite assíncrono REST inválido: %+v %v", accepted, err)
	}
	statusRequest, err := adapter.BuildStatusRequest(context.Background(), server.URL, contract, accepted.ProviderRequestID)
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Apply(context.Background(), http.DefaultClient, "account", providerauth.Config{AuthType: providerauth.APIKey, APIKeyHeader: "X-API-Key", SecretRef: "rest-secret", SecretVersion: "v7"}, statusRequest); err != nil {
		t.Fatal(err)
	}
	statusResponse, err := http.DefaultClient.Do(statusRequest)
	if err != nil {
		t.Fatal(err)
	}
	statusBody, err := io.ReadAll(statusResponse.Body)
	_ = statusResponse.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	final, err := adapter.DecodeResult(statusResponse.StatusCode, statusBody)
	if err != nil || final.Status != "SUCCEEDED" || final.Detail != "real-rest-marker" {
		t.Fatalf("resultado assíncrono REST não preservou o marcador: %+v %v", final, err)
	}
	if submitKey != "rest-api-key-value" || statusKey != "rest-api-key-value" || submitMode != "async_poll" {
		t.Fatalf("autenticação/modo REST inesperados: submit_key=%q status_key=%q mode=%q", submitKey, statusKey, submitMode)
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
