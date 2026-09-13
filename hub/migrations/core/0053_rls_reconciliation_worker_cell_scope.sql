-- R6-SEG-01: protocol_reconciliation_requests não tem coluna cell_id própria
-- (a célula é sempre a da operação correlata), então o worker_cell_scope
-- genérico de 0051/0052 (que exige a coluna) não a cobre. O worker de
-- reconciliação administrativa (ClaimReconciliation) precisa escrever nela
-- entre tenants dentro de uma única célula, tanto quanto precisa em
-- operations. Deriva a célula por relação com operations (mesmo
-- protocol_id+tenant_id), como o próprio JOIN da consulta já faz.
DROP POLICY IF EXISTS worker_cell_scope ON protocol_reconciliation_requests;
CREATE POLICY worker_cell_scope ON protocol_reconciliation_requests TO hub_runtime
  USING (
    current_setting('app.worker_cell_id', true) <> ''
    AND EXISTS (
      SELECT 1 FROM operations o
      WHERE o.protocol_id = protocol_reconciliation_requests.protocol_id
        AND o.tenant_id = protocol_reconciliation_requests.tenant_id
        AND o.cell_id = current_setting('app.worker_cell_id', true)
    )
  )
  WITH CHECK (
    current_setting('app.worker_cell_id', true) <> ''
    AND EXISTS (
      SELECT 1 FROM operations o
      WHERE o.protocol_id = protocol_reconciliation_requests.protocol_id
        AND o.tenant_id = protocol_reconciliation_requests.tenant_id
        AND o.cell_id = current_setting('app.worker_cell_id', true)
    )
  );
