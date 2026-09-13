-- R6-SEG-01: workers globais (varredura de fila por célula: claim/complete de
-- command_intents, contagem de operation_plans, liberação de operation_steps,
-- deadline sweep) processam múltiplos tenants por natureza — não são um
-- cliente pedindo dados de outro tenant. Em vez de reaproveitar o bypass do
-- migrador ou liberar acesso irrestrito, essa policy limita o worker a uma
-- única célula operacional por transação (app.worker_cell_id, LOCAL), que é
-- uma autoridade operacional específica (D-05/EXE-*: toda topologia já é
-- particionada por célula), não um tenant arbitrário fornecido pelo cliente.
-- Cobre todos os comandos (inclusive escrita), diferente de audited_scope
-- (somente leitura, para administrador nominal).
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
