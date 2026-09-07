# Configuração, Segurança e Manutenção — Delta de Especificação

## ADDED Requirements

### Requirement: CFG-01 — Portal e API administrativa
O sistema SHALL permitir cadastrar, buscar, validar, versionar, publicar, suspender e descontinuar clientes/aplicações, provedores/contas, serviços, produtos, planos, contratos, rotas, destinos, políticas de polling, limites e retenção, de modo que cada tela/capacidade informe campos obrigatórios, unidade, faixa válida, dependências e impacto em pedidos novos/em voo. Uma integração já existente SHALL ser configurável sem deploy para: endpoint homologado, mapeamento suportado, segredo referenciado, modo de obtenção do resultado final, ativação de polling, intervalos dentro de limites, política de rota e preço. O sistema NÃO SHALL permitir, via portal, introduzir nova semântica, protocolo, algoritmo ou transformação fora do conjunto seguro homologado; "configurável" NÃO SHALL ser interpretado como autorização para código arbitrário no portal.

#### Scenario: Reconfiguração de integração existente sem deploy
- GIVEN uma integração já homologada com endpoint, mapeamento e segredo referenciado válidos
- WHEN um operador ajusta intervalo de polling dentro dos limites permitidos, política de rota e preço pelo portal
- THEN o sistema SHALL aplicar a mudança sem exigir deploy, respeitando os limites configurados

#### Scenario: Tentativa de introduzir transformação fora do conjunto seguro é recusada
- GIVEN um operador tentando configurar, via portal, uma nova transformação ou algoritmo fora do conjunto seguro homologado
- WHEN a configuração é submetida
- THEN o sistema NÃO SHALL aceitar a mudança pelo portal, exigindo desenvolvimento e homologação prévios

#### Scenario: Tela de cadastro expõe campos obrigatórios, faixa válida e impacto
- GIVEN uma tela de cadastro de rota, destino, limite ou política de retenção
- WHEN o operador inicia o preenchimento
- THEN o sistema SHALL exibir campos obrigatórios, unidade, faixa válida, dependências e o impacto em pedidos novos e em voo antes da confirmação

### Requirement: CFG-02 — Publicação governada
O sistema SHALL impor o ciclo rascunho → validação estrutural → simulação com dados sintéticos → aprovação conforme risco → publicação versionada → observação → rollback por nova publicação, validando schemas, grafo, conta, capacidades, contratos, preços, quotas, prazo e retenção. O sistema NÃO SHALL publicar parcialmente um conjunto inconsistente. Um nó sem a versão necessária SHALL sincronizar ou recusar novas admissões, nunca misturar configuração antiga e nova; Atlas indisponível NÃO SHALL impedir execução com uma versão válida já conhecida. A revogação emergencial SHALL ter distribuição prioritária e independente da interface humana. A idade máxima da evidência de autorização SHALL ser explícita por perfil; vencida essa validade sem atualização confiável, o sistema SHALL bloquear somente as operações/escopos cuja autorização não possa ser comprovada, com prazo compatível com a continuidade aprovada em OPE-15 e sem prolongamento silencioso durante falha. Mudança de polling de operação em voo SHALL exigir ação explícita, faixa segura e trilha do motivo. Rollback de configuração SHALL criar nova versão referenciando o conteúdo anterior, sem apagar eventos de publicação ou pedidos já executados.

#### Scenario: Publicação parcial de conjunto inconsistente é recusada
- GIVEN um conjunto de configuração com schema, grafo, contrato, preço ou quota inconsistentes entre si
- WHEN a publicação é solicitada
- THEN o sistema NÃO SHALL publicar parcialmente o conjunto, recusando a publicação até a inconsistência ser corrigida

#### Scenario: Nó desatualizado recusa novas admissões em vez de misturar versões
- GIVEN um nó de execução que ainda não sincronizou a versão de configuração exigida por um pedido novo
- WHEN o pedido chega a esse nó
- THEN o sistema SHALL sincronizar a versão necessária ou recusar a nova admissão, nunca combinando configuração antiga e nova para o mesmo pedido

