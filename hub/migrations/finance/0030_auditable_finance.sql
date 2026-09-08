-- R2 additive migration. Legacy values retain exact NUMERIC provenance;
-- absent snapshots are never reconstructed from present-day tariffs.
ALTER TABLE economic_facts ALTER COLUMN amount TYPE NUMERIC(30,8);
ALTER TABLE reservations ALTER COLUMN amount TYPE NUMERIC(30,8);
ALTER TABLE credit_limits ALTER COLUMN limit_amount TYPE NUMERIC(30,8);
ALTER TABLE ledger_entries ALTER COLUMN amount TYPE NUMERIC(30,8);
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_state_check;
ALTER TABLE reservations ADD CONSTRAINT reservations_state_check CHECK(state IN ('RESERVED','CAPTURED','RELEASED','EXPIRED','UNCERTAIN_HOLD'));
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS evidence_id TEXT;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS unit_key TEXT;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS contract_id TEXT;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS contract_version INTEGER;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS evidence_id TEXT;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS operation_id TEXT;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS settlement_party TEXT;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS snapshot JSONB;
ALTER TABLE economic_facts ADD COLUMN IF NOT EXISTS provenance TEXT NOT NULL DEFAULT 'LEGACY_UNVERIFIED';
DROP INDEX IF EXISTS uq_economic_fact;
CREATE UNIQUE INDEX IF NOT EXISTS uq_economic_unit ON economic_facts(unit_key) WHERE unit_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS ix_finance_tenant_fact ON economic_facts(tenant_id,id);
CREATE INDEX IF NOT EXISTS ix_reservations_tenant ON reservations(tenant_id,currency,state);

CREATE TABLE IF NOT EXISTS finance_inbox(consumer TEXT NOT NULL,event_id TEXT NOT NULL,payload_hash TEXT NOT NULL,tenant_id TEXT NOT NULL,recorded_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),PRIMARY KEY(consumer,event_id));
CREATE TABLE IF NOT EXISTS finance_quarantine(id TEXT PRIMARY KEY,consumer TEXT NOT NULL,event_id TEXT NOT NULL,reason TEXT NOT NULL,payload_hash TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp());
CREATE TABLE IF NOT EXISTS finance_snapshots(tenant_id TEXT NOT NULL,protocol_id UUID NOT NULL,snapshot_hash TEXT NOT NULL,snapshot JSONB NOT NULL,PRIMARY KEY(tenant_id,protocol_id));
CREATE TABLE IF NOT EXISTS plan_consumption(tenant_id TEXT NOT NULL,plan_key TEXT NOT NULL,units BIGINT NOT NULL CHECK(units>=0),PRIMARY KEY(tenant_id,plan_key));
CREATE TABLE IF NOT EXISTS journal_batches(id TEXT PRIMARY KEY,tenant_id TEXT NOT NULL,unit_key TEXT UNIQUE,origin_batch_id TEXT REFERENCES journal_batches(id),reason TEXT,authorizer TEXT,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp());
ALTER TABLE ledger_entries ADD COLUMN IF NOT EXISTS tenant_id TEXT;
ALTER TABLE ledger_entries ADD COLUMN IF NOT EXISTS evidence_id TEXT;
ALTER TABLE ledger_entries ADD COLUMN IF NOT EXISTS contract_id TEXT;
CREATE INDEX IF NOT EXISTS ix_journal_tenant ON ledger_entries(tenant_id,id);
CREATE TABLE IF NOT EXISTS finance_adjustments(id TEXT PRIMARY KEY,tenant_id TEXT NOT NULL,origin_batch_id TEXT NOT NULL REFERENCES journal_batches(id),reason TEXT NOT NULL,prepared_by TEXT NOT NULL,approved_by TEXT,state TEXT NOT NULL CHECK(state IN ('PREPARED','APPROVED')),created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp());
CREATE TABLE IF NOT EXISTS finance_disputes(id TEXT PRIMARY KEY,tenant_id TEXT NOT NULL,fact_id BIGINT NOT NULL REFERENCES economic_facts(id),reason TEXT NOT NULL,evidence_id TEXT NOT NULL,amount NUMERIC(30,8) NOT NULL,created_by TEXT NOT NULL,state TEXT NOT NULL DEFAULT 'OPEN',created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp());
CREATE TABLE IF NOT EXISTS finance_watermarks(tenant_id TEXT NOT NULL,producer TEXT NOT NULL,complete_through TIMESTAMPTZ NOT NULL,evidence_id TEXT NOT NULL,PRIMARY KEY(tenant_id,producer));
CREATE TABLE IF NOT EXISTS settlement_periods(id TEXT PRIMARY KEY,tenant_id TEXT NOT NULL,period_start TIMESTAMPTZ NOT NULL,period_end TIMESTAMPTZ NOT NULL,layout_version INTEGER NOT NULL DEFAULT 1,export_body BYTEA NOT NULL,checksum TEXT NOT NULL,created_by TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),UNIQUE(tenant_id,period_start,period_end),CHECK(period_end>period_start));
CREATE TABLE IF NOT EXISTS export_receipts(export_id TEXT NOT NULL REFERENCES settlement_periods(id),receipt_id TEXT NOT NULL,checksum TEXT NOT NULL,actor TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),PRIMARY KEY(export_id,receipt_id));

