# Checklist de avaliação — hub-interoperabilidade-v4

Adaptado de `docs/openspec-docs/04-checklist-avaliador.md` para validar a **implementação futura** contra os artefatos desta mudança (`proposal.md`, `specs/*/spec.md`, `design.md`, `tasks.md`). Nesta entrega, todos os itens abaixo descrevem o que precisa ser verdade quando a fatia inicial (ou uma fatia subsequente) for implementada — nenhum código foi produzido ainda.

## 1. Alinhamento com `proposal.md`

- [ ] A implementação resolve o problema descrito (tradução da especificação v4.0 em artefatos rastreáveis por capability) sem reescrever os capítulos-fonte.
- [ ] Os goals do proposal foram atendidos: 98 requisitos traduzidos, `design.md` completo, `tasks.md` sequencial.
- [ ] Nenhum item de "Non-Goals" foi implementado sem aprovação explícita (nenhum código, nenhuma resolução de P-01–P-11, nenhuma execução de cenário de teste).
- [ ] Os atores impactados (Engenharia, Produto, Arquitetura, Segurança, QA, Plataforma/SRE, Financeiro, liderança técnica) foram considerados nas capabilities correspondentes.
- [ ] As 22 invariantes obrigatórias citadas em "Business Rules" continuam respeitadas em todas as specs.
- [ ] Os critérios de sucesso do proposal são verificáveis (contagem de requirements por ID, cobertura das seções do template em `design.md`, formato de `tasks.md`).
- [ ] As premissas (change-id único cobrindo toda a especificação; QUA/DEC tratados como capabilities de processo) continuam válidas.
- [ ] As perguntas em aberto (P-01 a P-11) permanecem explicitamente não resolvidas, não "resolvidas por omissão".

## 2. Conformidade com specs (`specs/*/spec.md`)

- [ ] Cada um dos 98 requisitos da matriz-fonte (`docs/MATRIZ_RASTREABILIDADE.csv`) tem um `### Requirement` correspondente, identificável pelo ID de origem, em exatamente uma capability.
- [ ] Cada `SHALL`/`MUST` tem cenário(s) `GIVEN/WHEN/THEN` correspondente(s).
- [ ] Nenhum comportamento externo à especificação-fonte foi adicionado sem justificativa (ex.: nenhuma tecnologia, algoritmo ou regra comercial inventada).
- [ ] Cenários cobrem caminho feliz, entrada inválida, permissão insuficiente, estado conflitante, falha de integração externa e concorrência sempre que o texto-fonte os descreve (especialmente em `execucao-e-integracoes`, `contratos-e-financeiro`, `configuracao-e-seguranca`).
- [ ] Valores numéricos e exemplos sintéticos (TTL, SLOs, exemplos financeiros de FIN-09, defaults de polling/webhook) foram preservados exatamente como no texto-fonte.
- [ ] `qualidade-e-aceite` e `decisoes-e-governanca` preservam o estado "NÃO EXECUTADO"/"NÃO QUALIFICADO" e os IDs de ADR/P-xx/GAP-xx sem reinterpretação.

## 3. Conformidade com `design.md`

- [ ] A arquitetura implementada segue as 5 decisões técnicas registradas (despacho direto SYNC, persistência por domínio/célula, UUIDv7 universal, credencial explícita, Redis opcional).
- [ ] Alternativas descartadas (gRPC/broker no SYNC, cluster global, workflow engine arbitrário, journal alternativo) não são reintroduzidas sem uma nova decisão registrada.
- [ ] Os fluxos principais (SYNC direto, ASYNC com callback/polling, agregação/composição) e os fluxos de erro (timeout Órbita↔Cometa, retorno tardio pós-SLA, Libra indisponível) foram implementados conforme descrito.
- [ ] Estratégia de autenticação/autorização (OAuth2/OIDC/mTLS, `hub_protocol_reader`) foi respeitada.
- [ ] Estratégia de persistência (hub_control/hub_core/hub_finance por célula, outbox/inbox) foi respeitada.
- [ ] Estratégia de observabilidade (métricas de baixa cardinalidade, alertas da tabela OPE-06) foi implementada.
- [ ] Riscos remanescentes listados em `design.md` (GAP-02, GAP-05, GAP-09, GAP-12, GAP-15) têm plano de mitigação ativo, não apenas citado.

## 4. Qualidade de `tasks.md`

