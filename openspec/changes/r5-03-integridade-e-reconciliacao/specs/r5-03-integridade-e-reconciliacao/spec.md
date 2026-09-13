# Delta for r5-03-integridade-e-reconciliacao

## ADDED Requirements

### Requirement: R5-DAD-01 — Precisão do resultado preservada em todas as fronteiras
O Hub SHALL preservar valor e tipo de números/identificadores em toda cadeia provedor→fato→estado→produto→representação→GET/webhook/console, sem coerção imprecisa. IDs inteiros fora da faixa segura do consumidor devem ter contrato textual explícito; resultados inválidos não são persistidos como sucesso.

#### Scenario: R5-DAD-01-S01 — resultado contém inteiro 9007199254740993 e decimal exato
- GIVEN resultado contém inteiro 9007199254740993 e decimal exato
- WHEN atravessar SYNC e ASYNC inclusive produto
- THEN valor é idêntico no resultado persistido, GET e webhook

#### Scenario: R5-DAD-01-S02 — ID financeiro maior que inteiro seguro JS
- GIVEN ID financeiro maior que inteiro seguro JS
- WHEN selecionar e enviar ação
- THEN identidade exata é enviada ou entrada é recusada sem arredondamento

#### Scenario: R5-DAD-01-S03 — campos não numéricos e schema rejeitado
- GIVEN campos não numéricos e schema rejeitado
- WHEN validar e consolidar
- THEN tipos preservados e falha do contrato não gera final de sucesso

### Requirement: R5-DAD-02 — Restore populado e reconciliação antes de retomada
O Hub SHALL demonstrar restore não vazio e retomada cercada de todas as autoridades e obrigações, incluindo planos, etapas, compensações, recibos, versões de objetos e financeiro. Antes de liberar tráfego deve reconciliar identidades/bytes/efeitos/valores; ausência de fixture ou divergência impede aprovação.

#### Scenario: R5-DAD-02-S01 — backup contém produto em trânsito, UNKNOWN, reserva e versões de arquivo
- GIVEN backup contém produto em trânsito, UNKNOWN, reserva e versões de arquivo
- WHEN restaurar e retomar em ambiente isolado
- THEN bytes/hash/referências e estados concordam; efeito e journal não duplicam

#### Scenario: R5-DAD-02-S02 — faltam versão de objeto ou etapa do plano
- GIVEN faltam versão de objeto ou etapa do plano
- WHEN validar restore
- THEN gate recusa retomada com obrigação identificada

#### Scenario: R5-DAD-02-S03 — falha regional ou restauração de ponto anterior ao efeito externo
- GIVEN falha regional ou restauração de ponto anterior ao efeito externo
- WHEN cercar e reconciliar
- THEN RPO/RTO medidos dentro do modelo declarado e sem replay cego

### Requirement: R5-DAD-03 — Liquidação por valor efetivo e completude demonstrável
O Hub SHALL liquidar saldo pelo valor contratado efetivo, preservando reserva/hold de incerteza e liberando excedente de forma auditável. Fechamento exige evidência durável de completude dos produtores, sem watermarks fabricados. Fatos tardios devem gerar disposição financeira consultável e ajustável sem alterar exportação fechada.

#### Scenario: R5-DAD-03-S01 — reserva 100 e consumo final 30
- GIVEN reserva 100 e consumo final 30
- WHEN capturar e consultar saldo
- THEN 30 liquidados e 70 liberados conforme contrato, sem duplicação

#### Scenario: R5-DAD-03-S02 — produtor/outbox atrasado e operador pede fechamento
- GIVEN produtor/outbox atrasado e operador pede fechamento
- WHEN avaliar corte
- THEN período fica bloqueado até completude demonstrada

#### Scenario: R5-DAD-03-S03 — fato tardio após exportação fechada
- GIVEN fato tardio após exportação fechada
- WHEN receber e conciliar
- THEN disputa/ajuste visível, exportação original imutável e conservação monetária

### Requirement: R5-DAD-04 — Quarentena financeira conserva mensagem recuperável
O Hub SHALL conservar conteúdo recuperável e proveniência de mensagens financeiras não aplicáveis antes de confirmar sua remoção do transporte. Reprocessamento exige autorização, idempotência e trilha de disposição; hash isolado não constitui custódia do conteúdo.

#### Scenario: R5-DAD-04-S01 — evento financeiro inválido com bytes originais
- GIVEN evento financeiro inválido com bytes originais
- WHEN consumir e confirmar mensagem
- THEN bytes ou referência durável verificável existem antes do ACK

#### Scenario: R5-DAD-04-S02 — falha da persistência de quarentena
- GIVEN falha da persistência de quarentena
- WHEN consumir
- THEN mensagem não é confirmada e redelivery permanece possível

#### Scenario: R5-DAD-04-S03 — contrato corrigido e operador autorizado pede replay
- GIVEN contrato corrigido e operador autorizado pede replay
- WHEN reprocessar quarentena
- THEN fato recuperado sem duplicar journal e histórico da disposição preservado
