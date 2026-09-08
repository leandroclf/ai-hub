ALTER TABLE operations ADD COLUMN IF NOT EXISTS tenant_id text NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE operations ADD COLUMN IF NOT EXISTS application_id text NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE operations ADD COLUMN IF NOT EXISTS cell_id text NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE operations ADD COLUMN IF NOT EXISTS command jsonb;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS command_hash text;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS result jsonb;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS evidence_id uuid;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS submit_owner text;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS submit_epoch bigint NOT NULL DEFAULT 0;
ALTER TABLE operations ADD COLUMN IF NOT EXISTS submit_lease_until timestamptz;
ALTER TABLE attempts ADD COLUMN IF NOT EXISTS prepared_at timestamptz NOT NULL DEFAULT clock_timestamp();
ALTER TABLE attempts ADD COLUMN IF NOT EXISTS owner text;
ALTER TABLE attempts ADD COLUMN IF NOT EXISTS epoch bigint;
CREATE TABLE IF NOT EXISTS operation_receipts (
 evidence_id uuid PRIMARY KEY,
 operation_id uuid NOT NULL,
 tenant_id text NOT NULL,
 source text NOT NULL,
 body jsonb NOT NULL,
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
