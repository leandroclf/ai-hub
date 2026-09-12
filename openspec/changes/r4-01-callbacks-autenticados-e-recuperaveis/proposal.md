# Proposal: Callbacks autenticados e recuperáveis
## Change ID
r4-01-callbacks-autenticados-e-recuperaveis
## Status
Em execução — autenticação por conta implementada e qualificada no laboratório; qualificação integral e promoção externa pendentes.
## Why
O handler agora aceita, no caminho atual, uma assinatura HMAC-SHA256 versionada
derivada da conta, operação, timestamp e hash do corpo, sem depender do JWT de
workload interno nem transportar capability em query string. A capability por
operação permanece somente como compatibilidade legada e fica fora do
armazenamento em claro. A prova local fecha o caminho Cometa/provider-sim,
custódia, repetição idempotente e MFA manual do laboratório; o caminho pelo
gateway Kong local também foi exercitado. Endpoint público TLS, rotação/mTLS e
homologação de provedor comercial continuam pendentes.
Quando a operação não existe, a entrada somente é admitida após autenticação
HMAC por conta (ou, no caminho legado, chave de ingress mais capability) e o
formato terminal passarem; a identidade de autenticação é parte da chave de
deduplicação, portanto tentativa incorreta não ocupa o recibo legítimo. Quota,
retenção, claim, lease e epoch estão implementados e cobertos localmente. A
A política de autenticação específica da conta em provedores comerciais e sua
composição com endpoint externo TLS/mTLS continuam como risco residual.
ReconcileCallbackInboxBatch reivindica lote limitado antes da aplicação, com
claim/lease/epoch e `SKIP LOCKED`; `RunCallbackInboxWorker` executa a recuperação
periodicamente sem depender de novo tráfego HTTP. A prova cobre limite e
disposição de capability inválida; reinício em múltiplas réplicas e gateway
externo TLS/mTLS permanecem fora desta rodada.
SUBMIT, polling e callback convergem em `ApplyExternalObservation`; a rotina
confere correlação, snapshot, schema e projeção, conserva o recibo bruto e
classifica conflito sem reabrir o terminal. A suíte local cobre correlação
divergente, contrato estrito e conflito poll/callback; provedor comercial ainda
não está homologado.
## Context
Brownfield do commit b9d0f90ce02aa0c27cad546745153d160ff5867f. A baseline tem 189 requisitos/696 cenários.
## Problem
As lacunas observadas impedem contrato/custódia/qualificação integral.
## Goals
- Rota de callback compatível com a autenticação do provedor
- Admissão e deduplicação segura da inbox órfã
- Recuperador autônomo limitado e independente do HTTP
- Uma validação de resultado para SUBMIT, polling e callback
## Non-Goals
Reescrever stack, aprovar contrato comercial ou reduzir semântica de SLA.
## Users / Actors Impacted
Clientes, provedores e operadores nominais. Responsável funcional: Integrações e Core.
## Scope
### In scope
Requisitos deste change e fechamento das obrigações herdadas vinculadas.
### Out of scope
Push/merge/deploy remoto e exclusão de dados existentes sem autorização.
## Product Requirements Summary
- R4-CBK-01: O Hub SHALL expor uma rota de callback com autenticação explicitamente homologada por conta e independente da credencial de workload interna. O caminho público/gateway/middleware/handler deve ser qualificado de ponta a ponta. Capability de operação pode complementar correlação, nunca substituir silenciosamente política da conta; segredos não devem aparecer em URLs registradas, logs ou traces.
- R4-CBK-02: O Hub SHALL autenticar a origem antes de admitir órfão e vincular custódia a tenant/conta/célula/política e identidade do evento. Uma tentativa inválida não pode ocupar a chave de deduplicação de recibo legítimo. Quota, retenção, autorização e disposição devem ser explícitas; conflito de identidade mantém evidência sem descartar o evento correto.
- R4-CBK-03: O Hub SHALL reconciliar inbox por worker durável com lote limitado, claim/lease/epoch e isolamento por célula/conta/tenant. Aceite de callback conhecido não deve aguardar drenagem global. A recuperação deve ocorrer após reinício/correlação sem depender de tráfego futuro; um item inválido deve ter disposição sem bloquear os seguintes.
- R4-CBK-04: O Hub SHALL aplicar uma única validação de resultado externo para todas as modalidades, verificando conta, operação, correlação, schema e snapshot antes de alterar o estado. Conservar recibo original protegido e resultado normalizado versionado; divergência recebe disposição inválida sem substituir correlação/final. Não truncar a resposta a campos de simulador.
## Business Rules
Custódia antes de ACK, UUIDv7 após aceite, identidade por recurso, resultado terminal único, snapshots e finanças exatas permanecem.
## Affected Capabilities
r4-01-callbacks-autenticados-e-recuperaveis
## Expected Impact
### Code
Fontes e responsabilidades em design/explore.
### Data
Migrações aditivas, recibos e evidência sem alteração cega de histórico.
### APIs / Contracts
Validação/autorização efetivas e versionamento compatível.
### Integrations
Provedor, callback, cofre e contratos existentes com fixtures externas.
### Operations
Um Compose do Hub ativo por vez; scripts, gates e runbooks verificáveis.
### Security / Privacy
Menor privilégio, escopo de custódia e evidência sem segredos.
## Risks and Mitigations
Ver risk-matrix.md.
## Success Criteria
Cenários passam com evidência atual e requisitos herdados correspondentes requalificados.
## Assumptions
Dados comerciais ausentes não impedem fixtures sintéticas; não equivalem a aprovação.
## Open Questions
D-01…D-07, P-01…P-11 e T-R2-01 continuam sujeitos ao estado real do registro; não presumir aprovação.
