CREATE TABLE IF NOT EXISTS protocol_access_audit (
 id bigserial PRIMARY KEY,
 subject text NOT NULL,
 requested_tenant text NOT NULL,
 resource text NOT NULL,
 action text NOT NULL,
 mfa boolean NOT NULL,
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
