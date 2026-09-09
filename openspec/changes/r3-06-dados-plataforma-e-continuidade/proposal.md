# Proposal: Dados, plataforma e continuidade
## Change ID
r3-06-dados-plataforma-e-continuidade
## Status
Draft — especificação pronta para implementação; nenhuma tarefa implementada por esta entrega.
## Why
FileRefs/pins/upload existem; Execute não consome FileRefs e persistência de resultado volumoso não está conectada à execução real.
Kubernetes contém cinco serviços de negócio; local-kind aponta dependências a IPs de containers Compose. Overlays remotos não materializam sozinhos toda configuração/dependências.
Probes/réplicas/autoscalers são avanços declarativos, ainda sem comprovação de escala de nós/dados, placement automático e drenagem integrada.
Custódia transacional avançou; RLS com papel real não proprietário e restore reconciliado continuam sem qualificação integrada. Fallback não pode transformar aceite em memória volátil.
Telemetria HTTP/manifests existem, mas API de SLA não funciona no painel e alertas de obrigação exigem produtor real. Render não prova scrape/log/trace/alerta.
## Context
Evolução do snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee; preservar mecanismos corretos v4/R2.
## Problem
As lacunas abaixo interrompem jornadas reais e deixam garantias normativas sem demonstração.
## Goals
- Objetos conectados ao fluxo e retenção
- Kind completo e ambientes reprodutíveis
- Escala e continuidade com envelope
- Autoridade durável, isolamento e restore
- SLA bilateral e telemetria verificável
## Non-Goals
Reescrever stack, renomear componentes ou relaxar SLA/custódia para obter teste verde.
## Users / Actors Impacted
Clientes, provedores, operadores, desenvolvedores nominais. Responsável funcional: Plataforma, Dados e SRE.
## Scope
### In scope
Requisitos e cenários deste change mais a requalificação herdada vinculada.
### Out of scope
Migração tecnológica não justificada e contratação comercial não autorizada.
## Product Requirements Summary
- R3-OPE-01: O Hub SHALL consumir referências imutáveis por streaming limitado, conservar resultado volumoso antes da publicação e integrar retenção/pins a protocolos, entregas e finanças. Classe/região restringem processamento e acesso.
- R3-OPE-02: A entrega SHALL oferecer Compose local completo e kind completo com dependências no cluster, UI/gateway/identidade/dados/mensageria/cofre emulável/observabilidade. Serviços gerenciados em dev/hom/ppd/prd exigem IaC e referências explícitas; imagens e endpoints devem ser reproduzíveis sem IP efêmero.
- R3-OPE-03: O Hub SHALL automatizar onboarding/placement e escala de pods/nós/dados dentro de envelope aprovado, separar readiness por capacidade e drenar workers/leases. Saturação além do envelope requer admissão controlada; escala infinita e failover instantâneo não são promessas válidas.
- R3-OPE-04: O Hub SHALL preservar autoridade durável replicada, papéis mínimos e isolamento tenant/aplicação/célula, e qualificar restore reconciliando mensagens/efeitos/objetos/finanças. Redis é dispensável. Falha total da autoridade exige recusa segura; fallback durável alternativo exige consistência/fencing comprovados.
- R3-OPE-05: O Hub SHALL emitir métricas reais de idade/custódia/lag/capacidade e SLA cliente-Hub/Hub-provedor, permitir consulta autorizada de violações e correlacionar logs/traces/protocolos sem segredos. Prometheus/Loki/Grafana devem ser testados com falhas injetadas e cardinalidade limitada.
## Business Rules
UUIDv7 recuperável após aceite, isolamento de tenant/aplicação, snapshot imutável, custódia antes de ACK, resultado terminal único e finanças exatas permanecem obrigatórios.
## Affected Capabilities
06-dados-plataforma-e-continuidade
## Expected Impact
### Code
hub/admin-ui/src/pages/OperationsPage.tsx, hub/deploy/r2, hub/deploy/r2/k8s/base/workloads.yaml, hub/deploy/r2/k8s/overlays/prd/kustomization.yaml, hub/deploy/r2/kind/render-runtime.py, hub/internal/cometa/executor.go, hub/internal/objectstore/catalog.go, hub/internal/objectstore/retention.go, hub/internal/orbita/admission.go, hub/internal/platform/httpserver, hub/internal/platform/httpserver/telemetry.go, hub/internal/platform/pg, hub/migrations/control, hub/migrations/core
### Data
Migrações aditivas e identidades/versionamento explícitos; sem perda de registros históricos.
### APIs / Contracts
Compatibilidade por versão, erros tipados, autorização em backend e sem sucesso fictício.
### Integrations
Caminho real com oráculos e credenciais de fixture, sem credenciais comerciais em artefatos.
### Operations
Métricas, recuperação, runbook e rollback devem acompanhar o comportamento.
### Security / Privacy
Menor privilégio; sem payload/segredo bruto em logs/evidência.
## Risks and Mitigations
Consultar risk-matrix.md; cada risco tem requisito, cenário e tarefa.
## Success Criteria
Todos os cenários deste change e regressão vinculada passam no SHA entregue; zero skip no gate obrigatório.
## Assumptions
Stack existente é mantida. Valores comerciais ausentes são fixtures explicitamente sintéticas.
## Open Questions
Decisões D-01…D-07/P-01…P-11 e T-R2-01 são preservadas; ver documento R3 05. Semântica não pode mudar por inferência.
