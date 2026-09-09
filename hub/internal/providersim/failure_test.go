package providersim

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDropAfterEffectRecordsExternalEffectBeforeResponseFailure(t *testing.T) {
	s := NewServer()
	body, err := json.Marshal(SubmitRequest{ProtocolID: "protocol-drop-after-effect", Mode: ModeSync, DropAfterEffect: true})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/operations", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	s.handleSubmit(recorder, req)

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.effects != 1 || s.protocols["protocol-drop-after-effect"] == "" {
		t.Fatalf("external effect was not recorded: effects=%d protocols=%v", s.effects, s.protocols)
	}
}
