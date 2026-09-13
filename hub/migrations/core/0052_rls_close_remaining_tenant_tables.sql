-- R6-SEG-01: achado adicional descoberto ao converter o caminho real para
-- hub_runtime — 11 tabelas de hub_core carregam tenant_id mas nunca entraram
-- na lista de 0033/0049/0050 (provider_receipts, object_retention_pins,
-- object_tombstones, orbita_fact_inbox, object_actions, operation_receipts,
-- capacity_permits, webhook_destination_versions, object_gc_runs,
-- protocol_reconciliation_requests, webhook_audit). Sem RLS nelas, um bug de
-- WHERE no código continuaria vazando entre tenants mesmo depois do corte de
-- credencial. Fecha a mesma lacuna, mesmo padrão.
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['provider_receipts','object_retention_pins','object_tombstones','orbita_fact_inbox','object_actions','operation_receipts','capacity_permits','webhook_destination_versions','object_gc_runs','protocol_reconciliation_requests','webhook_audit'] LOOP
    IF to_regclass(t) IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=t AND column_name='tenant_id') THEN
      EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
      EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
      EXECUTE format('DROP POLICY IF EXISTS tenant_runtime ON %I', t);
      EXECUTE format('CREATE POLICY tenant_runtime ON %I USING (tenant_id = current_setting(''app.tenant_id'', true)) WITH CHECK (tenant_id = current_setting(''app.tenant_id'', true))', t);
    END IF;
  END LOOP;
END $$;

-- Reaplica audited_scope e worker_cell_scope (idempotente) para que também
-- cubram as tabelas recém-habilitadas acima, sem duplicar a lógica de 0050/0051.
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

DO $$ DECLARE t text; BEGIN
  FOR t IN
    SELECT c.relname FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public' AND c.relkind = 'r' AND c.relrowsecurity
      AND EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_schema='public' AND table_name=c.relname AND column_name='cell_id')
  LOOP
    EXECUTE format('DROP POLICY IF EXISTS worker_cell_scope ON %I', t);
    EXECUTE format('CREATE POLICY worker_cell_scope ON %I TO hub_runtime
      USING (current_setting(''app.worker_cell_id'', true) <> '''' AND cell_id = current_setting(''app.worker_cell_id'', true))
      WITH CHECK (current_setting(''app.worker_cell_id'', true) <> '''' AND cell_id = current_setting(''app.worker_cell_id'', true))', t);
  END LOOP;
END $$;
