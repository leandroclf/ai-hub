-- Autenticação por conta do provedor para callbacks externos. A capability
-- legada permanece nas colunas existentes para compatibilidade; o fluxo novo
-- não grava segredo nem assinatura em URL.
ALTER TABLE callback_inbox
    ADD COLUMN IF NOT EXISTS provider_account_id TEXT,
    ADD COLUMN IF NOT EXISTS callback_auth_version TEXT NOT NULL DEFAULT 'legacy-capability',
    ADD COLUMN IF NOT EXISTS callback_timestamp TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS callback_signature TEXT;

CREATE INDEX IF NOT EXISTS ix_callback_inbox_account_auth
    ON callback_inbox (operation_id, provider_account_id, callback_auth_version)
    WHERE disposition = 'RECEIVED';
