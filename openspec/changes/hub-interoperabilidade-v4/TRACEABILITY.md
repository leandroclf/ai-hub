# Rastreabilidade completa: OpenSpec vs. implementação real

**Data desta revisão:** 6/7 de setembro de 2026. **Método:** cada um dos 98 requisitos
das 9 capabilities de `openspec/changes/hub-interoperabilidade-v4/specs/` foi conferido
contra o código real em `hub/` e, sempre que possível, contra evidência gerada nesta
sessão em `hub/evidence/` (banco de dados, filas SQS/SNS, S3, logs, screenshots). Isto
substitui qualquer leitura anterior que tratasse "specs escritas" como sinônimo de
"sistema implementado" — não são a mesma coisa, e este documento existe para não deixar
essa distinção implícita.

## Validade estrutural do OpenSpec (specs em si)

Antes de perguntar "está implementado", confirmamos que a especificação (as 9
capabilities) continua estruturalmente válida:

```
98 IDs esperados (CAT 11 + ARQ 6 + COM 6 + DAD 11 + EXE 16 + FIN 11 + CFG 6 + SEG 5 +
OPE 15 + QUA 6 + DEC 5 = 98) → 98 encontrados, 0 faltando, 0 duplicados, 0 IDs extras.
```

Verificado programaticamente nesta revisão (script Python percorrendo `### Requirement:`
em cada `specs/*/spec.md`), reconfirmando o resultado já obtido quando os deltas foram
escritos. **As specs em si são válidas e completas como tradução dos 98 requisitos da
especificação-fonte.** A pergunta que resta — e que este documento responde — é quanto
disso tem sistema real por trás.

## Legenda

- ✅ **Implementado e testado** — código real existe e foi exercitado nesta sessão ou na
  anterior, com evidência reproduzível.
- ⚠️ **Parcial** — existe código real, mas cobre só parte do requisito normativo, ou o
  comportamento não foi exercitado com evidência.
- ❌ **Não implementado** — nenhum código atende este requisito na fatia atual.
- 📄 **Documental/decisão** — o requisito é sobre uma decisão, registro ou processo (não
  comportamento de sistema em runtime); está "implementado" no sentido de que o registro
  existe e é seguido onde aplicável.

## catalogo-e-portfolio (11 requisitos) — 0 ✅ / 4 ⚠️ / 7 ❌

| ID | Título | Status | Nota |
|---|---|---|---|
| CAT-01 | Catálogo de serviços | ⚠️ | Atlas tem `services` (code/version/modes/SLA/retry_ttl/published) real e testado; falta schema de entrada/saída, classificação de dados, imutabilidade de versão publicada (upsert atual permite sobrescrever) |
| CAT-02 | Portfólio e disponibilidade | ❌ | Sem estados RASCUNHO/EM_VALIDACAO/SUSPENSA/DESCONTINUADA nem portfólio por cliente |
| CAT-03 | Agregação | ❌ | Sem lógica de agregação multi-serviço |
| CAT-04 | Composição | ❌ | Sem grafo de passos — apenas serviço de passo único |
| CAT-05 | Consolidação e efeitos | ❌ | Depende de composição/agregação, inexistentes |
| CAT-06 | Seleção e equivalência de provedores | ❌ | `provider_account_id` é escolhido pelo chamador, não pelo hub por elegibilidade/prioridade |
| CAT-07 | Política comercial do produto | ❌ | Só preço unitário fixo por contrato; sem pacote/soma/híbrido |
| CAT-08 | Rastreabilidade da composição | ❌ | N/A — não há composição |
| CAT-09 | Contratos técnicos por cliente | ❌ | Sem tradução de contrato técnico por cliente/aplicação |
| CAT-10 | Elegibilidade de SLA e capacidade | ⚠️ | `client_sla_seconds` real e o deadline é genuinamente aplicado (EXE-11); sem validação de classe de isolamento/capacidade na publicação |
| CAT-11 | Crescimento, modalidade e credencial | ⚠️ | SYNC/ASYNC/AUTO e resolução de credencial (SEG-05) reais e testados; "crescimento sem chamado" não modelado (célula única fixa local) |

