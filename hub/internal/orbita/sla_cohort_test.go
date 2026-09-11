package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

func TestAdminSLAReportSeparatesOpenFulfilledExpiredAndExcludedCohort(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tenant := "sla-cohort-" + idgen.New()
	config := json.RawMessage(`{"target":{"data":{"provider_sla_seconds":10,"provider_sla_policy":"MONITOR_ONLY"}}}`)
	statuses := []string{"WAITING_PROVIDER", "SUCCEEDED", "EXPIRED", "CANCELLED"}
	protocols := make([]string, 0, len(statuses))
	for index, status := range statuses {
		protocolID := idgen.New()
		protocols = append(protocols, protocolID)
		finalized := ""
		if status != "WAITING_PROVIDER" {
			finalized = ",finalized_at=clock_timestamp()"
		}
		query := `INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at,config_snapshot` + finalized + `)
			VALUES($1,$2,'sla-cohort-app','sla-cohort-cell',$3,$4,'{}','ASYNC','QUEUED',$5,$6,clock_timestamp()+interval '1 hour',$7` + func() string {
			if finalized != "" {
				return ""
			}
			return ""
		}() + `)`
		// Keep the fixture explicit and independent of finalization side effects.
		if finalized != "" {
			query = `INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at,config_snapshot,finalized_at)
				VALUES($1,$2,'sla-cohort-app','sla-cohort-cell',$3,$4,'{}','ASYNC','QUEUED',$5,$6,clock_timestamp()+interval '1 hour',$7,clock_timestamp())`
		} else {
			query = `INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at,config_snapshot)
				VALUES($1,$2,'sla-cohort-app','sla-cohort-cell',$3,$4,'{}','ASYNC','QUEUED',$5,$6,clock_timestamp()+interval '1 hour',$7)`
		}
		if _, err := db.Exec(query, protocolID, tenant, "cohort-key-"+idgen.New(), "cohort-hash-"+idgen.New(), idgen.New(), status, config); err != nil {
			t.Fatalf("inserir estado %d/%s: %v", index, status, err)
		}
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM protocol_access_audit WHERE requested_tenant=$1", tenant)
		db.Exec("DELETE FROM protocols WHERE tenant_id=$1", tenant)
	})

	principal := auth.Principal{Subject: "sla-cohort-reader", TenantID: tenant, MFA: true, Roles: []string{"hub_protocol_reader"}, Scopes: []string{"protocols:read"}, ExpiresAt: time.Now().Add(time.Hour)}
	mux := http.NewServeMux()
	NewHandlers(NewStore(db), nil, nil, nil, nil, nil).RegisterAdmin(mux)
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/sla-reports?tenant_id="+tenant+"&limit=100", nil).WithContext(auth.WithPrincipal(context.Background(), principal))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("relatório status=%d corpo=%s", response.Code, response.Body.String())
	}
	var report struct {
		Items  []map[string]any `json:"items"`
		Cohort struct {
			Eligible  int `json:"eligible"`
			Open      int `json:"open"`
			Fulfilled int `json:"fulfilled"`
			Expired   int `json:"expired"`
			Excluded  int `json:"excluded"`
		} `json:"cohort"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Items) != 4 || report.Cohort.Eligible != 4 || report.Cohort.Open != 1 || report.Cohort.Fulfilled != 1 || report.Cohort.Expired != 1 || report.Cohort.Excluded != 1 {
		t.Fatalf("coorte não separou estados: items=%d cohort=%+v body=%s", len(report.Items), report.Cohort, response.Body.String())
	}
	if len(protocols) != 4 {
		t.Fatal("fixture de protocolos incompleta")
	}
	t.Logf("painel de SLA separou elegíveis=%d, abertos=%d, cumpridos=%d, vencidos=%d e exclusões=%d sem contar aberto como sucesso", report.Cohort.Eligible, report.Cohort.Open, report.Cohort.Fulfilled, report.Cohort.Expired, report.Cohort.Excluded)
}
