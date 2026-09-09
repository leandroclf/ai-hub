-- Runtime role and tenant fence for core data. app.tenant_id is deliberately
-- transaction scoped; an unset context matches no tenant row.
DO $$ BEGIN
  CREATE ROLE hub_runtime LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS PASSWORD 'r2-runtime-fixture';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
GRANT USAGE ON SCHEMA public TO hub_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO hub_runtime;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO hub_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE hub IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO hub_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE hub IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO hub_runtime;

DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['protocols','operations','command_intents','protocol_audit','polling_schedule','deliveries','webhook_destinations','operation_facts','callback_inbox','file_refs','object_versions','object_pins'] LOOP
    IF to_regclass(t) IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=t AND column_name='tenant_id') THEN
      EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
      EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
      EXECUTE format('DROP POLICY IF EXISTS tenant_runtime ON %I', t);
      EXECUTE format('CREATE POLICY tenant_runtime ON %I USING (tenant_id = current_setting(''app.tenant_id'', true)) WITH CHECK (tenant_id = current_setting(''app.tenant_id'', true))', t);
    END IF;
  END LOOP;
END $$;
