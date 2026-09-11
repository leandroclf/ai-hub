# Evidência de execução integrada R4 — 10/09/2026

Rodada realizada no ambiente local `ai_hub_r3qual`, com dados sintéticos,
branch `main` e o único ecossistema Compose do AI Hub ativo. Senhas, OTPs,
bearers, cookies e chaves privadas não fazem parte desta evidência.

## Gates executados

| Gate | Comando/artefato | Resultado |
|---|---|---|
| Catálogo versionado | `node hub/deploy/r2/tests/catalog-seed.mjs` | PASS; seed idempotente via `admin/v1`, validação/publicação com `If-Match`, qualificação/célula sintéticas somente no fixture autorizado |
| Carga autorizada | `R2_COMPOSE_PROJECT=ai_hub_r3qual node hub/deploy/r2/tests/authorized-load.mjs` | PASS; oito caminhos, idempotência e log saneado em `hub/evidence/r2/execution/authorized-load-latest.log` (prefixo `r4-authorized-1789092849998`) |
| Callback | Carga acima + testes focados de inbox/validação terminal | PASS parcial; callback concluiu, ingress key/capability inválidos retornaram 401, deduplicação por capability, quota e retenção têm implementação e testes focados; cenários externos completos ainda abertos |
| Portal Playwright | `node hub/deploy/r2/tests/browser-smoke.mjs` | PASS; OIDC PKCE, senha+OTP, criação/leitura durável, SLA, destinos versionados, navegação, logout, storage sem sessão persistida e viewport móvel sem overflow |
| Entrega webhook com capacidade | `bash hub/deploy/r2/tests/webhook-capacity-runtime.sh` | PASS; worker Pulsar real entregou no sink local e encerrou o permit `r4-webhook` sem concessão aberta ou obrigação pendente; fixture removida ao final |
| Incidência financeira integrada | `bash hub/deploy/r2/tests/finance-runtime-proof.sh` após carga autorizada | PASS; `SUBMITTED`/`STATUS` com `attempt_id`, fatos de custo/receita, inbox persistida, journal balanceado por moeda e zero quarentena de incidência inválida; log em `hub/evidence/r2/execution/finance-runtime-latest.log` |
| Browser Harness | `browser-harness < hub/deploy/r2/tests/browser-harness/scenarios/admin-console.py` | BLOCKED-ENVIRONMENT; executável instalado, mas o daemon local não encontrou `DevToolsActivePort`/CDP utilizável |
| Cache de autenticação | `go test -race -count=1 ./internal/providerauth` | PASS; isolamento por binding, revogação, expiração, coordenação cancelável e máximo de 1024 locks rastreados |
| Redis opcional | Compose profile `cache`, `redis-cli ping` e `INFO keyspace` | PASS; `PONG`, keyspace vazio; Cometa permanece sem dependência autoritativa de Redis e sem token persistido |
| Ofertas | consulta `LIMIT 2` + `EXPLAIN (ANALYZE, BUFFERS)` | PASS de caminho; índice parcial `catalog_offer_eligibility_lookup` confirmado quando o planner usa índice; catálogo pequeno pode escolher seq scan por custo |
| Restore | `R2_COMPOSE_PROJECT=ai_hub_r3qual R2_RESTORE_SUFFIX=r4_20260911qual2 bash hub/deploy/r2/tests/restore-reconciliation.sh` | PASS; contagem/digest de control, core, finance e objetos S3 iguais; efeitos observados sem replay e re-admissão mantida desabilitada; alvos já existentes agora são recusados antes do restore |
| Kind/HA | `hub/deploy/r2/kind/bootstrap-independent.sh` + `continuity-runtime-proof.sh` | PASS local independente; PostgreSQL, LocalStack, Keycloak, Alloy, provider-sim, webhook-sink, observabilidade, Kong e UI materializados no cluster; cinco workloads em duas réplicas; Cometa/Pulsar recuperados após exclusão controlada com RTO observado de 4,453 ms/4,481 ms |
| Produto/DAG durável | `R2_CORE_TEST_DSN=... go test -race -count=1 ./internal/orbita -run 'TestProductPlan'` | PASS; etapas e dependências persistidas, limite de paralelismo, consolidação por fatos e compensação separada; evidência em `hub/evidence/r2/execution/product-dag-latest.log` |
| Resultado volumoso/FileRef | `go test -count=1 ./internal/orbita -run TestFinalizeMaterializesLargeResultAndLinksAfterCommit` + `go test -tags=e2e ... ./internal/objectstore -run TestPostgresS3MultipartFileRef` | PASS; resultado acima de 1 MiB usa upload streaming, fica `ORPHAN` até o protocolo confirmar e é exposto como `FileRef`; LocalStack real comprovou multipart, verificação/download e retenção |
| Backend | `go test -race -count=1 ./...` e `go vet ./...` | PASS; entrega Pulsar com domínio `r4-webhook` e lease seguro também coberta por teste PostgreSQL |
| Frontend | `npm run build` em `hub/admin-ui` | PASS; TypeScript e Vite, 41 módulos |
| Especificações | `openspec validate --all --strict --no-interactive --json` | PASS; 21/21 changes válidas, 0 falhas |