#### Scenario: Revogação emergencial é distribuída independentemente da interface humana
- GIVEN uma revogação emergencial de configuração ou credencial
- WHEN a revogação é disparada, mesmo com o portal administrativo indisponível
- THEN o sistema SHALL distribuir a revogação com prioridade, independente da interface humana estar acessível

#### Scenario: Evidência de autorização vencida bloqueia apenas o escopo afetado
- GIVEN uma evidência de autorização cuja idade máxima definida para o perfil foi ultrapassada sem atualização confiável
- WHEN operações dependem dessa evidência
- THEN o sistema SHALL bloquear somente as operações/escopos cuja autorização não possa ser comprovada, respeitando o prazo de continuidade aprovado em OPE-15, sem prolongar silenciosamente essa janela durante falha

### Requirement: CFG-03 — Manutenção e diagnóstico
O portal operacional SHALL permitir pesquisar por protocolo, tenant, cliente, conta de provedor, provider_request_id, período, estado e código de erro, respeitando a permissão do operador, exibindo linha do tempo com versões, passos, tentativas, resultado, entregas, uso e divergências. Payload sensível SHALL ser acessível somente por acesso excepcional auditado. O sistema SHALL oferecer como ações distintas: retomar timer, republicar evento, reprocessar inbox em quarentena, reconciliar operação, reenviar webhook, revisar resultado e reexecutar serviço, cada uma com escopo, pré-condição, motivo, responsável e efeito financeiro visível. Reexecução SHALL criar novo protocolo; replay técnico SHALL preservar a identidade do protocolo original. O sistema NÃO SHALL oferecer um botão genérico de "reprocessar tudo" que possa multiplicar chamadas e custos.

#### Scenario: Acesso a payload sensível exige acesso excepcional auditado
- GIVEN um protocolo cuja linha do tempo contém payload classificado como sensível
- WHEN um operador tenta visualizar esse payload
- THEN o sistema SHALL exigir acesso excepcional auditado, registrando essa exceção, em vez de exibir o payload por padrão

#### Scenario: Reexecução gera novo protocolo enquanto replay técnico preserva identidade
- GIVEN um protocolo concluído que precisa ser processado novamente
- WHEN o operador aciona reexecução em vez de replay técnico
- THEN o sistema SHALL criar um novo protocolo para a reexecução, e SHALL preservar a identidade do protocolo original apenas quando a ação for replay técnico

#### Scenario: Ação de reprocessamento em massa não está disponível
- GIVEN o conjunto de ações operacionais disponíveis no portal de manutenção
- WHEN um operador busca uma ação para reprocessar múltiplos protocolos de uma vez
- THEN o sistema NÃO SHALL oferecer um botão genérico "reprocessar tudo", exigindo que cada ação individual declare escopo, pré-condição, motivo, responsável e efeito financeiro

### Requirement: SEG-01 — Identidade e autorização
Clientes de máquina SHALL usar OAuth2 client credentials como padrão, com JWT de emissor/audiência/escopos homologados; usuários administrativos SHALL usar OIDC e MFA, com sessões e privilégios de prazo definido. API keys SHALL ser admitidas somente em integração legada explicitamente aprovada, com rotação e escopo. Comunicação HTTPS entre aplicações SHALL usar identidade de workload e mTLS; broker, banco e S3 SHALL usar TLS e os mecanismos de autenticação/menor privilégio suportados por cada serviço. O gateway SHALL eliminar cabeçalhos de identidade fornecidos pelo cliente e transmitir apenas contexto confiável. Os serviços SHALL validar identidade de workload e autorização de tenant/recurso, não confiando somente na rede interna. As permissões SHALL separar administrar catálogo, contrato, credenciais, executar, consultar, diagnosticar, ajustar financeiro, aprovar e confirmar liquidação. Um operador de runtime NÃO SHALL editar ledger publicado.

