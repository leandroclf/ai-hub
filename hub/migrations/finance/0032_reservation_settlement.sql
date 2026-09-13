-- Reserva e liquidação são valores distintos: o hold original nunca é
-- sobrescrito sem deixar o valor efetivamente capturado/liberado auditável.
ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS reserved_amount NUMERIC(30,8),
    ADD COLUMN IF NOT EXISTS captured_amount NUMERIC(30,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS released_amount NUMERIC(30,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS settlement_evidence TEXT;
UPDATE reservations SET reserved_amount=amount WHERE reserved_amount IS NULL;
ALTER TABLE reservations ALTER COLUMN reserved_amount SET NOT NULL;
CREATE TABLE IF NOT EXISTS reservation_settlements(
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    protocol_id UUID NOT NULL,
    reserved_amount NUMERIC(30,8) NOT NULL,
    captured_amount NUMERIC(30,8) NOT NULL DEFAULT 0,
    released_amount NUMERIC(30,8) NOT NULL DEFAULT 0,
    evidence_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE(protocol_id,evidence_id)
);
CREATE INDEX IF NOT EXISTS ix_reservation_settlements_tenant ON reservation_settlements(tenant_id,created_at);
