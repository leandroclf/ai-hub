package libra

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	TenantID  string  `json:"tenant_id"`
	Currency  string  `json:"currency"`
	Limit     Decimal `json:"limit"`
	Reserved  Decimal `json:"reserved"`
	Captured  Decimal `json:"captured"`
	Holds     Decimal `json:"holds"`
	Available Decimal `json:"available"`
}

func (s *Store) Accounts(ctx context.Context, tenant string) ([]Account, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT c.tenant_id,c.currency,c.limit_amount::text,COALESCE(SUM(r.amount) FILTER(WHERE r.state='RESERVED'),0)::text,COALESCE(SUM(r.amount) FILTER(WHERE r.state='CAPTURED'),0)::text,COALESCE(SUM(r.amount) FILTER(WHERE r.state='UNCERTAIN_HOLD'),0)::text,(c.limit_amount-COALESCE(SUM(r.amount) FILTER(WHERE r.state IN('RESERVED','CAPTURED','UNCERTAIN_HOLD')),0))::text FROM credit_limits c LEFT JOIN reservations r ON r.tenant_id=c.tenant_id AND r.currency=c.currency WHERE c.tenant_id=$1 GROUP BY c.tenant_id,c.currency,c.limit_amount`, tenant)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Account{}
	for rows.Next() {
		var a Account
		if e = rows.Scan(&a.TenantID, &a.Currency, &a.Limit, &a.Reserved, &a.Captured, &a.Holds, &a.Available); e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type Fact struct {
	ID         int64     `json:"id"`
	TenantID   string    `json:"tenant_id"`
	ProtocolID string    `json:"protocol_id"`
	Kind       string    `json:"kind"`
	Meter      string    `json:"meter"`
	Amount     Decimal   `json:"amount"`
	Currency   string    `json:"currency"`
	ContractID string    `json:"contract_id"`
	EvidenceID string    `json:"evidence_id"`
	UnitKey    string    `json:"unit_key"`
	Provenance string    `json:"provenance"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (s *Store) Facts(ctx context.Context, tenant, kind string, after int64, limit int) ([]Fact, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, e := s.db.QueryContext(ctx, `SELECT id,tenant_id,protocol_id,kind,meter,amount::text,currency,COALESCE(contract_id,''),COALESCE(evidence_id,''),COALESCE(unit_key,''),provenance,occurred_at FROM economic_facts WHERE tenant_id=$1 AND ($2='' OR kind=$2) AND id>$3 ORDER BY id LIMIT $4`, tenant, kind, after, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Fact{}
	for rows.Next() {
		var f Fact
		if e = rows.Scan(&f.ID, &f.TenantID, &f.ProtocolID, &f.Kind, &f.Meter, &f.Amount, &f.Currency, &f.ContractID, &f.EvidenceID, &f.UnitKey, &f.Provenance, &f.OccurredAt); e != nil {
			return nil, e
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

type JournalEntry struct {
	ID         int64   `json:"id"`
	BatchID    string  `json:"batch_id"`
	Account    string  `json:"account"`
	Direction  string  `json:"direction"`
	Amount     Decimal `json:"amount"`
	Currency   string  `json:"currency"`
	ProtocolID string  `json:"protocol_id"`
	ContractID string  `json:"contract_id"`
	EvidenceID string  `json:"evidence_id"`
}

func (s *Store) Journal(ctx context.Context, tenant string, after int64, limit int) ([]JournalEntry, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, e := s.db.QueryContext(ctx, `SELECT id,batch_id,account,direction,amount::text,currency,COALESCE(origin_protocol_id::text,''),COALESCE(contract_id,''),COALESCE(evidence_id,'') FROM ledger_entries WHERE tenant_id=$1 AND id>$2 ORDER BY id LIMIT $3`, tenant, after, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []JournalEntry{}
	for rows.Next() {
		var v JournalEntry
		if e = rows.Scan(&v.ID, &v.BatchID, &v.Account, &v.Direction, &v.Amount, &v.Currency, &v.ProtocolID, &v.ContractID, &v.EvidenceID); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

type Adjustment struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	OriginBatchID string `json:"origin_batch_id"`
	Reason        string `json:"reason"`
	PreparedBy    string `json:"prepared_by"`
	ApprovedBy    string `json:"approved_by,omitempty"`
	State         string `json:"state"`
}

func (s *Store) PrepareAdjustment(ctx context.Context, id, tenant, origin, reason, actor string) (Adjustment, error) {
	a := Adjustment{ID: id, TenantID: tenant, OriginBatchID: origin, Reason: reason, PreparedBy: actor, State: "PREPARED"}
	if id == "" || tenant == "" || reason == "" || actor == "" {
		return a, errors.New("libra: adjustment identity and reason required")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return a, e
	}
	defer tx.Rollback()
	if e = accountLock(ctx, tx, tenant); e != nil {
		return a, e
	}
	var owner string
	if e = tx.QueryRowContext(ctx, `SELECT tenant_id FROM journal_batches WHERE id=$1`, origin).Scan(&owner); errors.Is(e, sql.ErrNoRows) {
		return a, ErrNotFound
	}
	if e != nil {
		return a, e
	}
	if owner != tenant {
		return a, ErrNotFound
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO finance_adjustments(id,tenant_id,origin_batch_id,reason,prepared_by,state) VALUES($1,$2,$3,$4,$5,'PREPARED') ON CONFLICT DO NOTHING`, id, tenant, origin, reason, actor)
	if e != nil {
		return a, e
	}
	var old Adjustment
	e = tx.QueryRowContext(ctx, `SELECT id,tenant_id,origin_batch_id,reason,prepared_by,COALESCE(approved_by,''),state FROM finance_adjustments WHERE id=$1`, id).Scan(&old.ID, &old.TenantID, &old.OriginBatchID, &old.Reason, &old.PreparedBy, &old.ApprovedBy, &old.State)
	if e != nil {
		return a, e
	}
	if old.TenantID != tenant || old.OriginBatchID != origin || old.Reason != reason || old.PreparedBy != actor {
		return a, ErrConflict
	}
	return old, tx.Commit()
}
func (s *Store) ApproveAdjustment(ctx context.Context, id, tenant, actor string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = accountLock(ctx, tx, tenant); e != nil {
		return e
	}
	var origin, reason, prepared, state, approved string
	e = tx.QueryRowContext(ctx, `SELECT origin_batch_id,reason,prepared_by,state,COALESCE(approved_by,'') FROM finance_adjustments WHERE id=$1 AND tenant_id=$2 FOR UPDATE`, id, tenant).Scan(&origin, &reason, &prepared, &state, &approved)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrNotFound
	}
	if e != nil {
		return e
	}
	if actor == "" || actor == prepared {
		return errors.New("libra: approval requires distinct authenticated actor")
	}
	if state == "APPROVED" {
		if approved != actor {
			return ErrConflict
		}
		return tx.Commit()
	}
	batch := economicKey("adjustment", id)
	_, e = tx.ExecContext(ctx, `INSERT INTO journal_batches(id,tenant_id,origin_batch_id,reason,authorizer) VALUES($1,$2,$3,$4,$5)`, batch, tenant, origin, reason, actor)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO ledger_entries(batch_id,account,direction,amount,currency,origin_protocol_id,tenant_id,evidence_id,contract_id) SELECT $1,account,CASE direction WHEN 'DEBIT' THEN 'CREDIT' ELSE 'DEBIT' END,amount,currency,origin_protocol_id,tenant_id,$2,contract_id FROM ledger_entries WHERE batch_id=$3 AND tenant_id=$4`, batch, "adjustment:"+id, origin, tenant)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `UPDATE finance_adjustments SET state='APPROVED',approved_by=$3 WHERE id=$1 AND tenant_id=$2`, id, tenant, actor)
	if e != nil {
		return e
	}
	return tx.Commit()
}

type Export struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id"`
	Start         time.Time       `json:"period_start"`
	End           time.Time       `json:"period_end"`
	LayoutVersion int             `json:"layout_version"`
	Checksum      string          `json:"checksum"`
	Body          json.RawMessage `json:"body"`
}

func (s *Store) ClosePeriod(ctx context.Context, id, tenant, actor string, start, end time.Time) (Export, error) {
	out := Export{ID: id, TenantID: tenant, Start: start, End: end, LayoutVersion: 1}
	if id == "" || tenant == "" || actor == "" || start.IsZero() || !end.After(start) || end.After(time.Now().UTC()) {
		return out, errors.New("libra: invalid historical period")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return out, e
	}
	defer tx.Rollback()
	if e = accountLock(ctx, tx, tenant); e != nil {
		return out, e
	}
	var oldTenant string
	e = tx.QueryRowContext(ctx, `SELECT tenant_id,period_start,period_end,layout_version,checksum,export_body FROM settlement_periods WHERE id=$1`, id).Scan(&oldTenant, &out.Start, &out.End, &out.LayoutVersion, &out.Checksum, &out.Body)
	if e == nil {
		if oldTenant != tenant || !out.Start.Equal(start) || !out.End.Equal(end) {
			return out, ErrConflict
		}
		return out, tx.Commit()
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return out, e
	}
	var complete bool
	e = tx.QueryRowContext(ctx, `SELECT (SELECT count(*)=2 FROM finance_watermarks WHERE tenant_id=$1 AND producer IN ('orbita','cometa') AND complete_through>=$3) AND NOT EXISTS(SELECT 1 FROM reservations WHERE tenant_id=$1 AND state IN ('RESERVED','UNCERTAIN_HOLD') AND created_at<$3) AND NOT EXISTS(SELECT 1 FROM economic_facts f WHERE f.tenant_id=$1 AND occurred_at>=$2 AND occurred_at<$3 AND (provenance<>'SNAPSHOT_VERIFIED' OR ((kind='REVENUE' OR settlement_party='HUB') AND NOT EXISTS(SELECT 1 FROM journal_batches b WHERE b.unit_key=f.unit_key)))) AND NOT EXISTS(SELECT 1 FROM finance_disputes WHERE tenant_id=$1 AND state='OPEN') AND NOT EXISTS(SELECT 1 FROM finance_adjustments WHERE tenant_id=$1 AND state='PREPARED') AND NOT EXISTS(SELECT 1 FROM settlement_periods WHERE tenant_id=$1 AND period_start<$3 AND period_end>$2)`, tenant, start, end).Scan(&complete)
	if e != nil {
		return out, e
	}
	if !complete {
		return out, ErrIncomplete
	}
	type line struct {
		ID              int64   `json:"id"`
		Kind            string  `json:"kind"`
		UnitKey         string  `json:"unit_key"`
		ContractID      string  `json:"contract_id"`
		Amount          Decimal `json:"amount"`
		Currency        string  `json:"currency"`
		SettlementParty string  `json:"settlement_party"`
		EvidenceID      string  `json:"evidence_id"`
	}
	rows, e := tx.QueryContext(ctx, `SELECT id,kind,unit_key,contract_id,amount::text,currency,settlement_party,evidence_id FROM economic_facts WHERE tenant_id=$1 AND occurred_at>=$2 AND occurred_at<$3 ORDER BY id`, tenant, start, end)
	if e != nil {
		return out, e
	}
	lines := []line{}
	for rows.Next() {
		var l line
		if e = rows.Scan(&l.ID, &l.Kind, &l.UnitKey, &l.ContractID, &l.Amount, &l.Currency, &l.SettlementParty, &l.EvidenceID); e != nil {
			rows.Close()
			return out, e
		}
		lines = append(lines, l)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return out, e
	}
	out.Body, e = json.Marshal(struct {
		ExportID string    `json:"export_id"`
		Version  int       `json:"layout_version"`
		TenantID string    `json:"tenant_id"`
		Start    time.Time `json:"period_start"`
		End      time.Time `json:"period_end"`
		Lines    []line    `json:"lines"`
	}{id, 1, tenant, start, end, lines})
	if e != nil {
		return out, e
	}
	out.Checksum = digest(out.Body)
	_, e = tx.ExecContext(ctx, `INSERT INTO settlement_periods(id,tenant_id,period_start,period_end,export_body,checksum,created_by) VALUES($1,$2,$3,$4,$5,$6,$7)`, id, tenant, start, end, []byte(out.Body), out.Checksum, actor)
	if e != nil {
		return out, e
	}
	return out, tx.Commit()
}
func (s *Store) Export(ctx context.Context, tenant, id string) (Export, error) {
	var v Export
	e := s.db.QueryRowContext(ctx, `SELECT id,tenant_id,period_start,period_end,layout_version,checksum,export_body FROM settlement_periods WHERE tenant_id=$1 AND id=$2`, tenant, id).Scan(&v.ID, &v.TenantID, &v.Start, &v.End, &v.LayoutVersion, &v.Checksum, &v.Body)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrNotFound
	}
	return v, e
}
func (s *Store) Receipt(ctx context.Context, tenant, id, receipt, checksum, actor string) error {
	if receipt == "" || actor == "" {
		return errors.New("libra: receipt identity required")
	}
	v, e := s.Export(ctx, tenant, id)
	if e != nil {
		return e
	}
	if v.Checksum != checksum {
		return ErrConflict
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO export_receipts(export_id,receipt_id,checksum,actor) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, id, receipt, checksum, actor)
	return e
}
func (s *Store) Dispute(ctx context.Context, tenant string, fact int64, amount Decimal, reason, evidence, actor string) (string, error) {
	if amount.Sign() < 0 || reason == "" || evidence == "" || actor == "" {
		return "", errors.New("libra: invalid dispute")
	}
	id := uuid.NewString()
	res, e := s.db.ExecContext(ctx, `INSERT INTO finance_disputes(id,tenant_id,fact_id,amount,reason,evidence_id,created_by) SELECT $1,$2,id,$4,$5,$6,$7 FROM economic_facts WHERE tenant_id=$2 AND id=$3`, id, tenant, fact, string(amount), reason, evidence, actor)
	if e != nil {
		return "", e
	}
	n, e := res.RowsAffected()
	if e != nil {
		return "", e
	}
	if n != 1 {
		return "", ErrNotFound
	}
	return id, nil
}