#### Scenario: Gateway elimina cabeçalho de identidade forjado pelo cliente
- GIVEN uma requisição de cliente contendo um cabeçalho de identidade forjado (por exemplo, um tenant_id diferente do autenticado)
- WHEN a requisição passa pelo gateway
- THEN o sistema SHALL eliminar o cabeçalho fornecido pelo cliente e transmitir aos serviços somente o contexto de identidade confiável estabelecido pelo próprio gateway

#### Scenario: Serviço interno não confia apenas na origem de rede interna
- GIVEN uma chamada originada da rede interna sem identidade de workload válida ou sem autorização de tenant/recurso comprovada
- WHEN essa chamada chega a um serviço
- THEN o sistema SHALL recusar a chamada, validando identidade de workload e autorização de tenant/recurso independentemente da origem de rede

#### Scenario: Operador de runtime não pode editar ledger publicado
- GIVEN um operador com permissão de execução e diagnóstico de runtime, mas sem permissão de confirmação de liquidação
- WHEN esse operador tenta editar um registro do ledger já publicado
- THEN o sistema NÃO SHALL permitir a edição, mantendo a separação entre a permissão de operar runtime e a de confirmar liquidação

### Requirement: SEG-02 — Destinos, callbacks e dados
URLs de provedores e clientes SHALL ser registradas com domínio, finalidade e métodos permitidos. O sistema SHALL prevenir SSRF: bloquear destinos de metadados e redes internas não autorizadas, revalidar a resolução DNS, controlar egress, impedir redirecionamento para destino não validado e NÃO SHALL seguir URLs retornadas pelo provedor sem política explícita. Certificados TLS SHALL ser validados, com mTLS exigido por contrato quando aplicável. Segredos SHALL residir em cofre, nunca em eventos, tabelas de configuração em claro, logs ou arquivos versionados. Chaves SHALL ter dono, escopo, expiração, rotação e revogação; assinaturas de callback SHALL suportar key_id e janela de sobreposição controlada. Dado pessoal NÃO SHALL virar label de Prometheus. Conteúdo XML SHALL impedir resolução de entidades externas; uploads SHALL receber controles por tipo/risco.

#### Scenario: Registro de destino em rede de metadados é bloqueado
- GIVEN uma tentativa de registrar um destino de callback apontando para um endereço de metadados de nuvem ou rede interna não autorizada
- WHEN o cadastro é validado
- THEN o sistema SHALL bloquear o registro desse destino, prevenindo SSRF

#### Scenario: Redirecionamento para destino não validado não é seguido
- GIVEN um destino de callback registrado e validado
- WHEN a resposta desse destino retorna um redirecionamento para uma URL não validada ou retornada dinamicamente pelo provedor sem política explícita
- THEN o sistema NÃO SHALL seguir esse redirecionamento

#### Scenario: Conteúdo XML com entidade externa é rejeitado
- GIVEN um payload XML recebido de um provedor ou cliente contendo uma declaração de entidade externa
- WHEN o payload é processado
- THEN o sistema SHALL impedir a resolução dessa entidade externa, rejeitando o vetor de ataque XXE

#### Scenario: Segredo não aparece em log, evento ou arquivo versionado
- GIVEN um segredo referenciado por uma credencial ativa
- WHEN eventos, tabelas de configuração, logs ou arquivos versionados são gerados pelo sistema
- THEN o sistema NÃO SHALL registrar o valor do segredo em nenhum desses destinos, mantendo-o exclusivamente no cofre

### Requirement: SEG-03 — Isolamento e auditoria
O sistema SHALL separar tenants em consulta, objeto, contrato, consumo, dashboards e ações operacionais, e SHALL separar ambientes por identidades, bancos, filas, buckets, credenciais e domínios. Dado real em ambiente de não produção SHALL exigir autorização e proteção compatível. A autorização negativa SHALL ser testada por troca de IDs, não apenas por navegação do portal. O sistema SHALL auditar publicação, suspensão, alteração de destino, uso excepcional de dados, replay, reexecução, conciliação, ajustes e exportações, com registro contendo ator, instante UTC, motivo, correlação e diferença segura de configuração. A auditoria de negócio durável SHALL ser separada dos logs Loki; perda ou expurgo de logs NÃO SHALL eliminar evidência de cobrança.

