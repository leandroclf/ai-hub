# Índice de evidências

## Rodada integrada de 10/09/2026

O índice desta rodada está em
[`EXECUTION-2026-09-10.md`](EXECUTION-2026-09-10.md). Ele supersede as
descrições de bloqueio operacional registradas abaixo, sem superseder os
limites arquiteturais.

| Evidência | Escopo | Resultado |
|---|---|---|
| `hub/evidence/r2/execution/authorized-load-latest.log` | Carga autenticada com seed versionado | PASS; oito caminhos, falha controlada e idempotência |
| `hub/evidence/r2/execution/browser-smoke.json` | Portal Playwright | PASS; OIDC, erro de JSON inválido sem falso sucesso, CRUD/readback, editor de contrato REST declarativo, filtro local de referências, editor de produto com mapeamento entre etapas, SLA, destinos versionados, logout, HTTP 401 pós-logout e viewport móvel |
| `hub/evidence/r2/execution/browser-smoke.json` + `hub/evidence/r2/execution/r2-security.log` | R2-ADM-09-S01/S02/S03 e R2-ADM-11-S01/S03 | PASS local para redelivery auditado, reconciliação segura de `UNKNOWN`, negação de escrita ao leitor administrativo, leitura financeira persistida e bloqueio de autoaprovação; aprovação compensatória completa do ajuste permanece pendente |
| Browser Harness `scenarios/admin-console.py` | 16 rotas administrativas em CDP explícito `BU_CDP_URL=http://127.0.0.1:9222` | PASS-EXPLORATORY; descoberta automática do daemon headless ainda não funciona, evidência em `hub/evidence/r2/execution/browser-harness-latest.log` |
| `hub/evidence/r2/execution/r2-rest-adapter-latest.log` | Fronteira `ProviderAdapter` e contrato `rest-json-v1` | PASS seletivo; endpoint REST local independente, API Key por binding, SYNC/ASYNC, resposta estrita, submit/poll/reconciliação ligados ao adapter e Cometa reconstruído; provedor comercial permanece fora |
| `hub/internal/cometa/custody.go` + migração `0041` | Inbox de callback | PASS de implementação/testes focados; identidade por capability, limites de bytes/itens e retenção periódica; qualificação externa ainda aberta |
| `hub/internal/cometa/capacity.go` + executor/poller/reconciliation + `hub/internal/pulsar/worker.go` | Capacidade por domínio | PASS de integração local; concessões em `SUBMIT`/`STATUS`, reconciliação e `FETCH` de webhook; fencing validado antes da autenticação e do I/O externo; budget efetivo limitado por snapshot, contexto, lease e margem |
| `hub/internal/orbita/admission.go` + `hub/internal/pulsar/custody.go` + `custody_test.go` | Snapshot de destino webhook e orçamento de entrega | PASS de teste PostgreSQL; versão aceita é congelada, entrega usa o snapshot persistido e o permit é encerrado com sinal/evidência |
| `internal/providerauth` | Revogação, expiração e limite de locks | PASS com `go test -race -count=1` |
| `hub/internal/atlas/catalog_handlers.go` + `hub/internal/atlas/imports.go` + `hub/internal/libra/handlers.go` + `hub/internal/pulsar/handlers.go` | Fronteira administrativa Atlas/Libra/Pulsar | PASS local; rotas exigem sessão humana com MFA, papel compatível com a ação e bloqueiam workload; publicação e escritas financeiras exigem `hub_admin`, redelivery/destino exigem operador ou `hub_admin`, com regressão unitária/integrada |
| `hub/evidence/r2/execution/r2-security.log` | R2-SEG-01/02/03 | PASS local; nove cenários nomeados cobrem identidade e 401/403 sem efeito, workloads por célula, ausência de bypass, contexto sem vazamento entre requisições, leitura administrativa com MFA/justificativa e negativa sanitizada quando a auditoria está indisponível |
| `hub/evidence/r2/execution/r2-seg04-egress.log` | R2-SEG-04 | PASS local; HTTPS/TLS efetivo, bloqueio de SSRF/metadados e redirect entre origens, e whitelist privada limitada por host/porta; corrida real de rebinding DNS permanece fora desta prova |
| `hub/evidence/r2/execution/r2-seg05-admin.log` | R2-SEG-05 | PASS local; publicação atribuível em PostgreSQL, limpeza/logout no Chromium com API 401 e erros públicos correlacionados sem detalhes internos; injeção real de falha de cofre permanece fora desta prova |
| `hub/evidence/r2/execution/r2-int-exe-qualification.log` | R2-INT e R2-EXE selecionados | PASS local; credenciais dedicadas isoladas, mTLS, polling/callback, TTL/Retry-After, SLA bilateral, idempotência, fencing, inbox/outbox, AUTO e reconciliação; limites normativos não executados permanecem listados no próprio log |
| `hub/evidence/r2/execution/r2-cat-dad-qualification.log` | R2-CAT/R2-DAD selecionados | PASS local para publicação imutável, concorrência, DAG, JSON estrito, importação em staging, paginação, FileRef e pins; restore é parcial por não incluir objetos produtivos nem parceiro comercial |
| `hub/evidence/r2/execution/rls-runtime-latest.log` | R2-DAD-04-S03 | PASS runtime; `hub_runtime` sem privilégios de bypass confirmou isolamento cross-tenant em control/core/finance |
| `restore-reconciliation.sh` | Bancos control/core/finance e S3 | PASS com sufixos `r4_sequence_20260910c` e `r4_20260911qual2`; digest/contagem sem replay e colisão de alvos recusada |
| `hub/evidence/r2/execution/capacity-reconciliation-latest.log` | Reconciliação controlada das concessões locais após falhas de transporte | PASS local; quatro ausências comprovadas pelo oráculo sintético fechadas, zero efeitos presentes protegidos; não aplicável a provedor comercial |
| `hub/internal/pulsar/custody_test.go` | Entrega concorrente com capacidade `FETCH` | PASS com PostgreSQL real; duas entregas HMAC concorrentes, `transport_open=0` e `pending_external=0` |
| `hub/deploy/r2/tests/webhook-capacity-runtime.sh` + `webhook-capacity-latest.log` | Entrega no worker Pulsar real do Compose | PASS; delivery sintética chegou a `DELIVERED` e o permit `r4-webhook` terminou `open=0`, `pending=0`, com limpeza do registro de teste |
| `hub/deploy/r2/tests/finance-runtime-proof.sh` + `finance-runtime-latest.log` | Incidência financeira do workload autorizado | PASS; outbox com `SUBMITTED`/`STATUS` carregou `attempt_id`, Libra recebeu fatos de custo/receita, inbox foi persistida, journal balanceou por moeda e não houve quarentena de incidência inválida |
| `hub/internal/cometa/polling_custody.go` + `hub/internal/libra/consumers.go` | Separação entre estado operacional e incidência econômica | PASS com PostgreSQL real; aceitação UNKNOWN publica `SUBMITTED`, polling publica `STATUS` por tentativa e fencing não publica observação não autorizada |
| `hub/internal/cometa/polling_custody.go` + `polling_custody_test.go` | Fencing positivo de correlation no polling | PASS PostgreSQL real; `provider_request_id` divergente gera recibo auditável, não altera estado e não publica outbox |
| `hub/evidence/r2/execution/provider-output-custody-latest.log` | Perfil técnico, `OutputMapping` e separação de recibo bruto/resultado normalizado | PASS com PostgreSQL real; submit/poll/callback/reconciliação projetam a saída pelo snapshot e o corpo recebido é preservado em `provider_receipts` com hash |
| `hub/evidence/r2/execution/result-file-custody-latest.log` | Custódia de resultado volumoso | PASS com PostgreSQL real e LocalStack; upload streaming acima do limite inline, `ORPHAN` antes do commit, `FileRef` na representação e `LinkResult` pós-commit |
| `hub/evidence/r2/execution/r2-cat-dad-qualification.log` + `admission-postgres.log` | R2-DAD-01-S01/S02 | S01 parcial: pin de versão/hash na admissão, sem upload HTTP ponta a ponta; S02 PASS local: download e resolução de FileRef entre tenants são recusados |
| `hub/evidence/r2/execution/product-dag-latest.log` | Plano de produto, dependências, paralelismo, consolidação e compensação | PASS de integração durável em PostgreSQL e testes de concorrência; jornada HTTP local também comprovada, enquanto provedor comercial e matriz integral continuam abertos |
| `hub/evidence/r2/execution/product-http-latest.log` | Admissão HTTP de produto composto no provider-sim | PASS; duas etapas independentes produziram dois efeitos, GET finalizou `SUCCEEDED` e repetição idempotente não criou novo efeito |
| `hub/deploy/r2/tests/compose-recreation-proof.sh` + `hub/evidence/r2/execution/compose-recreation-latest.log` | R2-OPE-01-S02: recriação sem reset no Compose oficial | PASS integrado local; container Atlas foi recriado sem remover volumes, catálogo e protocolos mantiveram contagens e a jornada HTTP foi reexecutada com sucesso |
| `hub/evidence/r2/execution/product-dag-latest.log` + `product-http-latest.log` | R2-CAT-03-S01 | Parcial integrado; A/B independentes respeitam `max_parallel=2` e consolidam dois efeitos reais, mas a sobreposição temporal de traces ainda não foi medida |
| `hub/evidence/r2/execution/capacity-budget-latest.log` | Budget efetivo de I/O externo | PASS; snapshot de oferta, contexto, lease e margem limitam a janela; carga prolongada e provedor comercial continuam fora do gate local |
| `hub/internal/orbita/admission_test.go` | Isolamento do teste concorrente de admissão | PASS 20 repetições e suíte completa com o worker Orbita ativo; célula sintética não compartilhada com a execução do laboratório |
| `hub/evidence/r2/execution/kind-continuity-latest.log` | Kind independente, bootstrap offline, UI/gateway, dependências, perda controlada de Cometa/Pulsar | PASS; três nós, seis dependências cluster-owned, cinco workloads, Jobs de migração/OIDC, RTO observado de 4,398 ms/4,421 ms; volume efêmero de laboratório; recriação de nó/contêiner Kind permanece fora |
| `hub/deploy/r2/tests/kustomize-overlay-proof.sh` + `hub/evidence/r2/execution/kustomize-overlay-latest.log` | R3-OPE-02-S03: renderização dos ambientes local-kind/dev/hom/ppd/prd | PASS estrutural; cada overlay renderiza 20 recursos e cinco imagens `ai-hub-r2-<serviço>:r2`, com namespace determinístico; o manifesto `prd` não contém modo local |
| `hub/evidence/r2/execution/browser-harness-latest.log` | Browser Harness em CDP explícito | PASS-EXPLORATORY; 16 rotas e viewport 390×844, sem cloud auth ou gravação |
| Kind `ai-hub-r2` compatibilidade Compose | Prontidão, métricas, HPA/KEDA e recuperação | PASS histórico do perfil Compose-linked; o perfil independente agora é o gate de continuidade principal |

