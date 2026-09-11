-- A capability inválida não pode ocupar a identidade de um callback legítimo
-- que carregue os mesmos bytes. O hash da capability faz parte da
-- deduplicação; o valor em claro nunca é persistido.
ALTER TABLE callback_inbox
    DROP CONSTRAINT IF EXISTS callback_inbox_operation_id_body_sha256_key;

ALTER TABLE callback_inbox
    ADD CONSTRAINT callback_inbox_body_size_check
    CHECK (octet_length(body) BETWEEN 1 AND 524288);

CREATE UNIQUE INDEX IF NOT EXISTS ux_callback_inbox_capability_identity
    ON callback_inbox (operation_id, body_sha256, token_hash);

CREATE INDEX IF NOT EXISTS ix_callback_inbox_retention
    ON callback_inbox (processed_at)
    WHERE disposition IN ('APPLIED', 'REJECTED');