#### Scenario: Tentativa de acesso cross-tenant por troca de ID é negada
- GIVEN um cliente autenticado no tenant A que conhece ou deduz o identificador de um objeto pertencente ao tenant B
- WHEN o cliente troca o ID na requisição para tentar acessar esse objeto diretamente pela API, sem passar pela navegação do portal
- THEN o sistema SHALL negar o acesso, comprovando que o teste de autorização negativa cobre troca de IDs e não apenas o fluxo de portal

#### Scenario: Ação sensível gera registro de auditoria completo
- GIVEN uma ação de replay, reexecução, conciliação ou alteração de destino executada por um operador
- WHEN a ação é concluída
- THEN o sistema SHALL registrar ator, instante UTC, motivo, correlação e diferença segura de configuração na auditoria de negócio durável

#### Scenario: Expurgo de logs Loki não elimina evidência de cobrança
- GIVEN uma operação já registrada tanto nos logs Loki quanto na auditoria de negócio durável
- WHEN os logs Loki são perdidos ou expurgados por política de retenção operacional
- THEN o sistema NÃO SHALL perder a evidência de cobrança, pois ela permanece na auditoria de negócio durável, mantida separada dos logs

### Requirement: CFG-04 — Configuração dos novos limites
O sistema SHALL exigir, por família, a configuração obrigatória e a validação de publicação descritas no capítulo de origem para cliente/oferta, modalidade, credencial, expansão, retry, passo/provedor, polling, pressão, isolamento e administração. Todos os prazos configuráveis SHALL ser inteiros em segundos, podendo métricas manter precisão subsegundo; client_sla_seconds e provider_sla_seconds SHALL ser positivos; retry_ttl_seconds MAY ser zero; max_attempts SHALL incluir a primeira tentativa e SHALL ser ≥1; a reserva de finalização SHALL ser não negativa e menor que o prazo do cliente. Taxas e janelas de erro SHALL ter unidades e faixas explícitas; concorrência SHALL usar inteiros. O circuito aberto MAY zerar tráfego produtivo mantendo apenas sondagens autorizadas. O portal SHALL mostrar o valor resolvido e sua origem (contrato do cliente, produto, serviço, conta/provedor ou teto da plataforma), e NÃO SHALL permitir override silencioso do cliente que amplie sua autorização ou prazo acima do contrato. Alteração em voo NÃO SHALL estender deadline; exceções operacionais SHALL exigir novo protocolo/contrato, não reescrita histórica. Antes de publicar, o sistema SHALL simular caminho crítico, retry, polling, SLA e incidência financeira com casos sintéticos, declarando premissas sem comprovar capacidade de produção. O sistema SHALL permitir pausa e limite emergencial menor com motivo/auditoria; a retomada SHALL começar conservadoramente, sem restaurar instantaneamente o pico anterior.

#### Scenario: Publicação recusa configuração de retry com max_attempts inválido
- GIVEN uma configuração de retry com max_attempts menor que 1 ou com valores de taxa/concorrência negativos
- WHEN a publicação é validada
- THEN o sistema SHALL recusar a publicação até que os valores estejam dentro das faixas exigidas (inteiros, max_attempts ≥1, sem negativos)

#### Scenario: Portal exibe origem do valor resolvido e impede override silencioso
- GIVEN um limite efetivo resultante da combinação entre contrato do cliente, produto, serviço, conta/provedor e teto da plataforma
- WHEN o portal exibe esse valor ao operador ou ao cliente
- THEN o sistema SHALL mostrar o valor resolvido e sua origem, e NÃO SHALL permitir que o cliente amplie silenciosamente sua autorização ou prazo acima do contratado

