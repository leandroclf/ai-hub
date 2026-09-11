# Evidências da retomada R2 — execução em andamento

## Atualização da rodada R4 — 10/09/2026

| Artefato | Procedimento e resultado | Limite |
|---|---|---|
| `kustomize-overlay-latest.log` | `hub/deploy/r2/tests/kustomize-overlay-proof.sh`: PASS estrutural; cinco overlays (`local-kind`, `dev`, `hom`, `ppd`, `prd`) renderizaram 20 recursos e cinco imagens fixadas, com namespaces determinísticos e modo local ausente em `prd` | Não aplica os manifests nem qualifica máquina limpa, recriação de nós ou dependências gerenciadas |
| `authorized-load-latest.log` | `node hub/deploy/r2/tests/authorized-load.mjs` após `catalog-seed.mjs`: PASS; oito caminhos (`r4-authorized-1789123114890`), `SUCCEEDED`/`FAILED` esperados e mesma idempotência observada quatro vezes | Dados sintéticos, provider-sim e autoridade local |
| `browser-smoke.json` | Chromium real com OIDC PKCE, senha+OTP, editor de contrato REST declarativo, CRUD/readback, navegação, SLA, redelivery administrativo com recibo/auditoria durável, reconciliação segura sem correlação externa, destinos versionados, preparação financeira com bloqueio visual de autoaprovação, logout, leitor global cross-tenant com finalidade, `tenant_reader` sem mutações de destinos/protocolos/financeiro, HTTP 401 pós-logout e 390 px: PASS | Não substitui matriz completa de autorização nem entrega externa ponta a ponta |
| Browser Harness | Runner instalado; com `BU_CDP_URL=http://127.0.0.1:9222` percorreu 16 rotas e viewport 390×844: `PASS-EXPLORATORY` | Descoberta automática do daemon headless continua indisponível; Playwright é o gate determinístico |
| `r2-rest-adapter-latest.log` | `go test -count=1 -race -v ./internal/cometa -run 'TestAdapterRegistry\|TestVersionedRESTAdapter'`: PASS; `rest-json-v1` independente do simulador, paths declarativos `/analise` e `/consulta/{id}`, API Key por binding, SYNC/ASYNC, resposta estrita e Cometa reconstruído | Contrato local de homologação; não é provedor comercial nem qualificação produtiva |
| `r4-ctr-02-latest.log` | `go test -race -count=1 ./internal/atlas`, suíte Go/race, `go vet`, `catalog-seed` e OpenSpec strict: PASS; 21 changes válidas, subconjunto `ai-hub-json-schema-subset/v1`, igualdade matemática de números, integer exato, keywords aninhadas e limites de profundidade/nós | Rollout/rollback e cenários herdados completos permanecem abertos |
| `r4-ope-01-latest.log` | `go test -race -count=1 ./internal/providerauth`, suíte Go/race e `go vet`: PASS; L1 reutilizado com resolver indisponível, isolamento por binding/versão, espera cancelável e limite de 1.024 locks | Resolver sintético e HTTP local; não homologa cofre/AWS regional ou carga produtiva |
| `r4-ope-02-latest.log` | Runner real em três bancos temporários: instalação limpa, replay idempotente, checksum desconhecido rejeitado e variante conhecida de `0002` reconciliada com auditoria: PASS | Rollback de versão publicada permanece aberto |
| `r4-ope-03-latest.log` | PostgreSQL real em schema descartável + cliente Atlas: 1.500 ofertas irrelevantes, dois candidatos elegíveis, `LIMIT 2`, plano indexado e fallback de projeção válida com recusa da vencida: PASS; Atlas oficial reconstruído no Compose | Não mede carga produtiva prolongada, memória/latência sob tráfego externo ou provedor comercial |
| `r4-cbk-latest.log` | Cometa + PostgreSQL/HTTP local + `-race`: inbox limitada, capability inválida rejeitada, callback válido aplicado, custódia antes do ACK, revogação, correlação, conflito poll/callback e worker de reconciliação: PASS | Homologação de credencial por conta no gateway e provedor comercial continuam abertas |
| cache/HA/restore | Limite de locks, Redis vazio, duas réplicas kind com recuperação, restore final por digest: PASS | Kind usa dependências Compose; Redis é dispensável; sem AWS/provedor real |

## Atualização da retomada de 11/09/2026

