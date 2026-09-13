-- R6-SEG-01: remove o bypass nominal por role ('hub') herdado do compat de
-- 0034, pelo mesmo motivo descrito em migrations/control/0038: a aplicação
-- passa a conectar como hub_runtime (ver deploy/*.yml) e o compat deixa de
-- ser necessário.
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['economic_facts','reservations','credit_limits','ledger_entries','plan_consumption','finance_inbox','finance_snapshots','financial_disputes','settlement_batches'] LOOP
    IF to_regclass(t) IS NOT NULL AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=t AND column_name='tenant_id') THEN
      EXECUTE format('DROP POLICY IF EXISTS tenant_runtime ON %I', t);
      EXECUTE format('CREATE POLICY tenant_runtime ON %I USING (tenant_id = current_setting(''app.tenant_id'', true)) WITH CHECK (tenant_id = current_setting(''app.tenant_id'', true))', t);
    END IF;
  END LOOP;
END $$;

-- Acesso cruzado auditado (administrador nominal/financeiro): apenas
-- SELECT, apenas com motivo não vazio LOCAL à transação. Nunca cobre escrita.
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
