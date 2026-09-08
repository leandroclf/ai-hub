-- R2-CAT: authoritative versioned catalog; additive, preserving legacy identities.
CREATE TABLE IF NOT EXISTS catalog_resources (
    kind TEXT NOT NULL CHECK (kind IN ('clients','applications','services','products','offers','technical-profiles','providers','provider-accounts','credential-bindings','contracts','policies')),
    id TEXT NOT NULL,
    version INTEGER NOT NULL CHECK(version > 0),
    tenant_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'DRAFT' CHECK(state IN ('DRAFT','PUBLISHED','SUSPENDED')),
    revision BIGINT NOT NULL DEFAULT 1,
    data JSONB NOT NULL,
    content_hash TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY(kind,id,version)
);
CREATE INDEX IF NOT EXISTS catalog_tenant_page ON catalog_resources(kind,tenant_id,id,version);
CREATE INDEX IF NOT EXISTS catalog_state_page ON catalog_resources(kind,state,id,version);
CREATE TABLE IF NOT EXISTS catalog_publications (
    sequence BIGSERIAL PRIMARY KEY, kind TEXT NOT NULL, resource_id TEXT NOT NULL,
    resource_version INTEGER NOT NULL, revision BIGINT NOT NULL, content_hash TEXT NOT NULL,
    actor TEXT NOT NULL, reason TEXT NOT NULL, validation JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TABLE IF NOT EXISTS catalog_actions (
    actor TEXT NOT NULL, action_key TEXT NOT NULL, request_hash TEXT NOT NULL,
    response JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY(actor,action_key)
);
CREATE TABLE IF NOT EXISTS catalog_imports (
    id TEXT PRIMARY KEY, source_hash TEXT NOT NULL, source_name TEXT NOT NULL,
    actor TEXT NOT NULL, tenant_id TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'STAGED',
    items JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE(tenant_id,source_hash)
);
CREATE OR REPLACE FUNCTION catalog_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.state IN ('PUBLISHED','SUSPENDED') AND
    (NEW.data IS DISTINCT FROM OLD.data OR NEW.name IS DISTINCT FROM OLD.name OR
     NEW.tenant_id IS DISTINCT FROM OLD.tenant_id OR NEW.content_hash IS DISTINCT FROM OLD.content_hash) THEN
    RAISE EXCEPTION 'published catalog content is immutable' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS catalog_immutable_content ON catalog_resources;
CREATE TRIGGER catalog_immutable_content BEFORE UPDATE ON catalog_resources FOR EACH ROW EXECUTE FUNCTION catalog_immutable();
CREATE OR REPLACE FUNCTION legacy_service_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.published AND NEW IS DISTINCT FROM OLD THEN
    RAISE EXCEPTION 'published service is immutable; create a new version' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS legacy_service_immutable_content ON services;
CREATE TRIGGER legacy_service_immutable_content BEFORE UPDATE ON services FOR EACH ROW EXECUTE FUNCTION legacy_service_immutable();
-- Existing entries are inventory drafts: legacy publication is not homologation proof.
INSERT INTO catalog_resources(kind,id,version,name,data,author)
 SELECT 'services',code,version,description,jsonb_build_object('modes',modes,'client_sla_seconds',client_sla_seconds,
 'retry_ttl_seconds',retry_ttl_seconds,'input_schema',schema_input,'output_schema',schema_output,
 'provenance','legacy-service','legacy_published',published),'migration-0020' FROM services
 ON CONFLICT DO NOTHING;