| Artefato | Procedimento e resultado | Limite |
|---|---|---|
| `product-http-latest.log` | `R2_COMPOSE_PROJECT=ai_hub_r3qual node hub/deploy/r2/tests/product-runtime-proof.mjs`: PASS; produto com duas etapas independentes, duas operações/efeitos, GET final `SUCCEEDED` e duplicata sem novo efeito (`r4-product-http-1789127362271`) | Provider-sim local; não homologa provedor comercial ou matriz completa |
| `compose-recreation-latest.log` + `product-http-latest.log` | R2-OPE-01-S02: recriação sem reset | PASS; Atlas foi recriado sem remoção de volumes, catálogo/protocolos preservaram contagens e a jornada HTTP voltou a concluir (`r4-product-http-1789129264735`) |
| `capacity-budget-latest.log` | Testes de budget efetivo: PASS; snapshot, contexto, lease e margem limitam a janela de I/O | Não substitui carga prolongada nem expiração durante tráfego externo |
| `browser-smoke.json` | PASS; portal persiste e relê mapeamento de entrada entre etapas e a API retorna HTTP 401 após logout sem credencial | Não substitui matriz completa de autorização e negativos |
| `hub/internal/orbita/admin.go` + `finalize.go` | UNKNOWN sem correlação externa é rejeitado/auditado; finalizador usa snapshot do protocolo quando intent histórico está ausente | Não fornece confirmação positiva quando o `provider_request_id` foi perdido |

| `authorized-load-latest.log` | Reexecução após reconciliação local protegida e correção das CIDRs: oito caminhos autorizados, idempotência e falha controlada: PASS (`r4-authorized-1789127281502`) | Dados sintéticos e provider-sim local |
| `webhook-capacity-latest.log` + `finance-runtime-latest.log` | Entrega webhook com `open=0/pending=0` e financeiro balanceado sem quarentena: PASS | Não homologa endpoint comercial ou ERP |
| `rls-runtime-proof.sh` | Isolamento cross-tenant em control/core/finance com role não proprietária: PASS | Não substitui matriz HTTP completa de autorização |

### Qualificação integrada adicional — 11/09/2026

