ALTER TABLE outbox ADD COLUMN IF NOT EXISTS event_id text;
ALTER TABLE outbox ADD COLUMN IF NOT EXISTS occurred_at timestamptz;
CREATE UNIQUE INDEX IF NOT EXISTS uq_outbox_event_id ON outbox(event_id) WHERE event_id IS NOT NULL;
