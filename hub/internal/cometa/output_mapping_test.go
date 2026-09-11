package cometa

import (
	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/providersim"
	"bytes"
	"encoding/json"
	"testing"
)

func TestNormalizeProviderResultUsesFrozenOutputMapping(t *testing.T) {
	snapshot := atlas.OfferSnapshot{
		Target: atlas.Resource{Data: json.RawMessage(`{
			"output_schema":{"type":"object","properties":{"marker":{"type":"string"}},"required":["marker"]}
		}`)},
		TechnicalProfile: atlas.Resource{Data: json.RawMessage(`{
			"output_mapping":{"marker":"detail"}
		}`)},
	}
	got, err := normalizeProviderResult(snapshot, providersim.OperationResult{
		ProviderRequestID: "provider-123",
		Status:            "SUCCEEDED",
		Detail:            "resultado-normalizado",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := json.RawMessage(`{"marker":"resultado-normalizado"}`)
	if !bytes.Equal(got, want) {
		t.Fatalf("resultado não foi projetado pelo perfil: got=%s want=%s", got, want)
	}
	var body map[string]any
	if err := json.Unmarshal(got, &body); err != nil || body["provider_request_id"] != nil {
		t.Fatalf("metadado operacional vazou na representação pública: %s", got)
	}
}

func TestNormalizeProviderResultRejectsOutputContractWithoutMapping(t *testing.T) {
	snapshot := atlas.OfferSnapshot{
		Target: atlas.Resource{Data: json.RawMessage(`{
			"output_schema":{"type":"object","properties":{"required_marker":{"type":"string"}},"required":["required_marker"]}
		}`)},
	}
	if _, err := normalizeProviderResult(snapshot, providersim.OperationResult{Status: "SUCCEEDED"}); err == nil {
		t.Fatal("resultado sem o campo exigido foi aceito")
	}
}
