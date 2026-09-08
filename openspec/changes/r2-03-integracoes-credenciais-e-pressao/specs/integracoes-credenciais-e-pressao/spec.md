# Delta for integracoes-credenciais-e-pressao

## ADDED Requirements

### Requirement: R2-INT-01 — Adapter executável homologado por capacidade

Uma integração publicada SHALL indicar adapter/version, operações/métodos/paths, schemas, autenticação, mapeamento de erros e capacidades de submit/status/fetch/cancel/callback. Endpoint importado NÃO SHALL ser considerado executável por conter URL. Um adapter genérico SHALL executar apenas configuração declarativa limitada, sem código arbitrário.

Baseline relacionada: COM-02, CAT-06, CAT-11.

#### Scenario: R2-INT-01-S01 — REST real distinto do simulador

- GIVEN serviço homologado requer GET /consulta/{id} ou POST /analise com corpo próprio
- WHEN cliente usa a oferta
- THEN conector envia método/path/corpo homologados e valida resposta, sem chamar /v1/operations do simulador

#### Scenario: R2-INT-01-S02 — Capacidade ausente

- GIVEN catálogo importado não possui adapter ou contrato validado
- WHEN operador tenta publicar oferta
- THEN validação lista bloqueios e não anuncia o endpoint como integração pronta

#### Scenario: R2-INT-01-S03 — Protocolo especializado

- GIVEN perfil exige SOAP/XML, SFTP ou gRPC
- WHEN não existe adapter homologado para ele
- THEN capacidade fica não disponível; suporte futuro exige adapter e ensaio específico, sem simulação silenciosa via REST

### Requirement: R2-INT-02 — Credencial efetiva vinculada ao cliente e à conta

Cada tentativa SHALL resolver uma seleção explícita de ambiente, tenant/aplicação, oferta, conta, modo, binding e versão de segredo. TENANT_DEDICATED ausente/inválido NÃO SHALL cair em SHARED_HUB. Segredo efetivamente usado SHALL ser o do binding; tokens L1 SHALL separar binding, versão, audience, escopos e cliente pertinente. settlement_party SHALL permanecer decisão econômica independente.

Baseline relacionada: CFG-05, SEG-05, FIN-11, CAT-11.

#### Scenario: R2-INT-02-S01 — Dois clientes na mesma conta

- GIVEN A e B têm vínculos dedicados e credenciais diferentes
- WHEN executam simultaneamente na mesma provider_account
- THEN provedor identifica os principais corretos e nenhum token/certificado cruza clientes

#### Scenario: R2-INT-02-S02 — Vínculo dedicado indisponível

- GIVEN cliente exige modo dedicado e vínculo está revogado
- WHEN operação tenta iniciar
- THEN é recusada antes do submit; não usa credencial de conta compartilhada

#### Scenario: R2-INT-02-S03 — Rotação e cache

- GIVEN token existe para versão 1 e binding passa à versão 2
- WHEN nova tentativa é autorizada pela política de rotação
- THEN cache antigo não é reutilizado indevidamente; operação já aceita conserva conta, snapshot e pagador

### Requirement: R2-INT-03 — Segredos reais e cache dispensável

Segredos SHALL ser obtidos por identidade autorizada em cofre e utilizados em Basic/API key/OAuth/mTLS reais conforme perfil. Referência de cofre NÃO SHALL ser enviada como senha; header NÃO SHALL simular mTLS. Tokens e segredos SHALL permanecer somente no L1 privado dos workloads autorizados conforme a v4, sem Redis compartilhado. Falha de Redis NÃO SHALL bloquear autenticação ou bootstrap.

Baseline relacionada: SEG-05, DAD-10, OPE-13, ARQ-03.

#### Scenario: R2-INT-03-S01 — mTLS comprovado

- GIVEN provedor exige certificado cliente válido
- WHEN workload usa perfil homologado
- THEN handshake prova certificado correto; certificado ausente/inválido é rejeitado, mesmo com header de referência

#### Scenario: R2-INT-03-S02 — Redis ausente

- GIVEN cofre ou L1 ainda oferece material válido e provedor de identidade está elegível
- WHEN Redis cai ou Cometa reinicia sem Redis
- THEN operação segue sem consultar Redis para tokens e sem panic por cache

