# Delta for r4-01-callbacks-autenticados-e-recuperaveis

## ADDED Requirements

### Requirement: R4-CBK-01 — Rota de callback compatível com a autenticação do provedor
O Hub SHALL expor uma rota de callback com autenticação explicitamente homologada por conta e independente da credencial de workload interna. O caminho público/gateway/middleware/handler deve ser qualificado de ponta a ponta. Capability de operação pode complementar correlação, nunca substituir silenciosamente política da conta; segredos não devem aparecer em URLs registradas, logs ou traces.

#### Scenario: R4-CBK-01-S01 — provedor homologado com credencial própria e callback conhecido
- GIVEN provedor homologado com credencial própria e callback conhecido
- WHEN enviar pela URL realmente devolvida no SUBMIT sem token Hub
- THEN a política da conta autoriza e o recibo precede 2xx

#### Scenario: R4-CBK-01-S02 — token ausente/inválido ou identidade de outra conta
- GIVEN token ausente/inválido ou identidade de outra conta
- WHEN enviar pelo ingresso externo
- THEN requisição é recusada sem acesso a operação nem custódia de payload não autorizado

#### Scenario: R4-CBK-01-S03 — rotação de segredo e observação repetida
- GIVEN rotação de segredo e observação repetida
- WHEN entregar dentro/fora da janela de versão
- THEN somente versões elegíveis são aceitas e proxies/logs não expõem material de autenticação

### Requirement: R4-CBK-02 — Admissão e deduplicação segura da inbox órfã
O Hub SHALL autenticar a origem antes de admitir órfão e vincular custódia a tenant/conta/célula/política e identidade do evento. Uma tentativa inválida não pode ocupar a chave de deduplicação de recibo legítimo. Quota, retenção, autorização e disposição devem ser explícitas; conflito de identidade mantém evidência sem descartar o evento correto.

#### Scenario: R4-CBK-02-S01 — callback autenticado anterior à correlação
- GIVEN callback autenticado anterior à correlação
- WHEN admitir órfão e depois associar
- THEN 202 confirma custódia com escopo e o evento legítimo é aplicado uma vez

#### Scenario: R4-CBK-02-S02 — tentativa de token incorreto seguida de correto para mesmo ID/body
- GIVEN tentativa de token incorreto seguida de correto para mesmo ID/body
- WHEN receber e reconciliar
- THEN a primeira não bloqueia nem substitui a evidência legítima

#### Scenario: R4-CBK-02-S03 — origem excede quota ou tenta outro tenant
- GIVEN origem excede quota ou tenta outro tenant
- WHEN admitir payload
- THEN recusa controlada e retenção limitada preservam capacidade e isolamento dos demais

### Requirement: R4-CBK-03 — Recuperador autônomo limitado e independente do HTTP
O Hub SHALL reconciliar inbox por worker durável com lote limitado, claim/lease/epoch e isolamento por célula/conta/tenant. Aceite de callback conhecido não deve aguardar drenagem global. A recuperação deve ocorrer após reinício/correlação sem depender de tráfego futuro; um item inválido deve ter disposição sem bloquear os seguintes.

#### Scenario: R4-CBK-03-S01 — órfão admitido e correlação criada sem novo callback
- GIVEN órfão admitido e correlação criada sem novo callback
- WHEN reiniciar worker
- THEN o recibo é aplicado autonomamente dentro do orçamento configurado

#### Scenario: R4-CBK-03-S02 — item defeituoso antecede recibo válido de outro tenant
- GIVEN item defeituoso antecede recibo válido de outro tenant
- WHEN processar lote
- THEN o defeituoso é isolado e o válido progride sem bloquear seu HTTP

#### Scenario: R4-CBK-03-S03 — dois workers e pool SQL pequeno com perda de lease
- GIVEN dois workers e pool SQL pequeno com perda de lease
- WHEN processar concorrência
- THEN claims e fencing impedem posse inválida e não há espera circular por cursor/conexões

### Requirement: R4-CBK-04 — Uma validação de resultado para SUBMIT, polling e callback
O Hub SHALL aplicar uma única validação de resultado externo para todas as modalidades, verificando conta, operação, correlação, schema e snapshot antes de alterar o estado. Conservar recibo original protegido e resultado normalizado versionado; divergência recebe disposição inválida sem substituir correlação/final. Não truncar a resposta a campos de simulador.

#### Scenario: R4-CBK-04-S01 — mesmo resultado recebido por SYNC, poll e callback
- GIVEN mesmo resultado recebido por SYNC, poll e callback
- WHEN normalizar e finalizar
- THEN contrato e conteúdo de negócio equivalentes são preservados sob snapshot

#### Scenario: R4-CBK-04-S02 — callback com ID de provedor divergente ou schema inválido
- GIVEN callback com ID de provedor divergente ou schema inválido
- WHEN receber após autenticação
- THEN recibo é conservado como inválido sem trocar correlação nem produzir sucesso

#### Scenario: R4-CBK-04-S03 — resultado válido depois do final do cliente
- GIVEN resultado válido depois do final do cliente
- WHEN receber via qualquer observador
- THEN evidência externa é mantida sem reabrir protocolo nem duplicar incidência