#### Scenario: Alteração em voo não estende deadline do pedido
- GIVEN um pedido em voo cujo deadline já foi fixado no momento da admissão
- WHEN uma configuração de prazo é alterada durante a execução desse pedido
- THEN o sistema NÃO SHALL estender o deadline desse pedido em voo; uma exceção operacional SHALL exigir novo protocolo/contrato, sem reescrita histórica

#### Scenario: Retomada após limite emergencial começa conservadoramente
- GIVEN uma pausa ou limite emergencial reduzido aplicado com motivo e auditoria registrados
- WHEN a operação é retomada
- THEN o sistema SHALL iniciar a retomada de forma conservadora, sem restaurar instantaneamente o pico de tráfego anterior

### Requirement: SEG-04 — Administração entre tenants para desenvolvedores
Clientes SHALL permanecer estritamente limitados ao próprio tenant em protocolo, resultado, objeto, SLA, consumo e entrega; conhecer UUID, alterar header, usar URL de outra célula ou indicar tenant_id no corpo NÃO SHALL conferir acesso, e uma operação inexistente ou de outro tenant SHALL ter resposta que não revele existência. O sistema SHALL prover um perfil administrativo hub_protocol_reader, atribuído somente a identidades humanas individuais de desenvolvedores autorizados, capaz de localizar e consultar protocolos de qualquer tenant por API/rota administrativa própria, sem conta humana compartilhada, exigindo OIDC, MFA, grupos aprovados, expiração/revisão de concessão e separação entre produção/não produção. Consultar todos os tenants NÃO SHALL conceder automaticamente segredos, payload integral, edição, replay, reexecução ou poderes financeiros; por padrão dados sensíveis SHALL ser mascarados. Download, busca e visualização SHALL ser auditados com identidade, tenant alvo, protocolo, horário, finalidade e ambiente. O sistema NÃO SHALL aceitar função administrativa por flag em JWT emitido ao cliente ou header arbitrário. A aplicação SHALL autorizar explicitamente a consulta entre tenants e registrar o alvo; o sistema NÃO SHALL usar credencial de superusuário PostgreSQL nem desativar globalmente RLS para disponibilizar o portal. Relatórios globais SHALL usar projeções de leitura e paginação com limites, sem varreduras irrestritas ou fan-out ilimitado nos bancos de execução; se apenas parte das células responde, o sistema SHALL apresentar resultado parcial administrativo com indicação de incompletude, sem declarar protocolo inexistente globalmente.

#### Scenario: Tentativa de acesso cross-tenant por cliente é negada sem revelar existência
- GIVEN um cliente autenticado no tenant A que conhece o UUID de um protocolo do tenant B, altera um header ou indica outro tenant_id no corpo da requisição
- WHEN essa requisição é processada
- THEN o sistema NÃO SHALL conceder acesso, e a resposta para operação inexistente ou de outro tenant SHALL ser indistinguível, sem revelar a existência do protocolo alvo

#### Scenario: hub_protocol_reader consulta protocolo de outro tenant com dados mascarados e auditados
- GIVEN um desenvolvedor autorizado com identidade individual, OIDC e MFA válidos, atribuído ao perfil hub_protocol_reader
- WHEN esse desenvolvedor consulta um protocolo de um tenant diferente pela rota administrativa própria
- THEN o sistema SHALL permitir a consulta com dados sensíveis mascarados por padrão, e SHALL registrar identidade, tenant alvo, protocolo, horário, finalidade e ambiente na auditoria

#### Scenario: Privilégio administrativo não é aceito por flag em JWT de cliente
- GIVEN um JWT emitido para um cliente, contendo um campo ou header arbitrário que declara privilégio administrativo
- WHEN uma requisição usa esse JWT para tentar acessar a rota administrativa entre tenants
- THEN o sistema NÃO SHALL aceitar essa função administrativa, pois o perfil hub_protocol_reader só é reconhecido via OIDC/MFA de identidade humana aprovada

