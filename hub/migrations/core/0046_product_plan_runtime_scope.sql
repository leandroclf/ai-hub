-- As tabelas de composição carregam a mesma cerca explícita das demais
-- projeções do core. O vínculo por protocol_id continua necessário, mas não
-- é usado como única defesa contra uma consulta fora de tenant/célula.
ALTER TABLE operation_plans ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE operation_plans ADD COLUMN IF NOT EXISTS cell_id TEXT NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE operation_steps ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT 'legacy-unverified';
ALTER TABLE operation_steps ADD COLUMN IF NOT EXISTS cell_id TEXT NOT NULL DEFAULT 'legacy-unverified';

DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['operation_plans','operation_steps'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_runtime ON %I', t);
    EXECUTE format('CREATE POLICY tenant_runtime ON %I USING (current_user = ''hub'' OR tenant_id = current_setting(''app.tenant_id'', true)) WITH CHECK (current_user = ''hub'' OR tenant_id = current_setting(''app.tenant_id'', true))', t);
  END LOOP;
END $$;
