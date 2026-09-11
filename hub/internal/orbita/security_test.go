package orbita

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
)

func TestPublicAdmissionDistinguishesAuthenticationFromAuthorization(t *testing.T) {
	h := NewHandlers(nil, nil, nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	unauthenticated := httptest.NewRequest(http.MethodPost, "/v1/protocols", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, unauthenticated)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("identidade ausente retornou %d, esperado 401", rec.Code)
	}

	principal := auth.Principal{
		Subject:       "tenant-reader",
		TenantID:      "acme",
		ApplicationID: "app-acme",
		MFA:           true,
		Scopes:        []string{"protocols:read"},
		ExpiresAt:     time.Now().Add(time.Minute),
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/protocols", nil).WithContext(auth.WithPrincipal(context.Background(), principal))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("token válido sem escopo retornou %d, esperado 403", rec.Code)
	}
}
