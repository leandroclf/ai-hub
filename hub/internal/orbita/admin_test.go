package orbita

import (
	"bytes"
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

func TestAdminReconciliationRequestIsDurableAndAudited(t *testing.T) {
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
	tenant, protocolID := "admin-reconciliation-"+idgen.New(), idgen.New()
	if _, err = db.Exec(`INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at) VALUES($1,$2,'app-admin','r2-cell-a',$3,'hash','{}','ASYNC','QUEUED',$4,'RECONCILING',clock_timestamp()+interval '1 hour')`, protocolID, tenant, idgen.New(), idgen.New()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM protocol_access_audit WHERE requested_tenant=$1", tenant)
		db.Exec("DELETE FROM protocol_reconciliation_requests WHERE tenant_id=$1", tenant)
		db.Exec("DELETE FROM protocols WHERE tenant_id=$1", tenant)
	})

	principal := auth.Principal{Subject: "operator-admin", TenantID: tenant, MFA: true, Roles: []string{"hub_protocol_reader"}, Scopes: []string{"protocols:read", "protocols:reconcile"}, ExpiresAt: time.Now().Add(time.Hour)}
	mux := http.NewServeMux()
	NewHandlers(store, nil, nil, nil, nil, nil).RegisterAdmin(mux)
	request := func(reason string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/admin/v1/protocols/"+protocolID+"/reconcile?tenant_id="+tenant, bytes.NewBufferString(`{"reason":"`+reason+`"}`)).WithContext(auth.WithPrincipal(context.Background(), principal))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		return response
	}
	first := request("provider lookup required")
	if first.Code != http.StatusAccepted {
		t.Fatalf("first request status=%d body=%s", first.Code, first.Body.String())
	}
	var firstBody struct {
		RequestID string `json:"request_id"`
		State     string `json:"state"`
		Effect    string `json:"effect"`
	}
	if err = json.Unmarshal(first.Body.Bytes(), &firstBody); err != nil || firstBody.RequestID == "" || firstBody.State != "OPEN" || firstBody.Effect != "no_provider_replay" {
		t.Fatalf("invalid first response: %s (%v)", first.Body.String(), err)
	}
	second := request("provider lookup required again")
	if second.Code != http.StatusAccepted {
		t.Fatalf("second request status=%d body=%s", second.Code, second.Body.String())
	}
	var secondBody struct {
		RequestID string `json:"request_id"`
	}
	if err = json.Unmarshal(second.Body.Bytes(), &secondBody); err != nil || secondBody.RequestID != firstBody.RequestID {
		t.Fatalf("open request was not idempotent: first=%s second=%s", first.Body.String(), second.Body.String())
	}
	var count, audits int
	if err = db.QueryRow("SELECT count(*) FROM protocol_reconciliation_requests WHERE tenant_id=$1 AND protocol_id=$2 AND state='OPEN'", tenant, protocolID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("open requests=%d error=%v", count, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM protocol_access_audit WHERE requested_tenant=$1 AND resource=$2 AND action='RECONCILIATION_REQUEST'", tenant, protocolID).Scan(&audits); err != nil || audits != 2 {
		t.Fatalf("audit records=%d error=%v", audits, err)
	}

	noMFA := principal
	noMFA.MFA = false
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/protocols/"+protocolID+"/reconcile?tenant_id="+tenant, bytes.NewBufferString(`{"reason":"must be refused"}`)).WithContext(auth.WithPrincipal(context.Background(), noMFA))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("non-MFA reconciliation status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAdminReconciliationRejectsUnknownWithoutProviderCorrelation(t *testing.T) {
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
	tenant, protocolID, operationID := "admin-unknown-no-correlation-"+idgen.New(), idgen.New(), idgen.New()
	if _, err = db.Exec(`INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at) VALUES($1,$2,'app-admin','r2-cell-a',$3,'hash','{}','ASYNC','QUEUED',$4,'RECONCILING',clock_timestamp()+interval '1 hour')`, protocolID, tenant, idgen.New(), idgen.New()); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO operations(operation_id,protocol_id,provider_account_id,tenant_id,application_id,cell_id,command,state) VALUES($1,$2,'provider-account',$3,'app-admin','r2-cell-a','{}','UNKNOWN')`, operationID, protocolID, tenant); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM protocol_access_audit WHERE requested_tenant=$1", tenant)
		db.Exec("DELETE FROM protocol_reconciliation_requests WHERE tenant_id=$1", tenant)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", operationID)
		db.Exec("DELETE FROM protocols WHERE tenant_id=$1", tenant)
	})

	principal := auth.Principal{Subject: "operator-admin", TenantID: tenant, MFA: true, Roles: []string{"hub_protocol_reader"}, Scopes: []string{"protocols:read", "protocols:reconcile"}, ExpiresAt: time.Now().Add(time.Hour)}
	mux := http.NewServeMux()
	NewHandlers(store, nil, nil, nil, nil, nil).RegisterAdmin(mux)
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/protocols/"+protocolID+"/reconcile?tenant_id="+tenant, bytes.NewBufferString(`{"reason":"evidência externa ausente"}`)).WithContext(auth.WithPrincipal(context.Background(), principal))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusConflict {
		t.Fatalf("missing correlation status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		RequestID string `json:"request_id"`
		State     string `json:"state"`
		Effect    string `json:"effect"`
	}
	if err = json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.RequestID == "" || body.State != "REJECTED" || body.Effect != "no_provider_correlation" {
		t.Fatalf("invalid rejection response: %s (%v)", response.Body.String(), err)
	}
	var state, lastError string
	if err = db.QueryRow("SELECT state,last_error FROM protocol_reconciliation_requests WHERE request_id=$1", body.RequestID).Scan(&state, &lastError); err != nil {
		t.Fatal(err)
	}
	if state != "REJECTED" || lastError == "" {
		t.Fatalf("rejection was not durable: state=%s error=%q", state, lastError)
	}
}

func TestR2Seg03Scenarios(t *testing.T) {
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
	tenantA, tenantB, protocolID := "r2-seg03-a-"+idgen.New(), "r2-seg03-b-"+idgen.New(), idgen.New()
	if _, err = db.Exec(`INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at,final_representation) VALUES($1,$2,'app-seg03','r2-cell-a',$3,'hash','{}','ASYNC','QUEUED',$4,'SUCCEEDED',clock_timestamp()+interval '1 hour','{"status":"SUCCEEDED","safe":"value"}')`, protocolID, tenantB, idgen.New(), idgen.New()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM protocol_access_audit WHERE resource=$1 OR requested_tenant IN ($2,$3)", protocolID, tenantA, tenantB)
		_, _ = db.Exec("DELETE FROM protocols WHERE protocol_id=$1", protocolID)
	})

	principal := auth.Principal{Subject: "developer-seg03", TenantID: tenantA, MFA: true, Roles: []string{"hub_protocol_reader"}, Scopes: []string{"protocols:read", "admin:cross_tenant"}, ExpiresAt: time.Now().Add(time.Hour)}
	mux := http.NewServeMux()
	NewHandlers(store, nil, nil, nil, nil, nil).RegisterAdmin(mux)
	request := func(query string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/admin/v1/protocols/"+protocolID+"?tenant_id="+tenantB+query, nil).WithContext(auth.WithPrincipal(context.Background(), principal))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		return response
	}

	t.Run("R2-SEG-03-S01_diagnostico_de_desenvolvedor", func(t *testing.T) {
		response := request("&reason=triagem+de+incidente")
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), protocolID) {
			t.Fatalf("diagnóstico cross-tenant recusado: status=%d body=%s", response.Code, response.Body.String())
		}
		var reason string
		if err := db.QueryRow("SELECT reason FROM protocol_access_audit WHERE subject=$1 AND resource=$2 ORDER BY id DESC LIMIT 1", principal.Subject, protocolID).Scan(&reason); err != nil {
			t.Fatal(err)
		}
		if reason != "triagem de incidente" {
			t.Fatalf("justificativa não auditada: %q", reason)
		}
	})

	t.Run("R2-SEG-03-S02_sem_elevacao_implicita", func(t *testing.T) {
		response := request("")
		if response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("leitura sem justificativa status=%d body=%s", response.Code, response.Body.String())
		}
		noWrite := principal
		noWrite.Scopes = []string{"protocols:read", "admin:cross_tenant"}
		post := httptest.NewRequest(http.MethodPost, "/admin/v1/protocols/"+protocolID+"/reconcile?tenant_id="+tenantB, strings.NewReader(`{"reason":"não deve elevar"}`)).WithContext(auth.WithPrincipal(context.Background(), noWrite))
		record := httptest.NewRecorder()
		mux.ServeHTTP(record, post)
		if record.Code != http.StatusForbidden {
			t.Fatalf("leitor administrativo recebeu escrita: status=%d body=%s", record.Code, record.Body.String())
		}
	})

	t.Run("R2-SEG-03-S03_auditoria_indisponivel", func(t *testing.T) {
		unavailable, err := sql.Open("postgres", "postgres://hub:r2-local-fixture@127.0.0.1:1/hub_core?sslmode=disable&connect_timeout=1")
		if err != nil {
			t.Fatal(err)
		}
		defer unavailable.Close()
		unavailableMux := http.NewServeMux()
		NewHandlers(NewStore(unavailable), nil, nil, nil, nil, nil).RegisterAdmin(unavailableMux)
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/admin/v1/protocols/"+protocolID+"?tenant_id="+tenantB+"&reason=incidente+testado", nil).WithContext(auth.WithPrincipal(context.Background(), principal))
		unavailableMux.ServeHTTP(response, req)
		if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "audit_unavailable") || strings.Contains(response.Body.String(), dsn) {
			t.Fatalf("falha de auditoria não sanitizada: status=%d body=%s", response.Code, response.Body.String())
		}
	})
}
