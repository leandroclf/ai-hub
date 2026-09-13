-- Compensações são obrigações da plataforma, não extensões do SLA público.
-- Elas podem continuar depois que a resposta do cliente expirou.
ALTER TABLE command_intents
    ADD COLUMN IF NOT EXISTS continue_after_client_deadline BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS ix_command_intents_compensation
    ON command_intents(protocol_id, continue_after_client_deadline)
    WHERE continue_after_client_deadline;
