# Delta for 02-portfolio-e-contratos-efetivos

## ADDED Requirements

### Requirement: R3-CAT-01 — Precisão numérica e validação de schemas
O Hub SHALL preservar valores numéricos exatos, validar integralmente o dialeto declarado e rejeitar publicação de construções de schema não suportadas. Transformação deve ter limites de profundidade/tamanho/custo e não executar código ou acessar rede/segredos.

#### Scenario: R3-CAT-01-S01 — inteiro 9007199254740993 e decimal exato
- GIVEN inteiro 9007199254740993 e decimal exato
- WHEN transformar entrada e saída
- THEN valores permanecem exatos

#### Scenario: R3-CAT-01-S02 — enum, nested schema ou additionalProperties violados
- GIVEN enum, nested schema ou additionalProperties violados
- WHEN validar payload
- THEN violação é recusada com caminho de campo ou schema não suportado é recusado na publicação

#### Scenario: R3-CAT-01-S03 — payload abusivo A e cliente B saudável
- GIVEN payload abusivo A e cliente B saudável
- WHEN transformar em paralelo
- THEN limite encerra A e B preserva orçamento de isolamento

### Requirement: R3-CAT-02 — Política efetiva e representação por cliente
O Hub SHALL definir precedência serviço→oferta→perfil dentro de limites contratuais, validar versão solicitada e congelar política efetiva/hash por aplicação. Representação final transformada deve ser persistida uma vez e reutilizada com bytes idênticos em GET e corpo de webhook.

#### Scenario: R3-CAT-02-S01 — mesmo serviço com dois perfis e SLAs
- GIVEN mesmo serviço com dois perfis e SLAs
- WHEN executar e consultar/notificar
- THEN cada cliente recebe seu contrato e hashes GET/webhook coincidem

#### Scenario: R3-CAT-02-S02 — versão não ofertada ou override inválido
- GIVEN versão não ofertada ou override inválido
- WHEN admitir
- THEN recusa ocorre antes do efeito com erro de elegibilidade

#### Scenario: R3-CAT-02-S03 — nova publicação durante execução
- GIVEN nova publicação durante execução
- WHEN finalizar protocolo anterior
- THEN snapshot e formato originais permanecem preservados

### Requirement: R3-CAT-03 — Agregação e composição com executor de DAG
O Hub SHALL executar DAG versionado com dependências, mapeamentos, paralelismo limitado, parcialidade, compensações e estado durável por etapa. Sucesso do produto depende do critério contratado; compensação não é rollback automático de efeito externo.

#### Scenario: R3-CAT-03-S01 — A/B independentes e C depende de ambos
- GIVEN A/B independentes e C depende de ambos
- WHEN executar produto
- THEN A/B sobrepõem intervalos e C recebe entradas mapeadas

#### Scenario: R3-CAT-03-S02 — ciclo ou contrato incompatível entre etapas
- GIVEN ciclo ou contrato incompatível entre etapas
- WHEN publicar
- THEN o grafo é recusado antes de habilitação

#### Scenario: R3-CAT-03-S03 — queda após A e falha B
- GIVEN queda após A e falha B
- WHEN retomar/compensar
- THEN A não repete efeito e compensações têm identidade e estado próprios

### Requirement: R3-CAT-04 — Resolução indexada e projeção disponível
O Hub SHALL resolver ofertas por chave/vigência indexada sem teto artificial sobre o portfólio e distribuir projeções versionadas ao data plane. Snapshot válido sustenta caminho quente durante falha do controle dentro da validade/revogação definidas; cache frio não autoriza contrato desconhecido.

#### Scenario: R3-CAT-04-S01 — tenant com 150 ofertas
- GIVEN tenant com 150 ofertas
- WHEN resolver oferta 150
- THEN oferta correta é localizada sem truncamento

#### Scenario: R3-CAT-04-S02 — Atlas e Redis indisponíveis com projeção válida
- GIVEN Atlas e Redis indisponíveis com projeção válida
- WHEN executar pedido elegível
- THEN caminho quente mantém orçamento publicado

#### Scenario: R3-CAT-04-S03 — snapshot vencido ou revogado
- GIVEN snapshot vencido ou revogado
- WHEN admitir novo pedido
- THEN recusa seletiva não usa contrato de outro cliente
