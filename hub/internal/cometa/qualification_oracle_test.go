package cometa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/dispatch"
)

// TestQualificationEchoRegressionOracle comprova que a qualificação de
// contrato não pode aceitar somente SUCCEEDED: o mesmo input ecoado pelo
// conector é rejeitado pelo oráculo de conteúdo em SYNC e ASYNC.
func TestQualificationEchoRegressionOracle(t *testing.T) {
	input := `{"cpf":"input-echo"}`
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"provider_request_id":"echo-1","status":"SUCCEEDED","detail":"{\"cpf\":\"input-echo\"}"}`)
	}))
	defer provider.Close()

	adapter := restJSONAdapter{}
	command := dispatch.Command{ProtocolID: "qualification-echo", RequestBody: map[string]any{"cpf": "input-echo"}}
	contract := atlas.AdapterContract{SubmitPath: "/sync", StatusPath: "/async/{id}"}
	for _, mode := range []string{"sync", "async_poll"} {
		t.Run(mode, func(t *testing.T) {
			req, err := adapter.BuildSubmitRequest(context.Background(), provider.URL, command, contract, mode, "")
			if err != nil {
				t.Fatal(err)
			}
			response, err := provider.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, _ := io.ReadAll(response.Body)
			result, err := adapter.DecodeResult(response.StatusCode, body)
			if err != nil || result.Status != "SUCCEEDED" {
				t.Fatalf("fixture deveria simular sucesso superficial: status=%d result=%+v err=%v", response.StatusCode, result, err)
			}
			if strings.TrimSpace(result.Detail) == input {
				t.Logf("oráculo rejeitaria resposta ecoada em %s apesar de status=%s", mode, result.Status)
				return
			}
			t.Fatalf("oráculo não detectou eco: detail=%q input=%q", result.Detail, input)
		})
	}
}

func TestQualificationEchoRegressionFixtureIsIndependentJSON(t *testing.T) {
	var body map[string]string
	if err := json.Unmarshal([]byte(`{"cpf":"input-echo"}`), &body); err != nil || body["cpf"] != "input-echo" {
		t.Fatalf("fixture de input inválida: %+v %v", body, err)
	}
}
