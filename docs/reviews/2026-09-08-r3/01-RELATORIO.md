# Revisão técnica R3 — AI Hub
**A implementação evoluiu substancialmente, mas não está pronta para ser considerada cumprimento integral da v4/R2.**
A principal lacuna é conectar mecanismos já criados às jornadas reais com prova de recuperação, autorização e resultado.

Snapshot auditado: [a4a876a9f8e8](https://github.com/leandroclf/ai-hub/commit/a4a876a9f8e875db882f7ca45cf7dece24d57aee), merge da PR #1 r2-implementation.
A análise é do conteúdo fixado, não de uma branch móvel. O relatório R2 do próprio repositório informa execução em andamento.
Não foi criada PR nesta rodada; os novos arquivos são fornecidos para inclusão pelo usuário.

## Avanços confirmados no código
| Área | Avanço | Limite |
|---|---|---|
| Identidade | JWT RS256/JWKS, validação de escopos, PKCE e identidade de workload existem. | Fronteira administrativa/aplicação e fluxo callback precisam fechamento. |
| Aceite e efeitos | Admissão grava protocolo+intenção; PrepareSubmission e recibos possuem transações. | Recuperação UNKNOWN, fencing e topologia completa ainda precisam qualificação. |
| Polling | Claim com lease/epoch, timeout, autenticação e persistência do recibo foram implementados. | Combinação callback/polling, relógios e controle de pressão não estão completos. |
| Catálogo e console | Publicação/versionamento, ofertas, contratos, navegação e páginas administrativas avançaram. | Executor DAG e várias ações/formulários não cumprem a jornada real. |
| Financeiro | Valores exatos, ledger balanceado, reservas e snapshots foram implementados. | Reserva versus efetivo, unidades de tentativa e fechamento exigem integração. |
| Entregas/objetos | Representação persistida, destinos versionados, claims, FileRefs e pins existem. | Seleção congelada de destino e arquivos no fluxo do adapter continuam incompletos. |
| Operação | Compose ampliado, overlays, probes e telemetria HTTP foram adicionados. | Kind completo, continuidade, elasticidade e sinais de domínio ainda não estão qualificados. |

## Verificações realizadas nesta revisão
- go test ./...: exit 0.
- go test -race -json ./...: exit 0; **35 testes passaram e 18 foram pulados**. Onze pacotes passaram e dezoito pacotes não tinham testes executáveis. Os 18 testes pulados não são aprovação.
- npm ci --ignore-scripts e npm run build: exit 0.
- Duas provas novas por overlay: ambas falharam como esperado e demonstram perda numérica e enum ignorado.
- Docker/PostgreSQL integrado, navegador, kind, carga, failover e restore **não foram executados nesta sessão** por ausência do runtime/dependências disponíveis. Evidências antigas não foram reclassificadas como execução nova.
- A consulta de workflows retornou lista vazia no filtro disponível; isso não demonstra ausência de todo histórico de CI. Não foi localizado diretório .github neste checkout.

## Critérios de prioridade e evidência
P0: risco de custódia, integridade financeira/dados ou fronteira de acesso que impede homologação segura.
P1: lacuna de funcionalidade, operação, escala ou qualificação necessária ao escopo.
Achados estáticos são sustentados pelo caminho citado; impacto distribuído é hipótese de falha que deve ser reproduzida.
Somente F-R3-06 possui as duas reproduções adicionais executadas aqui. Nenhum teste de carga/HA é presumido.

## Achados priorizados

### F-R3-01 — Adapters executáveis e autenticação completa (P0)

O executor recusa AdapterID diferente de synthetic-provider. Configurar endpoint não cria integração executável. Config não recebe APIKeyHeader, embora Apply o exija.

**Correção exigida:** O Hub SHALL executar adapters versionados homologados, transmitir configuração completa de autenticação e distinguir inventário importado de capacidade executável. Publicação deve rejeitar capacidades não suportadas; resultado sintético nunca substitui resposta real.

**Evidência:** [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/cometa/poller.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/poller.go), [hub/internal/atlasclient/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlasclient/client.go).

**Rastreio:** R3-EXE-01; herdados R2-INT-01, R2-INT-02, R2-ADM-04; responsável: Core e Integrações.

### F-R3-02 — Callback com confirmação de custódia (P0)

Callback chama ApplyExternalObservation sem retorno de erro e responde 200; operação desconhecida, erro de consulta e observação após terminal não têm confirmação de recibo nesse caminho. A rota está sob JWT interno do Hub, não sob autenticação específica da conta provedora.

**Correção exigida:** O Hub SHALL autenticar callback pela política da conta, limitar método/tamanho/replay e emitir 2xx somente após custódia durável. Recibos órfãos, duplicados e tardios autenticados devem ter disposição recuperável sem substituir final terminal.

**Evidência:** [hub/internal/cometa/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/handlers.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/cmd/cometa/main.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/cmd/cometa/main.go).

**Rastreio:** R3-EXE-02; herdados R2-SEG-04, R2-EXE-04, R2-EXE-05, R2-INT-04; responsável: Core e Integrações.

### F-R3-03 — Recuperação de efeito incerto e fencing (P0)

Após SUBMITTING, replay pode devolver UNKNOWN durável sem recuperador geral desse estado. Envio não transmite chave idempotente homologada. Consulta inicial de replay antecede validação completa de hash/aplicação.

**Correção exigida:** O Hub SHALL manter obrigação recuperável para SUBMITTING/UNKNOWN, validar identidade completa/hash em todo replay e verificar lease/epoch/prazo antes de I/O. Reenvio de SUBMIT exige idempotência homologada ou prova positiva de ausência de efeito; UNKNOWN exige consulta ou reconciliação.

**Evidência:** [hub/internal/cometa/custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/custody.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/orbita/intents.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/intents.go).

**Rastreio:** R3-EXE-03; herdados R2-EXE-03, R2-EXE-09, R2-INT-05; responsável: Core e Integrações.

### F-R3-04 — Relógios de retry, polling e prazo final (P0)

RetryDeadline deriva do aceite e limita polling; provider_mode mantém polling/callback exclusivos. ADR R2 registra contraexemplo de commit posterior ao deadline e requisito não qualificado.

**Correção exigida:** O Hub SHALL separar SLA do cliente, SLA do provedor, TTL desde primeira falha transitória, timeout por tentativa e horizonte de reconciliação. Polling/callback devem coexistir; espera saudável não consome TTL de indisponibilidade. Não substituir confirmação durável no prazo por checagem SQL anterior ao commit sem mudança normativa autorizada.

**Evidência:** [hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/cometa/polling_custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/polling_custody.go), [hub/internal/orbita/finalize.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/finalize.go), [docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/docs/reviews/2026-09-07-r2/implementation/ADR_T_R2_01_DEADLINE.md).

**Rastreio:** R3-EXE-04; herdados R2-EXE-06, R2-INT-04, R2-INT-05, R2-INT-08; responsável: Core e Integrações.

### F-R3-05 — Topologia de mensagens e quarentena recuperável (P0)

Bootstraps independentes não estabelecem barreira de todas as assinaturas obrigatórias antes do relay. Quarentena financeira conserva hash/motivo sem payload para replay e então confirma a mensagem.

**Correção exigida:** O Hub SHALL verificar topologia durável obrigatória antes de liberar publicação, conservar cada obrigação até handoff comprovado e guardar bytes originais limitados ou referência imutável protegida em quarentena. Hash sozinho não é custódia recuperável; reconciliação cobre fan-out e retenção.

**Evidência:** [hub/internal/queue/bootstrap.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/queue/bootstrap.go), [hub/internal/queue/queue.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/queue/queue.go), [hub/internal/libra/consumers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/consumers.go), [hub/internal/libra/store.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/store.go).

**Rastreio:** R3-EXE-05; herdados R2-EXE-05, R2-QUA-02; responsável: Core e Integrações.

### F-R3-06 — Precisão numérica e validação de schemas (P0)

Provas executadas: TransformJSON altera 9007199254740993 para 9007199254740992 e aceita INVALID contra enum [OK]. Usa float64 e validação parcial.

**Correção exigida:** O Hub SHALL preservar valores numéricos exatos, validar integralmente o dialeto declarado e rejeitar publicação de construções de schema não suportadas. Transformação deve ter limites de profundidade/tamanho/custo e não executar código ou acessar rede/segredos.

**Evidência:** [hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go), [hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go).

**Rastreio:** R3-CAT-01; herdados R2-CAT-05, R2-EXE-04; responsável: Produto e Core.

### F-R3-07 — Política efetiva e representação por cliente (P1)

Admissão usa tempos/modos do target sem composição efetiva de overrides. service_version recebido não participa da resolução. Finalizer usa FinalBody fixo sem OutputMapping.

**Correção exigida:** O Hub SHALL definir precedência serviço→oferta→perfil dentro de limites contratuais, validar versão solicitada e congelar política efetiva/hash por aplicação. Representação final transformada deve ser persistida uma vez e reutilizada com bytes idênticos em GET e corpo de webhook.

**Evidência:** [hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/orbita/finalize.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/finalize.go), [hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go).

**Rastreio:** R3-CAT-02; herdados R2-CAT-02, R2-CAT-05, R2-CAT-06, R2-EXE-08; responsável: Produto e Core.

### F-R3-08 — Agregação e composição com executor de DAG (P1)

Catálogo contém steps, mas admissão produz um comando com StepID igual ao protocolo; não há executor integrado do DAG/agregação nesse caminho.

**Correção exigida:** O Hub SHALL executar DAG versionado com dependências, mapeamentos, paralelismo limitado, parcialidade, compensações e estado durável por etapa. Sucesso do produto depende do critério contratado; compensação não é rollback automático de efeito externo.

**Evidência:** [hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/atlas/catalog.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/catalog.go), [hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go).

**Rastreio:** R3-CAT-03; herdados R2-CAT-03, R2-CAT-04, R2-ADM-05; responsável: Produto e Core.

### F-R3-09 — Resolução indexada e projeção disponível (P1)

ResolveOffer recusa tenant com mais de 100 ofertas antes de filtrar aplicação/serviço. Caminho Offer consulta Atlas em cada requisição.

**Correção exigida:** O Hub SHALL resolver ofertas por chave/vigência indexada sem teto artificial sobre o portfólio e distribuir projeções versionadas ao data plane. Snapshot válido sustenta caminho quente durante falha do controle dentro da validade/revogação definidas; cache frio não autoriza contrato desconhecido.

**Evidência:** [hub/internal/atlas/offers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/offers.go), [hub/internal/atlasclient/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlasclient/client.go), [hub/internal/atlas/catalog.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/catalog.go).

**Rastreio:** R3-CAT-04; herdados R2-CAT-06, R2-CAT-08, R2-OPE-06; responsável: Produto e Core.

### F-R3-10 — Controle adaptativo conectado a todo I/O (P1)

Capacity possui código/testes, mas Execute/requestPoll não adquirem concessão nem realimentam o controlador. Relatório R2 admite a desconexão.

**Correção exigida:** O Hub SHALL adquirir concessão global por domínio antes de SUBMIT/STATUS/FETCH, devolver métricas de resultado e adaptar concorrência com redução por erros, recuperação amortecida e Retry-After. Limites contratuais duros coexistem com adaptação e justiça entre tenants.

**Evidência:** [hub/internal/cometa/capacity.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/capacity.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/cometa/poller.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/poller.go).

**Rastreio:** R3-INT-01; herdados R2-INT-06, R2-OPE-09; responsável: Integrações e Plataforma.

### F-R3-11 — Cache de autenticação isolado e sem tokens no Redis (P1)

Bearer resolve segredo antes do L1, mantém mutex do TokenCache durante Redis/OAuth e grava bearer token no Redis, contrariando a restrição R2.

**Correção exigida:** O Hub SHALL manter segredos em cofre/cache de memória limitado com expiração/revogação, eliminar tokens do Redis e coordenar renovação por binding sem bloquear contas independentes. L1 válido deve dispensar chamada remota; falha não permite token expirado ou fallback de outro cliente.

**Evidência:** [hub/internal/providerauth/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/providerauth/client.go).

**Rastreio:** R3-INT-02; herdados R2-INT-02, R2-INT-03, R2-OPE-06; responsável: Integrações e Plataforma.

### F-R3-12 — Pools HTTP e budgets de concorrência (P1)

NewClient cria Transport por submit/poll/OAuth. MaxConnsPerHost em transports distintos não limita consumo agregado e impede reaproveitamento eficiente.

**Correção exigida:** O Hub SHALL reutilizar pools limitados por origem/identidade TLS, manter timeout por operação e drenar pools obsoletos. Controlador global limita concorrência; certificados/bindings incompatíveis não compartilham estado de autenticação.

**Evidência:** [hub/internal/platform/egress](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/egress), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/cometa/poller.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/poller.go), [hub/internal/providerauth/client.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/providerauth/client.go).

