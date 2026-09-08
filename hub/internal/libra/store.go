// Package libra owns exact economic facts, strict reservations and the journal.
package libra

import (
	"ai-hub/hub/internal/contracts/economics"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLimitExceeded = economics.ErrLimitExceeded
	ErrLimitMissing  = errors.New("libra: explicit credit limit required")
	ErrConflict      = errors.New("libra: identity or immutable value conflict")
	ErrIncomplete    = errors.New("libra: financial obligations not reconciled")
	ErrNotFound      = errors.New("libra: resource not found")
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store                { return &Store{db: db} }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func digest(v []byte) string                    { h := sha256.Sum256(v); return hex.EncodeToString(h[:]) }
func economicKey(parts ...any) string           { b, _ := json.Marshal(parts); return digest(b) }
func accountLock(ctx context.Context, tx *sql.Tx, tenant string) error {
	_, e := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "finance:"+tenant)
	return e
}

// Legacy wrappers retain source compatibility. New producers use decimal strings.
func (s *Store) Reserve(ctx context.Context, tenant, protocol string, amount float64, currency string, _ float64) error {
	return s.ReserveExact(ctx, tenant, protocol, strconv.FormatFloat(amount, 'f', -1, 64), currency)
}
func (s *Store) CreditLimit(ctx context.Context, tenant string) (float64, error) {
	v, e := s.CreditLimitExact(ctx, tenant)
	if e != nil {
		return 0, e
	}
	return strconv.ParseFloat(string(v), 64)
}
func (s *Store) CreditLimitExact(ctx context.Context, tenant string) (Decimal, error) {
	var v string
	e := s.db.QueryRowContext(ctx, `SELECT limit_amount::text FROM credit_limits WHERE tenant_id=$1`, tenant).Scan(&v)
	if errors.Is(e, sql.ErrNoRows) {
		return "", ErrLimitMissing
	}
	return Decimal(v), e
}

