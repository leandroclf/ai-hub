-- R6-FIN-01: um replay que ainda encontra payload inválido precisa
-- permanecer visível como pendente, com a tentativa registrada — nunca
-- promovido a REPLAYED só porque ProcessEnvelope não retornou erro (uma
-- requarentena bem-sucedida também não retorna erro).
ALTER TABLE finance_quarantine
    ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_attempted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_attempted_by TEXT;
