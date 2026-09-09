package providersim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSubmitIsIdempotentAndExternalEffectCountedOnce(t *testing.T) {
	s := NewServer()
	mux := http.NewServeMux()
	s.Routes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()
	body := `{"protocol_id":"restore-proof-1","mode":"sync"}`
	for i := 0; i < 2; i++ {
		resp, err := http.Post(ts.URL+"/v1/operations", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		var result OperationResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || result.ProviderRequestID != "prov-req-000001" {
			t.Fatalf("attempt %d: status=%d result=%+v", i, resp.StatusCode, result)
		}
	}
	resp, err := http.Get(ts.URL + "/__qualification/effects")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var observed struct {
		Effects   int `json:"effects"`
		Protocols int `json:"protocols"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&observed); err != nil {
		t.Fatal(err)
	}
	if observed.Effects != 1 || observed.Protocols != 1 {
		t.Fatalf("external effect duplicated: %+v", observed)
	}
}