func (s *Store) ReserveExact(ctx context.Context, tenant, protocol, amount, currency string) error {
	d, e := ParseDecimal(amount)
	if e != nil || d.Sign() < 0 || tenant == "" || len(currency) != 3 {
		return errors.New("libra: invalid reservation")
	}
	if _, e = uuid.Parse(protocol); e != nil {
		return errors.New("libra: invalid protocol")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = accountLock(ctx, tx, tenant); e != nil {
		return e
	}
	var oldTenant, oldCurrency, oldAmount, state string
	e = tx.QueryRowContext(ctx, `SELECT tenant_id,currency,amount::text,state FROM reservations WHERE protocol_id=$1 FOR UPDATE`, protocol).Scan(&oldTenant, &oldCurrency, &oldAmount, &state)
	if e == nil {
		old, _ := Decimal(oldAmount).Canonical()
		want, _ := d.Canonical()
		if oldTenant != tenant || oldCurrency != currency || old != want {
			return ErrConflict
		}
		if state == "RELEASED" || state == "EXPIRED" {
			return ErrConflict
		}
		return tx.Commit()
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return e
	}
	var approved bool
	e = tx.QueryRowContext(ctx, `SELECT currency=$2 AND limit_amount >= $3::numeric + COALESCE((SELECT SUM(amount) FROM reservations WHERE tenant_id=$1 AND currency=$2 AND state IN ('RESERVED','CAPTURED','UNCERTAIN_HOLD')),0) FROM credit_limits WHERE tenant_id=$1 FOR UPDATE`, tenant, currency, amount).Scan(&approved)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrLimitMissing
	}
	if e != nil {
		return e
	}
	if !approved {
		return ErrLimitExceeded
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO reservations(protocol_id,tenant_id,amount,currency,state) VALUES($1,$2,$3,$4,'RESERVED')`, protocol, tenant, amount, currency)
	if e != nil {
		return e
	}
	return tx.Commit()
}

// ReconcileReservation requires positive evidence; client expiry alone creates a hold.
func (s *Store) ReconcileReservation(ctx context.Context, tenant, protocol, state, evidence string) error {
	if evidence == "" || (state != "CAPTURED" && state != "RELEASED" && state != "UNCERTAIN_HOLD") {
		return errors.New("libra: invalid reconciliation evidence")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = accountLock(ctx, tx, tenant); e != nil {
		return e
	}
	if e = reservationState(ctx, tx, tenant, protocol, state, evidence); e != nil {
		return e
	}
	return tx.Commit()
}
func reservationState(ctx context.Context, tx *sql.Tx, tenant, protocol, state, evidence string) error {
	var old string
	e := tx.QueryRowContext(ctx, `SELECT state FROM reservations WHERE tenant_id=$1 AND protocol_id=$2 FOR UPDATE`, tenant, protocol).Scan(&old)
	if errors.Is(e, sql.ErrNoRows) {
		return nil
	}
	if e != nil {
		return e
	}
	if old == state {
		return nil
	}
	if old == "CAPTURED" || old == "RELEASED" {
		if state == "UNCERTAIN_HOLD" {
			return nil
		}
		return ErrConflict
	}
	_, e = tx.ExecContext(ctx, `UPDATE reservations SET state=$3,evidence_id=$4,updated_at=clock_timestamp() WHERE tenant_id=$1 AND protocol_id=$2`, tenant, protocol, state, evidence)
	return e
}

// Old unrestricted mutation entry points deliberately refuse to invent evidence.
func (s *Store) Capture(context.Context, string) error {
	return errors.New("libra: use evidenced reconciliation")
}
func (s *Store) Release(context.Context, string) error {
	return errors.New("libra: use evidenced reconciliation")
}
func (s *Store) RecordFact(context.Context, string, string, string, string, float64, string) error {
	return errors.New("libra: frozen economic snapshot required")
}

// ApplyEvent confirms inbox, exact facts, plan consumption, journal and reservation
// in one transaction. Economic keys survive new transport event IDs and reordering.
func (s *Store) ApplyEvent(ctx context.Context, consumer, eventID string, ev EconomicEvent) error {
	if consumer != "revenue" && consumer != "cost" {
		return errors.New("libra: invalid consumer")
	}
	if ev.TenantID == "" || ev.EvidenceID == "" || eventID == "" || ev.OccurredAt.IsZero() {
		return errors.New("libra: economic identity/evidence required")
	}
	if _, e := uuid.Parse(ev.ProtocolID); e != nil {
		return errors.New("libra: invalid protocol identity")
	}
	if e := ev.EconomicSnapshot.Validate(); e != nil {
		return e
	}
	body, _ := json.Marshal(ev)
	hash := digest(body)
	snap, _ := json.Marshal(ev.EconomicSnapshot)
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = accountLock(ctx, tx, ev.TenantID); e != nil {
		return e
	}
	var old string
	e = tx.QueryRowContext(ctx, `SELECT payload_hash FROM finance_inbox WHERE consumer=$1 AND event_id=$2`, consumer, eventID).Scan(&old)
	if e == nil {
		if old != hash {
			return ErrConflict
		}
		return tx.Commit()
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO finance_snapshots(tenant_id,protocol_id,snapshot_hash,snapshot) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, ev.TenantID, ev.ProtocolID, digest(snap), string(snap))
	if e != nil {
		return e
	}
	if e = tx.QueryRowContext(ctx, `SELECT snapshot_hash FROM finance_snapshots WHERE tenant_id=$1 AND protocol_id=$2`, ev.TenantID, ev.ProtocolID).Scan(&old); e != nil {
		return e
	}
	if old != digest(snap) {
		return ErrConflict
	}
	rules := ev.EconomicSnapshot.Sell
	kind := "REVENUE"
	incidence := ev.Status
	if consumer == "cost" {
		rules = ev.EconomicSnapshot.Buy
		kind = "COST"
		incidence = ev.Kind
	}
	for _, rule := range rules {
		eligible := false
		for _, i := range rule.Incidence {
			if i == incidence {
				eligible = true
			}
		}
		if !eligible {
			continue
		}
		unit := ev.ProtocolID
		switch rule.UnitScope {
		case "STEP":
			unit = ev.StepID
		case "OPERATION":
			unit = ev.OperationID
		case "ATTEMPT":
			unit = ev.AttemptID
		}
		if unit == "" {
			return errors.New("libra: missing economic unit identity")
		}
		contract := rule.ContractID
		version := rule.Version
		if contract == "" {
			contract = ev.EconomicSnapshot.ContractID
			version = ev.EconomicSnapshot.Version
		}
		if version < 1 {
			return errors.New("libra: invalid contract version")
		}
		key := economicKey(ev.TenantID, contract, version, kind, rule.Meter, rule.UnitScope, unit)
		var exists bool
		if e = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM economic_facts WHERE unit_key=$1)`, key).Scan(&exists); e != nil {
			return e
		}
		if exists {
			continue
		}
		previous := int64(0)
		planKey := ""
		if rule.Plan != nil {
			planKey = economicKey(contract, version, kind, rule.Meter, rule.Plan.Period)
			e = tx.QueryRowContext(ctx, `SELECT units FROM plan_consumption WHERE tenant_id=$1 AND plan_key=$2`, ev.TenantID, planKey).Scan(&previous)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return e
			}
		}
		amount, e := Price(rule, previous, 1)
		if e != nil {
			return e
		}
		_, e = tx.ExecContext(ctx, `INSERT INTO economic_facts(tenant_id,protocol_id,kind,meter,amount,currency,occurred_at,unit_key,contract_id,contract_version,evidence_id,operation_id,settlement_party,snapshot,provenance) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'SNAPSHOT_VERIFIED')`, ev.TenantID, ev.ProtocolID, kind, rule.Meter, string(amount), ev.EconomicSnapshot.Currency, ev.OccurredAt, key, contract, version, ev.EvidenceID, ev.OperationID, ev.EconomicSnapshot.SettlementParty, string(snap))
		if e != nil {
			return e
		}
		if planKey != "" {
			_, e = tx.ExecContext(ctx, `INSERT INTO plan_consumption(tenant_id,plan_key,units) VALUES($1,$2,1) ON CONFLICT(tenant_id,plan_key) DO UPDATE SET units=plan_consumption.units+1`, ev.TenantID, planKey)
			if e != nil {
				return e
			}
		}
		if kind == "REVENUE" || ev.EconomicSnapshot.SettlementParty == "HUB" {
			if e = book(ctx, tx, key, ev.TenantID, ev.ProtocolID, kind, amount, ev.EconomicSnapshot.Currency, contract, ev.EvidenceID); e != nil {
				return e
			}
		}
	}
	if consumer == "revenue" {
		state := "UNCERTAIN_HOLD"
		if ev.Status == "SUCCEEDED" || ev.Status == "PARTIALLY_SUCCEEDED" {
			state = "CAPTURED"
		} else if ev.SafeToRelease && ev.ExternalState != "UNKNOWN" {
			state = "RELEASED"
		}
		if e = reservationState(ctx, tx, ev.TenantID, ev.ProtocolID, state, ev.EvidenceID); e != nil {
			return e
		}
	}
	// Any fact arriving into an already exported period opens a visible dispute;
	// closed export bytes are never rewritten.
	var closed bool
	if e = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM settlement_periods WHERE tenant_id=$1 AND period_start<=$2 AND period_end>$2)`, ev.TenantID, ev.OccurredAt).Scan(&closed); e != nil {
		return e
	}
	if closed {
		_, e = tx.ExecContext(ctx, `INSERT INTO finance_quarantine(id,consumer,event_id,reason,payload_hash) VALUES($1,$2,$3,'LATE_FACT_CLOSED_PERIOD',$4) ON CONFLICT DO NOTHING`, economicKey(consumer, eventID, "late"), consumer, eventID, hash)
		if e != nil {
			return e
		}
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO finance_inbox(consumer,event_id,payload_hash,tenant_id) VALUES($1,$2,$3,$4)`, consumer, eventID, hash, ev.TenantID)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func book(ctx context.Context, tx *sql.Tx, key, tenant, protocol, kind string, amount Decimal, currency, contract, evidence string) error {
	_, e := tx.ExecContext(ctx, `INSERT INTO journal_batches(id,tenant_id,unit_key) VALUES($1,$2,$1)`, key, tenant)
	if e != nil {
		return e
	}
	debit, credit := "CUSTOMER_RECEIVABLE", "REVENUE"
	if kind == "COST" {
		debit, credit = "PROVIDER_EXPENSE", "PROVIDER_PAYABLE"
	}
	a, e := amount.Rat()
	if e != nil {
		return e
	}
	if a.Sign() < 0 {
		debit, credit = credit, debit
		a.Neg(a)
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO ledger_entries(batch_id,account,direction,amount,currency,origin_protocol_id,tenant_id,evidence_id,contract_id) VALUES($1,$2,'DEBIT',$4,$5,$6,$7,$8,$9),($1,$3,'CREDIT',$4,$5,$6,$7,$8,$9)`, key, debit, credit, a.FloatString(8), currency, protocol, tenant, evidence, contract)
	return e
}
func (s *Store) Quarantine(ctx context.Context, consumer, eventID, reason string, payload []byte) error {
	hash := digest(payload)
	_, e := s.db.ExecContext(ctx, `INSERT INTO finance_quarantine(id,consumer,event_id,reason,payload_hash) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, economicKey(consumer, eventID, hash), consumer, eventID, reason, hash)
	return e
}

func (s *Store) SetWatermark(ctx context.Context, tenant, producer, evidence string, through time.Time) error {
	if tenant == "" || evidence == "" || (producer != "orbita" && producer != "cometa") || through.IsZero() {
		return errors.New("libra: explicit producer completeness evidence required")
	}
	_, e := s.db.ExecContext(ctx, `INSERT INTO finance_watermarks(tenant_id,producer,complete_through,evidence_id) VALUES($1,$2,$3,$4) ON CONFLICT(tenant_id,producer) DO UPDATE SET complete_through=GREATEST(finance_watermarks.complete_through,EXCLUDED.complete_through),evidence_id=EXCLUDED.evidence_id`, tenant, producer, through, evidence)
	return e
}

func invalidError(e error) bool {
	return errors.Is(e, ErrConflict) || errors.Is(e, ErrLimitExceeded) || errors.Is(e, ErrLimitMissing)
}

var _ = fmt.Sprintf
