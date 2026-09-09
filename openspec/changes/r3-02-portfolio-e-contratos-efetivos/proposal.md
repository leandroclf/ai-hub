# Proposal: Portfólio e contratos efetivos
## Change ID
r3-02-portfolio-e-contratos-efetivos
## Status
Draft — especificação pronta para implementação; nenhuma tarefa implementada por esta entrega.
## Why
Provas executadas: TransformJSON altera 9007199254740993 para 9007199254740992 e aceita INVALID contra enum [OK]. Usa float64 e validação parcial.
Admissão usa tempos/modos do target sem composição efetiva de overrides. service_version recebido não participa da resolução. Finalizer usa FinalBody fixo sem OutputMapping.
Catálogo contém steps, mas admissão produz um comando com StepID igual ao protocolo; não há executor integrado do DAG/agregação nesse caminho.
ResolveOffer recusa tenant com mais de 100 ofertas antes de filtrar aplicação/serviço. Caminho Offer consulta Atlas em cada requisição.
## Context
Evolução do snapshot a4a876a9f8e875db882f7ca45cf7dece24d57aee; preservar mecanismos corretos v4/R2.
## Problem
As lacunas abaixo interrompem jornadas reais e deixam garantias normativas sem demonstração.
## Goals
- Precisão numérica e validação de schemas
- Política efetiva e representação por cliente
- Agregação e composição com executor de DAG
- Resolução indexada e projeção disponível
## Non-Goals
Reescrever stack, renomear componentes ou relaxar SLA/custódia para obter teste verde.
## Users / Actors Impacted
Clientes, provedores, operadores, desenvolvedores nominais. Responsável funcional: Produto e Core.
## Scope
### In scope
Requisitos e cenários deste change mais a requalificação herdada vinculada.
### Out of scope
Migração tecnológica não justificada e contratação comercial não autorizada.
## Product Requirements Summary
- R3-CAT-01: O Hub SHALL preservar valores numéricos exatos, validar integralmente o dialeto declarado e rejeitar publicação de construções de schema não suportadas. Transformação deve ter limites de profundidade/tamanho/custo e não executar código ou acessar rede/segredos.
- R3-CAT-02: O Hub SHALL definir precedência serviço→oferta→perfil dentro de limites contratuais, validar versão solicitada e congelar política efetiva/hash por aplicação. Representação final transformada deve ser persistida uma vez e reutilizada com bytes idênticos em GET e corpo de webhook.
- R3-CAT-03: O Hub SHALL executar DAG versionado com dependências, mapeamentos, paralelismo limitado, parcialidade, compensações e estado durável por etapa. Sucesso do produto depende do critério contratado; compensação não é rollback automático de efeito externo.
- R3-CAT-04: O Hub SHALL resolver ofertas por chave/vigência indexada sem teto artificial sobre o portfólio e distribuir projeções versionadas ao data plane. Snapshot válido sustenta caminho quente durante falha do controle dentro da validade/revogação definidas; cache frio não autoriza contrato desconhecido.
## Business Rules
UUIDv7 recuperável após aceite, isolamento de tenant/aplicação, snapshot imutável, custódia antes de ACK, resultado terminal único e finanças exatas permanecem obrigatórios.
## Affected Capabilities
02-portfolio-e-contratos-efetivos
## Expected Impact
### Code
hub/internal/atlas/catalog.go, hub/internal/atlas/offers.go, hub/internal/atlasclient/client.go, hub/internal/orbita/finalize.go, hub/internal/orbita/handlers.go
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
