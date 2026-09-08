-- Órbita retains the observation before attempting terminal consolidation.
-- Broker redelivery retries RECEIVED rows; no acknowledgement precedes APPLIED.
CREATE TABLE orbita_fact_inbox (
 event_id text PRIMARY KEY,
 body_sha256 text NOT NULL,
 envelope bytea NOT NULL,
 protocol_id uuid NOT NULL REFERENCES protocols(protocol_id),
 tenant_id text NOT NULL,
 application_id text NOT NULL,
 cell_id text NOT NULL,
 disposition text NOT NULL DEFAULT 'RECEIVED' CHECK(disposition IN ('RECEIVED','APPLIED','OBSERVED')),
 received_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 processed_at timestamptz
);
CREATE INDEX ix_orbita_fact_pending ON orbita_fact_inbox(cell_id,received_at) WHERE disposition='RECEIVED';
