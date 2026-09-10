package libra

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/queue"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func TestExactPlanPrices(t *testing.T) {
	tests := []struct {
		name string
		rule PricingRule
		n    int64
		want Decimal
	}{
		{"marginal-four", PricingRule{Amount: "1", Plan: &Plan{Kind: "MARGINAL", Period: "fixture-2026-09", Tiers: []Tier{{UpTo: 2, Amount: "1"}, {UpTo: 0, Amount: "0.5"}}}}, 4, "3.00000000"},
		{"volume-four", PricingRule{Amount: "1", Plan: &Plan{Kind: "VOLUME", Period: "fixture-2026-09", Tiers: []Tier{{UpTo: 2, Amount: "1"}, {UpTo: 0, Amount: "0.5"}}}}, 4, "2.00000000"},
		{"four-decimal", PricingRule{Amount: "0.1234"}, 3, "0.37020000"},
		{"allowance", PricingRule{Amount: "0.8", Plan: &Plan{Kind: "ALLOWANCE", Period: "fixture-2026-09", IncludedUnits: 100, Excess: "CHARGE"}}, 103, "2.40000000"},
	}
	for _, x := range tests {
		t.Run(x.name, func(t *testing.T) {
			v, e := Price(x.rule, 0, x.n)
			if e != nil || v != x.want {
				t.Fatalf("got %s %v want %s", v, e, x.want)
			}
		})
	}
	for _, bad := range []string{"NaN", "1e3", "0.000000001", "-", "01", "1.2.3"} {
		if _, e := ParseDecimal(bad); e == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	var d Decimal
	if e := json.Unmarshal([]byte(`0.1`), &d); e != nil || d != "0.1" {
		t.Fatal(d, e)
	}
}
func financeDB(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("R2_FINANCE_DSN")
	if dsn == "" {
		t.Skip("requires real PostgreSQL R2_FINANCE_DSN")
	}
	db, e := sql.Open("postgres", dsn)
	if e != nil {
		t.Fatal(e)
	}
	db.SetMaxOpenConns(12)
	t.Cleanup(func() { db.Close() })
	if e = db.Ping(); e != nil {
		t.Fatal(e)
	}
	return NewStore(db)
}
func snapshot() Snapshot {
	return Snapshot{ContractID: "fixture-sale-v1", Version: 1, Currency: "BRL", SettlementParty: "HUB", Buy: []PricingRule{{ContractID: "fixture-buy", Version: 1, Meter: "submit", Amount: "0.2", Incidence: []string{"SUBMITTED", "SUCCEEDED"}, UnitScope: "OPERATION"}}, Sell: []PricingRule{{Meter: "product", Amount: "1", Incidence: []string{"SUCCEEDED", "PARTIALLY_SUCCEEDED"}, UnitScope: "PROTOCOL"}}}
}
func event(tenant string) EconomicEvent {
	return EconomicEvent{TenantID: tenant, ProtocolID: uuid.NewString(), OperationID: uuid.NewString(), AttemptID: uuid.NewString(), Status: "SUCCEEDED", Kind: "SUCCEEDED", EvidenceID: uuid.NewString(), OccurredAt: time.Now().UTC().Add(-time.Hour), EconomicSnapshot: snapshot()}
}
func tenant() string { return "r2-fin-" + uuid.NewString() }
func limit(t *testing.T, s *Store, tenant, amount string) {
	t.Helper()
	if _, e := s.db.Exec(`INSERT INTO credit_limits(tenant_id,limit_amount,currency) VALUES($1,$2,'BRL')`, tenant, amount); e != nil {
		t.Fatal(e)
	}
}
func apply(t *testing.T, s *Store, kind string, e EconomicEvent) {
	t.Helper()
	if err := s.ApplyEvent(context.Background(), kind, uuid.NewString(), e); err != nil {
		t.Fatal(err)
	}
}
func count(t *testing.T, s *Store, query string, args ...any) int {
	t.Helper()
	var n int
	if e := s.db.QueryRow(query, args...).Scan(&n); e != nil {
		t.Fatal(e)
	}
	return n
}
func TestFinancePostgresScenarios(t *testing.T) {
	s := financeDB(t)
	ctx := context.Background()
	t.Run("R2-FIN-UNKNOWN-no_capture", func(t *testing.T) {
		e := event(tenant())
		e.Status = "UNKNOWN"
		e.Kind = "UNKNOWN"
		apply(t, s, "cost", e)
		apply(t, s, "revenue", e)
		if got := count(t, s, `SELECT count(*) FROM economic_facts WHERE tenant_id=$1`, e.TenantID); got != 0 {
			t.Fatalf("UNKNOWN generated economic facts: %d", got)
		}
		if got := count(t, s, `SELECT count(*) FROM ledger_entries WHERE tenant_id=$1`, e.TenantID); got != 0 {
			t.Fatalf("UNKNOWN generated ledger entries: %d", got)
		}
	})
	t.Run("R2-FIN-01-S01_snapshot_frozen", func(t *testing.T) {
		e := event(tenant())
		apply(t, s, "revenue", e)
		e.EconomicSnapshot.Sell[0].Amount = "9"
		if err := s.ApplyEvent(ctx, "revenue", uuid.NewString(), e); !errors.Is(err, ErrConflict) {
			t.Fatal(err)
		}
		f, _ := s.Facts(ctx, e.TenantID, "", 0, 50)
		if len(f) != 1 || f[0].Amount != "1.00000000" {
			t.Fatal(f)
		}
	})
	t.Run("R2-FIN-01-S02_client_direct", func(t *testing.T) {
		e := event(tenant())
		e.EconomicSnapshot.SettlementParty = "CLIENT_DIRECT"
		apply(t, s, "cost", e)
		if count(t, s, `SELECT count(*) FROM economic_facts WHERE tenant_id=$1`, e.TenantID) != 1 || count(t, s, `SELECT count(*) FROM ledger_entries WHERE tenant_id=$1`, e.TenantID) != 0 {
			t.Fatal("client direct created Hub debt or lost usage")
		}
	})
	t.Run("R2-FIN-01-S03_submit_then_expiry", func(t *testing.T) {
		e := event(tenant())
		e.Kind = "SUBMITTED"
		e.Status = "EXPIRED"
		apply(t, s, "cost", e)
		apply(t, s, "revenue", e)
		f, _ := s.Facts(ctx, e.TenantID, "", 0, 50)
		if len(f) != 1 || f[0].Kind != "COST" || f[0].Amount != "0.20000000" {
			t.Fatal(f)
		}
	})
	t.Run("R2-FIN-02-S01_distinct_operations", func(t *testing.T) {
		e := event(tenant())
		other := e
		other.OperationID = uuid.NewString()
		apply(t, s, "cost", other)
		apply(t, s, "cost", e)
		apply(t, s, "cost", e)
		apply(t, s, "cost", other)
		if count(t, s, `SELECT count(*) FROM economic_facts WHERE tenant_id=$1`, e.TenantID) != 2 {
			t.Fatal("semantic collision or duplicate")
		}
	})
	t.Run("R2-FIN-02-S02_tariffed_polls", func(t *testing.T) {
		e := event(tenant())
		e.Kind = "STATUS"
		e.EconomicSnapshot.Buy = []PricingRule{{Meter: "status", Amount: "0.01", Incidence: []string{"STATUS"}, UnitScope: "ATTEMPT"}}
		for i := 0; i < 3; i++ {
			e.AttemptID = uuid.NewString()
			apply(t, s, "cost", e)
			apply(t, s, "cost", e)
		}
		if count(t, s, `SELECT count(*) FROM economic_facts WHERE tenant_id=$1`, e.TenantID) != 3 {
			t.Fatal("poll evidence dedup mismatch")
		}
	})
	t.Run("R2-FIN-02-S03_exact_arithmetic", func(t *testing.T) {
		e := event(tenant())
		e.EconomicSnapshot.Buy = []PricingRule{{Meter: "one", Amount: "0.1", Incidence: []string{"SUCCEEDED"}, UnitScope: "OPERATION"}, {Meter: "two", Amount: "0.2", Incidence: []string{"SUCCEEDED"}, UnitScope: "OPERATION"}}
		apply(t, s, "cost", e)
		var sum string
		s.db.QueryRow(`SELECT sum(amount)::text FROM economic_facts WHERE tenant_id=$1`, e.TenantID).Scan(&sum)
		if sum != "0.30000000" {
			t.Fatal(sum)
		}
	})
	t.Run("R2-FIN-03-S01_cumulative_balance", func(t *testing.T) {
		ten := tenant()
		limit(t, s, ten, "2")
		for i := 0; i < 3; i++ {
			e := event(ten)
			err := s.ReserveExact(ctx, ten, e.ProtocolID, "1", "BRL")
			if i == 2 {
				if !errors.Is(err, ErrLimitExceeded) {
					t.Fatal(err)
				}
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			apply(t, s, "revenue", e)
		}
		a, _ := s.Accounts(ctx, ten)
		if a[0].Available != "0.00000000" {
			t.Fatal(a)
		}
	})
	t.Run("R2-FIN-03-S02_concurrent_last_balance", func(t *testing.T) {
		ten := tenant()
		limit(t, s, ten, "1")
		errs := make(chan error, 20)
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); errs <- s.ReserveExact(ctx, ten, uuid.NewString(), "1", "BRL") }()
		}
		wg.Wait()
		close(errs)
		granted := 0
		for e := range errs {
			if e == nil {
				granted++
			} else if !errors.Is(e, ErrLimitExceeded) {
				t.Fatal(e)
			}
		}
		if granted != 1 {
			t.Fatal(granted)
		}
		if e := s.ReserveExact(ctx, ten, uuid.NewString(), "-1", "BRL"); e == nil {
			t.Fatal("negative reserve")
		}
		if e := s.ReserveExact(ctx, ten, uuid.NewString(), "1", "USD"); e == nil {
			t.Fatal("currency mismatch")
		}
	})
	t.Run("R2-FIN-03-S03_missing_limit", func(t *testing.T) {
		if e := s.ReserveExact(ctx, tenant(), uuid.NewString(), "1", "BRL"); !errors.Is(e, ErrLimitMissing) {
			t.Fatal(e)
		}
	})
	t.Run("R2-FIN-03-S04_unknown_hold", func(t *testing.T) {
		e := event(tenant())
		limit(t, s, e.TenantID, "1")
		if er := s.ReserveExact(ctx, e.TenantID, e.ProtocolID, "1", "BRL"); er != nil {
			t.Fatal(er)
		}
		e.Status = "EXPIRED"
		e.ExternalState = "UNKNOWN"
		apply(t, s, "revenue", e)
		apply(t, s, "revenue", e)
		a, _ := s.Accounts(ctx, e.TenantID)
		if a[0].Holds != "1.00000000" {
			t.Fatal(a)
		}
		if er := s.ReserveExact(ctx, e.TenantID, uuid.NewString(), "1", "BRL"); !errors.Is(er, ErrLimitExceeded) {
			t.Fatal(er)
		}
	})
	t.Run("R2-FIN-04-S01_marginal_journal", func(t *testing.T) {
		e := event(tenant())
		e.EconomicSnapshot.Sell[0].Plan = &Plan{Kind: "MARGINAL", Period: "fixture", Tiers: []Tier{{UpTo: 2, Amount: "1"}, {UpTo: 0, Amount: "0.5"}}}
		for i := 0; i < 4; i++ {
			e.ProtocolID = uuid.NewString()
			apply(t, s, "revenue", e)
		}
		var sum string
		s.db.QueryRow(`SELECT sum(amount)::text FROM economic_facts WHERE tenant_id=$1`, e.TenantID).Scan(&sum)
		if sum != "3.00000000" {
			t.Fatal(sum)
		}
	})
	t.Run("R2-FIN-04-S02_package", func(t *testing.T) {
		e := event(tenant())
		e.EconomicSnapshot.Sell[0].Amount = "5"
		e.EconomicSnapshot.Sell[0].Plan = &Plan{Kind: "PACKAGE"}
		apply(t, s, "cost", e)
		e.OperationID = uuid.NewString()
		apply(t, s, "cost", e)
		apply(t, s, "revenue", e)
		apply(t, s, "revenue", e)
		f, _ := s.Facts(ctx, e.TenantID, "REVENUE", 0, 50)
		if len(f) != 1 || f[0].Amount != "5.00000000" {
			t.Fatal(f)
		}
	})
	t.Run("R2-FIN-04-S03_concurrent_allowance", func(t *testing.T) {
		e := event(tenant())
		e.EconomicSnapshot.Sell[0].Plan = &Plan{Kind: "ALLOWANCE", Period: "fixture", IncludedUnits: 1, Excess: "CHARGE"}
		errs := make(chan error, 2)
		for i := 0; i < 2; i++ {
			copy := e
			copy.ProtocolID = uuid.NewString()
			go func() { errs <- s.ApplyEvent(ctx, "revenue", uuid.NewString(), copy) }()
		}
		for i := 0; i < 2; i++ {
			if err := <-errs; err != nil {
				t.Fatal(err)
			}
		}
		f, _ := s.Facts(ctx, e.TenantID, "", 0, 50)
		if len(f) != 2 || count(t, s, `SELECT count(*) FROM economic_facts WHERE tenant_id=$1 AND amount=0`, e.TenantID) != 1 {
			t.Fatal(f)
		}
	})
	t.Run("R2-FIN-05-S01_balanced_and_immutable", func(t *testing.T) {
		e := event(tenant())
		apply(t, s, "revenue", e)
		if count(t, s, `SELECT count(*) FROM (SELECT batch_id,currency FROM ledger_entries WHERE tenant_id=$1 GROUP BY batch_id,currency HAVING SUM(CASE direction WHEN 'DEBIT' THEN amount ELSE -amount END)<>0) b`, e.TenantID) != 0 {
			t.Fatal("unbalanced")
		}
		if _, err := s.db.Exec(`UPDATE economic_facts SET amount=9 WHERE tenant_id=$1`, e.TenantID); err == nil {
			t.Fatal("facts editable")
		}
		if _, err := s.db.Exec(`INSERT INTO ledger_entries(batch_id,account,direction,amount,currency,tenant_id) VALUES($1,'broken','DEBIT',1,'BRL',$2)`, uuid.NewString(), e.TenantID); err == nil {
			t.Fatal("DB accepted unbalanced commit")
		}
	})
	t.Run("R2-FIN-05-S02_replay_inbox_atomic", func(t *testing.T) {
		e := event(tenant())
		id := uuid.NewString()
		if er := s.ApplyEvent(ctx, "revenue", id, e); er != nil {
			t.Fatal(er)
		}
		if er := s.ApplyEvent(ctx, "revenue", id, e); er != nil {
			t.Fatal(er)
		}
		apply(t, s, "revenue", e)
		if count(t, s, `SELECT count(*) FROM ledger_entries WHERE tenant_id=$1`, e.TenantID) != 2 {
			t.Fatal("replay generated entries")
		}
		if count(t, s, `SELECT count(*) FROM economic_facts WHERE tenant_id=$1`, e.TenantID) != 1 {
			t.Fatal("semantic redelivery generated a second economic fact")
		}
		if count(t, s, `SELECT count(*) FROM finance_inbox WHERE tenant_id=$1`, e.TenantID) != 2 {
			t.Fatal("redelivery was not durably recorded per event identity")
		}
	})
	t.Run("R2-FIN-05-S03_authorized_compensation", func(t *testing.T) {
		e := event(tenant())
		apply(t, s, "revenue", e)
		j, _ := s.Journal(ctx, e.TenantID, 0, 50)
		a, err := s.PrepareAdjustment(ctx, uuid.NewString(), e.TenantID, j[0].BatchID, "synthetic correction", "preparer")
		if err != nil {
			t.Fatal(err)
		}
		if err = s.ApproveAdjustment(ctx, a.ID, e.TenantID, "preparer"); err == nil {
			t.Fatal("self approval")
		}
		if err = s.ApproveAdjustment(ctx, a.ID, e.TenantID, "approver"); err != nil {
			t.Fatal(err)
		}
		if err = s.ApproveAdjustment(ctx, a.ID, e.TenantID, "approver"); err != nil {
			t.Fatal(err)
		}
		if count(t, s, `SELECT count(*) FROM ledger_entries WHERE tenant_id=$1`, e.TenantID) != 4 {
			t.Fatal("compensation not immutable/idempotent")
		}
	})
	t.Run("R2-FIN-06-S01_incomplete_watermark", func(t *testing.T) {
		e := event(tenant())
		apply(t, s, "revenue", e)
		_, err := s.ClosePeriod(ctx, uuid.NewString(), e.TenantID, "closer", time.Now().Add(-2*time.Hour), time.Now())
		if !errors.Is(err, ErrIncomplete) {
			t.Fatal(err)
		}
	})
	t.Run("R2-FIN-06-S02_export_repeated", func(t *testing.T) {
		e := event(tenant())
		apply(t, s, "revenue", e)
		start, end := time.Now().UTC().Add(-2*time.Hour).Truncate(time.Second), time.Now().UTC().Truncate(time.Second)
		for _, producer := range []string{"orbita", "cometa"} {
			if err := s.SetWatermark(ctx, e.TenantID, producer, "synthetic-completeness", end); err != nil {
				t.Fatal(err)
			}
		}
		id := uuid.NewString()
		v, err := s.ClosePeriod(ctx, id, e.TenantID, "closer", start, end)
		if err != nil {
			t.Fatal(err)
		}
		again, err := s.ClosePeriod(ctx, id, e.TenantID, "closer", start, end)
		if err != nil || v.Checksum != again.Checksum || string(v.Body) != string(again.Body) {
			t.Fatal("export changed", err)
		}
		if err = s.Receipt(ctx, e.TenantID, id, "receipt", v.Checksum, "erp-fixture"); err != nil {
			t.Fatal(err)
		}
		if err = s.Receipt(ctx, e.TenantID, id, "receipt", v.Checksum, "erp-fixture"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("R2-FIN-06-S03_sla_dispute", func(t *testing.T) {
		e := event(tenant())
		apply(t, s, "cost", e)
		f, _ := s.Facts(ctx, e.TenantID, "", 0, 50)
		if _, err := s.Dispute(ctx, e.TenantID, f[0].ID, "0.2", "late provider", "sla-receipt", "reviewer"); err != nil {
			t.Fatal(err)
		}
		if count(t, s, `SELECT count(*) FROM finance_disputes WHERE tenant_id=$1 AND state='OPEN'`, e.TenantID) != 1 {
			t.Fatal("dispute missing")
		}
		f, _ = s.Facts(ctx, e.TenantID, "", 0, 50)
		if f[0].Amount != "0.20000000" {
			t.Fatal("dispute changed fact")
		}
	})
	t.Run("quarantine_and_tenant_authorization", func(t *testing.T) {
		e := event(tenant())
		raw, _ := json.Marshal(e)
		env := queue.Envelope{EventID: uuid.NewString(), SchemaVersion: 1, TenantID: "wrong", ProtocolID: e.ProtocolID, Producer: "orbita", Payload: raw}
		if er := s.ProcessEnvelope(ctx, "revenue", env); er != nil {
			t.Fatal(er)
		}
		if count(t, s, `SELECT count(*) FROM finance_quarantine WHERE event_id=$1`, env.EventID) != 1 {
			t.Fatal("missing quarantine")
		}
		h := NewHandlers(s, slog.New(slog.NewTextHandler(io.Discard, nil)))
		req := httptest.NewRequest("GET", "/admin/v1/finance/facts?tenant_id="+e.TenantID, nil)
		req = req.WithContext(auth.WithPrincipal(ctx, auth.Principal{Subject: "user-b", TenantID: "b", Scopes: []string{"finance:read"}, ExpiresAt: time.Now().Add(time.Hour)}))
		w := httptest.NewRecorder()
		h.handleAdmin(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("failed_commit_rolls_back_inbox_and_journal", func(t *testing.T) {
		e := event(tenant())
		e.EconomicSnapshot.Sell = append(e.EconomicSnapshot.Sell, PricingRule{Meter: "bad-step", Amount: "1", Incidence: []string{"SUCCEEDED"}, UnitScope: "STEP"})
		id := uuid.NewString()
		if err := s.ApplyEvent(ctx, "revenue", id, e); err == nil || !strings.Contains(err.Error(), "identity") {
			t.Fatal(err)
		}
		if count(t, s, `SELECT count(*) FROM economic_facts WHERE tenant_id=$1`, e.TenantID) != 0 || count(t, s, `SELECT count(*) FROM finance_inbox WHERE event_id=$1`, id) != 0 {
			t.Fatal("partial financial commit")
		}
	})
}
