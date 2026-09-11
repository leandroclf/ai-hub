package cometa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"ai-hub/hub/internal/contracts/files"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/providersim"
)

var ErrAdapterUnavailable = errors.New("cometa: adapter não habilitado no runtime")

// AdapterRegistry é a fronteira entre a homologação publicada no Atlas e a
// capacidade efetivamente carregada no Cometa. O catálogo pode conservar
// versões históricas, mas somente IDs explicitamente instalados chegam ao
// transporte externo.
type AdapterRegistry struct {
	adapters map[string]ProviderAdapter
}

var compiledAdapters = map[string]struct{}{
	"provider-sim":       {},
	"synthetic-provider": {},
	"rest-json-v1":       {},
}

// ProviderAdapter é a fronteira executável entre um contrato de catálogo e
// uma API externa. O adapter constrói e interpreta HTTP; autenticação,
// capacidade, fencing e custódia continuam sob responsabilidade do Cometa.
// Isso impede que um ID de catálogo seja confundido com transporte disponível
// e permite homologar um contrato REST real sem importar o servidor simulado.
type ProviderAdapter interface {
	BuildSubmitRequest(context.Context, string, dispatch.Command, string, string) (*http.Request, error)
	BuildStatusRequest(context.Context, string, string) (*http.Request, error)
	DecodeResult(int, []byte) (providersim.OperationResult, error)
}

type providerSimAdapter struct{}
type restJSONAdapter struct{}

type restJSONSubmitRequest struct {
	IdempotencyKey string            `json:"idempotency_key"`
	Mode           string            `json:"mode"`
	Input          any               `json:"input"`
	FileRefs       []files.Reference `json:"file_refs,omitempty"`
	CallbackURL    string            `json:"callback_url,omitempty"`
}

func (providerSimAdapter) BuildSubmitRequest(ctx context.Context, baseURL string, cmd dispatch.Command, providerMode, callbackURL string) (*http.Request, error) {
	req := providersim.SubmitRequest{
		ProtocolID:      externalIdempotencyKey(cmd),
		FileRefs:        cmd.FileRefs,
		Mode:            providersim.Mode(providerMode),
		DelayMs:         requestedDelayMs(cmd.RequestBody),
		Fail:            shouldFail(cmd.RequestBody),
		DropAfterEffect: shouldDropAfterEffect(cmd.RequestBody),
		CallbackURL:     callbackURL,
	}
	return newJSONRequest(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/operations", req)
}

func (providerSimAdapter) BuildStatusRequest(ctx context.Context, baseURL, providerRequestID string) (*http.Request, error) {
	return newStatusRequest(ctx, baseURL, providerRequestID)
}

func (providerSimAdapter) DecodeResult(status int, body []byte) (providersim.OperationResult, error) {
	return decodeOperationResult(status, body)
}

// rest-json-v1 é um contrato REST genérico e versionado para homologação de
// provedores que expõem POST/GET JSON. Ele não conhece o provider-sim: envia
// uma chave de idempotência, a entrada transformada e referências de arquivos
// e exige a resposta estrita provider_request_id/status.
func (restJSONAdapter) BuildSubmitRequest(ctx context.Context, baseURL string, cmd dispatch.Command, providerMode, callbackURL string) (*http.Request, error) {
	payload := restJSONSubmitRequest{
		IdempotencyKey: externalIdempotencyKey(cmd),
		Mode:           providerMode,
		Input:          cmd.RequestBody,
		FileRefs:       cmd.FileRefs,
		CallbackURL:    callbackURL,
	}
	return newJSONRequest(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/operations", payload)
}

func (restJSONAdapter) BuildStatusRequest(ctx context.Context, baseURL, providerRequestID string) (*http.Request, error) {
	return newStatusRequest(ctx, baseURL, providerRequestID)
}

func (restJSONAdapter) DecodeResult(status int, body []byte) (providersim.OperationResult, error) {
	return decodeOperationResult(status, body)
}

func newJSONRequest(ctx context.Context, method, endpoint string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("adapter request serialization failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("adapter request invalid")
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func newStatusRequest(ctx context.Context, baseURL, providerRequestID string) (*http.Request, error) {
	if strings.TrimSpace(providerRequestID) == "" {
		return nil, errors.New("provider request ID is required")
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/v1/operations/" + url.PathEscape(providerRequestID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.New("adapter status request invalid")
	}
	return req, nil
}

func decodeOperationResult(status int, body []byte) (providersim.OperationResult, error) {
	if status != http.StatusOK && status != http.StatusAccepted {
		return providersim.OperationResult{}, fmt.Errorf("provider status %d", status)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var result providersim.OperationResult
	if err := decoder.Decode(&result); err != nil || result.ProviderRequestID == "" {
		return providersim.OperationResult{}, errors.New("provider result invalid")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return providersim.OperationResult{}, errors.New("provider result must contain one JSON document")
	}
	if result.Status != "SUCCEEDED" && result.Status != "FAILED" && result.Status != "PENDING" {
		return providersim.OperationResult{}, errors.New("provider result status invalid")
	}
	return result, nil
}

func NewAdapterRegistry(ids ...string) *AdapterRegistry {
	r := &AdapterRegistry{adapters: map[string]ProviderAdapter{}}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if _, compiled := compiledAdapters[id]; !compiled {
			continue
		}
		switch id {
		case "provider-sim", "synthetic-provider":
			r.adapters[id] = providerSimAdapter{}
		case "rest-json-v1":
			r.adapters[id] = restJSONAdapter{}
		}
	}
	return r
}

// AdapterRegistryFromEnv mantém somente adapters explicitamente instalados
// quando COMETA_ADAPTERS existe. Sem configuração, as implementações
// compiladas localmente ficam disponíveis, mas uma identidade desconhecida
// continua falhando fechado.
func AdapterRegistryFromEnv() *AdapterRegistry {
	raw := strings.TrimSpace(os.Getenv("COMETA_ADAPTERS"))
	if raw == "" {
		return NewAdapterRegistry("provider-sim", "synthetic-provider", "rest-json-v1")
	}
	return NewAdapterRegistry(strings.Split(raw, ",")...)
}

func (r *AdapterRegistry) Supports(id string) bool {
	if r == nil {
		return false
	}
	_, ok := r.adapters[strings.TrimSpace(id)]
	return ok
}

func (r *AdapterRegistry) Get(id string) (ProviderAdapter, error) {
	if r == nil {
		return nil, ErrAdapterUnavailable
	}
	adapter, ok := r.adapters[strings.TrimSpace(id)]
	if !ok {
		return nil, ErrAdapterUnavailable
	}
	return adapter, nil
}