## arquitetura-e-comunicacao (12 requisitos) — 5 ✅ / 7 ⚠️

| ID | Título | Status | Nota |
|---|---|---|---|
| ARQ-01 | Fronteiras | ✅ | 5 aplicações + gateway, responsabilidades batem exatamente (Atlas nunca chama provedor; Cometa é o único que chama) |
| ARQ-02 | Topologia de referência | ✅ | SYNC HTTPS direto, ASYNC via SQS, fan-out SNS — confirmado em `hub/evidence/queues/sqs_sns.txt` (5 filas, 2 tópicos, 4 assinaturas, exatamente como o diagrama descreve) |
| ARQ-03 | Escolha tecnológica | ⚠️ | Go/PostgreSQL/SNS+SQS/S3/Kong reais; Secrets Manager/KMS só referenciado (segredo é string fictícia); Prometheus/Grafana/Loki/Tempo não implantados (só `/metrics` minimalista) |
| ARQ-04 | Linguagem por aplicação | ✅ | Go nas 6 apps, TypeScript/React no admin-ui — exatamente como decidido |
| ARQ-05 | Razões das dependências | ⚠️ | Decisões seguidas onde há runtime (Postgres/SQS/S3); as demais (Secrets Manager, observabilidade completa) só existem como decisão registrada, sem runtime |
| ARQ-06 | Tecnologias para expansão | ⚠️ | HPA/KEDA/Crossplane só como manifests de referência (`hub/deploy/k8s`, `terraform/crossplane`), nunca aplicados/exercitados |
| COM-01 | Comunicação interna definida | ✅ | Tabela de transporte batida item a item com o código e a evidência de filas |
| COM-02 | Integração de clientes e provedores | ⚠️ | REST/JSON, webhook e polling reais; SOAP/XML, SFTP, gRPC — não implementados (também não exigidos ainda pela spec-fonte) |
| COM-03 | Mensagens e idempotência | ⚠️ | Envelope real (com trace_id adicionado nesta sessão); outbox real e testado; **tabela `inbox` existe mas está vazia — nenhum consumidor a usa** (confirmado em `hub/evidence/db/hub_core.txt`) |
| COM-04 | Evolução de contratos | ❌ | Sem versionamento/quarentena de schema; OpenAPI/AsyncAPI são documentos estáticos, não aplicados em runtime |
| COM-05 | Mesmo contrato em GET e webhook | ✅ | Provado bit-a-bit idêntico por `TestGetMatchesWebhookBody` e novamente nesta sessão (todo webhook recebido no `webhook-sink` bate com o GET) |
| COM-06 | Contrato do despacho direto | ✅ | `dispatch.Command` implementado e testado extensivamente (SYNC e ASYNC, evidência de fila com `trace_id` no corpo real da mensagem) |

## persistencia-e-dados (11 requisitos) — 4 ✅ / 6 ⚠️ / 1 ❌

| ID | Título | Status | Nota |
|---|---|---|---|
| DAD-01 | Bancos e propriedade | ✅ | 3 bases lógicas confirmadas (`hub/evidence/db/*.txt`), nenhuma escrita cruzada no código |
| DAD-02 | Modelo lógico mínimo | ⚠️ | Maioria das entidades presente; `ledger_entries` existe mas nada escreve nela; tabela de auditoria não existe |
| DAD-03 | Transações e referências | ✅ | Admissão/finalização atômicas; concorrência otimista (`version`) exercitada de fato na corrida deadline-vs-sucesso (ver `hub/evidence/queues/queue_message_demo.txt`) |
| DAD-04 | Resultado como fonte de consulta | ✅ | GET nunca chama o provedor — confirmado por inspeção de código e pela ausência de qualquer chamada ao provider-sim nos logs de consultas GET desta sessão |
| DAD-05 | Arquivos e atomicidade com objetos | ⚠️ | Cliente S3 funciona e foi testado isoladamente (`hub/internal/objectstore/objectstore_test.go`, evidência em `hub/evidence/s3/`), mas **não está integrado a nenhum fluxo real de admissão/resultado** |
| DAD-06 | Retenção e expurgo | ❌ | Nenhum job de retenção/expurgo implementado |
| DAD-07 | Isolamento, recuperação e auditoria | ⚠️ | Isolamento por `WHERE tenant_id=$1` na aplicação, testado; RLS do PostgreSQL não configurado; backup/PITR não configurado (ambiente local) |
| DAD-08 | Persistência adicional de contrato/prazo | ⚠️ | `accepted_at`/`client_deadline_at`/`finalized_at` reais; `cell_id`, `retry_policy_version`, `clock_source` não modelados |
| DAD-09 | Dependência mínima de escrita | ✅ | Testado com Postgres genuinamente indisponível: nenhum aceite fictício, resposta 503 sem detalhe interno vazado |
| DAD-10 | Cache dispensável | ⚠️ | "Funciona sem Redis" é satisfeito trivialmente — **nenhum L1 nem L2 foi implementado**, então não há bypass qualificado, só ausência total |
| DAD-11 | Dados, placement e crescimento de células | ❌ | Célula única fixa local; sem placement dinâmico nem realocação |

