-- Additive fencing columns required by the polling custody implementation.
-- Migration 0017 may already be applied in existing laboratories.
ALTER TABLE polling_schedule ADD COLUMN IF NOT EXISTS epoch bigint NOT NULL DEFAULT 0;
ALTER TABLE polling_schedule ADD COLUMN IF NOT EXISTS timeout_seconds integer NOT NULL DEFAULT 5 CHECK(timeout_seconds BETWEEN 1 AND 60);
ALTER TABLE polling_schedule ADD COLUMN IF NOT EXISTS jitter_percent integer NOT NULL DEFAULT 10 CHECK(jitter_percent BETWEEN 0 AND 50);
ALTER TABLE polling_schedule ADD COLUMN IF NOT EXISTS max_attempts integer NOT NULL DEFAULT 100 CHECK(max_attempts BETWEEN 1 AND 100000);
ALTER TABLE attempts ADD COLUMN IF NOT EXISTS owner text;
ALTER TABLE attempts ADD COLUMN IF NOT EXISTS epoch bigint;
ALTER TABLE attempts ADD COLUMN IF NOT EXISTS prepared_at timestamptz NOT NULL DEFAULT clock_timestamp();
CREATE INDEX IF NOT EXISTS polling_due_lease ON polling_schedule(next_run_at,lease_expires_at);
