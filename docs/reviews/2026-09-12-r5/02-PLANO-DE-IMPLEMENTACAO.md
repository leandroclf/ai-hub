# Plano integrado de implementação

Sem cronograma artificial; ordem por dependência e risco.

| Ordem | Fatia | Saída verificável |
|---|---|---|
| 1 | Baseline/laboratório | SHA/diff/instruções, ferramentas fixadas, um Compose e fixtures identificadas |
| 2 | R5-SEG-01/02 e DAD-04 | Autenticação antes de inbox, negação respeitada e quarentena antes de ACK |
| 3 | R5-SEG-03 e EXE-01/05 | Stores escopados, prazos pré-I/O e topologia confirmada |
| 4 | R5-EXE-02/03/04 e DAD-01 | Slots atômicos, contratos por etapa, compensação independente e precisão |
| 5 | R5-DAD-03 e OPE-01 | Liquidação/corte reais e permits recuperáveis |
| 6 | R5-UX-01 | Jornadas completas sobre APIs corrigidas, intenção durável e busca paginada |
| 7 | R5-DAD-02 e OPE-02/03 | Restore populado, promoção por artefato e perfis operacionais |
| 8 | R5-QUA-01 | Matriz integral por prova e relatório consistente |

UI e harness podem evoluir cedo; fechamento depende da API/autoridade real.
Cada requisito possui tarefas contrato/reprodução, implementação integrada e qualificação. Decompor por transação/endpoint/worker se necessário, preservando IDs.

## Fatias herdadas obrigatórias
| Tarefa | Requisito | Avanço a preservar e próximo passo |
|---|---|---|
| B-R5-01 | R3-EXE-01 | rest-json-v1 instalado e teste HTTP independente do simulador; preservar registry e homologar provedores/contratos reais e capacidades por etapa. |
| B-R5-02 | R3-EXE-02 | Rota pública/HMAC/inbox/worker implementados. F-R5-01 demonstra caminho órfão sem verificação prévia; escopo/rotação/retention continuam exigindo prova. |
| B-R5-03 | R3-EXE-03 | Fencing de submissão e reconciliação por status implementados. UNKNOWN sem correlação não é reenviado; falta reconciliação automática com oráculo por chave quando suportado e fechamento de efeito/financeiro sob crash. |
| B-R5-04 | R3-EXE-04 | Polling saudável agora usa StepDeadline; TTL de primeira falha persistido. Barreira retry_until pré-despacho e semântica UI ainda pendentes; T-R2-01 não presumidamente resolvido. |
| B-R5-05 | R3-EXE-05 | Worker Cometa/Pulsar conserva quarentena. Libra ainda só hash; startup dos publishers não confirma todas as assinaturas obrigatórias. |
| B-R5-06 | R3-CAT-01 | Probes JSON anteriores PASS no novo SHA. Não reimplementar TransformJSON. Perda numérica no consumidor de fatos segue em R5-DAD-01. |
| B-R5-07 | R3-CAT-02 | Perfil técnico/output mapping/representação e versão solicitada avançaram. Produtos precisam snapshot por etapa e consolidação efetiva, preservar GET/webhook congelados. |
| B-R5-08 | R3-CAT-03 | Plano/etapas/compensações persistidos e produto HTTP local executados. Concorrência de slots, recuperação pós-RUNNING, rotas por etapa e compensação fora do prazo do cliente são gaps atuais. |
| B-R5-09 | R3-CAT-04 | Oferta seletiva indexada implementada e cache fallback existe. Fallback em 403 reproduzido e remoto antes do cache mantém custo no caminho quente. |
| B-R5-10 | R3-INT-01 | Capacidade agora ligada aos transportes e métricas de feedback. Qualificar política obrigatória, settlement recuperável e custo/justiça em escala, sem tratar implementação como ausente. |
| B-R5-11 | R3-INT-02 | L1 antes do cofre, locks limitados/canceláveis, revogação e mTLS têm código/testes. Probe adaptado PASS. Homologação de rotação multi-réplica/provedor continua gate específico. |
| B-R5-12 | R3-INT-03 | Pool de egress por origem implementado. Preservar testes SSRF/mTLS e qualificar budgets end-to-end, memória e concorrência global sob carga real. |
| B-R5-13 | R3-ADM-01 | MFA/papel nominal/escopo e reason auditados em admin; smokes de tenant/application. Adoção RLS e escopo de endpoints auxiliares precisam completar defesa. |
| B-R5-14 | R3-ADM-02 | OperationsPage corrige delivery ID e APIs de SLA/reconcile existem. Preservar melhorias; evoluir cenários de autorização, conflito, escala e frescor com dados reais. |
| B-R5-15 | R3-ADM-03 | FinancePage corrige tenant/DTO/datas/ator e aprovação distinta. Falta intenção persistida sob duplo timeout/reload, IDs grandes e fechamento ligado a produtores. |
| B-R5-16 | R3-ADM-04 | Multimodalidade, auth condicional, fuso, importação/simulação e mensagens de erro avançaram. Lookup falha acima de 1000 e DTO runtime ainda genérico. |
| B-R5-17 | R3-FIN-01 | Ledger/dedupe e testes UNKNOWN avançaram. Captura não demonstra diferença reserva versus valor efetivo; franquia pré-efeito e compensação exigem prova integral. |
| B-R5-18 | R3-FIN-02 | Incidências SUBMITTED/STATUS por attempt implementadas. SetWatermark apenas chamado em teste; faltam fechamento por completude e disputa tardia operacional. |
| B-R5-19 | R3-FIN-03 | Destinos congelados no aceite e representação no fato final implementados; preservar igualdade de bytes e isolamento. Requalificar retry/rotação/consumo legado sem destino. |
| B-R5-20 | R3-OPE-01 | FileRefs enviados ao adapter, resultado volumoso conservado/vinculado e retenção tem worker. Restore versionado/pins e prova produtiva de acesso a bytes externos continuam. |
| B-R5-21 | R3-OPE-02 | Kind independente com dependências/UI/gateway implementado: não repetir achado de ausência. Laboratório usa recursos efêmeros; ambientes remotos completos/IaC/durabilidade precisam qualificação. |
| B-R5-22 | R3-OPE-03 | HPA/KEDA e perda de pods ensaiados localmente. Não equivalem a escala de nós/dados/placement nem justiça/RPO sob falha de zona. |
| B-R5-23 | R3-OPE-04 | RLS helper sem adoção nos stores e prova negativa com rollback permanecem. Restore digests melhora cópia, não comprova retomada populada reconciliada. |
| B-R5-24 | R3-OPE-05 | SLA API e sinais reais/outbox/alertas têm novas provas. Qualificar coortes/budgets bilaterais, métricas de obrigação e SLO sob carga/falha por contrato. |
| B-R5-25 | R3-QUA-01 | Inventário remoto 201/732 e 231 linhas associadas/501 não qualificadas registrados honestamente. Ponte da R4 regenerada ausente e proveniência por SHA/digest exigem fechamento. |

## Definition of Done por fatia
Cenário reproduzido → regra compatível → implementação ligada → teste de erro/concorrência → migração/rollback → evidência do artefato.
P0 fechado só com prova do risco específico; unitário isolado não fecha propriedade distribuída. NOT_APPLICABLE exige justificativa verificável, não atalho.
Resultado esperado por classe: falha de ambiente não vira PASS; contrato comercial não aprovado não impede fixture técnica; impossibilidade normativa não autoriza inventar garantia.
