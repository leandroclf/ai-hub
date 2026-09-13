# Delta for r5-02-produtos-e-prazos-duraveis

## ADDED Requirements

### Requirement: R5-EXE-01 — Horizonte de retry e SLA verificados antes do despacho
O Hub SHALL impedir novos despachos com possibilidade de efeito após o horizonte aplicável e conservar separadamente prazo do cliente, provedor, tentativa e retry de indisponibilidade desde primeira falha. Pendência assíncrona legítima não inicia TTL de falha. Takeover não renova prazos; observação tardia conserva evidência sem alterar final fechado.

#### Scenario: R5-EXE-01-S01 — retry_until vencido e upstream voltou
- GIVEN retry_until vencido e upstream voltou
- WHEN reivindicar e despachar
- THEN nenhum novo efeito é iniciado; intenção recebe disposição durável

#### Scenario: R5-EXE-01-S02 — polling saudável demora além do SLA e callback chega depois
- GIVEN polling saudável demora além do SLA e callback chega depois
- WHEN aplicar política de SLA congelada
- THEN final coerente, recibo tardio preservado e GET sem nova chamada

#### Scenario: R5-EXE-01-S03 — pausa após claim, lease expirada e comando legado sem novo campo
- GIVEN pausa após claim, lease expirada e comando legado sem novo campo
- WHEN retomar
- THEN barreira pré-I/O e migração impedem execução fora do horizonte

### Requirement: R5-EXE-02 — Limite de execução de produto atômico e retomável
O Hub SHALL reservar vaga de execução e posse da etapa atomicamente por produto, respeitando max_parallel entre réplicas. Uma etapa com posse expirada deve ser recuperável sem disputar uma segunda vaga nem repetir efeito confirmado. Cancelamento impede novo efeito e estado terminal não é ressuscitado.

#### Scenario: R5-EXE-02-S01 — max_parallel=1 e dois publishers simultâneos
- GIVEN max_parallel=1 e dois publishers simultâneos
- WHEN reivindicar duas etapas independentes
- THEN somente uma reserva válida existe

#### Scenario: R5-EXE-02-S02 — crash após marcar RUNNING e antes de publicar
- GIVEN crash após marcar RUNNING e antes de publicar
- WHEN expirar lease e retomar
- THEN a própria vaga é recuperada e etapa progride sem deadlock

#### Scenario: R5-EXE-02-S03 — cancelamento ou final antes de novo I/O
- GIVEN cancelamento ou final antes de novo I/O
- WHEN retomar owner antigo e novo
- THEN nenhum efeito novo após barreira e estado terminal preservado

### Requirement: R5-EXE-03 — Contrato, provedor e incidência próprios por etapa
O Hub SHALL congelar por etapa serviço/versão, rota homologada, conta, binding, contrato de compra, perfil técnico, prazo e identidade econômica coerentes. A venda do produto e os custos das etapas devem manter escopos próprios sem duplicação. Consolidação deve seguir política publicada e preservar proveniência do plano.

#### Scenario: R5-EXE-03-S01 — produto com serviços de dois provedores e bindings diferentes
- GIVEN produto com serviços de dois provedores e bindings diferentes
- WHEN admitir e executar
- THEN cada etapa usa conta/contrato/credencial corretos e venda não duplica

#### Scenario: R5-EXE-03-S02 — catálogo/rota mudam depois do aceite
- GIVEN catálogo/rota mudam depois do aceite
- WHEN retomar etapa e consultar resultado
- THEN snapshot original é respeitado e hash é verificável

#### Scenario: R5-EXE-03-S03 — serviço ou contrato da etapa não é elegível
- GIVEN serviço ou contrato da etapa não é elegível
- WHEN publicar ou admitir
- THEN recusa antes de efeito e erro aponta etapa incompatível

### Requirement: R5-EXE-04 — Compensação tem ordem causal e prazo próprio
O Hub SHALL conservar e executar compensações em ordem causal reversa com prazo, retry e autorização próprios, independentemente do encerramento da resposta ao cliente. Falha ou incerteza de compensação deve permanecer reconciliável sem ser convertida em sucesso nem reabrir protocolo.

#### Scenario: R5-EXE-04-S01 — A precede B e etapa posterior falha
- GIVEN A precede B e etapa posterior falha
- WHEN compensar
- THEN B é compensada antes de A, com comandos idempotentes

#### Scenario: R5-EXE-04-S02 — cliente expira com compensação pendente
- GIVEN cliente expira com compensação pendente
- WHEN reiniciar workers
- THEN obrigação continua sob seu próprio orçamento

#### Scenario: R5-EXE-04-S03 — compensação falha após possível efeito
- GIVEN compensação falha após possível efeito
- WHEN reconciliar
- THEN UNKNOWN conservado e ausência de replay cego demonstrada

### Requirement: R5-EXE-05 — Topologia de mensagens pronta antes de publicar obrigações
O Hub SHALL confirmar a topologia e políticas de todas as assinaturas obrigatórias antes de liberar publicação de fatos duráveis. Indisponibilidade da topologia deve manter obrigações no outbox sem bloquear rotas independentes. Alteração ou recriação de recurso exige revalidação antes de descartar custódia local.

#### Scenario: R5-EXE-05-S01 — ambiente limpo e consumidor financeiro inicia por último
- GIVEN ambiente limpo e consumidor financeiro inicia por último
- WHEN aceitar pedido e produzir fato antes dele
- THEN fato permanece no outbox até assinatura obrigatória confirmada

#### Scenario: R5-EXE-05-S02 — assinatura removida ou policy incorreta durante execução
- GIVEN assinatura removida ou policy incorreta durante execução
- WHEN publicar novos fatos
- THEN falha detectada e custódia não é descartada como entrega completa

#### Scenario: R5-EXE-05-S03 — broker recupera e processos reiniciam fora de ordem
- GIVEN broker recupera e processos reiniciam fora de ordem
- WHEN retomar
- THEN fatos chegam a todos consumidores obrigatórios sem duplicação de efeito
