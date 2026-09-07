# Execução e Integrações — Delta de Especificação

## ADDED Requirements

### Requirement: EXE-01 — Admissão durável
O sistema SHALL admitir uma requisição de execução somente quando consumidor autenticado/autorizado, oferta e contrato elegíveis, payload válido, objetos disponíveis, capacidade de admissão e versão de configuração válida estiverem satisfeitos, e SHALL atribuir protocol_id UUIDv7 a toda requisição aceita; entrada recusada SHALL gerar trilha técnica segura, sem obrigação de execução ou receita, salvo evento comercial explicitamente contratado e distinto de consumo do serviço. Idempotency-Key SHALL ser obrigatória na criação, com unicidade por tenant, operação pública e chave, associando hash semântico de oferta/versão, contrato técnico de entrada/transformação fixado, input canônico e objetos imutáveis; esse hash NÃO SHALL incluir timestamps de trace ou URLs temporárias. Repetição após timeout do cliente SHALL usar a mesma chave. O sistema SHALL confirmar aceite, em qualquer modalidade, somente após commit; em ASYNC, SHALL responder 202 apenas nessa condição. Sem banco de admissão disponível, o sistema SHALL responder indisponibilidade e NÃO SHALL confirmar aceitação apenas porque a mensagem entrou em memória. UUID NÃO SHALL ser tratado como senha nem substituir autorização.

#### Scenario: Mesma chave e mesmo hash recuperam o protocolo existente
- GIVEN um consumidor reenvia a mesma Idempotency-Key com o mesmo hash semântico de oferta/versão e input canônico
- WHEN a requisição é processada pela Órbita
- THEN o sistema SHALL retornar o mesmo protocol_id e o mesmo estado já existente, sem criar novo protocolo

#### Scenario: Mesma chave com payload diferente retorna conflito
- GIVEN a mesma Idempotency-Key é reenviada pelo mesmo tenant e operação, mas com hash semântico diferente do payload original
- WHEN a Órbita compara a nova requisição com a chave já registrada
- THEN o sistema SHALL retornar conflito, sem reutilizar ou sobrescrever o protocolo original

#### Scenario: Resposta perdida após commit é recuperada pela repetição da mesma chave
- GIVEN o commit de admissão ocorreu, mas a resposta HTTP se perdeu antes de chegar ao cliente
- WHEN o cliente repete a requisição com a mesma Idempotency-Key
- THEN o sistema SHALL recuperar o mesmo protocolo já commitado, sem executar novo efeito nem gerar novo protocol_id

#### Scenario: Ausência de banco de admissão impede confirmar aceite em memória
- GIVEN o banco de admissão está indisponível no momento da requisição
- WHEN a Órbita tenta processar o aceite
- THEN o sistema SHALL responder indisponibilidade, NÃO SHALL confirmar aceitação apenas porque a mensagem chegou a residir em memória

### Requirement: EXE-02 — Modos de atendimento
O sistema SHALL atender cada combinação de provedor síncrono/assíncrono e modo escolhido pelo cliente (SYNC, ASYNC, AUTO) conforme a tabela normativa: provedor síncrono em SYNC SHALL despachar diretamente e retornar final durável na mesma requisição quando concluído no orçamento, sem espera obrigatória por fila e sem conversão para 202; provedor síncrono em ASYNC SHALL responder 202 após aceite, com worker chamando o provedor e resultado local mais webhook conforme contrato; provedor síncrono em AUTO SHALL aguardar o processamento em fila conforme contratado, retornando 200 final se disponível ou 202 com o mesmo protocolo, sem equivaler a SYNC; provedor assíncrono em ASYNC SHALL responder 202, com obtenção do final por callback, polling ou combinação; provedor assíncrono em AUTO SHALL aguardar de forma limitada pelo resultado durável, com final ou 202, sem manter conexão por todo o prazo do provedor; provedor assíncrono em SYNC NÃO SHALL ser ofertado na referência v4, exigindo nova decisão e qualificação de orçamento para exceção. Toda admissão SHALL receber UUIDv7. No perfil canônico, conclusão válida SHALL retornar 200 com estado final e corpo persistido, e falha de negócio também SHALL estar representada no contrato. Prazo SYNC esgotado SHALL retornar 504 com o final de expiração se já persistido; falha de infraestrutura que impeça finalização SHALL ser indicada como incerteza, com o protocolo conhecido, e o sistema NÃO SHALL anunciar como final um corpo que só exista em memória. Em AUTO, o prazo da conexão MAY terminar antes do prazo de negócio, produzindo 202; em SYNC, o prazo efetivo do protocolo SHALL caber na conexão conforme EXE-14. O limite de oito segundos da v3 NÃO SHALL ser tratado como default universal; a publicação SHALL exigir valores do cliente, da borda, do serviço e da reserva final. Encerrar a conexão NÃO SHALL cancelar efeito externo já enviado; webhook adicional SHALL ser opção explícita, usando o mesmo resultado e sem receita duplicada.

#### Scenario: Provedor síncrono em SYNC não converte para 202
- GIVEN o provedor é síncrono e o cliente escolheu o modo SYNC
- WHEN a execução conclui dentro do orçamento da conexão
- THEN o sistema SHALL retornar o final durável na mesma requisição, SHALL NOT converter a resposta para 202 nem esperar obrigatoriamente por fila

#### Scenario: Provedor síncrono em AUTO aguarda fila sem equivaler a SYNC
- GIVEN o provedor é síncrono e o cliente escolheu o modo AUTO
- WHEN o processamento ocorre em fila conforme contratado
- THEN o sistema SHALL retornar 200 final se disponível dentro da espera contratada, ou 202 com o mesmo protocolo caso contrário, SHALL NOT tratar esse comportamento como equivalente a SYNC

#### Scenario: Prazo SYNC esgotado retorna 504 com final de expiração já persistido
- GIVEN uma execução em modo SYNC atinge o prazo da conexão
- WHEN o final de expiração já foi persistido antes do esgotamento
- THEN o sistema SHALL retornar 504 com esse final persistido, e caso a finalização não tenha sido possível por falha de infraestrutura, SHALL indicar incerteza com o protocolo conhecido, sem anunciar como final um corpo que só exista em memória

#### Scenario: Provedor assíncrono em SYNC não é ofertado sem nova decisão
- GIVEN um provedor assíncrono e uma solicitação de atendê-lo no modo SYNC
- WHEN essa combinação é avaliada contra a referência v4
- THEN o sistema NÃO SHALL ofertar essa combinação; qualquer exceção SHALL exigir nova decisão e qualificação de orçamento, não uma simples flag de configuração

