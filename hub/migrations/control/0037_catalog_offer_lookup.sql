-- Índice do conjunto elegível; evita varredura/materialização do portfólio
-- para cada resolução de oferta.
CREATE INDEX IF NOT EXISTS catalog_offer_eligibility_lookup
    ON catalog_resources (tenant_id, state, (data->>'application_id'), (data->>'target_id'), id, version)
    WHERE kind='offers';