CREATE OR REPLACE FUNCTION finance_immutable() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'confirmed financial evidence is append-only'; END $$;
DROP TRIGGER IF EXISTS finance_facts_immutable ON economic_facts;
CREATE TRIGGER finance_facts_immutable BEFORE UPDATE OR DELETE ON economic_facts FOR EACH ROW EXECUTE FUNCTION finance_immutable();
DROP TRIGGER IF EXISTS finance_ledger_immutable ON ledger_entries;
CREATE TRIGGER finance_ledger_immutable BEFORE UPDATE OR DELETE ON ledger_entries FOR EACH ROW EXECUTE FUNCTION finance_immutable();
DROP TRIGGER IF EXISTS finance_batches_immutable ON journal_batches;
CREATE TRIGGER finance_batches_immutable BEFORE UPDATE OR DELETE ON journal_batches FOR EACH ROW EXECUTE FUNCTION finance_immutable();
DROP TRIGGER IF EXISTS finance_snapshots_immutable ON finance_snapshots;
CREATE TRIGGER finance_snapshots_immutable BEFORE UPDATE OR DELETE ON finance_snapshots FOR EACH ROW EXECUTE FUNCTION finance_immutable();
DROP TRIGGER IF EXISTS finance_period_immutable ON settlement_periods;
CREATE TRIGGER finance_period_immutable BEFORE UPDATE OR DELETE ON settlement_periods FOR EACH ROW EXECUTE FUNCTION finance_immutable();

CREATE OR REPLACE FUNCTION finance_check_balanced() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE batch TEXT;
BEGIN
 batch:=NEW.batch_id;
 IF EXISTS(SELECT 1 FROM ledger_entries WHERE batch_id=batch GROUP BY currency HAVING SUM(CASE direction WHEN 'DEBIT' THEN amount ELSE -amount END)<>0) THEN RAISE EXCEPTION 'unbalanced journal batch'; END IF;
 RETURN NULL;
END $$;
DROP TRIGGER IF EXISTS finance_ledger_balance ON ledger_entries;
CREATE CONSTRAINT TRIGGER finance_ledger_balance AFTER INSERT ON ledger_entries DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION finance_check_balanced();
CREATE TABLE IF NOT EXISTS finance_action_audit(id BIGSERIAL PRIMARY KEY,tenant_id TEXT NOT NULL,actor TEXT NOT NULL,action TEXT NOT NULL,resource TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp());
DROP TRIGGER IF EXISTS finance_audit_immutable ON finance_action_audit;
CREATE TRIGGER finance_audit_immutable BEFORE UPDATE OR DELETE ON finance_action_audit FOR EACH ROW EXECUTE FUNCTION finance_immutable();
