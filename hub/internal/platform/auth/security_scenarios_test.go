package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestR2Seg01Scenarios(t *testing.T) {
	t.Run("R2-SEG-01-S01_cliente_autorizado", func(t *testing.T) {
		principal := Principal{
			Subject:       "client-a",
			TenantID:      "tenant-a",
			ApplicationID: "application-a",
			Scopes:        []string{"protocols:read", "protocols:write"},
			ExpiresAt:     time.Now().Add(time.Minute),
		}
		ctx := WithPrincipal(context.Background(), principal)
		if !Authorize(ctx, "protocols:read", "tenant-a") || !Authorize(ctx, "protocols:write", "tenant-a") {
			t.Fatal("cliente autorizado foi recusado")
		}
		if tenant, ok := PublicTenant(ctx, "protocols:read"); !ok || tenant != "tenant-a" {
			t.Fatalf("contexto público inválido: tenant=%q ok=%v", tenant, ok)
		}
		if Authorize(ctx, "protocols:read", "tenant-b") {
			t.Fatal("autorização ultrapassou o tenant concedido")
		}
	})

	t.Run("R2-SEG-01-S02_tentativa_entre_tenants", func(t *testing.T) {
		verifier, sign := fixture(t)
		called := false
		handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Tenant-Id") != "" || r.Header.Get("X-Application-Id") != "" {
				t.Fatal("header fornecido pelo cliente permaneceu confiável")
			}
			if Authorize(r.Context(), "protocols:read", "tenant-b") {
				called = true
				t.Fatal("filtro cross-tenant autorizou recurso alheio")
			}
			Error(w, http.StatusNotFound, "not_found")
		}))
		request := httptest.NewRequest(http.MethodGet, "/v1/protocols/protocol-b?tenant_id=tenant-b", nil)
		request.Header.Set("Authorization", "Bearer "+sign(validClaims()))
		request.Header.Set("X-Tenant-Id", "tenant-b")
		request.Header.Set("X-Application-Id", "application-b")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound || called {
			t.Fatalf("recurso cross-tenant não foi ocultado: status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("R2-SEG-01-S03_token_invalido", func(t *testing.T) {
		verifier, sign := fixture(t)
		effects := 0
		handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !Authorize(r.Context(), "protocols:write", "tenant-a") {
				Error(w, http.StatusForbidden, "forbidden")
				return
			}
			effects++
			w.WriteHeader(http.StatusNoContent)
		}))

		expired := validClaims()
		expired["exp"] = time.Now().Add(-time.Minute).Unix()
		request := httptest.NewRequest(http.MethodPost, "/v1/protocols", nil)
		request.Header.Set("Authorization", "Bearer "+sign(expired))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized || effects != 0 {
			t.Fatalf("token expirado produziu efeito: status=%d effects=%d", response.Code, effects)
		}

		withoutScope := validClaims()
		withoutScope["scope"] = "protocols:read"
		request = httptest.NewRequest(http.MethodPost, "/v1/protocols", nil)
		request.Header.Set("Authorization", "Bearer "+sign(withoutScope))
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden || effects != 0 {
			t.Fatalf("token sem escopo não recebeu 403 antes do efeito: status=%d effects=%d", response.Code, effects)
		}
	})
}

func TestR2Seg02Scenarios(t *testing.T) {
	t.Setenv("CELL_ID", "cell-a")

	t.Run("R2-SEG-02-S01_chamada_interna_permitida", func(t *testing.T) {
		principal := Principal{Subject: "orbita-workload", Workload: true, CellID: "cell-a", Scopes: []string{"cometa:execute"}, ExpiresAt: time.Now().Add(time.Minute)}
		if !Authorize(WithPrincipal(context.Background(), principal), "cometa:execute", "") {
			t.Fatal("workload da mesma célula foi recusado")
		}
	})

	t.Run("R2-SEG-02-S02_bypass_pela_porta_interna", func(t *testing.T) {
		for _, principal := range []Principal{
			{Subject: "human", TenantID: "tenant-a", Scopes: []string{"cometa:execute"}, ExpiresAt: time.Now().Add(time.Minute)},
			{Subject: "foreign-workload", Workload: true, CellID: "cell-b", Scopes: []string{"cometa:execute"}, ExpiresAt: time.Now().Add(time.Minute)},
			{Subject: "wrong-scope", Workload: true, CellID: "cell-a", Scopes: []string{"protocols:read"}, ExpiresAt: time.Now().Add(time.Minute)},
		} {
			if Authorize(WithPrincipal(context.Background(), principal), "cometa:execute", "") {
				t.Fatalf("identidade sem a concessão correta foi aceita: %+v", principal)
			}
		}
	})

	t.Run("R2-SEG-02-S03_pool_reutilizado", func(t *testing.T) {
		first := Principal{Subject: "client-a", TenantID: "tenant-a", ApplicationID: "application-a", Scopes: []string{"protocols:read"}, ExpiresAt: time.Now().Add(time.Minute)}
		second := Principal{Subject: "client-b", TenantID: "tenant-b", ApplicationID: "application-b", Scopes: []string{"protocols:read"}, ExpiresAt: time.Now().Add(time.Minute)}
		firstContext := WithPrincipal(context.Background(), first)
		secondContext := WithPrincipal(context.Background(), second)
		if got, ok := FromContext(secondContext); !ok || got.TenantID != "tenant-b" || got.ApplicationID != "application-b" {
			t.Fatalf("contexto reutilizado conservou identidade anterior: got=%+v ok=%v", got, ok)
		}
		if Authorize(secondContext, "protocols:read", "tenant-a") || !Authorize(firstContext, "protocols:read", "tenant-a") {
			t.Fatal("isolamento de contexto não foi preservado entre requisições")
		}
	})
}