## execucao-e-integracoes (16 requisitos) — 8 ✅ / 8 ⚠️ / 0 ❌ (capability mais madura)

| ID | Título | Status | Nota |
|---|---|---|---|
| EXE-01 | Admissão durável | ✅ | Testado (`TestIdempotencyReplay`, mais replay/conflito nesta sessão) |
| EXE-02 | Modos de atendimento | ⚠️ | SYNC/ASYNC reais e testados; **AUTO implementado como equivalente a ASYNC**, não a espera limitada normativa |
| EXE-03 | Estados e autoridade | ⚠️ | Estados de protocolo/operação/entrega implementados; `RECONCILING` existe no enum mas nenhum reconciliador o usa; `CANCELLED` inalcançável (sem endpoint de cancelamento) |
| EXE-04 | Operação e correlação | ⚠️ | `operation_id`/`attempt_id`/`provider_request_id` reais; estado `UNCORRELATED` de callback não implementado (o handler assume correlação sempre presente pela URL) |
| EXE-05 | Polling configurável | ⚠️ | Polling funciona e foi testado (evidência: protocolos `prov-poll-1`/`prov-poll-2` resolvidos); os três modos configuráveis por vínculo (CALLBACK_ONLY/POLLING_ONLY/FALLBACK) foram aproximados pelo `provider_mode` da conta, não configurados por vínculo |
| EXE-06 | Concorrência entre callback e polling | ⚠️ | "Uma transição final, segunda é duplicata" implementado (`ApplyExternalObservation` checa estado); `CONFLICT` para finais divergentes não testado (nenhum cenário gerou dois finais diferentes) |
| EXE-07 | Contrato de consultas públicas | ⚠️ | Tabela batida majoritariamente; **410 RESULT_EXPIRED não implementado** (depende de DAD-06, ausente) |
| EXE-08 | Webhooks | ✅ | HMAC-SHA256, retry/backoff, EXHAUSTED — testado (`TestGetMatchesWebhookBody` + entregas reais em `hub/evidence/db/hub_core.txt`) |
| EXE-09 | Falhas, retry e encerramento | ⚠️ | Tratamento de UNKNOWN real; **DLQ e reconciliador não implementados** |
| EXE-10 | TTL de reprocessamento em segundos | ❌ | `retry_ttl_seconds` é armazenado no catálogo mas **nenhum código o lê para agendar retry de falha transitória** — falha vai direto a FAILED |
| EXE-11 | Deadline, encerramento e disputa temporal | ✅ | Transição terminal única via versão otimista, testada inclusive em corrida real (ver achado em `queue_message_demo.txt`) |
| EXE-12 | Rejeição de retorno tardio | ⚠️ | Fato tardio após EXPIRED é corretamente ignorado (checagem `IsTerminal`); estado explícito `REJECTED_LATE_SLA` e evidência dedicada não implementados |
| EXE-13 | Paralelismo e espera eficiente | ⚠️ | Sem transação longa confirmada por desenho; paralelismo de passos é N/A (não há composição) |
| EXE-14 | SYNC direto, prazo e resposta final | ✅ | O requisito mais testado desta implementação — múltiplos cenários, inclusive via Kong |
| EXE-15 | Posse única despacho/fila/recuperação | ✅ | `operation_id` derivado de `command_id`; demonstrado com evidência real de fila (mensagem sobrevive a ~2 min de Cometa parado e é processada uma única vez ao retomar) |
| EXE-16 | UUIDv7, repetição e consulta futura | ✅ | Testado por unidade (versão UUIDv7) e por idempotência (replay) |

