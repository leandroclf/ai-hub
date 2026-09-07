# Tasks: Hub de Interoperabilidade "Constelação" v4.0

Estas tarefas decompõem a fatia inicial recomendada por QUA-04 (`docs/08_QUALIDADE_E_ACEITE.md`): serviço síncrono de provedor disponível em SYNC direto e ASYNC, com UUIDv7, resultado local, credencial compartilhada/dedicada e uma regra simples de receita/custo — seguida por polling/callback concorrentes e, por fim, produto composto e planos avançados. Nenhuma tarefa abaixo foi executada nesta mudança; `tasks.md` é o plano para a implementação futura, condicionado ao gate G0 (revisão) descrito em `qualidade-e-aceite/spec.md`.

Toda tarefa que envolva dependência de infraestrutura real (conta AWS, região, orçamento) permanece bloqueada por P-01 até sua resolução; tarefas sintéticas/locais (Compose, kind, provedor simulado) não dependem dessa resolução.

## 1. Discovery and setup

- [ ] 1.1 Confirmar responsáveis nominais e registrar aprovação do gate G0 para a fatia inicial
  - Objective: Formalizar quem aprova domínio, estados, contratos econômicos e critérios de aceite da fatia (QUA-04 G0).
  - Likely files/components: `openspec/changes/hub-interoperabilidade-v4/` (este pacote), registro de decisão (P-09).
  - Depends on: nenhuma (usa este proposal/specs/design já produzidos).
  - Validation: manual
  - Completion criteria: responsáveis nominais registrados para Produto, Arquitetura, Engenharia, QA, Segurança, Financeiro; nenhuma pendência bloqueante da fatia (P-09) em aberto.

- [ ] 1.2 Preparar ambiente local (Docker Compose) e laboratório Kubernetes (kind)
  - Objective: Disponibilizar PostgreSQL, emulação AWS (SQS/SNS/S3 local) e observabilidade em containers, identificados como emulação (OPE-04).
  - Likely files/components: infraestrutura local de desenvolvimento (fora do escopo desta mudança documental; a implementar).
  - Depends on: 1.1
  - Validation: manual
  - Completion criteria: ambiente local sobe com PostgreSQL, fila/objeto emulados e stack de observabilidade; documentado como não comprovando HA multi-AZ.

- [ ] 1.3 Definir provedor simulado determinístico para a coleção sintética de referência
  - Objective: Simulador local capaz de reproduzir SYNC, ASYNC (callback/polling), falhas injetadas e comportamento determinístico (QUA-01).
  - Likely files/components: simulador de provedor (a implementar).
  - Depends on: 1.2
  - Validation: integration
  - Completion criteria: simulador cobre ao menos os casos da coleção sintética mínima de QUA-01 (serviço síncrono, assíncrono callback/polling/ambos, operação sem idempotência, callback antecipado, resultado grande, provedor que cobra status).

## 2. Contract and specs

- [ ] 2.1 Especificar OpenAPI do endpoint de admissão/consulta do serviço síncrono da fatia inicial
  - Objective: Contrato HTTP para criação (SYNC/ASYNC) e consulta (GET unificado) conforme EXE-07, COM-05.
  - Likely files/components: especificação OpenAPI (a criar em local a definir na implementação).
  - Depends on: 1.1
  - Validation: contract
  - Completion criteria: contrato cobre 200/202/404/410/504 conforme EXE-02/EXE-07; GET pendente e GET final usam o mesmo schema do webhook.
  - Notes: capability associada — `execucao-e-integracoes`, `arquitetura-e-comunicacao`.

- [ ] 2.2 Especificar AsyncAPI dos eventos de fato externo e fato final
  - Objective: Envelope de evento (event_id, tenant_id, protocol_id, causation_id, versões) conforme COM-03.
  - Likely files/components: especificação AsyncAPI (a criar).
  - Depends on: 2.1
  - Validation: contract
  - Completion criteria: schema versionado; consumidor com versão desconhecida especifica quarentena, não descarte (COM-04).

- [ ] 2.3 Especificar contrato interno de despacho direto (Órbita→Cometa)
  - Objective: Payload do comando `DIRECT`/`QUEUED` (tenant_id, protocol_id, step_id, command_id, dispatch_mode, epoch, prazos) conforme COM-06.
  - Likely files/components: contrato interno HTTPS (a criar).
  - Depends on: 2.1
  - Validation: contract
  - Completion criteria: contrato nunca inclui segredo em claro; inclui identidade estável entre timeout/retry/recuperação.

## 3. Data model and persistence

