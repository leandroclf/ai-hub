# 04 · Execução, polling e entrega

## EXE-01 · Admissão durável — proprietário: Órbita

Pré-condições: consumidor autenticado/autorizado, oferta e contrato elegíveis, payload válido, objetos disponíveis, capacidade de admissão e versão de configuração válida. Requisição aceita recebe protocol_id UUIDv7. Entrada recusada gera trilha técnica segura, mas não obrigação de execução ou receita, salvo evento comercial explicitamente contratado e distinto de consumo do serviço.

Idempotency-Key é obrigatória na criação. Unicidade por tenant, operação pública e chave; associar hash semântico de oferta/versão, contrato técnico de entrada/transformação fixado, input canônico e objetos imutáveis. Mesma chave e mesmo hash retornam protocolo e estado existentes; mesma chave com payload diferente retorna conflito. Inclusão de timestamps de trace ou URLs temporárias no hash é proibida. Repetição após timeout do cliente usa a mesma chave.

Confirmar aceite em qualquer modalidade somente após commit; em ASYNC, responder 202 apenas nessa condição. Se o commit ocorreu e a resposta HTTP se perdeu, a repetição recupera o mesmo protocolo. Sem banco de admissão, responder indisponibilidade; não confirmar aceitação somente porque a mensagem entrou em memória. UUID não é senha e não substitui autorização.

## EXE-02 · Modos de atendimento — proprietário: Órbita / Produto

| Provedor | Modo escolhido pelo cliente | Comportamento obrigatório |
| --- | --- | --- |
| Síncrono | SYNC | Despacho interno direto, final durável na mesma requisição quando concluído no orçamento; sem espera obrigatória por fila e sem conversão para 202 |
| Síncrono | ASYNC | 202 após aceite; worker chama o provedor síncrono; resultado local e webhook conforme contrato |
| Síncrono | AUTO | Espera contratada pelo processamento em fila; 200 final se disponível ou 202 com o mesmo protocolo; não equivale a SYNC |
| Assíncrono | ASYNC | 202; obtenção do final por callback, polling ou combinação; resultado local e notificação |
| Assíncrono | AUTO | Espera limitada pelo resultado durável, com final ou 202; não mantém conexão por todo o prazo do provedor |
| Assíncrono | SYNC | Não ofertado na referência v4; exceção requer nova decisão e qualificação do orçamento, não simples flag de configuração |

Toda admissão recebe UUIDv7. No perfil canônico, conclusão válida retorna 200 com estado final e corpo persistido; falha de negócio também está representada no contrato. Prazo SYNC esgotado retorna 504 com o final de expiração se ele já foi persistido; falha de infraestrutura pode impedir a finalização e deve ser indicada como incerteza, com o protocolo conhecido. Não anunciar como final um corpo que só exista em memória. Perda da conexão pode impedir qualquer resposta HTTP; consulta/retry com a mesma chave recupera o protocolo, sem repetir efeito inseguro.

Em AUTO, prazo da conexão pode terminar antes do prazo de negócio, produzindo 202. Em SYNC, o prazo efetivo do protocolo deve caber na conexão, conforme EXE-14. O limite de oito segundos da v3 deixa de ser default universal: a publicação exige os valores do cliente, da borda, do serviço e da reserva final. Encerrar conexão não cancela efeito externo já enviado. Webhook adicional é opção explícita, usando o mesmo resultado e sem receita duplicada.

## EXE-03 · Estados e autoridade — proprietário: Órbita / Cometa

| Dimensão | Estados normativos e significado |
| --- | --- |
| Protocolo | ACCEPTED, RUNNING, WAITING_PROVIDER, RECONCILING; finais SUCCEEDED, PARTIALLY_SUCCEEDED, FAILED, EXPIRED, CANCELLED |
| Passo | PENDING, READY, RUNNING, WAITING_PROVIDER, RECONCILING; finais SUCCEEDED, FAILED, SKIPPED, CANCELLED |
| Operação externa | PREPARED, SUBMITTING, ACCEPTED_EXTERNAL, WAITING_FINAL, UNKNOWN; finais SUCCEEDED, FAILED, CANCELLED |
| Recibo externo | RECEIVED, UNCORRELATED, APPLIED, DUPLICATE, CONFLICT, REJECTED, REJECTED_LATE_SLA |
| Entrega | PENDING, DELIVERING, RETRY_SCHEDULED, DELIVERED, SUSPENDED, EXHAUSTED |
| Resultado | Versões finais imutáveis com motivo e vínculo da revisão; ponteiro atual controlado pela Órbita |

