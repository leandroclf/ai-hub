# Matriz de risco

| Risco | Prioridade | Mitigação | Responsável | Evidência exigida |
|---|---|---|---|---|
| Tokens deixaram Redis e lock passou a ser por chave. Contudo, Resolve é chamado antes do L1; prova com L1 válido e cofre indisponível falha. locks cresce sem remoção por binding/versão e mutex não respeita cancelamento durante espera. | P1 | R4-OPE-01 | Plataforma, Dados e Integrações | R4-OPE-01-S01/S02/S03 |
| 0002_provider_auth.sql já existente na R2 foi alterada para API_KEY e header. O runner checksum-guardado para em 0002 de um banco previamente migrado, antes de executar a nova 0004. Comparação dos bytes/hashes comprova alteração; falha SQL integrada ainda não foi executada nesta auditoria. | P0 | R4-OPE-02 | Plataforma, Dados e Integrações | R4-OPE-02-S01/S02/S03 |
| Paginação eliminou recusa acima de 100, mas agrega todas as páginas em resources antes de filtrar aplicação/serviço. Portanto CPU/memória/round-trips continuam proporcionais ao total de ofertas do tenant e o control plane ainda é consultado por pedido. | P1 | R4-OPE-03 | Plataforma, Dados e Integrações | R4-OPE-03-S01/S02/S03 |

Probabilidade quantitativa não medida. Evidência estática não é incidente observado.
