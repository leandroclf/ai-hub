-- R6-SEG-02 (S02): callback_inbox_disposition_check só permitia
-- RECEIVED/APPLIED/REJECTED. A obrigação recuperável introduzida em
-- 0056 (QUARANTINED, distinta de REJECTED) violava essa constraint — o
-- INSERT/UPDATE falhava e o erro era descartado silenciosamente pelo
-- chamador (_ = dispose(...)), deixando a linha presa em RECEIVED com o
-- lease expirado sem nunca progredir. Amplia a constraint para incluir
-- QUARANTINED.
ALTER TABLE callback_inbox DROP CONSTRAINT IF EXISTS callback_inbox_disposition_check;
ALTER TABLE callback_inbox ADD CONSTRAINT callback_inbox_disposition_check
    CHECK (disposition = ANY (ARRAY['RECEIVED','APPLIED','REJECTED','QUARANTINED']));
