# Delta for r6-01-isolamento-e-custodia

## ADDED Requirements

### Requirement: R6-SEG-01 — RLS no caminho real e prova com dados existentes
O Hub SHALL aplicar escopo autenticado por transação e por item de trabalho com roles runtime sem bypass; tabelas sem tenant direto devem ter autorização por relação ou autoridade operacional específica. Testes de isolamento SHALL afirmar existência e acesso próprio, invisibilidade alheia e recusa de escrita cruzada antes de remover fixtures.

#### Scenario: R6-SEG-01-S01 — fixtures A e B confirmadas e role runtime
- GIVEN fixtures A e B confirmadas e role runtime
- WHEN consultar pela API de A e SQL escopado
- THEN A existe e é visível, B existe e é invisível, escrita em B falha

#### Scenario: R6-SEG-01-S02 — pool de uma conexão e A/B alternados
- GIVEN pool de uma conexão e A/B alternados
- WHEN executar sucesso, rollback e reconexão
- THEN não sobra identidade na conexão nem acesso owner

#### Scenario: R6-SEG-01-S03 — worker de múltiplos tenants e administrador nominal
- GIVEN worker de múltiplos tenants e administrador nominal
- WHEN processar claims e consultar escopo global
- THEN obrigações progridem com autorização específica e auditoria sem acesso global implícito

### Requirement: R6-SEG-02 — Callback aceito sobrevive à rotação e à indisponibilidade
O Hub SHALL preservar a atestação de autenticação obtida no ingresso e a disposição recuperável de cada callback aceito. Rotação e falha transitória não podem invalidar custódia legítima; esgotamento de retry deve manter obrigação em quarentena reprocessável, distinta de rejeição de contrato.

#### Scenario: R6-SEG-02-S01 — callback legítimo aceito antes da correlação
- GIVEN callback legítimo aceito antes da correlação
- WHEN rotacionar chave, reiniciar e criar operação
- THEN recibo original é aplicado uma vez com autenticação preservada

#### Scenario: R6-SEG-02-S02 — três falhas transitórias no apply de callback aceito
- GIVEN três falhas transitórias no apply de callback aceito
- WHEN restaurar dependência e reprocessar
- THEN obrigação continua recuperável e resultado chega sem nova emissão do provedor

#### Scenario: R6-SEG-02-S03 — muitos órfãos de uma conta e conta saudável
- GIVEN muitos órfãos de uma conta e conta saudável
- WHEN executar ingestão e expirar janela de correlação
- THEN conta saudável progride e órfãos recebem disposição auditável sem lock global
