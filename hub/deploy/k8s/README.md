# Kubernetes de referência — Hub de Interoperabilidade (Constelação)

## Status: NÃO APLICADO

Os manifests desta pasta são **material de referência para revisão técnica**,
não uma entrega operacional. Nenhum `kubectl apply` foi executado contra
nenhum cluster — não há conta cloud nem cluster Kubernetes disponível para
este repositório nesta etapa, e a própria decisão de usar Kubernetes
gerenciado (ou não) ainda está em aberto (ver "Decisões pendentes" abaixo).

A única validação feita sobre estes arquivos foi sintática/client-side
(`kubectl --dry-run=client` e/ou parsing YAML), sem qualquer contato com um
API server real.

## Escopo

Um arquivo por serviço de negócio real, todos escritos em Go e construídos
pelo mesmo `hub/deploy/Dockerfile` genérico (build arg `SERVICE`):

| Arquivo | Serviço | Porta | Papel (OPE-05) |
|---|---|---|---|
| `atlas.yaml` | Atlas | 8081 | API — controle/config/SLA — escalado por HPA |
| `orbita.yaml` | Órbita | 8080 | API — admissão/consulta pública — escalado por HPA |
| `cometa.yaml` | Cometa | 8082 | Worker — controlador AIMD/despacho a provedor — KEDA (comentado) |
| `pulsar.yaml` | Pulsar | 8083 | Worker — entrega de webhook/callback — KEDA (exemplo incluso) |
| `libra.yaml` | Libra | 8084 | Worker — apuração/reservas financeiras — KEDA (comentado) |

`provider-sim` (8090) e `webhook-sink` (8091) existem apenas no
`docker-compose.yml` local para simular um provedor externo e um receptor
de webhook nos testes; **não têm manifest aqui** porque não devem ir para
nenhum ambiente Kubernetes real, incluindo ppd/prd.

Cada arquivo de serviço contém, na ordem: `ConfigMap` (configuração não
secreta), `Secret` (placeholder de string de conexão), `Deployment`,
`Service` (ClusterIP), `PodDisruptionBudget` e, apenas para Atlas/Órbita,
`HorizontalPodAutoscaler`. `kustomization.yaml` agrega os 5 arquivos sob o
namespace `hub` (também placeholder).

## Como isto se relaciona com o `docker-compose.yml` local

- **Mesma imagem, dockerfile genérico**: `hub/deploy/Dockerfile` recebe
  `ARG SERVICE` e compila `./cmd/<nome>`. Os manifests não reinventam a
  imagem — apontam para `<registry-placeholder>/hub-<nome>:<tag-placeholder>`,
  a ser resolvido quando P-01 definir registro/pipeline de publicação.
- **Mesmas variáveis de ambiente**: os nomes em `ConfigMap`/`Secret`
  (`CORE_DSN`, `CONTROL_DSN`, `FINANCE_DSN`, `ATLAS_URL`, `LIBRA_URL`,
  `COMETA_URL`, `ORBITA_URL`, `SELF_URL`, `QUEUE_ENDPOINT`, `QUEUE_REGION`,
  `HTTP_ADDR`) são exatamente os mesmos lidos pelos binários no
  `docker-compose.yml`. Só a origem do valor muda: no compose, valores
  literais no próprio arquivo; aqui, `envFrom` (`configMapRef`/`secretRef`),
  separando o que é configuração (URLs internas) do que é segredo (strings
  de conexão com usuário/senha).
- **Probes já existem no código**: os três endpoints (`/healthz/startup`,
  `/healthz/ready`, `/healthz/live`) já são implementados de forma genérica
  em `hub/internal/platform/httpserver/httpserver.go` para todos os
  serviços — os manifests apenas apontam para eles, não inventam contrato
  novo.
- **QUEUE_ENDPOINT** no compose aponta para o container `localstack`
  (SQS/SNS/S3 emulados). Nos manifests, esse valor é mantido como
  placeholder explícito porque o desenho real de filas gerenciadas
  (endpoint da AWS real, VPC endpoint, IRSA) depende de P-01.

## Decisões pendentes que bloqueiam uso real

Ver `openspec/changes/hub-interoperabilidade-v4/proposal.md`, seção
"Open Questions", e `design.md` ("nenhuma decisão pendente pode ser
tratada como resolvida por este design").

- **P-01** — Conta/região/Kubernetes gerenciado ou infraestrutura própria,
  orçamento, registro OCI e licenças (Plataforma). Bloqueia: `image`
  (registry/tag reais), `QUEUE_ENDPOINT` real, estratégia de IRSA/OIDC
  para acesso a SQS/SNS/S3, e a própria decisão de usar Kubernetes
  gerenciado em vez de outra topologia.
- **P-02** — Catálogo real, volumes, fan-out, perfis legados, mix real e
  classes de isolamento (Produto). Bloqueia: valores de `resources.requests/
  limits` (marcados como placeholder de referência), `minReplicas`/
  `maxReplicas` dos `HorizontalPodAutoscaler`, limiares do `ScaledObject`
  de exemplo do Pulsar, e o orçamento de conexões ao PostgreSQL somado nas
  réplicas (OPE-02) que qualquer envelope de escala real precisa respeitar.
- **P-11** — Envelopes de escala, quotas pré-autorizadas, reserva e
  política de células (Plataforma). Bloqueia: teto real de autoscaling
  (HPA e KEDA), estratégia de namespace/isolamento por célula (OPE-08) —
  o `namespace: hub` único usado em `kustomization.yaml` é apenas um
  placeholder de referência, não uma decisão de topologia de células.

Nenhuma dessas pendências é resolvida por estes manifests — elas
permanecem citadas como pendentes, conforme a mesma regra que rege o
restante do pacote OpenSpec desta mudança.

## Decisão HPA vs. KEDA (OPE-05/OPE-12)

Conforme OPE-05 ("HPA para APIs, KEDA para workers de fila... sem HPA
paralelo concorrente"): Atlas e Órbita são APIs síncronas (request/response
direto) e recebem `HorizontalPodAutoscaler` (`autoscaling/v2`, CPU +
memória). Cometa, Pulsar e Libra consomem fila/backlog como sinal primário
de carga — escalá-los por CPU/memória isolados ignoraria backlog e idade
de obrigação pendente (o sinal que OPE-05 pede explicitamente). Por isso
nenhum dos três tem `HorizontalPodAutoscaler` neste pacote; cada
`Deployment` documenta em comentário que o controlador de escala real
seria um KEDA `ScaledObject`. Um exemplo completo (porém **não aplicado,
não instalado**, dependente do operador KEDA e de valores ainda não
qualificados) está em `pulsar.yaml`, escalando por profundidade de fila
SQS — não duplicado em `cometa.yaml`/`libra.yaml` para evitar redundância,
já que o padrão é o mesmo.

## Validação executada

Sintaxe/schema client-side apenas (sem cluster): ver saída do agente que
gerou este pacote. Nenhum `kubectl apply` real, nenhum `kubectl diff`,
nenhum contato com API server.
