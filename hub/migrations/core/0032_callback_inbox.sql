-- Callback recebido antes da correlação completa não pode ser descartado.
-- O corpo é limitado pelo handler e fica recuperável até reconciliação.
CREATE TABLE IF NOT EXISTS callback_inbox (
    inbox_id UUID PRIMARY KEY,
    operation_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    body_sha256 TEXT NOT NULL,
    body BYTEA NOT NULL,
    disposition TEXT NOT NULL DEFAULT 'RECEIVED'
        CHECK (disposition IN ('RECEIVED','APPLIED','REJECTED')),
    received_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    processed_at TIMESTAMPTZ,
    occurrences BIGINT NOT NULL DEFAULT 1,
    UNIQUE (operation_id, body_sha256)
);
CREATE INDEX IF NOT EXISTS ix_callback_inbox_pending
    ON callback_inbox (received_at) WHERE disposition = 'RECEIVED';
