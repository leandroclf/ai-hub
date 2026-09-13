# Proposal: Integridade de dados e reconciliação
## Change ID
r5-03-integridade-e-reconciliacao
## Status
Draft — planejamento solicitado; código de produção não alterado nesta revisão.
## Why
TransformJSON agora passa todos os probes anteriores. A perda reaparece depois: operationFact.ResponseBody é any e json.Unmarshal usa float64. Probe reproduziu 9007199254740993→9007199254740992 na desserialização/serialização desse fato. Consolidação de produto também usa any; frontend converte fact_id com Number. Correção do validador não protege essas fronteiras.
Harness agora evita reutilizar alvo e compara digests SQL; permanece s3 sync/list-objects-v2 e uma única consulta ao oráculo antes de PASS. Não comprova versões históricas referenciadas, pins/tombstones, retomada nem reconciliação financeira/externa após replay. Lista de tabelas comparadas não inclui operation_plans/steps. Evidência histórica de cópia não encerra restore do produto atualizado.
Incidências SUBMITTED/STATUS, dedupe e ledger balanceado avançaram. Captura ainda muda estado da reserva sem apurar liberação da diferença para o valor real; SetWatermark aparece chamado em teste, sem produtor integrado de completude. Evento tardio insere finance_quarantine, apesar do comentário prometer disputa. São lacunas de fechamento operacional, não ausência de ledger.
ProcessEnvelope valida e chama Quarantine para envelope inválido; Quarantine persiste apenas payload_hash. O consumidor apaga a mensagem após retorno nil. Portanto o conteúdo inválido pode deixar de ser recuperável depois do ACK. O avanço do worker Cometa, que conserva bytes, não foi aplicado à autoridade financeira.
## Context
Brownfield f87ce33034ae29c9431b1910dcc6a633b545e330. R4 remota contém quatro changes; avanços preservados.
## Problem
Falhas nas fronteiras reais impedem cumprir os requisitos vinculados.
## Goals
- Precisão do resultado preservada em todas as fronteiras
- Restore populado e reconciliação antes de retomada
- Liquidação por valor efetivo e completude demonstrável
- Quarentena financeira conserva mensagem recuperável
## Non-Goals
Reescrever stack, fabricar aprovação comercial ou executar produção nesta revisão.
## Users / Actors Impacted
Clientes, provedores, operadores nominais e equipes de suporte. Responsável: Dados e Financeiro.
## Scope
### In scope
Comportamentos abaixo e fatias herdadas vinculadas, com testes de falha e integração.
### Out of scope
Provisionamento pago/deploy remoto e alteração de contratos normativos sem aprovação.
## Product Requirements Summary
- R5-DAD-01: O Hub SHALL preservar valor e tipo de números/identificadores em toda cadeia provedor→fato→estado→produto→representação→GET/webhook/console, sem coerção imprecisa. IDs inteiros fora da faixa segura do consumidor devem ter contrato textual explícito; resultados inválidos não são persistidos como sucesso.
- R5-DAD-02: O Hub SHALL demonstrar restore não vazio e retomada cercada de todas as autoridades e obrigações, incluindo planos, etapas, compensações, recibos, versões de objetos e financeiro. Antes de liberar tráfego deve reconciliar identidades/bytes/efeitos/valores; ausência de fixture ou divergência impede aprovação.
- R5-DAD-03: O Hub SHALL liquidar saldo pelo valor contratado efetivo, preservando reserva/hold de incerteza e liberando excedente de forma auditável. Fechamento exige evidência durável de completude dos produtores, sem watermarks fabricados. Fatos tardios devem gerar disposição financeira consultável e ajustável sem alterar exportação fechada.
- R5-DAD-04: O Hub SHALL conservar conteúdo recuperável e proveniência de mensagens financeiras não aplicáveis antes de confirmar sua remoção do transporte. Reprocessamento exige autorização, idempotência e trilha de disposição; hash isolado não constitui custódia do conteúdo.
## Business Rules
UUIDv7 após aceite, custódia antes de ACK, isolamento, resultado final único, snapshot contratual e valores exatos.
## Affected Capabilities
r5-03-integridade-e-reconciliacao
## Expected Impact
### Code
Fontes em explore.md e tarefas discriminadas por requisito.
### Data
Migrações aditivas; obrigações antigas preservadas, backfill validado e sem atualização cega do ledger.
### APIs / Contracts
Semântica explícita, versionada e compatível; erros não viram sucesso.
### Integrations
Qualificação externa por conta/adapter; fixture não equivale a contrato comercial.
### Operations
Scripts reproduzíveis e evidências do conteúdo atual; um Compose conforme AGENTS.
### Security / Privacy
Identidade do token/work item, menor privilégio, sem segredo em artefatos.
## Risks and Mitigations
Ver risk-matrix.md, design.md e cenários negativos.
## Success Criteria
Todos os cenários deste change e herdados afetados comprovados com código, oráculo, comando e digest.
## Assumptions
Defaults sintéticos autorizados no registro R3 podem apoiar laboratório; não aprovam produção.
## Open Questions
D-01…D-07/P-01…P-11/T-R2-01: conferir registro real. Somente gates dependentes ficam bloqueados.
