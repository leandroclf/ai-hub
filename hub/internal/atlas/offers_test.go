package atlas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
)

func TestOfferResolveRequiresVersionedTarget(t *testing.T) {
	h := NewHandlers(nil)
	mux := http.NewServeMux()
	h.Register(mux)
	principal := auth.Principal{
		Subject:       "operator-a",
		TenantID:      "acme",
		ApplicationID: "app-acme",
		Scopes:        []string{"catalog:read"},
		ExpiresAt:     time.Now().Add(time.Minute),
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/offers/resolve?service_code=consulta-cadastral", nil).WithContext(auth.WithPrincipal(context.Background(), principal))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oferta sem versão retornou %d, esperado 400", rec.Code)
	}
}