| Artefato | Resultado observado | Limite |
|---|---|---|
| `finance-runtime-proof.sh` | `FINANCE_RUNTIME_PROOF=PASS`: outbox 58, fatos 21, custo 8, receita 13, inbox 58, journal balanceado e zero quarentena inválida | Janela local de 900s; não homologa ERP |
| `webhook-capacity-runtime.sh` | `PASS`: delivery `d38017c0-2e2f-4ed1-b46b-5796569733cf`, `open=0`, `pending=0` | Endpoint webhook sintético |
| `restore-reconciliation.sh` | `PASS`: três bancos restaurados por contagem/digest, 0 objetos S3 e oráculo externo observado sem replay | Restore local; sem AWS/provedor comercial |
| `continuity-runtime-proof.sh` | `PASS`: 3 nós, 5 workloads cluster-owned; reexecução mais recente com RTO Cometa 4398ms e Pulsar 4421ms | Ensaio Kind local; recriação de nó/contêiner Kind permanece fora |
| `browser-smoke.json` | Execução mais recente: 24 verificações, 23 `PASS` e 1 `OBSERVED`, sem `FAIL`; inclui conflito concorrente `If-Match` com diff local/servidor | Não substitui matriz integral |
| `admin-client-lifecycle-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; pendência de capacidade e suspensão com 1 protocolo `RUNNING` sintético | Fixture local; não substitui aprovação de capacidade cloud nem operação comercial |
| `admin-import-preview-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; preview `STAGED`, diff de endpoints e `IMPORTED_NOT_EXECUTABLE` sem retenção de segredo | Não publica serviço nem qualifica adapter externo |
| `admin-service-publication-smoke.json` | Ensaio Chromium curto: 3 verificações `PASS`; qualificação vigente, publicação v1 sem chamadas ao provedor e rejeição de mutação com HTTP 409 | Fixture local; não substitui homologação comercial externa |
| `admin-product-simulation-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; DAG/tabela, ausência de efeitos no provider-sim e ciclo bloqueado | Não executa chamadas faturáveis nem substitui qualificação de carga |
| `admin-integration-health-smoke.json` | Ensaio Chromium curto: 3 verificações `PASS`; binding/pagador sem segredo, rotação v2 com vigência e domínio `r4-sync` com pressão adaptativa | Consulta local; não autoriza alteração comercial de limite |
| `admin-technical-policy-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; TTL efetivo e incompatibilidade SYNC/async_poll | Contrato legado com prova GET/webhook ainda não qualificado |
| `admin-protocol-scope-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; resultado materializado sem provider e isolamento de filtro tenant-only | Não substitui ensaio comercial de todas as timelines |
| `admin-sla-freshness-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; geração, watermark, atraso observado e exportação CSV limitada/auditada | Fixture local; não homologa retenção/entrega comercial do relatório |
| `admin-finance-adjustment-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; ajuste compensatório preparado com razão, autoaprovação bloqueada e aprovação por `auditor-global` distinto com recibo `approved` | Fixture local; a auditoria financeira produtiva e o ERP comercial permanecem fora |
| `admin-usability-recovery-smoke.json` | Ensaio Chromium curto: 2 verificações `PASS`; teclado preservou JSON inválido com foco no alerta e revisão obsoleta produziu HTTP 412 recuperável com texto/diff preservados | Fixture local; não substitui auditoria WCAG formal nem rede externa degradada |
| `admin-session-scope-smoke.json` | Ensaio Chromium/API curto: `PASS`; autoridade respondeu 403 a consulta cross-tenant e UI preservou `acme` sem contexto global | Fixture OIDC local; não substitui prova regional de IdP |
| `admin-legacy-contract-smoke.json` | R2-ADM-07-S01; ensaio Chromium: `PASS`; perfil JSON/polling, oferta polling/callback e mesmo corpo final entre protocolo e webhook | Fixture OIDC, PostgreSQL e provider-sim locais; não substitui homologação comercial externa |
| Browser Harness | `PASS-EXPLORATORY`: sessão OIDC/OTP qualificada, 16 rotas, viewport 390×844, sem overflow | Exploração; não bloqueia CI |

## Atualização operacional posterior — 11/09/2026

| `identity-global-reader.json` | `python3 hub/deploy/r2/tests/identity-scope-proof.py`: `IDENTITY_SCOPE_PROOF=PASS`; `auditor-global` emite sujeito, MFA e `admin:cross_tenant`, Atlas/Órbita permitem leitura administrativa mascarada, escrita retorna 403 e `leitor-a` não cruza tenant | Fixture OIDC local, PostgreSQL e dados sintéticos; não substitui homologação produtiva |

O runner de reconciliação local fechou quatro permits `pending_external` após
confirmar HTTP 404 no provider-sim (`closed=4 protected=0`). A prova não fecha
obrigações cujo efeito externo esteja presente. O worker Cometa também passou a
usar a URL interna do LocalStack, quarentenar envelopes completos antes do ACK
e preservar rejeições recuperáveis para redelivery; detalhes e testes estão no
commit `0a57029`.

Origem: HEAD `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`, branch `r2-implementation`, alterações não commitadas. Ambiente: laboratório isolado `ai-hub-r2`, PostgreSQL16, LocalStack3.8, Keycloak26.7.3; dados sintéticos. Coleta em 07–08/09/2026. A existência de teste com nome de cenário não encerra requisito sem revisar seu oráculo e a integração.

| Artefato | Procedimento e resultado | Limite |
|---|---|---|
| admission-postgres.log | `R2_CORE_TEST_DSN=<fixture hub_core> go test -v ./internal/orbita -run TestAdmissionPostgresAtomicIdempotency -count=1`: PASS | DB real, 24 concorrentes, intent lease/takeover; não testa crash de processo ou broker |
| cometa-custody-postgres.log | `R2_CORE_TEST_DSN=<fixture hub_core> go test -v ./internal/cometa -run TestPostgresSubmissionAndObservationCustody -count=1`: PASS | DB real, posse/tentativa/resposta/outbox; não homologa adapter remoto |
| atlas-postgres.log | `ATLAS_TEST_DSN=<fixture hub_control> go test -v ./internal/atlas -count=1`: PASS | Publicação/paginação/staging em schemas descartáveis; DAG apenas validado, não executado |
| finance-postgres.log | `R2_FINANCE_DSN=<fixture hub_finance> go test -v ./internal/libra -count=1`: PASS | Fixtures de eventos no domínio; produtor de snapshot econômico ainda não integrado |
| provider-auth-unit.log | `go test -v ./internal/providerauth`: PASS nos casos Basic/OAuth sem Redis | Resolver sintético; não comprova cofre neste arquivo |
| provider-vault-localstack.log | `go test -v ./internal/providerauth -run TestAWSVaultLocalStackVersion -count=1`, AWSVault contra LocalStack: PASS | SigV4, valor decifrado e versão exata; mTLS e provedor real não testados |
| compose-bootstrap.log | `bash hub/deploy/r2/scripts/bootstrap.sh`: exit 0 | Subida local completa; não comprova recuperação/elasticidade |
| http-auth-smoke.json | HTTP em portas18080–18084: probes200, sem token401; `/admin/v1/me` e serviços200 com token real da fixture | Não substitui matriz de autorização/isolamento |
| browser-layout-investigation.log | Chromium153/Playwright1.63, OIDC senha+OTP e navegação: PASS; largura390px: FAIL | Causa: `.app-shell` herdava flex-direction row; alteração isolada para column confirmou hipótese |
| browser-smoke.json | Execução mais recente do script `hub/deploy/r2/tests/browser-smoke.mjs`; inclui preparação real de ajuste e ausência de `Aprovar` para o próprio `prepared_by` | Consultar resultado atual; HTTP 401 pós-logout foi comprovado, revogação/reautorização completa por jornada ainda não |
| go-suite-current.log / go-vet-current.log | `go test ./...` e `go vet ./...`: exit0 no estado da coleta | Testes condicionais de integração podem estar SKIP; logs específicos acima são a prova com serviços reais |

As migrations0010–0013 foram executadas via runner com checksum;0014 administrativa ainda requer aplicação. T-R2-01 continua aberto: comparar clock_timestamp antes do commit rejeita resultado já atrasado, mas NÃO prova commit estritamente anterior ao limite.

## Evidências adicionadas na retomada de 08/09/2026

| Artefato | Procedimento e resultado | Limite |
|---|---|---|
| `orbita-fact-custody-postgres.log` | PostgreSQL real, `go test -race ./internal/orbita -run TestOperationFactPostgresCustody -count=3`: PASS; persistência final indisponível não permite ACK, 12 redeliveries convergem para um final/outbox, escopos inválidos e identidade conflitante vão para quarentena | Não qualifica broker real nem deadline estrito |
| `cometa-pending-custody-postgres.log` | PostgreSQL real, custódia `PENDING` atômica: falha no orçamento de polling faz rollback; aceitação válida grava correlação, recibo, agenda, resultado e outbox | Não qualifica provedor comercial |
| `cometa-scope-null-result-failure.log` / `resume-custody-and-scope-postgres.log` | Falha inicial de leitura de resultado `NULL` registrada; correção validada com isolamento tenant/aplicação/célula e sem exposição de binding; suíte passou | Não substitui matriz completa de autorização |
| `auto-wait-postgres.log` | PostgreSQL real: AUTO rápido devolve representação persistida; AUTO lento devolve 202 após espera configurada; retry não renova espera; cancelamento conserva intenção | Não prova execução SYNC/AUTO ponta a ponta |
| `polling-fencing-postgres-v3.log` | PostgreSQL real + `-race`: 24 claims/um dono, takeover epoch, stale fenced, deadline e backoff preservados, callback/poll concorrentes, autenticação Basic e binding revogado | Não mede carga prolongada ou disponibilidade externa |
| `capacity-authority-postgres.log` | PostgreSQL real + `-race`: 60 concorrentes em 3 identidades/2 células, limites agregados, fairness por tenant, rate rolling, UNKNOWN pendente e AIMD | Integração com executor ainda não conectada; políticas são fixtures |
| `migrations-0019.log` | Runner aditivo com advisory lock/checksum após 0017/0018 | Não é prova de restore ou migração em banco histórico distinto |

Mensageria foi endurecida para descoberta somente leitura fora de `ENVIRONMENT=local`: `EnsureQueue` usa `GetQueueUrl` e `EnsureTopic` descobre ARN declarado; criação implícita permanece restrita ao laboratório. `go test ./internal/queue ./cmd/...` passou. Não há ensaio AWS remoto.

`product-build-current.log` registra a construção das imagens atuais de Atlas, Órbita, Cometa, Pulsar e Libra; `product-up-current.log` registra migrations e subida. `healthz/ready` retornou 200 em 18080–18084. `http-current-unauthenticated.log` comprova 401 nas rotas protegidas existentes; `http-current-auth-me.json` permanece 401 porque o volume Keycloak recusou a fixture até o diagnóstico de credencial persistente, portanto não é prova de login atual.

Os contratos OpenAPI/AsyncAPI foram alinhados ao principal OIDC e à identidade de aplicação/célula; headers arbitrários não são fonte de tenant.
