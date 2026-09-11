-- Recibo bruto do provedor separado do resultado normalizado (R4-CTR-03).
CREATE TABLE IF NOT EXISTS provider_receipts(
 evidence_id UUID PRIMARY KEY,
 operation_id UUID NOT NULL,
 tenant_id TEXT NOT NULL,
 source TEXT NOT NULL,
 provider_request_id TEXT NOT NULL,
 body JSONB NOT NULL,
 body_sha256 TEXT NOT NULL,
 recorded_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX IF NOT EXISTS ix_provider_receipts_operation
 ON provider_receipts(tenant_id,operation_id,recorded_at);
CREATE OR REPLACE FUNCTION provider_receipt_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'provider receipt custody is append-only';
END $$;
DROP TRIGGER IF EXISTS provider_receipt_immutable ON provider_receipts;
CREATE TRIGGER provider_receipt_immutable BEFORE UPDATE OR DELETE ON provider_receipts
 FOR EACH ROW EXECUTE FUNCTION provider_receipt_immutable();
