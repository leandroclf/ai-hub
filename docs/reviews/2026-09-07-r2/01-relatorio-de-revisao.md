# Revisão completa de implementação — AI Hub / Constelação — R2

Conclusão em 2026-09-08; snapshot `a39d394b0d87185ed4cc3861c12ec45f2c302d9e` de 07/09/2026. **A implementação é uma fatia inicial, não o atendimento integral dos requisitos levantados.** Os gaps do administrativo são extensos, mas dependem de correções críticas de identidade, custódia, execução, conteúdo e financeiro.

## Síntese verificável

Foram identificados **42 achados: 13 P0, 26 P1 e 3 P2**, todos com evidência de código/estrutura. A mitigação foi especificada em **9 changes, 66 requisitos R2 e 205 cenários**. Nenhuma correção de aplicação foi implementada nesta entrega.

P0 significa bloqueador de exposição/ativação do fluxo afetado no uso pretendido, não afirma que um incidente já ocorreu. P1 significa gap funcional/operacional relevante; P2 evolução de experiência/governança. Todas as constatações de comportamento são de análise estática, salvo as verificações explicitamente executadas no documento de qualidade. Riscos são consequências deduzidas dos caminhos observados.

## O que já existe e deve ser aproveitado

Cinco aplicações Go e gateway; UI React/TypeScript com quatro cadastros; três bancos lógicos; UUIDv7; protocolo consultável localmente; despacho HTTP direto e SQS assíncrono; outbox atômica na finalização da Órbita; filtros SQL por tenant; agenda de polling/webhook; estrutura inicial de fatos financeiros; probes e manifests de referência. Esses mecanismos são bases úteis, mas a presença do mecanismo não demonstra a garantia completa.

Build do frontend e go test ./... passaram nesta revisão. Não foram executados E2E com dependências, carga, cluster ou HA. A revisão visual usou screenshot existente no repo; não houve nova sessão autenticada do console.

## Decisões prioritárias

1. Fechar identidade/tenant e autenticação efetiva do provedor antes de exposição externa.
2. Preservar comando/observação/fato/entrega e ACK só após custódia; garantir posse da operação e resposta real.
3. Corrigir deadline/TTL, saldo/contratos e recursos isolados antes de anunciar SLA ou cobrança correta.
4. Evoluir catálogo executável e produtos/perfis, oferecendo APIs para jornadas administrativas reais.
5. Completar laboratório, escala e observabilidade; qualificar manutenção/recuperação no perfil antes de prd/crítico.

## Índice de achados

| ID | Prioridade | Achado | Change |
|---|---|---|---|
| F-01 | P0 | Identidade do cliente controlada pelo próprio chamador | r2-01-identidade-e-isolamento |
| F-02 | P0 | APIs internas e administrativas sem identidade de workload | r2-01-identidade-e-isolamento |
| F-03 | P0 | Callback confirma recebimento sem custódia comprovada | r2-03-integracoes-credenciais-e-pressao |
| F-04 | P0 | 202 após falha de envio de comando, sem retomada de despacho | r2-02-execucao-duravel-e-resultados |
| F-05 | P0 | Consumidores removem mensagens mesmo quando o efeito falha | r2-02-execucao-duravel-e-resultados |
| F-06 | P0 | Resultado retornado não é o resultado do provedor | r2-02-execucao-duravel-e-resultados |
| F-07 | P0 | Concorrência pode disparar mais de uma operação externa | r2-02-execucao-duravel-e-resultados |
| F-08 | P0 | Estado externo e fato são gravados em transações separadas | r2-02-execucao-duravel-e-resultados |
| F-09 | P0 | Corrida temporal permite sucesso depois do deadline | r2-02-execucao-duravel-e-resultados |
| F-10 | P1 | TTL de retry configurado não governa execução | r2-03-integracoes-credenciais-e-pressao |
| F-11 | P1 | SYNC, AUTO e UUID em erros não cumprem todo o contrato | r2-02-execucao-duravel-e-resultados |
| F-12 | P0 | Credencial dedicada resolvida não é a usada na chamada | r2-03-integracoes-credenciais-e-pressao |
| F-13 | P0 | Autenticação do simulador não equivale a integração segura real | r2-03-integracoes-credenciais-e-pressao |
| F-14 | P1 | Redis e broker bloqueiam inclusive reinício do caminho SYNC | r2-08-operacao-elastica-e-observavel |
| F-15 | P1 | Polling sem autenticação, lease e orçamento configurável | r2-03-integracoes-credenciais-e-pressao |
| F-16 | P1 | Não há amortecimento adaptativo nem isolamento de carga | r2-03-integracoes-credenciais-e-pressao |
| F-17 | P1 | Catálogo publicado pode ser sobrescrito e não governa admissão | r2-04-catalogo-produtos-e-contratos |
| F-18 | P1 | Produtos, DAG e contratos legados ainda não existem | r2-04-catalogo-produtos-e-contratos |
| F-19 | P1 | Importação não ativa adaptadores e substitui configurações locais | r2-04-catalogo-produtos-e-contratos |
| F-20 | P1 | Console limitado a quatro formulários e histórico efêmero | r2-05-console-administrativo |
| F-21 | P1 | Faltam jornadas administrativas de produto, operação e financeiro | r2-05-console-administrativo |
| F-22 | P2 | Formulários expõem detalhes técnicos e estados pouco guiados | r2-05-console-administrativo |
| F-23 | P0 | Custo e receita não usam contratos econômicos congelados | r2-06-financeiro-auditavel |
| F-24 | P0 | Saldo estrito não contabiliza consumo já capturado | r2-06-financeiro-auditavel |
| F-25 | P1 | Ledger, precisão monetária e fechamento não estão implementados | r2-06-financeiro-auditavel |
| F-26 | P1 | Webhook usa URL como identidade do segredo | r2-03-integracoes-credenciais-e-pressao |
| F-27 | P1 | Webhook sem claim, recibos completos e política por cliente | r2-03-integracoes-credenciais-e-pressao |
| F-28 | P1 | Resultado final não é materializado como representação imutável completa | r2-02-execucao-duravel-e-resultados |
| F-29 | P1 | Arquivos grandes não participam dos fluxos reais | r2-07-objetos-dados-e-retencao |
| F-30 | P1 | Persistência sem isolamento de papéis, expurgo e auditoria durável | r2-07-objetos-dados-e-retencao |
| F-31 | P1 | Docker local não inclui toda a plataforma e não garante persistência na recriação | r2-08-operacao-elastica-e-observavel |
| F-32 | P1 | Kubernetes é referência incompleta, sem os cinco ambientes | r2-08-operacao-elastica-e-observavel |
| F-33 | P1 | Autoscaling de células e reconciliação de IaC não fecham o circuito | r2-08-operacao-elastica-e-observavel |
| F-34 | P1 | Clientes AWS e bootstrap são fixos ao ambiente local | r2-08-operacao-elastica-e-observavel |
| F-35 | P1 | Probes não acompanham capacidades e não há drenagem graciosa | r2-08-operacao-elastica-e-observavel |
| F-36 | P1 | Observabilidade não mede SLA, pressão nem custódia | r2-08-operacao-elastica-e-observavel |
| F-37 | P1 | Ensaios de HA, isolamento e recuperação não foram demonstrados | r2-09-qualificacao-e-rastreabilidade |
| F-38 | P1 | Cobertura existente não prova as garantias que os nomes dos testes sugerem | r2-09-qualificacao-e-rastreabilidade |
| F-39 | P2 | Rastreabilidade histórica contém conclusões que não refletem o snapshot | r2-09-qualificacao-e-rastreabilidade |
| F-40 | P2 | Versões e contratos de ferramenta carecem de qualificação operacional | r2-09-qualificacao-e-rastreabilidade |
| F-41 | P1 | Destinos configuráveis não possuem defesa SSRF | r2-01-identidade-e-isolamento |
| F-42 | P1 | Recepção e pools não limitam custo por tenant | r2-08-operacao-elastica-e-observavel |

