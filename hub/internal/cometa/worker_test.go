package cometa

import (
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
)

func TestCommandResultQuarantineReason(t *testing.T) {
	tests := []struct {
		name   string
		result dispatch.Result
		want   string
	}{
		{
			name:   "rejeicao deterministica",
			result: dispatch.Result{Kind: dispatch.FactRejected, ErrorCode: "invalid_snapshot"},
			want:   "rejected_command:invalid_snapshot",
		},
		{
			name:   "adapter nao homologado",
			result: dispatch.Result{Kind: dispatch.FactRejected, ErrorCode: "adapter_not_qualified"},
			want:   "rejected_command:adapter_not_qualified",
		},
		{
			name:   "capacidade permanece redeliverable",
			result: dispatch.Result{Kind: dispatch.FactRejected, ErrorCode: "capacity_unavailable"},
		},
		{
			name:   "custodia incerta permanece redeliverable",
			result: dispatch.Result{Kind: dispatch.FactUnknown, ErrorCode: "custody_unavailable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandResultQuarantineReason(tt.result); got != tt.want {
				t.Fatalf("quarantine reason = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommandDeadlineExpired(t *testing.T) {
	now := time.Date(2026, time.September, 11, 6, 0, 0, 0, time.UTC)
	if !commandDeadlineExpired(dispatch.Command{StepDeadline: now.Add(-time.Second)}, now) {
		t.Fatal("expired command deadline was not detected")
	}
	if commandDeadlineExpired(dispatch.Command{StepDeadline: now.Add(time.Second)}, now) {
		t.Fatal("future command deadline was classified as expired")
	}
	if commandDeadlineExpired(dispatch.Command{}, now) {
		t.Fatal("zero command deadline was classified as expired")
	}
}
