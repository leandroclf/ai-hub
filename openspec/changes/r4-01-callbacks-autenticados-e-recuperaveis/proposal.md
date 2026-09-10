# Proposal: Callbacks autenticados e recuperáveis
## Change ID
r4-01-callbacks-autenticados-e-recuperaveis
## Status
Em execução — implementação parcial; qualificação integral pendente.
## Why
O handler ganhou capability por operação e custódia antes de 2xx. Entretanto, toda a rota /internal/ continua envolvida pelo JWT do Hub; a URL entregue ao provedor contém apenas capability. O simulador não envia JWT Hub. A correção do handler não fecha o caminho de rede. Isso é incompatibilidade de composição, não prova de endpoint publicamente desprotegido.
Quando a operação não existe, qualquer token não vazio que alcance o handler permite inserir body e receber 202. A chave única usa operação+hash do body, sem identidade autenticada/token: primeira tentativa com token incorreto pode ocupar a identidade de posterior recibo correto, que só incrementa occurrences. A tabela também não registra tenant/conta/célula, TTL ou quota. Exploração externa depende da rota/autenticação atual; o risco permanece ao corrigir essa rota.
ReconcileCallbackInbox só é chamado ao receber outro callback conhecido. A consulta percorre todos os RECEIVED associados, sem LIMIT, claim, lease ou escopo de célula. O cursor permanece aberto enquanto são feitas outras operações SQL e aplicação. O HTTP de uma operação pode depender do backlog/erro de outra, e sem novo callback não há gatilho autônomo.
finalize valida OutputSchema e conserva OperationResult. ApplyExternalObservation transforma callback somente em detail, não chama a mesma validação e não confere provider_request_id contra correlação armazenada. ConserveObservation pode sobrescrever a correlação com qualquer ID não vazio recebido. Token de operação não torna o body semanticamente correto.
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
