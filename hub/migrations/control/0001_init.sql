-- hub_control (Atlas) — DAD-01/DAD-02.
-- Catalogo de servicos, contratos e vinculos de credencial. Nenhuma
-- tabela aqui referencia hub_core/hub_finance por FK: apenas por ID.

CREATE TABLE IF NOT EXISTS services (
    code            TEXT NOT NULL,
    version         INTEGER NOT NULL,
    description     TEXT NOT NULL,
    schema_input    JSONB NOT NULL,
    schema_output   JSONB NOT NULL,
    modes           TEXT[] NOT NULL,              -- SYNC, ASYNC, AUTO
    client_sla_seconds INTEGER NOT NULL CHECK (client_sla_seconds > 0),
    retry_ttl_seconds  INTEGER NOT NULL DEFAULT 0 CHECK (retry_ttl_seconds >= 0),
    published       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (code, version)
);

CREATE TABLE IF NOT EXISTS provider_accounts (
    provider_account_id TEXT PRIMARY KEY,
    provider_id          TEXT NOT NULL,
    environment          TEXT NOT NULL DEFAULT 'local',
    base_url              TEXT NOT NULL,
    -- Natureza declarada do provedor (CAT-11): sync (elegivel a SYNC
    -- direto), async_poll ou async_callback. Nao e escolhida pelo
    -- comando, e propriedade homologada da conta.
    provider_mode          TEXT NOT NULL DEFAULT 'sync'
                            CHECK (provider_mode IN ('sync','async_poll','async_callback')),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS credential_bindings (
    binding_id       TEXT PRIMARY KEY,
    credential_mode  TEXT NOT NULL CHECK (credential_mode IN ('SHARED_HUB', 'TENANT_DEDICATED')),
    tenant_id        TEXT,                         -- obrigatorio e unico quando dedicado
    provider_account_id TEXT NOT NULL REFERENCES provider_accounts(provider_account_id),
    secret_ref        TEXT NOT NULL,               -- referencia ao cofre; nunca o segredo em si
    settlement_party  TEXT NOT NULL CHECK (settlement_party IN ('HUB', 'CLIENT_DIRECT')),
    state             TEXT NOT NULL DEFAULT 'ATIVO'
                       CHECK (state IN ('RASCUNHO','EM_VALIDACAO','ATIVO','EM_ROTACAO','SUSPENSO','REVOGADO','EXPIRADO')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_credential_dedicated
    ON credential_bindings (tenant_id, provider_account_id)
    WHERE credential_mode = 'TENANT_DEDICATED';

CREATE TABLE IF NOT EXISTS contracts (
    tenant_id            TEXT PRIMARY KEY,
    plan                  TEXT NOT NULL DEFAULT 'unit',
    unit_price            NUMERIC(18,4) NOT NULL DEFAULT 1.0000,
    strict_balance        BOOLEAN NOT NULL DEFAULT FALSE,
    client_sla_seconds    INTEGER NOT NULL DEFAULT 30 CHECK (client_sla_seconds > 0),
    -- Modalidade de credencial exigida por este tenant (CFG-05, SEG-05):
    -- decide se o Cometa pode usar SHARED_HUB ou exige TENANT_DEDICATED.
    -- Nunca inferida pela ausencia de vinculo.
    credential_mode_required TEXT NOT NULL DEFAULT 'SHARED_HUB'
                              CHECK (credential_mode_required IN ('SHARED_HUB','TENANT_DEDICATED')),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS placements (
    tenant_id    TEXT PRIMARY KEY,
    cell_id       TEXT NOT NULL DEFAULT 'cell-local-1',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