ACCEPTED→RUNNING após despacho direto ou agendamento; RUNNING→WAITING_PROVIDER quando existe operação assíncrona; incerteza vai para RECONCILING. Cometa decide apenas o fato externo; Órbita decide a conclusão do passo/produto. Sucesso exige saída válida e disponível. Parcialidade exige política publicada. Eventos intermediários atrasados não regridem estado final.

FAILED exige falha conhecida segundo contrato; UNKNOWN não vira FAILED por conveniência. EXPIRED encerra o prazo de negócio, podendo manter operação incerta em reconciliação e obrigação financeira separada. CANCELLED exige inexistência comprovada de efeitos pendentes ou confirmação externa apropriada; pedido de cancelamento fica como solicitação até então. Reconciliar pode revelar custo mesmo depois de expiração.

## EXE-04 · Operação e correlação — proprietário: Cometa

Antes de enviar, registrar operation_id, protocol_id, step_id, provedor, conta, versão do vínculo, modo e binding_id de credencial, secret_version_id efetivamente usado, contrato de aquisição e chave de idempotência externa quando suportada. Cada chamada física cria attempt_id e tipo de interação. Reenvio de transporte preserva operation_id e chave externa; troca autorizada de provedor cria outra operação vinculada ao mesmo passo.

provider_request_id é associado no escopo documentado do provedor/conta/ambiente; nunca presumir unicidade global. Quando possível, enviar token de correlação opaco previamente persistido. Callback sem referência externa ainda associada é persistido como UNCORRELATED, com reconciliação posterior; não associar por nome, CPF ou proximidade de horário.

Registrar timestamps de preparação/envio/recebimento, resultado de autenticação, erro canônico, hash/evidência e classificação de efeito conhecido ou incerto. Se a chamada saiu mas o processo caiu antes de persistir resposta, reconstruir a situação por chave idempotente, status externo ou investigação. Não inventar uma resposta final.

## EXE-05 · Polling configurável — proprietário: Cometa / Integrações

“Polling” significa consulta periódica ao provedor; não confundir com pool de conexões. Modos por vínculo: CALLBACK_ONLY, POLLING_ONLY e CALLBACK_WITH_POLLING_FALLBACK. Ativação no portal só é permitida se o adaptador declarar capacidade de consulta homologada e se o contrato permitir seu consumo.

| Parâmetro obrigatório | Regra de configuração |
| --- | --- |
| Referência da consulta | Origem do provider_request_id e mapeamento no endpoint; autenticação por referência de segredo |
| Estados externos | Mapa completo para pendente/sucesso/falha/desconhecido; valores novos vão para diagnóstico |
| Obtenção do resultado | Resultado junto do status ou operação FETCH separada; limites e eventual custo de ambas |
| Agendamento | Atraso inicial, intervalo mínimo/máximo, fator de backoff, jitter e fallback após ausência de callback |
| Limites | Deadline, máximo de tentativas, consultas por segundo e concorrência por conta/serviço |
| Erros | 429/Retry-After, timeout, 401/403, inexistente, payload inválido e indisponibilidade |
| Economia | Medidor e tarifa por consulta/fetch, franquia e teto de custo quando aplicável |
| Operação | Pausa, retomada, aging, versão e autorização para mudar política de operações em voo |

Defaults propostos para qualificação, nunca aplicados sem aprovação do vínculo e compatibilidade com o deadline: atraso inicial 2 s; intervalo 5 s; backoff fator 2 até 60 s; jitter ±20%; callback com fallback após 30 s. Deadline e número máximo dependem do serviço e são obrigatórios, sem default universal. Respeitar Retry-After dentro do prazo; se ultrapassa deadline, tratar expiração/reconciliação.

Agenda durável guarda next_run_at, contadores, lease, token de posse e versão. Um worker reclama trabalho de forma exclusiva, não mantém thread ou transação dormindo por toda operação. Após reinício retoma vencidos com controle de avalanche. Final persistido ou deadline atingido encerra agendamentos produtivos futuros; consultas residuais de reconciliação têm finalidade e orçamento separados. Pausar polling não cancela operação nem apaga prazo; retomada considera tempo transcorrido.

