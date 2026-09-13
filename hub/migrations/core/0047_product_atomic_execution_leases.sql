-- A capacidade de produto pertence ao plano, não ao instante em que um
-- publisher leu uma contagem. O contador e o lease da etapa são atualizados
-- na mesma transação da reivindicação do intent.
ALTER TABLE operation_plans
    ADD COLUMN IF NOT EXISTS running_count INTEGER NOT NULL DEFAULT 0,
    ADD CONSTRAINT operation_plans_running_count_check CHECK (running_count >= 0);

UPDATE operation_plans p
SET running_count = q.running_count
FROM (
    SELECT protocol_id, count(*)::integer AS running_count
    FROM operation_steps
    WHERE state IN ('RUNNING','WAITING_PROVIDER')
    GROUP BY protocol_id
) q
WHERE p.protocol_id = q.protocol_id;

ALTER TABLE operation_steps
    ADD COLUMN IF NOT EXISTS lease_owner TEXT,
    ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS lease_epoch BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS dispatch_started_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS ix_operation_steps_lease
    ON operation_steps(protocol_id, state, lease_until);
