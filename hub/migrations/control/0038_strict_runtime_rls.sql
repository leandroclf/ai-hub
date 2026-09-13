-- R6-SEG-01: remove o bypass nominal por role ('hub') herdado do compat de
-- 0034. O migrador (role hub, superuser local) já ignora RLS por definição
-- de Postgres; o bypass explícito na policy só permitia que uma futura
-- conexão runtime reaproveitando esse nome herdasse acesso irrestrito. A
-- aplicação passa a conectar como hub_runtime (ver deploy/*.yml), então o
-- compat deixa de ser necessário.
DROP POLICY IF EXISTS catalog_resources_tenant ON catalog_resources;
CREATE POLICY catalog_resources_tenant ON catalog_resources
  USING (tenant_id = '' OR tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (tenant_id = '' OR tenant_id = current_setting('app.tenant_id', true));

-- Acesso cruzado auditado (administrador nominal): apenas SELECT, apenas com
-- motivo não vazio LOCAL à transação. Nunca cobre escrita.
DO $$ DECLARE t text; BEGIN
  FOR t IN
    SELECT c.relname FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public' AND c.relkind = 'r' AND c.relrowsecurity
  LOOP
    EXECUTE format('DROP POLICY IF EXISTS audited_scope ON %I', t);
    EXECUTE format('CREATE POLICY audited_scope ON %I FOR SELECT TO hub_runtime USING (current_setting(''app.access_reason'', true) <> '''')', t);
  END LOOP;
END $$;
