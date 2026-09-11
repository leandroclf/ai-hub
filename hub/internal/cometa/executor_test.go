package cometa

import (
	"testing"

	"ai-hub/hub/internal/dispatch"
)

func TestExternalIdempotencyKeySeparatesProductSteps(t *testing.T) {
	base := dispatch.Command{ProtocolID: "protocol-1", CommandID: "command-1", StepID: "protocol-1"}
	if got := externalIdempotencyKey(base); got != "protocol-1" {
		t.Fatalf("operação simples deve usar protocolo: %q", got)
	}
	step := base
	step.CommandID = "command-step-a"
	step.StepID = "step_a"
	if got := externalIdempotencyKey(step); got != "command-step-a" {
		t.Fatalf("etapa de produto deve usar comando estável: %q", got)
	}
}