## Achados detalhados

### F-01 — Identidade do cliente controlada pelo próprio chamador

**Prioridade:** P0. **Responsável funcional:** Segurança e Engenharia de Plataforma. **Estado:** aberto; mitigação proposta.

**Fato observado:** POST e GET confiam em X-Tenant-Id. O gateway declara autenticação como placeholder. O filtro SQL por tenant existe, mas sua identidade não é autenticada.

**Evidências:** [hub/internal/orbita/handlers.go:95](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L95), [hub/internal/orbita/handlers.go:279](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L279), [hub/deploy/kong/kong.yml:4](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/kong/kong.yml#L4).

**Risco / inferência:** Um chamador que indique outro tenant pode atravessar a barreira de autorização; UUID não é controle de acesso.

**Correção e prova esperada:** Token do cliente A com header, corpo, URL ou cursor do cliente B não permite consultar nem executar em nome de B; chamadas sem identidade são negadas.

**Rastreio:** baseline SEG-01, SEG-03, EXE-07, EXE-16; [r2-01-identidade-e-isolamento](../../../openspec/changes/r2-01-identidade-e-isolamento/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-02 — APIs internas e administrativas sem identidade de workload

**Prioridade:** P0. **Responsável funcional:** Segurança e Engenharia de Plataforma. **Estado:** aberto; mitigação proposta.

**Fato observado:** A consulta interna usa GetByID sem tenant; Atlas, Cometa, Pulsar e Libra não verificam identidade/autorização nas rotas internas. Compose publica portas no host.

**Evidências:** [hub/internal/orbita/handlers.go:53](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L53), [hub/cmd/orbita/main.go:119](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/orbita/main.go#L119), [hub/internal/atlas/handlers.go:24](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/handlers.go#L24), [hub/deploy/docker-compose.yml:57](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L57).

**Risco / inferência:** Acesso à rede equivale a poderes sobre configuração, resultados e financeiro. Não há o perfil administrativo individual solicitado.

**Correção e prova esperada:** Identidades de workload com audience/escopo; APIs administrativas com OIDC/MFA e RBAC; hub_protocol_reader lê entre tenants pela rota administrativa auditada e não escreve.

**Rastreio:** baseline SEG-01, SEG-04, CFG-01, DAD-07, COM-01; [r2-01-identidade-e-isolamento](../../../openspec/changes/r2-01-identidade-e-isolamento/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-03 — Callback confirma recebimento sem custódia comprovada

**Prioridade:** P0. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Callback não autentica origem. ApplyObservation não retorna erro ao handler, e o handler responde 200; operação desconhecida ou falha de leitura é apenas ignorada pelo executor.

**Evidências:** [hub/internal/cometa/handlers.go:73](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/handlers.go#L73), [hub/internal/cometa/executor.go:267](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L267).

**Risco / inferência:** Pode haver observação forjada ou perda de resposta válida após ACK. Callback anterior à correlação também não tem inbox de órfãos.

**Correção e prova esperada:** Autenticar e persistir recibo antes de ACK; falha de persistência recebe erro retryable; órfão válido fica recuperável; rejeição tardia de negócio não apaga recibo.

**Rastreio:** baseline EXE-06, EXE-12, SEG-02, COM-03; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-04 — 202 após falha de envio de comando, sem retomada de despacho

**Prioridade:** P0. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** Protocolo é inserido antes de SendCommand. Erro de publicação só gera log e a resposta continua 202. Não há outbox de comando atômica nem scanner de recuperação do despacho.

**Evidências:** [hub/internal/orbita/handlers.go:178](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L178), [hub/internal/orbita/handlers.go:213](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L213), [hub/internal/orbita/dispatcher.go:78](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/dispatcher.go#L78), [hub/internal/orbita/store.go:74](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/store.go#L74).

**Risco / inferência:** Pedido aceito pode nunca chegar ao provedor e apenas expirar. A existência de uma outbox para fatos finais não resolve essa janela.

**Correção e prova esperada:** Crash entre aceite e broker e indisponibilidade do broker deixam uma obrigação durável que é retomada sem outro pedido do cliente.

**Rastreio:** baseline EXE-01, EXE-15, COM-03, DAD-03, OPE-11, DAD-02; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-05 — Consumidores removem mensagens mesmo quando o efeito falha

**Prioridade:** P0. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** Há Delete após Execute, falha de finalização, falha de agendamento de webhook e falha de apuração. A tabela inbox existe, mas nenhum consumidor a usa. Payload inválido também é descartado em caminhos.

**Evidências:** [hub/internal/cometa/worker.go:34](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/worker.go#L34), [hub/internal/orbita/factconsumer.go:51](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/factconsumer.go#L51), [hub/internal/pulsar/consumer.go:38](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/consumer.go#L38), [hub/internal/libra/consumers.go:45](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/consumers.go#L45).

**Risco / inferência:** SQS at-least-once não protege uma mensagem que a aplicação confirmou antes do commit. Pode desaparecer trabalho, notificação ou receita/custo.

**Correção e prova esperada:** Falha em cada commit impede ACK; replay aplica efeito uma vez; poison message ganha quarentena durável, diagnóstico e replay autorizado.

**Rastreio:** baseline COM-03, EXE-08, FIN-04, OPE-11; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-06 — Resultado retornado não é o resultado do provedor

**Prioridade:** P0. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** finalize atribui ResponseBody a cmd.RequestBody. Callback/polling recompõem comando sem esse corpo. Operações não persistem o resultado externo; replay terminal retorna apenas Kind. Decode e estado desconhecido podem resultar em SUCCEEDED.

**Evidências:** [hub/internal/cometa/executor.go:120](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L120), [hub/internal/cometa/executor.go:192](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L192), [hub/internal/cometa/executor.go:267](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L267), [hub/internal/cometa/store.go:45](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/store.go#L45).

**Risco / inferência:** Cliente pode receber eco da entrada, corpo vazio ou sucesso sem resposta válida, incompatível com a finalidade central do Hub.

**Correção e prova esperada:** Provedor retorna marcador diferente da entrada; SYNC, ASYNC, GET, webhook e replay preservam esse marcador, versão e hash; resposta inválida nunca vira sucesso.

**Rastreio:** baseline DAD-04, COM-05, EXE-04, EXE-14; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-07 — Concorrência pode disparar mais de uma operação externa

**Prioridade:** P0. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** CreateOperation usa ON CONFLICT DO NOTHING sem informar posse ao executor; os concorrentes continuam. UpdateState não usa CAS/epoch. A tentativa é registrada depois do envio, com erros de escrita ignorados.

**Evidências:** [hub/internal/cometa/store.go:32](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/store.go#L32), [hub/internal/cometa/executor.go:100](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L100), [hub/internal/cometa/store.go:85](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/store.go#L85), [hub/internal/dispatch/contract.go:40](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/dispatch/contract.go#L40).

**Risco / inferência:** Redelivery ou timeout pode duplicar efeito e custo no provedor. Epoch está no contrato, mas não cerca o executor.

**Correção e prova esperada:** Dois pods, fila e recuperação disputam a mesma operação: só um possui permissão de envio; crash após envio conserva UNKNOWN e não autoriza failover cego.

**Rastreio:** baseline EXE-04, EXE-09, EXE-15, DAD-03, COM-01, EXE-03; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-08 — Estado externo e fato são gravados em transações separadas

**Prioridade:** P0. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** finalize atualiza operação e limpa polling antes de abrir outra transação para outbox. Falha de publicação local apenas é registrada em log.

**Evidências:** [hub/internal/cometa/executor.go:192](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L192), [hub/internal/cometa/executor.go:219](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L219).

**Risco / inferência:** Cometa pode declarar sucesso sem fato recuperável para Órbita e Libra; final externo fica desconectado do protocolo e do custo.

**Correção e prova esperada:** Estado, resposta imutável, recibo e outbox compartilham commit; falha em qualquer fronteira não anuncia final durável.

**Rastreio:** baseline COM-03, DAD-03, EXE-14, FIN-04; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-09 — Corrida temporal permite sucesso depois do deadline

**Prioridade:** P0. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** UPDATE terminal compara versão e estado, mas não client_deadline_at. O timer periódico apenas disputa a mesma versão. Observações após terminal são ignoradas, sem recibo tardio.

**Evidências:** [hub/internal/orbita/store.go:204](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/store.go#L204), [hub/internal/orbita/deadline_timer.go:22](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/deadline_timer.go#L22), [hub/internal/cometa/executor.go:267](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L267).

**Risco / inferência:** Se o timer atrasar, final tardio pode ganhar e gerar receita; se o timer ganhar, a evidência de quebra pode desaparecer.

**Correção e prova esperada:** Com timer parado, resposta na igualdade ou depois do prazo nunca produz sucesso; disputa é decidida por regra temporal durável e preserva REJECTED_LATE_SLA.

**Rastreio:** baseline EXE-11, EXE-12, FIN-10, OPE-10, DAD-08; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-10 — TTL de retry configurado não governa execução

**Prioridade:** P1. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** retry_ttl_seconds está no catálogo; não é lido no caminho de despacho nem há first_transient_at/retry_until ou scheduler de retry. Falha de transporte vai para UNKNOWN sem reconciliador.

**Evidências:** [hub/internal/atlas/store.go:40](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/store.go#L40), [hub/internal/orbita/handlers.go:166](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L166), [hub/internal/cometa/executor.go:182](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L182).

**Risco / inferência:** O parâmetro exibido na UI não produz o comportamento prometido; indisponibilidade transitória não tem janela de recuperação contratada.

**Correção e prova esperada:** TTL=0 impede retry; primeiro erro transitório fixa a janela absoluta; reinício e troca de provedor não a renovam; UNKNOWN exige reconciliação antes de novo efeito.

**Rastreio:** baseline EXE-09, EXE-10, CFG-04; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-11 — SYNC, AUTO e UUID em erros não cumprem todo o contrato

**Prioridade:** P1. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** AUTO segue ASYNC sem espera limitada; não há elegibilidade SYNC por oferta/natureza do provedor; respostas 402/504 e alguns 500 após criar protocolo não o identificam. Prazo HTTP fixo pode divergir do SLA.

**Evidências:** [hub/internal/orbita/handlers.go:168](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L168), [hub/internal/orbita/handlers.go:187](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L187), [hub/internal/orbita/handlers.go:242](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L242), [hub/internal/platform/httpserver/httpserver.go:90](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/httpserver.go#L90).

**Risco / inferência:** Cliente pode perder o identificador de operação já aceita ou receber comportamento diferente do modo contratado.

**Correção e prova esperada:** SYNC elegível retorna final na conexão, nunca 202 implícito; AUTO espera seu orçamento; todo aceite tem UUIDv7 recuperável, inclusive erro posterior e repetição.

**Rastreio:** baseline EXE-02, EXE-14, EXE-16, COM-06, CAT-11; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-12 — Credencial dedicada resolvida não é a usada na chamada

**Prioridade:** P0. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** binding_id é gravado, mas Apply usa os campos de autenticação da conta de provedor; secret_ref do vínculo não chega ao transporte. Cache OAuth usa apenas providerAccountID.

**Evidências:** [hub/internal/cometa/executor.go:83](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L83), [hub/internal/cometa/executor.go:142](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L142), [hub/internal/providerauth/client.go:52](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/providerauth/client.go#L52).

**Risco / inferência:** Clientes dedicados na mesma conta podem executar com credencial compartilhada ou token de outro vínculo. Metadado correto não prova identidade externa correta.

**Correção e prova esperada:** Provedor de ensaio identifica principal externo por cliente; A e B compartilham conta, mas nunca token/certificado; rotação não reutiliza token de versão anterior.

**Rastreio:** baseline CFG-05, SEG-05, FIN-11; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-13 — Autenticação do simulador não equivale a integração segura real

**Prioridade:** P0. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Basic usa a referência como senha, OAuth envia client_secret_ref e mTLS é representado por um header. Não há resolução de cofre nem certificado no transporte TLS.

**Evidências:** [hub/internal/providerauth/client.go:57](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/providerauth/client.go#L57), [hub/internal/providerauth/client.go:102](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/providerauth/client.go#L102), [hub/internal/providerauth/client.go:117](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/providerauth/client.go#L117).

**Risco / inferência:** A integração real pode falhar ou aparentar segurança sem autenticação criptográfica efetiva.

**Correção e prova esperada:** Resolver referência apenas no workload autorizado; ensaiar OAuth e mTLS reais com certificado inválido/revogado e confirmar ausência de segredos em UI, eventos e logs.

**Rastreio:** baseline SEG-01, SEG-05, ARQ-03, COM-02; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-14 — Redis e broker bloqueiam inclusive reinício do caminho SYNC

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** Cometa dá panic se Redis não responde; erro de Set do token aborta autenticação. Órbita só inicia HTTP após criar/localizar recursos no broker. Resolução de credencial consulta Atlas por operação.

**Evidências:** [hub/cmd/cometa/main.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/cometa/main.go#L43), [hub/internal/providerauth/client.go:88](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/providerauth/client.go#L88), [hub/cmd/orbita/main.go:54](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/orbita/main.go#L54), [hub/internal/atlasclient/client.go:131](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlasclient/client.go#L131).

**Risco / inferência:** Manutenção do cache, broker ou controle derruba operações que deveriam funcionar com projeções válidas e persistência disponível.

**Correção e prova esperada:** Ensaiar Redis desligado, broker indisponível e Atlas parado com projeção válida, em processo aquecido e reiniciado; SYNC/GET mantêm o SLO do perfil qualificado.

**Rastreio:** baseline DAD-09, DAD-10, OPE-13, OPE-14, ARQ-02; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-15 — Polling sem autenticação, lease e orçamento configurável

**Prioridade:** P1. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Polling usa HTTP Get sem perfil de autenticação; não verifica status HTTP/erro de decode adequadamente; agenda não reivindica lease. Processa sequencialmente e descarta o backoff calculado.

**Evidências:** [hub/internal/cometa/poller.go:48](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/poller.go#L48), [hub/internal/cometa/store.go:134](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/store.go#L134).

**Risco / inferência:** Duas réplicas podem consultar a mesma operação; resposta de erro pode ser interpretada como final; um provedor lento atrasa os demais.

**Correção e prova esperada:** Polling privado autenticado, claim com fencing, jitter e Retry-After; taxa e custo contabilizados; callback simultâneo decide um final e preserva ambas observações.

**Rastreio:** baseline EXE-05, EXE-06, EXE-13, OPE-07; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-16 — Não há amortecimento adaptativo nem isolamento de carga

**Prioridade:** P1. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Workers fazem loops sequenciais compartilhados; não há controlador AIMD, bulkheads por tenant/conta, fairness ou limite global distribuído. Gateway tem quota local fixa.

**Evidências:** [hub/internal/cometa/worker.go:34](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/worker.go#L34), [hub/internal/cometa/poller.go:41](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/poller.go#L41), [hub/internal/pulsar/worker.go:51](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/worker.go#L51), [hub/deploy/kong/kong.yml:18](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/kong/kong.yml#L18).

**Risco / inferência:** Carga ruidosa e provedor degradado podem monopolizar filas/pools; aumentar pods pode multiplicar a pressão externa.

**Correção e prova esperada:** Controlador reduz pressão por 429/503/timeouts e recupera gradualmente; três réplicas obedecem um orçamento por domínio; tenant saudável mantém SLO no ensaio ruidoso.

**Rastreio:** baseline OPE-07, OPE-08, OPE-09, EXE-13, DAD-08; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-17 — Catálogo publicado pode ser sobrescrito e não governa admissão

**Prioridade:** P1. **Responsável funcional:** Produto, Atlas e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Upsert altera versão publicada, apesar do texto da UI. Schemas são gravados como objetos vazios e não expostos. Admissão valida contrato do tenant, sem catálogo/portfólio/rota elegível.

**Evidências:** [hub/internal/atlas/store.go:48](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/store.go#L48), [hub/internal/atlas/handlers.go:42](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/handlers.go#L42), [hub/internal/orbita/handlers.go:157](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L157).

**Risco / inferência:** Mudança retroativa pode alterar significado de pedidos; serviço inexistente ou não contratado pode alcançar um provedor indicado pelo cliente.

**Correção e prova esperada:** Publicado é imutável, rascunho exige controle de concorrência; oferta inexistente/inativa/incompatível é negada antes de efeitos e snapshot fixa execução.

**Rastreio:** baseline CAT-01, CAT-02, CAT-10, CFG-02; [r2-04-catalogo-produtos-e-contratos](../../../openspec/changes/r2-04-catalogo-produtos-e-contratos/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-18 — Produtos, DAG e contratos legados ainda não existem

**Prioridade:** P1. **Responsável funcional:** Produto, Atlas e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Há um serviço por comando, com StepID igual ao protocolo. Não existem produto composto, dependências, política de parcialidade, compensação nem transformações por cliente. Importação usa services como produtos.

**Evidências:** [hub/internal/orbita/handlers.go:23](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L23), [hub/internal/orbita/handlers.go:200](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L200), [hub/migrations/control/0001_init.sql:5](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/control/0001_init.sql#L5), [hub/migrations/control/0003_provider_api_catalog.sql:25](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/control/0003_provider_api_catalog.sql#L25).

**Risco / inferência:** O núcleo de agregação/composição e interoperabilidade legada não é coberto pelo scaffold.

**Correção e prova esperada:** Produto A+B paralelo e C dependente, versão fixa, erro obrigatório/opcional, compensação e dois perfis de contrato são executados com resultados e custos reconciliáveis.

**Rastreio:** baseline CAT-03, CAT-04, CAT-05, CAT-07, CAT-08, CAT-09, COM-04, ARQ-01, CAT-06; [r2-04-catalogo-produtos-e-contratos](../../../openspec/changes/r2-04-catalogo-produtos-e-contratos/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-19 — Importação não ativa adaptadores e substitui configurações locais

**Prioridade:** P1. **Responsável funcional:** Produto, Atlas e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Script local apaga catálogo, vínculos, contratos e contas; publica todos os itens em três modos. Metadados importados não são consumidos pelo executor, que chama /v1/operations do simulador. Collection/environment externos não estão no clone.

**Evidências:** [hub/deploy/import_hiveplace_collection.sh:38](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/import_hiveplace_collection.sh#L38), [hub/deploy/import_hiveplace_collection.sh:58](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/import_hiveplace_collection.sh#L58), [hub/internal/cometa/executor.go:120](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L120), [hub/migrations/control/0003_provider_api_catalog.sql:4](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/control/0003_provider_api_catalog.sql#L4).

**Risco / inferência:** Histórico pode perder contexto de controle e reinicialização não é reprodutível. Contagem de endpoints não equivale a integração homologada.

**Correção e prova esperada:** Importação em staging com diff e IDs estáveis, sem delete global; credenciais sanitizadas; nenhum endpoint se publica sem mapeamento, contrato e teste real/sandbox homologado.

**Rastreio:** baseline CAT-01, CAT-02, COM-02, CFG-02, QUA-01; [r2-04-catalogo-produtos-e-contratos](../../../openspec/changes/r2-04-catalogo-produtos-e-contratos/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-20 — Console limitado a quatro formulários e histórico efêmero

**Prioridade:** P1. **Responsável funcional:** Produto, Frontend e UX. **Estado:** aberto; mitigação proposta.

**Fato observado:** Não há listagens persistentes/paginação, edição por detalhe com conflito, pesquisa global, navegação por URL nem cadastro completo de clientes e aplicações. As tabelas vivem no estado da aba.

**Evidências:** [hub/admin-ui/src/App.tsx:7](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L7), [hub/admin-ui/src/pages/ServicesPage.tsx:15](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/ServicesPage.tsx#L15), [hub/admin-ui/src/api/atlasClient.ts:70](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/api/atlasClient.ts#L70).

**Risco / inferência:** Operador não encontra nem administra a população real de clientes, ofertas e provedores; refresh elimina o histórico visual.

**Correção e prova esperada:** Listagens do servidor, filtros paginados, URLs retomáveis e detail/edit auditáveis; refresh e troca de operador não dependem de dados da sessão anterior.

**Rastreio:** baseline CFG-01, CAT-01, CFG-06; [r2-05-console-administrativo](../../../openspec/changes/r2-05-console-administrativo/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-21 — Faltam jornadas administrativas de produto, operação e financeiro

**Prioridade:** P1. **Responsável funcional:** Produto, Frontend e UX. **Estado:** aberto; mitigação proposta.

**Fato observado:** Sem produto/DAG, publicação/diff, contratos por cliente, protocolos/timeline, SLA bilateral, webhook/replay, pressão, ledger/fechamento, auditoria e onboarding de capacidade.

**Evidências:** [hub/admin-ui/src/App.tsx:9](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/App.tsx#L9), [hub/internal/atlas/handlers.go:24](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/handlers.go#L24), [hub/internal/libra/handlers.go:25](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/handlers.go#L25).

**Risco / inferência:** Evoluir só a aparência dos quatro formulários não atende o escopo operacional do Hub.

**Correção e prova esperada:** Executar as jornadas ADM-01 a ADM-12 documentadas, com APIs autoritativas, papéis, estados de erro, observabilidade e trilha de ações.

**Rastreio:** baseline CFG-01, CFG-03, CFG-04, CFG-06, SEG-04, OPE-10, FIN-08; [r2-05-console-administrativo](../../../openspec/changes/r2-05-console-administrativo/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-22 — Formulários expõem detalhes técnicos e estados pouco guiados

**Prioridade:** P2. **Responsável funcional:** Produto, Frontend e UX. **Estado:** aberto; mitigação proposta.

**Fato observado:** Modo, estado e parte de liquidação usam texto livre em campos com enum; SLA aceita zero no HTML, mas backend exige positivo. Mensagem de publicação aparece mesmo ao salvar published=false. Faltam alertas acessíveis de resultado e conflitos.

**Evidências:** [hub/admin-ui/src/pages/ServicesPage.tsx:81](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/ServicesPage.tsx#L81), [hub/admin-ui/src/pages/ServicesPage.tsx:122](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/ServicesPage.tsx#L122), [hub/admin-ui/src/pages/ProviderAccountsPage.tsx:97](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/ProviderAccountsPage.tsx#L97), [hub/admin-ui/src/pages/CredentialBindingsPage.tsx:116](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/src/pages/CredentialBindingsPage.tsx#L116).

**Risco / inferência:** Operação manual fica propensa a erro e depende de conhecimento de endpoints/IDs. Acessibilidade não foi qualificada.

**Correção e prova esperada:** Controles de escolha e ajuda de negócio, validação coerente servidor/UI, foco e mensagens acessíveis, confirmação de impacto e estados draft/published distintos.

**Rastreio:** baseline CFG-01, CFG-02, QUA-03; [r2-05-console-administrativo](../../../openspec/changes/r2-05-console-administrativo/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-23 — Custo e receita não usam contratos econômicos congelados

**Prioridade:** P0. **Responsável funcional:** Financeiro, Comercial e Engenharia Libra. **Estado:** aberto; mitigação proposta.

**Fato observado:** Custo é constante 0,20 e apenas no sucesso. Receita busca o contrato atual do tenant. Não há aquisição versionada, incidência por tentativa/poll/fetch, vigência nem settlement_party aplicado.

**Evidências:** [hub/internal/libra/consumers.go:17](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/consumers.go#L17), [hub/internal/libra/consumers.go:67](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/consumers.go#L67), [hub/internal/libra/consumers.go:101](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/consumers.go#L101), [hub/internal/atlas/store.go:240](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/store.go#L240).

**Risco / inferência:** Troca de tarifa antes de consumir evento muda cobrança passada; CLIENT_DIRECT pode gerar custo do Hub; produto composto teria custo incorreto.

**Correção e prova esperada:** Snapshot de compra/venda no aceite; custos por unidade elegível e responsabilidade; mudança de preço, falha, parcialidade e resultado tardio reconciliam conforme exemplos.

**Rastreio:** baseline FIN-01, FIN-02, FIN-03, FIN-04, FIN-05, FIN-10, FIN-11; [r2-06-financeiro-auditavel](../../../openspec/changes/r2-06-financeiro-auditavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-24 — Saldo estrito não contabiliza consumo já capturado

**Prioridade:** P0. **Responsável funcional:** Financeiro, Comercial e Engenharia Libra. **Estado:** aberto; mitigação proposta.

**Fato observado:** Reserva soma somente estado RESERVED; CAPTURED sai do cálculo sem débito de saldo. Ausência de limite usa 1.000.000. Expiração do protocolo libera reserva sem checar incerteza externa.

**Evidências:** [hub/internal/libra/store.go:73](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/store.go#L73), [hub/internal/libra/store.go:100](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/store.go#L100), [hub/internal/libra/consumers.go:75](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/consumers.go#L75).

**Risco / inferência:** Consumo sequencial pode ultrapassar teto e uma reserva pode ser liberada enquanto o provedor ainda produzir custo.

**Correção e prova esperada:** Limite ausente nega oferta estrita; saldo considera liquidações, reservas e retenções incertas; captura repetida é idempotente; EXPIRED+UNKNOWN mantém retenção até reconciliação.

**Rastreio:** baseline FIN-06, FIN-10, DAD-03; [r2-06-financeiro-auditavel](../../../openspec/changes/r2-06-financeiro-auditavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-25 — Ledger, precisão monetária e fechamento não estão implementados

**Prioridade:** P1. **Responsável funcional:** Financeiro, Comercial e Engenharia Libra. **Estado:** aberto; mitigação proposta.

**Fato observado:** Valores trafegam em float64; economic_facts deduplica por protocolo/kind/meter. ledger_entries existe sem escritor. Sem partidas balanceadas, franquias/faixas, estornos, conciliação ou exportação fechada.

**Evidências:** [hub/internal/libra/store.go:27](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/store.go#L27), [hub/migrations/finance/0001_init.sql:19](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/finance/0001_init.sql#L19), [hub/migrations/finance/0001_init.sql:44](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/finance/0001_init.sql#L44), [hub/internal/libra/handlers.go:25](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/libra/handlers.go#L25).

**Risco / inferência:** Auditoria contábil não consegue explicar saldo; custos de passos distintos podem colidir e arredondamento não é governado.

**Correção e prova esperada:** Decimal exato fim a fim; identidade econômica por obrigação; lotes balanceados por moeda; ajustes compensatórios e fechamento/exportação idempotente com completude comprovada.

**Rastreio:** baseline FIN-02, FIN-04, FIN-07, FIN-08, FIN-09; [r2-06-financeiro-auditavel](../../../openspec/changes/r2-06-financeiro-auditavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-26 — Webhook usa URL como identidade do segredo

**Prioridade:** P1. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Segredo HMAC está em texto no banco e é resolvido por URL, não tenant/destino/versão. Dois tenants podem ter a mesma URL; rotação altera segredo de tentativas existentes. Não há assinatura com timestamp anti-replay.

**Evidências:** [hub/internal/pulsar/store.go:64](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/store.go#L64), [hub/internal/pulsar/store.go:91](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/store.go#L91), [hub/internal/pulsar/worker.go:71](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/worker.go#L71), [hub/migrations/core/0001_init.sql:101](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/core/0001_init.sql#L101).

**Risco / inferência:** Entrega pode ser assinada com chave de outro cliente ou perder verificabilidade após alteração de cadastro.

**Correção e prova esperada:** Delivery fixa tenant, destination_id/version, key_id e representação; mesma URL para dois tenants não cruza chaves; timestamp e assinatura dos bytes são verificáveis.

**Rastreio:** baseline EXE-08, SEG-02, COM-05, CFG-05; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-27 — Webhook sem claim, recibos completos e política por cliente

**Prioridade:** P1. **Responsável funcional:** Integrações e Engenharia de Core. **Estado:** aberto; mitigação proposta.

**Fato observado:** Workers consultam entregas devidas sem lease; timeout e maxAttempts fixos; envio é sequencial. Tentativas/recebimentos não têm registros completos. Corpo é buscado novamente e reserializado na Órbita.

**Evidências:** [hub/internal/pulsar/store.go:105](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/store.go#L105), [hub/internal/pulsar/worker.go:15](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/worker.go#L15), [hub/internal/pulsar/worker.go:59](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/worker.go#L59).

**Risco / inferência:** Réplicas disputam entregas sem coordenação, clientes lentos atrasam outros e reentrega não tem diagnóstico operacional suficiente.

**Correção e prova esperada:** Claims cercados, política versionada, recibos por tentativa, retry respeitando prazo e 429; reentrega só reutiliza resultado imutável e nunca chama provedor.

**Rastreio:** baseline EXE-08, EXE-13, COM-05, CFG-03; [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-28 — Resultado final não é materializado como representação imutável completa

**Prioridade:** P1. **Responsável funcional:** Engenharia de Core e Dados. **Estado:** aberto; mitigação proposta.

**Fato observado:** GET e webhook passam pelo mesmo respondWithProtocol, portanto não foi identificado wrapper diferente entre eles. Porém o wrapper é reconstruído, final_body.result_version fica zero e o teste compara somente o campo final_body, não todo o corpo nem HMAC.

**Evidências:** [hub/internal/orbita/finalize.go:49](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/finalize.go#L49), [hub/internal/orbita/handlers.go:303](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L303), [hub/test/e2e/e2e_test.go:163](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/test/e2e/e2e_test.go#L163).

**Risco / inferência:** Não há prova da identidade contratual completa, sobretudo após upgrade, perfis legados e rotação de destino.

**Correção e prova esperada:** Materializar o corpo completo da versão negociada, inclusive wrapper v1 quando aplicável; status/versão coerentes; comparar bytes completos, hash e assinatura antes/depois de restart e upgrade.

**Rastreio:** baseline COM-05, COM-04, DAD-04, CAT-09; [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-29 — Arquivos grandes não participam dos fluxos reais

**Prioridade:** P1. **Responsável funcional:** Dados, Segurança e Plataforma. **Estado:** aberto; mitigação proposta.

**Fato observado:** Cliente S3 é isolado; Put/Get usam buffers completos e EnsureBucket ignora qualquer erro. Não há upload direto, multipart, catálogo file_ref, checksum, quarentena ou autorização por objeto.

**Evidências:** [hub/internal/objectstore/objectstore.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/objectstore/objectstore.go#L43), [hub/internal/objectstore/objectstore.go:56](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/objectstore/objectstore.go#L56), [hub/internal/objectstore/objectstore.go:79](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/objectstore/objectstore.go#L79), [hub/internal/orbita/handlers.go:114](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L114).

**Risco / inferência:** Payload grande pode esgotar memória; sucesso do teste S3 não prova custódia/atomicidade de arquivo referenciado por protocolo.

**Correção e prova esperada:** Upload direto e finalização validam proprietário/tamanho/checksum; referência imutável antes de sucesso; broker leva só metadados; falha de S3 afeta apenas operações que dele dependem.

**Rastreio:** baseline DAD-05, SEG-02, COM-02, OPE-02; [r2-07-objetos-dados-e-retencao](../../../openspec/changes/r2-07-objetos-dados-e-retencao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-30 — Persistência sem isolamento de papéis, expurgo e auditoria durável

**Prioridade:** P1. **Responsável funcional:** Dados, Segurança e Plataforma. **Estado:** aberto; mitigação proposta.

**Fato observado:** Três bancos existem, mas o ambiente local usa o mesmo usuário hub. Não há RLS/papéis por domínio, trilha administrativa, políticas de retenção/tombstone ou restore coordenado no runtime.

**Evidências:** [hub/deploy/postgres-init/01-init.sh:8](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/postgres-init/01-init.sh#L8), [hub/deploy/docker-compose.yml:56](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L56), [hub/migrations/core/0001_init.sql:7](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/core/0001_init.sql#L7), [hub/migrations/core/0001_init.sql:122](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/core/0001_init.sql#L122).

**Risco / inferência:** Permissões excessivas e exclusão sem retenção de obrigações comprometem isolamento e futura recuperação.

**Correção e prova esperada:** Papéis de aplicação/migração separados; RLS testada com reutilização de pool; expurgo protege obrigações; restore com egress bloqueado reconcilia core/finance/objetos antes de retomar efeitos.

**Rastreio:** baseline DAD-01, DAD-06, DAD-07, SEG-03; [r2-07-objetos-dados-e-retencao](../../../openspec/changes/r2-07-objetos-dados-e-retencao/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-31 — Docker local não inclui toda a plataforma e não garante persistência na recriação

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** Compose contém backend, simuladores, Postgres, Redis, LocalStack, Kong e Swagger, mas não admin-ui, IdP, Prometheus, Loki, Alloy, Grafana ou Tempo. Não nomeia volume de dados PostgreSQL nem persistência LocalStack.

**Evidências:** [hub/deploy/docker-compose.yml:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L3), [hub/deploy/docker-compose.yml:12](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L12), [hub/deploy/docker-compose.yml:25](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L25), [hub/admin-ui/README.md:1](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/README.md#L1).

**Risco / inferência:** Subida completa e recuperação após recriar containers não são reproduzíveis; não se deve assumir perda em todo restart, mas os volumes anônimos não garantem reanexação no novo projeto.

**Correção e prova esperada:** Bootstrap documentado inicia todas as dependências e UI; volumes explícitos sobrevivem ao ciclo de recriação previsto; reset destrutivo é uma ação distinta de subir/atualizar.

**Rastreio:** baseline OPE-04, OPE-14, QUA-01, DAD-07; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-32 — Kubernetes é referência incompleta, sem os cinco ambientes

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** Manifests têm probes, recursos, réplicas, PDB e HPA em Atlas/Órbita. KEDA está comentado; faltam overlays local-kind/dev/hom/ppd/prd, imagens resolvidas, serviços auxiliares, identidade e rede. Não há kind config.

**Evidências:** [hub/deploy/k8s/README.md:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/README.md#L3), [hub/deploy/k8s/kustomization.yaml:16](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/kustomization.yaml#L16), [hub/deploy/k8s/cometa.yaml:8](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/cometa.yaml#L8), [hub/deploy/k8s/orbita.yaml:86](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/orbita.yaml#L86).

**Risco / inferência:** Não é uma instalação operacional, e declarar três réplicas não prova HA nem escala de workers.

**Correção e prova esperada:** Renderizar e subir os overlays em kind apropriado; ppd/prd qualificados em topologia real; um scaler por workload e dependências/probes efetivamente verificadas.

**Rastreio:** baseline OPE-04, OPE-05, OPE-12, ARQ-06; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-33 — Autoscaling de células e reconciliação de IaC não fecham o circuito

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** Placement não é usado pelo roteamento. Nomes de filas/tópicos são globais, embora bancos sejam por célula. Crossplane substitui assumeRolePolicy inteiro por ARN via fmt %s; não é JSON de trust policy. Não há controlador de onboarding/headroom ou instalação de Karpenter.

**Evidências:** [hub/migrations/control/0001_init.sql:62](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/migrations/control/0001_init.sql#L62), [hub/deploy/terraform/main.tf:517](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/terraform/main.tf#L517), [hub/deploy/terraform/crossplane/composition.yaml:272](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/terraform/crossplane/composition.yaml#L272), [hub/deploy/terraform/main.tf:736](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/terraform/main.tf#L736).

**Risco / inferência:** Células não são isoladas de ponta a ponta; adicionar clientes não aciona provisionamento seguro automático. O exemplo Crossplane não é aplicável como está.

**Correção e prova esperada:** Nova capacidade nasce por política, é homologada e só então recebe placement; histórico mantém célula de origem; recursos têm dono único Terraform/Crossplane e trust policy válida com audience/subject.

**Rastreio:** baseline CFG-06, DAD-11, ARQ-06, OPE-12; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-34 — Clientes AWS e bootstrap são fixos ao ambiente local

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** SDK usa credenciais estáticas local/local em qualquer ambiente; aplicações criam recursos e aplicam política permissiva ao iniciar. Não há workload IAM configurado nos Deployments.

**Evidências:** [hub/internal/queue/queue.go:58](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/queue/queue.go#L58), [hub/internal/queue/queue.go:126](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/queue/queue.go#L126), [hub/internal/objectstore/objectstore.go:27](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/objectstore/objectstore.go#L27), [hub/cmd/orbita/main.go:74](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/orbita/main.go#L74).

**Risco / inferência:** Trocar o endpoint para AWS não é migração suficiente. Aplicação não deve reescrever políticas que IaC gerencia.

**Correção e prova esperada:** Modo local explicitamente separado; remoto usa identidade de workload e URLs/ARNs provisionados, sem CreateQueue/SetQueueAttributes no runtime; testes negam credencial local em ambiente remoto.

**Rastreio:** baseline ARQ-03, ARQ-05, SEG-01, OPE-04; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-35 — Probes não acompanham capacidades e não há drenagem graciosa

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** Há três endpoints; readiness/startup apenas pingam banco, liveness é independente. Código não captura SIGTERM nem chama Shutdown; pods declaram terminationGracePeriod sem rotina de drenagem.

**Evidências:** [hub/internal/platform/httpserver/httpserver.go:86](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/httpserver.go#L86), [hub/cmd/orbita/main.go:115](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/orbita/main.go#L115), [hub/cmd/cometa/main.go:52](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/cmd/cometa/main.go#L52), [hub/deploy/k8s/orbita.yaml:82](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/orbita.yaml#L82).

**Risco / inferência:** Readiness pode declarar pronto sem configuração válida ou derrubar consulta final por falha exclusiva de writer. Atualização interrompe chamadas e workers em fronteiras perigosas.

**Correção e prova esperada:** Separar papéis de leitura/admissão/workers com probes próprias; drenar HTTP/claims em SIGTERM; persistir incerteza e permitir retomada cercada.

**Rastreio:** baseline OPE-05, OPE-13, OPE-14, DAD-10; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-36 — Observabilidade não mede SLA, pressão nem custódia

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** Contador HTTP manual e logs com IDs existem; não há histogramas, spans OTel, scrape/painéis/alertas, métricas de lag/UNKNOWN, dual SLA ou runbooks implantados.

**Evidências:** [hub/internal/platform/httpserver/metrics.go:16](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/metrics.go#L16), [hub/internal/platform/httpserver/httpserver.go:115](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/httpserver/httpserver.go#L115), [hub/internal/dispatch/contract.go:35](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/dispatch/contract.go#L35), [hub/deploy/docker-compose.yml:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/docker-compose.yml#L3).

**Risco / inferência:** Operações não consegue aferir contrato, detectar aceite sem despacho ou distinguir lentidão do Hub e do provedor.

**Correção e prova esperada:** Dashboards e consultas por coorte/contrato com denominadores definidos, exemplares de trace e alertas exercitados para atraso, erro de custódia, pressão, DLQ e saldo.

**Rastreio:** baseline OPE-03, OPE-06, OPE-10, CFG-03; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-37 — Ensaios de HA, isolamento e recuperação não foram demonstrados

**Prioridade:** P1. **Responsável funcional:** Qualidade, Liderança técnica e SRE. **Estado:** aberto; mitigação proposta.

**Fato observado:** Evidências locais e manifests não demonstram perda de zona/região, autoscaling integral, manutenção sob carga ou restore sem reexecutar efeitos. Nenhum perfil crítico está qualificado.

**Evidências:** [hub/deploy/k8s/README.md:106](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/README.md#L106), [hub/evidence/EVIDENCE.md:1](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/evidence/EVIDENCE.md#L1), [hub/test/e2e/e2e_test.go:9](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/test/e2e/e2e_test.go#L9).

**Risco / inferência:** Promessas de ausência de perda, indisponibilidade imperceptível e uso crítico permanecem sem evidência suficiente.

**Correção e prova esperada:** Registrar resultados e limites dos ensaios por perfil; impedir ativação de SLA/criticidade que não passou gates; não apresentar RTO/RPO de referência como medidos.

**Rastreio:** baseline OPE-01, OPE-03, OPE-11, OPE-15, QUA-06, DEC-02; [r2-09-qualificacao-e-rastreabilidade](../../../openspec/changes/r2-09-qualificacao-e-rastreabilidade/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-38 — Cobertura existente não prova as garantias que os nomes dos testes sugerem

**Prioridade:** P1. **Responsável funcional:** Qualidade, Liderança técnica e SRE. **Estado:** aberto; mitigação proposta.

**Fato observado:** Sucesso SYNC verifica apenas status; igualdade verifica campo interno; limite excedido só verifica uma reserva permitida com limite default. Só Atlas e UUID têm testes unitários. Seed necessário aos E2E foi desabilitado.

**Evidências:** [hub/test/e2e/e2e_test.go:78](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/test/e2e/e2e_test.go#L78), [hub/test/e2e/e2e_test.go:152](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/test/e2e/e2e_test.go#L152), [hub/test/e2e/e2e_test.go:230](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/test/e2e/e2e_test.go#L230), [hub/internal/atlas/handlers_test.go:5](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/handlers_test.go#L5).

**Risco / inferência:** Suíte verde pode manter os bugs de conteúdo, saldo, segurança e replay. Ambiente limpo não tem as fixtures históricas esperadas.

**Correção e prova esperada:** Testes com oráculos de payload/identidade/saldo/custódia, concorrência e falhas, fixtures isoladas versionadas e pipeline que publica evidencia por SHA.

**Rastreio:** baseline QUA-01, QUA-02, QUA-03, QUA-04, QUA-05, QUA-06; [r2-09-qualificacao-e-rastreabilidade](../../../openspec/changes/r2-09-qualificacao-e-rastreabilidade/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-39 — Rastreabilidade histórica contém conclusões que não refletem o snapshot

**Prioridade:** P2. **Responsável funcional:** Qualidade, Liderança técnica e SRE. **Estado:** aberto; mitigação proposta.

**Fato observado:** TRACEABILITY afirma cache ausente, SEG-05 implementado e prova bit-a-bit completa; código atual tem Redis obrigatório, autenticação simulada e teste parcial. Tasks v4 permanecem abertas apesar de outros relatórios de conclusão.

**Evidências:** [openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md:68](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md#L68), [openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md:84](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md#L84), [openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md:138](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md#L138), [openspec/changes/hub-interoperabilidade-v4/tasks.md:1](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/openspec/changes/hub-interoperabilidade-v4/tasks.md#L1).

**Risco / inferência:** Próxima rodada pode herdar falso pronto. Não foi encontrada árvore canônica openspec/specs nem configuração de workflow; baseline vive em uma change aberta.

**Correção e prova esperada:** Separar alvo normativo, entrega e prova; manter 98 IDs; atualizar status só com evidência; definir reconciliação de baseline sem arquivar a v4 inteira como implementada.

**Rastreio:** baseline DEC-03, DEC-04, DEC-05, QUA-03, QUA-04; [r2-09-qualificacao-e-rastreabilidade](../../../openspec/changes/r2-09-qualificacao-e-rastreabilidade/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-40 — Versões e contratos de ferramenta carecem de qualificação operacional

**Prioridade:** P2. **Responsável funcional:** Qualidade, Liderança técnica e SRE. **Estado:** aberto; mitigação proposta.

**Fato observado:** Há versões e lockfiles, mas sem matriz de suporte, SBOM/scan automatizado, compatibilidade testada de controllers ou contratos gerados/validados contra consumidores. Não se inferiu vulnerabilidade apenas pela idade.

**Evidências:** [hub/go.mod:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/go.mod#L3), [hub/deploy/Dockerfile:5](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/Dockerfile#L5), [hub/deploy/terraform/variables.tf:121](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/terraform/variables.tf#L121), [hub/admin-ui/package.json:6](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/admin-ui/package.json#L6), [hub/api/asyncapi.yaml:83](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/api/asyncapi.yaml#L83).

**Risco / inferência:** Atualizações podem quebrar adapters, manifests e UI; versões de exemplo podem ser confundidas com versões aprovadas.

**Correção e prova esperada:** Fixar toolchain/imagens por versão/digest e verificar suporte/licença no momento da promoção; contract tests e diff de OpenAPI/AsyncAPI acompanham mudanças.

**Rastreio:** baseline ARQ-04, ARQ-05, COM-04, QUA-04; [r2-09-qualificacao-e-rastreabilidade](../../../openspec/changes/r2-09-qualificacao-e-rastreabilidade/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-41 — Destinos configuráveis não possuem defesa SSRF

**Prioridade:** P1. **Responsável funcional:** Segurança e Engenharia de Plataforma. **Estado:** aberto; mitigação proposta.

**Fato observado:** URLs de provedor/token/webhook não têm política de redes autorizadas, validação DNS no envio nem bloqueio de redirecionamento. O cliente HTTP padrão pode seguir redirects.

**Evidências:** [hub/internal/atlas/handlers.go:100](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlas/handlers.go#L100), [hub/internal/pulsar/handlers.go:24](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/handlers.go#L24), [hub/internal/pulsar/worker.go:80](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/pulsar/worker.go#L80), [hub/internal/cometa/executor.go:135](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/cometa/executor.go#L135).

**Risco / inferência:** Cadastro privilegiado ou destino comprometido pode acessar serviços internos/metadados ou desviar credenciais. Exposição pública depende do ambiente, não foi explorada nesta revisão.

**Correção e prova esperada:** Validar cadastro e conexão efetiva, destinos privados só por perfil aprovado, impedir rebinding/redirect não autorizado e não propagar Authorization para outra origem.

**Rastreio:** baseline SEG-02, CFG-05, COM-02; [r2-01-identidade-e-isolamento](../../../openspec/changes/r2-01-identidade-e-isolamento/proposal.md). Cenários e tarefas da change tornam essa prova executável.

### F-42 — Recepção e pools não limitam custo por tenant

**Prioridade:** P1. **Responsável funcional:** Plataforma, SRE e Engenharia. **Estado:** aberto; mitigação proposta.

**Fato observado:** ReadAll sem limite no POST, cache Atlas sem política visível de evicção e pools globais com limites por processo. Workers recebem lotes e executam sequencialmente sem extensão de visibility timeout.

**Evidências:** [hub/internal/orbita/handlers.go:114](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/orbita/handlers.go#L114), [hub/internal/atlasclient/client.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/atlasclient/client.go#L43), [hub/internal/platform/pg/pg.go:29](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/platform/pg/pg.go#L29), [hub/internal/queue/queue.go:199](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/queue/queue.go#L199).

**Risco / inferência:** Memória, conexões e visibilidade de mensagens podem saturar antes de CPU indicar escala; redelivery cresce em operações lentas.

**Correção e prova esperada:** Limitar bytes/complexidade/tempo, caches por tamanho, quotas de fila/pool por classe, renovar visibility com lease e medir pressão antes de elevar réplicas.

**Rastreio:** baseline OPE-02, OPE-08, OPE-09, EXE-13; [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md). Cenários e tarefas da change tornam essa prova executável.

## Cobertura dos 98 requisitos anteriores

| Situação no snapshot | Quantidade |
|---|---|
| DOCUMENTAL | 5 |
| ESCOLHA_CONSTATADA | 1 |
| NAO_IMPLEMENTADO | 29 |
| PARCIAL_NAO_QUALIFICADO | 63 |

A classificação não é percentual de implementação. NAO_IMPLEMENTADO significa ausência da obrigação completa, mesmo quando há campos ou comentários relacionados. PARCIAL_NAO_QUALIFICADO inclui código existente com lacunas de regra/prova. ESCOLHA_CONSTATADA limita-se à linguagem/stack encontrada; DOCUMENTAL confirma registro normativo, sem presumir decisões aprovadas. A matriz [02-rastreabilidade-98.csv](02-rastreabilidade-98.csv) identifica cada requisito, evidência, incremento, cenário e tarefa; não utiliza o total histórico de “implementados” como conclusão desta revisão.

## Limites e prontidão

Esta auditoria não confirma vulnerabilidade explorada, estado de produção, HA, carga ou certificação. Há bloqueios demonstráveis no caminho de código para as garantias pretendidas. P-01 a P-11 e T-R2-* permanecem onde exigem decisão/prova; não foram resolvidos por artefatos ou avanço do repositório. Nenhuma ausência de banco é resolvida por aceite fictício; nenhuma disponibilidade regional é inferida de três réplicas ou backup.
