# Contratos, dados, estados e fronteiras de custódia

Status: modelo lógico e contratos propostos; nenhum DDL ou código de aplicação é entregue. Nomes de tabelas abaixo indicam entidades de destino e não pressupõem que todas existam no snapshot.

## Contrato público e compatibilidade

Manter `/v1/protocols` e `/v1/protocols/{uuid}` para o perfil atual, acrescentando autenticação real e validação de oferta. Hoje o corpo exige service_code/version e provider_account_id. O futuro contrato por oferta não deve remover campos sem migração: adaptar a v1 para uma oferta elegível ou introduzir versão pública negociada. Campo provider_account_id jamais autoriza a conta por si só. A escolha do perfil legado e sua data de descontinuação são explícitas.

| Situação | Resposta contratada |
|---|---|
| Recusa antes de aceite conhecido | 4xx ou 503 com código estável; sem protocolo fictício. Commit incerto instrui repetir mesma chave, sem afirmar inexistência |
| ASYNC aceito | 202, UUIDv7 e Location; somente após custódia de protocolo/intenção/snapshot |
| SYNC final no prazo | Final da modalidade na mesma conexão, materializado e identificado; falha de negócio também é final conforme perfil |
| SYNC sem prova de final no prazo | Erro de timeout/indisponibilidade com UUID e Location quando aceite conhecido; sem 202 implícito, sem sucesso não comprovado |
| AUTO | Aguarda auto_wait_seconds; final se pronto, senão 202 com a mesma identidade; um comando QUEUED |
| GET pendente | 200 com estado local autorizado, timestamps e próxima ação contratada; não consulta provedor |
| GET final | 200 e corpo final materializado; erro de negócio/SLA está no corpo. Protocolo alheio e inexistente não se distinguem |
| Repetição idempotente | Mesmo UUID e estado/representação aplicável; hash diferente gera 409. Perfil de status no replay deve estar documentado, sem reinterpretar estado |
| FileRef de resultado | Referência estável no corpo; obtenção de URL temporária é outra chamada autorizada |

UUIDv7 é identidade persistida, não relógio de deadline, chave de autorização ou trace_id. Requisição perdida é recuperada por Idempotency-Key; não depende de conseguir ler UUID de resposta que não chegou.

### Representação materializada

Guardar identidade do protocolo/cliente/aplicação, result_version, output_profile_id/version, media_type, bytes completos ou FileRef, hash, final_event_id e proveniência. Para v1, o wrapper completo atual (`protocol_id`, `status`, `result_version`, `mode`, `final_body`) faz parte desses bytes. Corrigir result_version interno para nova materialização sem reescrever silenciosamente resultados históricos. Perfis personalizados podem mapear conteúdo, mas devem garantir UUID por campo contratado ou header explícito quando formato legado não comportar mudança. GET e webhook do mesmo perfil reutilizam a mesma representação; diferenças de headers de transporte não quebram esse compromisso.

## Modelo lógico e acesso

