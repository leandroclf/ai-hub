package atlas

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
)

func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func TestDAGContract(t *testing.T) {
	steps := []Step{{ID: "A", ServiceID: "a", ServiceVersion: 1}, {ID: "B", ServiceID: "b", ServiceVersion: 1}, {ID: "C", ServiceID: "c", ServiceVersion: 1, DependsOn: []string{"A", "B"}, InputMapping: map[string]string{"marker": "A.marker"}}}
	layers, err := PlanDAG(steps)
	if err != nil || len(layers) != 2 || len(layers[0]) != 2 || layers[1][0] != "C" {
		t.Fatalf("dependency plan: %v %v", layers, err)
	}
	steps[0].DependsOn = []string{"C"}
	if _, err := PlanDAG(steps); err == nil {
		t.Fatal("cycle accepted")
	}
	steps[0].DependsOn = nil
	steps[2].InputMapping["marker"] = "D.marker"
	if _, err := PlanDAG(steps); err == nil {
		t.Fatal("unrelated output accepted")
	}
	tooMany := make([]Step, 21)
	if _, err := PlanDAG(tooMany); err == nil {
		t.Fatal("fanout accepted")
	}
}
func TestTechnicalProjection(t *testing.T) {
	schema := raw(map[string]any{"type": "object", "properties": map[string]any{"resultado": map[string]string{"type": "string"}}, "required": []string{"resultado"}})
	out, err := TransformJSON(raw(map[string]string{"marker": "PROVIDER_ACTUAL", "private": "excluded"}), map[string]string{"resultado": "marker"}, schema)
	if err != nil || string(out) != `{"resultado":"PROVIDER_ACTUAL"}` {
		t.Fatalf("output %s %v", out, err)
	}
	if _, err := TransformJSON(raw(map[string]string{}), map[string]string{"resultado": "absent"}, schema); err == nil {
		t.Fatal("missing source accepted")
	}
	if _, err := TransformJSON(raw(map[string]string{"marker": "value"}), map[string]string{"resultado": "https://untrusted"}, schema); err == nil {
		t.Fatal("arbitrary mapping accepted")
	}
}

func TestTechnicalProjectionPreservesExactNumbersAndEnums(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"id":{"type":"integer"},"status":{"type":"string","enum":["OK"]}}}`)
	input := json.RawMessage(`{"id":9007199254740993,"status":"OK"}`)
	out, err := TransformJSON(input, nil, schema)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(input, out) {
		t.Fatalf("representação numérica alterada: input=%s output=%s", input, out)
	}
	if _, err := TransformJSON(json.RawMessage(`{"id":1,"status":"INVALID"}`), nil, schema); err == nil {
		t.Fatal("valor fora do enum aceito")
	}
	nested := json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"meta":{"type":"object","additionalProperties":false,"properties":{"code":{"type":"string"}}}}}`)
	if _, err := TransformJSON(json.RawMessage(`{"meta":{"code":"ok","extra":true}}`), nil, nested); err == nil {
		t.Fatal("campo adicional aninhado aceito")
	}
}

func TestTechnicalProjectionAcceptsOpenObjectRuntimeSchema(t *testing.T) {
	schema := json.RawMessage(`{"type":"object"}`)
	input := json.RawMessage(`{"provider_request_id":"provider-correlation","status":"SUCCEEDED","result":{"marker":"actual-provider"}}`)
	out, err := TransformJSON(input, nil, schema)
	if err != nil {
		t.Fatalf("schema de objeto aberto rejeitado no runtime: %v", err)
	}
	if !bytes.Equal(input, out) {
		t.Fatalf("projeção de objeto aberto alterou bytes: input=%s output=%s", input, out)
	}
}

