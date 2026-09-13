-- R6-SEG-02 (S02): esgotamento de retry no apply de um callback JA
-- AUTENTICADO (nao um problema de assinatura/contrato) marcava a obrigacao
-- como 'REJECTED' -- o mesmo estado terminal usado para capability invalida.
-- Isso descartava uma obrigacao legitima apos uma falha transitoria (ex.:
-- indisponibilidade momentanea de uma dependencia), quando deveria continuar
-- recuperavel. 'QUARANTINED' e um estado distinto, nunca elegivel para
-- prune automatico (PruneCallbackInbox so remove APPLIED/REJECTED), e so sai
-- dele por replay explicitamente autorizado.
ALTER TABLE callback_inbox
    ADD COLUMN IF NOT EXISTS replayed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS replayed_by TEXT;
