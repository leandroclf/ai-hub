# Delta for execucao-duravel-e-resultados

## ADDED Requirements

### Requirement: R2-EXE-01 — Aceite e intenção recuperáveis em todas as modalidades

Antes de confirmar aceite, o Hub SHALL conservar atomicamente UUIDv7, identidade, chave/hash de idempotência, snapshot necessário, prazo e intenção de execução. ASYNC/AUTO SHALL publicar comando a partir dessa intenção durável. SYNC SHALL conservar intenção DIRECT e recuperação cercada, sem depender de broker para responder. Nenhuma custódia SHALL ser confirmada apenas em memória ou cache.

Baseline relacionada: EXE-01, EXE-15, DAD-03, COM-03, DAD-09, DAD-02.

#### Scenario: R2-EXE-01-S01 — Crash após aceite

- GIVEN intenção e protocolo foram confirmados mas broker ainda não recebeu
- WHEN o processo reinicia
- THEN a obrigação é recuperada e executada conforme prazo, sem novo pedido do cliente

#### Scenario: R2-EXE-01-S02 — Broker indisponível

- GIVEN writer e capacidade de backlog estão elegíveis
- WHEN é admitido ASYNC durante falha temporária do broker
- THEN 202 só ocorre com intenção recuperável; ao alcançar limite de capacidade, novas admissões são recusadas explicitamente

#### Scenario: R2-EXE-01-S03 — Commit incerto

- GIVEN conexão cai durante commit do aceite
- WHEN o cliente repete a mesma chave
- THEN o Hub resolve pela autoridade de idempotência, sem afirmar não execução nem criar outro efeito

### Requirement: R2-EXE-02 — Idempotência concorrente com UUID recuperável

Admissões concorrentes com a mesma identidade de tenant/aplicação/operação e Idempotency-Key SHALL convergir para o mesmo protocolo quando o hash semântico coincidir; divergência SHALL retornar 409 sem novo efeito. Todos os retornos posteriores a aceite conhecido SHALL incluir o UUIDv7 e rota de consulta, inclusive erros. Recusas anteriores ao aceite NÃO SHALL inventar protocolo.

Baseline relacionada: EXE-01, EXE-16, DAD-03.

#### Scenario: R2-EXE-02-S01 — Concorrência real

- GIVEN duas conexões enviam mesma chave e conteúdo simultaneamente
- WHEN ambas disputam unicidade
- THEN as respostas identificam um único protocolo; conflito SQL não vira segunda execução nem 503 indevido

#### Scenario: R2-EXE-02-S02 — Chave divergente

- GIVEN uma chave já possui aceite
- WHEN chega conteúdo semanticamente diferente
- THEN 409 identifica conflito sem substituir conteúdo anterior nem revelar dados de outro tenant

#### Scenario: R2-EXE-02-S03 — Erro após aceite

- GIVEN protocolo foi conservado
- WHEN reserva, transporte SYNC ou finalização falha
- THEN o erro expõe UUID e consulta quando o aceite é conhecido; retry usa a mesma chave

### Requirement: R2-EXE-03 — Um dono de despacho e tentativa anterior ao efeito

Operação externa SHALL ter identidade estável, reivindicação exclusiva, epoch/fencing e tentativa persistida antes do envio. Duplicata, timeout de transporte, perda de lease ou takeover NÃO SHALL autorizar novo efeito sem resolver o anterior. Falha em gravar a tentativa SHALL impedir o envio.

Baseline relacionada: EXE-04, EXE-09, EXE-15, DAD-03, COM-06.

#### Scenario: R2-EXE-03-S01 — Dois executores

- GIVEN fila e DIRECT/recuperação apresentam mesmo command_id
- WHEN dois pods tentam executar
- THEN apenas o dono válido envia; outro consulta a operação durável

#### Scenario: R2-EXE-03-S02 — Crash após envio

- GIVEN provedor pode ter executado, mas resposta não foi conservada
- WHEN worker morre e lease expira
- THEN operação fica UNKNOWN para reconciliação; não há novo submit nem failover cego

#### Scenario: R2-EXE-03-S03 — Fencing antigo

- GIVEN executor antigo volta após takeover
- WHEN tenta gravar estado ou enviar usando epoch vencido
- THEN sua ação é rejeitada; lease expirada não prova ausência de efeito já enviado

### Requirement: R2-EXE-04 — Resultado externo válido e durável

Cometa SHALL conservar resposta externa validada, correlação, recibo, estado e intenção de fato no mesmo commit local; SHALL retornar essa resposta em execução direta e replay. Input do cliente, HTTP 200 sem semântica válida ou falha de decode NÃO SHALL ser usados como resultado de sucesso.

