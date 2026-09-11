-- A compensação é uma nova operação de negócio, não uma mutação silenciosa
-- do efeito original. O comando fica congelado no aceite e só vira intenção
-- quando uma falha exigir a política COMPENSATE.
ALTER TABLE operation_steps ADD COLUMN IF NOT EXISTS compensation_command JSONB;
ALTER TABLE operation_steps ADD COLUMN IF NOT EXISTS compensates_step_id TEXT;
CREATE INDEX IF NOT EXISTS ix_operation_steps_compensation
    ON operation_steps(protocol_id, compensates_step_id)
    WHERE compensates_step_id IS NOT NULL;
