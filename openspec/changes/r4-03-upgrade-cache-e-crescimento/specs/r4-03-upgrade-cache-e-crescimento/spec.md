# Delta for r4-03-upgrade-cache-e-crescimento

## ADDED Requirements

### Requirement: R4-OPE-01 — L1 utilizável na falha de cofre e coordenação limitada
O Hub SHALL consultar cache válido por identidade/versionamento autorizado antes de buscar segredo remoto, respeitar expiração/revogação e coordenar renovação com espera cancelável por chave. Entradas de cache e estruturas de coordenação devem ter limites/evicção seguros sob rotação e crescimento, sem fallback para outro binding.

#### Scenario: R4-OPE-01-S01 — token válido e versão ainda autorizada em L1 com cofre fora
- GIVEN token válido e versão ainda autorizada em L1 com cofre fora
- WHEN autenticar
- THEN chamada usa L1 sem consulta remota obrigatória

#### Scenario: R4-OPE-01-S02 — token vencido/revogado ou binding distinto
- GIVEN token vencido/revogado ou binding distinto
- WHEN autenticar durante falha
- THEN escopo afetado é recusado sem usar credencial de outro cliente

#### Scenario: R4-OPE-01-S03 — muitas rotações e chamadas canceladas aguardando refresh
- GIVEN muitas rotações e chamadas canceladas aguardando refresh
- WHEN medir memória e duração
- THEN coordenação permanece limitada e cancelamento respeita budget

### Requirement: R4-OPE-02 — Upgrade com migração histórica imutável
A entrega SHALL preservar migrações aplicadas e implementar mudanças por migrações aditivas. Qualificar instalação limpa e upgrade de volume com checksum R2 anterior sem apagar dados ou desabilitar verificação. Bancos já inicializados com variante modificada precisam reconciliação explícita e restrita a hashes/estados conhecidos, não atualização cega do ledger.

#### Scenario: R4-OPE-02-S01 — banco R2 com checksum histórico e dados
- GIVEN banco R2 com checksum histórico e dados
- WHEN aplicar upgrade
- THEN migração avança preservando dados/ledger e API_KEY fica disponível

#### Scenario: R4-OPE-02-S02 — banco limpo e banco criado com variante R3 conhecida
- GIVEN banco limpo e banco criado com variante R3 conhecida
- WHEN executar instalação/convergência
- THEN ambos chegam ao schema alvo com proveniência e plano explícito

#### Scenario: R4-OPE-02-S03 — checksum desconhecido/adulterado
- GIVEN checksum desconhecido/adulterado
- WHEN executar migrador
- THEN processo falha antes de modificar schema; não troca checksum para forçar verde

### Requirement: R4-OPE-03 — Consulta de oferta seletiva sem materializar o portfólio
O Hub SHALL resolver pelo conjunto elegível indexado de tenant/aplicação/alvo/versão/vigência, sem materializar catálogo inteiro por pedido, e usar projeção válida conforme a baseline. A prova de crescimento deve medir round-trips/memória/latência e ambiguidade verdadeira, não apenas capacidade de encontrar registro 101.

#### Scenario: R4-OPE-03-S01 — milhares de ofertas não relacionadas e uma elegível
- GIVEN milhares de ofertas não relacionadas e uma elegível
- WHEN resolver o alvo
- THEN plano seletivo e uso de memória/round-trips seguem o budget definido

#### Scenario: R4-OPE-03-S02 — duas ofertas realmente elegíveis e vigentes
- GIVEN duas ofertas realmente elegíveis e vigentes
- WHEN resolver
- THEN ambiguidade é diagnosticada sem escolher arbitrariamente

#### Scenario: R4-OPE-03-S03 — Atlas fora com projeção válida versus projeção vencida
- GIVEN Atlas fora com projeção válida versus projeção vencida
- WHEN admitir
- THEN caminho quente autorizado continua e contrato vencido é recusado seletivamente
