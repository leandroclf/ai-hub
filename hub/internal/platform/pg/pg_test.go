package pg

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/idgen"
	_ "github.com/lib/pq"
)

func TestRuntimeDSNFixesTenantSessionSetting(t *testing.T) {
	got, err := RuntimeDSN("postgres://runtime:secret@db.example/hub_core?sslmode=require", "tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" || !contains(got, "app.tenant_id%3Dtenant-a") {
		t.Fatalf("tenant session setting missing from DSN: %s", got)
	}
	if _, err = RuntimeDSN("postgres://runtime/db", "bad\nvalue"); err == nil {
		t.Fatal("invalid tenant accepted")
	}
}

func TestWithTenantTxRejectsMissingScopeBeforeOpeningDatabase(t *testing.T) {
	if err := WithTenantTx(nil, nil, "", nil); err == nil {
		t.Fatal("missing tenant accepted")
	}
}

// TestWithTenantTxAndAuditedScopeEnforceRLSAgainstRealPostgres prova R6-SEG-01
// (S01/S03) contra um PostgreSQL real com o role runtime não-superuser: tenant
// A existe e é visível a si mesmo, B existe e é invisível a A, escrita de A em
// B falha, e leitura cruzada só é possível com um motivo auditável (nunca
// escrita cruzada). Não faz ROLLBACK das fixtures antes de afirmar: cada
// asserção roda em sua própria transação committada, contra dado que
// realmente persistiu.
func TestWithTenantTxAndAuditedScopeEnforceRLSAgainstRealPostgres(t *testing.T) {
	dsn := os.Getenv("R2_RUNTIME_TEST_DSN")
	if dsn == "" {
		t.Skip("requires PostgreSQL migrado com hub_runtime em R2_RUNTIME_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	tenantA, tenantB := "pg-rls-a-"+randSuffix(), "pg-rls-b-"+randSuffix()
	t.Cleanup(func() {
		_ = WithAuditedScopeTx(ctx, db, "test-cleanup", func(tx *sql.Tx) error {
			_, _ = tx.ExecContext(ctx, "DELETE FROM webhook_destinations WHERE tenant_id IN ($1,$2)", tenantA, tenantB)
			return nil
		})
	})

	if err := WithTenantTx(ctx, db, tenantA, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO webhook_destinations(tenant_id,url,hmac_secret) VALUES ($1,'https://a.invalid/hook','fixture-a')", tenantA)
		return err
	}); err != nil {
		t.Fatalf("tenant A insert: %v", err)
	}
	if err := WithTenantTx(ctx, db, tenantB, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO webhook_destinations(tenant_id,url,hmac_secret) VALUES ($1,'https://b.invalid/hook','fixture-b')", tenantB)
		return err
	}); err != nil {
		t.Fatalf("tenant B insert: %v", err)
	}

	var ownVisible, crossVisible int
	if err := WithTenantTx(ctx, db, tenantA, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM webhook_destinations WHERE tenant_id=$1", tenantA).Scan(&ownVisible); err != nil {
			return err
		}
		return tx.QueryRowContext(ctx, "SELECT count(*) FROM webhook_destinations WHERE tenant_id=$1", tenantB).Scan(&crossVisible)
	}); err != nil {
		t.Fatal(err)
	}
	if ownVisible != 1 {
		t.Fatalf("tenant A deve ver sua própria fixture: got %d", ownVisible)
	}
	if crossVisible != 0 {
		t.Fatalf("tenant A não deve ver fixture de B: got %d", crossVisible)
	}

	var crossWriteRows int64
	if err := WithTenantTx(ctx, db, tenantA, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE webhook_destinations SET url='https://hijack.invalid' WHERE tenant_id=$1", tenantB)
		if err != nil {
			return err
		}
		crossWriteRows, err = res.RowsAffected()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if crossWriteRows != 0 {
		t.Fatalf("escrita cruzada de A em B deveria afetar 0 linhas, afetou %d", crossWriteRows)
	}

	var auditedVisible int
	if err := WithAuditedScopeTx(ctx, db, "R6-SEG-01 test proof", func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT count(*) FROM webhook_destinations WHERE tenant_id IN ($1,$2)", tenantA, tenantB).Scan(&auditedVisible)
	}); err != nil {
		t.Fatal(err)
	}
	if auditedVisible != 2 {
		t.Fatalf("motivo auditado deve enxergar ambos os tenants: got %d", auditedVisible)
	}

	var auditedCrossWriteRows int64
	if err := WithAuditedScopeTx(ctx, db, "R6-SEG-01 test proof", func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE webhook_destinations SET url='https://hijack2.invalid' WHERE tenant_id=$1", tenantB)
		if err != nil {
			return err
		}
		auditedCrossWriteRows, err = res.RowsAffected()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if auditedCrossWriteRows != 0 {
		t.Fatalf("audited_scope é somente leitura; escrita cruzada deveria afetar 0 linhas, afetou %d", auditedCrossWriteRows)
	}

	var noReasonVisible int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM webhook_destinations WHERE tenant_id IN ($1,$2)", tenantA, tenantB).Scan(&noReasonVisible); err != nil {
		t.Fatal(err)
	}
	if noReasonVisible != 0 {
		t.Fatalf("sem tenant_id nem access_reason definidos, nenhuma linha deve ser visível: got %d", noReasonVisible)
	}
}

