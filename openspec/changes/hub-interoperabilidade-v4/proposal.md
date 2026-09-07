# Proposal: Hub de Interoperabilidade "Constelação" — especificação de engenharia v4.0 pronta para implementação

## Change ID
`hub-interoperabilidade-v4`

## Status
Draft

## Context

A HivePlace mantém, em `docs/`, uma especificação de engenharia de requisitos v4.0 (nome de trabalho "Constelação", datada de 6 de setembro de 2026) para um hub de interoperabilidade entre clientes e provedores de serviços externos. O pacote é composto por dez capítulos Markdown (`00_LEIA_PRIMEIRO.md` a `09_DECISOES_E_REFERENCIAS.md`), uma matriz de rastreabilidade (`MATRIZ_RASTREABILIDADE.csv`) com 98 requisitos e 98 cenários mínimos de aceite, um diagrama SVG e uma edição HTML consolidada. O pacote também inclui, em `docs/openspec-docs/`, um método recomendado ("Spec-Driven Development com OpenSpec") para transformar essa especificação em artefatos versionáveis (`proposal.md`, deltas de specs por capability, `design.md`, `tasks.md`).

Até esta mudança, o repositório continha apenas a documentação-fonte: nenhum artefato OpenSpec, nenhum código de aplicação, nenhum manifesto de infraestrutura. A especificação é explicitamente documental — "não afirma implementação, teste de carga ou publicação" — e todos os 98 cenários da matriz estão no estado **NÃO EXECUTADO**.

Esta mudança aplica o fluxo recomendado em `docs/openspec-docs/01-prompt-final-completo.md` e `README.md` sobre o conteúdo dos capítulos 01 a 09, produzindo o primeiro conjunto de artefatos OpenSpec (`proposal.md`, specs por capability, `design.md`, `tasks.md` e artefatos auxiliares) que tornam a especificação "implementation-ready" sem, ainda, implementar código.

## Problem

A especificação v4.0 é extensa (~31 mil palavras), normativa e escrita em prosa corrida por capítulo. Isso a torna difícil de:

1. **Rastrear por capability** — um requisito de execução (EXE-14) depende de decisões de arquitetura (COM-01/06), dados (DAD-09) e financeiro (FIN-06), espalhadas em capítulos diferentes, sem um contrato de comportamento observável isolado por domínio.
2. **Validar de forma incremental** — não há tarefas pequenas, sequenciais e verificáveis; a matriz lista requisitos e cenários, mas não uma ordem de implementação nem critérios de conclusão por tarefa.
3. **Auditar decisões pendentes** — 11 decisões de negócio/infraestrutura (P-01 a P-11) e 15 gaps de risco (GAP-01 a GAP-15) bloqueiam partes específicas do escopo, mas não estão conectados a um plano de execução.

Sem uma tradução para o formato OpenSpec (specs comportamentais + design técnico + tarefas), qualquer implementação — humana ou por agente de IA — corre o risco de reinterpretar requisitos, perder invariantes críticos (ex.: aceite durável, isolamento de tenant, não duplicação de cobrança) ou não saber quando uma fatia está de fato pronta para promoção entre ambientes.

## Goals

- Traduzir os 98 requisitos normativos dos capítulos 01–09 em requirements OpenSpec (`SHALL`/`MUST`/`SHOULD`/`MAY`) com cenários `GIVEN/WHEN/THEN` verificáveis, organizados por capability.
- Produzir um `design.md` único que descreva a arquitetura de referência (5 aplicações + gateway, célula, SYNC direto vs. ASYNC/AUTO, persistência por domínio, segurança de credenciais, observabilidade) sem introduzir decisões técnicas não presentes na especificação-fonte.
- Produzir um `tasks.md` com tarefas pequenas, sequenciais e verificáveis, organizadas segundo a fatia inicial recomendada em QUA-04 (serviço síncrono em SYNC+ASYNC → polling/callback concorrentes → produto composto/planos avançados) e os gates G0–G4.
- Preservar fielmente as ambiguidades, decisões pendentes (P-01–P-11) e gaps de risco (GAP-01–GAP-15) já registrados no capítulo 09, sem resolvê-los por conta própria.
- Deixar o pacote pronto para os gates G0 (revisão) sem exigir decisões comerciais/infraestruturais que a especificação já marca como pendentes.

## Non-Goals