## Evidências históricas

| Evidência | Escopo |
|---|---|
| `go test ./...` | compilação e testes unitários atuais |
| `hub/internal/atlas/catalog_test.go` | null/string, string/integer, EOF, minimum, dialeto e inteiro exato |
| `hub/migrations/control/0002_provider_auth.sql` | checksum histórico R2 `37241e604376471efd7de394e9394673feaec0a32476a517c38363890dad1541` |
| `hub/deploy/r2/scripts/migrate.sh` | reconciliação restrita e falha para hash desconhecido |
| `hub/migrations/core/0036_callback_inbox_leases.sql` | lease/epoch/claim aditivo para recuperação da inbox |
| `hub/internal/cometa/worker.go` | worker autônomo, lote limitado, quarentena durável de poison item/comando expirado e redelivery preservada para falhas recuperáveis |
| `hub/migrations/control/0037_catalog_offer_lookup.sql` | índice parcial do conjunto elegível de ofertas |
| `hub/internal/atlas/offers.go` | consulta seletiva com limite de ambiguidade |
| `COMPOSE-INTEGRATED.md` | bootstrap, OIDC, upgrade, probes e testes integrados locais |
| `git diff --check` | higiene do diff |
| `docker compose ls`, `docker ps -a` | proveniência do runtime local |
| `../evidence/openspec-strict-20260910.json` | OpenSpec strict reexecutado: 21 changes, 0 falhas; a revisão deve considerar o SHA registrado no artefato após o commit |
| `hub/deploy/r2/tests/generate-openspec-inventory.py` + `INVENTORY-732-CENARIOS.csv` | Inventário derivado diretamente das 29 specs: 201 requisitos, 732 cenários, IDs sem duplicidade e digest SHA-256 das fontes |
| `hub/evidence/r2/execution/r2-security.log` + `r2-seg04-egress.log` + `r2-seg05-admin.log` | quinze cenários R2-SEG-01/02/03/04/05 nomeados; PostgreSQL, HTTPS/TLS e Chromium local nas provas correspondentes |
| `hub/deploy/r2/tests/generate-openspec-results.py` + `RESULT-MATRIX-732-CENARIOS.csv` | Matriz derivada do inventário: 732 linhas, 143 com evidência existente e 589 explicitamente `NAO_QUALIFICADO_NESTA_RODADA`; nenhum cenário sem prova é promovido |
| `OPENSPEC-AUDIT-2026-09-10.md` | auditoria sequencial das quatro changes R4 e critérios para não encerrar por inferência |
| `hub/deploy/r2/tests/browser-harness/README.md` | procedimento permanente para validação exploratória do console com browser real |
| `hub/deploy/r2/tests/browser-harness/run.sh` e `scenarios/admin-console.py` | runner e cenário somente leitura do Browser Harness; resultado é exploratório, não substitui Playwright |
| 10/09/2026: kind, readiness, métricas/HPA e recuperação de pod | laboratório `ai-hub-r2` | PASS local; dependências completas ainda não estão dentro do cluster |
| 10/09/2026: smoke Chromium autenticado | `hub/evidence/r2/execution/browser-smoke.json` | PASS: OIDC PKCE, senha/OTP, CRUD, readback, logout, API administrativa sem credencial retornando 401 e 390px |
| 10/09/2026: carga autorizada | `hub/deploy/r2/tests/authorized-load.mjs` | PASS: oito caminhos autorizados, falha controlada, polling/callback/AUTO, saldo estrito, credencial dedicada e quatro observações idempotentes |
| 11/09/2026: Browser Harness | `BU_CDP_URL=http://127.0.0.1:9222 hub/deploy/r2/tests/browser-harness/run.sh` | PASS-EXPLORATORY: 16 rotas em viewport 390×844; descoberta automática headless requer CDP explícito |

