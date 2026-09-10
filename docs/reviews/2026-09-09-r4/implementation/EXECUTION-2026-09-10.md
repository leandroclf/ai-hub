# Evidência de execução integrada R4 — 10/09/2026

Rodada realizada no ambiente local `ai_hub_r3qual`, com dados sintéticos,
branch `main` e o único ecossistema Compose do AI Hub ativo. Senhas, OTPs,
bearers, cookies e chaves privadas não fazem parte desta evidência.

## Gates executados

| Gate | Comando/artefato | Resultado |
|---|---|---|
| Catálogo versionado | `node hub/deploy/r2/tests/catalog-seed.mjs` | PASS; seed idempotente via `admin/v1`, validação/publicação com `If-Match`, qualificação/célula sintéticas somente no fixture autorizado |
| Carga autorizada | `R2_EVIDENCE_PREFIX=r4-authorized-20260910-final node hub/deploy/r2/tests/authorized-load.mjs` | PASS; oito caminhos e idempotência, log saneado em `hub/evidence/r2/execution/authorized-load-latest.log` |
| Callback | Carga acima + duas negativas HTTP no `/callbacks/{operation}` | PASS; callback concluiu, chave de ingresso inválida e capability inválida retornaram 401 |
| Portal Playwright | `node hub/deploy/r2/tests/browser-smoke.mjs` | PASS; OIDC PKCE, senha+OTP, criação/leitura durável, navegação, logout, storage sem sessão persistida e viewport móvel sem overflow |
| Browser Harness | `browser-harness < hub/deploy/r2/tests/browser-harness/scenarios/admin-console.py` | PASS-EXPLORATORY; 16 rotas, CDP dedicado, 390×844 aplicado via Emulation, conteúdo útil e sem overflow |
| Cache de autenticação | `go test -race -count=1 ./internal/providerauth` | PASS; isolamento por binding, revogação, expiração, coordenação cancelável e máximo de 1024 locks rastreados |
| Redis opcional | Compose profile `cache`, `redis-cli ping` e `INFO keyspace` | PASS; `PONG`, keyspace vazio; Cometa permanece sem dependência autoritativa de Redis e sem token persistido |
| Ofertas | consulta `LIMIT 2` + `EXPLAIN (ANALYZE, BUFFERS)` | PASS de caminho; índice parcial `catalog_offer_eligibility_lookup` confirmado quando o planner usa índice; catálogo pequeno pode escolher seq scan por custo |
| Restore | `R2_RESTORE_SUFFIX=r4_final_20260910 bash hub/deploy/r2/tests/restore-reconciliation.sh` | PASS; contagem/digest de control, core, finance e objetos S3 iguais; 16 efeitos observados sem replay e re-admissão mantida desabilitada |
| Kind/HA | cluster `ai-hub-r2`, três nós, réplicas e exclusão controlada de pods | PASS local; Atlas, Órbita, Cometa, Pulsar e Libra recuperaram réplicas em workers distintos; métricas, HPA e KEDA disponíveis |
| Backend | `go test -race -count=1 ./...` e `go vet ./...` | PASS |
| Frontend | `npm run build` em `hub/admin-ui` | PASS; TypeScript e Vite, 38 módulos |
| Especificações | `openspec validate --all --strict --no-interactive --json` | PASS; 21/21 changes válidas, 0 falhas |

## Detalhes observados

- A carga final produziu `SUCCEEDED` para SYNC, dois polling, callback, AUTO,
  saldo estrito e credencial dedicada; o caso de falha forçada produziu
  `FAILED` conforme esperado.
- A repetição da chave do SYNC de sucesso retornou o mesmo `protocol_id` em
  quatro observações.
- A inbox de callback terminou sem itens `RECEIVED`; os efeitos do simulador
  foram observados sem reenvio durante o restore.
- O cenário exploratório do Browser Harness continua sendo auxiliar. O gate
  bloqueador do frontend permanece o Playwright determinístico.

## Limites e pendências explícitas

Esta rodada fecha a sequência operacional local de seed, carga, callback,
portal, cache, ofertas, restore e HA do laboratório. Ela não equivale à
conclusão integral dos 201 requisitos/732 cenários nem à homologação de
provedores reais, AWS regional, multi-célula ou produção.

Ainda permanecem fora do fechamento integral, entre outros, o executor de DAG
conectado ao atendimento, capacidade adaptativa ligada a todo I/O, pools HTTP
reutilizáveis, integração completa de FileRefs/financeiro/webhooks, fencing
geral de efeito incerto, projeção de catálogo em escala, dependências próprias
do kind sem endpoints Compose e a matriz integral de requisitos/cenários.
Esses itens continuam marcados como abertos no backlog R4 e não foram
convertidos em PASS por inferência a partir desta execução local.
