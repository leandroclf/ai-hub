# 03 · Persistência e dados

## DAD-01 · Bancos e propriedade — proprietário: Arquitetura / Dados

Persistência em PostgreSQL com três funções lógicas: **hub_control** para Atlas, **hub_core** para Órbita/Cometa/Pulsar e **hub_finance** para Libra. Cada célula de execução possui seu core e financeiro; o controle é independente e distribui projeções versionadas. A separação por célula limita contenção e falhas. Um contrato com saldo estrito tem uma única autoridade/célula financeira ativa; não se fragmenta o mesmo saldo entre células sem protocolo específico. Cada domínio tem usuário de aplicação, usuário de migração e permissões próprios. Mesmo no mesmo servidor, são proibidas escritas, joins operacionais e chaves estrangeiras entre domínios; referências externas são IDs e projeções.

Local/dev/hom podem compartilhar servidor entre bancos para reduzir custo, sem alegar isolamento físico. Em ppd/prd, core e financeiro têm instâncias/clusters separados por célula; classe dedicada usa também recursos próprios de nó/rede conforme OPE-08. hub_control tem persistência independente para que consultas administrativas e publicação não esgotem o banco de execução. O core usa primário com redundância multi-AZ; não há escrita ativa-ativa presumida. A separação por célula introduzida na v3 é mantida; a v4 acrescenta expansão automática por perfis, sem um banco por novo cliente por padrão.

Replicação melhora continuidade, mas não substitui backup e não assegura RPO zero em qualquer desastre. A autoridade de escrita atende aceite, transições, saldo estrito e leitura que exija atualidade. Réplicas servem relatórios e histórico com atraso explícito; leitura de final imutável comprovado e autorizado pode continuar por réplica/cache sob DAD-10. Ausência em cópia atrasada não demonstra inexistência. Pool de conexões é limitado por aplicação e dimensionado considerando o máximo de réplicas Kubernetes.

## DAD-02 · Modelo lógico mínimo — proprietário: Dados / Engenharia