O smoke Playwright mais recente também verificou a pesquisa local de
referências no editor de ofertas (`Admin reference editor filters loaded
catalog options locally`). O filtro atua somente sobre as páginas carregadas
pelo frontend, limitadas a 1.000 registros por catálogo; a autoridade de
tenant e a consulta do backend permanecem inalteradas.

Nenhuma evidência histórica foi promovida como PASS de integração.

## Atualização da retomada de 11/09/2026

O inventário de fonte foi regenerado pela ferramenta
`hub/deploy/r2/tests/generate-openspec-inventory.py` após o seed versionado das
specs: 201 requisitos, 732 cenários, sem chaves duplicadas e com digest
`363db7fd0613b909c75beaeb9eb7dfe1e3cb52c8f4ecbbbbb697241e381dde9b`. A matriz
de resultados dos cenários ainda é mantida separadamente e não é promovida por
esse inventário.

O produto composto foi exercitado pela API pública após seed versionado do
catálogo: o probe `product-runtime-proof.mjs` observou duas operações e dois
efeitos distintos no provider-sim, finalização `SUCCEEDED` e repetição da mesma
chave sem aumento de efeitos. A reconciliação administrativa agora rejeita e
audita `UNKNOWN` sem `provider_request_id`, em vez de sugerir replay sem
correlação. O finalizador também usa o snapshot do protocolo quando o intent
histórico está ausente, preservando a confirmação durável.

