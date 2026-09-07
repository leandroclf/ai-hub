-- hub_finance (Libra) — DAD-01/DAD-02, FIN-04/FIN-06/FIN-07.
-- Fatos de uso imutaveis e reservas financeiras. Nenhuma tabela aqui
-- referencia hub_core por FK: apenas por protocol_id/operation_id.

CREATE TABLE IF NOT EXISTS economic_facts (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id        TEXT NOT NULL,
    protocol_id      UUID NOT NULL,
    kind             TEXT NOT NULL CHECK (kind IN ('REVENUE','COST')),
    meter            TEXT NOT NULL,
    amount           NUMERIC(18,4) NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'BRL',
    occurred_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Dedup semantica alem do event_id (FIN-04): uma unidade elegivel por
-- protocolo/medidor/natureza, mesmo com reentrega do evento de origem.
CREATE UNIQUE INDEX IF NOT EXISTS uq_economic_fact
    ON economic_facts (protocol_id, kind, meter);

CREATE TABLE IF NOT EXISTS reservations (
    protocol_id      UUID PRIMARY KEY,
    tenant_id         TEXT NOT NULL,
    amount            NUMERIC(18,4) NOT NULL,
    currency          TEXT NOT NULL DEFAULT 'BRL',
    state             TEXT NOT NULL DEFAULT 'RESERVED'
                       CHECK (state IN ('RESERVED','CAPTURED','RELEASED','EXPIRED')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Limite estrito por tenant (FIN-06): teto financeiro que a reserva
-- atomica nao pode ultrapassar. Ausencia de linha equivale a um teto
-- alto por padrao neste ambiente de referencia (ver seed local); em
-- producao toda oferta com saldo estrito exigiria linha explicita
-- aprovada comercialmente (P-03).
CREATE TABLE IF NOT EXISTS credit_limits (
    tenant_id      TEXT PRIMARY KEY,
    limit_amount    NUMERIC(18,4) NOT NULL,
    currency        TEXT NOT NULL DEFAULT 'BRL'
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id               BIGSERIAL PRIMARY KEY,
    batch_id          TEXT NOT NULL,
    account           TEXT NOT NULL,
    direction         TEXT NOT NULL CHECK (direction IN ('DEBIT','CREDIT')),
    amount            NUMERIC(18,4) NOT NULL,
    currency          TEXT NOT NULL DEFAULT 'BRL',
    origin_protocol_id UUID,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