**Rastreio:** R3-INT-03; herdados R2-INT-06, R2-OPE-09; responsável: Integrações e Plataforma.

### F-R3-13 — Fronteira administrativa e aplicação (P0)

Leitura administrativa no mesmo tenant aceita protocols:read sem papel administrativo específico/MFA e lista todas as aplicações. Entre tenants já existe verificação MFA/papel; não se trata de afirmar bypass global irrestrito.

**Correção exigida:** O Hub SHALL distinguir consumidor de operador e aplicar escopo por tenant/aplicação em listagens, detalhe e payload. Leitura entre tenants exige identidade nominal autorizada, MFA, auditoria e mascaramento; desenvolvedor não usa conta compartilhada nem herda escrita financeira.

**Evidência:** [hub/internal/orbita/admin.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admin.go), [hub/internal/platform/auth/auth.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/auth/auth.go).

**Rastreio:** R3-ADM-01; herdados R2-SEG-01, R2-SEG-03, R2-ADM-01; responsável: Frontend e Segurança.

### F-R3-14 — Ações operacionais com API efetiva (P1)

Seleção usa protocol_id antes de delivery_id. Menu SLA aponta para rota não registrada; botão de reconciliação POST aponta para handler de leitura.

**Correção exigida:** O console SHALL usar contratos validados por operação, delivery_id para entregas e endpoints efetivos para reconciliação e SLA. Ação não implementada não pode simular disponibilidade. Redelivery repete bytes da entrega e não executa novamente provedor.