### Requirement: EXE-03 — Estados e autoridade
O sistema SHALL controlar a máquina de estados normativa: Protocolo SHALL transitar entre ACCEPTED, RUNNING, WAITING_PROVIDER, RECONCILING e os finais SUCCEEDED, PARTIALLY_SUCCEEDED, FAILED, EXPIRED, CANCELLED; Passo SHALL usar PENDING, READY, RUNNING, WAITING_PROVIDER, RECONCILING e finais SUCCEEDED, FAILED, SKIPPED, CANCELLED; Operação externa SHALL usar PREPARED, SUBMITTING, ACCEPTED_EXTERNAL, WAITING_FINAL, UNKNOWN e finais SUCCEEDED, FAILED, CANCELLED; Recibo externo SHALL usar RECEIVED, UNCORRELATED, APPLIED, DUPLICATE, CONFLICT, REJECTED, REJECTED_LATE_SLA; Entrega SHALL usar PENDING, DELIVERING, RETRY_SCHEDULED, DELIVERED, SUSPENDED, EXHAUSTED; Resultado SHALL manter versões finais imutáveis com motivo e vínculo da revisão, com ponteiro atual controlado pela Órbita. ACCEPTED SHALL transitar para RUNNING após despacho direto ou agendamento; RUNNING SHALL transitar para WAITING_PROVIDER quando existir operação assíncrona; incerteza SHALL ir para RECONCILING. Cometa SHALL decidir apenas o fato externo; Órbita SHALL decidir a conclusão do passo/produto. Sucesso SHALL exigir saída válida e disponível; parcialidade SHALL exigir política publicada. Eventos intermediários atrasados NÃO SHALL regredir estado final. FAILED SHALL exigir falha conhecida segundo contrato; UNKNOWN NÃO SHALL virar FAILED por conveniência. EXPIRED SHALL encerrar o prazo de negócio, podendo manter operação incerta em reconciliação e obrigação financeira separada. CANCELLED SHALL exigir inexistência comprovada de efeitos pendentes ou confirmação externa apropriada; pedido de cancelamento SHALL permanecer como solicitação até então. Reconciliar MAY revelar custo mesmo depois de expiração.

#### Scenario: Cometa decide o fato externo, Órbita decide a conclusão do passo
- GIVEN uma operação externa observada por Cometa
- WHEN Cometa registra o fato correspondente
- THEN o sistema SHALL restringir a decisão sobre esse fato externo a Cometa, e a decisão sobre a conclusão do passo/produto SHALL permanecer exclusiva da Órbita

#### Scenario: UNKNOWN não vira FAILED por conveniência
- GIVEN uma operação externa permanece em estado UNKNOWN sem falha conhecida segundo o contrato
- WHEN o sistema avalia a conclusão do passo
- THEN o sistema NÃO SHALL transicionar esse estado para FAILED apenas por conveniência operacional, mantendo a incerteza registrada

#### Scenario: CANCELLED exige inexistência comprovada de efeitos pendentes
- GIVEN um pedido de cancelamento é recebido para um protocolo com operação externa em curso
- WHEN o sistema avalia se pode concluir CANCELLED
- THEN o sistema SHALL exigir inexistência comprovada de efeitos pendentes ou confirmação externa apropriada antes de CANCELLED, mantendo o pedido apenas como solicitação enquanto isso não ocorrer

#### Scenario: Evento intermediário atrasado não regride estado final
- GIVEN um protocolo já alcançou um estado final, como SUCCEEDED
- WHEN um evento intermediário atrasado chega posteriormente
- THEN o sistema NÃO SHALL regredir o estado final já alcançado por causa desse evento atrasado

### Requirement: EXE-04 — Operação e correlação
Antes de enviar uma chamada externa, o sistema SHALL registrar operation_id, protocol_id, step_id, provedor, conta, versão do vínculo, modo e binding_id de credencial, secret_version_id efetivamente usado, contrato de aquisição e chave de idempotência externa quando suportada. Cada chamada física SHALL criar attempt_id e tipo de interação. Reenvio de transporte SHALL preservar operation_id e chave externa; troca autorizada de provedor SHALL criar outra operação vinculada ao mesmo passo. provider_request_id SHALL ser associado no escopo documentado do provedor/conta/ambiente, e o sistema NÃO SHALL presumir unicidade global desse identificador. Quando possível, o sistema SHALL enviar token de correlação opaco previamente persistido. Callback sem referência externa ainda associada SHALL ser persistido como UNCORRELATED, com reconciliação posterior; o sistema NÃO SHALL associar esse callback por nome, CPF ou proximidade de horário. O sistema SHALL registrar timestamps de preparação/envio/recebimento, resultado de autenticação, erro canônico, hash/evidência e classificação de efeito conhecido ou incerto. Se a chamada saiu mas o processo caiu antes de persistir a resposta, o sistema SHALL reconstruir a situação por chave idempotente, status externo ou investigação, e NÃO SHALL inventar uma resposta final.

#### Scenario: Callback sem correlação vira UNCORRELATED, não associado por heurística
- GIVEN um callback chega sem referência externa ainda associada a uma operação conhecida
- WHEN o Cometa processa esse callback
- THEN o sistema SHALL persisti-lo como UNCORRELATED com reconciliação posterior, NÃO SHALL associá-lo por nome, CPF ou proximidade de horário

#### Scenario: Troca autorizada de provedor cria nova operação vinculada ao mesmo passo
- GIVEN uma troca de provedor autorizada para um passo já iniciado
- WHEN a nova chamada é preparada
- THEN o sistema SHALL criar uma nova operação vinculada ao mesmo passo, preservando a rastreabilidade, em vez de reutilizar a operação anterior

#### Scenario: Crash após envio sem resposta persistida exige reconstrução, não invenção de final
- GIVEN a chamada externa foi enviada mas o processo caiu antes de persistir a resposta
- WHEN o sistema tenta recuperar essa operação
- THEN o sistema SHALL reconstruir a situação por chave idempotente, status externo ou investigação, NÃO SHALL inventar uma resposta final

### Requirement: EXE-05 — Polling configurável
O sistema SHALL tratar "polling" como consulta periódica ao provedor, distinta de pool de conexões, e SHALL suportar por vínculo os modos CALLBACK_ONLY, POLLING_ONLY e CALLBACK_WITH_POLLING_FALLBACK. Ativação de polling no portal SHALL ser permitida somente se o adaptador declarar capacidade de consulta homologada e o contrato permitir seu consumo. A configuração por vínculo SHALL incluir referência da consulta (origem do provider_request_id, mapeamento no endpoint, autenticação por referência de segredo), mapa completo de estados externos para pendente/sucesso/falha/desconhecido (valores novos SHALL ir para diagnóstico), forma de obtenção do resultado (junto do status ou FETCH separada, com limites e custo de ambas), agendamento (atraso inicial, intervalo mínimo/máximo, fator de backoff, jitter, fallback após ausência de callback), limites (deadline, máximo de tentativas, consultas por segundo, concorrência por conta/serviço), tratamento de erros (429/Retry-After, timeout, 401/403, inexistente, payload inválido, indisponibilidade), economia (medidor e tarifa por consulta/fetch, franquia, teto de custo) e operação (pausa, retomada, aging, versão, autorização para mudar política em voo). Defaults propostos para qualificação — atraso inicial 2 s; intervalo 5 s; backoff fator 2 até 60 s; jitter ±20%; callback com fallback após 30 s — NÃO SHALL ser aplicados sem aprovação do vínculo e compatibilidade com o deadline; deadline e número máximo de tentativas SHALL ser obrigatórios e dependentes do serviço, sem default universal. O sistema SHALL respeitar Retry-After dentro do prazo; se ultrapassa o deadline, SHALL tratar como expiração/reconciliação. A agenda durável SHALL guardar next_run_at, contadores, lease, token de posse e versão; um worker SHALL reclamar trabalho de forma exclusiva, sem manter thread ou transação dormindo por toda a operação, e SHALL retomar vencidos após reinício com controle de avalanche. Final persistido ou deadline atingido SHALL encerrar agendamentos produtivos futuros; consultas residuais de reconciliação SHALL ter finalidade e orçamento separados. Pausar polling NÃO SHALL cancelar operação nem apagar prazo; retomada SHALL considerar o tempo transcorrido.

