CREATE TABLE IF NOT EXISTS webhook_destination_versions (
 id uuid NOT NULL,
 version integer NOT NULL CHECK(version>0),
 tenant_id text NOT NULL,
 cell_id text NOT NULL,
 url text NOT NULL,
 secret_ref text NOT NULL,
 secret_version text NOT NULL,
 max_attempts integer NOT NULL CHECK(max_attempts BETWEEN 1 AND 20),
 timeout_seconds integer NOT NULL CHECK(timeout_seconds BETWEEN 1 AND 15),
 state text NOT NULL CHECK(state IN ('ACTIVE','SUSPENDED')),
 created_by text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 PRIMARY KEY(id,version)
);
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS tenant_id text NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS cell_id text NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS destination_id uuid;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS destination_version integer;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS representation bytea;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS body_sha256 text;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS lease_owner text;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS lease_until timestamptz;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS epoch bigint NOT NULL DEFAULT 0;
DROP INDEX IF EXISTS uq_delivery_event_destination;
CREATE UNIQUE INDEX IF NOT EXISTS uq_delivery_event_version ON deliveries(event_id,destination_id,destination_version);
CREATE TABLE IF NOT EXISTS webhook_attempts (
 id uuid PRIMARY KEY,
 delivery_id uuid NOT NULL REFERENCES deliveries(delivery_id),
 owner text NOT NULL,
 epoch bigint NOT NULL,
 state text NOT NULL CHECK(state IN ('PREPARED','ACKNOWLEDGED','FAILED','UNKNOWN')),
 prepared_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 received_at timestamptz,
 http_status integer,
 error_code text,
 body_sha256 text NOT NULL
);
CREATE TABLE IF NOT EXISTS webhook_audit (
 id bigserial PRIMARY KEY,tenant_id text NOT NULL,subject text NOT NULL,action text NOT NULL,resource text NOT NULL,reason text NOT NULL,recorded_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