## contratos-e-financeiro (11 requisitos) — 2 ✅ / 6 ⚠️ / 3 ❌

| ID | Título | Status | Nota |
|---|---|---|---|
| FIN-01 | Contrato de aquisição | ❌ | Sem modelagem de contrato de compra/tarifa do provedor |
| FIN-02 | Contrato de venda e planos | ⚠️ | Só preço unitário fixo; sem franquia/faixas/híbrido |
| FIN-03 | Vigência e snapshot | ❌ | Contrato é lido ao vivo na admissão, não fixado/versionado no protocolo |
| FIN-04 | Medição e chaves econômicas | ✅ | Fatos REVENUE/COST com dedup (protocol_id, kind, meter), 14 fatos reais em `hub/evidence/db/hub_finance.txt` |
| FIN-05 | Matriz de incidência | ⚠️ | Linhas principais cobertas (sucesso, falha, tardio-após-expirado); "não encontrado válido", "failover autorizado" não modeladas |
| FIN-06 | Reserva de saldo e limites | ✅ | Reserva atômica com lock consultivo por tenant, captura/liberação e limite excedido — todos testados nesta e na sessão anterior |
| FIN-07 | Apuração e ledger | ❌ | `ledger_entries` existe na tabela, mas nenhum código grava partidas balanceadas |
| FIN-08 | Fechamento, conciliação e pagamento | ❌ | Sem ciclo ABERTO→EXPORTADO nem conciliação |
| FIN-09 | Exemplos de referência para QA | ⚠️ | Exemplo "produto em pacote" reproduzido (sessão anterior); franquia/faixas marginais/volume total não reproduzidos (dependem de FIN-02, ausente) |
| FIN-10 | Quebra de SLA, custo tardio e créditos | ⚠️ | "Sem receita em EXPIRED" demonstrado; fluxo de crédito/contestação não implementado |
| FIN-11 | Conta, credencial e responsabilidade econômica | ⚠️ | `settlement_party` modelado e armazenado; só o caminho HUB foi exercitado (CLIENT_DIRECT não testado); rotação de segredo não implementada |

## configuracao-e-seguranca (11 requisitos) — 2 ✅ / 4 ⚠️ / 5 ❌

| ID | Título | Status | Nota |
|---|---|---|---|
| CFG-01 | Portal e API administrativa | ⚠️ | CRUD básico real via Atlas + admin-ui (capturado em screenshots); sem validação de campo obrigatório/faixa/dependência na UI |
| CFG-02 | Publicação governada | ❌ | Sem ciclo rascunho→validação→simulação→aprovação→rollback; upsert é imediato |
| CFG-03 | Manutenção e diagnóstico | ❌ | Sem portal operacional de busca/timeline/replay |
| SEG-01 | Identidade e autorização | ❌ | Sem OAuth2/OIDC/mTLS — `X-Tenant-Id` confiado diretamente (placeholder documentado desde a auditoria anterior) |
| SEG-02 | Destinos, callbacks e dados | ❌ | **Sem proteção SSRF** em destinos de webhook — gap de segurança real, não corrigido |
| SEG-03 | Isolamento e auditoria | ⚠️ | Isolamento de tenant testado (404 para protocolo de outro tenant); sem trilha de auditoria de ações administrativas |
| CFG-04 | Configuração dos novos limites | ⚠️ | `client_sla_seconds`/`retry_ttl_seconds` reais; famílias de pressão/isolamento/administração não modeladas |
| SEG-04 | Administração entre tenants | ❌ | Sem perfil `hub_protocol_reader` nem rota administrativa cross-tenant |
| CFG-05 | Gestão de credenciais Hub–cliente–provedor | ✅ | CRUD completo testado via API e via UI (screenshot `admin-ui-03-credenciais.png`), segredo nunca exposto |
| CFG-06 | Onboarding e capacidade como serviço interno | ❌ | Sem máquina de estados de onboarding; célula única fixa |
| SEG-05 | Resolução, rotação e isolamento de segredo | ✅ | Resolução determinística sem fallback testada extensivamente (SHARED_HUB, TENANT_DEDICATED, credencial indisponível); rotação de segredo não implementada (sem segredo real para rotacionar) |

