-- Compatibility policy: migration owner remains usable until all services
-- propagate tenant context transactionally; hub_runtime stays fenced.
DROP POLICY IF EXISTS catalog_resources_tenant ON catalog_resources;
CREATE POLICY catalog_resources_tenant ON catalog_resources
  USING (current_user = 'hub' OR tenant_id = '' OR tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (current_user = 'hub' OR tenant_id = '' OR tenant_id = current_setting('app.tenant_id', true));
