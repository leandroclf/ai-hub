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
	stepOffer := func(account string) atlas.StepOffer {
		return atlas.StepOffer{
			Account:       atlas.Resource{Kind: "provider-accounts", ID: account, Version: 1, State: "PUBLISHED", Data: json.RawMessage(`{}`)},
			Binding:       atlas.Resource{Kind: "credential-bindings", ID: "binding-" + account, Version: 1, State: "PUBLISHED", Data: json.RawMessage(`{"provider_account_id":"` + account + `"}`)},
			SelectedRoute: atlas.Route{ProviderAccountID: account, ProviderAccountVersion: 1, BindingID: "binding-" + account, BindingVersion: 1},
		}
	}
	snapshot := atlas.OfferSnapshot{
		Target: atlas.Resource{Kind: "products", ID: "product-1", Version: 1, Data: productData}, Services: []atlas.Resource{service("service-a"), service("service-b")},
		StepOffers: map[string]atlas.StepOffer{"A": stepOffer("account-a"), "B": stepOffer("account-b")},
	}
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
	var childA, childB atlas.OfferSnapshot
	if err := json.Unmarshal(commands[0].ConfigSnapshot, &childA); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(commands[1].ConfigSnapshot, &childB); err != nil {
		t.Fatal(err)
	}
	if childA.Target.Kind != "services" || childA.Target.ID != "service-a" {
		t.Fatalf("child snapshot not service scoped: %+v", childA.Target)
	}
	// R6-EXE-01: cada etapa carrega SUA PRÓPRIA conta/rota/vínculo, nunca a
	// da etapa vizinha nem a do produto — o probe original reproduziu rota B
	// com conta A.
	if childA.Account.ID != "account-a" || childA.SelectedRoute.ProviderAccountID != "account-a" || commands[0].ProviderAccountID != "account-a" {
		t.Fatalf("etapa A não usou sua própria conta: snapshot=%+v command=%+v", childA.Account, commands[0].ProviderAccountID)
	}
	if childB.Account.ID != "account-b" || childB.SelectedRoute.ProviderAccountID != "account-b" || commands[1].ProviderAccountID != "account-b" {
		t.Fatalf("etapa B não usou sua própria conta: snapshot=%+v command=%+v", childB.Account, commands[1].ProviderAccountID)
	}
}

func TestBuildProductPlanRejectsStepWithoutOwnResolvedRoute(t *testing.T) {
	serviceSchema := json.RawMessage(`{"type":"object"}`)
	data, _ := json.Marshal(atlas.CatalogData{InputSchema: serviceSchema, OutputSchema: serviceSchema, AdapterID: "adapter-a", DataClass: "SYNTHETIC"})
	service := atlas.Resource{Kind: "services", ID: "service-a", Version: 1, State: "PUBLISHED", Data: data}
	productData, _ := json.Marshal(atlas.CatalogData{
		MaxParallel: 2, Consolidation: "ALL_REQUIRED", FailurePolicy: "STOP",
		Steps: []atlas.Step{{ID: "A", ServiceID: "service-a", ServiceVersion: 1, Required: true}},
	})
	// No StepOffers at all — a snapshot resolved before R6-EXE-01, or a
	// step Atlas genuinely could not resolve a route for.
	snapshot := atlas.OfferSnapshot{Target: atlas.Resource{Kind: "products", ID: "product-1", Version: 1, Data: productData}, Services: []atlas.Resource{service}}
	base := dispatch.Command{ProtocolID: "00000000-0000-0000-0000-000000000002", TenantID: "tenant", ApplicationID: "app", CellID: "cell", DispatchMode: dispatch.DispatchQueued, ConfigSnapshot: json.RawMessage(`{"product":"snapshot"}`), RequestBody: json.RawMessage(`{"marker":"input"}`)}
	if _, _, err := BuildProductPlan(snapshot, base, json.RawMessage(`{"marker":"input"}`)); err == nil {
		t.Fatal("admission accepted a step without its own resolved account/binding")
	}
}