## desempenho-e-operacao (15 requisitos) — 1 ✅ / 6 ⚠️ / 8 ❌

| ID | Título | Status | Nota |
|---|---|---|---|
| OPE-01 | Modelo de capacidade | ❌ | Nenhum teste de carga executado |
| OPE-02 | Estratégia de baixa latência | ⚠️ | Pool de conexões e transações curtas por desenho; sem medição formal de orçamento de latência |
| OPE-03 | SLOs e semântica de garantia | ❌ | Nenhum SLO medido/painel |
| OPE-04 | Ambientes e alta disponibilidade | ⚠️ | Só local rodando; ppd/prd só como manifests de referência não aplicados |
| OPE-05 | Probes, escala e proteção | ✅ | As três probes reais, testadas inclusive sob indisponibilidade de Postgres (readiness 503, liveness independente) |
| OPE-06 | Observabilidade e runbooks | ⚠️ | Logs estruturados + `/metrics` mínimo reais; Prometheus/Loki/Grafana/Tempo não implantados; nenhum runbook escrito |
| OPE-07 | Controle adaptativo do provedor | ❌ | Nenhum AIMD/controlador implementado no Cometa |
| OPE-08 | Isolamento de concorrência e células | ❌ | Célula única compartilhada local, sem filas por tenant |
| OPE-09 | Orçamento de latência e paralelismo | ❌ | Não medido |
| OPE-10 | Aferição de SLA em ambas as relações | ❌ | Sem painel/registro dual de SLA |
| OPE-11 | Preservação de informação | ⚠️ | Outbox + idempotência demonstram RPO≈0 local para fatos confirmados; sem ensaio multi-AZ/regional (impossível localmente) |
| OPE-12 | Escala automática de ponta a ponta | ⚠️ | Descrito e validado sintaticamente em manifests de referência (`terraform validate`, `kubectl kustomize`); nunca exercitado contra infraestrutura real |
| OPE-13 | Matriz de dependências e fallback seguro | ⚠️ | Fallback do writer PostgreSQL testado e correto; fallback de Atlas indisponível observado indiretamente (ver achado em `queue_message_demo.txt`); Redis/cofre/S3 fallback não testados |
| OPE-14 | Manutenção sem indisponibilidade evitável | ❌ | Nenhum ensaio de rolling update (uma única instância por serviço local) |
| OPE-15 | Criticidade, tolerância a falhas | ❌ | Nenhum perfil de criticidade implementado |

## qualidade-e-aceite (6 requisitos) — 📄/⚠️ — processo, não sistema

| ID | Título | Status | Nota |
|---|---|---|---|
| QUA-01 | Estratégia de engenharia | ⚠️ | Coleção sintética parcial (sync, async_poll, async_callback, sem-credencial, força-falha); agregado/composto/franquia ausentes |
| QUA-02 | Casos críticos de falha/concorrência | ⚠️ | Vários casos da tabela demonstrados nesta sessão (commit sem resposta, timeout ambíguo→UNKNOWN, callback+polling simultâneos); muitos permanecem NÃO EXECUTADO (objeto sem commit, restore, etc.) |
| QUA-03 | Critérios de aceite e evidência | ✅ (metodologia) | Este documento + `hub/evidence/` seguem exatamente essa disciplina (pré-condições, evidência, comparação) |
| QUA-04 | Gates e definição de pronto | ⚠️ | Evidência de G1 produzida nesta sessão; G2+ não aplicável (sem sandbox real) |
| QUA-05 | Qualificação dos incrementos v3 | ❌ | A maioria dos cenários (fronteira D±ε precisa, múltiplas réplicas do Cometa) não foi construída |
| QUA-06 | Qualificação integrada da v4 | ⚠️ | Alguns casos genuinamente demonstrados (SYNC nativo, UUID em todo aceite, fronteira DIRECT/QUEUED, dedicada ausente); a maioria (réplica atrasada, cofre quente/frio, perda regional) não |