| Autoridade | Entidades propostas | Identidades, campos e invariantes | Acessos/índices principais |
|---|---|---|---|
| Atlas / hub_control | client, application, grant | tenant/application, estado, escopo, revisão; histórico de concessões | tenant + application; lookup autorizado; paginação de clientes |
| Atlas | service_version, product_version, offer_version, technical_profile | schemas/hashes, DAG, modos, vigência, compra/venda, classe, transformação; publicado imutável | ID/versão; oferta por aplicação/vigência; filtros por estado |
| Atlas | provider_account, adapter_version, credential_binding_version | provider/account, ambiente externo, capacidades, binding, modo, tenant opcional conforme modo, secret_ref/version, settlement_party | unicidade por escopo/vigência; um vínculo elegível sem escolha ambígua |
| Atlas | publication, projection_manifest, import_batch | autor, diff, hash, relatório, destino/versão, estado de distribuição; importação em staging | versão monotônica e estado; reimport por origem/identidade estável |
| Atlas / placement | tenant_placement, capacity_profile, onboarding | célula, epoch, autoridade financeira, headroom, estado, razão | tenant/aplicação/oferta; histórico para protocolos anteriores |
| Órbita / core | protocol, idempotency_record, command_intent | UUID, tenant/application, operação/chave/hash; snapshot completo; accepted_at/deadline; dispatch_mode, command_id/epoch; compromisso de custódia | unicidade por escopo de idempotência; pendentes por idade/deadline; consulta tenant+UUID |
| Órbita | step_execution, dependency, compensation_obligation | grafo/version, estado por passo, pré-condições, resultado, deadlines, required/optional | protocolo+step_id; índice de passos prontos; dependências sem ciclo |
| Órbita | result_representation, terminal_decision | corpo/hash/versão/perfil, decisão de prazo e prova temporal; um final contratual | protocolo+versão+perfil único; final imutável por tenant |
| Cometa / core | operation, dispatch_claim, attempt | operation_id, command_id, account/binding/secret_version, provider_request_id, epoch, owner, lease; SUBMIT/STATUS/FETCH/CANCEL | command_id e unidade externa únicos; pendências por domínio/idade |
| Cometa | external_receipt, observation, reconciliation | event_id externo ou chave/hash da observação, recebido/persistido/alegado, verdict, late_reason, corpo permitido/ref; órfão preservado | origem+conta+ID externo; operação; recibos não associados |
| Cometa | polling_schedule, retry_schedule, capacity_domain/lease | next_run, retry_until, deadlines, owner/epoch, taxa, concorrência, pendências externas, validade | agenda por due/estado; claim indexado; orçamento agregado por domínio |
| Pulsar / core | destination_version, delivery, delivery_attempt/receipt | tenant+destino+versão, key_id, evento/result_version/hash, política, claim, attempts, horários/status | único evento+destino+versão; agenda indexada; consulta tenant+protocolo |
| Libra / hub_finance | economic_unit/fact, contract_snapshot, account_balance | natureza, unidade, medidor, compra/venda/versão, moeda/decimal, settlement_party, evidência | dedup semântica por obrigação, não apenas protocolo; conta por moeda |
| Libra | reservation, uncertain_hold, journal_batch/entry | reserva identidade/valor/moeda, estado, epoch financeiro, origem; journal balanceado, ajustes por referência | unicidade reserva; lock de conta; lote e soma por moeda |
| Libra | settlement_period, reconciliation_item, export_batch/receipt | período/watermarks, completos/pendentes, layout/versão/hash, status, recibo | competência/contrato; export_id único; lote fechado imutável |
| Domínio dono | outbox, inbox, quarantine, action_audit | IDs/timestamps/casualidade fixos; efeito + inbox atômico; quarentena conserva origem | consumer+event_id; outbox pendente por idade; audit por alvo/tempo |
| Domínio dono / S3 | file_ref, upload_session, retention_pin, tombstone | tenant, objeto/versão/hash/tipo/tamanho, estado, classe/região, prazo/hold, proveniência | file_id/version; pins por obrigação; expurgo por classe/estado |

