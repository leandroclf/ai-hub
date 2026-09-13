-- R6-SEG-01: achado do subagente que converteu internal/atlas — a policy
-- tenant_runtime de credential_bindings (0039) não tinha exceção para
-- tenant_id NULL, usado por bindings SHARED_HUB (credencial compartilhada
-- entre tenants, ver hub/internal/atlas/store.go ResolveCredential/
-- UpsertCredentialBinding). tenant_id = current_setting(...) nunca é
-- verdadeiro quando tenant_id é NULL, então essas linhas ficariam invisíveis
-- para leitura e toda escrita seria rejeitada pelo WITH CHECK, com qualquer
-- escopo de tenant. Mesmo padrão de catalog_resources (tenant_id = '' para
-- recursos globais), adaptado para NULL.
DROP POLICY IF EXISTS tenant_runtime ON credential_bindings;
CREATE POLICY tenant_runtime ON credential_bindings
  USING (tenant_id IS NULL OR tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (tenant_id IS NULL OR tenant_id = current_setting('app.tenant_id', true));