#### Scenario: R2-INT-03-S03 — Cofre indisponível

- GIVEN material L1 está expirado ou revogado
- WHEN nova tentativa precisa de segredo
- THEN operação afetada aguarda/recusa dentro da política sem credencial inválida; outras contas elegíveis continuam

### Requirement: R2-INT-04 — Polling e callback combináveis e coordenados

Oferta assíncrona SHALL permitir habilitar polling, callback ou ambos por configuração homologada. Polling SHALL usar autenticação da operação, agenda durável, lease/fencing, intervalo/jitter/limites e deadline. Callback SHALL autenticar, limitar corpo e conservar recibo antes de ACK, incluindo duplicatas, órfãos e tardios. Observações concorrentes SHALL produzir um resultado de negócio, preservando divergências.

Baseline relacionada: EXE-05, EXE-06, SEG-02.

#### Scenario: R2-INT-04-S01 — Ativação de polling

- GIVEN provedor ASYNC homologado possui status endpoint
- WHEN operador habilita polling por nova versão de configuração
- THEN novas operações recebem agenda retomável e consultas autenticadas, sem deploy de código para o caso suportado

#### Scenario: R2-INT-04-S02 — Resposta pendente saudável

- GIVEN status 200 indica PENDING e callback não chegou
- WHEN cliente atinge deadline
- THEN protocolo expira conforme contrato; polling saudável não renova o SLA

#### Scenario: R2-INT-04-S03 — Polling e callback concorrentes

- GIVEN ambos observam final da mesma operação
- WHEN três réplicas processam recibos duplicados e conflitantes
- THEN uma transição vence conforme identidade/versão/prazo; demais recibos ficam classificados e divergência gera diagnóstico

#### Scenario: R2-INT-04-S04 — Falha de recibo

- GIVEN callback autenticado chegou
- WHEN writer não confirma a custódia
- THEN ACK 2xx não é emitido; o parceiro pode retransmitir conforme contrato

### Requirement: R2-INT-05 — Janela absoluta de retry e classificação de falha

retry_ttl_seconds SHALL ser inteiro não negativo, congelado na política do aceite. O primeiro erro transitório elegível SHALL fixar first_transient_at e retry_until uma única vez. Novas tentativas SHALL caber no menor prazo de retry, cliente e provedor aplicável, respeitar Retry-After/jitter e comprovação de segurança do efeito. Timeout ambíguo SHALL ir para UNKNOWN, não retry cego.

Baseline relacionada: EXE-09, EXE-10, CFG-04.

#### Scenario: R2-INT-05-S01 — TTL zero

- GIVEN política tem retry_ttl_seconds=0
- WHEN ocorre primeira indisponibilidade transitória comprovadamente sem efeito
- THEN não se agenda novo submit e o encerramento segue contrato

#### Scenario: R2-INT-05-S02 — Reinício no meio da janela

- GIVEN primeiro erro fixou retry_until
- WHEN pod reinicia ou rota equivalente muda
- THEN prazo original é preservado e nenhuma tentativa começa sem orçamento restante

#### Scenario: R2-INT-05-S03 — 429 com espera incompatível

- GIVEN Retry-After é maior que a janela restante
- WHEN controlador considera nova tentativa
- THEN não viola o pedido de espera nem ultrapassa prazo para tentar; encerra ou reconcilia conforme efeito

#### Scenario: R2-INT-05-S04 — Timeout após envio

- GIVEN não é possível provar se provedor executou
- WHEN há TTL ainda disponível
- THEN somente consulta/reconciliação segura é permitida antes de autorizar nova execução

### Requirement: R2-INT-06 — Controle adaptativo global por domínio de capacidade

O Hub SHALL controlar taxa de submit/status/fetch/cancel, concorrência HTTP e pendências externas por capacity_domain homologado, compartilhado entre contas quando o limite do parceiro assim exigir. Feedback de latência, 429, indisponibilidade e timeout SHALL reduzir pressão e recuperar capacidade gradualmente dentro do teto autorizado. Réplicas NÃO SHALL multiplicar quota; mudança de segredo NÃO SHALL reiniciar o domínio.

Baseline relacionada: OPE-07, OPE-08, EXE-13, CAT-06.

#### Scenario: R2-INT-06-S01 — Provedor degrada

