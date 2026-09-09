DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['protocols','operations','command_intents','protocol_audit','polling_schedule','deliveries','webhook_destinations','file_refs','object_versions','object_pins'] LOOP
    IF to_regclass(t) IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=t AND column_name='tenant_id') THEN
      EXECUTE format('DROP POLICY IF EXISTS tenant_runtime ON %I', t);
      EXECUTE format('CREATE POLICY tenant_runtime ON %I USING (current_user = ''hub'' OR tenant_id = current_setting(''app.tenant_id'', true)) WITH CHECK (current_user = ''hub'' OR tenant_id = current_setting(''app.tenant_id'', true))', t);
    END IF;
  END LOOP;
END $$;
