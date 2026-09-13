-- R6-SEG-01: acesso cruzado de tenant para administrador nominal e workers
-- globais (ex.: consulta interna por protocol_id, varredura de deadlines por
-- celula) passa a exigir uma policy adicional, restrita a SELECT, que só
-- libera linhas quando app.access_reason (LOCAL a transação) é não vazio.
-- Sem motivo auditado, nenhuma linha de outro tenant fica visível — não há
-- bypass implícito por nome de role. Escrita cruzada continua vedada: esta
-- policy nunca cobre INSERT/UPDATE/DELETE.
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
