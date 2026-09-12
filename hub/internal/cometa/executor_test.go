package cometa

import (
	"log/slog"
	"testing"

	"ai-hub/hub/internal/atlasclient"
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

func TestCallbackURLUsesConfiguredPublicGatewayWithoutCapabilityQuery(t *testing.T) {
	t.Setenv("CALLBACK_PUBLIC_URL", "http://kong:8000")
	e := NewExecutor(nil, (*atlasclient.Client)(nil), slog.Default(), "http://cometa:8082", nil)
	if got := e.callbackURLFor("operation-1"); got != "http://kong:8000/callbacks/operation-1" {
		t.Fatalf("callback deve usar o gateway público sem capability na query: %q", got)
	}
}
