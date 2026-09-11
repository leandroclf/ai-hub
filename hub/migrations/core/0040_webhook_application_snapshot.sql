-- R4: o destino de webhook pode ser global do tenant ou específico da
-- aplicação. A linha versionada continua imutável; o aceite copia sua
-- identidade para o comando/fato final e nunca consulta o destino corrente.
ALTER TABLE webhook_destination_versions
    ADD COLUMN IF NOT EXISTS application_id text NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS ix_webhook_destination_scope
    ON webhook_destination_versions(tenant_id, cell_id, application_id, id, version);
