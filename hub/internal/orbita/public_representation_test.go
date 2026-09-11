package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

// TestPublicLookupUsesLocalPendingAndExpiredRepresentation qualifica
// R2-EXE-08-S03: a consulta pública serve somente a custódia local, devolve
// 200 nos estados contratados e nunca mascara um protocolo pendente ou
// expirado como sucesso.
func TestPublicLookupUsesLocalPendingAndExpiredRepresentation(t *testing.T) {
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
	tenant := "representation-state-" + idgen.New()
	application := "app-representation-state"
	pendingID := idgen.New()
	expiredID := idgen.New()
	insert := func(protocolID, status string, representation []byte) {
		t.Helper()
		_, err = db.Exec(`
			INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,result_version,final_body,final_representation,client_deadline_at)
			VALUES($1,$2,$3,'r2-representation-state',$4,'representation-hash','{}','ASYNC','QUEUED',$5,$6,CASE WHEN $6='EXPIRED' THEN 1 ELSE 0 END,$7::jsonb,$8::bytea,clock_timestamp()+interval '1 hour')
		`, protocolID, tenant, application, idgen.New(), idgen.New(), status, nullableJSON(representation), representation)
		if err != nil {
			t.Fatal(err)
		}
	}
	insert(pendingID, string(StatusWaitingProvider), nil)
	expiredRepresentation := []byte(`{"protocol_id":"` + expiredID + `","status":"EXPIRED","result_version":1,"error_code":"SLA_EXCEEDED"}`)
	insert(expiredID, string(StatusExpired), expiredRepresentation)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id IN ($1,$2)", pendingID, expiredID)
	})

	h := NewHandlers(store, nil, nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)
	principal := auth.Principal{
		Subject:       "representation-reader",
		TenantID:      tenant,
		ApplicationID: application,
		Scopes:        []string{"protocols:read"},
		ExpiresAt:     time.Now().Add(time.Minute),
	}
	get := func(protocolID string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/v1/protocols/"+protocolID, nil).WithContext(auth.WithPrincipal(context.Background(), principal))
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		return res
	}

	pending := get(pendingID)
	if pending.Code != http.StatusOK {
		t.Fatalf("resultado pendente retornou HTTP %d: %s", pending.Code, pending.Body.String())
	}
	var pendingBody map[string]any
	if err := json.Unmarshal(pending.Body.Bytes(), &pendingBody); err != nil {
		t.Fatalf("contrato pendente inválido: %v", err)
	}
	if pendingBody["status"] != string(StatusWaitingProvider) || pendingBody["protocol_id"] != pendingID {
		t.Fatalf("estado pendente não foi preservado: %s", pending.Body.String())
	}
	if strings.Contains(strings.ToUpper(pending.Body.String()), "SUCCEEDED") {
		t.Fatal("consulta pendente mascarou atraso como sucesso")
	}

	expired := get(expiredID)
	if expired.Code != http.StatusOK {
		t.Fatalf("resultado expirado retornou HTTP %d: %s", expired.Code, expired.Body.String())
	}
	if expired.Body.String() != string(expiredRepresentation) {
		t.Fatalf("consulta expirado regenerou representação: got=%s want=%s", expired.Body.String(), expiredRepresentation)
	}
	if expired.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("consulta expirado não preservou media type JSON: %q", expired.Header().Get("Content-Type"))
	}

	t.Log("PostgreSQL + HTTP público: estado pendente foi servido localmente como 200 sem sucesso fictício; estado EXPIRED retornou 200 com os mesmos bytes finais persistidos")
}

func TestPublicLookupReportsAuthorityUnavailableWithout404(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	h := NewHandlers(NewStore(db), nil, nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)
	principal := auth.Principal{
		Subject:       "representation-reader",
		TenantID:      "authority-unavailable",
		ApplicationID: "app-authority-unavailable",
		Scopes:        []string{"protocols:read"},
		ExpiresAt:     time.Now().Add(time.Minute),
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/protocols/00000000-0000-4000-8000-000000000001", nil).WithContext(auth.WithPrincipal(context.Background(), principal))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("autoridade indisponível retornou HTTP %d, esperado 503: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"error":"protocol_unavailable"`) {
		t.Fatalf("erro de autoridade não foi explícito: %s", res.Body.String())
	}
	if strings.Contains(res.Body.String(), `"error":"not_found"`) {
		t.Fatal("falha da autoridade foi convertida em 404 conclusivo")
	}
	t.Log("PostgreSQL indisponível: GET público retornou protocol_unavailable/503, sem falso inexistente")
}

func nullableJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}
