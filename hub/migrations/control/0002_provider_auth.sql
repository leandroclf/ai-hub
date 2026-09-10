-- Perfil de autenticacao outbound por conta de provedor.
-- Segredos permanecem como referencias; valores em claro nao sao persistidos.
ALTER TABLE provider_accounts
    ADD COLUMN IF NOT EXISTS auth_type TEXT NOT NULL DEFAULT 'NONE'
        CHECK (auth_type IN ('NONE','BASIC','OAUTH_CLIENT_CREDENTIALS','MTLS_OAUTH')),
    ADD COLUMN IF NOT EXISTS auth_username TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS auth_secret_ref TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS oauth_token_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS oauth_client_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS oauth_client_secret_ref TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mtls_certificate_ref TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS token_ttl_seconds INTEGER NOT NULL DEFAULT 300
        CHECK (token_ttl_seconds > 0);
