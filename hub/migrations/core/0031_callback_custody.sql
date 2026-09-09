-- Callback capabilities are random per operation. Only their SHA-256 digest
-- is durable, so a database read cannot be replayed as a provider callback.
ALTER TABLE operations ADD COLUMN IF NOT EXISTS callback_token_hash TEXT NOT NULL DEFAULT '';