- [ ] 3.1 Modelar schema mínimo de `hub_core` para Protocolo, Passo e Resultado
  - Objective: Tabelas/índices para as entidades da fatia inicial (DAD-02/03), com unicidade tenant/protocolo e tenant/chave idempotente.
  - Likely files/components: migrations de `hub_core` (a criar).
  - Depends on: 1.1
  - Validation: integration
  - Completion criteria: constraint de unicidade por tenant/chave idempotente; versão para concorrência otimista no Resultado.

- [ ] 3.2 Modelar schema mínimo de `hub_core` para Operação externa, Tentativa e Agenda de polling
  - Objective: Suportar correlação por `provider_request_id` e posse exclusiva de agenda (DAD-02, EXE-04/05).
  - Likely files/components: migrations de `hub_core` (Cometa).
  - Depends on: 3.1
  - Validation: integration
  - Completion criteria: lease/token de posse impede duplo consumo do mesmo item de agenda.

- [ ] 3.3 Modelar schema mínimo de `hub_finance` para Uso/fato econômico e Reserva de franquia
  - Objective: Suportar a regra simples de receita/custo da fatia inicial (DAD-02, FIN-04/06).
  - Likely files/components: migrations de `hub_finance` (Libra).
  - Depends on: 3.1
  - Validation: integration
  - Completion criteria: chave econômica mínima (cliente/contrato/protocolo/medidor) com dedup semântica além do event_id.

- [ ] 3.4 Modelar schema mínimo de `hub_control` para Vínculo de credencial e Placement
  - Objective: Suportar resolução SHARED_HUB/TENANT_DEDICATED e mapa tenant→célula (DAD-02, CFG-05, DAD-11).
  - Likely files/components: migrations de `hub_control` (Atlas).
  - Depends on: 1.1
  - Validation: integration
  - Completion criteria: binding dedicado exige tenant_id único; nunca há valor de segredo na entidade.

- [ ] 3.5 Implementar outbox/inbox transacional nos três domínios da fatia
  - Objective: Publicação de fatos na mesma transação do estado local, consumo com inbox antes do ack (COM-03).
  - Likely files/components: outbox/inbox por domínio (Órbita, Cometa, Libra).
  - Depends on: 3.1, 3.2, 3.3
  - Validation: integration
  - Completion criteria: crash entre commit e ack não duplica efeito nem perde o fato (QUA-02: "consumidor antes/depois do ack").

## 4. Core execution — SYNC direto e ASYNC

- [ ] 4.1 Implementar admissão durável com UUIDv7 e idempotência
  - Objective: Endpoint de criação persistindo protocolo/idempotência/pedido antes de qualquer efeito externo (EXE-01, EXE-16).
  - Likely files/components: Órbita — admissão.
  - Depends on: 3.1, 2.1
  - Validation: integration
  - Completion criteria: repetição da mesma `Idempotency-Key` recupera o mesmo protocolo mesmo após perda de resposta; payload diferente com mesma chave retorna conflito.

- [ ] 4.2 Implementar despacho DIRECT (SYNC) com posse única
  - Objective: Envio de comando idempotente ao Cometa, com watchdog durável e sem duplicação como comando QUEUED (EXE-14/15, COM-06).
  - Likely files/components: Órbita — dispatcher; Cometa — recepção de comando direto.
  - Depends on: 4.1, 2.3
  - Validation: integration
  - Completion criteria: crash após o final do Cometa e antes da resposta direta é resolvido por leitura da operação existente, sem reenvio ao provedor.

- [ ] 4.3 Implementar resolução de credencial determinística no Cometa
  - Objective: Resolver conta/binding/segredo autorizado antes de qualquer chamada externa, sem fallback implícito (SEG-05, CFG-05).
  - Likely files/components: Cometa — resolução de credencial.
  - Depends on: 3.4, 4.2
  - Validation: unit
  - Completion criteria: credencial dedicada ausente/revogada não usa SHARED_HUB nem credencial de outro tenant; vínculo é suspenso/recusado antes do envio.

- [ ] 4.4 Implementar deadline absoluto e arbitragem terminal (EXE-11)
  - Objective: Temporizador durável que encerra o protocolo ao atingir `client_deadline_at`, com transição terminal serializável única entre final e timeout.
  - Likely files/components: Órbita — temporizador de deadline.
  - Depends on: 4.1
  - Validation: integration
  - Completion criteria: cenário D−ε/D/D+ε (QUA-05) produz exatamente uma transição terminal, sem sucesso tardio publicado.