- GIVEN taxa está estável e há demanda
- WHEN aumentam timeouts/503 e p95 de resposta
- THEN controlador diminui admissão externa, preserva piso de reconciliação e expõe motivo/limite/medidas

#### Scenario: R2-INT-06-S02 — Provedor escala

- GIVEN erros cessam, latência recupera e há demanda
- WHEN passam janelas de estabilidade configuradas
- THEN a taxa sobe gradualmente por sondagem até capacidade segura/teto, sem ficar presa ao limite antigo

#### Scenario: R2-INT-06-S03 — Três réplicas e partição

- GIVEN mesmo domínio é usado por três executores
- WHEN coordenador de quota fica indisponível
- THEN soma das permissões ainda válidas não excede orçamento; sem lease válido não há novos envios e outras contas isoladas seguem

#### Scenario: R2-INT-06-S04 — Ruído de cliente

- GIVEN A excede dez vezes sua quota e B mantém carga contratada
- WHEN A usa provedor degradado por 15 minutos
- THEN B mantém SLO do perfil e reservas; A recebe contenção/recusa explícita em vez de consumir capacidade de B

### Requirement: R2-INT-07 — Entrega ao cliente com identidade e política próprias

Pulsar SHALL criar obrigação por evento final e destination_id/version do tenant, com key_id, corpo/hash e política de entrega fixados. Envio SHALL ter claim exclusivo, tentativas/recibos duráveis, assinatura dos bytes com timestamp e IDs verificáveis, retry dentro do prazo e EXHAUSTED consultável. HTTP 2xx SHALL significar recibo de transporte, não processamento do cliente. Reentrega NÃO SHALL executar provedor.

Baseline relacionada: EXE-08, COM-05, SEG-02, CFG-03.

#### Scenario: R2-INT-07-S01 — Mesma URL em dois tenants

- GIVEN A e B usam mesma URL com chaves distintas
- WHEN seus resultados são entregues
- THEN Pulsar seleciona chave por tenant/destino/versão, nunca por URL isolada

#### Scenario: R2-INT-07-S02 — Rotação e replay

- GIVEN entrega antiga falhou e destino foi alterado
- WHEN operador autorizado pede reentrega
- THEN ação explicita a versão/chave elegível e conserva resultado original; não redireciona silenciosamente evento histórico

#### Scenario: R2-INT-07-S03 — Destino lento ou indisponível

- GIVEN webhook de A retorna 429/timeout
- WHEN worker agenda nova tentativa e B também tem entrega
- THEN lease e retry respeitam política de A; B não fica bloqueado; esgotamento fica visível sem perder final local

#### Scenario: R2-INT-07-S04 — Duplicata no receptor

- GIVEN primeiro recibo HTTP se perdeu após o cliente receber
- WHEN Pulsar retransmite
- THEN event_id/result_version identificam o mesmo final e assinatura/timestamp são verificáveis

### Requirement: R2-INT-08 — SLA do provedor separado de SLA do cliente

Cada operação SHALL registrar marco contratual do SLA do provedor e timestamps de submit, aceite externo, observações, resposta final durável e resultado do cliente. MONITOR_ONLY e REJECT_LATE SHALL ter comportamentos distintos explícitos e não ampliar deadline do cliente. Latência de submit, duração final e entrega de webhook SHALL ser aferidas separadamente.

Baseline relacionada: OPE-10, EXE-11, FIN-10, CAT-10.

#### Scenario: R2-INT-08-S01 — Violação apenas do provedor

- GIVEN prazo do provedor venceu e do cliente ainda não
- WHEN final chega sob MONITOR_ONLY
- THEN registra quebra do provedor e pode concluir cliente no prazo, sem ocultar a violação

#### Scenario: R2-INT-08-S02 — Rejeição contratada

- GIVEN mesmo atraso ocorre sob REJECT_LATE
- WHEN final chega
- THEN passo não vira sucesso do produto; recibo e custo elegível permanecem

#### Scenario: R2-INT-08-S03 — Coorte ainda aberta

- GIVEN há protocolos rápidos, lentos e ainda no prazo
- WHEN operador consulta cumprimento do SLA
- THEN painel mostra elegíveis, abertos, cumpridos, vencidos e exclusões, sem contar abertos como sucesso
