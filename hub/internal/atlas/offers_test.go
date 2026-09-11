package atlas

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestOfferResolutionBoundsIndexedCandidates(t *testing.T) {
	s := integrationStore(t)
	ctx := context.Background()

	for i := 0; i < 1500; i++ {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO catalog_resources(kind,id,version,tenant_id,name,state,data,author,content_hash)
			VALUES ('offers',$1,1,'acme',$1,'PUBLISHED',$2,'qualification',$3)
		`, fmt.Sprintf("unrelated-offer-%04d", i), raw(map[string]any{
			"application_id": fmt.Sprintf("other-application-%04d", i),
			"target_id":      "growth-service",
			"target_version": "1",
		}), fmt.Sprintf("hash-unrelated-%04d", i))
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"eligible-offer-a", "eligible-offer-b"} {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO catalog_resources(kind,id,version,tenant_id,name,state,data,author,content_hash)
			VALUES ('offers',$1,1,'acme',$1,'PUBLISHED',$2,'qualification',$3)
		`, id, raw(map[string]any{
			"application_id": "growth-application",
			"target_id":      "growth-service",
			"target_version": "1",
			"valid_from":     time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
			"valid_until":    time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
			"routes":         []map[string]string{{"provider_account_id": "account-a"}},
		}), "hash-"+id)
		if err != nil {
			t.Fatal(err)
		}
	}

	candidates, err := s.listEligibleOffers(ctx, "acme", "growth-application", "growth-service", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, candidate.ID)
	}
	if len(candidates) != 2 || ids[0] != "eligible-offer-a" || ids[1] != "eligible-offer-b" {
		t.Fatalf("candidatos elegíveis incorretos ou não limitados: len=%d ids=%v", len(candidates), ids)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "SET LOCAL enable_seqscan=off"); err != nil {
		t.Fatal(err)
	}
	var plan string
	err = tx.QueryRowContext(ctx, `EXPLAIN (FORMAT JSON)
		SELECT id,version FROM (
			SELECT id,version FROM catalog_resources
			WHERE kind='offers' AND state='PUBLISHED' AND tenant_id=$1
			  AND data->>'application_id'=$2 AND data->>'target_id'=$3 AND data->>'target_version'=$4
			UNION ALL
			SELECT id,version FROM catalog_resources
			WHERE kind='offers' AND state='PUBLISHED' AND tenant_id=''
			  AND data->>'application_id'=$2 AND data->>'target_id'=$3 AND data->>'target_version'=$4
		) eligible ORDER BY id,version LIMIT 2`, "acme", "growth-application", "growth-service", "1").Scan(&plan)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plan, "catalog_offer_eligibility_lookup") {
		t.Fatalf("plano não utilizou índice de elegibilidade: %s", plan)
	}
}
