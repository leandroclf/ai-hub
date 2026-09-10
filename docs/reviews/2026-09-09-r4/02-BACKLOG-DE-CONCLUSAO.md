# Backlog integrado e ordem de execução

A R4 adiciona correções específicas; não reinicia o projeto nem elimina tarefas v4/R2/R3.
O escopo de conclusão inclui a união de 201 requisitos e 732 cenários.
Correções pontuais comprovadas devem ser preservadas e não refeitas desnecessariamente.

## Sequência de engenharia
1. Inventariar HEAD/working tree e ambiente, obedecer AGENTS e preparar gates de teste.
2. Corrigir upgrade de migrações antes de usar dados existentes.
3. Fixar contratos JSON e ingresso/custódia de callbacks com worker autônomo.
4. Fechar adapter real, UNKNOWN/fencing e topologia; só então afirmar jornada externa segura.
5. Separar relógios e investigar T-R2-01 sem enfraquecer confirmação durável.
6. Ligar pressão adaptativa/pools/cache/projeções ao envio real.
7. Concluir DAG, política efetiva, representação, finanças e objetos.
8. Concluir UI com contratos backend já efetivos, testando efeitos no navegador.
9. Qualificar Compose/kind, sinais, escala, falhas e restore, mantendo um ecossistema Compose ativo.
10. Consolidar evidência de toda baseline; registrar gates comerciais/regionalmente externos separadamente.

O trabalho de UI e harness pode começar cedo; fechamento depende da jornada real.
Não existe dependência que obrigue parar todas as tarefas até D-01…D-07 estarem aprovadas.
T-R2-01 não autoriza reduzir EXE-11. Manter o gate normativo específico aberto se solução
exigir alteração de requisito; avançar nos demais.

## Fatias herdadas obrigatórias
| Tarefa | Requisito | Estado atual e próximo passo | Responsável |
|---|---|---|---|
| B-R4-01 | R3-EXE-01 | API_KEY agora atravessa modelo/projeção/executor/poller; gate synthetic-provider persiste. Fechar adapter real e homologação. | Core e Integrações |
| B-R4-02 | R3-EXE-02 | Capability e recibo antes de ACK foram adicionados. Nova inbox tem lacunas de autenticação/deduplicação e recuperador; rota segue sob JWT Hub. | Core e Integrações |
| B-R4-03 | R3-EXE-03 | PrepareSubmission/UNKNOWN sem recuperador geral e sem fencing de egress continuam. Inbox não resolve incerteza de SUBMIT. | Core e Integrações |
| B-R4-04 | R3-EXE-04 | Relógios, exclusividade polling/callback e T-R2-01 permanecem; sem mudança no finalizador. | Core e Integrações |
| B-R4-05 | R3-EXE-05 | Topologia/outboxes/consumidor financeiro não mudaram. Quarentena financeira continua só hash. | Core e Integrações |
| B-R4-06 | R3-CAT-01 | Precisão grande/enum simples corrigidos e reproduzidos positivamente. Novas violações de tipos/EOF/schema ainda abertas. | Produto e Core |
| B-R4-07 | R3-CAT-02 | Handler de admissão e finalizer não mudaram; política efetiva/versão e OutputMapping pendentes. | Produto e Core |
| B-R4-08 | R3-CAT-03 | Executor de DAG não conectado; PlanDAG de catálogo não é execução de produto. | Produto e Core |
| B-R4-09 | R3-CAT-04 | Teto de 100 removido por paginação; todas as ofertas ainda materializadas antes do filtro e Atlas consultado por pedido. | Produto e Core |
| B-R4-10 | R3-INT-01 | Capacity continua sem ligação a Execute/requestPoll; código desse controlador e consumidores permanece igual. | Integrações e Plataforma |
| B-R4-11 | R3-INT-02 | Tokens removidos do Redis; lock por chave existe. L1 ainda depende do cofre; crescimento de locks não limitado. | Integrações e Plataforma |
| B-R4-12 | R3-INT-03 | Transports por chamada permanecem; arquivo egress não mudou e chamadas NewClient continuam. | Integrações e Plataforma |
| B-R4-13 | R3-ADM-01 | Papel hub_protocol_reader e MFA agora obrigatórios no admin; avanço de autorização constatado. Requalificar política de aplicação/mascaramento e browser sem afirmar bypass corrigido ainda existente. | Frontend e Segurança |
| B-R4-14 | R3-ADM-02 | Nenhuma alteração no frontend; IDs de entrega/rotas SLA/reconcile continuam incompatíveis. | Frontend e Segurança |
| B-R4-15 | R3-ADM-03 | Frontend/handlers financeiros não mudaram; payload/tenant/datas e fluxo idempotente continuam pendentes. | Frontend e Segurança |
| B-R4-16 | R3-ADM-04 | Catálogo UI não mudou; multimodalidade, campos auth, fuso e lookups continuam pendentes. | Frontend e Segurança |
| B-R4-17 | R3-FIN-01 | Financeiro não mudou; reserva versus efetivo e franquia pré-efeito precisam fechamento. | Financeiro e Core |
| B-R4-18 | R3-FIN-02 | Financeiro/contratos econômicos não mudaram; incidência e fechamento integrados continuam pendentes. | Financeiro e Core |
| B-R4-19 | R3-FIN-03 | Pulsar não mudou; seleção de destinos atuais em vez de snapshot continua. | Financeiro e Core |
| B-R4-20 | R3-OPE-01 | Objetos/admissão não mudaram; integração de FileRefs/resultados no adapter permanece pendente. | Plataforma, Dados e SRE |
| B-R4-21 | R3-OPE-02 | Kind/overlays não mudaram; dependências ligadas a Compose não equivalem a cluster completo. | Plataforma, Dados e SRE |
| B-R4-22 | R3-OPE-03 | Autoscaling/placement/drenagem não receberam fechamento neste delta; exigem ensaio. | Plataforma, Dados e SRE |
| B-R4-23 | R3-OPE-04 | Migrações de callback adicionadas, mas alteração de migração histórica cria novo risco de upgrade. RLS/restore sem prova integrada atual. | Plataforma, Dados e SRE |
| B-R4-24 | R3-OPE-05 | UI SLA/telemetria não mudaram; sinais reais de domínio e consulta bilateral pendentes. | Plataforma, Dados e SRE |
| B-R4-25 | R3-QUA-01 | OpenSpec strict atual 17/17 e build/race passam nesta auditoria; 18 testes skip. Matriz omite 416 cenários v4 e relatórios são contraditórios. | Engenharia e Qualidade |