**Evidência:** [hub/admin-ui/src/pages/OperationsPage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/OperationsPage.tsx), [hub/internal/pulsar/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/pulsar/handlers.go), [hub/internal/orbita/admin.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admin.go).

**Rastreio:** R3-ADM-02; herdados R2-ADM-08, R2-ADM-09, R2-ADM-10; responsável: Frontend e Segurança.

### F-R3-15 — Comandos financeiros compatíveis (P1)

UI envia tenant_id/prepared_by a decoder estrito incompatível e datas YYYY-MM-DD para time.Time. Ações não mantêm tenant selecionado na query e criam idempotency key nova em cada execução.

**Correção exigida:** O console SHALL emitir contrato financeiro válido, propagar tenant autorizado, converter períodos com fuso explícito e preservar chave da intenção em retries. Autor deriva da identidade; aprovação, disputa e exportação devem completar jornadas com segregação.

**Evidência:** [hub/admin-ui/src/pages/FinancePage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/FinancePage.tsx), [hub/internal/libra/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/handlers.go).

**Rastreio:** R3-ADM-03; herdados R2-ADM-11, R2-FIN-05, R2-FIN-06; responsável: Frontend e Segurança.

### F-R3-16 — Formulários completos e tempo local (P1)

Editor de modos reduz array a seleção única; faltam campos condicionais de autenticação; lookups limitados à primeira centena. datetime-local corta UTC e o reinterpreta no fuso local ao salvar.

