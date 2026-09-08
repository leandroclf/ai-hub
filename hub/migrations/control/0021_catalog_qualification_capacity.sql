-- Homologation evidence and capacity must be provisioned by the qualified lab or
-- authorized platform workflow, never inferred from an operator text field.
CREATE TABLE IF NOT EXISTS catalog_qualifications (
 id TEXT PRIMARY KEY, adapter_id TEXT NOT NULL, environment TEXT NOT NULL,
 evidence_ref TEXT NOT NULL, valid_until TIMESTAMPTZ NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('QUALIFIED','REVOKED'))
);
CREATE TABLE IF NOT EXISTS catalog_capacity_cells (
 cell_id TEXT PRIMARY KEY, environment TEXT NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('PROVISIONING','QUALIFYING','READY','DRAINING')),
 total_units INTEGER NOT NULL CHECK(total_units>=0), reserved_units INTEGER NOT NULL DEFAULT 0 CHECK(reserved_units>=0),
 qualification_ref TEXT NOT NULL, CHECK(reserved_units<=total_units)
);
CREATE TABLE IF NOT EXISTS catalog_onboardings (
 tenant_id TEXT PRIMARY KEY, cell_id TEXT NOT NULL, capacity_units INTEGER NOT NULL CHECK(capacity_units>0),
 state TEXT NOT NULL CHECK(state IN ('ACTIVE','PROVISIONING','BLOCKED')), reason TEXT NOT NULL,
 actor TEXT NOT NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TABLE IF NOT EXISTS catalog_provisioning_requests (
 request_id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, cell_id TEXT NOT NULL,
 required_units INTEGER NOT NULL, state TEXT NOT NULL DEFAULT 'REQUESTED',
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
