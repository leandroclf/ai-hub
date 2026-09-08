-- Capacity domains deliberately exclude cell, account and credential version:
-- every replica using the same upstream budget must use this one authority.
CREATE TABLE capacity_domains (
 domain_id text PRIMARY KEY,
 policy jsonb NOT NULL,
 policy_hash text NOT NULL,
 current_limit integer NOT NULL CHECK(current_limit > 0),
 epoch bigint NOT NULL DEFAULT 0,
 rate_started_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 rate_used integer NOT NULL DEFAULT 0,
 stable_since timestamptz NOT NULL DEFAULT clock_timestamp(),
 last_feedback text NOT NULL DEFAULT 'INITIAL',
 updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE TABLE capacity_permits (
 domain_id text NOT NULL REFERENCES capacity_domains(domain_id),
 permit_id text NOT NULL,
 tenant_id text NOT NULL,
 cell_id text NOT NULL,
 owner_id text NOT NULL,
 action text NOT NULL CHECK(action IN ('SUBMIT','STATUS','FETCH','CANCEL')),
 epoch bigint NOT NULL,
 lease_until timestamptz NOT NULL,
 transport_open boolean NOT NULL DEFAULT true,
 pending_external boolean NOT NULL,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 completed_at timestamptz,
 evidence_ref text,
 PRIMARY KEY(domain_id,permit_id)
);
CREATE INDEX capacity_permits_occupied ON capacity_permits(domain_id,tenant_id) WHERE transport_open OR pending_external;
CREATE TABLE capacity_feedback (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 domain_id text NOT NULL REFERENCES capacity_domains(domain_id),
 permit_id text NOT NULL,
 signal text NOT NULL,
 observed_latency_ms bigint NOT NULL,
 resulting_limit integer NOT NULL,
 evidence_ref text NOT NULL,
 occurred_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