- Não implementar código de produção (Go, TypeScript/React, manifests Kubernetes, Terraform/Crossplane, migrations SQL). Esta mudança é exclusivamente documental/OpenSpec.
- Não resolver as decisões pendentes P-01 a P-11 (conta/região AWS, volumes reais, tarifas comerciais, retenção legal, RPO/RTO regional, perfil de criticidade, etc.). Essas permanecem em aberto e são citadas como tal.
- Não criar testes automatizados nem executar nenhum cenário da matriz — todos os 98 cenários continuam **NÃO EXECUTADO**.
- Não alterar o conteúdo normativo dos capítulos `docs/*.md`; esta mudança apenas os traduz para o formato OpenSpec, sem adicionar, remover ou reinterpretar requisitos.
- Não cobrir SFTP em lote, gRPC/Protobuf ou broker externo de parceiro além do que COM-02 já descreve como "evolução futura" — esses permanecem fora de escopo, como no texto-fonte.

## Users / Actors Impacted

- **Engenharia** (implementará Portal, Atlas, Órbita, Cometa, Pulsar, Libra a partir destes artefatos).
- **Produto / Comercial** (aprovam semântica de serviços, planos e condições comerciais).
- **Arquitetura** (dona das decisões ADR-01 a ADR-25 e das fronteiras entre componentes).
- **Segurança** (credenciais, isolamento de tenant, SSRF, auditoria).
- **QA** (estratégia de testes, gates G0–G4, evidência).
- **Plataforma/SRE** (ambientes, HA, escala automática, runbooks).
- **Financeiro** (medição econômica, ledger, conciliação).
- **Liderança técnica** (aprovação da v4, responsáveis pelas decisões pendentes P-01–P-11).
- **Clientes e provedores externos** (atores de domínio descritos nos requisitos, não usuários diretos deste artefato).

## Scope

### In scope
- Deltas de spec OpenSpec (`ADDED Requirements`) para as 9 capabilities derivadas dos capítulos 01–09: `catalogo-e-portfolio`, `arquitetura-e-comunicacao`, `persistencia-e-dados`, `execucao-e-integracoes`, `contratos-e-financeiro`, `configuracao-e-seguranca`, `desempenho-e-operacao`, `qualidade-e-aceite`, `decisoes-e-governanca`.
- `design.md` único, cobrindo arquitetura, decisões técnicas, fluxos principais/erro, modelo de dados, segurança, observabilidade e estratégia de testes de referência.
- `tasks.md` com tarefas pequenas e sequenciais cobrindo a fatia inicial recomendada por QUA-04 até a preparação para os gates subsequentes.
- Artefatos auxiliares: matriz de riscos consolidada (gaps DEC-05 + pendências P-01–P-11) e checklist de avaliador adaptado a esta mudança.

### Out of scope
- Qualquer código de aplicação, migração de banco, manifest de infraestrutura ou pipeline de CI/CD.
- Execução de qualquer cenário de teste da matriz (`MATRIZ_RASTREABILIDADE.csv`) ou de QUA-02/05/06.
- Resolução das decisões pendentes P-01 a P-11 e dos gaps GAP-01 a GAP-15 (permanecem registrados como pendentes/parcialmente mitigados no desenho, não encerrados).
- Contratação de conta/região AWS, licenciamento de Kong, ou qualquer decisão comercial/orçamentária.

## Product Requirements Summary

- Hub multi-tenant que intermedia clientes e provedores externos, com portfólio de serviços/produtos versionados (CAT-01 a CAT-11), oferecendo modos de atendimento SYNC, ASYNC e AUTO (EXE-02) sobre uma única máquina de estados durável.
- Toda admissão recebe `protocol_id` UUIDv7 (EXE-16), com idempotência obrigatória por `Idempotency-Key` e deadline absoluto não renovável (EXE-11).
- Consulta (`GET`) e webhook usam sempre a mesma representação final persistida (COM-05); o hub nunca reconsulta o provedor para responder a um `GET`.
- Medição econômica (receita do cliente, custo do provedor) é independente da confirmação técnica de entrega e nunca duplicada (FIN-04/FIN-05).
- Isolamento de tenant é absoluto por padrão, com um perfil administrativo individual e auditado (`hub_protocol_reader`) como única exceção (SEG-04).
- Capacidade cresce por automação (HPA/KEDA/Karpenter/Crossplane) dentro de perfis homologados, sem exigir chamados manuais de infraestrutura para crescimento rotineiro de clientes/provedores (OPE-12).
- Redis é sempre dispensável: todo caminho crítico deve continuar operando, com SLO qualificado, quando o Redis está desligado (DAD-10).

## Business Rules

As 22 invariantes obrigatórias do pacote (`docs/00_LEIA_PRIMEIRO.md`) aplicam-se a todas as capabilities e são citadas aqui como regras de negócio transversais, sem reinterpretação:

1. Aceite durável apenas após persistir protocolo, pedido, versão de configuração e intenção de execução.
2. Consulta do cliente lê apenas o hub; nunca dispara reexecução no provedor.
3. Protocolo, passo, operação externa, tentativa física, evento, resultado, entrega e fato econômico têm identidades próprias e vínculos verificáveis.
4. Callback, polling e respostas imediatas concorrem para a mesma operação; só transições válidas consolidam resultado.
5. Repetição de mensagem/webhook não repete efeito de negócio nem cobra novamente.
6. Faturamento do cliente, custo do provedor e confirmação de entrega são dimensões independentes.
7. Cada execução fixa as versões de produto/contrato aplicáveis; alterações posteriores não reescrevem a história.
8. Pedido aceito terá conclusão contratual ou pendência operacional explícita — nunca abandono silencioso.
9. Isolamento: nenhum cliente acessa dados de outro cliente por conhecer um identificador.
10. Exactly-once externo depende das capacidades do provedor; resultado desconhecido bloqueia reexecução insegura.
11. Prazo absoluto: retry/polling/fila/reinício/troca de provedor não renovam o SLA do cliente.
12. Protocolo encerrado por prazo não é reaberto por retorno tardio; evidência é preservada e segregada.
13. GET e webhook usam a mesma representação contratada e a mesma versão final.
14. Taxa e concorrência efetivas variam com métricas, dentro de limites técnicos e contratuais (pressão adaptativa).
15. Tráfego de um tenant não consome a capacidade reservada de outro; classe de isolamento e domínio de falha são explícitos.
16. Clientes não cruzam tenants; administradores individuais podem consultar qualquer tenant, com perfil auditado.
17. Serviço SYNC elegível devolve o final na mesma requisição; fila não é espera obrigatória desse caminho.
18. Toda admissão (SYNC ou ASYNC) recebe UUIDv7 persistido; consultar de novo não repete a operação.
19. Toda operação resolve explicitamente conta/modo (compartilhado/dedicado)/vínculo/versão de segredo/responsável econômico, sem fallback silencioso.
20. Redis é dispensável: nenhuma cópia única de protocolo, resultado, idempotência, saldo, outbox, agenda, autorização ou quota global vive só nele.
21. Escala de clientes/provedores usa perfis homologados e automação de capacidade, preservando a arquitetura e as quotas explícitas.
22. Fallback só usa informação válida, autorizada e durável quando a obrigação exigir; nunca inventa resposta ou repete efeito incerto para aparentar disponibilidade.

## Affected Capabilities
- `catalogo-e-portfolio`
- `arquitetura-e-comunicacao`
- `persistencia-e-dados`
- `execucao-e-integracoes`
- `contratos-e-financeiro`
- `configuracao-e-seguranca`
- `desempenho-e-operacao`
- `qualidade-e-aceite`
- `decisoes-e-governanca`

## Expected Impact

### Code
Nenhum nesta mudança. `design.md` e `tasks.md` preparam a implementação futura de cinco aplicações Go (Portal/Kong, Atlas, Órbita, Cometa, Pulsar, Libra) e uma interface TypeScript/React (Atlas), conforme ARQ-04.

### Data
Nenhuma migração nesta mudança. `persistencia-e-dados/spec.md` e `design.md` descrevem o modelo lógico (hub_control, hub_core, hub_finance por célula) que orientará o desenho físico futuro (DAD-01/02).

### APIs / Contracts
Nenhum contrato publicado nesta mudança. Os requirements de `execucao-e-integracoes` e `arquitetura-e-comunicacao` definem o comportamento observável que os futuros contratos OpenAPI/AsyncAPI deverão implementar (COM-04).

### Integrations
Nenhuma integração real nesta mudança. `configuracao-e-seguranca/spec.md` e `contratos-e-financeiro/spec.md` definem os requisitos de credenciais, modalidades e apuração que qualquer adaptador futuro deve satisfazer.

### Operations
Nenhuma operação de infraestrutura nesta mudança. `desempenho-e-operacao/spec.md` e `design.md` descrevem SLOs propostos, ambientes, probes, observabilidade e runbooks de referência para qualificação futura (gates G1–G4).