| Entidade e proprietário | Informações obrigatórias | Integridade e acesso principal |
| --- | --- | --- |
| Cliente/aplicação — Atlas | tenant, aplicação, estado, credenciais referenciadas, permissões | Identidade dentro do tenant; sem credenciais em claro |
| Serviço/produto versão — Atlas | código, versão, schemas, grafo, modos, política final e retenção | Versão publicada imutável; consulta por código/versão |
| Conta/vínculo de provedor — Atlas | provedor, conta, ambiente, capacidades, contrato, endpoint, quotas, titular e responsável econômico | Separar identidades de contas, contratos e ambientes |
| Vínculo de credencial — Atlas | binding_id/version, SHARED_HUB ou TENANT_DEDICATED, tenant quando dedicado, conta externa, serviço, secret_ref, estado/validade, política de rotação | Um vínculo elegível resolvido; dedicado pertence a um único tenant; nunca valor secreto na entidade |
| Placement/capacidade — Atlas | tenant, célula, versão/epoch, estado de ativação, perfil, reservas e limites | Uma autoridade ativa de admissão por tenant; mapa distribuído autenticado e versionado |
| Contrato/plano versão — Atlas | partes, escopo, vigência, moeda, medidores, preço, franquia, limites | Unicidade da regra aplicável; vigências não ambíguas |
| Protocolo — Órbita | tenant, protocolo, chave idempotente, hash, pedido canônico, estado, datas, deadline, versões, resultado atual | Unicidade tenant/operação/chave; versão para concorrência |
| Passo — Órbita | protocolo, passo, serviço/versão, dependências, estado, input, regra e decisão de rota | Identidade por protocolo; grafo publicado fixado |
| Resultado — Órbita | protocolo, versão, estado final, schema, conteúdo/referência, checksum, datas, motivo de revisão | Imutável por versão; somente apontador atual é atualizado com auditoria |
| Operação externa — Cometa | operação, passo, provedor/conta, chave externa, provider_request_id, estado, contrato de compra | Correlação única no escopo declarado pelo provedor |
| Tentativa — Cometa | operação, tentativa, tipo SUBMIT/STATUS/FETCH/CANCEL, envio, término, erro, bytes e observação de custo | Nova tentativa não cria nova operação lógica por si só |
| Recibo externo — Cometa | identidade do evento, autenticação verificada, horário, referência de evidência, correlação | Recibo durável antes do ack; órfãos pesquisáveis |
| Agenda de polling — Cometa | operação, política/versão, próxima execução, deadline, lease, token de posse, contadores | Uma posse válida por operação; retomada após crash |
| Entrega — Pulsar | evento final, protocolo, resultado/versão, destino/versão, estado, próxima tentativa | Única por evento/destino; tentativas HTTP separadas |
| Uso/fato econômico — Libra | chave econômica, origem, contrato, quantidade, unidade, marco, datas, moeda | Dedup semântica além do event_id |
| Lançamento — Libra | lote, conta gerencial, débito/crédito, valor, moeda, regra, origem, reversão | Lote balanceado por moeda; sem UPDATE do lançamento publicado |
| Reserva/consumo de franquia — Libra | contrato, período, unidade, quantidade, estado, protocolo, expiração | Concorrência serializada no recurso financeiro estrito |
| Fatura/conciliação — Libra | período, contraparte, versão, linhas, evidências, divergências, total e estado | Não misturar moeda; fechamento rastreável |
| Objeto — domínio proprietário | object_id, tenant, finalidade, bucket/chave/versão, checksum, tamanho, tipo, estado, retenção | Referência imutável; URL temporária não é identidade |
| Outbox/inbox/timer — cada domínio | ID, tipo, agregado, payload/referência, agendamento, estado, tentativas | Mesma transação do efeito ou intenção; replay controlado |
| Intenção de despacho — Órbita / Cometa em seu domínio | command_id, modo DIRECT/QUEUED, epoch, posse/validade, operation_id, estado e reconciliação | Uma intenção lógica; transição de transporte explícita, sem emissão concorrente automática |
| Auditoria — cada domínio | ator, ação, recurso, antes/depois seguros, motivo, data, correlação | Registro append-only e acesso segregado |

São entidades lógicas e regras de integridade, não DDL. Modelo físico, índices finais e migrations exigem projeto e testes posteriores. Campos de negócio frequentemente consultados são tipados; JSONB é reservado a payloads canônicos/versionados e extensões. Não esconder tenant, estado, prazo ou chave idempotente somente dentro de JSON.

## DAD-03 · Transações e referências — proprietário: Engenharia

Na admissão, persistir atomicamente protocolo, idempotência, pedido ou referência já validada, versões aplicáveis e intenção de trabalho. Na conclusão, preparar e validar a representação do cliente e suas referências; persistir atomicamente estado, versão final, decisão de prazo, histórico e evento final. A arbitragem temporal segue EXE-11. O cliente nunca observa candidato final não confirmado. Não manter transação aberta enquanto espera HTTP, provedor, arquivo ou saldo remoto.

Índices previstos: tenant/protocolo; tenant/chave idempotente; tenant/estado/data; operação/conta/provider_request_id; agenda/next_run_at/estado; outbox pendente/data; entrega/estado/próxima tentativa; contrato/período/unidade; chave econômica. Histórico volumoso pode ser particionado por tempo após medição, preservando unicidade de idempotência em estrutura não sujeita a expurgo prematuro.

Valores monetários e quantidades fracionárias usam decimal exato; timestamps são UTC com fuso contratual preservado para fechamento. Toda atualização concorrente de estado verifica a versão. Lease e token de posse impedem worker antigo de consolidar localmente; não impedem um provedor de executar chamada já recebida.

## DAD-04 · Resultado como fonte de consulta — proprietário: Órbita