#### Scenario: Portal administrativo não usa superusuário PostgreSQL nem desativa RLS globalmente
- GIVEN a necessidade de disponibilizar consulta administrativa entre tenants
- WHEN o mecanismo de acesso é implementado no banco/serviço
- THEN o sistema NÃO SHALL usar credencial de superusuário PostgreSQL nem desativar globalmente RLS, oferecendo em vez disso uma política de leitura administrativa controlada, mantendo todos os testes negativos de cliente obrigatórios

#### Scenario: Resposta parcial quando apenas parte das células responde
- GIVEN uma busca administrativa entre células na qual apenas parte delas responde dentro do prazo
- WHEN o resultado é consolidado
- THEN o sistema SHALL apresentar um resultado parcial administrativo com indicação explícita de incompletude, sem declarar o protocolo inexistente globalmente

### Requirement: CFG-05 — Gestão de credenciais Hub–cliente–provedor
O portal SHALL administrar metadados e ciclo de vida do vínculo de credencial, com o segredo escrito no cofre por fluxo restrito e SEM leitura posterior em claro pela interface; a máscara exibida NÃO SHALL revelar parte suficiente para reutilizar o segredo. O sistema SHALL exigir credential_mode explícito (SHARED_HUB ou TENANT_DEDICATED, nunca derivado da ausência de campo), relacionamento (provider_id, provider_account_id, serviço/vínculo, ambiente, com tenant_id obrigatório e único no modo dedicado), autenticação homologada, referência (credential_binding_id, versão de metadados, secret_ref, política de versão/rotação, com valores secretos só no cofre), estado do ciclo de vida (RASCUNHO, EM_VALIDACAO, ATIVO, EM_ROTACAO, SUSPENSO, REVOGADO, EXPIRADO), continuidade, capacidade e economia. O cadastro SHALL validar que o operador pode administrar aquele tenant/provedor, que a conta pertence ao escopo e que não existe resolução ambígua. Teste de credencial SHALL usar endpoint sem efeito quando fornecido; a ausência de endpoint de teste NÃO SHALL autorizar executar pedido real invisível. Ativação SHALL exigir credencial válida, contrato coerente, isolamento de segredo, política de rotação e teste de acesso. Troca de segredo NÃO SHALL trocar automaticamente de conta externa. Cliente MAY fornecer/rotacionar sua própria credencial por permissão específica, mas NÃO SHALL escolher secret_ref arbitrário nem ler o cofre ou a chave compartilhada do Hub.

#### Scenario: Interface não permite leitura de segredo em claro
- GIVEN um vínculo de credencial ativo com segredo armazenado no cofre
- WHEN o gestor visualiza os metadados desse vínculo no portal
- THEN o sistema SHALL exibir apenas identificador, tipo, titular, conta externa, escopo, ambiente, validade e estado, e NÃO SHALL permitir leitura posterior do segredo em claro nem exibir máscara suficiente para reutilizá-lo

#### Scenario: Ausência de endpoint de teste não autoriza execução real invisível
- GIVEN um vínculo de credencial sem endpoint de teste sem efeito configurado
- WHEN o operador tenta validar essa credencial
- THEN o sistema NÃO SHALL executar um pedido real e invisível ao cliente apenas para testar a credencial, exigindo operação de homologação explícita com custo identificado caso o teste precise executar o serviço

#### Scenario: Cliente não pode escolher secret_ref arbitrário nem ler cofre do Hub
- GIVEN um cliente com permissão específica para fornecer/rotacionar sua própria credencial
- WHEN esse cliente tenta especificar um secret_ref arbitrário ou acessar o cofre/chave compartilhada do Hub
- THEN o sistema NÃO SHALL permitir essa escolha ou leitura, restringindo o cliente à rotação de sua própria credencial dentro do fluxo autorizado