- [ ] 4.5 Implementar TTL de retry em segundos para falha transitória (EXE-10)
  - Objective: `retry_ttl_seconds` resolvido antes do aceite, com `first_transient_at` persistido uma única vez.
  - Likely files/components: Órbita/Cometa — política de retry.
  - Depends on: 4.4
  - Validation: unit
  - Completion criteria: reproduzir o exemplo numérico de EXE-10 (t=0, SLA 120s, erro em t=5, TTL 60s, reserva 5s ⇒ retry_until=65) exatamente.

- [ ] 4.6 Implementar admissão ASYNC com fila durável e worker de execução
  - Objective: Resposta 202 apenas após commit da intenção QUEUED; worker consome e executa via Cometa (EXE-02).
  - Likely files/components: Órbita — publicação em SQS; Cometa — worker.
  - Depends on: 4.1, 3.5
  - Validation: integration
  - Completion criteria: perda de resposta HTTP do 202 é recuperável pela mesma chave, sem repetir efeito.

- [ ] 4.7 Implementar concorrência entre callback e polling (EXE-05/06)
  - Objective: Consolidar observações concorrentes em uma única transição, preservando evidência duplicada.
  - Likely files/components: Cometa — consolidação de operação externa.
  - Depends on: 4.6
  - Validation: integration
  - Completion criteria: cenário "callback e polling simultâneos" (QUA-02) resulta em uma conclusão, uma entrega por destino, uma unidade de receita.

- [ ] 4.8 Implementar GET unificado de protocolo (pendente/final) e endpoint de entregas
  - Objective: Endpoint que nunca consulta o provedor; serve resultado do hub (DAD-04, EXE-07).
  - Likely files/components: Órbita — consulta.
  - Depends on: 4.1
  - Validation: contract
  - Completion criteria: instrumentar contagem de chamadas ao provedor antes/depois do GET e exigir zero novas chamadas atribuíveis (QUA-03).

## 5. Delivery — webhook

- [ ] 5.1 Implementar materialização única da representação final (COM-05)
  - Objective: Órbita materializa o corpo final uma única vez, com hash e versão; Pulsar não transforma novamente.
  - Likely files/components: Órbita — projeção final; Pulsar — entrega.
  - Depends on: 4.4
  - Validation: contract
  - Completion criteria: comparar hash/schema/media type do corpo servido por GET com o corpo de todas as tentativas de webhook do mesmo evento (QUA-03, QUA-05 "GET = webhook").

- [ ] 5.2 Implementar agenda de retry de webhook com assinatura HMAC e janela de deduplicação
  - Objective: Entrega com defaults propostos (conexão 2s, total 10s, janela 72h) e assinatura HMAC-SHA256 (EXE-08).
  - Likely files/components: Pulsar — scheduler de entrega.
  - Depends on: 5.1
  - Validation: integration
  - Completion criteria: esgotamento gera EXHAUSTED e alerta; reenvio preserva event_id e identidade semântica.

## 6. Financial — regra simples de receita/custo

- [ ] 6.1 Implementar medição econômica por chave (FIN-04)
  - Objective: Registrar receita por produto executado e custo por operação, com dedup semântica.
  - Likely files/components: Libra — medição.
  - Depends on: 3.3, 4.4
  - Validation: unit
  - Completion criteria: reproduzir o exemplo "produto em pacote" de FIN-09 (receita R$1,00; custo R$0,52; diferença R$0,48) exatamente.

- [ ] 6.2 Implementar reserva estrita para contrato com saldo estrito (FIN-06)
  - Objective: Reserva idempotente ligada ao protocolo antes de liberar qualquer passo externo.
  - Likely files/components: Libra — reserva; Órbita — chamada de reserva.
  - Depends on: 6.1
  - Validation: integration
  - Completion criteria: cenário "saldo disputado" (QUA-02) — nenhum efeito sem reserva confirmada, consumo não excede o limite.

## 7. Observability and resilience validation

- [ ] 7.1 Instrumentar métricas de baixa cardinalidade e probes por domínio
  - Objective: Métricas de OPE-06 (admissões, recusas, idade de outbox, operações UNKNOWN) e probes de OPE-05 (startup/readiness/liveness diferenciados por dependência).
  - Likely files/components: todos os domínios da fatia.
  - Depends on: 4.1–4.8
  - Validation: observability
  - Completion criteria: readiness da Órbita reflete capacidade de persistir/consultar; liveness não depende de provedor externo.

- [ ] 7.2 Qualificar operação com Redis totalmente desligado
  - Objective: Provar que SYNC/ASYNC/GET mantêm SLO qualificado sem L2 (DAD-10, QUA-06 "Redis totalmente desligado").
  - Likely files/components: caminho de leitura de resultado.
  - Depends on: 4.8, 7.1
  - Validation: performance
  - Completion criteria: pico contratado com cache L1 frio e perda de uma zona simulada não degrada abaixo do SLO qualificado.

