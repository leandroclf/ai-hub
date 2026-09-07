-- hub_core (Orbita/Cometa/Pulsar) — DAD-01/DAD-02/DAD-03.
-- Cada dominio grava apenas seu proprio estado; referencias entre
-- tabelas desta base sao por ID (protocol_id/operation_id), nao FK
-- cruzando autoridade de escrita (Orbita decide protocolo, Cometa
-- decide operacao/tentativa, Pulsar decide entrega).

CREATE TABLE IF NOT EXISTS protocols (
    protocol_id        UUID PRIMARY KEY,
    tenant_id           TEXT NOT NULL,
    idempotency_key     TEXT NOT NULL,
    request_hash        TEXT NOT NULL,
    request_body        JSONB NOT NULL,
    mode                TEXT NOT NULL CHECK (mode IN ('SYNC','ASYNC','AUTO')),
    dispatch_mode       TEXT NOT NULL CHECK (dispatch_mode IN ('DIRECT','QUEUED')),
    command_id          UUID NOT NULL,
    status              TEXT NOT NULL DEFAULT 'ACCEPTED'
                         CHECK (status IN ('ACCEPTED','RUNNING','WAITING_PROVIDER','RECONCILING',
                                            'SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED')),
    result_version      INTEGER NOT NULL DEFAULT 0,
    final_body          JSONB,
    final_event_id      UUID,
    terminal_reason     TEXT,
    accepted_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_deadline_at  TIMESTAMPTZ NOT NULL,
    finalized_at        TIMESTAMPTZ,
    version             INTEGER NOT NULL DEFAULT 0,   -- concorrencia otimista (DAD-03)
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unicidade de idempotencia por tenant/operacao/chave (EXE-01, DAD-03).
CREATE UNIQUE INDEX IF NOT EXISTS uq_protocol_idempotency
    ON protocols (tenant_id, idempotency_key);

CREATE INDEX IF NOT EXISTS ix_protocol_tenant_status
    ON protocols (tenant_id, status, created_at);

-- Timers de deadline vencidos (EXE-11, DAD-08).
CREATE INDEX IF NOT EXISTS ix_protocol_open_deadline
    ON protocols (client_deadline_at)
    WHERE status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED');

CREATE TABLE IF NOT EXISTS operations (
    operation_id         UUID PRIMARY KEY,
    protocol_id           UUID NOT NULL,
    provider_account_id   TEXT NOT NULL,
    provider_request_id   TEXT,
    credential_binding_id TEXT,
    secret_version_id     TEXT,
    state                 TEXT NOT NULL DEFAULT 'PREPARED'
                           CHECK (state IN ('PREPARED','SUBMITTING','ACCEPTED_EXTERNAL','WAITING_FINAL','UNKNOWN',
                                             'SUCCEEDED','FAILED','CANCELLED')),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_operations_protocol ON operations (protocol_id);
CREATE INDEX IF NOT EXISTS ix_operations_provider_request
    ON operations (provider_account_id, provider_request_id);

CREATE TABLE IF NOT EXISTS attempts (
    attempt_id     UUID PRIMARY KEY,
    operation_id    UUID NOT NULL REFERENCES operations(operation_id),
    attempt_type    TEXT NOT NULL CHECK (attempt_type IN ('SUBMIT','STATUS','FETCH','CANCEL')),
    sent_at         TIMESTAMPTZ,
    received_at     TIMESTAMPTZ,
    error_code      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS polling_schedule (
    operation_id    UUID PRIMARY KEY REFERENCES operations(operation_id),
    next_run_at      TIMESTAMPTZ NOT NULL,
    interval_seconds INTEGER NOT NULL DEFAULT 5,
    max_interval_seconds INTEGER NOT NULL DEFAULT 60,
    attempts_count    INTEGER NOT NULL DEFAULT 0,
    deadline_at       TIMESTAMPTZ NOT NULL,
    lease_owner       TEXT,
    lease_expires_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS deliveries (
    delivery_id     UUID PRIMARY KEY,
    protocol_id      UUID NOT NULL,
    event_id         UUID NOT NULL,
    destination_url  TEXT NOT NULL,
    state            TEXT NOT NULL DEFAULT 'PENDING'
                      CHECK (state IN ('PENDING','DELIVERING','RETRY_SCHEDULED','DELIVERED','SUSPENDED','EXHAUSTED')),
    attempts_count   INTEGER NOT NULL DEFAULT 0,
    next_attempt_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_delivery_event_destination
    ON deliveries (event_id, destination_url);

CREATE INDEX IF NOT EXISTS ix_deliveries_state_next
    ON deliveries (state, next_attempt_at);

CREATE TABLE IF NOT EXISTS webhook_destinations (
    tenant_id     TEXT PRIMARY KEY,
    url            TEXT NOT NULL,
    hmac_secret    TEXT NOT NULL
);

-- Outbox transacional (COM-03): publicacao de fato na mesma transacao
-- do estado local; um relay separado publica no SNS/SQS e marca
-- published = true somente apos confirmacao do broker.
CREATE TABLE IF NOT EXISTS outbox (
    id              BIGSERIAL PRIMARY KEY,
    aggregate_type   TEXT NOT NULL,
    aggregate_id     TEXT NOT NULL,
    event_type       TEXT NOT NULL,
    payload          JSONB NOT NULL,
    published        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_outbox_pending ON outbox (published, created_at);

-- Inbox de deduplicacao de consumo (COM-03): registra event_id ja
-- aplicado por este dominio antes do ack ao broker.
CREATE TABLE IF NOT EXISTS inbox (
    consumer        TEXT NOT NULL,
    event_id         TEXT NOT NULL,
    processed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);