Baseline relacionada: EXE-04, EXE-14, DAD-04, COM-06.

#### Scenario: R2-EXE-04-S01 — Conteúdo real

- GIVEN entrada contém input_marker e provedor devolve result_marker diferente
- WHEN operação termina em SYNC ou ASYNC
- THEN resultado contém result_marker e não substitui a resposta pelo input

#### Scenario: R2-EXE-04-S02 — Resposta inválida

- GIVEN provedor devolve 200 com corpo vazio, estado desconhecido ou schema inválido
- WHEN Cometa valida a observação
- THEN não publica SUCCEEDED; registra erro de contrato e evidência permitida para reconciliação

#### Scenario: R2-EXE-04-S03 — Commit falha

- GIVEN resposta externa válida foi recebida
- WHEN persistência de estado, corpo ou outbox falha
- THEN Cometa não declara fato durável; mantém obrigação recuperável e sem novo submit automático

### Requirement: R2-EXE-05 — Inbox, outbox e confirmação de mensagens

Todo consumidor SHALL confirmar a mensagem somente após commit de inbox e efeito ou obrigação durável equivalente. Publicação SHALL preservar event_id, timestamps originais, versão, identidade e causalidade nas reentregas. Mensagem inválida ou schema desconhecido SHALL ir para quarentena durável com diagnóstico; não SHALL ser descartada silenciosamente.

Baseline relacionada: COM-03, COM-04, OPE-11, FIN-04, EXE-08.

#### Scenario: R2-EXE-05-S01 — Efeito falha

- GIVEN Órbita, Cometa, Pulsar ou Libra não consegue confirmar seu efeito
- WHEN recebe uma mensagem válida
- THEN não remove a mensagem; retry posterior conclui o efeito uma vez

#### Scenario: R2-EXE-05-S02 — Reentrega após commit

- GIVEN commit concluiu, ACK se perdeu
- WHEN broker entrega novamente event_id
- THEN inbox reconhece efeito confirmado e ACK não duplica operação, delivery ou lançamento

#### Scenario: R2-EXE-05-S03 — Schema desconhecido

- GIVEN evento não é compatível com consumidor
- WHEN o evento chega
- THEN original sanitizado e metadados ficam em quarentena; alerta e replay autorizado preservam a identidade de origem

### Requirement: R2-EXE-06 — Deadline por confirmação durável e final tardio

Elegibilidade de sucesso SHALL exigir representação final validada e confirmação durável estritamente anterior ao client_deadline_at, conforme EXE-11 v4. A arbitragem SHALL considerar commit lento e não o horário de execução do timer nem o início da transação. Igualdade SHALL ser atraso. Se não houver prova de elegibilidade, NÃO SHALL sair sucesso. Resultado tardio SHALL permanecer como evidência REJECTED_LATE_SLA sem reabrir protocolo, gerar receita de sucesso ou eliminar eventual custo.

Baseline relacionada: EXE-11, EXE-12, FIN-10, OPE-10.

#### Scenario: R2-EXE-06-S01 — Timer atrasado

- GIVEN timer está pausado e o deadline já passou
- WHEN chega final tecnicamente bem-sucedido
- THEN o Hub não publica sucesso e materializa/recupera encerramento por SLA com recibo tardio

#### Scenario: R2-EXE-06-S02 — Commit cruza o limite

- GIVEN transação começou antes do deadline e sua confirmação durável termina depois
- WHEN o Hub arbitra final versus expiração
- THEN o início antecipado não torna sucesso elegível; evento de sucesso não escapa pelo relay

#### Scenario: R2-EXE-06-S03 — Final e expiração concorrentes

- GIVEN callback, polling e timer disputam o protocolo
- WHEN a confirmação ocorre antes, na igualdade ou depois do prazo em execuções separadas
- THEN há exatamente um terminal contratual por execução e todas as observações são preservadas

#### Scenario: R2-EXE-06-S04 — Resultado tardio custoso

- GIVEN protocolo EXPIRED possui compra por execução externa
- WHEN provedor confirma final depois do prazo
- THEN GET e webhook conservam o erro final; Libra considera apenas o custo elegível e a contestação

### Requirement: R2-EXE-07 — SYNC direto e AUTO com espera limitada

