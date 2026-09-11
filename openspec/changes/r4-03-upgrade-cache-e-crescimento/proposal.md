# Proposal: Upgrade, cache e crescimento
## Change ID
r4-03-upgrade-cache-e-crescimento
## Status
Em execução — upgrade/migração e os cenários focados de cache/coordenação foram
qualificados; integração regional do cofre e crescimento em produção ainda
permanecem parciais.
## Why
Tokens deixaram Redis e lock passou a ser por chave. A prova atual confirma L1
válido durante indisponibilidade do resolver, isolamento por binding/versão,
limite de locks e cancelamento durante espera; a validação de cofre/AWS regional
e crescimento sob tráfego produtivo ainda não foi realizada.
0002_provider_auth.sql já existente na R2 foi alterada para API_KEY e header. O runner checksum-guardado para em 0002 de um banco previamente migrado, antes de executar a nova 0004. Comparação dos bytes/hashes comprova alteração; falha SQL integrada ainda não foi executada nesta auditoria.
Paginação eliminou recusa acima de 100, mas agrega todas as páginas em resources antes de filtrar aplicação/serviço. Portanto CPU/memória/round-trips continuam proporcionais ao total de ofertas do tenant e o control plane ainda é consultado por pedido.
## Context
Brownfield do commit b9d0f90ce02aa0c27cad546745153d160ff5867f. A baseline tem 189 requisitos/696 cenários.
## Problem
As lacunas observadas impedem contrato/custódia/qualificação integral.
## Goals
- L1 utilizável na falha de cofre e coordenação limitada
- Upgrade com migração histórica imutável
- Consulta de oferta seletiva sem materializar o portfólio
## Non-Goals
Reescrever stack, aprovar contrato comercial ou reduzir semântica de SLA.
## Users / Actors Impacted
Clientes, provedores e operadores nominais. Responsável funcional: Plataforma, Dados e Integrações.
## Scope
### In scope
Requisitos deste change e fechamento das obrigações herdadas vinculadas.
### Out of scope
Push/merge/deploy remoto e exclusão de dados existentes sem autorização.
## Product Requirements Summary
- R4-OPE-01: O Hub SHALL consultar cache válido por identidade/versionamento autorizado antes de buscar segredo remoto, respeitar expiração/revogação e coordenar renovação com espera cancelável por chave. Entradas de cache e estruturas de coordenação devem ter limites/evicção seguros sob rotação e crescimento, sem fallback para outro binding.
- R4-OPE-02: A entrega SHALL preservar migrações aplicadas e implementar mudanças por migrações aditivas. Qualificar instalação limpa e upgrade de volume com checksum R2 anterior sem apagar dados ou desabilitar verificação. Bancos já inicializados com variante modificada precisam reconciliação explícita e restrita a hashes/estados conhecidos, não atualização cega do ledger.
- R4-OPE-03: O Hub SHALL resolver pelo conjunto elegível indexado de tenant/aplicação/alvo/versão/vigência, sem materializar catálogo inteiro por pedido, e usar projeção válida conforme a baseline. A prova de crescimento deve medir round-trips/memória/latência e ambiguidade verdadeira, não apenas capacidade de encontrar registro 101.
## Business Rules
Custódia antes de ACK, UUIDv7 após aceite, identidade por recurso, resultado terminal único, snapshots e finanças exatas permanecem.
## Affected Capabilities
r4-03-upgrade-cache-e-crescimento
## Expected Impact
### Code
Fontes e responsabilidades em design/explore.
### Data
Migrações aditivas, recibos e evidência sem alteração cega de histórico.
### APIs / Contracts
Validação/autorização efetivas e versionamento compatível.
### Integrations
Provedor, callback, cofre e contratos existentes com fixtures externas.
### Operations
Um Compose do Hub ativo por vez; scripts, gates e runbooks verificáveis.
### Security / Privacy
Menor privilégio, escopo de custódia e evidência sem segredos.
## Risks and Mitigations
Ver risk-matrix.md.
## Success Criteria
Cenários passam com evidência atual e requisitos herdados correspondentes requalificados.
## Assumptions
Dados comerciais ausentes não impedem fixtures sintéticas; não equivalem a aprovação.
## Open Questions
D-01…D-07, P-01…P-11 e T-R2-01 continuam sujeitos ao estado real do registro; não presumir aprovação.