## EXE-06 · Concorrência entre callback e polling — proprietário: Cometa

Ambos submetem observações ao mesmo consolidado da operação externa. Se callback e polling apresentam o mesmo final, aplicar uma única transição e tratar o segundo como evidência duplicada. Uma consulta já em voo pode terminar após cancelamento da agenda; sua tentativa e possível custo continuam registráveis, sem republicar conclusão equivalente.

Se houver finais conflitantes, preservar ambos e marcar CONFLICT para resolução com a regra homologada do provedor. Não prevalece automaticamente o último recebido. Se o provedor possui sequência/revisão autoritativa, ela participa da resolução; sem isso, exige reconciliação. Resposta “não encontrado” logo após submit pode ser consistência eventual: usar janela documentada, não reiniciar a operação às cegas.

## EXE-07 · Contrato de consultas públicas — proprietário: Órbita

| Operação lógica | Comportamento definido |
| --- | --- |
| Criar pedido | SYNC direto com final na mesma conexão conforme EXE-14; ASYNC 202; AUTO final ou 202; todo aceite identifica UUIDv7 e consulta |
| Consultar protocolo | 200 com a representação contratada de estado/resultado, igual ao corpo do webhook quando final; 404 para inexistente ou outro tenant |
| Consultar resultado pendente | 200 com a variante pendente do contrato unificado; o endpoint de resultado, se mantido, é alias |
| Consultar resultado final | 200 com a mesma representação final persistida usada pelo webhook; inclui falha/EXPIRED no contrato do cliente |
| Consultar resultado expurgado | 410 RESULT_EXPIRED se metadado autorizado permite reconhecê-lo; sem nova busca externa |
| Solicitar cancelamento | Registra intenção durável; retorno não afirma cancelamento externo concluído |
| Consultar entregas | Estado e tentativas em projeção, informando instante de atualização |
| Atualizar/reexecutar | Nova criação com novo protocolo e referência ao anterior; avaliação comercial explícita |

Resultado contém protocol_id, client_request_id, status, result_version, schema_version, submitted_at, completed_at, partes do produto, códigos canônicos, conteúdo pequeno ou object_ids e informações autorizadas de download. Estado HTTP de consulta bem-sucedida não confunde erro de negócio com falha da infraestrutura HTTP.

Nenhuma dessas consultas públicas incrementa o medidor de execução do provedor. Cobrança pelo acesso a resultado armazenado só existe como medidor separado e explicitamente vendido; não integra o plano padrão v4. Cache miss, resultado expurgado e provedor indisponível não mudam essa regra.

## EXE-08 · Webhooks — proprietário: Cometa / Pulsar

Recepção de provedor: autenticar conforme capacidade homologada (HMAC, mTLS ou token dedicado), limitar tamanho/taxa, persistir recibo e só depois responder 2xx. IP permitido é proteção complementar, não substituto de autenticação quando esta é possível. Callback inválido não altera protocolo.

Envio ao cliente: destino previamente cadastrado, verificado e vinculado ao tenant; não aceitar URL arbitrária por pedido. Criar uma entrega por evento final e destino/versão. O corpo é a representação materializada pela Órbita conforme COM-05, com a correlação exigida no contrato; não é produzido um segundo envelope. Resultado grande é acessível por referência estável no hub. Assinatura HMAC-SHA256 usa bytes exatos do corpo e timestamp; key_id permite rotação. A homologação do cliente deve comprovar deduplicação durável por event_id e validação de timestamp dentro da janela acordada (proposta de cinco minutos); o hub não presume que o cliente implemente isso só por retornar 2xx. Reenvio mantém event_id e identidade semântica, com assinatura e timestamp da tentativa novos.

2xx confirma recebimento HTTP, não processamento interno do cliente. Timeout, rede, 429 e 5xx permitem retry com jitter/Retry-After; 401/403 suspendem para correção; 404/410 desabilitam destino para revisão; outros 4xx exigem diagnóstico. Defaults propostos: conexão 2 s, total 10 s, janela automática 72 h, atrasos 1/5/15 min e 1/4/8 h, depois a cada 12 h dentro da janela. Esgotamento gera EXHAUSTED e alerta; resultado continua disponível segundo retenção. Destino recuperado admite reenvio autorizado, sem reexecução nem receita duplicada.