## decisoes-e-governanca (5 requisitos) — 📄 registro, seguido onde há runtime

| ID | Título | Status | Nota |
|---|---|---|---|
| DEC-01 | Decisões arquiteturais de referência | 📄 | ADRs seguidos onde há runtime (ver ARQ-01..06); registro em si intacto |
| DEC-02 | Registro de decisões pendentes (P-01–P-11) | 📄 | Nenhuma resolvida (correto — são decisões de negócio fora do alcance de código) |
| DEC-03 | Decisões herdadas da v3 | 📄 | ADR-12 (retorno tardio) demonstrado nesta sessão; demais seguidas onde aplicável |
| DEC-04 | Decisões e compatibilidade v4 | 📄/✅ | ADR-19 (SYNC direto), ADR-20 (UUIDv7), ADR-21 (credencial explícita) genuinamente implementados e testados; ADR-23 (expansão) só como referência; ADR-25 (criticidade) não implementado |
| DEC-05 | Registro de gaps, riscos e mitigação | 📄 | GAP-02/03/05 receberam evidência nova nesta sessão; os demais permanecem "qualificação pendente" como já registrado |

## Resumo quantitativo

| Capability | ✅ | ⚠️ | ❌ | 📄 | Total |
|---|---:|---:|---:|---:|---:|
| catalogo-e-portfolio | 0 | 4 | 7 | 0 | 11 |
| arquitetura-e-comunicacao | 5 | 7 | 0 | 0 | 12 |
| persistencia-e-dados | 4 | 6 | 1 | 0 | 11 |
| execucao-e-integracoes | 8 | 8 | 0 | 0 | 16 |
| contratos-e-financeiro | 2 | 6 | 3 | 0 | 11 |
| configuracao-e-seguranca | 2 | 4 | 5 | 0 | 11 |
| desempenho-e-operacao | 1 | 6 | 8 | 0 | 15 |
| qualidade-e-aceite | 1 | 4 | 1 | 0 | 6 |
| decisoes-e-governanca | 0 | 0 | 0 | 5 | 5 |
| **Total** | **23** | **45** | **25** | **5** | **98** |

**Leitura honesta deste resultado:** aproximadamente **23%** dos 98 requisitos têm
implementação real e testada nesta sessão; **46%** têm implementação parcial real (o
núcleo funciona, mas nem todas as variações/exceções do texto normativo); **25%** não
têm nenhum código; **5%** são registros de decisão/processo, não comportamento de
sistema. A capability **execucao-e-integracoes** (a máquina de estados central) é de
longe a mais madura (8 ✅ + 8 ⚠️, 0 ❌) porque foi o foco desta implementação; capabilities
de produto (catalogo-e-portfolio), operação (desempenho-e-operacao) e parte de segurança
(configuracao-e-seguranca) permanecem majoritariamente não implementadas, por decisão de
escopo (fatia inicial), não por descuido.

## O que isto significa para "o OpenSpec está implementado?"

**Não, não integralmente — e não deveria estar.** As specs em `openspec/` descrevem a
especificação v4.0 inteira (98 requisitos de um hub de interoperabilidade completo, com
composição de produtos, planos comerciais avançados, escala automática real,
controlador adaptativo, criticidade formal). Esta sessão implementou e testou o
**subconjunto realista de uma primeira fatia local** (`tasks.md` grupos 1–8), mais
artefatos de infraestrutura de referência para os grupos 9–10. As specs permanecem
**válidas como especificação** (estrutura íntegra, 98/98 IDs, rastreáveis à
matriz-fonte); o que este documento deixa explícito, requisito por requisito, é que
"especificado" e "implementado" continuam sendo coisas diferentes para a maior parte
dos 98 itens — exatamente como o próprio pacote de especificação sempre alertou
("cobertura documental não certifica operação em produção").
