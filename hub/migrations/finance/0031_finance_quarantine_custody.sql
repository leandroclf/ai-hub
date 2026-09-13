-- Quarentena financeira precisa ser recuperável antes do ACK da fila. O
-- hash continua sendo a identidade; o corpo original permite replay
-- autorizado, auditoria e investigação sem reconstruir o evento.
ALTER TABLE finance_quarantine
    ADD COLUMN IF NOT EXISTS payload BYTEA,
    ADD COLUMN IF NOT EXISTS disposition TEXT NOT NULL DEFAULT 'QUARANTINED',
    ADD COLUMN IF NOT EXISTS replayed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS replayed_by TEXT;

ALTER TABLE finance_quarantine DROP CONSTRAINT IF EXISTS finance_quarantine_disposition_check;
ALTER TABLE finance_quarantine ADD CONSTRAINT finance_quarantine_disposition_check
    CHECK (disposition IN ('QUARANTINED','REPLAYED','DISCARDED'));
CREATE INDEX IF NOT EXISTS ix_finance_quarantine_pending
    ON finance_quarantine(consumer,created_at) WHERE disposition='QUARANTINED';