## Aceite por fatia
Contrato publicado → admissão autenticada → tentativa/concessão duráveis → provedor externo →
recibo/fato → final do Hub → GET/webhook → fato econômico → consulta administrativa.
Nem toda fatia percorre todos os passos, mas seu efeito precisa ser visível nos consumidores pertinentes.

Uma biblioteca de capacidade sem chamada em Execute não satisfaz controle adaptativo.
Uma tabela de inbox sem worker não satisfaz recuperação.
Um ledger sem produtores de unidade/corte não satisfaz faturamento.
Uma página sem endpoint não satisfaz operação administrativa.
Uma spec válida não significa implementação conforme.

## Casos obrigatórios do console
- Duas entregas do mesmo protocolo: detalhe por delivery_id e redelivery da entrega escolhida.
- Financeiro com tenant selecionado: comando aceito pelo DTO real, ator do token e datas válidas.
- Timeout após commit: retry da mesma intenção, sem novo ajuste financeiro.
- Conta OAuth/API_KEY/mTLS: campos condicionais e referências corretas; nunca revelar segredo.
- Oferta multimodal: SYNC/ASYNC/AUTO preservados após editar/salvar/reabrir.
- Vigência UTC em America/Sao_Paulo: instante preservado e fuso exibido.
- Catálogo acima de 100 itens: pesquisa/paginação acessível e resolução seletiva no runtime.
- Protocolo: GET final sem provedor, reconciliação autorizada e timeline real.
- SLA: rota efetiva, relógio cliente-Hub e Hub-provedor separados, filtros e evidência de quebra.
- Consumidor contra admin e tenant/aplicação cruzados: API recusa, mesmo com UI manipulada.
- Operador nominal global: MFA/papel/escopo, leitura auditada e política de mascaramento explícita.
- Acessibilidade: teclado, foco, labels e estados carregando/vazio/erro/conflito/sucesso confirmado.

## Segurança de retomada e migração
Não executar SUBMIT às cegas para recuperar UNKNOWN.
Não reabrir final expirado para acomodar callback tardio.
Não liberar reserva sem evidência positiva de ausência de efeito.
Não editar ledger/checksum histórico ou apagar volume para fazer gate passar.
Não usar token expirado/revogado para mascarar indisponibilidade do cofre.