As rotas administrativas Atlas, Libra e Pulsar também foram endurecidas: principal de
workload ou sessão sem MFA não alcança a administração mesmo com escopo
compatível; as escritas financeiras exigem adicionalmente o papel `hub_admin`.
No Atlas, o papel também é confrontado com a permissão da ação; publicação e
escrita financeira não são concedidas a leitores. No Pulsar, redelivery e
publicação de destino exigem `tenant_operator` ou `hub_admin`. O comportamento
foi reproduzido por testes RED→GREEN e validado na suíte Go completa, race dos
módulos críticos e `go vet`.

A matriz integral foi regenerada a partir do inventário no mesmo conteúdo de
fonte: 732 cenários, dos quais 143 possuem resultado/evidência já registrada e
589 permanecem explicitamente não qualificados. O artefato é de rastreabilidade
e não substitui a execução dos cenários restantes.

Na continuação de 11/09/2026, `r2-int-exe-qualification.log` registrou as
provas locais de credencial por binding, mTLS, polling/callback, TTL zero,
`Retry-After` além do prazo, SLA `MONITOR_ONLY`/`REJECT_LATE`, idempotência,
fencing, custódia de mensagens, AUTO e reconciliação sem replay. A matriz
associou 34 cenários R2-INT/R2-EXE a essas provas; crash/broker, provedor
comercial, partição, coorte completa e perda de recibo 2xx continuam fora da
promoção.