O hub conserva resultado normalizado, schema aplicado, versão, origem, horários e evidências necessárias. O endpoint público nunca usa um GET do cliente como gatilho para consultar o provedor. Isso vale para PENDENTE, FINAL, EXPIRADO, conteúdo expurgado e indisponibilidade externa. Polling do provedor é um processo de Cometa independente.

Um resultado final é servido do PostgreSQL ou de objeto pertencente ao hub; não devolver URL de provedor como única custódia da resposta. Se o provedor retorna link temporário, Cometa deve capturar o conteúdo permitido pelo contrato antes de considerá-lo disponível no hub. Não é permitido prometer custódia se o contrato do provedor proíbe armazenar os dados: nesse caso, o serviço não atende a esta oferta e precisa de decisão explícita antes de publicação.

Resultados expurgados retornam informação de indisponibilidade conforme retenção, sem recriar o pedido. “Atualizar dados” é uma nova execução com novo protocolo e avaliação de custo, não consulta do resultado antigo. Cache opcional usa tenant/protocolo/versão; cache miss lê persistência do hub.

## DAD-05 · Arquivos e atomicidade com objetos — proprietário: Órbita / Cometa

Entradas e saídas volumosas ficam no S3. Metadados e pequenos resultados tipados ficam no PostgreSQL. Estados de objeto: RESERVADO, RECEBIDO, EM_VALIDACAO, DISPONIVEL, REJEITADO, EXPURGADO. Reservar exige proprietário, tamanho máximo, tipo aceito, finalidade e prazo. Concluir upload verifica tamanho, checksum, tipo real e controles de conteúdo aplicáveis; só objeto DISPONIVEL pode integrar pedido executável.

Usar autorização temporária para chave específica, multipart quando apropriado e streaming com buffers limitados. Após validação, preservar versão imutável ou promover a objeto definitivo sem sobrescrita. Associar somente objetos do tenant com permissão e finalidade compatíveis. Links externos arbitrários não são uma forma alternativa de upload.

Como objeto e PostgreSQL não têm transação conjunta, primeiro confirmar conteúdo íntegro, depois confirmar referência e conclusão no banco. Falha entre etapas gera órfão para coleta, não resultado final com referência inexistente. Reconciliação deve detectar órfãos, referências quebradas e objetos presos em validação. Saída obrigatória indisponível impede publicação de sucesso completo.

Tetos iniciais propostos: JSON público 1 MiB; mensagem interna 128 KiB; objeto 1 GiB; soma por pedido 5 GiB. São limites do hub, não limites máximos do serviço de armazenamento. Transformações Base64 e fornecedores que exigem corpo integral têm limites próprios e qualificação de memória. Download é autorizado a cada emissão de URL; chave não constitui segredo de autorização.

## DAD-06 · Retenção e expurgo — proprietário: responsáveis pelos dados / Segurança

| Classe | Regra inicial da proposta | Bloqueio de expurgo |
| --- | --- | --- |
| Protocolos, metadados e resultados | Proposta de 90 dias online após conclusão; conteúdo pode exigir prazo distinto por contrato | Não publicar serviço sem prazo explícito e compatível com obrigação ao cliente |
| Protocolos ativos e operações incertas | Até conclusão/reconciliação, com aging e escalonamento | Não usar TTL para sumir com obrigação pendente |
| Resultado necessário a entrega pendente | Até término da janela contratual de entrega/reenvio, além do prazo normal se necessário | Entrega não pode depender de resultado já removido |
| Idempotência, inbox e correlação | Pelo menos 35 dias e todo período ativo; nunca menor que replay, retry e callbacks aceitos | Chaves financeiras seguem retenção financeira |
| Replay de eventos | 30 dias, em histórico durável fora da SQS quando exceder retenção da fila | Expiração na fila não apaga obrigação durável |
| Payload de callback/evidência | Mínimo necessário ao diagnóstico e conciliação; prazo por classificação | Preservação autorizada ou disputa |
| Uso, ledger, fatura, conciliação e auditoria financeira | Prazo contratual e regulatório a aprovar em P-04; sem número universal presumido | Bloquear ativação financeira sem política de retenção |
| Logs / traces / métricas | Propostas de 30 / 7 / 90 dias, sem payload sensível | Incidente pode exigir preservação autorizada |
| Upload abandonado | Proposta de remoção após 24 horas sem vínculo | Não remover versão referenciada por execução ativa |

