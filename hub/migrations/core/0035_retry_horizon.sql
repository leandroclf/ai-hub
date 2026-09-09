-- A retry window starts at the first transient transport failure and is
-- immutable afterwards. It must survive process restart and lease takeover.
ALTER TABLE command_intents ADD COLUMN IF NOT EXISTS retry_started_at timestamptz;
ALTER TABLE command_intents ADD COLUMN IF NOT EXISTS retry_until timestamptz;
CREATE INDEX IF NOT EXISTS ix_intents_retry_horizon ON command_intents(state,retry_until) WHERE state='READY';
