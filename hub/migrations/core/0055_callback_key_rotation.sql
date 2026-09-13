-- R6-SEG-02: um callback aceito deve sobreviver à rotação da chave de
-- ingresso. Hoje só existe uma chave (CALLBACK_INGRESS_KEY, sem id/versão) e
-- a reconciliação revalida contra a chave *atual*, não contra a que
-- originalmente autenticou o callback — rotacionar a env var invalidaria
-- reconciliação de callbacks legítimos ainda em trânsito. Persistimos qual
-- key_id verificou o callback e o instante da verificação; a reconciliação
-- volta a usar essa mesma chave (ver internal/callbackauth.KeyRing).
ALTER TABLE callback_inbox
    ADD COLUMN IF NOT EXISTS callback_key_id TEXT,
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ;
