-- Leases boundedas para o reconciliador administrativo. A solicitação só
-- pode ser resolvida depois de evidência positiva do status externo.
ALTER TABLE protocol_reconciliation_requests
    ADD COLUMN IF NOT EXISTS claim_owner TEXT,
    ADD COLUMN IF NOT EXISTS claim_epoch BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_error TEXT;
CREATE INDEX IF NOT EXISTS ix_protocol_reconciliation_claimable
    ON protocol_reconciliation_requests(created_at, lease_until)
    WHERE state = 'OPEN';