**Correção exigida:** O console SHALL preservar multimodalidade, oferecer campos condicionais de autenticação, lookup pesquisável/paginado e conversão correta de fuso. Respostas devem ser validadas em runtime e jornadas acessíveis por teclado com erros vinculados aos campos.

**Evidência:** [hub/admin-ui/src/pages/CatalogPage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/CatalogPage.tsx), [hub/admin-ui/src](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src).

**Rastreio:** R3-ADM-04; herdados R2-ADM-02, R2-ADM-03, R2-ADM-06, R2-ADM-07, R2-ADM-12; responsável: Frontend e Segurança.

### F-R3-17 — Reserva estrita e franquia antes do efeito (P0)

Reserva exata existe, mas captura não ajusta seu valor ao valor efetivo. DENY por franquia ocorre na incidência após execução externa.

**Correção exigida:** O Hub SHALL autorizar saldo/franquia antes do efeito, reservar exposição máxima ou obter autorização incremental prévia e conciliar captura/liberação com valor efetivo exato. UNKNOWN mantém retenção até evidência positiva; expiração do cliente não prova ausência de custo.

**Evidência:** [hub/internal/libra/store.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/store.go), [hub/internal/orbita/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/handlers.go), [hub/internal/contracts/economics/publication.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/contracts/economics/publication.go).