// TestWithWorkerCellTxScopesToOneCellNotAnyTenant prova R6-SEG-01-S03: um
// worker de fila (claim/complete de intents, deadline sweep) enxerga e
// escreve protocolos de múltiplos tenants dentro da própria célula, mas não
// de outra célula — autoridade operacional específica, não acesso global.
func TestWithWorkerCellTxScopesToOneCellNotAnyTenant(t *testing.T) {
	dsn := os.Getenv("R2_RUNTIME_TEST_DSN")
	if dsn == "" {
		t.Skip("requires PostgreSQL migrado com hub_runtime em R2_RUNTIME_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	protocolX, protocolY := idgen.New(), idgen.New()
	cellX, cellY := "pg-cell-x-"+randSuffix(), "pg-cell-y-"+randSuffix()
	t.Cleanup(func() {
		_ = WithWorkerCellTx(ctx, db, cellX, func(tx *sql.Tx) error {
			_, _ = tx.ExecContext(ctx, "DELETE FROM protocols WHERE protocol_id=$1", protocolX)
			return nil
		})
		_ = WithWorkerCellTx(ctx, db, cellY, func(tx *sql.Tx) error {
			_, _ = tx.ExecContext(ctx, "DELETE FROM protocols WHERE protocol_id=$1", protocolY)
			return nil
		})
	})

	insert := func(cell, protocolID string) error {
		return WithWorkerCellTx(ctx, db, cell, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `INSERT INTO protocols(protocol_id,tenant_id,application_id,cell_id,idempotency_key,request_hash,request_body,mode,dispatch_mode,command_id,status,client_deadline_at)
				VALUES ($1,$2,'app',$3,$4,'hash','{}','ASYNC','QUEUED',$5,'ACCEPTED',clock_timestamp()+interval '1 hour')`,
				protocolID, "tenant-"+protocolID, cell, "key-"+protocolID, idgen.New())
			return err
		})
	}
	if err := insert(cellX, protocolX); err != nil {
		t.Fatalf("insert cellX: %v", err)
	}
	if err := insert(cellY, protocolY); err != nil {
		t.Fatalf("insert cellY: %v", err)
	}

	var visibleInCellX int
	if err := WithWorkerCellTx(ctx, db, cellX, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT count(*) FROM protocols WHERE protocol_id IN ($1,$2)", protocolX, protocolY).Scan(&visibleInCellX)
	}); err != nil {
		t.Fatal(err)
	}
	if visibleInCellX != 1 {
		t.Fatalf("worker da célula X só deve ver o protocolo da própria célula: got %d", visibleInCellX)
	}

	var crossCellWriteRows int64
	if err := WithWorkerCellTx(ctx, db, cellX, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE protocols SET status='CANCELLED' WHERE protocol_id=$1", protocolY)
		if err != nil {
			return err
		}
		crossCellWriteRows, err = res.RowsAffected()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if crossCellWriteRows != 0 {
		t.Fatalf("worker da célula X não pode escrever protocolo da célula Y: afetou %d", crossCellWriteRows)
	}

	var sameCellWriteRows int64
	if err := WithWorkerCellTx(ctx, db, cellY, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE protocols SET status='CANCELLED' WHERE protocol_id=$1", protocolY)
		if err != nil {
			return err
		}
		sameCellWriteRows, err = res.RowsAffected()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if sameCellWriteRows != 1 {
		t.Fatalf("worker da própria célula Y deve conseguir escrever entre tenants da mesma célula: afetou %d", sameCellWriteRows)
	}
}

func randSuffix() string {
	return time.Now().UTC().Format("150405.000000000")
}

func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
