package pulsar

import (
	"net/http"
	"testing"
	"time"
)

func TestWebhookCapacityLeaseAndSignals(t *testing.T) {
	deadline := time.Now().Add(2 * time.Second)
	if capacityLeaseCovers(deadline.Add(500*time.Millisecond), deadline) {
		t.Fatal("lease sem margem deveria ser recusado")
	}
	if !capacityLeaseCovers(deadline.Add(2*time.Second), deadline) {
		t.Fatal("lease com margem deveria ser aceito")
	}
	tests := []struct {
		status int
		signal string
	}{
		{http.StatusNoContent, "SUCCESS"},
		{http.StatusTooManyRequests, "THROTTLED"},
		{http.StatusBadGateway, "UNAVAILABLE"},
	}
	for _, tt := range tests {
		if got := webhookCapacitySignal(tt.status); got != tt.signal {
			t.Errorf("status %d: sinal=%q, esperado=%q", tt.status, got, tt.signal)
		}
	}
}