- [ ] Tarefas do grupo 1–7 (fatia inicial) foram concluídas ou explicitamente canceladas antes de iniciar os grupos 8–9 (slices 2 e 3).
- [ ] Cada tarefa concluída tem evidência de validação do tipo declarado (unit/integration/contract/observability/performance/manual).
- [ ] Dependências entre tarefas (`Depends on`) foram respeitadas na ordem de execução.
- [ ] Tarefas de observabilidade (7.1) e de qualificação de resiliência (7.2, 7.3) não foram puladas antes de avançar para o gate G1/G2.
- [ ] Tarefa 10.3 (decisões pendentes) foi revisitada antes de qualquer promoção a ppd/prd.

## 5. Testes

- [ ] Testes unitários cobrem os invariantes transacionais de cada domínio (Órbita, Cometa, Pulsar, Libra).
- [ ] Testes de integração cobrem PostgreSQL/broker/objetos por célula.
- [ ] Testes de contrato existem para o OpenAPI/AsyncAPI da fatia inicial.
- [ ] Testes end-to-end cobrem SYNC direto e ASYNC com callback/polling concorrentes.
- [ ] Testes de segurança cobrem isolamento de tenant, SSRF e credencial cruzada.
- [ ] Testes de performance reproduzem os perfis de OPE-01 (1.000 admissões/s, rajada 2.000/s, 2.000 GETs/s, 8h a 50%).
- [ ] Nenhum teste desta lista foi marcado como "passou" sem evidência registrada — a matriz-fonte permanece NÃO EXECUTADO até haver essa evidência.

## 6. API, contratos e compatibilidade

- [ ] Contratos HTTP/eventos estão documentados em OpenAPI/AsyncAPI (tarefas 2.1–2.3).
- [ ] GET de protocolo e corpo de webhook usam exatamente o mesmo schema/hash (COM-05), verificado por comparação automatizada.
- [ ] Idempotência (`Idempotency-Key`, UUIDv7) foi implementada e testada com payload repetido/diferente.
- [ ] Mensagens de erro não vazam segredo, dado de outro tenant ou detalhe de infraestrutura interna.

## 7. Dados e migração

- [ ] Schemas de `hub_control`/`hub_core`/`hub_finance` seguem o modelo lógico de `persistencia-e-dados` (DAD-02), sem FK cruzada entre domínios.
- [ ] Índices previstos em DAD-03 (tenant/protocolo, tenant/chave idempotente, etc.) estão presentes.
- [ ] Retenção por classe de dado (DAD-06) está configurada e documentada, mesmo que os prazos finais ainda dependam de P-04.
- [ ] Realocação de célula (se implementada) segue o workflow de DAD-11, sem duas autoridades ativas simultaneamente.

## 8. Segurança e privacidade

- [ ] Autenticação (OAuth2/OIDC/mTLS) e autorização (perfis de SEG-01) foram validadas em cenários positivos e negativos.
- [ ] Isolamento de tenant foi testado por troca de ID/header/body, não apenas por navegação da interface.
- [ ] Segredos nunca aparecem em logs, eventos ou configuração em claro.
- [ ] SSRF é mitigado em todo destino configurável (webhook de cliente, callback de provedor).
- [ ] Perfil `hub_protocol_reader` é individual, auditado e não concede automaticamente segredo/edição/replay.

## 9. Observabilidade e operação

- [ ] Logs, métricas e traces relevantes foram adicionados conforme a seção "Observability" de `design.md`.
- [ ] Alertas da tabela OPE-06 estão configurados com responsável e ação associados.
- [ ] Runbooks obrigatórios (provedor indisponível, banco indisponível, broker indisponível, DLQ, etc.) foram redigidos antes do gate G3.
- [ ] Degradação/fallback foi testado conforme a matriz de dependências (OPE-13): Redis, Atlas, writer PostgreSQL, banco financeiro, SNS/SQS, S3, cofre, IdP.

## 10. Revisão final

- [ ] A implementação está alinhada com o problema (interoperabilidade multi-tenant rastreável), não com uma solução presumida além do texto-fonte.
- [ ] Nenhum requisito foi inventado durante a implementação além dos 98 já rastreados.
- [ ] Riscos remanescentes (`risk-matrix.md`) foram revisitados e, quando aceitos, têm responsável nominal.
- [ ] Decisões pendentes (P-01–P-11) resolvidas até o momento estão registradas com data e responsável; as não resolvidas continuam bloqueando apenas o escopo/gate correspondente.
- [ ] O conjunto de artefatos (`proposal.md`, `specs/`, `design.md`, `tasks.md`, `risk-matrix.md`) pode ser revisado por humano ou agente sem depender de contexto implícito desta conversa.