#### Scenario: Defaults propostos não se aplicam sem aprovação do vínculo
- GIVEN os defaults propostos de agendamento — atraso inicial 2 s, intervalo 5 s, backoff fator 2 até 60 s, jitter ±20%, fallback após 30 s — para um vínculo em qualificação
- WHEN esse vínculo é avaliado para ativação em produção
- THEN o sistema NÃO SHALL aplicar esses defaults sem aprovação explícita do vínculo e compatibilidade verificada com o deadline do serviço

#### Scenario: Retry-After respeitado dentro do prazo, expiração tratada além do deadline
- GIVEN o provedor retorna 429 com um Retry-After que ultrapassa o deadline restante da operação
- WHEN o Cometa avalia se deve reagendar a consulta
- THEN o sistema SHALL respeitar o Retry-After enquanto couber no prazo, e SHALL tratar como expiração/reconciliação quando o Retry-After ultrapassar o deadline

#### Scenario: Worker reclama trabalho de forma exclusiva sem manter transação dormindo
- GIVEN uma agenda durável de polling com next_run_at, lease e token de posse
- WHEN um worker reclama esse trabalho para execução
- THEN o sistema SHALL garantir reclamação exclusiva sem manter thread ou transação dormindo por toda a operação, retomando trabalhos vencidos após reinício com controle de avalanche

#### Scenario: Pausar polling não cancela a operação nem apaga o prazo
- GIVEN o polling de uma operação é pausado por decisão operacional
- WHEN a operação é retomada posteriormente
- THEN o sistema NÃO SHALL cancelar a operação nem apagar o prazo por causa da pausa, e a retomada SHALL considerar o tempo já transcorrido

### Requirement: EXE-06 — Concorrência entre callback e polling
O sistema SHALL tratar callback e polling como submissões concorrentes ao mesmo consolidado da operação externa. Se callback e polling apresentam o mesmo final, o sistema SHALL aplicar uma única transição e tratar o segundo como evidência duplicada. Uma consulta já em voo MAY terminar após cancelamento da agenda; sua tentativa e possível custo SHALL continuar registráveis, sem republicar conclusão equivalente. Se houver finais conflitantes, o sistema SHALL preservar ambos e marcar CONFLICT para resolução com a regra homologada do provedor; o último recebido NÃO SHALL prevalecer automaticamente. Se o provedor possui sequência/revisão autoritativa, ela SHALL participar da resolução; sem isso, o sistema SHALL exigir reconciliação. Resposta "não encontrado" logo após submit MAY ser consistência eventual: o sistema SHALL usar a janela documentada, NÃO SHALL reiniciar a operação às cegas.

#### Scenario: Callback e polling com o mesmo final aplicam transição única
- GIVEN callback e polling reportam o mesmo final para a mesma operação externa
- WHEN ambas as observações chegam ao consolidado
- THEN o sistema SHALL aplicar uma única transição de estado, tratando a segunda observação como evidência duplicada

#### Scenario: Finais conflitantes preservam ambos e marcam CONFLICT
- GIVEN callback e polling reportam finais conflitantes para a mesma operação
- WHEN o sistema consolida essas observações
- THEN o sistema SHALL preservar ambos os registros e marcar CONFLICT para resolução segundo a regra homologada do provedor, NÃO SHALL prevalecer automaticamente o último recebido

#### Scenario: "Não encontrado" logo após submit usa janela documentada, sem reiniciar às cegas
- GIVEN uma consulta imediatamente após o submit retorna "não encontrado" por possível consistência eventual do provedor
- WHEN o sistema avalia essa resposta
- THEN o sistema SHALL usar a janela documentada de tolerância antes de qualquer conclusão, NÃO SHALL reiniciar a operação às cegas

### Requirement: EXE-07 — Contrato de consultas públicas
O sistema SHALL implementar o contrato de consultas públicas: criar pedido SHALL seguir SYNC direto com final na mesma conexão conforme EXE-14, ASYNC SHALL responder 202, AUTO SHALL responder final ou 202, e todo aceite SHALL identificar UUIDv7 e endereço de consulta; consultar protocolo SHALL retornar 200 com a representação contratada de estado/resultado, igual ao corpo do webhook quando final, e 404 para protocolo inexistente ou de outro tenant; consultar resultado pendente SHALL retornar 200 com a variante pendente do contrato unificado, sendo o endpoint de resultado, se mantido, um alias; consultar resultado final SHALL retornar 200 com a mesma representação final persistida usada pelo webhook, incluindo falha/EXPIRED no contrato do cliente; consultar resultado expurgado SHALL retornar 410 RESULT_EXPIRED quando metadado autorizado permitir reconhecê-lo, sem nova busca externa; solicitar cancelamento SHALL registrar intenção durável, e o retorno NÃO SHALL afirmar cancelamento externo concluído; consultar entregas SHALL retornar estado e tentativas em projeção, informando instante de atualização; atualizar/reexecutar SHALL criar nova criação com novo protocolo e referência ao anterior, mediante avaliação comercial explícita. O resultado SHALL conter protocol_id, client_request_id, status, result_version, schema_version, submitted_at, completed_at, partes do produto, códigos canônicos, conteúdo pequeno ou object_ids e informações autorizadas de download. O estado HTTP de consulta bem-sucedida NÃO SHALL confundir erro de negócio com falha da infraestrutura HTTP. Nenhuma consulta pública SHALL incrementar o medidor de execução do provedor; cobrança pelo acesso a resultado armazenado SHALL existir apenas como medidor separado e explicitamente vendido, e NÃO SHALL integrar o plano padrão v4, mesmo em caso de cache miss, resultado expurgado ou provedor indisponível.

#### Scenario: Consultar resultado expurgado retorna 410 sem nova busca externa
- GIVEN um resultado foi expurgado legitimamente e o metadado autorizado permite reconhecê-lo
- WHEN o cliente consulta esse resultado
- THEN o sistema SHALL retornar 410 RESULT_EXPIRED, NÃO SHALL realizar nova busca externa ao provedor

#### Scenario: Solicitar cancelamento registra intenção durável sem afirmar conclusão externa
- GIVEN um cliente solicita cancelamento de um protocolo em execução
- WHEN o sistema processa esse pedido
- THEN o sistema SHALL registrar a intenção durável de cancelamento, e o retorno NÃO SHALL afirmar que o cancelamento externo foi efetivamente concluído

#### Scenario: Consultas públicas não incrementam o medidor de execução do provedor
- GIVEN um cliente realiza uma consulta de protocolo, resultado ou entregas
- WHEN essa consulta é processada, inclusive em caso de cache miss ou resultado expurgado
- THEN o sistema NÃO SHALL incrementar o medidor de execução do provedor por causa dessa consulta; eventual cobrança de acesso a resultado armazenado SHALL existir apenas como medidor separado explicitamente vendido

#### Scenario: Atualizar/reexecutar cria novo protocolo com referência ao anterior
- GIVEN um cliente solicita atualização ou reexecução de um protocolo já concluído
- WHEN essa solicitação é avaliada comercialmente
- THEN o sistema SHALL criar um novo protocolo com referência ao protocolo anterior, mediante avaliação comercial explícita, sem reutilizar o protocolo original

