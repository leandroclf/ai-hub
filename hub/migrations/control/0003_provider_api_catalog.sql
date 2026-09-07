-- Catálogo de endpoints importados de collections externas.
-- Apenas metadados sanitizados são persistidos: valores de headers, bodies
-- e credenciais não entram nesta tabela.
CREATE TABLE IF NOT EXISTS provider_api_catalog (
    api_id              TEXT PRIMARY KEY,
    product_code        TEXT NOT NULL,
    product_version     INTEGER NOT NULL DEFAULT 1,
    provider_account_id TEXT NOT NULL REFERENCES provider_accounts(provider_account_id) ON DELETE CASCADE,
    source_collection   TEXT NOT NULL,
    source_folder       TEXT NOT NULL,
    request_name        TEXT NOT NULL,
    http_method         TEXT NOT NULL CHECK (http_method IN ('GET','POST','PUT','PATCH','DELETE','OPTIONS','HEAD')),
    path_template       TEXT NOT NULL,
    auth_type           TEXT NOT NULL DEFAULT 'NONE' CHECK (auth_type IN ('NONE','BASIC','BEARER','API_KEY')),
    request_metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider_account_id, source_collection, source_folder, request_name)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_provider_api_product'
    ) THEN
        ALTER TABLE provider_api_catalog
            ADD CONSTRAINT fk_provider_api_product
            FOREIGN KEY (product_code, product_version) REFERENCES services(code, version) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_provider_api_catalog_product ON provider_api_catalog (product_code);
CREATE INDEX IF NOT EXISTS idx_provider_api_catalog_provider ON provider_api_catalog (provider_account_id);

-- Collections podem conter cenários homônimos dentro do mesmo grupo; api_id é
-- a identidade importada e permite manter todas as variações.
ALTER TABLE provider_api_catalog
    DROP CONSTRAINT IF EXISTS provider_api_catalog_provider_account_id_source_collection__key;