#### Scenario: Suspensão de vínculo afeta apenas os protocolos daquele vínculo
- GIVEN múltiplos vínculos de credencial ativos para o mesmo tenant/provedor
- WHEN um vínculo específico é suspenso
- THEN o sistema SHALL suspender apenas esse vínculo, localizando os protocolos em voo afetados e exibindo a repercussão em SLA, polling e custo, sem afetar os demais vínculos

### Requirement: CFG-06 — Onboarding e capacidade como serviço interno
O sistema SHALL conduzir o onboarding por fluxo declarativo RASCUNHO → VALIDADO → RESERVANDO_CAPACIDADE → PROVISIONANDO (quando necessário) → QUALIFICANDO → ATIVO, mantendo uma falha em PENDENTE_COM_MOTIVO com preservação de recursos/progresso. Atlas SHALL calcular o placement usando perfil, região permitida, criticidade, carga prevista, capacidade já reservada e limites de conta de provedor, reservando capacidade de forma concorrente e idempotente para impedir que duas ativações prometam a mesma folga. Antes de ATIVO, o sistema SHALL confirmar dependências prontas, schema compatível, probes, quotas, credenciais, conectividade, objetos, telemetry/alertas, capacidade N−1 exigida e ensaio sintético do perfil, sem que os testes de prontidão substituam a qualificação prévia do template. O processo SHALL ser retomável após crash, sem criar contas/filas duplicadas nem publicar rota antes de persistir o estado aprovado. Teto de nuvem, ausência de região permitida, restrição comercial e orçamento esgotado SHALL aparecer como impedimentos explícitos antes da saturação. Clientes ativos SHALL conservar sua reserva enquanto novas ativações aguardam. Desativar cliente SHALL drenar/reconciliar obrigações e aplicar retenção, sem destruir automaticamente recursos compartilhados.

#### Scenario: Reserva de capacidade concorrente e idempotente evita dupla promessa
- GIVEN duas ativações concorrentes competindo pela mesma folga de capacidade em uma célula
- WHEN ambas solicitam reserva de capacidade simultaneamente
- THEN o sistema SHALL garantir, de forma idempotente, que apenas uma ativação obtenha a folga reservada, sem prometer a mesma capacidade duas vezes

#### Scenario: Onboarding retomado após crash não duplica recursos nem publica rota prematuramente
- GIVEN um onboarding interrompido por uma falha do processo após a reserva de capacidade, mas antes da persistência do estado ATIVO
- WHEN o processo é retomado
- THEN o sistema SHALL continuar do estado persistido sem criar contas ou filas duplicadas e sem publicar a rota de tráfego antes de persistir o estado aprovado

#### Scenario: Orçamento esgotado aparece como impedimento explícito antes da saturação
- GIVEN uma nova ativação cujo orçamento de capacidade já está esgotado
- WHEN a ativação é solicitada
- THEN o sistema SHALL apresentar o orçamento esgotado como impedimento explícito antes de atingir saturação, preservando a reserva dos clientes já ativos enquanto a nova ativação aguarda

#### Scenario: Desativação de cliente drena obrigações sem destruir recursos compartilhados
- GIVEN um cliente com protocolos em voo e capacidade parcialmente compartilhada com outros tenants
- WHEN o cliente é desativado
- THEN o sistema SHALL drenar e reconciliar as obrigações pendentes e aplicar a retenção definida, sem destruir automaticamente os recursos compartilhados usados por outros tenants