func TestTechnicalProjectionRejectsInvalidJSONDocumentsAndTypes(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"count":{"type":"integer"}},"required":["name"]}`)
	for name, input := range map[string]json.RawMessage{
		"null as string":        json.RawMessage(`{"name":null}`),
		"number as string":      json.RawMessage(`{"name":123}`),
		"string as integer":     json.RawMessage(`{"name":"ok","count":"1"}`),
		"concatenated document": json.RawMessage(`{"name":"ok"}{"name":"later"}`),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := TransformJSON(input, nil, schema); err == nil {
				t.Fatal("documento/tipo inválido aceito")
			}
		})
	}
}

func TestTechnicalProjectionEnforcesPublishedDialectAndBounds(t *testing.T) {
	valid := json.RawMessage(`{"type":"object","properties":{"amount":{"type":"number","minimum":0}},"required":["amount"]}`)
	if _, err := TransformJSON(json.RawMessage(`{"amount":-1}`), nil, valid); err == nil {
		t.Fatal("minimum foi ignorado")
	}
	unsupported := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","minLength":2}}}`)
	if validSchema(unsupported) {
		t.Fatal("keyword não suportada publicada")
	}
	if _, err := TransformJSON(json.RawMessage(`{"name":"x"}`), nil, unsupported); err == nil {
		t.Fatal("keyword não suportada aceita em runtime")
	}
	integer := json.RawMessage(`{"type":"object","properties":{"count":{"type":"integer"}}}`)
	if _, err := TransformJSON(json.RawMessage(`{"count":1.0}`), nil, integer); err != nil {
		t.Fatalf("integer JSON semanticamente exato rejeitado: %v", err)
	}
}

func TestImportSecretSanitization(t *testing.T) {
	input := json.RawMessage(`{"variable":[{"key":"secret","value":"never-retain-me"}],"item":[{"request":{"method":"POST","url":{"raw":"https://name:password@api.example.test/service?token=never-retain-me"},"header":[{"key":"Authorization","value":"Bearer never-retain-me"}],"body":{"raw":"never-retain-me"},"auth":{"type":"bearer","bearer":[{"value":"never-retain-me"}]}}}]}`)
	items, err := SanitizeImport(input)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(raw(items))
	for _, secret := range []string{"never-retain-me", "password", "name:password", "api.example.test"} {
		if strings.Contains(serialized, secret) {
			t.Fatalf("secret retained: %s", secret)
		}
	}
	if len(items) != 1 || items[0].Path != "/service" || items[0].State != "IMPORTED_NOT_EXECUTABLE" {
		t.Fatalf("wrong inventory: %s", serialized)
	}
	again, err := SanitizeImport(input)
	if err != nil || contentHash(items) != contentHash(again) {
		t.Fatal("unstable import identity")
	}
}
func TestCatalogValidation(t *testing.T) {
	r := Resource{Kind: "services", ID: "a", Version: 1, Name: "A", Data: raw(CatalogData{Modes: []string{"SYNC"}, ClientSLASeconds: 10, ProviderSLASeconds: 20, ProviderMode: "async_poll", SyncHTTPBudgetSeconds: 5})}
	v := ValidateResource(r)
	if v.Valid || v.FieldErrors["modes"] == "" || v.FieldErrors["qualification_id"] == "" {
		t.Fatalf("invalid service accepted: %+v", v)
	}
	a := Resource{Data: json.RawMessage(`{"z":1,"a":2}`)}
	b := Resource{Data: json.RawMessage(`{ "a":2,"z":1 }`)}
	if resourceHash(a) != resourceHash(b) {
		t.Fatal("hash depends on json whitespace/order")
	}
	invalidOffer := Resource{Kind: "offers", ID: "offer", Version: 1, TenantID: "acme", Name: "Offer", Data: raw(CatalogData{
		Modes: []string{"ASYNC"}, ClientSLASeconds: 30, FinalizationReserveSeconds: 1,
		ApplicationID: "app", TargetKind: "services", TargetID: "service", TargetVersion: 1,
		TechnicalProfileID: "profile", TechnicalProfileVersion: 1,
		PurchaseContractID: "purchase", PurchaseContractVersion: 1,
		SaleContractID: "sale", SaleContractVersion: 1,
		Routes: []Route{{ProviderAccountID: "account", ProviderAccountVersion: 0, BindingID: "binding", BindingVersion: 1, CapacityDomain: "cell-a"}},
	})}
	if v := ValidateResource(invalidOffer); v.Valid || v.FieldErrors["routes"] == "" {
		t.Fatalf("rota sem versão da conta aceita: %+v", v)
	}
	policySchema := raw(map[string]any{"type": "object", "properties": map[string]any{"marker": map[string]string{"type": "string"}}})
	validPolicy := Resource{Kind: "services", ID: "sla-policy", Version: 1, Name: "SLA policy", Data: raw(CatalogData{Modes: []string{"ASYNC"}, ClientSLASeconds: 30, ProviderSLASeconds: 5, ProviderSLAPolicy: "REJECT_LATE", ProviderMode: "async_poll", AdapterID: "provider-sim", QualificationID: "fixture", DataClass: "SYNTHETIC", InputSchema: policySchema, OutputSchema: policySchema})}
	if v := ValidateResource(validPolicy); !v.Valid {
		t.Fatalf("política de SLA válida rejeitada: %+v", v)
	}
	invalidPolicy := validPolicy
	invalidPolicy.ID = "invalid-sla-policy"
	invalidPolicy.Data = raw(CatalogData{Modes: []string{"ASYNC"}, ClientSLASeconds: 30, ProviderSLASeconds: 5, ProviderSLAPolicy: "RETRY_FOREVER", ProviderMode: "async_poll", AdapterID: "provider-sim", QualificationID: "fixture", DataClass: "SYNTHETIC", InputSchema: policySchema, OutputSchema: policySchema})
	if v := ValidateResource(invalidPolicy); v.Valid || v.FieldErrors["provider_sla_policy"] == "" {
		t.Fatalf("política de SLA inválida aceita: %+v", v)
	}
}

