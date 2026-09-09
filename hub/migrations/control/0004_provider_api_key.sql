-- API Key é configuração de transporte; o valor continua exclusivamente no cofre.
ALTER TABLE provider_accounts
    DROP CONSTRAINT IF EXISTS provider_accounts_auth_type_check;

ALTER TABLE provider_accounts
    ADD COLUMN IF NOT EXISTS api_key_header TEXT NOT NULL DEFAULT '';

ALTER TABLE provider_accounts
    ADD CONSTRAINT provider_accounts_auth_type_check
    CHECK (auth_type IN ('NONE','BASIC','API_KEY','OAUTH_CLIENT_CREDENTIALS','MTLS_OAUTH'));