### Security / Privacy
Nenhuma mudança de sistema. `configuracao-e-seguranca/spec.md` preserva os requisitos de identidade, isolamento de tenant, prevenção de SSRF e rotação de segredo já definidos na especificação-fonte.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Tradução para OpenSpec introduzir requisito não presente no texto-fonte | Divergência entre a spec "fonte da verdade" e a intenção original de negócio | Cada requirement cita o ID de origem (ex. CAT-01) e cada capability tem seção "Notas de origem"; revisão cruzada contra `MATRIZ_RASTREABILIDADE.csv` antes de qualquer implementação |
| Ambiguidade documental ser resolvida silenciosamente durante a tradução | Perda de um "ponto em aberto" que deveria bloquear implementação (regra de parada da metodologia OpenSpec) | Pontos em aberto e decisões pendentes (P-01–P-11) são preservados literalmente em `design.md` e no artefato auxiliar de riscos, não "resolvidos" nesta mudança |
| Gaps críticos do DEC-05 (GAP-02, GAP-05, GAP-09, GAP-10, GAP-11, GAP-12, GAP-13, GAP-15) serem tratados como já mitigados | Falsa sensação de prontidão para produção | `decisoes-e-governanca/spec.md` e a matriz de riscos auxiliar preservam a distinção entre "resolvido no desenho" e "qualificação pendente" |
| Volume de 98 requisitos gerar specs inconsistentes entre si (nomenclatura, granularidade de cenários) | Dificulta a QA e a leitura por agentes de IA na implementação | Todas as capabilities seguem o mesmo template (`### Requirement: <ID> — <título>` + `#### Scenario`) e a mesma convenção de verbos normativos |

## Success Criteria

- Todos os 98 requisitos da matriz têm um `### Requirement` correspondente em algum arquivo `openspec/changes/hub-interoperabilidade-v4/specs/*/spec.md`, identificável pelo ID de origem.
- `design.md` cobre todas as seções do template (arquitetura, decisões técnicas, fluxos, dados, segurança, observabilidade, testes, migração, rollback, compatibilidade, riscos remanescentes) sem contradizer os capítulos-fonte.
- `tasks.md` decompõe a fatia inicial recomendada (QUA-04) em tarefas com objetivo, arquivos/componentes prováveis, dependências, tipo de validação e critério de conclusão.
- Nenhuma decisão pendente (P-01–P-11) nem gap crítico (DEC-05) é apresentado como resolvido nesta mudança.
- O checklist de avaliador (`docs/openspec-docs/04-checklist-avaliador.md`) pode ser aplicado a este proposal/specs/design/tasks sem apontar itens fora de escopo cobertos silenciosamente.

## Assumptions

- O change-id `hub-interoperabilidade-v4` cobre a especificação inteira como uma única mudança inicial (greenfield), em vez de mudanças separadas por capability, dado que não existe ainda nenhuma spec publicada em `openspec/specs/`.
- A tradução para specs comportamentais pode nomear os componentes de domínio (Portal, Atlas, Órbita, Cometa, Pulsar, Libra) e tecnologias já decididas na especificação (PostgreSQL, SNS/SQS, S3, Kong, Kubernetes, UUIDv7) como vocabulário de domínio, não como detalhe de implementação a esconder.
- QUA-01 a QUA-06 (capítulo 08) e DEC-01 a DEC-05 (capítulo 09) são tratados como capabilities de processo/governança (comportamento exigido do processo de qualidade e de decisão), não como funcionalidades do sistema-produto, já que descrevem gates, evidência e registro de decisões.

## Open Questions

Estas são as decisões pendentes já registradas em DEC-02 do capítulo 09 (não resolvidas por esta mudança):

- P-01 — Conta/região/Kubernetes gerenciado ou infraestrutura própria, orçamento, registro OCI e licenças (Plataforma).
- P-02 — Catálogo real, volumes, fan-out, perfis legados, mix real e classes de isolamento (Produto).
- P-03 — Contratos de compra/venda, tarifas, planos, marcos, faixas, parcialidade e riscos de failover (Comercial).
- P-04 — Retenção, finalidade, região permitida e direitos de custódia dos resultados (responsável pelos dados).
- P-05 — Idempotência, SYNC/polling/callback, credenciais compartilhadas/dedicadas e equivalência de provedores (Integrações).
- P-06 — ERP, layout/API, plano de contas gerencial, arredondamento, competência e liquidação (Financeiro).
- P-07 — SLA bilateral, TTL em segundos, enforcement do provedor e compromisso de entrega (Produto).
- P-08 — Domínio de falhas, RPO zero contratado, confirmação entre regiões e RTO por célula (SRE).
- P-09 — Responsáveis, perfis administrativos dos desenvolvedores e aprovação formal da v4 (liderança técnica).
- P-10 — Perfil de criticidade para uso com impacto em vida/segurança (Produto).
- P-11 — Envelopes de escala, quotas preautorizadas, reserva e política de células (Plataforma).

Estas pendências bloqueiam apenas a produção/oferta específica a que se referem — não bloqueiam a elaboração dos artefatos OpenSpec desta mudança.
