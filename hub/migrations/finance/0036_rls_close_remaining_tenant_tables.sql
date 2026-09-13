-- R6-SEG-01: mesma lacuna descrita em migrations/core/0052 — tabelas
-- financeiras com tenant_id que nunca entraram na lista de RLS de
-- hub_finance (inclusive settlement_periods, journal_batches e
-- reservation_settlements).
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['finance_action_audit','finance_disputes','journal_batches','finance_adjustments','finance_watermarks','reservation_settlements','settlement_periods'] LOOP
    IF to_regclass(t) IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=t AND column_name='tenant_id') THEN
      EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
      EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
      EXECUTE format('DROP POLICY IF EXISTS tenant_runtime ON %I', t);
      EXECUTE format('CREATE POLICY tenant_runtime ON %I USING (tenant_id = current_setting(''app.tenant_id'', true)) WITH CHECK (tenant_id = current_setting(''app.tenant_id'', true))', t);
    END IF;
  END LOOP;
END $$;

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
