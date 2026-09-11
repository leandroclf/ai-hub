package cometa

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/idgen"
)

func TestPostgresOperationResourceScope(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	t.Setenv("CELL_ID", "scope-cell")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)
	id := idgen.New()
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "scope-tenant", ApplicationID: "scope-app", CellID: "scope-cell", ProviderAccountID: "synthetic"}
	if _, _, err = s.PrepareSubmission(context.Background(), cmd, "secret-binding-do-not-expose", "v1"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	}()
	h := NewHandlers(nil, s)
	principal := auth.Principal{Subject: "workload-fixture", Workload: true, CellID: "scope-cell", Scopes: []string{"protocols:read"}, ExpiresAt: time.Now().Add(time.Minute)}
	for _, tc := range []struct {
		name, tenant, app, cell string
		human                   bool
		want                    int
	}{
		{"authorized", "scope-tenant", "scope-app", "scope-cell", false, 200},
		{"other tenant", "other", "scope-app", "scope-cell", false, 404},
		{"other application", "scope-tenant", "other", "scope-cell", false, 404},
		{"other cell", "scope-tenant", "scope-app", "other", false, 403},
		{"human token", "scope-tenant", "scope-app", "scope-cell", true, 403},
		{"missing scope", "", "", "scope-cell", false, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := principal
			p.CellID = tc.cell
			p.Workload = !tc.human
			r := httptest.NewRequest(http.MethodGet, "/internal/operations/"+id+"?tenant_id="+tc.tenant+"&application_id="+tc.app, nil)
			r = r.WithContext(auth.WithPrincipal(r.Context(), p))
			w := httptest.NewRecorder()
			h.handleGetOperation(w, r)
			if w.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tc.want, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "secret-binding-do-not-expose") {
				t.Fatal("credential metadata exposed")
			}
		})
	}
	t.Log("real PostgreSQL + HTTP handler: workload resource scope enforced for tenant/application/cell; humans denied; binding references omitted")
}

func TestCallbackRejectsNonTerminalObservationBeforeOrphanCustody(t *testing.T) {
	t.Setenv("CALLBACK_INGRESS_KEY", "fixture-ingress")
	h := NewHandlers(nil, nil)
	r := httptest.NewRequest(http.MethodPost, "/callbacks/550e8400-e29b-41d4-a716-446655440000?token=fixture-capability", strings.NewReader(`{"provider_request_id":"provider-correlation","status":"PENDING"}`))
	r.Header.Set("X-Provider-Callback-Key", "fixture-ingress")
	w := httptest.NewRecorder()
	h.handleCallback(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("non-terminal callback entered orphan path: status=%d body=%s", w.Code, w.Body.String())
	}
}
