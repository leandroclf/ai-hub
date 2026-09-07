# Auditoria da implementação — Hub de Interoperabilidade "Constelação"

**Data:** 6/7 de setembro de 2026. **Escopo:** implementação de código real (não apenas
documentação) da fatia inicial definida em
`openspec/changes/hub-interoperabilidade-v4/tasks.md`, a partir das specs em
`openspec/changes/hub-interoperabilidade-v4/specs/*/spec.md` e do `design.md` da mesma
mudança. Código em `hub/`.

## Como ler este documento

Cada item é marcado como:
- ✅ **Implementado e testado** — código real, executado contra a stack local (Docker
  Compose), com evidência de teste automatizado e/ou manual reproduzível.
- ⚠️ **Implementado parcialmente / simplificado** — existe código funcional, mas com
  uma simplificação deliberada em relação ao texto normativo, documentada explicitamente.
- ❌ **Não implementado** — não existe código para este item nesta entrega.

Nada aqui foi "marcado como pronto" sem verificação: todo item ✅ tem um comando ou
teste reproduzível citado. Consistente com QUA-03 ("um caso só passa com pré-condições,
massa... e comparação com esperado") e com o registro **NÃO EXECUTADO** que a própria
especificação-fonte usa para tudo que não tem evidência.

## Resumo executivo

Esta entrega **não é** uma implementação completa dos 98 requisitos da especificação
v4.0. Isso seria irrealista sem as decisões de negócio ainda pendentes (P-01 a P-11,
ver `openspec/changes/hub-interoperabilidade-v4/proposal.md`) — conta/região AWS,
tarifas comerciais reais, política de retenção, perfil de criticidade, etc. — e sem
meses de engenharia de um time real.

O que **foi** entregue: uma implementação real, executável e testada da **fatia inicial**
recomendada em QUA-04 e detalhada em `tasks.md` grupos 1–8 — as seis aplicações do
domínio (Portal/Kong, Atlas, Órbita, Cometa, Pulsar, Libra) rodando localmente via Docker
Compose, com PostgreSQL (três bases lógicas), LocalStack emulando SQS/SNS/S3, um
provedor externo simulado com três modalidades (síncrono, polling, callback), e os
invariantes mais críticos da especificação verificados com testes automatizados que
passam contra a stack real, não apenas descritos.

Ao final: **44 arquivos Go, ~4.750 linhas**, 3 migrations SQL (248 linhas), 6 testes de
integração ponta a ponta (todos passando), 2 testes unitários, contratos OpenAPI/AsyncAPI
formais, e artefatos de infraestrutura de referência (Kubernetes, Terraform/Crossplane)
não aplicados a nenhum ambiente real.

## Cobertura por grupo de `tasks.md`

| Grupo | Status | Evidência |
|---|---|---|
| 1. Discovery and setup | ✅ | `hub/deploy/docker-compose.yml` sobe Postgres+LocalStack+Kong; `hub/internal/providersim` é o simulador determinístico com falha injetável (`force_fail`, `delay_ms`) |
| 2. Contract and specs | ✅ | `hub/api/openapi.yaml`, `hub/api/openapi-internal.yaml`, `hub/api/asyncapi.yaml` — validados sintaticamente (`yaml.safe_load`) |
| 3. Data model and persistence | ⚠️ | Migrations completas (3.1–3.4) e outbox transacional (3.5) implementados e testados; **tabela `inbox` existe mas nenhum consumidor a usa** para dedup de `event_id` — dedup real hoje depende de `command_id`/`protocol_id`/chaves únicas específicas de cada domínio, não do padrão inbox genérico descrito em COM-03 |
| 4. Core execution (SYNC/ASYNC) | ✅ | Todas as 8 tarefas (admissão, despacho DIRECT, credencial, deadline, TTL, ASYNC, concorrência callback/polling, GET) implementadas e cobertas por `hub/test/e2e/e2e_test.go` + testes manuais documentados abaixo |
| 5. Delivery (webhook) | ✅ | `hub/internal/pulsar` — materialização única + HMAC-SHA256 + retry; `TestGetMatchesWebhookBody` prova que GET e webhook devolvem bytes idênticos |
| 6. Financial | ✅ | `hub/internal/libra` — medição por chave econômica e reserva estrita com lock consultivo por tenant; `TestStrictBalanceLimitExceeded` |
| 7. Observability and resilience | ⚠️ | Probes diferenciadas (startup/ready/live) e `/metrics` minimalista implementados (7.1); Redis nunca foi implementado — "funciona sem Redis" é satisfeito trivialmente por ausência total de L1/L2, não pela qualificação de um cache real desligado (7.2); indisponibilidade do writer testada manualmente duas vezes, não automatizada em teste de regressão (7.3) |
| 8. Slice 2 (polling avançado, credencial dedicada) | ⚠️ | Credencial `TENANT_DEDICATED` testada e funcionando (8.2); os três modos explícitos de polling por vínculo de EXE-05 (`CALLBACK_ONLY`/`POLLING_ONLY`/`CALLBACK_WITH_POLLING_FALLBACK`) foram aproximados pelo `provider_mode` da conta, não implementados como configuração por vínculo (8.1) |
| 9. Slice 3 (produto composto, planos avançados) | ❌ | Grafo de composição (CAT-04), agregação com merge (CAT-03) e planos de faixas/franquia (FIN-02) **não implementados** — apenas serviço de passo único e preço unitário |
| 10. Rollout and gates | ⚠️ | Este documento + os testes e2e constituem evidência de G1 (local/dev); G2+ exige sandbox real de cliente/provedor e decisões P-01/P-02/etc., fora do alcance desta entrega |

## O que foi verificado (testes reais, não apenas descrição)

### Testes automatizados

```
cd hub
go test ./...                       # 1 pacote com teste unitário (idgen: UUIDv7 válido e único)
go test -tags e2e ./test/e2e/...    # 6 testes de integração — todos PASS
```

Testes e2e (`hub/test/e2e/e2e_test.go`), cada um mapeado ao requisito que verifica:

| Teste | Requisito | Resultado |
|---|---|---|
| `TestSyncSuccess` | EXE-14 (SYNC direto, sucesso na mesma conexão) | PASS |
| `TestSyncFailure` | EXE-02/EXE-03 (falha de negócio representada no contrato) | PASS |
| `TestIdempotencyReplay` | EXE-01 (mesma chave/hash recupera protocolo; hash diferente é conflito) | PASS |
| `TestGetMatchesWebhookBody` | COM-05 (mesmo corpo em GET e em toda tentativa de webhook) | PASS |
| `TestCredentialUnavailableNoFallback` | SEG-05 (sem vínculo elegível, falha sem fallback implícito) | PASS |
| `TestStrictBalanceLimitExceeded` | FIN-06 (reserva estrita respeita teto do tenant) | PASS |

### Testes manuais reproduzidos e documentados nesta sessão (fora do pacote e2e)

- **ASYNC via polling e via callback**: dois `provider_mode` distintos (`async_poll`,
  `async_callback`) admitidos, aceitos com 202, finalizados corretamente como
  `SUCCEEDED` após o fato assíncrono ser observado — confirmado por `GET` e por entrega
  de webhook recebida no `webhook-sink`.
- **AUTO**: aceito com 202 — implementado como equivalente a ASYNC (ver limitações).
- **TENANT_DEDICATED com sucesso**: contrato exigindo credencial dedicada, vínculo
  cadastrado, execução SYNC concluída com sucesso.
- **Indisponibilidade do writer PostgreSQL** (`docker compose stop postgres`): nova
  admissão retorna **503 `admission_unavailable`** sem aceite fictício; readiness da
  Órbita reporta 503; após restaurar o Postgres, o sistema volta a operar sem
  intervenção manual e sem efeito duplicado.
- **Kong (Portal/gateway)**: requisição de admissão SYNC roteada por
  `http://localhost:8000/v1/protocols` (porta do Kong) concluída com sucesso,
  confirmando que o papel de gateway (ARQ-01) está de fato no caminho.
- **Fatos econômicos**: `SELECT * FROM economic_facts` confirma um registro `REVENUE`
  e um `COST` por protocolo bem-sucedido, sem duplicação em GETs repetidos.

## Bugs encontrados e corrigidos durante esta auditoria

Nenhum destes bugs existia na especificação — foram introduzidos durante a
implementação e corrigidos após serem detectados pelos próprios testes/ensaios
manuais desta sessão:

1. **Structs do Atlas sem `json` tags** (`Service`, `ProviderAccount`,
   `CredentialBinding`, `Contract`) faziam `encoding/json` ignorar campos como
   `client_sla_seconds`, silenciosamente gravando `0`. Corrigido adicionando tags
   explícitas a todos os campos. Efeito antes da correção: deadlines calculados como
   "agora", expirando protocolos instantaneamente.
2. **`pulsar.Destination` sem `json` tags** — mesmo problema, causava
   `tenant_id`/`hmac_secret` vazios ao cadastrar destino de webhook.
3. **Scan de coluna `final_body` (JSONB nulo) em `json.RawMessage`** causava erro
   `unsupported Scan` sempre que um protocolo ainda não finalizado era lido. Corrigido
   escaneando para `[]byte` primeiro.
4. **`SELECT ... FOR UPDATE` sobre uma agregação (`SUM`)** no Libra — proibido pelo
   PostgreSQL, causava 500 em toda reserva financeira. Corrigido com
   `pg_advisory_xact_lock(hashtext(tenant_id))` para serializar reservas concorrentes
   do mesmo tenant sem depender de uma linha física para lockar.
5. **Vazamento de erro interno de infraestrutura ao cliente**: quando o Postgres caía
   durante a checagem de idempotência (antes mesmo da tentativa de admissão), a Órbita
   respondia 500 com a mensagem literal do driver (`dial tcp: lookup postgres...`).
   Corrigido para responder 503 `admission_unavailable` sem detalhe interno, com o erro
   completo apenas no log do servidor.

Script de init do Postgres também precisou de correção (conectar a `-d postgres` em vez
do banco default do usuário, que não existe até a primeira criação).

## Simplificações e placeholders assumidos (por decisão explícita, não por descuido)

| Item | O que foi simplificado | Por quê / o que falta para produção |
|---|---|---|
| Autenticação (SEG-01) | `X-Tenant-Id` aceito diretamente pela Órbita, sem OAuth2/OIDC/mTLS real | Requer IdP real e configuração de Kong (plugin `openid-connect`/`key-auth`), fora do alcance local |
| Credenciais/segredos (CFG-05, SEG-05) | `secret_ref` é uma string fictícia (`vault://...`); nenhuma integração com Secrets Manager/KMS | Depende de conta AWS real (P-01) |
| SSRF em destinos de webhook (SEG-02) | **Nenhuma validação** de destino é aplicada ao cadastrar `webhook_destinations` ou ao executar a entrega | Gap de segurança real, não apenas simplificação — deve ser implementado antes de aceitar URLs fornecidas por clientes reais |
| Cache (DAD-10) | Nenhum L1 nem L2 implementado; toda leitura vai direto ao PostgreSQL | Satisfaz trivialmente "funciona sem Redis", mas não qualifica o comportamento de bypass/degradação de um cache real desligado |
| Polling por vínculo (EXE-05) | Três modos aproximados por `provider_mode` da conta, não configuráveis por vínculo cliente-provedor | Exigiria campo adicional em `credential_bindings` ou tabela própria de política de polling |
| AUTO (EXE-02) | Tratado como equivalente a ASYNC (202 imediato) | EXE-02 exige espera limitada com possível 200 síncrono; não implementado |
| Composição/agregação (CAT-03/04) | Não implementado — apenas serviço de passo único | Escopo explicitamente adiado para a "slice 3" em QUA-04 |
| Planos comerciais avançados (FIN-02) | Apenas preço unitário fixo por contrato | Faixas marginais/volume total e franquia não implementados |
| Inbox (COM-03) | Tabela existe, não é usada para dedup genérica | Dedup hoje depende de chaves únicas específicas por domínio (idempotency_key, command_id, event_id+meter) |
| RLS (DAD-07) | Isolamento de tenant apenas por `WHERE tenant_id = $1` na aplicação | Nenhuma política RLS do PostgreSQL foi criada nesta referência |
| Observabilidade (OPE-06) | `/metrics` é um registry próprio minimalista (texto compatível com Prometheus), sem OpenTelemetry/Tempo/Loki reais | Suficiente para demonstrar o contrato de probes; não há stack de observabilidade real no compose |
| Controle adaptativo (OPE-07) | Não implementado — nenhum AIMD, sem coordenação distribuída de quota | Fora do escopo desta fatia inicial |

## Gaps críticos do DEC-05 — status nesta implementação

Nenhum destes gaps foi encerrado por este código; onde há evidência nova, ela está
anotada:

- **GAP-02** (duplicação entre despacho direto e fila): mitigado na implementação —
  `operation_id` deriva deterministicamente de `command_id`, testado via idempotência
  de comando reenviado (código revisado; não há teste automatizado de reentrega de
  mensagem SQS especificamente).
- **GAP-03** (credencial de outro cliente/conta compartilhada indevida): testado e
  confirmado — `TestCredentialUnavailableNoFallback`.
- **GAP-05** (custódia sem prova): testado manualmente — writer indisponível não gera
  aceite fictício.
- **GAP-09, GAP-10, GAP-11, GAP-12, GAP-13, GAP-15**: **não endereçados** por esta
  implementação — exigem, respectivamente, realocação de célula, ensaio de fronteira
  temporal D±ε automatizado, perfil de criticidade real, topologia regional, contrato
  de recuperação com provedor real, e qualificação formal — nenhum desses existe nesta
  fatia local.

## Infraestrutura de referência (não aplicada a nenhum ambiente real)

Nenhum `kubectl apply`, `terraform apply`/`plan` ou provisionamento real foi executado
contra qualquer cluster ou conta AWS — não há credenciais nem conta/cluster aprovados
(P-01). Tudo abaixo foi apenas **validado sintática/estruturalmente**, de forma
totalmente offline/local.

### Kubernetes (`hub/deploy/k8s/`)

Um manifest por serviço de negócio (`atlas.yaml`, `orbita.yaml`, `cometa.yaml`,
`pulsar.yaml`, `libra.yaml`) mais `kustomization.yaml` e `README.md`. Cada um traz
`ConfigMap` (config não sensível) + `Secret` (DSNs com placeholder de senha),
`Deployment` com 3 réplicas e `topologySpreadConstraints` por zona/nó (OPE-04),
`startupProbe`/`readinessProbe`/`livenessProbe` apontando para os endpoints reais já
implementados (`/healthz/startup|ready|live`), `resources.requests/limits` marcados como
placeholder pendente de P-02, `Service` ClusterIP e `PodDisruptionBudget`. Atlas/Órbita
recebem `HorizontalPodAutoscaler` (API); Cometa/Libra têm comentário explícito de que
usariam KEDA em vez de HPA (OPE-12); Pulsar inclui um `ScaledObject` KEDA de exemplo.
Todas as env vars espelham exatamente os nomes usados no `docker-compose.yml`.

Validação: `kubectl kustomize .` (renderização local, sem contato com API server) e
parsing YAML de cada arquivo — ambos sem erro, 28 recursos renderizados. O `kubectl`
local aponta para um cluster EKS real da organização com credenciais SSO expiradas;
nenhuma tentativa de contato com esse cluster foi feita ou teve sucesso.

### Terraform / Crossplane (`hub/deploy/terraform/`)

VPC multi-AZ (3 zonas), EKS multi-AZ, RDS PostgreSQL separado por base lógica (um
`hub_control` global + `hub_core`/`hub_finance` por célula, multi-AZ, PITR de 35 dias,
criptografado por KMS, com `prevent_destroy`), bucket S3 versionado com KMS e bloqueio de
acesso público, Secrets Manager (apenas referências — valores via `random_password`,
nunca hardcoded) e a topologia **exata** de filas/tópicos SQS/SNS já usada pelo código
(`cometa-commands`, `hub-operation-facts`, `orbita-operation-facts`,
`libra-cost-facts`, `hub-protocol-facts`, `pulsar-protocol-facts`,
`libra-revenue-facts`), com políticas de fila restritas a `aws:SourceArn` — mais
restritivas do que o atalho `Principal: "*"` usado apenas localmente contra o
LocalStack em `hub/internal/queue/queue.go` (`AllowSNSDelivery`), diferença documentada
explicitamente no README do módulo. Um módulo Crossplane de referência
(`crossplane/xrd.yaml` + `composition.yaml` + `claim-example.yaml`) demonstra como uma
nova célula seria provisionada declarativamente (ARQ-06, CFG-06, OPE-12).

Validação **executada de fato e reproduzida de forma independente nesta auditoria**:
`terraform fmt -check` (limpo), `terraform init -backend=false` (baixa os providers reais
`hashicorp/aws` 5.100.0, `hashicorp/random` 3.9.0, `hashicorp/tls` 4.4.0 do registry
público, sem credenciais AWS) e `terraform validate` → **`Success! The configuration is
valid.`**, sem warnings. Nenhum `terraform plan`/`apply` foi executado. Os arquivos
Crossplane foram validados apenas como YAML sintaticamente correto — não contra CRDs
reais de um cluster com Crossplane instalado (nenhum cluster disponível).

### Interface administrativa (`hub/admin-ui/`)

Scaffold real (não estático) em TypeScript + React + Vite (ARQ-04), com cliente HTTP
tipado (`src/api/atlasClient.ts`) cobrindo as 7 rotas reais do Atlas — nomes de campo em
snake_case espelhando exatamente as tags `json:"..."` do Go, sem inventar endpoint. Telas
para catálogo de serviços, contas de provedor, vínculos de credencial (sem nenhum campo
para o segredo em si, apenas `secret_ref` — CFG-05) e contratos por tenant. Proxy do Vite
(`/api` → `http://localhost:8081`) evita CORS no backend Go. Autenticação real
(SEG-01/OIDC) **não implementada** — placeholder documentado no próprio topo da UI.

Validação: `npm install` e `npm run build` executados **duas vezes de forma
independente** (uma pelo agente, uma nesta auditoria, com o mesmo resultado) — ambos sem
erro de TypeScript, bundle de produção gerado (~162 KB / 50 KB gzip). `npm audit` acusa 2
avisos conhecidos do dev server do `esbuild`/Vite (não afeta o build de produção),
não corrigidos por exigirem um upgrade maior (Vite 8) fora do escopo pedido.

## Estado de gate

Esta entrega produz evidência compatível com o **gate G1** (local/dev) de
`qualidade-e-aceite/spec.md`: implementação da fatia, testes de contrato/integração,
Compose, probes e telemetria demonstrados, emulação identificada como tal (LocalStack,
provedor simulado). **Não** produz evidência de G2 (exige sandbox real de
cliente/provedor) nem de gates superiores.

## Recomendação

1. Resolver P-01 (conta/região) e P-02 (volumes/perfis reais) antes de qualquer
   avanço para hom/ppd.
2. Tratar a lista de "Simplificações e placeholders" acima como um backlog explícito,
   priorizando SSRF (SEG-02) e autenticação real (SEG-01) antes de qualquer exposição
   externa, mesmo em ambiente de homologação.
3. Tratar composição/agregação (slice 3) como o próximo incremento funcional, não como
   trabalho já coberto por esta entrega.
