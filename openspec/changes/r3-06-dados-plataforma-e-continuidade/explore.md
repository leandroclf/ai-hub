# Explore: Dados, plataforma e continuidade

Snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee. Escopo brownfield v4/R2. Instruções: AGENTS.md e docs/openspec-docs/.

## F-R3-20
FileRefs/pins/upload existem; Execute não consome FileRefs e persistência de resultado volumoso não está conectada à execução real.

[hub/internal/objectstore/catalog.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/objectstore/catalog.go), [hub/internal/objectstore/retention.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/objectstore/retention.go), [hub/internal/cometa/executor.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/cometa/executor.go), [hub/internal/orbita/admission.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admission.go)

## F-R3-21
Kubernetes contém cinco serviços de negócio; local-kind aponta dependências a IPs de containers Compose. Overlays remotos não materializam sozinhos toda configuração/dependências.

[hub/deploy/r2/kind/render-runtime.py](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/kind/render-runtime.py), [hub/deploy/r2/k8s/base/workloads.yaml](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/k8s/base/workloads.yaml), [hub/deploy/r2/k8s/overlays/prd/kustomization.yaml](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/k8s/overlays/prd/kustomization.yaml)

## F-R3-22
Probes/réplicas/autoscalers são avanços declarativos, ainda sem comprovação de escala de nós/dados, placement automático e drenagem integrada.

[hub/deploy/r2/k8s/base/workloads.yaml](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2/k8s/base/workloads.yaml), [hub/internal/platform/httpserver](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/httpserver)

## F-R3-23
Custódia transacional avançou; RLS com papel real não proprietário e restore reconciliado continuam sem qualificação integrada. Fallback não pode transformar aceite em memória volátil.

[hub/migrations/core](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/migrations/core), [hub/migrations/control](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/migrations/control), [hub/internal/platform/pg](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/pg), [hub/internal/orbita/admission.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/orbita/admission.go)

## F-R3-24
Telemetria HTTP/manifests existem, mas API de SLA não funciona no painel e alertas de obrigação exigem produtor real. Render não prova scrape/log/trace/alerta.

[hub/internal/platform/httpserver/telemetry.go](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/internal/platform/httpserver/telemetry.go), [hub/admin-ui/src/pages/OperationsPage.tsx](https://github.com/leandroclf/ai-hub/blob/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/admin-ui/src/pages/OperationsPage.tsx), [hub/deploy/r2](https://github.com/leandroclf/ai-hub/tree/a4a876a9f8e875db882f7ca45cf7dece24d57aee/hub/deploy/r2)

Não confundir presença do mecanismo com qualificação integrada. Ver relatório R3 e evidência de testes.
