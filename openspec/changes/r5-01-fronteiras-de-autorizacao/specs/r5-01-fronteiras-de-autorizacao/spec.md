# Delta for r5-01-fronteiras-de-autorizacao

## ADDED Requirements

### Requirement: R5-SEG-01 — Autenticação antes de custódia de callback órfão
O Hub SHALL autenticar a origem antes de confirmar ou reservar custódia órfã. Cada recibo deve possuir escopo de origem autenticada, identidade estável de evento e disposição recuperável. Tentativas não autenticadas não podem consumir quota durável de obrigações válidas. Quotas, retenção e claims devem isolar contas/células; eventos irrelacionáveis devem chegar a disposição auditável sem bloqueio global.

#### Scenario: R5-SEG-01-S01 — operação desconhecida e assinatura falsa não vazia
- GIVEN operação desconhecida e assinatura falsa não vazia
- WHEN POST pela rota pública real
- THEN 401/403 sem inbox e sem consumo de quota de aceites

#### Scenario: R5-SEG-01-S02 — callback autêntico precede correlação e há rotação da chave
- GIVEN callback autêntico precede correlação e há rotação da chave
- WHEN receber, reiniciar e associar
- THEN versão autenticada é preservada, recibo aplicado uma vez e HTTP não drena backlog

#### Scenario: R5-SEG-01-S03 — conta agressora envia órfãos e repete evento com timestamp novo
- GIVEN conta agressora envia órfãos e repete evento com timestamp novo
- WHEN processar junto de outra conta
- THEN dedupe de evento estável, quota por escopo e progresso da conta saudável

### Requirement: R5-SEG-02 — Fallback de oferta respeita negação e revogação
O Hub SHALL distinguir indisponibilidade transitória de negação autoritativa ao resolver ofertas. Negação, revogação, conflito ou resposta inválida não podem ser convertidos em autorização por cache. Projeções válidas devem atender o caminho quente dentro de janela de autorização explicitamente publicada, com invalidação e limites de armazenamento mensuráveis.

#### Scenario: R5-SEG-02-S01 — oferta em cache e Atlas retorna 403 ou 409
- GIVEN oferta em cache e Atlas retorna 403 ou 409
- WHEN resolver novamente
- THEN negação prevalece e cache não autoriza nova admissão

#### Scenario: R5-SEG-02-S02 — projeção elegível e Atlas temporariamente indisponível
- GIVEN projeção elegível e Atlas temporariamente indisponível
- WHEN admitir sob janela publicada
- THEN fallback permitido sem ultrapassar validade e sem consulta remota obrigatória em cada hit

#### Scenario: R5-SEG-02-S03 — oferta suspensa e cache em múltiplas réplicas
- GIVEN oferta suspensa e cache em múltiplas réplicas
- WHEN propagar revogação e testar
- THEN janela máxima respeitada, rejeição seletiva e auditoria da versão

### Requirement: R5-SEG-03 — Adoção real de RLS e autorização de dados auxiliares
O Hub SHALL aplicar identidade autenticada por transação/work item com roles runtime sem propriedade nem bypass, incluindo tabelas auxiliares e caminhos administrativos. Escopo global exige autorização nominal específica e auditoria. Provas de isolamento devem demonstrar dados próprios existentes, dados alheios existentes e invisíveis, escrita cruzada recusada e ausência de vazamento ao reutilizar conexões.

#### Scenario: R5-SEG-03-S01 — dados A/B existentes e credencial runtime real
- GIVEN dados A/B existentes e credencial runtime real
- WHEN ler e escrever via API/SQL com A
- THEN A visível, B invisível, escrita em B negada por assertions

#### Scenario: R5-SEG-03-S02 — pool reutilizado e worker executando tenant alternado
- GIVEN pool reutilizado e worker executando tenant alternado
- WHEN commit, rollback e reconexão
- THEN não há contexto residual nem fallback para owner

#### Scenario: R5-SEG-03-S03 — operador com leitura local e operador global autorizado
- GIVEN operador com leitura local e operador global autorizado
- WHEN consultar protocolos/capacidade/dados auxiliares
- THEN escopos distintos respeitados, global auditado e segredos mascarados
