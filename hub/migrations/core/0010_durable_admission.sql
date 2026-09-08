-- Historical rows retain their original evidence; no provider result backfill.
ALTER TABLE protocols ADD COLUMN IF NOT EXISTS application_id text NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE protocols ADD COLUMN IF NOT EXISTS cell_id text NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE protocols ADD COLUMN IF NOT EXISTS admitted_subject text;
ALTER TABLE protocols ADD COLUMN IF NOT EXISTS config_snapshot jsonb;
ALTER TABLE protocols ADD COLUMN IF NOT EXISTS final_representation bytea;
ALTER TABLE protocols ADD COLUMN IF NOT EXISTS final_sha256 text;
ALTER TABLE protocols ADD COLUMN IF NOT EXISTS final_media_type text;
CREATE UNIQUE INDEX IF NOT EXISTS uq_protocol_application_idempotency ON protocols(tenant_id,application_id,idempotency_key);
DROP INDEX IF EXISTS uq_protocol_idempotency;

CREATE TABLE IF NOT EXISTS command_intents (
 command_id uuid PRIMARY KEY,
 protocol_id uuid NOT NULL REFERENCES protocols(protocol_id),
 tenant_id text NOT NULL,
 application_id text NOT NULL,
 cell_id text NOT NULL,
 dispatch_mode text NOT NULL CHECK(dispatch_mode IN ('DIRECT','QUEUED')),
 command jsonb NOT NULL,
 state text NOT NULL CHECK(state IN ('READY','WAITING_RESERVATION','DELIVERED','EXPIRED')),
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 next_attempt_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 lease_owner text,
 lease_until timestamptz,
 epoch bigint NOT NULL DEFAULT 0,
 attempts integer NOT NULL DEFAULT 0,
 last_error text
);
CREATE INDEX IF NOT EXISTS ix_intents_ready ON command_intents(next_attempt_at) WHERE state='READY';

CREATE TABLE IF NOT EXISTS protocol_audit (
 id bigserial PRIMARY KEY,
 protocol_id uuid NOT NULL,
 tenant_id text NOT NULL,
 subject text NOT NULL,
 action text NOT NULL,
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 details jsonb NOT NULL DEFAULT '{}'
);