**Rastreio:** R3-FIN-01; herdados R2-FIN-03, R2-FIN-04; responsável: Financeiro e Core.

### F-R3-18 — Incidência completa e fechamento operacional (P1)

Fatos não transportam todas as unidades/identidades ATTEMPT; STATUS/FETCH carecem de incidência integrada. SetWatermark tem chamada em teste sem fluxo produtor de completude; ciclo de disputa/fechamento não está completo.

**Correção exigida:** O Hub SHALL medir unidades de compra/venda do snapshot, incluindo SUBMIT/STATUS/FETCH contratados, com attempt_id/evidence_id e deduplicação por unidade. Watermarks derivam de completude demonstrável; disputa, conciliação e exportação devem funcionar sem editar banco manualmente.

**Evidência:** [hub/internal/cometa/custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/custody.go), [hub/internal/cometa/polling_custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/polling_custody.go), [hub/internal/libra/store.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/store.go), [hub/internal/libra/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/libra/handlers.go).

**Rastreio:** R3-FIN-02; herdados R2-FIN-01, R2-FIN-02, R2-FIN-06; responsável: Financeiro e Core.

### F-R3-19 — Destino de webhook congelado por aplicação (P1)

Pulsar escolhe versões ACTIVE atuais por tenant no consumo, em vez de destinos/aplicação congelados na admissão.

**Correção exigida:** O Hub SHALL congelar destinos autorizados e contrato de entrega, conservar representação final imutável e distinguir reentrega de execução. Ausência de destino tem disposição explícita. Tentativas, leases e comando manual respeitam budgets auditáveis.

**Evidência:** [hub/internal/pulsar/custody.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/pulsar/custody.go), [hub/internal/orbita/finalize.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/finalize.go), [hub/internal/pulsar/handlers.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/pulsar/handlers.go).

**Rastreio:** R3-FIN-03; herdados R2-INT-07, R2-EXE-08, R2-ADM-09; responsável: Financeiro e Core.

### F-R3-20 — Objetos conectados ao fluxo e retenção (P1)

FileRefs/pins/upload existem; Execute não consome FileRefs e persistência de resultado volumoso não está conectada à execução real.

**Correção exigida:** O Hub SHALL consumir referências imutáveis por streaming limitado, conservar resultado volumoso antes da publicação e integrar retenção/pins a protocolos, entregas e finanças. Classe/região restringem processamento e acesso.

**Evidência:** [hub/internal/objectstore/catalog.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/objectstore/catalog.go), [hub/internal/objectstore/retention.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/objectstore/retention.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/orbita/admission.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admission.go).

**Rastreio:** R3-OPE-01; herdados R2-DAD-01, R2-DAD-02, R2-DAD-03; responsável: Plataforma, Dados e SRE.

### F-R3-21 — Kind completo e ambientes reprodutíveis (P1)

Kubernetes contém cinco serviços de negócio; local-kind aponta dependências a IPs de containers Compose. Overlays remotos não materializam sozinhos toda configuração/dependências.

**Correção exigida:** A entrega SHALL oferecer Compose local completo e kind completo com dependências no cluster, UI/gateway/identidade/dados/mensageria/cofre emulável/observabilidade. Serviços gerenciados em dev/hom/ppd/prd exigem IaC e referências explícitas; imagens e endpoints devem ser reproduzíveis sem IP efêmero.

**Evidência:** [hub/deploy/r2/kind/render-runtime.py](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/kind/render-runtime.py), [hub/deploy/r2/k8s/base/workloads.yaml](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/k8s/base/workloads.yaml), [hub/deploy/r2/k8s/overlays/prd/kustomization.yaml](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/k8s/overlays/prd/kustomization.yaml).

**Rastreio:** R3-OPE-02; herdados R2-OPE-01, R2-OPE-02, R2-OPE-08; responsável: Plataforma, Dados e SRE.

### F-R3-22 — Escala e continuidade com envelope (P1)