### Requirement: EXE-08 — Webhooks
Na recepção de callback de provedor, o sistema SHALL autenticar conforme capacidade homologada (HMAC, mTLS ou token dedicado), limitar tamanho/taxa, persistir o recibo e só então responder 2xx; IP permitido SHALL ser proteção complementar, não substituto de autenticação quando esta for possível, e callback inválido NÃO SHALL alterar o protocolo. No envio ao cliente, o sistema SHALL usar destino previamente cadastrado, verificado e vinculado ao tenant, NÃO SHALL aceitar URL arbitrária por pedido; SHALL criar uma entrega por evento final e destino/versão, com o corpo sendo a representação materializada pela Órbita conforme COM-05, sem produzir um segundo envelope; resultado grande SHALL ser acessível por referência estável no hub. A assinatura HMAC-SHA256 SHALL usar bytes exatos do corpo e timestamp, com key_id permitindo rotação. A homologação do cliente SHALL comprovar deduplicação durável por event_id e validação de timestamp dentro da janela acordada — proposta de cinco minutos; o hub NÃO SHALL presumir que o cliente implementa isso só por retornar 2xx. Reenvio SHALL manter event_id e identidade semântica, com assinatura e timestamp da tentativa novos. 2xx SHALL confirmar apenas recebimento HTTP, não processamento interno do cliente. Timeout, rede, 429 e 5xx SHALL permitir retry com jitter/Retry-After; 401/403 SHALL suspender para correção; 404/410 SHALL desabilitar o destino para revisão; outros 4xx SHALL exigir diagnóstico. Defaults propostos: conexão 2 s, total 10 s, janela automática 72 h, atrasos 1/5/15 min e 1/4/8 h, depois a cada 12 h dentro da janela. Esgotamento SHALL gerar EXHAUSTED e alerta; o resultado SHALL continuar disponível segundo retenção. Destino recuperado SHALL admitir reenvio autorizado, sem reexecução nem receita duplicada.

#### Scenario: Recibo é persistido antes da resposta 2xx ao provedor
- GIVEN um callback autenticado chega do provedor conforme capacidade homologada
- WHEN o Cometa processa esse callback
- THEN o sistema SHALL persistir o recibo antes de responder 2xx, e callback inválido NÃO SHALL alterar o protocolo

#### Scenario: Destino de webhook precisa estar previamente cadastrado, sem URL arbitrária por pedido
- GIVEN um evento final precisa ser entregue por webhook a um cliente
- WHEN o Pulsar seleciona o destino de entrega
- THEN o sistema SHALL usar apenas destino previamente cadastrado, verificado e vinculado ao tenant, NÃO SHALL aceitar uma URL arbitrária informada no pedido

#### Scenario: Erros de entrega são classificados por família de status HTTP
- GIVEN uma tentativa de entrega de webhook recebe timeout, 429, 401 ou 404
- WHEN o Pulsar classifica essa resposta
- THEN o sistema SHALL agendar retry com jitter/Retry-After para timeout/rede/429/5xx, suspender o destino para correção em 401/403, e desabilitá-lo para revisão em 404/410

#### Scenario: Esgotamento das tentativas gera EXHAUSTED e alerta, resultado permanece disponível
- GIVEN todas as tentativas de entrega dentro da janela automática de 72 horas se esgotam, seguindo os atrasos 1/5/15 minutos e 1/4/8 horas e depois a cada 12 horas
- WHEN o sistema encerra as tentativas dessa entrega
- THEN o sistema SHALL marcar a entrega como EXHAUSTED e emitir alerta, mantendo o resultado disponível segundo a política de retenção

### Requirement: EXE-09 — Falhas, retry e encerramento
O sistema SHALL permitir retry seguro quando a falha for comprovadamente anterior ao envio; após envio possível, o sistema SHALL classificar a operação como UNKNOWN e consultar status ou reenviar com a mesma chave apenas se a garantia idempotente do provedor for suficiente, abrindo reconciliação manual quando não houver mecanismo confiável — a disponibilidade de outro provedor NÃO SHALL eliminar o risco de duplicação. O sistema SHALL separar deadline de negócio, timeout por tentativa, janela de polling, retry de broker e entrega ao cliente; erro permanente NÃO SHALL receber retry infinito. DLQ SHALL ser tratada como fila de diagnóstico com responsável, aging e reprocessamento controlado, NÃO SHALL ser tratada como encerramento da obrigação. Um reconciliador SHALL verificar protocolos sem progresso, intenção sem publicação, callback órfão, polling vencido, resultado sem entrega e fato sem apuração. Resultado tardio após encerramento por SLA SHALL ser obrigatoriamente rejeitado para publicação nesse protocolo, conforme EXE-12; preservar evidência e conciliar NÃO SHALL autorizar reabrir o protocolo, revisá-lo para sucesso ou enviar novo webhook de sucesso. Correção legítima de outro tipo de resultado final MAY seguir revisão contratual auditada, mas NÃO SHALL contornar um encerramento por prazo; dados expurgados legitimamente NÃO SHALL ser ressuscitados.

#### Scenario: Falha anterior ao envio permite retry seguro
- GIVEN uma falha ocorre comprovadamente antes do envio da chamada ao provedor
- WHEN o sistema decide sobre repetir a tentativa
- THEN o sistema SHALL permitir retry seguro dessa tentativa

#### Scenario: Falha após envio possível exige UNKNOWN e consulta/reenvio idempotente ou reconciliação manual
- GIVEN uma falha ocorre depois que o envio ao provedor já era possível
- WHEN o sistema não possui garantia idempotente suficiente do provedor
- THEN o sistema SHALL classificar a operação como UNKNOWN e SHALL abrir reconciliação manual, NÃO SHALL reenviar sem essa garantia nem tratar a disponibilidade de outro provedor como eliminação do risco de duplicação

#### Scenario: Resultado tardio após encerramento por SLA não reabre o protocolo
- GIVEN um resultado chega depois do encerramento do protocolo por SLA
- WHEN o sistema avalia esse resultado tardio
- THEN o sistema SHALL rejeitar obrigatoriamente esse resultado para publicação nesse protocolo conforme EXE-12, NÃO SHALL reabrir passos, revisar para sucesso ou enviar novo webhook de sucesso

#### Scenario: DLQ é diagnóstico, não encerramento da obrigação
- GIVEN uma mensagem cai na DLQ após esgotar retries de broker
- WHEN o sistema trata essa mensagem
- THEN o sistema SHALL tratar a DLQ como fila de diagnóstico com responsável, aging e reprocessamento controlado, NÃO SHALL considerá-la encerramento da obrigação original

