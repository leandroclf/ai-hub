# Risk matrix

| Risco | Prioridade | Mitigação | Prova | Responsável |
|---|---|---|---|---|
| Leitura administrativa no mesmo tenant aceita protocols:read sem papel administrativo específico/MFA e lista todas as aplicações. Entre tenants já existe verificação MFA/papel; não se trata de afirmar bypass global irrestrito. | P0 | R3-ADM-01 | R3-ADM-01-S01/S02/S03 | Frontend e Segurança |
| Seleção usa protocol_id antes de delivery_id. Menu SLA aponta para rota não registrada; botão de reconciliação POST aponta para handler de leitura. | P1 | R3-ADM-02 | R3-ADM-02-S01/S02/S03 | Frontend e Segurança |
| UI envia tenant_id/prepared_by a decoder estrito incompatível e datas YYYY-MM-DD para time.Time. Ações não mantêm tenant selecionado na query e criam idempotency key nova em cada execução. | P1 | R3-ADM-03 | R3-ADM-03-S01/S02/S03 | Frontend e Segurança |
| Editor de modos reduz array a seleção única; faltam campos condicionais de autenticação; lookups limitados à primeira centena. datetime-local corta UTC e o reinterpreta no fuso local ao salvar. | P1 | R3-ADM-04 | R3-ADM-04-S01/S02/S03 | Frontend e Segurança |

P0 bloqueia qualificação de uso; P1 bloqueia conclusão do escopo. Probabilidade quantitativa não foi medida.
