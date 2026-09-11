# Evidências da retomada R2 — execução em andamento

## Atualização da rodada R4 — 10/09/2026

| Artefato | Procedimento e resultado | Limite |
|---|---|---|
| `authorized-load-latest.log` | `node hub/deploy/r2/tests/authorized-load.mjs` após `catalog-seed.mjs`: PASS; oito caminhos (`r4-authorized-1789123114890`), `SUCCEEDED`/`FAILED` esperados e mesma idempotência observada quatro vezes | Dados sintéticos, provider-sim e autoridade local |
| `browser-smoke.json` | Chromium real com OIDC PKCE, senha+OTP, editor de contrato REST declarativo, CRUD/readback, navegação, SLA, redelivery administrativo com recibo/auditoria durável, reconciliação segura sem correlação externa, destinos versionados, preparação financeira com bloqueio visual de autoaprovação, logout, leitor global cross-tenant com finalidade, `tenant_reader` sem mutações de destinos/protocolos/financeiro, HTTP 401 pós-logout e 390 px: PASS | Não substitui matriz completa de autorização nem entrega externa ponta a ponta |
| Browser Harness | Runner instalado; com `BU_CDP_URL=http://127.0.0.1:9222` percorreu 16 rotas e viewport 390×844: `PASS-EXPLORATORY` | Descoberta automática do daemon headless continua indisponível; Playwright é o gate determinístico |
| `r2-rest-adapter-latest.log` | `go test -count=1 -race -v ./internal/cometa -run 'TestAdapterRegistry\|TestVersionedRESTAdapter'`: PASS; `rest-json-v1` independente do simulador, paths declarativos `/analise` e `/consulta/{id}`, API Key por binding, SYNC/ASYNC, resposta estrita e Cometa reconstruído | Contrato local de homologação; não é provedor comercial nem qualificação produtiva |
| `r4-ctr-02-latest.log` | `go test -race -count=1 ./internal/atlas`, suíte Go/race, `go vet`, `catalog-seed` e OpenSpec strict: PASS; 21 changes válidas, subconjunto `ai-hub-json-schema-subset/v1`, igualdade matemática de números, integer exato, keywords aninhadas e limites de profundidade/nós | Rollout/rollback e cenários herdados completos permanecem abertos |
| `r4-ope-01-latest.log` | `go test -race -count=1 ./internal/providerauth`, suíte Go/race e `go vet`: PASS; L1 reutilizado com resolver indisponível, isolamento por binding/versão, espera cancelável e limite de 1.024 locks | Resolver sintético e HTTP local; não homologa cofre/AWS regional ou carga produtiva |
| `r4-ope-03-latest.log` | PostgreSQL real em schema descartável: 1.500 ofertas irrelevantes, dois candidatos elegíveis, `LIMIT 2`, ambiguidade preservável e plano com `catalog_offer_eligibility_lookup`: PASS; Atlas oficial reconstruído no Compose | Não mede carga produtiva prolongada, memória/latência sob tráfego externo ou provedor comercial |
| cache/HA/restore | Limite de locks, Redis vazio, duas réplicas kind com recuperação, restore final por digest: PASS | Kind usa dependências Compose; Redis é dispensável; sem AWS/provedor real |

## Atualização da retomada de 11/09/2026

| Artefato | Procedimento e resultado | Limite |
|---|---|---|
| `product-http-latest.log` | `R2_COMPOSE_PROJECT=ai_hub_r3qual node hub/deploy/r2/tests/product-runtime-proof.mjs`: PASS; produto com duas etapas independentes, duas operações/efeitos, GET final `SUCCEEDED` e duplicata sem novo efeito (`r4-product-http-1789118405229`) | Provider-sim local; não homologa provedor comercial ou matriz completa |
| `capacity-budget-latest.log` | Testes de budget efetivo: PASS; snapshot, contexto, lease e margem limitam a janela de I/O | Não substitui carga prolongada nem expiração durante tráfego externo |
| `browser-smoke.json` | PASS; portal persiste e relê mapeamento de entrada entre etapas e a API retorna HTTP 401 após logout sem credencial | Não substitui matriz completa de autorização e negativos |
| `hub/internal/orbita/admin.go` + `finalize.go` | UNKNOWN sem correlação externa é rejeitado/auditado; finalizador usa snapshot do protocolo quando intent histórico está ausente | Não fornece confirmação positiva quando o `provider_request_id` foi perdido |

| `authorized-load-latest.log` | Reexecução após correção de CIDR e rebuild do Cometa: oito caminhos autorizados, idempotência e falha controlada: PASS (`r4-authorized-1789118415310`) | Dados sintéticos e provider-sim local |
| `webhook-capacity-latest.log` + `finance-runtime-latest.log` | Entrega webhook com `open=0/pending=0` e financeiro balanceado sem quarentena: PASS | Não homologa endpoint comercial ou ERP |
| `rls-runtime-proof.sh` | Isolamento cross-tenant em control/core/finance com role não proprietária: PASS | Não substitui matriz HTTP completa de autorização |

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