### Requirement: EXE-10 — TTL de reprocessamento em segundos
retry_ttl_seconds SHALL ser um inteiro não negativo obrigatório na política publicada do serviço/produto e na resolução do contrato do cliente; zero SHALL desabilitar retry de falha transitória, e ausência NÃO SHALL significar infinito. O valor efetivo SHALL ser resolvido antes do aceite, limitado pelos tetos do serviço/produto e pela capacidade homologada; o sistema SHALL persistir também o máximo de tentativas e o orçamento de retries por domínio de capacidade. TTL NÃO SHALL ser tratado como retenção de dados, timeout HTTP ou prazo de entrega do webhook. Ao primeiro erro transitório recuperável do passo, o sistema SHALL persistir first_transient_at uma única vez; a janela SHALL terminar no menor instante entre first_transient_at + retry_ttl_seconds, o deadline do passo e o deadline do cliente menos reserva de finalização. Reinício, backoff, duplicata, novo worker ou alternância de provedor NÃO SHALL redefinir esse início. O produto SHALL manter limite global e cada passo MAY ter teto menor; o sistema NÃO SHALL somar janelas de provedores para alongar o SLA do produto. Enquanto a janela estiver aberta, houver orçamento de tentativa e a chamada for segura, o sistema SHALL agendar retry durável com backoff/jitter respeitando controle adaptativo, e NÃO SHALL enviar erro final ao cliente por indisponibilidade transitória ainda recuperável — ASYNC já possui aceite 202 e GET informa pendência. Erro permanente, cancelamento válido ou impossibilidade contratual conhecida MAY concluir antes do TTL. Se next_run_at mais orçamento mínimo da tentativa/consolidação exceder o prazo restante, o sistema NÃO SHALL iniciar nova tentativa inviável. Falha comprovadamente sem efeito até esgotar TTL SHALL resultar em FAILED com motivo PROVIDER_UNAVAILABLE_RETRY_EXHAUSTED, se antes do deadline do cliente; se o deadline já venceu, o sistema SHALL concluir EXPIRED/SLA_EXCEEDED. Operação externa ainda incerta ao encerrar a janela SHALL terminar para o cliente como EXPIRED/RETRY_WINDOW_EXHAUSTED, mantendo reconciliação separada; esse motivo NÃO SHALL provar quebra do SLA temporal do cliente nem ausência de custo. Quando o provedor volta a aceitar a operação e ela segue pendente de conclusão normal, o fim do TTL de retry, isoladamente, NÃO SHALL ser causa de término, pois esse TTL limita tentativas de recuperação enquanto o SLA absoluto limita o processamento normal. A janela original SHALL permanecer registrada e NÃO SHALL ser renovada por recuperação momentânea; nova falha posterior NÃO SHALL ganhar outra janela automaticamente. Operação UNKNOWN SHALL continuar sujeita à regra de encerramento/reconciliação. Um timeout de chamada NÃO SHALL, por si só, dar segurança para repetir — o sistema SHALL aplicar EXE-09. Circuito aberto SHALL preservar a obrigação e MAY adiar tentativa, mas NÃO SHALL pausar o relógio.

#### Scenario: Exemplo numérico de retry_until com TTL 60s e reserva final 5s
- GIVEN um protocolo aceito em t=0 com client_sla_seconds=120, primeiro erro transitório em t=5 e retry_ttl_seconds=60 com reserva final de 5 s
- WHEN o sistema calcula o fim da janela de retry
- THEN o sistema SHALL definir retry_until em t=65 (first_transient_at=5 mais TTL=60), e SHALL finalizar a falha conhecida nesse instante se não houver êxito, sem esperar indefinidamente até t=120

#### Scenario: Zero desabilita retry de falha transitória
- GIVEN retry_ttl_seconds é configurado como zero para um serviço/produto
- WHEN ocorre um erro transitório recuperável nesse passo
- THEN o sistema SHALL desabilitar o retry de falha transitória para esse passo, sem tratar a ausência de configuração em outro caso como infinito

#### Scenario: Reinício ou troca de provedor não redefine first_transient_at
- GIVEN first_transient_at já foi persistido para um passo após o primeiro erro transitório
- WHEN ocorre reinício, backoff, duplicata, novo worker ou alternância de provedor
- THEN o sistema NÃO SHALL redefinir first_transient_at por causa desses eventos, mantendo a janela original de retry

#### Scenario: Falha sem efeito até esgotar TTL antes do deadline resulta em FAILED com motivo específico
- GIVEN uma falha permanece comprovadamente sem efeito externo até o TTL se esgotar, e o deadline do cliente ainda não venceu
- WHEN o sistema conclui o passo
- THEN o sistema SHALL resultar em FAILED com motivo PROVIDER_UNAVAILABLE_RETRY_EXHAUSTED; se o deadline já tiver vencido, o sistema SHALL concluir EXPIRED/SLA_EXCEEDED

#### Scenario: Fim do TTL não é, sozinho, causa de término quando o provedor volta a aceitar
- GIVEN o provedor volta a aceitar a operação e ela segue pendente de conclusão normal após o fim da janela de retry de recuperação
- WHEN o sistema avalia se deve encerrar o passo
- THEN o sistema NÃO SHALL tratar isoladamente o fim do TTL de retry como causa de término, pois esse TTL limita tentativas de recuperação e não o SLA absoluto de processamento normal

### Requirement: EXE-11 — Deadline, encerramento e disputa temporal
client_sla_seconds SHALL ser o máximo contratual entre aceite durável e disponibilização da representação final no hub. O limite contratual máximo SHALL ser accepted_at + client_sla_seconds; client_deadline_at SHALL ser esse limite em ASYNC/AUTO e o menor limite compatível com a conexão em SYNC, conforme EXE-14. A redução SHALL ser publicada na política de modalidade e registrada no aceite, e NÃO SHALL ser ampliada em voo; SHALL incluir fila, reserva financeira, provedores, captura de objetos, composição e transformação personalizada. O SLA de tentativa de webhook SHALL ser outro relógio; chegar ao deadline NÃO SHALL dispensar avisar o cliente do encerramento. provider_sla_seconds SHALL definir obrigação de conclusão do provedor, com marco inicial explicitamente contratado (envio do hub ou aceite externo duravelmente observado); métrica de resposta ao submit e prazo final externo SHALL ser separados. Timestamp informado pelo provedor SHALL ser preservado, mas NÃO SHALL demonstrar sozinho quando o hub recebeu e conservou o resultado; se o contrato usa conclusão declarada pelo provedor, o sistema SHALL aferir adicionalmente o atraso de transmissão, sem estender o deadline do cliente. Um temporizador durável da Órbita SHALL encerrar protocolo aberto ao atingir o deadline, mesmo com polling saudável, HTTP 200 do provedor e ausência de erro técnico. Final e timeout SHALL disputar uma única transição terminal serializável; a elegibilidade SHALL exigir representação final validada e confirmada de forma durável antes do limite, e igualdade ao deadline SHALL ser tratada como atraso. O sistema NÃO SHALL decidir pelo instante em que o worker do timeout conseguiu executar, e NÃO SHALL usar relógio local divergente, timestamp de enqueue ou hora alegada pelo provedor para aceitar sucesso tardio. A implementação futura SHALL demonstrar o ponto autoritativo de ordenação/commit e a arbitragem na fronteira do prazo, inclusive commit lento; iniciar uma transação antes do deadline NÃO SHALL provar que o resultado estava durável no prazo. O evento de sucesso NÃO SHALL sair antes dessa verificação. O ensaio SHALL comprovar que não há publicação concorrente de sucesso e expiração; se a infraestrutura não conseguir confirmar a elegibilidade temporal, o sistema NÃO SHALL emitir sucesso não comprovado, registrando a indisponibilidade para recuperação/reconciliação. Na expiração, o sistema SHALL persistir EXPIRED, SLA_EXCEEDED, prazo e causa; SHALL materializar representação final de erro do cliente; SHALL publicar a obrigação de notificação e fatos de SLA; SHALL cancelar timers produtivos e impedir novos passos; reconciliar chamadas já enviadas e reservas/custos SHALL ocorrer em fluxo separado. Scheduler atrasado SHALL ser tratado como falha do hub mensurada; GET NÃO SHALL executar provedor nem apresentar como sucesso um resultado fora do prazo. Enforcement de SLA de provedor SHALL ser política distinta: MONITOR_ONLY SHALL registrar violação podendo ainda concluir no prazo do cliente; REJECT_LATE SHALL fechar o passo para uso no produto ao vencer o prazo do provedor. Failover SHALL seguir segurança de efeito e prazo restante. Em ambos os casos, o deadline do cliente SHALL ser rígido e não renovável; a comparação dessas políticas SHALL ser obrigatória na publicação para não prometer obrigação contraditória.