## EXE-09 · Falhas, retry e encerramento — proprietário: Órbita / Cometa / SRE

Falha comprovadamente anterior ao envio permite retry seguro. Depois de envio possível, classificar UNKNOWN e consultar status ou reenviar com a mesma chave se a garantia idempotente do provedor for suficiente. Sem mecanismo confiável, abrir reconciliação manual; disponibilidade de outro provedor não elimina risco de duplicação.

Separar deadline de negócio, timeout por tentativa, janela de polling, retry de broker e entrega ao cliente. Erro permanente não recebe retry infinito. DLQ é fila de diagnóstico com responsável, aging e reprocessamento controlado, não encerramento da obrigação. Um reconciliador verifica protocolos sem progresso, intenção sem publicação, callback órfão, polling vencido, resultado sem entrega e fato sem apuração.

Resultado tardio após encerramento por SLA é obrigatoriamente rejeitado para publicação nesse protocolo, conforme EXE-12. Preservar evidência e conciliar não autoriza reabrir, revisar para sucesso ou enviar novo webhook de sucesso. Correção legítima de outro tipo de resultado final pode seguir revisão contratual auditada; ela não pode contornar um encerramento por prazo. Dados expurgados legitimamente não são ressuscitados.

## EXE-10 · TTL de reprocessamento em segundos — proprietário: Órbita / Cometa / Produto

retry_ttl_seconds é um inteiro não negativo obrigatório na política publicada do serviço/produto e na resolução do contrato do cliente. Zero desabilita retry de falha transitória; ausência não significa infinito. O valor efetivo é resolvido antes do aceite, limitado pelos tetos do serviço/produto e pela capacidade homologada. Persistir também máximo de tentativas e orçamento de retries por domínio de capacidade. TTL não é retenção de dados, timeout HTTP nem prazo de entrega do webhook.

Ao primeiro erro transitório recuperável do passo, persistir first_transient_at uma única vez. A janela termina no menor instante entre first_transient_at + retry_ttl_seconds, deadline do passo e deadline do cliente menos reserva de finalização. Reinício, backoff, duplicata, novo worker ou alternância de provedor não redefinem esse início. Produto mantém limite global e cada passo pode ter teto menor; não somar janelas de provedores para alongar o SLA do produto.

Enquanto a janela estiver aberta, houver orçamento de tentativa e a chamada for segura, agendar retry durável com backoff/jitter e respeitar controle adaptativo. Não enviar erro final ao cliente por uma indisponibilidade transitória ainda recuperável; ASYNC já possui aceite 202 e GET informa pendência. Erro permanente, cancelamento válido ou impossibilidade contratual conhecida pode concluir antes do TTL. Se next_run_at mais orçamento mínimo da tentativa/consolidação exceder o prazo restante, não iniciar nova tentativa inviável.

Falha comprovadamente sem efeito até esgotar TTL resulta em FAILED com motivo PROVIDER_UNAVAILABLE_RETRY_EXHAUSTED, se antes do deadline do cliente. Se o deadline venceu, concluir EXPIRED/SLA_EXCEEDED. Operação externa ainda incerta ao encerrar a janela termina para o cliente como EXPIRED/RETRY_WINDOW_EXHAUSTED e mantém reconciliação separada; esse motivo não prova quebra do SLA temporal do cliente nem ausência de custo.

Quando o provedor volta a aceitar a operação e ela segue pendente de conclusão normal, o fim do TTL de retry não é, sozinho, causa de término: esse TTL limita tentativas de recuperação, enquanto o SLA absoluto limita o processamento normal. A janela original permanece registrada e não é renovada por recuperação momentânea; nova falha posterior não ganha outra janela automaticamente. Operação UNKNOWN continua sujeita à regra de encerramento/reconciliação acima.

Um timeout de chamada não dá segurança para repetir: aplicar EXE-09. Circuito aberto preserva a obrigação e pode adiar tentativa, mas não pausa o relógio. Exemplo: aceite t=0, SLA do cliente 120 s, primeiro erro t=5, TTL 60 s e reserva final 5 s → retry_until=t=65; sem êxito, finalizar a falha conhecida nesse instante, não esperar indefinidamente até t=120.

## EXE-11 · Deadline, encerramento e disputa temporal — proprietário: Órbita / Produto

