# Delta for 01-custodia-e-integracao-externa

## ADDED Requirements

### Requirement: R3-EXE-01 — Adapters executáveis e autenticação completa
O Hub SHALL executar adapters versionados homologados, transmitir configuração completa de autenticação e distinguir inventário importado de capacidade executável. Publicação deve rejeitar capacidades não suportadas; resultado sintético nunca substitui resposta real.

#### Scenario: R3-EXE-01-S01 — conta REST homologada com API Key
- GIVEN conta REST homologada com API Key
- WHEN executar SYNC e ASYNC
- THEN header contratado chega ao provedor e ambos os resultados conservam seu marcador real

#### Scenario: R3-EXE-01-S02 — adapter ausente ou autenticação incompleta
- GIVEN adapter ausente ou autenticação incompleta
- WHEN publicar oferta
- THEN a validação impede ativação e identifica a lacuna

#### Scenario: R3-EXE-01-S03 — dois tenants com bindings específicos
- GIVEN dois tenants com bindings específicos
- WHEN executar em paralelo
- THEN cada chamada usa exclusivamente a credencial elegível congelada

### Requirement: R3-EXE-02 — Callback com confirmação de custódia
O Hub SHALL autenticar callback pela política da conta, limitar método/tamanho/replay e emitir 2xx somente após custódia durável. Recibos órfãos, duplicados e tardios autenticados devem ter disposição recuperável sem substituir final terminal.

#### Scenario: R3-EXE-02-S01 — callback válido de operação conhecida
- GIVEN callback válido de operação conhecida
- WHEN receber resultado
- THEN recibo e obrigação de aplicação são confirmados antes do 2xx

#### Scenario: R3-EXE-02-S02 — banco indisponível ou assinatura inválida
- GIVEN banco indisponível ou assinatura inválida
- WHEN receber callback
- THEN erro retentável não promete custódia e assinatura inválida é recusada

#### Scenario: R3-EXE-02-S03 — callback anterior à correlação e outro após final
- GIVEN callback anterior à correlação e outro após final
- WHEN reconciliar
- THEN o primeiro é recuperado de inbox órfã e o segundo fica como evidência tardia

### Requirement: R3-EXE-03 — Recuperação de efeito incerto e fencing
O Hub SHALL manter obrigação recuperável para SUBMITTING/UNKNOWN, validar identidade completa/hash em todo replay e verificar lease/epoch/prazo antes de I/O. Reenvio de SUBMIT exige idempotência homologada ou prova positiva de ausência de efeito; UNKNOWN exige consulta ou reconciliação.

#### Scenario: R3-EXE-03-S01 — queda após intenção antes do HTTP
- GIVEN queda após intenção antes do HTTP
- WHEN reiniciar
- THEN a obrigação é retomada com decisão segura e não fica abandonada

#### Scenario: R3-EXE-03-S02 — provedor executou e a resposta se perdeu
- GIVEN provedor executou e a resposta se perdeu
- WHEN recuperar UNKNOWN
- THEN oráculo externo comprova um efeito e reconciliação recupera o resultado

#### Scenario: R3-EXE-03-S03 — owner antigo ou replay com aplicação/hash distintos
- GIVEN owner antigo ou replay com aplicação/hash distintos
- WHEN tentar enviar/consultar
- THEN fencing ou conflito impede efeito e vazamento

### Requirement: R3-EXE-04 — Relógios de retry, polling e prazo final
O Hub SHALL separar SLA do cliente, SLA do provedor, TTL desde primeira falha transitória, timeout por tentativa e horizonte de reconciliação. Polling/callback devem coexistir; espera saudável não consome TTL de indisponibilidade. Não substituir confirmação durável no prazo por checagem SQL anterior ao commit sem mudança normativa autorizada.

#### Scenario: R3-EXE-04-S01 — provedor saudável demora com polling e callback ativos
- GIVEN provedor saudável demora com polling e callback ativos
- WHEN receber final no SLA
- THEN observadores convergem em um final único

#### Scenario: R3-EXE-04-S02 — primeira falha em t1 com TTL 30 segundos
- GIVEN primeira falha em t1 com TTL 30 segundos
- WHEN agendar retry
- THEN limite t1+30s não se renova e deadline do cliente limita atendimento

#### Scenario: R3-EXE-04-S03 — pausa entre decisão e commit atravessa SLA
- GIVEN pausa entre decisão e commit atravessa SLA
- WHEN executar contraexemplo do ADR
- THEN nenhum sucesso tardio é tratado como tempestivo; limitação remanescente permanece aberta explicitamente

### Requirement: R3-EXE-05 — Topologia de mensagens e quarentena recuperável
O Hub SHALL verificar topologia durável obrigatória antes de liberar publicação, conservar cada obrigação até handoff comprovado e guardar bytes originais limitados ou referência imutável protegida em quarentena. Hash sozinho não é custódia recuperável; reconciliação cobre fan-out e retenção.

#### Scenario: R3-EXE-05-S01 — primeira inicialização sem assinatura financeira
- GIVEN primeira inicialização sem assinatura financeira
- WHEN publicar final
- THEN obrigação permanece recuperável até destinatário obrigatório receber

#### Scenario: R3-EXE-05-S02 — envelope inválido ou versão desconhecida
- GIVEN envelope inválido ou versão desconhecida
- WHEN quarentenar e confirmar origem
- THEN payload original permanece recuperável com autorização e retenção

#### Scenario: R3-EXE-05-S03 — broker cai e retorna
- GIVEN broker cai e retorna
- WHEN retomar relays e consumidores
- THEN obrigações têm disposição provada sem repetir efeitos ou lançamentos