#### Scenario: Temporizador encerra o protocolo no deadline mesmo com polling saudável
- GIVEN um protocolo aberto atinge client_deadline_at enquanto o polling está saudável e o provedor retornou HTTP 200 sem erro técnico
- WHEN o temporizador durável da Órbita verifica o prazo
- THEN o sistema SHALL encerrar o protocolo aberto nesse instante, mesmo nessas condições favoráveis aparentes

#### Scenario: Igualdade exata ao deadline é tratada como atraso
- GIVEN a representação final se torna disponível exatamente no instante do client_deadline_at
- WHEN o sistema avalia a elegibilidade temporal para sucesso
- THEN o sistema SHALL tratar essa igualdade como atraso, exigindo confirmação durável estritamente antes do limite para elegibilidade de sucesso

#### Scenario: Infraestrutura incapaz de confirmar elegibilidade temporal não emite sucesso
- GIVEN a infraestrutura não consegue confirmar se a representação final foi validada e confirmada de forma durável antes do deadline
- WHEN o sistema decide a transição terminal
- THEN o sistema NÃO SHALL emitir sucesso não comprovado, registrando a indisponibilidade para recuperação/reconciliação em vez de aceitar sucesso tardio

#### Scenario: MONITOR_ONLY e REJECT_LATE aplicam políticas distintas sem afetar o deadline do cliente
- GIVEN o prazo do provedor (provider_sla_seconds) vence antes da conclusão do passo
- WHEN a política de enforcement configurada é MONITOR_ONLY ou REJECT_LATE
- THEN o sistema SHALL registrar a violação e permitir conclusão dentro do prazo do cliente sob MONITOR_ONLY, ou SHALL fechar o passo para uso no produto sob REJECT_LATE; em ambos os casos o deadline do cliente SHALL permanecer rígido e não renovável

### Requirement: EXE-12 — Rejeição de retorno tardio sem apagar evidência
Ao receber final depois de encerramento por SLA, o sistema SHALL persistir recibo autenticado como REJECTED_LATE_SLA, com correlação, horários, hash e evidência permitida; o sistema NÃO SHALL substituir o erro final do protocolo, NÃO SHALL reenviar sucesso, NÃO SHALL reabrir passos e NÃO SHALL capturar receita por sucesso. O retorno MAY comprovar efeito externo ou custo e alimentar contestação contratual. "Recusar" SHALL ser decisão de negócio, não necessariamente devolução de HTTP 4xx ao provedor; se o webhook foi autenticado e seu recibo duravelmente registrado, o transporte MAY responder 2xx para evitar tempestade de retransmissões, mas o recibo SHALL continuar rejeitado para conclusão. Responder código específico de rejeição SHALL ocorrer somente quando o contrato externo o suportar e seu efeito estiver homologado; falha de persistência NÃO SHALL receber ack de sucesso. Callback e polling que cruzam o deadline SHALL ser arbitrados pela mesma regra; mesmo resultado já iniciado no provedor antes do prazo MAY chegar tarde ao hub. Resultado recebido no Cometa antes do limite, mas consolidado/publicável no hub depois, MAY configurar quebra do hub e não necessariamente do provedor; processamento externo NÃO SHALL desaparecer porque o protocolo acabou. Reexecução solicitada depois SHALL criar novo protocolo, nova chave e avaliação econômica, sem reutilizar conteúdo rejeitado como sucesso do protocolo vencido. GET posterior e todas as tentativas do webhook de encerramento SHALL devolver a representação final de erro original. Coleta residual para reconciliação SHALL ter teto, finalidade e permissão distintos, NÃO SHALL se mascarar como polling produtivo após expiração.

#### Scenario: Retorno tardio é persistido como REJECTED_LATE_SLA sem substituir o erro final
- GIVEN um final chega ao Cometa depois do encerramento do protocolo por SLA
- WHEN o sistema processa esse retorno tardio
- THEN o sistema SHALL persistir o recibo autenticado como REJECTED_LATE_SLA com correlação, horários, hash e evidência, NÃO SHALL substituir o erro final do protocolo, reenviar sucesso ou reabrir passos

#### Scenario: Recusa pode responder 2xx ao provedor sem significar aceitação
- GIVEN um webhook de retorno tardio foi autenticado e seu recibo duravelmente registrado
- WHEN o hub decide o código de transporte a devolver ao provedor
- THEN o sistema MAY responder 2xx para evitar tempestade de retransmissões, mas o recibo SHALL continuar rejeitado para conclusão do protocolo

#### Scenario: Reexecução solicitada cria novo protocolo, não reutiliza conteúdo rejeitado
- GIVEN um cliente solicita reexecução depois de um retorno tardio rejeitado
- WHEN o sistema processa essa solicitação
- THEN o sistema SHALL criar novo protocolo, nova chave de idempotência e nova avaliação econômica, NÃO SHALL reutilizar o conteúdo rejeitado como sucesso do protocolo vencido

#### Scenario: GET e webhook de encerramento continuam devolvendo a representação de erro original
- GIVEN um protocolo foi encerrado por SLA e um retorno tardio foi rejeitado posteriormente
- WHEN o cliente consulta via GET ou recebe as tentativas do webhook de encerramento
- THEN o sistema SHALL devolver, em todas essas interações, a representação final de erro original do protocolo vencido

### Requirement: EXE-13 — Paralelismo e espera eficiente
A Órbita SHALL agendar em paralelo apenas passos READY sem dependências pendentes, dentro de limites de protocolo, produto, tenant e célula; dependências causais SHALL permanecer sequenciais. A consolidação do produto SHALL ocorrer depois da política de agregação, sem segurar thread ou transação enquanto aguarda todos os provedores; paralelismo SHALL ser avaliado pelo caminho crítico, não pela simples quantidade de goroutines. Cometa e Pulsar SHALL usar pools limitados, I/O de rede concorrente e contexto de cancelamento. Esperas longas de polling/retry SHALL ficar em timers persistidos, NÃO SHALL ficar em thread dedicada, loop ocupado ou goroutine dormindo por toda a duração do protocolo. Transformações/arquivos intensivos em CPU SHALL ser executados em pools isolados com orçamento, sem monopolizar o processamento de outras ofertas. Uma chamada pendente NÃO SHALL manter transação PostgreSQL aberta. Limites de memória, fila, conexões e tarefas em voo SHALL ser respeitados antes de reclamar mais trabalho. Desligamento SHALL drenar tarefas ou devolver posse de forma segura; cancelar espera local NÃO SHALL significar que o provedor cancelou sua operação. Bibliotecas que bloqueiam threads SHALL ser identificadas por profiling e isoladas/substituídas conforme evidência.