client_sla_seconds é o máximo contratual entre aceite durável e disponibilização da representação final no hub. O limite contratual máximo é accepted_at + client_sla_seconds; client_deadline_at é esse limite em ASYNC/AUTO e o menor limite compatível com a conexão em SYNC, conforme EXE-14. A redução é publicada na política de modalidade e registrada no aceite; nunca ampliada em voo. Inclui fila, reserva financeira, provedores, captura de objetos, composição e transformação personalizada. O SLA de tentativa de webhook é outro relógio; chegar ao deadline não dispensa avisar o cliente do encerramento.

provider_sla_seconds define obrigação de conclusão do provedor, com marco inicial explicitamente contratado (envio do hub ou aceite externo duravelmente observado). Métrica de resposta ao submit e prazo final externo são separados. Timestamp informado pelo provedor é preservado, mas não demonstra sozinho quando o hub recebeu e conservou o resultado. Se o contrato usa conclusão declarada pelo provedor, aferir adicionalmente o atraso de transmissão, sem estender o deadline do cliente.

Um temporizador durável da Órbita encerra protocolo aberto ao atingir o deadline, mesmo com polling saudável, HTTP 200 do provedor e ausência de erro técnico. Final e timeout disputam uma única transição terminal serializável. A elegibilidade exige representação final validada e confirmada de forma durável antes do limite; igualdade ao deadline é atraso. Não decidir pelo instante em que o worker do timeout conseguiu executar. Não usar relógio local divergente, timestamp de enqueue ou hora alegada pelo provedor para aceitar sucesso tardio.

A implementação futura deve demonstrar o ponto autoritativo de ordenação/commit e a arbitragem na fronteira do prazo, inclusive commit lento. Iniciar uma transação antes do deadline não prova que o resultado estava durável no prazo. O evento de sucesso não pode sair antes dessa verificação. O ensaio deve comprovar que não há publicação concorrente de sucesso e expiração; se a infraestrutura não consegue confirmar a elegibilidade temporal, não emitir sucesso não comprovado e registrar a indisponibilidade para recuperação/reconciliação. Isto é requisito a implementar e qualificar, não garantia já demonstrada.

Na expiração: persistir EXPIRED, SLA_EXCEEDED, prazo e causa; materializar representação final de erro do cliente; publicar a obrigação de notificação e fatos de SLA; cancelar timers produtivos e impedir novos passos. Reconciliar chamadas já enviadas e reservas/custos em fluxo separado. Scheduler atrasado é falha do hub mensurada; GET não executa provedor nem apresenta como sucesso um resultado fora do prazo. Se a finalização operacional do erro atrasar por falha de infraestrutura, registrar também esse atraso, sem retroagir falsamente o instante do commit.

Enforcement de SLA de provedor é política distinta: MONITOR_ONLY registra violação e pode ainda concluir no prazo do cliente; REJECT_LATE fecha o passo para uso no produto ao vencer o prazo do provedor. Failover segue segurança de efeito e prazo restante. Em ambos os casos, o deadline do cliente é rígido e não renovável. Comparação dessas políticas é obrigatória na publicação para não prometer obrigação contraditória.

## EXE-12 · Rejeição de retorno tardio sem apagar evidência — proprietário: Cometa / Órbita / Financeiro

Ao receber final depois de encerramento por SLA, persistir recibo autenticado como REJECTED_LATE_SLA, correlação, horários, hash e evidência permitida. Não substituir o erro final do protocolo, não reenviar sucesso, não reabrir passos e não capturar receita por sucesso. O retorno pode comprovar efeito externo ou custo e alimentar contestação contratual.

“Recusar” é decisão de negócio, não necessariamente devolver HTTP 4xx ao provedor. Se o webhook foi autenticado e seu recibo foi duravelmente registrado, o transporte pode responder 2xx para evitar uma tempestade de retransmissões; o recibo continua rejeitado para conclusão. Responder código específico de rejeição somente quando o contrato externo o suportar e seu efeito estiver homologado. Falha de persistência não recebe ack de sucesso.

Callback e polling que cruzam o deadline são arbitrados pela mesma regra. Mesmo resultado já iniciado no provedor antes do prazo pode chegar tarde ao hub. Resultado recebido no Cometa antes do limite, mas consolidado/publicável no hub depois, pode configurar quebra do hub e não necessariamente do provedor; a investigação usa os marcos de OPE-10. Processamento externo não desaparece porque o protocolo acabou.