- [ ] 7.3 Qualificar falha do writer PostgreSQL antes do aceite
  - Objective: Provar que nenhum efeito externo/aceite fictício ocorre sem autoridade durável (DAD-09, QUA-06 "writer indisponível antes do aceite").
  - Likely files/components: Órbita — admissão.
  - Depends on: 4.1, 7.1
  - Validation: integration
  - Completion criteria: cortar o writer e enviar criação resulta em erro de transporte, nunca aceite sem persistência.

## 8. Slice 2 — polling/callback concorrentes avançado e credencial dedicada

- [ ] 8.1 Implementar os três modos de polling por vínculo (EXE-05)
  - Objective: `CALLBACK_ONLY`, `POLLING_ONLY`, `CALLBACK_WITH_POLLING_FALLBACK` configuráveis por vínculo homologado.
  - Likely files/components: Cometa — configuração de polling.
  - Depends on: 4.7
  - Validation: contract
  - Completion criteria: ativação recusada se o adaptador não declarar capacidade de consulta homologada.

- [ ] 8.2 Implementar credencial TENANT_DEDICATED com isolamento de rotação
  - Objective: Vínculo dedicado por tenant, com rotação sem afetar outros vínculos saudáveis (SEG-05, CFG-05).
  - Likely files/components: Atlas — gestão de credenciais; Cometa — resolução.
  - Depends on: 4.3
  - Validation: integration
  - Completion criteria: cenário "dedicada ausente" (QUA-06) — tenant A com dedicada inválida não afeta tenant B saudável.

## 9. Slice 3 — produto composto e planos avançados

- [ ] 9.1 Implementar grafo de composição com política de falha (CAT-04)
  - Objective: Suportar até 20 passos / 5 simultâneos, com dependências causais e política de falha publicada.
  - Likely files/components: Órbita — motor de grafo.
  - Depends on: 4.4, 4.7
  - Validation: integration
  - Completion criteria: publicação com ciclo ou dependência inexistente é bloqueada antes de qualquer execução.

- [ ] 9.2 Implementar agregação com regra de merge e parcialidade (CAT-03/05)
  - Objective: Consolidar resultados independentes com precedência/deduplicação definidas; parte obrigatória ausente impede sucesso completo.
  - Likely files/components: Órbita — consolidação.
  - Depends on: 9.1
  - Validation: integration
  - Completion criteria: cenário CAT-03 (dois opcionais, um falha) resulta em estado e partes conforme a política publicada.

- [ ] 9.3 Implementar planos de cobrança avançados (faixas marginais/volume total, franquia)
  - Objective: Suportar os modelos comerciais adicionais de FIN-02 além do preço unitário simples.
  - Likely files/components: Libra — cálculo de plano.
  - Depends on: 6.1
  - Validation: unit
  - Completion criteria: reproduzir os exemplos numéricos de FIN-09 (faixas marginais 120 un. = R$116,00; volume total 120 un. = R$96,00; franquia 103 sucessos = R$52,40).

## 10. Rollout and gates

- [ ] 10.1 Preparar evidência para o gate G1 (local/dev)
  - Objective: Demonstrar testes de contrato/integração, Compose/kind, probes e telemetria da fatia inicial (QUA-04 G1).
  - Likely files/components: relatório de evidência G1.
  - Depends on: 4.1–7.3
  - Validation: manual
  - Completion criteria: emulação identificada como tal; nenhuma alegação de HA multi-AZ a partir do ambiente local.

- [ ] 10.2 Preparar evidência para o gate G2 (hom)
  - Objective: Demonstrar TTL/deadline, rejeição tardia, contratos legados/GET-webhook equivalentes, polling/callback e apuração aprovados (QUA-04 G2).
  - Likely files/components: relatório de evidência G2; sandbox de cliente/provedor homologado.
  - Depends on: 10.1, 8.1, 8.2
  - Validation: integration
  - Completion criteria: testes negativos entre tenants aprovados; acesso administrativo (`hub_protocol_reader`) auditado.

- [ ] 10.3 Registrar decisões pendentes ainda bloqueadoras antes de ppd/prd
  - Objective: Confirmar que P-01, P-02, P-05, P-07, P-08, P-10, P-11 (as que afetam a fatia/ambiente alvo) estão endereçadas ou explicitamente aceitas como risco antes de avançar para G3/G4.
  - Likely files/components: registro de decisão (fora deste pacote documental).
  - Depends on: 10.2
  - Validation: manual
  - Completion criteria: nenhuma promessa de SLO/capacidade/RPO regional feita sem a decisão correspondente resolvida (DEC-02).
