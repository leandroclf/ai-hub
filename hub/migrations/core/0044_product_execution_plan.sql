-- Plano de produto e etapas duráveis (R2-CAT-04/R3-CAT-03).
-- O plano pertence à autoridade Orbita; cada etapa conserva seu próprio
-- command_id para que reentrega, retomada e observação externa sejam
-- idempotentes sem transformar o DAG em estado de memória.
CREATE TABLE IF NOT EXISTS operation_plans (
    protocol_id      UUID PRIMARY KEY,
    target_kind      TEXT NOT NULL CHECK (target_kind = 'products'),
    target_id        TEXT NOT NULL,
    target_version   INTEGER NOT NULL CHECK (target_version > 0),
    max_parallel     INTEGER NOT NULL CHECK (max_parallel BETWEEN 1 AND 5),
    allow_partial    BOOLEAN NOT NULL DEFAULT FALSE,
    consolidation    TEXT NOT NULL CHECK (consolidation = 'ALL_REQUIRED'),
    failure_policy   TEXT NOT NULL CHECK (failure_policy IN ('STOP','COMPENSATE')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE IF NOT EXISTS operation_steps (
    protocol_id              UUID NOT NULL,
    step_id                  TEXT NOT NULL,
    command_id               UUID NOT NULL,
    service_id               TEXT NOT NULL,
    service_version          INTEGER NOT NULL CHECK (service_version > 0),
    required                  BOOLEAN NOT NULL DEFAULT TRUE,
    depends_on               JSONB NOT NULL DEFAULT '[]'::jsonb,
    input_mapping            JSONB NOT NULL DEFAULT '{}'::jsonb,
    compensation_service_id  TEXT,
    state                    TEXT NOT NULL DEFAULT 'PENDING'
                             CHECK (state IN ('PENDING','READY','RUNNING','WAITING_PROVIDER',
                                               'SUCCEEDED','FAILED','SKIPPED','COMPENSATING',
                                               'COMPENSATED','COMPENSATION_FAILED')),
    command                  JSONB NOT NULL,
    result                   JSONB,
    error_code               TEXT,
    error_message            TEXT,
    evidence_id              UUID,
    version                  INTEGER NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (protocol_id, step_id),
    UNIQUE (protocol_id, command_id)
);

CREATE INDEX IF NOT EXISTS ix_operation_steps_ready
    ON operation_steps(protocol_id, state);

-- Etapas ainda não liberadas não são publicáveis pelo intent publisher.
ALTER TABLE command_intents DROP CONSTRAINT IF EXISTS command_intents_state_check;
ALTER TABLE command_intents ADD CONSTRAINT command_intents_state_check
    CHECK (state IN ('PENDING','READY','WAITING_RESERVATION','DELIVERED','EXPIRED'));
