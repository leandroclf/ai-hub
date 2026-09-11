package pulsar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
)

func TestDeliveryAdminRequiresInteractiveRole(t *testing.T) {
	base := auth.Principal{
		Subject:   "operator",
		TenantID:  "acme",
		MFA:       true,
		Scopes:    []string{"deliveries:read"},
		ExpiresAt: time.Now().Add(time.Hour),
	}
	tests := map[string]auth.Principal{
		"sem papel": base,
		"sem mfa": func() auth.Principal {
			p := base
			p.MFA = false
			p.Roles = []string{"tenant_reader"}
			return p
		}(),
		"workload": func() auth.Principal {
			p := base
			p.Workload = true
			p.CellID = "r2-cell-a"
			p.Roles = []string{"workload_pulsar"}
			return p
		}(),
	}
	for name, principal := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/v1/deliveries?tenant_id=acme", nil).WithContext(auth.WithPrincipal(context.Background(), principal))
			response := httptest.NewRecorder()
			_, _, ok := adminScope(response, req)
			if ok || response.Code != http.StatusForbidden {
				t.Fatalf("identidade administrativa indevidamente aceita: ok=%v status=%d body=%s", ok, response.Code, response.Body.String())
			}
		})
	}

	valid := base
	valid.Roles = []string{"tenant_reader"}
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/deliveries?tenant_id=acme", nil).WithContext(auth.WithPrincipal(context.Background(), valid))
	response := httptest.NewRecorder()
	if _, _, ok := adminScope(response, req); !ok || response.Code != http.StatusOK {
		t.Fatalf("sessão administrativa válida recusada: ok=%v status=%d body=%s", ok, response.Code, response.Body.String())
	}

	reader := valid
	reader.Scopes = []string{"deliveries:write"}
	req = httptest.NewRequest(http.MethodPost, "/admin/v1/deliveries/id/redeliver?tenant_id=acme", nil).WithContext(auth.WithPrincipal(context.Background(), reader))
	response = httptest.NewRecorder()
	if _, _, ok := adminScope(response, req); ok || response.Code != http.StatusForbidden {
		t.Fatalf("leitor administrativo recebeu escrita: ok=%v status=%d body=%s", ok, response.Code, response.Body.String())
	}

	operator := valid
	operator.Roles = []string{"tenant_operator"}
	operator.Scopes = []string{"deliveries:write"}
	req = httptest.NewRequest(http.MethodPost, "/admin/v1/deliveries/id/redeliver?tenant_id=acme", nil).WithContext(auth.WithPrincipal(context.Background(), operator))
	response = httptest.NewRecorder()
	if _, _, ok := adminScope(response, req); !ok || response.Code != http.StatusOK {
		t.Fatalf("operador administrativo válido recusado para escrita: ok=%v status=%d body=%s", ok, response.Code, response.Body.String())
	}
}
