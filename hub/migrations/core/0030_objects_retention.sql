-- FileRef remains stable; temporary access credentials never enter result bytes.
CREATE TABLE IF NOT EXISTS object_retention_policies(class TEXT NOT NULL,region TEXT NOT NULL,purpose TEXT NOT NULL,retention_seconds BIGINT NOT NULL CHECK(retention_seconds>0),max_bytes BIGINT NOT NULL CHECK(max_bytes>0 AND max_bytes<=1073741824),allowed_types JSONB NOT NULL,approved_by TEXT NOT NULL,PRIMARY KEY(class,region,purpose));
CREATE TABLE IF NOT EXISTS file_refs(id UUID PRIMARY KEY,tenant_id TEXT NOT NULL,object_key TEXT NOT NULL UNIQUE,object_version TEXT,sha256 TEXT NOT NULL,size_bytes BIGINT NOT NULL CHECK(size_bytes>0),content_type TEXT NOT NULL,purpose TEXT NOT NULL,class TEXT NOT NULL,region TEXT NOT NULL,state TEXT NOT NULL CHECK(state IN ('UPLOADING','VALIDATING','READY','ORPHAN','PURGING','PURGED','REJECTED')),upload_id TEXT,expires_at TIMESTAMPTZ NOT NULL,retention_until TIMESTAMPTZ NOT NULL,obligation_id TEXT,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),UNIQUE(tenant_id,id));
CREATE INDEX IF NOT EXISTS ix_object_retention ON file_refs(state,retention_until,id);
CREATE INDEX IF NOT EXISTS ix_object_tenant ON file_refs(tenant_id,id);
CREATE TABLE IF NOT EXISTS object_retention_pins(file_id UUID NOT NULL REFERENCES file_refs(id),tenant_id TEXT NOT NULL,obligation_id TEXT NOT NULL,reason TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),PRIMARY KEY(file_id,obligation_id));
CREATE TABLE IF NOT EXISTS object_tombstones(file_id UUID PRIMARY KEY,tenant_id TEXT NOT NULL,sha256 TEXT NOT NULL,purged_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),reason TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS object_actions(id BIGSERIAL PRIMARY KEY,tenant_id TEXT NOT NULL,file_id UUID NOT NULL,action TEXT NOT NULL,actor TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp());
CREATE TABLE IF NOT EXISTS object_gc_runs(id UUID PRIMARY KEY,tenant_id TEXT NOT NULL,actor TEXT NOT NULL,candidates JSONB NOT NULL,expires_at TIMESTAMPTZ NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp());
CREATE OR REPLACE FUNCTION object_protect_content() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
 IF OLD.state IN ('READY','ORPHAN','PURGING','PURGED') AND (NEW.tenant_id,NEW.object_key,NEW.object_version,NEW.sha256,NEW.size_bytes,NEW.content_type,NEW.purpose,NEW.class,NEW.region) IS DISTINCT FROM (OLD.tenant_id,OLD.object_key,OLD.object_version,OLD.sha256,OLD.size_bytes,OLD.content_type,OLD.purpose,OLD.class,OLD.region) THEN RAISE EXCEPTION 'confirmed file reference is immutable'; END IF;
 RETURN NEW;END $$;
DROP TRIGGER IF EXISTS object_content_immutable ON file_refs;
CREATE TRIGGER object_content_immutable BEFORE UPDATE ON file_refs FOR EACH ROW EXECUTE FUNCTION object_protect_content();
CREATE OR REPLACE FUNCTION object_audit_immutable() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'object custody evidence is append-only'; END $$;
DROP TRIGGER IF EXISTS object_tombstone_immutable ON object_tombstones;
CREATE TRIGGER object_tombstone_immutable BEFORE UPDATE OR DELETE ON object_tombstones FOR EACH ROW EXECUTE FUNCTION object_audit_immutable();
DROP TRIGGER IF EXISTS object_action_immutable ON object_actions;
CREATE TRIGGER object_action_immutable BEFORE UPDATE OR DELETE ON object_actions FOR EACH ROW EXECUTE FUNCTION object_audit_immutable();
