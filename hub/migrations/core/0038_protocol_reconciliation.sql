-- Solicitações administrativas de reconciliação não autorizam novo envio.
-- Elas preservam a intenção, o motivo e o operador para que um reconciliador
-- separado obtenha evidência positiva do provedor antes de resolver o efeito.
CREATE TABLE IF NOT EXISTS protocol_reconciliation_requests (
    request_id  UUID PRIMARY KEY,
    protocol_id UUID NOT NULL,
    tenant_id   TEXT NOT NULL,
    requested_by TEXT NOT NULL,
    reason      TEXT NOT NULL,
    state       TEXT NOT NULL DEFAULT 'OPEN'
                CHECK (state IN ('OPEN','RESOLVED','REJECTED')),
    evidence_id TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_protocol_reconciliation_open
    ON protocol_reconciliation_requests(tenant_id, protocol_id)
    WHERE state = 'OPEN';
CREATE INDEX IF NOT EXISTS ix_protocol_reconciliation_state
    ON protocol_reconciliation_requests(tenant_id, state, created_at);