Todo dado deve ter classe, finalidade, retenção, proprietário e política de backup. Prazos são parâmetros sujeitos a aprovação, não orientação jurídica. Expurgo deve abranger objetos, índices, caches e projeções; backups seguem seu ciclo e acesso restrito. Solicitação de exclusão conflitando com preservação contratual vai ao responsável pelos dados, com decisão auditável. Tombstone preserva apenas metadados mínimos autorizados; não conservar payload sob outro nome.

## DAD-07 · Isolamento, recuperação e auditoria — proprietário: Dados / SRE

Aplicar tenant_id nas entidades de cliente e nas constraints adequadas. RLS funciona como defesa adicional; contas de aplicação não podem ter BYPASSRLS e o papel de migração não é o de runtime. Contexto de tenant em conexões reutilizadas deve ser transacional e limpo a cada uso. Proprietários de tabelas e papéis privilegiados têm exceções que precisam ser consideradas, conforme a documentação de [RLS do PostgreSQL](https://www.postgresql.org/docs/current/ddl-rowsecurity.html).

Política inicial proposta: backups automáticos diários e PITR de 35 dias para PostgreSQL, criptografia e restauração isolada ensaiada mensalmente; proteção de versões de objetos e cópia para recuperação regional conforme retenção aprovada. Backups de core, financeiro e objetos devem formar um ponto de recuperação reconciliável. Não prometer snapshot atômico entre bancos distintos e S3. Ao restaurar: bloquear saída de chamadas, recuperar dados, confrontar outboxes e fatos, consultar operações incertas de forma segura, reconciliar ledger e somente então retomar efeitos. Um restore não autoriza reexecutar todos os pedidos do período.

## DAD-08 · Persistência adicional de contrato, prazo e pressão — proprietário: Dados / Engenharia

| Agregado | Campos/evidências conservados e ampliados na v4 | Regra de conservação |
| --- | --- | --- |
| Protocolo | accepted_at, client_deadline_at, contract_version, retry_policy_version, output_contract_version, cell_id, terminal_reason, finalized_at | Deadline absoluto não renovado por mensagem, migração ou retry |
| Passo/operação | first_transient_at, retry_until, provider_sla_start_at, provider_deadline_at, budgets, clock_source | Distinguir tempo total da operação, tempo de tentativa e tempo de fila |
| Observação externa | received_at, persisted_at, external_reported_at, eligibility_verdict, late_reason, evidência/checksum | Timestamps do provedor são evidência separada, nunca substituem o relógio autoritativo do hub |
| Representação do cliente | contrato/versão, canonical_result_version, media_type, bytes/referência imutável, hash, final_event_id | Uma representação final usada por GET e webhook; projeção não muda em replay |
| SLA | parte obrigada, contrato/versão, métrica, início/fim, prazo, breach, causa atribuída, evidência e revisão | Detalhe auditável por protocolo/operação, independente de métricas amostradas |
| Domínio adaptativo | conta/grupo de capacidade, controlador/epoch, taxa e concorrência atuais, amostra, regra, decisão, validade | Checkpoint durável e lease exclusivo; sinais em memória são reconstruíveis |
| Permissão operacional | concessão administrativa, identidade humana, escopo, validade, motivo e acessos | Consulta entre tenants via política explícita, não flag manipulável pelo cliente |

Indexar deadline de protocolos abertos, timers vencidos, fatos de SLA por período/contrato e relações tenant/célula. Índices e projeções de relatórios têm orçamento próprio; não varrer tabela de protocolo a cada tela de SLA. Uma tentativa tardia continua vinculada à operação original e ao eventual custo; sua evidência não entra na resposta de sucesso daquele protocolo expirado.

## DAD-09 · Dependência mínima de escrita, sem custódia fictícia — proprietário: Dados / Engenharia

Minimizar dependência significa eliminar consultas repetidas e operações duráveis desnecessárias; não eliminar a evidência indispensável de execução. Configuração, contratos e mapas já publicados são projeções locais válidas. Não consultar catálogo, cofre, relatórios, broker e financeiro pós-pago em série a cada chamada. O caminho SYNC simples tem as seguintes fronteiras lógicas, consolidadas por transação quando pertencem ao mesmo domínio:

| Fronteira | Registro necessário antes de prosseguir | Pode ser assíncrona? |
| --- | --- | --- |
| Aceite Órbita | UUIDv7, idempotência/hash, pedido/referências, snapshot, prazo e intenção DIRECT | Não; antecede confirmação e efeito externo |
| Preparação Cometa | Operação, vínculo/conta, tentativa, chave externa e posse válida | Não; permite identificar efeito possivelmente enviado |
| Observação Cometa | Retorno/evidência, resultado externo, custo observado e outbox | Não para afirmar que há final externo conservado |
| Final Órbita | Representação contratada, decisão terminal, hash e outbox | Não antes da resposta final ao cliente |
| Publicação e apuração pós-paga | Transporte da outbox, inbox, entregas e cálculo financeiro | Sim, dentro de prazo, capacidade durável e limite de risco |

Essas fronteiras não são um orçamento fixo de quatro comandos SQL: o projeto físico pode agrupar escritas locais, usar índices e reduzir round trips sem misturar autoridades. Produto com vários passos e contrato de saldo estrito adiciona trabalho. Leituras pós-commit desnecessárias são evitadas pelo uso dos dados confirmados; toda otimização conserva a prova de commit. Medir custo de cada fronteira em OPE-09.

Durante falha do writer, tentar recuperação de conexão/failover com orçamento curto e idempotência, sem manter transação aberta em espera externa. Commit com resposta perdida é estado incerto: consultar sua chave na mesma autoridade recuperada, não assumir rollback. Sem cópia autoritativa apta a confirmar, novas execuções não são enviadas ao provedor. Não trocar escrita para Redis, disco de pod, memória, um banco independente ou uma fila avulsa para emitir aceite fictício.

Um journal distribuído alternativo só poderia assumir a admissão se fosse projetado como autoridade única de idempotência, ordem, resultados e recuperação, com transição exclusiva e qualificação próprias. Isso não é o fallback da referência v4: duplicaria o problema de consenso e introduziria outra arquitetura. PostgreSQL gerenciado multi-AZ permanece a autoridade definida. A disponibilidade desse compromisso é tratada por redundância, capacidade e recuperação, com limites expostos em OPE-13/15.

## DAD-10 · Cache dispensável e continuidade de leitura — proprietário: Engenharia / Dados / Segurança

L1 em memória é limitado por bytes/entradas/idade; Redis L2 é opcional. Fluxo de leitura: resultado imutável válido no L1; L2 se habilitado e saudável; persistência do Hub. Nenhuma consulta pública aciona o provedor. A chave inclui tenant, protocolo, versão da representação e contrato técnico. Idempotência de criação continua autoritativa no banco, independentemente do cache.

Redis usa timeout pequeno, circuit breaker, bypass automático e recuperação gradual. Orçamento inicial de ensaio: até 5 ms para lookup L2 saudável; até 10 ms de tentativa degradada antes do bypass, contados na latência total. Valores não são SLA universal. Coalescer consultas iguais em voo, aplicar TTLs com jitter e aquecimento limitado para evitar avalanche contra a origem. A origem deve suportar o pico contratado com caches frios e falha de uma zona; cache não justifica subdimensionar o núcleo. O aceite inclui corte/reinício/latência de Redis e comparação com operação sem Redis.

Durante indisponibilidade de escrita, GET pode servir final imutável já confirmado de réplica ou snapshot local íntegro, dentro de retenção e autorização válidas. Para pedido de versão específica, a prova é o registro daquela versão. Para GET da versão corrente, é necessária projeção atual do ponteiro ou garantia publicada de que aquela classe de final não admite revisão; cache antigo de final revisável não é prova de atualidade. Estado EXPIRED por SLA conserva a regra de não reabertura. Não retornar 404 definitivo, pendência supostamente atual ou sucesso novo porque uma réplica atrasada não conhece o registro.

Consulta sem dado comprovadamente adequado retorna indisponibilidade transitória com correlação quando conhecida; não fabrica resultado. Expurgo/revogação invalidam caches conforme prazo contratado. Autorização é verificada antes de entregar bytes, inclusive em fallback. Quando não for possível garantir a atualidade mínima exigida de autorização/retenção, suspender a leitura afetada. Disponibilidade de cache não permite acesso entre tenants ou ressuscitar conteúdo apagado.

## DAD-11 · Dados, placement e crescimento de células — proprietário: Arquitetura / Dados / Plataforma

Novos tenants são alocados em células prontas com capacidade reservável. O crescimento do número de clientes se atende por novas células do mesmo perfil, sem criar tabela, schema, fila física ou deployment exclusivo para cada cliente compartilhado. Projeções e índices incluem tenant; listas administrativas usam paginação/cursor. Históricos e objetos crescem por particionamento/retenção e armazenamento gerenciado. HPA não amplia throughput de escrita do PostgreSQL; storage autoscaling, quando disponível, também não prova aumento de CPU/IOPS/locks.

O mapa tenant→célula tem versão/epoch e origem autenticada, distribuído pelo Atlas aos gateways e runtimes. Cliente não escolhe célula por header. UUIDv7 permanece conforme RFC, sem bits privados de roteamento. Para uma chamada, tenant autorizado e mapa determinam a autoridade de protocolo; não se adiciona consulta a um diretório global por UUID no caminho habitual. Busca administrativa entre células usa projeções com indicação de incompletude, não fan-out ilimitado.

Na referência inicial, um tenant possui uma célula ativa de admissão e uma autoridade financeira estrita. Crescimento dentro de um tenant ocorre pelos recursos/limites homologados dessa célula. Um tenant que se aproxime do envelope aciona aumento de capacidade ou realocação automática planejada antes de saturar. Fragmentar um único saldo/protocolo entre várias células ativas não é comportamento implícito; workloads acima do maior perfil homologado precisam de novo perfil qualificado antes de oferta comercial. Esse limite não impede expansão contínua por quantidade de clientes.

Realocação, quando necessária, é workflow automatizado: preparar destino e cópia de estado; drenar operações externas e reservas ou reconciliar as incertas; estabelecer barreira de admissão com espera limitada; completar cópia incremental e chaves de idempotência ainda válidas; revogar/fencear epoch de escrita/egress da origem; ativar destino e publicar novo mapa; conferir resultados/objetos/outboxes antes de retirar origem. A barreira não transforma SYNC em ASYNC nem ignora deadline. Se não couber no orçamento homologado, manter a autoridade antiga e abortar a migração antes do corte; uma operação incerta pode impedir a migração, mas não autoriza duas autoridades. A origem encaminha consultas/retries atrasados ao destino após o corte, sem executar localmente.

Automação deve provar recuperação de interrupção em cada fase e exclusividade efetiva de escrita/saída, não só troca de DNS. A expansão por novas células deve ocorrer preventivamente para que realocação não seja a resposta emergencial rotineira a todo pico. Cópias temporárias e índices de idempotência seguem a mesma proteção de dados. Mudança regional acrescenta as restrições de RPO e partição de OPE-15; não equivale a migração local comum.
