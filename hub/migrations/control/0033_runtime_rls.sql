-- Runtime role and tenant fence for control-plane data. The migration role
-- remains the owner; application connections use hub_runtime and must set
-- app.tenant_id inside a transaction before reading or writing tenant data.
DO $$ BEGIN
  CREATE ROLE hub_runtime LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS PASSWORD 'r2-runtime-fixture';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
GRANT USAGE ON SCHEMA public TO hub_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO hub_runtime;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO hub_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE hub IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO hub_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE hub IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO hub_runtime;

ALTER TABLE catalog_resources ENABLE ROW LEVEL SECURITY;
ALTER TABLE catalog_resources FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS catalog_resources_tenant ON catalog_resources;
CREATE POLICY catalog_resources_tenant ON catalog_resources
  USING (tenant_id = '' OR tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (tenant_id = '' OR tenant_id = current_setting('app.tenant_id', true));