#### Scenario: Passos READY sem dependências pendentes são agendados em paralelo dentro de limites
- GIVEN múltiplos passos de um protocolo estão em READY sem dependências pendentes entre si
- WHEN a Órbita agenda a execução
- THEN o sistema SHALL agendá-los em paralelo, respeitando os limites de protocolo, produto, tenant e célula, mantendo sequenciais apenas as dependências causais

#### Scenario: Espera de polling/retry usa timers persistidos, não thread dedicada
- GIVEN uma operação aguarda o resultado de um polling ou retry de longa duração
- WHEN o sistema mantém essa espera
- THEN o sistema SHALL usar timers persistidos para essa espera, NÃO SHALL manter thread dedicada, loop ocupado ou goroutine dormindo por toda a duração do protocolo

#### Scenario: Cancelar espera local não prova que o provedor cancelou a operação
- GIVEN o desligamento do worker cancela a espera local de uma chamada em andamento
- WHEN o sistema trata esse cancelamento
- THEN o sistema NÃO SHALL presumir que o provedor cancelou sua operação apenas porque a espera local foi cancelada

### Requirement: EXE-14 — SYNC direto, prazo e resposta final
SYNC SHALL ser tratado como contrato de interação, não como nome alternativo para polling do cliente. Pré-condições SHALL incluir todas as etapas necessárias síncronas no provedor, o caminho crítico cabendo no orçamento do cliente e capacidade reservada para chamada e commits. A admissão SHALL registrar modo e limites; a execução SHALL observar o mesmo grafo e as mesmas regras de segurança do ASYNC. Passos independentes MAY chamar Cometa em paralelo, limitados por protocolo/tenant/provedor, e sucessores SHALL aguardar apenas suas dependências. A sequência normativa para serviço simples SHALL ser: validar cliente/contrato/entrada e capacidade; confirmar protocolo/intenção no core; reservar financeiro se estrito; enviar comando DIRECT idempotente ao Cometa; Cometa confirma operação/tentativa e resolve segredo autorizado; chama provedor; confirma observação e intenção de fatos; retorna o final externo à Órbita; Órbita captura/consolida o que faltar, valida a projeção, decide o prazo e confirma seu final; responde pela conexão original. Publicação de fatos, cálculo pós-pago e webhook NÃO SHALL ser esperas obrigatórias desse percurso. Registro prévio de tentativa NÃO SHALL significar que ela foi enviada — seu estado SHALL distinguir preparação, envio possível e observação. connection_deadline_at SHALL ser received_at na borda confiável + sync_wait_seconds homologado, e o pedido MAY reduzir mas NÃO SHALL ampliar o limite autorizado; client_deadline_at SYNC SHALL ser o menor entre accepted_at + client_sla_seconds e connection_deadline_at − margem de retorno HTTP; o deadline do provedor/passo NÃO SHALL ultrapassar o deadline do protocolo menos reserva de consolidação, incluindo a segurança da próxima tentativa; o retry transitório SHALL ser limitado por first_transient_at + retry_ttl_seconds e pelo prazo efetivo do passo/protocolo, e o TTL NÃO SHALL prolongar a conexão. A política publicada SHALL explicitar essa diferença de relógios. Se o aceite já consumir o orçamento útil, o sistema SHALL recusar antes de efeito externo; se já houve commit, o sistema SHALL conservar o UUID e concluir falha/expiração de forma durável quando possível. Cliente que precisa de janela de recuperação maior que sua conexão SHALL escolher ASYNC ou AUTO. SYNC NÃO SHALL retornar 202 ao encontrar fila, provedor lento ou banco em failover; nessas situações MAY ocorrer erro contratual/transporte com protocolo, sem esconder a quebra de disponibilidade. Enquanto aguarda I/O, a chamada MAY manter goroutine/contexto limitado, sem transação aberta nem thread dedicada por espera de rede suportada. Retries diferidos SHALL manter timer/intenção persistidos; o observador HTTP NÃO SHALL se tornar a autoridade do timer. Fechamento da conexão SHALL interromper a espera local e propagar cancelamento quando seguro, mas NÃO SHALL provar cancelamento do provedor. Sem final conhecido após possível efeito, o sistema SHALL manter operação UNKNOWN/reconciliação; ao prazo, EXPIRED SHALL continuar impedindo sucesso tardio. O mesmo corpo final persistido SHALL ser retornado na criação SYNC, no GET posterior e no webhook contratado, preservando COM-05 e a versão do cliente; headers/status HTTP MAY diferenciar criação, consulta e timeout. O compromisso normal SHALL ser resposta final na mesma chamada; não existe garantia física de entrega se o cliente fechar a conexão ou a rede falhar.

#### Scenario: Pedido pode reduzir mas não ampliar connection_deadline_at
- GIVEN connection_deadline_at é calculado como received_at na borda confiável mais sync_wait_seconds homologado
- WHEN um pedido específico solicita um prazo diferente
- THEN o sistema MAY aceitar uma redução desse limite pelo pedido, NÃO SHALL permitir ampliação além do limite autorizado

#### Scenario: SYNC não retorna 202 diante de fila, provedor lento ou banco em failover
- GIVEN uma execução SYNC encontra fila, provedor lento ou banco em failover
- WHEN o sistema decide como responder
- THEN o sistema NÃO SHALL retornar 202 nessas situações; MAY ocorrer erro contratual/transporte com protocolo, sem esconder a quebra de disponibilidade

#### Scenario: Aceite que já consome o orçamento útil é recusado antes de efeito externo
- GIVEN o orçamento útil da conexão já foi consumido antes de qualquer efeito externo ser enviado
- WHEN o sistema avalia se deve prosseguir com a chamada SYNC
- THEN o sistema SHALL recusar a execução antes de gerar efeito externo; se já houve commit, o sistema SHALL conservar o UUID e concluir falha/expiração de forma durável quando possível

#### Scenario: Fechamento da conexão interrompe espera local sem provar cancelamento do provedor
- GIVEN o cliente fecha a conexão durante uma execução SYNC em andamento
- WHEN o sistema trata esse fechamento
- THEN o sistema SHALL interromper a espera local e propagar cancelamento quando seguro, mas NÃO SHALL tratar isso como prova de que o provedor cancelou sua operação, mantendo UNKNOWN/reconciliação até final conhecido ou EXPIRED ao prazo

#### Scenario: Mesmo corpo final é retornado na criação SYNC, no GET e no webhook
- GIVEN um protocolo SYNC conclui com um corpo final persistido
- WHEN esse resultado é acessado na resposta de criação, no GET posterior e no webhook contratado
- THEN o sistema SHALL retornar o mesmo corpo final persistido nos três canais, preservando COM-05 e a versão do cliente, podendo diferenciar apenas headers/status HTTP entre criação, consulta e timeout