Reexecução solicitada depois cria novo protocolo, nova chave e avaliação econômica; não reutiliza conteúdo rejeitado como sucesso do protocolo vencido. GET posterior e todas as tentativas do webhook de encerramento devolvem a representação final de erro original. Coleta residual para reconciliação deve ter teto, finalidade e permissão distintos; não se mascara como polling produtivo após expiração.

## EXE-13 · Paralelismo e espera eficiente — proprietário: Engenharia

Órbita agenda em paralelo apenas passos READY sem dependências pendentes, dentro de limites de protocolo, produto, tenant e célula. Dependências causais permanecem sequenciais. Consolidar produto depois da política de agregação, sem segurar thread ou transação enquanto aguarda todos os provedores. Paralelismo é avaliado pelo caminho crítico e não pela simples quantidade de goroutines.

Cometa e Pulsar usam pools limitados, I/O de rede concorrente e contexto de cancelamento. Esperas longas de polling/retry ficam em timers persistidos, não em thread dedicada, loop ocupado ou goroutine dormindo por toda a duração do protocolo. Transformações/arquivos intensivos em CPU são executados em pools isolados com orçamento, sem monopolizar o processamento de outras ofertas.

Uma chamada pendente não mantém transação PostgreSQL aberta. Limites de memória, fila, conexões e tarefas em voo são respeitados antes de reclamar mais trabalho. Desligamento drena tarefas ou devolve posse de forma segura; cancelar espera local não significa que o provedor cancelou sua operação. Bibliotecas que bloqueiam threads devem ser identificadas por profiling e isoladas/substituídas conforme evidência.

## EXE-14 · SYNC direto, prazo e resposta final — proprietário: Órbita / Cometa / Produto

SYNC é contrato de interação, não nome alternativo para polling do cliente. Pré-condições: todas as etapas necessárias são síncronas no provedor, o caminho crítico cabe no orçamento do cliente e existe capacidade reservada para chamada e commits. A admissão registra modo e limites; execução observa o mesmo grafo e as mesmas regras de segurança do ASYNC. Passos independentes podem chamar Cometa em paralelo, limitados por protocolo/tenant/provedor, e sucessores aguardam apenas suas dependências.

Sequência normativa para serviço simples: validar cliente/contrato/entrada e capacidade; confirmar protocolo/intenção no core; reservar financeiro se estrito; enviar comando DIRECT idempotente ao Cometa; Cometa confirma operação/tentativa e resolve segredo autorizado; chama provedor; confirma observação e intenção de fatos; retorna o final externo à Órbita; Órbita captura/consolida o que faltar, valida a projeção, decide o prazo e confirma seu final; responde pela conexão original. Publicação de fatos, cálculo pós-pago e webhook não são esperas obrigatórias desse percurso. Registro prévio de tentativa não significa que ela foi enviada: seu estado distingue preparação, envio possível e observação.

| Relógio/limite | Regra |
| --- | --- |
| connection_deadline_at | received_at na borda confiável + sync_wait_seconds homologado; pedido pode reduzir, nunca ampliar o limite autorizado |
| client_deadline_at SYNC | Menor entre accepted_at + client_sla_seconds e connection_deadline_at − margem de retorno HTTP |
| Deadline do provedor/passo | Não ultrapassa deadline do protocolo menos reserva de consolidação; inclui a segurança da próxima tentativa |
| Retry transitório | first_transient_at + retry_ttl_seconds, limitado pelo prazo efetivo do passo/protocolo; o TTL não prolonga a conexão |

A política publicada explicita essa diferença de relógios. Se o aceite já consumir o orçamento útil, recusar antes de efeito externo; se já houve commit, conservar o UUID e concluir falha/expiração de forma durável quando possível. Cliente que precisa de janela de recuperação maior que sua conexão escolhe ASYNC ou AUTO. SYNC não retorna 202 ao encontrar fila, provedor lento ou banco em failover; nessas situações pode ocorrer erro contratual/transporte com protocolo, sem esconder a quebra de disponibilidade.

