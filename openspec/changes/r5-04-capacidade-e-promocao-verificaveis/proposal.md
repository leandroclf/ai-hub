# Proposal: Capacidade, ambientes e promoção verificáveis
## Change ID
r5-04-capacidade-e-promocao-verificaveis
## Status
Draft — planejamento solicitado; código de produção não alterado nesta revisão.
## Why
Controle adaptativo e pools agora estão ligados a SUBMIT/STATUS/reconciliação e webhook. Contudo domínio vazio/controller nil desabilita controle; permissões externas expiradas não são recicladas automaticamente e falha de settlement apenas gera log. Scripts históricos precisaram reconciliar permits por 404 do simulador. Contagens percorrem histórico de permits por domínio a cada Acquire sob lock global do domínio.
Probe local do script devolveu ALLOW para prd com profile=unverified, três nomes de aprovação e isolation=PASS, sem manifesto de evidência. Script é gate, não deploy: nenhuma implantação foi feita. Validador exige formato de SHA, mas isso não vincula sozinho execução ao artefato promovido. HEAD altera bases Go/Nginx após evidências anteriores; Node build não está fixado por digest.
Kind independente agora inclui dependências/UI/gateway: o achado antigo de ausência deve ser encerrado nesse escopo. Overlays remotos continuam centrados nas réplicas de cinco serviços; laboratório independente não constitui IaC regional nem durabilidade/escala de dados. Recuperar dois pods não mede continuidade de negócio sob perda de nó/zona e backlog financeiro.
## Context
Brownfield f87ce33034ae29c9431b1910dcc6a633b545e330. R4 remota contém quatro changes; avanços preservados.
## Problem
Falhas nas fronteiras reais impedem cumprir os requisitos vinculados.
## Goals
- Capacidade sem bypass e recuperação de permits
- Promoção exige evidência vinculada ao artefato
- Ambientes elásticos com dados duráveis e isolamento completo
## Non-Goals
Reescrever stack, fabricar aprovação comercial ou executar produção nesta revisão.
## Users / Actors Impacted
Clientes, provedores, operadores nominais e equipes de suporte. Responsável: Plataforma e Operações.
## Scope
### In scope
Comportamentos abaixo e fatias herdadas vinculadas, com testes de falha e integração.
### Out of scope
Provisionamento pago/deploy remoto e alteração de contratos normativos sem aprovação.
## Product Requirements Summary
- R5-OPE-01: O Hub SHALL exigir política de capacidade qualificada para cada rota ativa e recuperar concessões pendentes com evidência durável de transporte/efeito, sem reciclar apenas por timeout. Crescimento de histórico e número de tenants não deve violar orçamento de concessão publicado; isolamento e limite agregado devem valer entre réplicas.
- R5-OPE-02: O processo de promoção SHALL recusar artefato sem evidência íntegra e compatível com SHA/conteúdo/imagens efetivos e sem aprovações verificáveis do ambiente aplicável. Uma string PASS ou nome de aprovação não constitui prova. Mudança de toolchain/base exige requalificação material antes da promoção.
- R5-OPE-03: O Hub SHALL disponibilizar perfis local/dev/hom/ppd/prd reproduzíveis com identidade, dados, segredos, rede, observabilidade e capacidade isolados. Escala automática deve abranger pods/nós/placement e budgets das dependências dentro de quotas explícitas; continuidade é aferida por requisições/obrigações reconciliadas, não somente readiness.
## Business Rules
UUIDv7 após aceite, custódia antes de ACK, isolamento, resultado final único, snapshot contratual e valores exatos.
## Affected Capabilities
r5-04-capacidade-e-promocao-verificaveis
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
