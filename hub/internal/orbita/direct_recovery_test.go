package orbita

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/platform/pg"
	_ "github.com/lib/pq"
)

// TestPostgresDirectRecoveryNeverResubmitsAfterRetryExpiryEvenWithFutureSLA
// prova R6-EXE-02-S01: uma intenção DIRECT com retry_until já vencido, mesmo
// com o prazo do cliente ainda no futuro, nunca gera novo SUBMIT ao
// recuperador assumi-la — ela deve encerrar como EXPIRED sem tocar o
// transporte.
func TestPostgresDirectRecoveryNeverResubmitsAfterRetryExpiryEvenWithFutureSLA(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL R2_CORE_TEST_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var submits int64
	cometa := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&submits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer cometa.Close()

	ctx := context.Background()
	tenant, protocolID, commandID := "direct-recovery-"+idgen.New(), idgen.New(), idgen.New()
	t.Cleanup(func() {
		pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
			tx.ExecContext(ctx, "DELETE FROM command_intents WHERE tenant_id=$1", tenant)
			tx.ExecContext(ctx, "DELETE FROM protocols WHERE tenant_id=$1", tenant)
			return nil
		})
	})

	now := time.Now().UTC()
	p := Protocol{
		ProtocolID: protocolID, TenantID: tenant, ApplicationID: "app-direct", CellID: "cell-direct",
		IdempotencyKey: "intent-" + idgen.New(), RequestHash: "hash-direct", RequestBody: []byte(`{"mode":"SYNC"}`),
		Mode: "SYNC", DispatchMode: string(dispatch.DispatchDirect), CommandID: commandID,
		Status: StatusAccepted, AcceptedAt: now, ClientDeadlineAt: now.Add(time.Hour),
	}
	cmd := dispatch.Command{
		TenantID: tenant, ApplicationID: p.ApplicationID, CellID: p.CellID, ProtocolID: protocolID,
		StepID: "step-direct", CommandID: commandID, DispatchMode: dispatch.DispatchDirect,
		ConfigSnapshot: []byte(`{"synthetic":true}`), RequestBody: map[string]any{"x": 1},
		AcceptedAt: now, StepDeadline: now.Add(time.Hour), RetryTTLSeconds: 30,
	}
	s := NewStore(db)
	if _, created, err := s.Admit(ctx, p, cmd, "subject-direct", false); err != nil || !created {
		t.Fatalf("admit created=%v err=%v", created, err)
	}
	if err := pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, `UPDATE command_intents SET retry_started_at=clock_timestamp()-interval '1 hour',retry_until=clock_timestamp()-interval '1 second' WHERE command_id=$1`, commandID)
		return execErr
	}); err != nil {
		t.Fatal(err)
	}

	d := NewDispatcher(cometa.URL, nil, "")
	f := NewFinalizer(s, slog.Default())
	recoveryCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	RunDirectRecovery(recoveryCtx, s, d, f, p.CellID, slog.Default())

	if got := atomic.LoadInt64(&submits); got != 0 {
		t.Fatalf("DIRECT recovery resubmitted an expired intent: submits=%d", got)
	}
	var state string
	if err := pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT state FROM command_intents WHERE command_id=$1", commandID).Scan(&state)
	}); err != nil {
		t.Fatal(err)
	}
	if state != "EXPIRED" {
		t.Fatalf("expired DIRECT intent with future client SLA was not closed as EXPIRED: state=%s", state)
	}
}