Objetos e tabelas podem compartilhar infraestrutura em local, mas o isolamento lógico e permissões devem ser reais. Contas de aplicação não devem ser donas capazes de ignorar RLS; revisar exceções de owners/superusers conforme [PostgreSQL RLS](https://www.postgresql.org/docs/current/ddl-rowsecurity.html). Não usar FK, join de runtime ou escrita cruzando autoridades de domínio; usar IDs e projeções. Toda migration nova é adicionada, não edição retroativa de migration já aplicada.

## Estados, autoridade e encerramento

| Agregado | Estados/decisões principais | Dono e regra |
|---|---|---|
| Protocolo | ACCEPTED, RUNNING, WAITING_PROVIDER, RECONCILING; SUCCEEDED, PARTIALLY_SUCCEEDED, FAILED, EXPIRED, CANCELLED | Órbita; um terminal; cancelamento exige política explícita; não editar terminal manualmente |
| Operação externa | PREPARED, SUBMITTING, ACCEPTED_EXTERNAL, WAITING_FINAL, UNKNOWN; SUCCEEDED, FAILED, CANCELLED | Cometa; UNKNOWN não significa falha sem efeito; terminal externo pode coexistir com EXPIRED do protocolo |
| Observação | RECEIVED, ORPHAN, APPLIED, DUPLICATE, CONFLICTING, REJECTED_LATE_SLA, INVALID_CONTRACT | Cometa conserva recibo; somente observação validada e elegível pode afetar produto |
| Entrega | PENDING, DELIVERING, RETRY_SCHEDULED, DELIVERED, SUSPENDED, EXHAUSTED | Pulsar; 2xx confirma HTTP; reentrega usa mesmo final |
| Reserva | RESERVED, CAPTURED, RELEASED, UNCERTAIN_HOLD; estados de reconciliação pertinentes | Libra; EXPIRED de protocolo não libera automaticamente hold externo |
| Publicação | RASCUNHO, EM_VALIDACAO, APROVADO, PUBLICADO, SUSPENSO, DESCONTINUADO | Atlas; transições por permissão/versão; suspensão de oferta não apaga publicado histórico |
| Onboarding/célula | REQUESTED, VALIDATING, PROVISIONING, QUALIFYING, READY/ACTIVE, BLOCKED, DRAINING | Atlas/Plataforma; tráfego só em capacidade qualificada; segredo ou quota ausentes explicam bloqueio |

## Quatro fronteiras do SYNC simples

1. **Órbita:** aceite, idempotência, snapshot, deadline e intenção DIRECT no mesmo commit.
2. **Cometa:** operação, posse e tentativa antes do efeito externo; somente após commit permite envio.
3. **Cometa:** observação validada, resposta e outbox no mesmo commit; resposta direta informa fato conservado. A chamada ao parceiro ocorre fora da transação.
4. **Órbita:** representação final/decisão temporal/estado/outbox coerentes, antes de retornar sucesso ao cliente. A prova de commit na fronteira do deadline é T-R2-01, não resolvida por comparar now() antes de Commit.

Reserva estrita é uma fronteira adicional necessária, antes de qualquer efeito. Falha entre reserva e despacho deixa intenção retomável. Um retry não inventa nova reserva/operação. ASYNC/AUTO compartilham invariantes; transporte em fila não altera autoridade.

Outbox elimina o dual write somente quando participa da mesma transação do estado, como descrito no [padrão Transactional Outbox da AWS](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html). Delivery pode repetir e consumidores precisam deduplicar. Uma tabela chamada outbox em outra transação não basta.

## Envelope de mensagem e identidade

Fixar no commit: event_id, type, schema_version, producer, tenant_id/application_id quando aplicável, cell_id, protocol_id, step_id, operation_id, attempt_id, causation_id, occurred_at, recorded_at, aggregate_version, config_versions e payload pequeno/ref. Reenvio preserva conteúdo lógico e timestamps originais; publish_attempt_at é metadado distinto. Eventos sem tenant por natureza administrativa possuem schema explícito, não campo vazio por bug. Órbita e Cometa não devem gerar IDs baseados apenas no número local de uma outbox que colida entre células.

SQS Standard pode repetir entrega, conforme [documentação oficial](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/standard-queues-at-least-once-delivery.html). Validar envelope e versão antes do efeito; inbox+efeito/obrigação atomizados; ACK posterior. Visibility heartbeat impede repetição prematura, mas não substitui fencing. Quarentena e DLQ têm retenção, alerta e replay autorizado; replay preserva origem e usa a mesma idempotência semântica.

## Migração de dados existentes

Inventariar protocolos, resultados, operações, vínculos, contratos e fatos. Classificar por proveniência: confirmado com corpo externo, somente estado, sem snapshot, exemplo de simulador. Não corrigir resultado perdido copiando request_body nem reexecutando parceiro. Incluir marca de qualidade histórica fora dos bytes públicos já entregues; se necessário corrigir interpretação, fazê-lo por procedimento e perfil versionados, preservando evidência original.

Aplicar expand/contract, backfill em lotes retomáveis e comparações antes de obrigatoriedade de novos campos. Validar unicidades antes de constraints. Conservar histórico de origens e snapshots; isolar fatos financeiros não reconciliáveis como abertura legada, sem inventar compra/venda passada. Rollback conserva novas obrigações e impede antigo worker sem fencing de assumir trabalho novo.
