-- Claim durável para recuperação autônoma; não altera o corpo nem a
-- identidade de deduplicação dos callbacks já recebidos.
ALTER TABLE callback_inbox
    ADD COLUMN IF NOT EXISTS claim_owner TEXT,
    ADD COLUMN IF NOT EXISTS claim_epoch BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS processing_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_error TEXT;
CREATE INDEX IF NOT EXISTS ix_callback_inbox_claimable
    ON callback_inbox (received_at, lease_until)
    WHERE disposition = 'RECEIVED';
