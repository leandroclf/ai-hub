-- JSONB normaliza ordem e espaçamento. Recibo bruto precisa preservar os
-- bytes recebidos para que body_sha256 seja um oráculo independente.
ALTER TABLE provider_receipts
  ALTER COLUMN body TYPE BYTEA
  USING convert_to(body::text, 'UTF8');
