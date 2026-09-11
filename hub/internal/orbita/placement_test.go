package orbita

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

// TestPublicProtocolLookupSurvivesTenantCellRelocation qualifica OPE-05-S01:
// depois de a autoridade do tenant mudar de A para B, o cliente ainda
// consulta protocolos antigos por UUID, enquanto uma rota interna não pode
// cruzar o fence da célula que possui o protocolo.
func TestPublicProtocolLookupSurvivesTenantCellRelocation(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	t.Setenv("CELL_ID", "r2-cell-b")
	tenant := "placement-relocation-" + idgen.New()
	otherTenant := "placement-other-" + idgen.New()
	oldProtocol := idgen.New()
	newProtocol := idgen.New()
	foreignProtocol := idgen.New()
	insert := func(protocolID, protocolTenant, cell, key string) {
		t.Helper()
		_, err = db.Exec(`
			INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at)
			VALUES($1,$2,'app-relocation',$3,$4,'placement-hash','{}','ASYNC','QUEUED',$5,'SUCCEEDED',clock_timestamp()+interval '1 hour')
		`, protocolID, protocolTenant, cell, key, idgen.New())
		if err != nil {
			t.Fatal(err)
		}
	}
	insert(oldProtocol, tenant, "r2-cell-a", "old")
	insert(newProtocol, tenant, "r2-cell-b", "new")
	insert(foreignProtocol, otherTenant, "r2-cell-b", "foreign")
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id IN ($1,$2,$3)", oldProtocol, newProtocol, foreignProtocol)
	})

	h := NewHandlers(store, nil, nil, nil, nil, nil)
	publicMux := http.NewServeMux()
	h.Register(publicMux)
	principal := auth.Principal{
		Subject:       "relocated-client",
		TenantID:      tenant,
		ApplicationID: "app-relocation",
		CellID:        "r2-cell-b",
		MFA:           true,
		Scopes:        []string{"protocols:read"},
		ExpiresAt:     time.Now().Add(time.Hour),
	}
	getPublic := func(protocolID string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/v1/protocols/"+protocolID, nil).
			WithContext(auth.WithPrincipal(context.Background(), principal))
		res := httptest.NewRecorder()
		publicMux.ServeHTTP(res, req)
		return res
	}
	if res := getPublic(oldProtocol); res.Code != http.StatusOK {
		t.Fatalf("protocolo antigo na célula A não foi resolvido após realocação para B: status=%d body=%s", res.Code, res.Body.String())
	}
	if res := getPublic(newProtocol); res.Code != http.StatusOK {
		t.Fatalf("protocolo novo na célula B não foi resolvido: status=%d body=%s", res.Code, res.Body.String())
	}
	if res := getPublic(foreignProtocol); res.Code != http.StatusNotFound {
		t.Fatalf("protocolo de outro tenant vazou após realocação: status=%d body=%s", res.Code, res.Body.String())
	}

	internalMux := http.NewServeMux()
	h.RegisterInternal(internalMux)
	workload := auth.Principal{
		Subject:   "orbita-cell-b",
		CellID:    "r2-cell-b",
		Workload:  true,
		Scopes:    []string{"protocols:read"},
		ExpiresAt: time.Now().Add(time.Hour),
	}
	req := httptest.NewRequest(http.MethodGet, "/internal/protocols/"+oldProtocol, nil).
		WithContext(auth.WithPrincipal(context.Background(), workload))
	res := httptest.NewRecorder()
	internalMux.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("workload da célula B cruzou fence do protocolo da célula A: status=%d body=%s", res.Code, res.Body.String())
	}
}
