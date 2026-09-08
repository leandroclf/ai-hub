CREATE TABLE IF NOT EXISTS message_quarantine (
 consumer text NOT NULL,
 body_sha256 text NOT NULL,
 body bytea NOT NULL,
 reason text NOT NULL,
 first_seen_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 last_seen_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 occurrences bigint NOT NULL DEFAULT 1,
 PRIMARY KEY(consumer,body_sha256)
);