func TestCatalogPublicationRejectsBindingFromAnotherProviderAccount(t *testing.T) {
	s := integrationStore(t)
	ctx := context.Background()
	for _, resource := range []Resource{
		{Kind: "services", ID: "service", Version: 1, TenantID: "acme", Name: "Service", Data: raw(map[string]any{})},
		{Kind: "technical-profiles", ID: "profile", Version: 1, TenantID: "acme", Name: "Profile", Data: raw(map[string]any{})},
		{Kind: "contracts", ID: "purchase", Version: 1, TenantID: "acme", Name: "Purchase", Data: raw(map[string]any{})},
		{Kind: "contracts", ID: "sale", Version: 1, TenantID: "acme", Name: "Sale", Data: raw(map[string]any{})},
		{Kind: "provider-accounts", ID: "account-a", Version: 1, TenantID: "acme", Name: "Account A", Data: raw(map[string]any{})},
		{Kind: "credential-bindings", ID: "binding-b", Version: 1, TenantID: "acme", Name: "Binding B", Data: raw(CatalogData{ProviderAccountID: "account-b", CredentialMode: "SHARED_HUB"})},
	} {
		if _, err := s.SaveResource(ctx, resource, 0, "fixture"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE catalog_resources SET state='PUBLISHED' WHERE kind=$1 AND id=$2 AND version=$3`, resource.Kind, resource.ID, resource.Version); err != nil {
			t.Fatal(err)
		}
	}
	offer := Resource{Kind: "offers", ID: "offer", Version: 1, TenantID: "acme", Name: "Offer", Data: raw(CatalogData{
		Modes: []string{"ASYNC"}, ClientSLASeconds: 30, FinalizationReserveSeconds: 1,
		ApplicationID: "app", TargetKind: "services", TargetID: "service", TargetVersion: 1,
		TechnicalProfileID: "profile", TechnicalProfileVersion: 1,
		PurchaseContractID: "purchase", PurchaseContractVersion: 1,
		SaleContractID: "sale", SaleContractVersion: 1,
		Routes: []Route{{ProviderAccountID: "account-a", ProviderAccountVersion: 1, BindingID: "binding-b", BindingVersion: 1, CapacityDomain: "cell-a"}},
	})}
	v := s.ValidatePublication(ctx, offer)
	if v.Valid || v.FieldErrors["routes/account-a"] == "" {
		t.Fatalf("vínculo de conta diferente aceito: %+v", v)
	}
}

func TestR2Seg05Scenarios(t *testing.T) {
	t.Run("R2-SEG-05-S01_alteracao_atribuivel", func(t *testing.T) {
		s := integrationStore(t)
		ctx := context.Background()
		draft, err := s.SaveResource(ctx, Resource{Kind: "providers", ID: "audited-provider", Version: 1, TenantID: "acme", Name: "Audited provider", Data: raw(map[string]string{"environment": "sandbox"})}, 0, "operator-a")
		if err != nil {
			t.Fatal(err)
		}
		validation := s.ValidatePublication(ctx, draft)
		published, err := s.PublishResource(ctx, draft, draft.Revision, "operator-a", "homologação da conta", validation)
		if err != nil {
			t.Fatal(err)
		}
		if published.State != "PUBLISHED" || published.Revision != draft.Revision+1 || published.Hash == "" {
			t.Fatalf("publicação sem versão/hash atribuíveis: %+v", published)
		}
		var actor, reason, hash string
		var revision int64
		if err := s.db.QueryRowContext(ctx, `SELECT actor,reason,content_hash,revision FROM catalog_publications WHERE resource_id=$1 AND resource_version=$2`, published.ID, published.Version).Scan(&actor, &reason, &hash, &revision); err != nil {
			t.Fatal(err)
		}
		if actor != "operator-a" || reason != "homologação da conta" || hash != published.Hash || revision != published.Revision {
			t.Fatalf("trilha de publicação incompleta: actor=%q reason=%q hash=%q revision=%d", actor, reason, hash, revision)
		}
	})
}

func TestAdminScopeRequiresInteractiveMFA(t *testing.T) {
	ctx := context.Background()
	base := auth.Principal{Subject: "operator", TenantID: "acme", Scopes: []string{"catalog:read"}, Roles: []string{"tenant_reader"}, ExpiresAt: time.Now().Add(time.Hour)}
	for name, principal := range map[string]auth.Principal{
		"sem mfa": base,
		"workload": func() auth.Principal {
			p := base
			p.MFA = true
			p.Workload = true
			p.CellID = "r2-cell-a"
			return p
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/v1/providers?tenant_id=acme", nil).WithContext(auth.WithPrincipal(ctx, principal))
			if tenant, ok := adminScope(req); ok || tenant != "" {
				t.Fatalf("identidade administrativa indevidamente aceita: tenant=%q ok=%v", tenant, ok)
			}
		})
	}

	valid := base
	valid.MFA = true
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/providers?tenant_id=acme", nil).WithContext(auth.WithPrincipal(ctx, valid))
	if tenant, ok := adminScope(req); !ok || tenant != "acme" {
		t.Fatalf("sessão administrativa válida recusada: tenant=%q ok=%v", tenant, ok)
	}
}

func TestCatalogPermissionRequiresCompatibleRole(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		permission string
		want       bool
	}{
		{name: "reader read", role: "tenant_reader", permission: "catalog:read", want: true},
		{name: "reader write", role: "tenant_reader", permission: "catalog:write"},
		{name: "protocol reader publish", role: "hub_protocol_reader", permission: "catalog:publish"},
		{name: "operator integrations write", role: "tenant_operator", permission: "integrations:write", want: true},
		{name: "operator finance write", role: "tenant_operator", permission: "finance:write"},
		{name: "admin finance write", role: "hub_admin", permission: "finance:write", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			principal := auth.Principal{Roles: []string{tt.role}}
			if got := roleAllows(principal, tt.permission); got != tt.want {
				t.Fatalf("role=%s permission=%s: got=%v want=%v", tt.role, tt.permission, got, tt.want)
			}
		})
	}
}

func integrationStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("ATLAS_TEST_DSN")
	if dsn == "" {
		t.Skip("ATLAS_TEST_DSN required for PostgreSQL qualification")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("r2_atlas_%d", time.Now().UnixNano())
	if _, err = db.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); db.Close() })
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	isolated, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { isolated.Close() })
	paths, err := filepath.Glob("../../migrations/control/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = isolated.Exec(string(body)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	return NewStore(isolated)
}
func TestCatalogPostgresConcurrencyPublicationAndIsolation(t *testing.T) {
	s := integrationStore(t)
	ctx := context.Background()
	r := Resource{Kind: "providers", ID: "fixture-provider", Version: 1, TenantID: "acme", Name: "Provider", Data: raw(map[string]string{"environment": "sandbox"})}
	saved, err := s.SaveResource(ctx, r, 0, "operator-a")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	passed, conflicted := 0, 0
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			edited := saved
			edited.Name = fmt.Sprintf("edited-%d", i)
			_, e := s.SaveResource(ctx, edited, 1, fmt.Sprintf("operator-%d", i))
			mu.Lock()
			defer mu.Unlock()
			if e == nil {
				passed++
			} else if errors.Is(e, ErrRevision) {
				conflicted++
			} else {
				t.Errorf("unexpected concurrency: %v", e)
			}
		}(i)
	}
	wg.Wait()
	if passed != 1 || conflicted != 1 {
		t.Fatalf("writes=%d conflicts=%d", passed, conflicted)
	}
	current, err := s.GetResource(ctx, r.Kind, r.ID, r.Version)
	if err != nil {
		t.Fatal(err)
	}
	v := s.ValidatePublication(ctx, current)
	published, err := s.PublishResource(ctx, current, current.Revision, "operator-a", "fixture publication", v)
	if err != nil {
		t.Fatal(err)
	}
	published.Name = "tampered"
	if _, err = s.SaveResource(ctx, published, published.Revision, "operator-b"); !errors.Is(err, ErrConflict) {
		t.Fatalf("published mutation: %v", err)
	}
	if _, err = s.db.Exec(`UPDATE catalog_resources SET data='{"tampered":true}' WHERE id=$1`, r.ID); err == nil {
		t.Fatal("DB allowed published mutation")
	}
	var count int
	if err = s.db.QueryRow(`SELECT count(*) FROM catalog_publications WHERE resource_id=$1`, r.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("publication audit count %d %v", count, err)
	}
	h := NewHandlers(s)
	mux := http.NewServeMux()
	h.Register(mux)
	principal := auth.Principal{Subject: "operator-b", TenantID: "beta", MFA: true, Roles: []string{"tenant_operator"}, Scopes: []string{"catalog:read", "catalog:write", "integrations:read", "integrations:write"}, ExpiresAt: time.Now().Add(time.Hour)}
	req := httptest.NewRequest("GET", "/admin/v1/providers/fixture-provider/1", nil).WithContext(auth.WithPrincipal(ctx, principal))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != 404 || strings.Contains(response.Body.String(), "edited-") {
		t.Fatalf("cross tenant disclosure %d %s", response.Code, response.Body.String())
	}
	req = httptest.NewRequest("GET", "/admin/v1/providers?tenant_id=acme", nil).WithContext(auth.WithPrincipal(ctx, principal))
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != 403 {
		t.Fatalf("forged filter %d", response.Code)
	}
}
func TestCatalogPostgresPaginationAndStaging(t *testing.T) {
	s := integrationStore(t)
	ctx := context.Background()
	for i := 0; i < 37; i++ {
		_, err := s.SaveResource(ctx, Resource{Kind: "providers", ID: fmt.Sprintf("provider-%03d", i), Version: 1, TenantID: "acme", Name: "Synthetic provider", Data: json.RawMessage(`{}`)}, 0, "fixture")
		if err != nil {
			t.Fatal(err)
		}
	}
	a, err := s.ListResources(ctx, "providers", "acme", "Synthetic", "", "", 0, 25)
	if err != nil || len(a) != 25 {
		t.Fatalf("page1 %d %v", len(a), err)
	}
	b, err := s.ListResources(ctx, "providers", "acme", "Synthetic", "", a[24].ID, a[24].Version, 25)
	if err != nil || len(b) != 12 || a[24].ID >= b[0].ID {
		t.Fatalf("page2 %d %v", len(b), err)
	}
	p := auth.Principal{Subject: "operator-a", TenantID: "acme", MFA: true, Roles: []string{"tenant_operator"}, Scopes: []string{"catalog:read", "catalog:write"}, ExpiresAt: time.Now().Add(time.Hour)}
	mux := http.NewServeMux()
	NewHandlers(s).Register(mux)
	body := []byte(`{"collection":{"item":[{"request":{"method":"GET","url":"https://api.example.test/test?token=secret"}}]}}`)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/admin/v1/imports", bytes.NewReader(body)).WithContext(auth.WithPrincipal(ctx, p))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != 201 || strings.Contains(rec.Body.String(), "token=") {
			t.Fatalf("import %d %s", rec.Code, rec.Body.String())
		}
	}
	var n int
	if err = s.db.QueryRow(`SELECT count(*) FROM catalog_imports`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("import duplicated: %d %v", n, err)
	}
	if err = s.db.QueryRow(`SELECT count(*) FROM catalog_resources`).Scan(&n); err != nil || n != 37 {
		t.Fatal("import changed catalog")
	}
}

func TestCatalogImportDiffStates(t *testing.T) {
	s := integrationStore(t)
	ctx := context.Background()
	p := auth.Principal{Subject: "operator-a", TenantID: "acme", MFA: true, Roles: []string{"tenant_operator"}, Scopes: []string{"catalog:read", "catalog:write"}, ExpiresAt: time.Now().Add(time.Hour)}
	mux := http.NewServeMux()
	NewHandlers(s).Register(mux)
	post := func(t *testing.T, body string) ImportBatch {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/admin/v1/imports?tenant_id=acme", strings.NewReader(body)).WithContext(auth.WithPrincipal(ctx, p))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("importação retornou %d: %s", rec.Code, rec.Body.String())
		}
		var batch ImportBatch
		if err := json.Unmarshal(rec.Body.Bytes(), &batch); err != nil {
			t.Fatal(err)
		}
		return batch
	}
	first := post(t, `{"collection":{"item":[{"request":{"method":"GET","url":{"raw":"https://api.example.test/old"},"auth":{"type":"none"}}}]}}`)
	if len(first.Items) != 1 || first.Items[0].Difference != "NEW" {
		t.Fatalf("primeiro inventário não identificado como novo: %+v", first.Items)
	}
	second := post(t, `{"collection":{"item":[{"request":{"method":"GET","url":{"raw":"https://api.example.test/old"},"auth":{"type":"bearer"}}},{"request":{"method":"POST","url":{"raw":"https://api.example.test/new"},"auth":{"type":"none"}}}]}}`)
	byPath := map[string]string{}
	for _, item := range second.Items {
		byPath[item.Path] = item.Difference
	}
	if byPath["/old"] != "CHANGED" || byPath["/new"] != "NEW" {
		t.Fatalf("diff da segunda importação incorreto: %+v", byPath)
	}
	third := post(t, `{"collection":{"item":[{"request":{"method":"GET","url":{"raw":"https://api.example.test/old"},"auth":{"type":"bearer"}}},{"request":{"method":"POST","url":{"raw":"https://api.example.test/new"},"auth":{"type":"none"}}}]}}`)
	for _, item := range third.Items {
		if item.Difference != "UNCHANGED" {
			t.Fatalf("reimportação idêntica não foi marcada como inalterada: %+v", third.Items)
		}
	}
	var resources int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM catalog_resources`).Scan(&resources); err != nil {
		t.Fatal(err)
	}
	if resources != 0 {
		t.Fatalf("importação criou catálogo executável: %d", resources)
	}
}
