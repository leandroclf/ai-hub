package orbita

import (
	"encoding/json"
	"testing"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/dispatch"
)

func TestBuildProductPlanCreatesExecutableCommandsPerStep(t *testing.T) {
	serviceSchema := json.RawMessage(`{"type":"object","properties":{"marker":{"type":"string"}},"required":["marker"]}`)
	service := func(id string) atlas.Resource {
		data, _ := json.Marshal(atlas.CatalogData{InputSchema: serviceSchema, OutputSchema: serviceSchema, AdapterID: "adapter-" + id, QualificationID: "qualification-" + id, DataClass: "SYNTHETIC"})
		return atlas.Resource{Kind: "services", ID: id, Version: 1, State: "PUBLISHED", Data: data}
	}
	productData, _ := json.Marshal(atlas.CatalogData{
		MaxParallel: 2, AllowPartial: true, Consolidation: "ALL_REQUIRED", FailurePolicy: "STOP",
		Steps: []atlas.Step{
			{ID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true},
			{ID: "B", ServiceID: "service-b", ServiceVersion: 1, Required: true, DependsOn: []string{"A"}, InputMapping: map[string]string{"marker": "A.marker"}},
		},
	})
	snapshot := atlas.OfferSnapshot{Target: atlas.Resource{Kind: "products", ID: "product-1", Version: 1, Data: productData}, Services: []atlas.Resource{service("service-a"), service("service-b")}}
	base := dispatch.Command{ProtocolID: "00000000-0000-0000-0000-000000000001", TenantID: "tenant", ApplicationID: "app", CellID: "cell", DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"product":"snapshot"}`), RequestBody: json.RawMessage(`{"marker":"input"}`)}
	plan, commands, err := BuildProductPlan(snapshot, base, json.RawMessage(`{"marker":"input"}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 2 || len(commands) != 2 {
		t.Fatalf("plan=%+v commands=%d", plan, len(commands))
	}
	if commands[0].StepID != "A" || commands[0].ServiceCode != "service-a" || commands[0].CommandID == "" {
		t.Fatalf("first command not materialized: %+v", commands[0])
	}
	if commands[1].StepID != "B" || commands[1].ServiceCode != "service-b" || commands[1].CommandID == commands[0].CommandID {
		t.Fatalf("second command identity invalid: %+v", commands[1])
	}
	var child atlas.OfferSnapshot
	if err := json.Unmarshal(commands[0].ConfigSnapshot, &child); err != nil {
		t.Fatal(err)
	}
	if child.Target.Kind != "services" || child.Target.ID != "service-a" {
		t.Fatalf("child snapshot not service scoped: %+v", child.Target)
	}
}
