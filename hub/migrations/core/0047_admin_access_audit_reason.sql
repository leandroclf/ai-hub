ALTER TABLE protocol_access_audit
  ADD COLUMN IF NOT EXISTS reason text NOT NULL DEFAULT '';

ALTER TABLE protocol_access_audit
  ADD CONSTRAINT protocol_access_audit_reason_length
  CHECK (char_length(reason) <= 512);