### Requirement: EXE-15 — Posse única entre despacho direto, fila e recuperação
Na admissão, o sistema SHALL persistir command_id e dispatch_mode DIRECT para SYNC ou QUEUED para ASYNC/AUTO. O despachante de outbox SHALL publicar somente comandos elegíveis QUEUED. Intenção DIRECT SHALL ter observação de progresso e watchdog durável, mas NÃO SHALL ser simultaneamente duplicada como comando de execução em SQS. Fatos de domínio MAY ser publicados em qualquer modalidade, sendo diferentes do comando que causa efeito. Cometa SHALL reclamar a intenção com chave única, lease/epoch e estado de operação. Repetição HTTPS ou mensagem de recuperação SHALL devolver/acompanhar a operação existente; um worker vencido NÃO SHALL consolidar estado local. Antes de reassumir, o sistema SHALL distinguir comando nunca enviado, resultado local já registrado e envio possível com resposta desconhecida. Para UNKNOWN, o sistema SHALL consultar status seguro ou repetir somente sob idempotência externa homologada; lease expirado, isoladamente, NÃO SHALL tornar seguro um segundo submit. Transferir trabalho remanescente para fila SHALL exigir transição durável exclusiva, liberação/fencing da posse anterior e conservação de command_id/operation_id; se há chamada externa incerta, o novo trabalho SHALL ser de reconciliação, não novo efeito. Essa transferência NÃO SHALL mudar o modo escolhido pelo cliente nem estender o deadline. Em SYNC com conexão perdida, a obrigação durável SHALL continuar até finalização/encerramento contratual; o sistema NÃO SHALL prometer retornar resposta a uma conexão já encerrada. Crash depois do final Cometa e antes da resposta direta SHALL ser resolvido por leitura da operação existente; o evento posterior MAY aplicar o mesmo fato, e a consolidação Órbita SHALL deduplicar ambos por operação/versão. Crash após o final Órbita e antes da resposta pública SHALL ser resolvido pela mesma chave/protocolo. O watchdog, a fila e a resposta direta SHALL compartilhar pré-condições e chaves econômicas; NÃO SHALL haver implementação negocial divergente por modalidade.

#### Scenario: Intenção DIRECT não é duplicada como comando de execução em SQS
- GIVEN uma intenção DIRECT foi publicada com watchdog durável e observação de progresso
- WHEN o despachante de outbox avalia o que publicar
- THEN o sistema SHALL publicar apenas comandos elegíveis QUEUED, NÃO SHALL tratar essa intenção DIRECT como um comando de execução também enviado a SQS

#### Scenario: Lease expirado sozinho não torna seguro um segundo submit
- GIVEN a lease de uma operação em UNKNOWN expira sem confirmação de resposta
- WHEN um worker considera repetir o submit
- THEN o sistema NÃO SHALL tratar o lease expirado, isoladamente, como suficiente para autorizar um segundo submit; SHALL exigir status seguro consultado ou idempotência externa homologada

#### Scenario: Transferência para fila conserva command_id/operation_id sem mudar modo ou estender deadline
- GIVEN um trabalho remanescente precisa ser transferido do despacho direto para a fila
- WHEN essa transição ocorre
- THEN o sistema SHALL exigir transição durável exclusiva com liberação/fencing da posse anterior, conservando command_id e operation_id, NÃO SHALL mudar o modo escolhido pelo cliente nem estender o deadline

#### Scenario: Crash após final do Cometa e antes da resposta direta é resolvido por leitura da operação existente
- GIVEN o Cometa já registrou o final da operação mas o processo caiu antes de responder diretamente à Órbita
- WHEN o sistema recupera esse estado
- THEN o sistema SHALL resolver a situação por leitura da operação já existente, e a consolidação da Órbita SHALL deduplicar esse fato por operação/versão caso o evento correspondente também chegue posteriormente

#### Scenario: Em SYNC com conexão perdida, a obrigação durável continua sem prometer resposta à conexão encerrada
- GIVEN a conexão do cliente se perde durante uma execução SYNC ainda em andamento
- WHEN o sistema continua processando essa obrigação
- THEN o sistema SHALL manter a obrigação durável até finalização/encerramento contratual, NÃO SHALL prometer retornar uma resposta à conexão já encerrada

### Requirement: EXE-16 — UUIDv7, repetição e consulta futura
protocol_id SHALL usar UUIDv7 conforme RFC 9562 em todas as requisições duravelmente aceitas, SYNC, ASYNC ou AUTO, inclusive protocolos que terminem em falha/expiração; o sistema NÃO SHALL substituir por UUIDv4/ULID em um canal. A unicidade SHALL ser protegida pela persistência; o componente temporal do UUIDv7 NÃO SHALL ser tratado como identidade de tenant, prova de autorização ou prova de commit/deadline. Entropia, gerador concorrente e comportamento sob regressão de relógio SHALL ser qualificados. A resposta de criação SHALL informar o UUIDv7 e o endereço de consulta em headers documentados e no corpo canônico; perfis legados SHALL preservar a correlação por mapeamento homologado conforme COM-05. Repetição da mesma Idempotency-Key pelo mesmo tenant/operação SHALL recuperar o mesmo UUID, mesmo se o cliente não recebeu a primeira resposta; a janela de deduplicação SHALL ser contratada e não menor que o período operacional relevante, e após expurgo legítimo da chave o sistema NÃO SHALL prometer identificar a repetição antiga. GET SHALL validar o tenant e ler exclusivamente o ecossistema do Hub; para final existente e autorizado, SHALL devolver a representação contratada conservada; para indisponibilidade de fonte suficiente, SHALL aplicar DAD-10. Nenhuma credencial de cliente SHALL consultar outro tenant por conhecer o UUID. O perfil administrativo individual SHALL usar rota própria e auditoria. Pedido inválido ou recusado antes do commit MAY receber request_id técnico, mas esse identificador NÃO SHALL ser anunciado como protocolo aceito e consultável.

#### Scenario: Repetição da mesma Idempotency-Key recupera o mesmo UUID mesmo sem resposta anterior recebida
- GIVEN o mesmo tenant repete a mesma Idempotency-Key para a mesma operação, sem ter recebido a primeira resposta
- WHEN a Órbita processa essa repetição dentro da janela de deduplicação contratada
- THEN o sistema SHALL recuperar o mesmo UUIDv7 do protocolo original, sem gerar um novo protocolo

#### Scenario: Componente temporal do UUIDv7 não prova autorização nem commit
- GIVEN um protocol_id UUIDv7 contém um componente temporal decodificável
- WHEN esse componente é usado em alguma verificação
- THEN o sistema NÃO SHALL tratar esse componente temporal como identidade de tenant, prova de autorização ou prova de commit/deadline

#### Scenario: Nenhuma credencial de cliente consulta outro tenant por conhecer o UUID
- GIVEN uma credencial de cliente conhece o UUIDv7 de um protocolo pertencente a outro tenant
- WHEN essa credencial tenta consultar esse protocolo via GET
- THEN o sistema SHALL validar o tenant e retornar 404, NÃO SHALL permitir que essa credencial consulte outro tenant apenas por conhecer o UUID

#### Scenario: Request_id técnico de pedido recusado antes do commit não é anunciado como protocolo aceito
- GIVEN um pedido é recusado antes do commit e recebe apenas um request_id técnico
- WHEN esse identificador é exposto ao cliente
- THEN o sistema NÃO SHALL anunciar esse request_id como um protocolo aceito e consultável

## Notas de origem

Este delta deriva integralmente do capítulo `docs/04_EXECUCAO_E_INTEGRACOES.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos EXE-01 a EXE-16 conforme texto normativo de origem.