Enquanto aguarda I/O, a chamada pode manter goroutine/contexto limitado, sem transação aberta nem thread dedicada por espera de rede suportada. Retries diferidos mantêm timer/intenção persistidos; o observador HTTP não se torna a autoridade do timer. Fechamento da conexão interrompe a espera local e propaga cancelamento quando seguro, mas não prova cancelamento do provedor. Sem final conhecido após possível efeito, manter operação UNKNOWN/reconciliação; ao prazo, EXPIRED continua impedindo sucesso tardio.

O mesmo corpo final persistido pode ser retornado na criação SYNC, no GET posterior e no webhook contratado, preservando COM-05 e a versão do cliente. Headers/status HTTP podem diferenciar criação, consulta e timeout. O compromisso normal é resposta final na mesma chamada; não existe garantia física de entrega se o cliente fechar a conexão ou a rede falhar. Esse risco tem comportamento observável e recuperação idempotente, sem alegar conclusão não confirmada.

## EXE-15 · Posse única entre despacho direto, fila e recuperação — proprietário: Engenharia

Na admissão, persistir command_id e dispatch_mode DIRECT para SYNC ou QUEUED para ASYNC/AUTO. O despachante de outbox só publica comandos elegíveis QUEUED. Intenção DIRECT tem observação de progresso e watchdog durável, mas não é simultaneamente duplicada como comando de execução em SQS. Fatos de domínio podem ser publicados em qualquer modalidade; são diferentes do comando que causa efeito.

Cometa reclama a intenção com chave única, lease/epoch e estado de operação. Repetição HTTPS ou mensagem de recuperação devolve/acompanha a operação existente. Um worker vencido não pode consolidar estado local. Antes de reassumir, distinguir comando nunca enviado, resultado local já registrado e envio possível com resposta desconhecida. Para UNKNOWN, consultar status seguro ou repetir somente sob idempotência externa homologada; lease expirado sozinho não torna seguro um segundo submit.

Transferir trabalho remanescente para fila exige transição durável exclusiva, liberação/fencing da posse anterior e conservação de command_id/operation_id. Se há chamada externa incerta, o novo trabalho é de reconciliação, não novo efeito. Essa transferência não muda o modo escolhido pelo cliente nem estende o deadline. Em SYNC com conexão perdida, a obrigação durável continua até finalização/encerramento contratual; não se promete retornar resposta a uma conexão já encerrada.

Crash depois do final Cometa e antes da resposta direta é resolvido por leitura da operação existente; o evento posterior também pode aplicar o mesmo fato. A consolidação Órbita deduplica ambos por operação/versão. Crash após o final Órbita e antes da resposta pública é resolvido pela mesma chave/protocolo. O watchdog, a fila e a resposta direta compartilham pré-condições e chaves econômicas; não há implementação negocial divergente por modalidade.

## EXE-16 · UUIDv7, repetição e consulta futura — proprietário: Órbita / Integrações

protocol_id usa UUIDv7 conforme RFC 9562 em todas as requisições duravelmente aceitas, SYNC, ASYNC ou AUTO, inclusive protocolos que terminem em falha/expiração. Não substituir por UUIDv4/ULID em um canal. Unicidade é protegida pela persistência; componente temporal não é identidade de tenant, prova de autorização ou prova de commit/deadline. Entropia, gerador concorrente e comportamento sob regressão de relógio serão qualificados. [RFC 9562 — UUIDv7](https://www.rfc-editor.org/rfc/rfc9562.html#section-5.7).

Resposta de criação informa UUIDv7 e endereço de consulta em headers documentados e no corpo canônico; perfis legados preservam a correlação por mapeamento homologado conforme COM-05. Repetição da mesma Idempotency-Key pelo mesmo tenant/operação recupera o mesmo UUID, mesmo se o cliente não recebeu a primeira resposta. A janela de deduplicação é contratada e não menor que o período operacional relevante; após expurgo legítimo da chave não se promete identificar a repetição antiga.

GET valida o tenant e lê exclusivamente o ecossistema do Hub. Para final existente e autorizado, devolve a representação contratada conservada; para indisponibilidade de fonte suficiente, aplica DAD-10. Nenhuma credencial de cliente consulta outro tenant por conhecer o UUID. O perfil administrativo individual usa rota própria e auditoria. Pedido inválido ou recusado antes do commit pode receber request_id técnico, mas esse identificador não deve ser anunciado como protocolo aceito e consultável.