Probes/réplicas/autoscalers são avanços declarativos, ainda sem comprovação de escala de nós/dados, placement automático e drenagem integrada.

**Correção exigida:** O Hub SHALL automatizar onboarding/placement e escala de pods/nós/dados dentro de envelope aprovado, separar readiness por capacidade e drenar workers/leases. Saturação além do envelope requer admissão controlada; escala infinita e failover instantâneo não são promessas válidas.

**Evidência:** [hub/deploy/r2/k8s/base/workloads.yaml](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/k8s/base/workloads.yaml), [hub/internal/platform/httpserver](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/httpserver).

**Rastreio:** R3-OPE-03; herdados R2-OPE-03, R2-OPE-04, R2-OPE-05, R2-OPE-09; responsável: Plataforma, Dados e SRE.

### F-R3-23 — Autoridade durável, isolamento e restore (P0)

Custódia transacional avançou; RLS com papel real não proprietário e restore reconciliado continuam sem qualificação integrada. Fallback não pode transformar aceite em memória volátil.

**Correção exigida:** O Hub SHALL preservar autoridade durável replicada, papéis mínimos e isolamento tenant/aplicação/célula, e qualificar restore reconciliando mensagens/efeitos/objetos/finanças. Redis é dispensável. Falha total da autoridade exige recusa segura; fallback durável alternativo exige consistência/fencing comprovados.

**Evidência:** [hub/migrations/core](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/migrations/core), [hub/migrations/control](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/migrations/control), [hub/internal/platform/pg](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/pg), [hub/internal/orbita/admission.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admission.go).

**Rastreio:** R3-OPE-04; herdados R2-DAD-04, R2-DAD-05, R2-OPE-06; responsável: Plataforma, Dados e SRE.

### F-R3-24 — SLA bilateral e telemetria verificável (P1)

Telemetria HTTP/manifests existem, mas API de SLA não funciona no painel e alertas de obrigação exigem produtor real. Render não prova scrape/log/trace/alerta.

**Correção exigida:** O Hub SHALL emitir métricas reais de idade/custódia/lag/capacidade e SLA cliente-Hub/Hub-provedor, permitir consulta autorizada de violações e correlacionar logs/traces/protocolos sem segredos. Prometheus/Loki/Grafana devem ser testados com falhas injetadas e cardinalidade limitada.

**Evidência:** [hub/internal/platform/httpserver/telemetry.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/httpserver/telemetry.go), [hub/admin-ui/src/pages/OperationsPage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/OperationsPage.tsx), [hub/deploy/r2](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2).

**Rastreio:** R3-OPE-05; herdados R2-OPE-07, R2-INT-08, R2-ADM-10; responsável: Plataforma, Dados e SRE.

### F-R3-25 — Qualificação integral do SHA sem skips ocultos (P1)

Go/race passam nesta revisão com 18 testes SKIP. Relatório R2 admite integração incompleta e fixture OIDC bloqueada. Build da UI não detecta incompatibilidades de comandos.

**Correção exigida:** A engenharia SHALL qualificar o SHA entregue com integração obrigatória sem skips, fixture OIDC reproduzível, navegador real e oráculos externos, cobrindo toda baseline v4/R2/R3. Evidência histórica não qualifica outro SHA; tarefas só fecham com provas correspondentes e bloqueios materiais permanecem explícitos.

**Evidência:** [hub/internal/atlas/catalog_test.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/atlas/catalog_test.go), [hub/internal/cometa/custody_test.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/custody_test.go), [docs/reviews/2026-09-07-r2/implementation/FINAL_REPORT.md](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/docs/reviews/2026-09-07-r2/implementation/FINAL_REPORT.md).

**Rastreio:** R3-QUA-01; herdados R2-QUA-01, R2-QUA-02, R2-QUA-03, R2-QUA-04; responsável: Engenharia e Qualidade.

## Conclusão de engenharia
A próxima rodada deve privilegiar jornadas verticais verificadas: contrato publicado → admissão → efeito externo →
resultado durável → consulta/webhook → consumo financeiro → investigação administrativa.
Não basta acrescentar classes, tabelas, menus ou manifests desconectados.
A ausência de perda deve ser formulada como invariantes de custódia dentro de um modelo de falhas explícito,
e não como promessa absoluta de disponibilidade ou de execução exatamente uma vez em qualquer provedor.