## Detalhes observados

- A carga final produziu `SUCCEEDED` para SYNC, dois polling, callback, AUTO,
  saldo estrito e credencial dedicada; o caso de falha forçada produziu
  `FAILED` conforme esperado.
- A repetição da chave do SYNC de sucesso retornou o mesmo `protocol_id` em
  quatro observações.
- A inbox de callback terminou sem itens `RECEIVED`; os efeitos do simulador
  foram observados sem reenvio durante o restore.
- O ingresso de callback rejeitou observações sem `provider_request_id` ou não
  terminais antes de criar custódia órfã; a inbox usa `(operation_id,
  body_sha256, token_hash)`, limite de bytes/itens e limpeza limitada de itens
  processados.
- A capacidade está ligada às concessões de `SUBMIT`, `STATUS` de polling,
  reconciliação e `FETCH` de webhook. Cada caminho valida o fencing antes da
  autenticação e imediatamente antes do I/O externo; submissão conserva
  `UNKNOWN` sem reenviar quando a lease cai, enquanto observadores e webhook
  liberam o permit para retomada segura. O Pulsar falha fechado quando o lease
  não cobre o timeout mais a margem de segurança; respostas/timeout fecham-no
  com sinal, latência e evidência. O runtime Compose instala o domínio
  `r4-webhook` sem alterar os volumes existentes.
- Resultado acima de 1 MiB não é colocado no JSONB do protocolo: o finalizador
  faz upload streaming pela autoridade de objetos, conserva a referência como
  `ORPHAN` antes do commit e executa `LinkResult` depois da representação final.
  Falha de vínculo permanece uma obrigação reconciliável, sem desfazer o
  protocolo já confirmado.
- O aceite congelou destinos de webhook por tenant/aplicação e o Pulsar
  consumiu somente o snapshot persistido no fato, sem consultar uma versão
  dinâmica posterior.
- A incidência econômica foi separada do estado operacional: a aceitação
  externa permanece `UNKNOWN` para a Órbita, mas publica `economic_kind=SUBMITTED`;
  cada polling publica `economic_kind=STATUS` com `attempt_id`. O oráculo
  financeiro confirmou fatos de custo/receita e journal balanceado no mesmo
  workload.
- O cenário exploratório do Browser Harness continua auxiliar e bloqueado por
  CDP nesta máquina; o gate bloqueador do frontend é o Playwright determinístico.

## Limites e pendências explícitas

Esta rodada fecha a sequência operacional local de seed, carga, callback,
portal, cache, ofertas, restore e HA do laboratório. Ela não equivale à
conclusão integral dos 201 requisitos/732 cenários nem à homologação de
provedores reais, AWS regional, multi-célula ou produção.

A execução durável do DAG agora está conectada à admissão de produtos e tem
evidência própria; a jornada HTTP com catálogo/provedor real ainda precisa de
qualificação dedicada. Permanecem fora do fechamento integral, entre outros,
budgets integrais de I/O, qualificação de adapter/provedor real, vinculação
produtiva de financeiro/webhooks, fencing de efeitos cujo provider_request_id
já foi perdido, projeção de catálogo em escala e a matriz integral de
requisitos/cenários. O perfil Kind independente fecha o gate local de
dependências e endpoints; não substitui IaC/HA regional dos ambientes remotos.
Esses itens continuam marcados como abertos no backlog R4 e não foram
convertidos em PASS por inferência a partir desta execução local.