### Requirement: SEG-05 — Resolução, rotação e isolamento de segredo
O sistema SHALL resolver credenciais de forma determinística: obter tenant da identidade confiável, fixar contrato técnico/comercial e rota, resolver vínculo elegível de conta/serviço/ambiente, exigir o modo previsto, validar estado/validade/permissão, obter a versão de segredo autorizada e registrar a referência da versão efetivamente utilizada na tentativa; o cliente NÃO SHALL escolher outra conta por parâmetro. Se o modo é TENANT_DEDICATED e a credencial está ausente, expirada ou revogada, o sistema NÃO SHALL usar SHARED_HUB nem credencial de outro tenant como fallback, suspendendo o vínculo/recusando antes do envio ou conservando a pendência segundo prazo, sem alterar o pagador; failover entre provedores SHALL exigir também um vínculo de credencial explicitamente elegível para o mesmo cliente, e a indisponibilidade de uma credencial dedicada NÃO SHALL desligar conectores de outros vínculos saudáveis. O cache de segredo/tokens SHALL ser local ao workload autorizado, limitado e separado por ambiente, provedor/conta, binding, versão, audience e escopos, e NÃO SHALL ser armazenado em Redis comum, eventos, backups de configuração, logs, traces ou interface. A rotação SHALL preparar versão nova na mesma conta, testar, pré-aquecer workloads, publicar política/versionamento, usar a nova versão em novas tentativas, manter sobreposição somente quando provedor e segurança permitirem, e encerrar a anterior após validação das operações em voo; a revogação emergencial SHALL prevalecer sobre conveniência operacional, sem esperar o término do TTL de retry. Com o cofre fora do ar, o material já obtido MAY ser usado apenas até o menor limite de expiração externa, validade local e política de revogação, com atualização antecipada e jitter para evitar tempestade de refresh. O perfil hub_protocol_reader MAY ver metadados permitidos e diagnósticos, mas NÃO SHALL obter o valor secreto.

#### Scenario: Credencial dedicada indisponível não recorre a fallback compartilhado ou de outro tenant
- GIVEN um vínculo TENANT_DEDICATED cuja credencial está ausente, expirada ou revogada no momento da resolução
- WHEN uma tentativa de envio depende dessa credencial
- THEN o sistema NÃO SHALL usar SHARED_HUB nem a credencial de outro tenant como fallback, suspendendo o vínculo ou recusando o envio conforme a política de prazo, sem alterar o pagador da operação

#### Scenario: Rotação de segredo mantém sobreposição controlada e preserva operações em voo
- GIVEN uma rotação de segredo com nova versão preparada, testada e pré-aquecida na mesma conta
- WHEN a nova versão é publicada e usada em novas tentativas
- THEN o sistema SHALL manter a versão anterior ativa apenas durante a sobreposição controlada necessária para validar as operações em voo, encerrando-a somente depois dessa validação

#### Scenario: Cofre indisponível limita uso do material cacheado ao menor limite de validade
- GIVEN o cofre de segredos fora do ar e uma instância de workload com material de credencial previamente obtido em cache
- WHEN essa instância tenta usar o material cacheado
- THEN o sistema MAY usá-lo apenas até o menor entre o limite de expiração externa, a validade local e a política de revogação, recusando o uso além desse limite mesmo sem contato com o cofre

#### Scenario: Perfil hub_protocol_reader não obtém valor secreto
- GIVEN um administrador com perfil hub_protocol_reader consultando diagnósticos de um protocolo de outro tenant
- WHEN esse administrador visualiza os metadados de credencial associados ao protocolo
- THEN o sistema SHALL exibir apenas metadados e diagnósticos permitidos, e NÃO SHALL expor o valor do segredo em nenhuma circunstância

#### Scenario: Revogação emergencial prevalece sobre TTL de retry em voo
- GIVEN uma operação em voo aguardando nova tentativa dentro do retry_ttl_seconds configurado, usando uma credencial que sofre revogação emergencial
- WHEN a revogação emergencial é aplicada
- THEN o sistema SHALL interromper o uso dessa credencial imediatamente, sem esperar o término do TTL de retry, registrando o impedimento e reconciliando a operação sem redirecioná-la para outra conta

## Notas de origem

Este delta deriva integralmente do capítulo `docs/06_CONFIGURACAO_E_SEGURANCA.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos CFG-01, CFG-02, CFG-03, SEG-01, SEG-02, SEG-03, CFG-04, SEG-04, CFG-05, CFG-06 e SEG-05 conforme texto normativo de origem.
