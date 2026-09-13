package atlas

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/platform/pg"
	_ "github.com/lib/pq"
)

func TestRequestCapacityKeepsPartialCellOutOfPlacementAndIsIdempotent(t *testing.T) {
	dsn := os.Getenv("R2_CONTROL_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CONTROL_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	tenant := "provisioning-partial-" + idgen.New()
	cell := "provisioning-cell-" + idgen.New()
	resource := Resource{
		Kind:     "clients",
		TenantID: tenant,
		Data:     raw(CatalogData{CellID: cell, CapacityUnits: 2}),
	}
	t.Cleanup(func() {
		_ = pg.WithTenantTx(context.Background(), db, tenant, func(tx *sql.Tx) error {
			_, _ = tx.Exec("DELETE FROM placements WHERE tenant_id=$1", tenant)
			_, _ = tx.Exec("DELETE FROM catalog_provisioning_requests WHERE tenant_id=$1", tenant)
			_, _ = tx.Exec("DELETE FROM catalog_onboardings WHERE tenant_id=$1", tenant)
			return nil
		})
	})

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if err := store.requestCapacity(ctx, resource, "provisioner-test"); err != nil {
			t.Fatalf("requestCapacity attempt %d: %v", i+1, err)
		}
	}

	var state, reason string
	var onboardingCount, requestCount, placementCount int
	if err := pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		if err := tx.QueryRow("SELECT state,reason FROM catalog_onboardings WHERE tenant_id=$1", tenant).Scan(&state, &reason); err != nil {
			return err
		}
		if err := tx.QueryRow("SELECT count(*) FROM catalog_onboardings WHERE tenant_id=$1", tenant).Scan(&onboardingCount); err != nil {
			return err
		}
		if err := tx.QueryRow("SELECT count(*) FROM catalog_provisioning_requests WHERE tenant_id=$1", tenant).Scan(&requestCount); err != nil {
			return err
		}
		return tx.QueryRow("SELECT count(*) FROM placements WHERE tenant_id=$1", tenant).Scan(&placementCount)
	}); err != nil {
		t.Fatal(err)
	}
	if state != "PROVISIONING" || reason != "insufficient-qualified-headroom" {
		t.Fatalf("partial cell was not held out of traffic: state=%s reason=%s", state, reason)
	}
	if onboardingCount != 1 || requestCount != 1 || placementCount != 0 {
		t.Fatalf("reconcile duplicated or assigned partial placement: onboardings=%d requests=%d placements=%d", onboardingCount, requestCount, placementCount)
	}
}
