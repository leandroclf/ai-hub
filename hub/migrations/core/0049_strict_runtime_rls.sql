-- Endurecimento aditivo: migrations históricas mantêm seus checksums, mas a
-- role de runtime não recebe bypass por nome. O contexto de tenant é sempre
-- transacional e a lista inclui as tabelas auxiliares que carregam escopo.
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['protocols','operations','command_intents','protocol_audit','polling_schedule','deliveries','webhook_destinations','operation_facts','callback_inbox','file_refs','object_versions','object_pins','operation_plans','operation_steps'] LOOP
    IF to_regclass(t) IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=t AND column_name='tenant_id') THEN
      EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
      EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
      EXECUTE format('DROP POLICY IF EXISTS tenant_runtime ON %I', t);
      EXECUTE format('CREATE POLICY tenant_runtime ON %I USING (tenant_id = current_setting(''app.tenant_id'', true)) WITH CHECK (tenant_id = current_setting(''app.tenant_id'', true))', t);
    END IF;
  END LOOP;
END $$;
