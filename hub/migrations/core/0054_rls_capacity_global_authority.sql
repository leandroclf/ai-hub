-- R6-SEG-01: capacity_permits/capacity_domains implementam um orçamento
-- agregado por DOMÍNIO DE PROVEDOR — não por tenant nem por célula. O teste
-- TestPostgresCapacityAggregateReplicasCells prova que o mesmo domínio é
-- compartilhado entre células (o limite é do provedor/conta, não da
-- topologia interna). worker_cell_scope (0051, restrito a uma célula) e
-- tenant_runtime (restrito a um tenant) fragmentariam esse agregado
-- incorretamente. Esta policy é específica de capacity_permits: cobre todos
-- os comandos, exige um motivo auditável (mesma primitiva de audited_scope),
-- mas não restringe por célula — é a autoridade do próprio controlador de
-- capacidade, uma superfície muito mais estreita que um bypass geral.
DROP POLICY IF EXISTS capacity_authority ON capacity_permits;
CREATE POLICY capacity_authority ON capacity_permits TO hub_runtime
  USING (current_setting('app.access_reason', true) <> '')
  WITH CHECK (current_setting('app.access_reason', true) <> '');