Esses resultados fecham os gates locais correspondentes, mas não promovem como
concluídos o adapter/provedor comercial, a qualificação regional ou a matriz
integral de requisitos e cenários.

### Atualização operacional posterior — 11/09/2026

O worker Cometa passou a receber a URL interna
`http://localstack:4566/...` mesmo quando o LocalStack devolve a autoridade
externa `sqs.*.localstack.cloud`; o comportamento é restrito a
`ENVIRONMENT=local` e tem testes unitários, mantendo URLs produtivas
inalteradas. A execução também detecta `StepDeadline` vencido antes de abrir
custódia/egresso e grava o envelope completo em `message_quarantine` antes de
remover a mensagem. Rejeições determinísticas são isoladas após quarentena;
falhas recuperáveis de capacidade/credencial continuam elegíveis a redelivery.

O backlog antigo da fila continha 15 comandos expirados e dez permits de
submissão com ausência comprovada no oráculo local. Os primeiros foram
quarentenados como `expired_command_deadline`; os permits foram fechados pelo
runner de reconciliação somente após HTTP 404 do provider-sim, resultando em
`closed=10 protected=0`. Depois disso, `product-runtime-proof.mjs` passou com a
chave `r4-product-http-1789127362271`, duas etapas/operações/efeitos e
duplicata idempotente sem novo efeito. Essa prova permanece limitada à
fixture local e não encerra o fencing geral nem a matriz integral.

Na mesma qualificação, a carga autorizada passou com o prefixo
`r4-authorized-1789127281502`; os gates financeiro, webhook, RLS, restore por
digest, continuidade Kind e smoke Playwright também passaram. O Browser
Harness percorreu as 16 rotas com sessão OIDC/OTP autenticada, viewport
390×844 e resultado `PASS-EXPLORATORY`. As correções de runtime incluíram a
renovação do upstream do `admin-ui` após rebuild da Órbita e a recriação do
Pulsar com as CIDRs internas reais; nenhum segredo foi versionado.