SYNC SHALL ser admitido apenas em oferta elegível com orçamento HTTP compatível e SHALL devolver final contratual na mesma requisição quando concluído no prazo; NÃO SHALL converter silenciosamente para 202. AUTO SHALL aguardar até auto_wait_seconds, limitado pelo prazo, e retornar final ou 202 com a mesma identidade. ASYNC SHALL retornar aceite durável sem esperar o provedor. Fila NÃO SHALL ser percurso obrigatório do comando/final SYNC.

Baseline relacionada: EXE-02, EXE-14, COM-01, COM-06, CAT-11.

#### Scenario: R2-EXE-07-S01 — SYNC com broker parado

- GIVEN oferta SYNC homologada e dependências autoritativas disponíveis
- WHEN cliente solicita SYNC durante falha de broker
- THEN recebe final durável na mesma conexão; fatos posteriores ficam na outbox

#### Scenario: R2-EXE-07-S02 — Natureza incompatível

- GIVEN provedor só conclui assincronamente e oferta não qualifica SYNC
- WHEN cliente solicita SYNC
- THEN admissão é recusada com motivo antes de efeito, sem escolher ASYNC por conta própria

#### Scenario: R2-EXE-07-S03 — AUTO rápido e lento

- GIVEN oferta AUTO tem espera configurada e intenção QUEUED
- WHEN uma execução conclui dentro da espera e outra depois
- THEN a primeira retorna final; a segunda 202 com UUID; ambas têm um único despacho e deadline inalterado

#### Scenario: R2-EXE-07-S04 — Cliente desconecta

- GIVEN efeito externo foi enviado em SYNC
- WHEN conexão do cliente é encerrada
- THEN custódia/reconciliação continua com contexto interno limitado; retry e GET recuperam a mesma identidade

### Requirement: R2-EXE-08 — Representação final única por contrato de cliente

O corpo completo da resposta final SHALL ser materializado uma vez por protocolo/versão/perfil, com media_type, bytes ou referência imutável e hash. POST final, GET final e corpo de webhook SHALL usar essa representação negociada; headers de entrega podem variar. Consultas SHALL usar apenas custódia do Hub e não consultar provedor. Versões e status internos/externos SHALL ser coerentes. Compatibilidade com wrapper v1 SHALL ser preservada por perfil, não por projeção regenerada em cada leitura.

Baseline relacionada: COM-05, COM-04, DAD-04, CAT-09, EXE-07.

#### Scenario: R2-EXE-08-S01 — Igualdade completa

- GIVEN protocolo possui representação final v1 ou perfil legado
- WHEN são obtidos POST final, GET e duas tentativas de webhook
- THEN corpos completos são idênticos ao hash persistido, inclusive wrapper e result_version, e HMAC verifica os mesmos bytes

#### Scenario: R2-EXE-08-S02 — Upgrade e mudança de contrato

- GIVEN um final foi materializado na versão antiga
- WHEN software e contrato são atualizados
- THEN consulta e reentrega desse protocolo preservam bytes antigos; novos protocolos usam o perfil novo

#### Scenario: R2-EXE-08-S03 — Resultado pendente ou expirado

- GIVEN protocolo ainda está em andamento ou já EXPIRED
- WHEN cliente autorizado consulta
- THEN recebe 200 com estado contratado local; consulta não faz polling no provedor nem mascara atraso como sucesso

### Requirement: R2-EXE-09 — Reconciliação de obrigações sem reexecutar efeitos

O Hub SHALL detectar protocolos sem despacho, operações UNKNOWN, finais sem consolidação e obrigações sem efeito esperado por idade/estado. Recuperação SHALL respeitar identidade, epoch, prazo e evidência do parceiro. Cancelar prazo do cliente NÃO SHALL eliminar reconciliação externa ou financeira. Falta de recurso em réplica atrasada NÃO SHALL ser interpretada como inexistência comprovada.

Baseline relacionada: EXE-03, EXE-09, EXE-15, OPE-11, DAD-09.

#### Scenario: R2-EXE-09-S01 — Protocolo órfão de despacho

- GIVEN aceite existe e nenhum executor o recebeu
- WHEN scanner verifica idade e intenção
- THEN recupera mesma obrigação se elegível ou encerra contratualmente, com diagnóstico da causa

#### Scenario: R2-EXE-09-S02 — Parceiro sem consulta idempotente

- GIVEN operação ficou UNKNOWN após envio
- WHEN recuperação não consegue demonstrar ausência de efeito
- THEN mantém pendência explícita e acionável; não reenvia nem libera hold por suposição

#### Scenario: R2-EXE-09-S03 — Callback antes da associação

- GIVEN recibo válido chega antes de provider_request_id ser associado
- WHEN associação torna-se durável depois
- THEN recibo é reconciliado sem perda ou segunda operação
