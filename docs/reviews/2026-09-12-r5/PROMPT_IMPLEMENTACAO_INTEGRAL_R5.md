# Implementação e qualificação integral — AI Hub R5

Atue como responsável técnico pela execução desta rodada. Implemente o planejamento R5 e conclua as obrigações herdadas aplicáveis. Trabalhe de forma autônoma nas ações locais autorizadas, resolvendo código, ferramentas, fixtures e integração até o fechamento verificável. Não substitua implementação por novos relatórios de pendência.

## 1. Entrada obrigatória e estado real

Repositório: https://github.com/leandroclf/ai-hub.
Snapshot auditado: `f87ce33034ae29c9431b1910dcc6a633b545e330`, de 12/09/2026.

Leia nesta ordem:
1. AGENTS.md e demais instruções aplicáveis nos diretórios.
2. docs/openspec-docs/ e os templates/checklists do projeto.
3. docs/reviews/2026-09-12-r5/00-LEIA-PRIMEIRO.md, relatório, plano, arquitetura, gates, aceite e ponte R4.
4. Matrizes de requisitos/cenários, reavaliações de R2/R3/R4 e achados R5.
5. Os seis changes `openspec/changes/r5-*`, incluindo proposal, design, specs, tasks e rastreabilidade.
6. Requisitos herdados v4/R2/R3/R4 vinculados e evidências reais correspondentes.

Inspecione HEAD, branch, git status/diff e conteúdo não commitado. Se HEAD avançou, revalide os achados contra esse delta antes de editar. Preserve alterações preexistentes e correções comprovadas. Não use reset/clean destrutivos.

A baseline remota auditada contém **201 requisitos e 732 cenários**. A R5 adiciona **17 requisitos e 51 cenários**: união **218/783**. Recalcule as contagens a partir das specs se o conteúdo mudou; não force contagem por remoção de cenários.

A R4 regenerada de 10/09 tinha quatro requisitos RUN que não entraram no repositório. Estão mapeados na R5. Não copiar o quinto change antigo junto com R5 e duplicar obrigações; use a ponte do documento 06. Preserve histórico v4/R2/R3/R4.

## 2. Autonomia e segurança do trabalho

Está autorizado a implementar código, contratos, UI, testes, migrações aditivas, scripts, manifests, documentação e ferramentas locais necessárias. Prepare fixtures sintéticas e execute testes/carga/falhas controladas no ambiente de laboratório. Tome decisões técnicas reversíveis compatíveis com requisitos, documentando motivo, alternativa e consequência.

Não peça confirmação para tarefas locais já autorizadas. Não pare apenas porque o relatório anterior registrou teste não executado: investigue o impedimento concreto e resolva-o quando possível.

Não invente aprovações comerciais/normativas, não apague dados preexistentes, não faça operações financeiras reais, não provisione recursos pagos e não faça push/merge/deploy remoto nesta missão. Respeite permissões de plataforma; indisponibilidade de acesso não autoriza contornar controles.

O AGENTS exige **no máximo um ecossistema Compose do AI Hub ativo por vez**. Antes de iniciar/substituir, execute `docker compose ls` e `docker ps -a`. Identifique o projeto oficial, use nome estável, pare somente o Hub identificado quando necessário e confirme encerramento antes de substituir. Preserve volumes/dados e componentes de terceiros. Não use prune amplo ou curingas destrutivos. Registre/limpe somente recursos exatos dispensáveis de sua execução.

Kind independente já existe. Preserve-o e use esse perfil para provar independência de Compose; não atribua HA regional ao laboratório local.

## 3. Plano executável e evidência

Crie `docs/reviews/2026-09-12-r5/implementation/` com:
- EXECUTION_PLAN.md: fatias, dependências e próximo passo concreto.
- REQUIREMENTS_STATUS.csv e SCENARIO_RESULTS.csv: resultado por cenário com proveniência.
- FINDINGS_STATUS.md: achados, correções, regressões e prova de fechamento.
- DECISIONS_AND_BLOCKERS.md: apenas decisões/bloqueios reais, escopo, responsável e alternativa.
- CHECKPOINT.md: retomada exata, comandos e recursos ativos.
- EVIDENCE_INDEX.md: logs saneados, hashes, versões, imagens e comandos.
- FINAL_REPORT.md: conclusão derivada das provas, coerente com as matrizes.

Cada evidência identifica SHA, hash do working tree quando houver alterações, digest de imagem/base, toolchain, ambiente, comando, cenário, resultado e oráculo. HEAD sozinho não identifica código local alterado. Evidência histórica é reaproveitável apenas com análise explícita de compatibilidade; nunca promover por título ou existência de arquivo.

## 4. Prepare o laboratório e os gates

Inventarie Go/Node, Docker/Compose, PostgreSQL, LocalStack, OIDC, kubectl/kind, OpenSpec e Playwright. Instale versões fixadas compatíveis quando necessário e permitido. OpenSpec 1.12.0 validou a baseline nesta auditoria sem config.yaml; não trate ausência de config como impedimento automático.

O Dockerfile atual usa Go 1.26 e base Nginx atualizada; as provas próprias da auditoria usaram Go 1.24.13. Construa e teste a **imagem candidata efetiva**, validando tag/digest e arquitetura. Corrija pins/reprodutibilidade com evidência, sem apenas trocar por latest.

Separe unitários e integração obrigatória. Um gate integrado deve falhar quando dependências necessárias estiverem ausentes; não contar t.Skip como PASS e não remover assertions para passar. Corrija daemon/socket/fixture/seed/reconciliação quando permitido. Se houver impedimento de plataforma incontornável, documente erro exato, prepare gate reproduzível e conclua todo trabalho independente.

## 5. Prioridades de implementação

### A. Autorizações e custódia

1. **R5-SEG-01:** callback de operação desconhecida não pode gravar inbox com assinatura arbitrária. Verifique origem antes da custódia, sem depender da existência da operação. Prove caminho completo ingress→handler→DB. Defina chave/versão por conta sem distribuir segredo raiz a todos os provedores. Dedupe por evento estável, quota por escopo, rotação e disposição de órfãos permanentes.
2. **R5-SEG-02:** resposta 403/409 do Atlas não pode ser mascarada por cache. Tipar erros e publicar semântica de lease/revogação; fallback somente para indisponibilidade transitória elegível. Resolver caminho quente sem chamada remota obrigatória por hit, respeitando invalidação.
3. **R5-SEG-03:** migrar stores/workers para tenant-scoped transactions usando helper existente e roles não proprietárias. Inventariar tabelas auxiliares/grants e admin global nominal. Prova negativa exige fixtures A/B existentes, controle positivo e escrita cruzada recusada, antes de rollback.
4. **R5-DAD-04:** Libra deve conservar payload bruto ou referência durável recuperável antes de ACK em quarentena. Hash isolado não serve. Replay autorizado reprocessa fato, nunca reenviar SUBMIT.
5. **R5-EXE-05:** gate de topologia confirma todas as assinaturas obrigatórias antes do relay. Testar namespace limpo, consumidor tardio e assinatura removida; Publish bem-sucedido não comprova entrega aos consumidores necessários.

### B. Execução e produtos

6. **R5-EXE-01:** cercar retry_until antes de claim/dispatch/egress conforme modelo; delivered=true depois do prazo não valida efeito indevido. Preservar correção de polling que usa StepDeadline. Harmonizar TTL de primeira falha, prazo do cliente/provedor/tentativa e texto administrativo; tratar comandos legados.
7. **R5-EXE-02:** reservar vaga do produto e posse da etapa atomicamente entre réplicas. Testar max_parallel=1 com dois publishers e crash entre RUNNING/publicação. Reclaim reutiliza a vaga; não bloquear a própria recuperação. O helper ExecuteDAG antigo não é prova do runtime persistido: remover caminho morto ou qualificá-lo sem duplicar implementação.
8. **R5-EXE-03:** cada etapa recebe conta/rota/binding/compra/perfil/prazo próprios e hash verificável. Não clonar conta do produto para serviços distintos. Provar dois provedores e separar venda do produto de custo de cada efeito. Consolidação segue contrato publicado.
9. **R5-EXE-04:** compensações persistidas em ordem causal reversa, com prazo próprio; continuam mesmo após fechar resposta ao cliente. Não depender do client_deadline ou contexto HTTP já cancelado. UNKNOWN de compensação exige reconciliação.
10. **R5-DAD-01:** eliminar perda numérica em operationFact.ResponseBody/any, consolidação, dispatch/resultados e frontend. Provar 9007199254740993 e decimais exatos atravessando a cadeia inteira. TransformJSON já foi corrigido: preserve seus testes.

### C. Financeiro, capacidade e console

11. **R5-DAD-03:** captura usa valor efetivo contratado e libera diferença da reserva sem perder hold de incerteza. Ligar produtores de watermark/completude; período não fecha com mensagens pendentes. Fato tardio gera disposição/ajuste consultável e não altera exportação antiga. Provar compra/venda por etapa, franquia, compensação e CLIENT_DIRECT.
12. **R5-OPE-01:** rota ativa requer política qualificada de capacidade; não desabilitar controle por campo vazio. Falha de settlement de permit é obrigação durável recuperável, não só log. Não reciclar lease por timeout sem prova de transporte/efeito. Medir custo de Acquire com histórico grande e justiça entre tenants/pods.
13. **R5-UX-01:** pesquisa/paginação server-side de referências, sem teto funcional de 1000. Validar DTOs runtime por jornada. Persistir identidade de intenção para timeout/reload com consulta de recibo e escopo/payload hash; nunca guardar segredos no browser. Concluir configuração autorizada de capacidade/qualificação e jornadas de operador. Visitar menus não equivale a provar efeito.

### D. Operação e qualificação

14. **R5-DAD-02:** restore populado com plano/etapas/compensações, UNKNOWN, reserva, callbacks e múltiplas versões de objeto. Reconciliação no ambiente restaurado, comparando efeitos/ledger antes e depois da retomada. Testar referências, pins/tombstones e barreira de tráfego. Count/digest de cópia e uma consulta ao simulador não encerram restore.
15. **R5-OPE-02:** promoção deve bloquear sem manifesto íntegro vinculado a SHA/digests/toolchain e aprovações verificáveis do pipeline autorizado. Strings PASS e nomes de aprovação não são prova. Não promover imagem diferente da testada.
16. **R5-OPE-03:** preservar kind independente e completar perfis local/dev/hom/ppd/prd com isolamento/identidade/dados/observabilidade. Planejar e implementar IaC sem provisionamento pago nesta missão; qualificar PV/backup, drains, nós/placement/autoscaling e SLO por tráfego sob falhas. EmptyDir local não é durabilidade de produção.
17. **R5-QUA-01:** regenerar inventários a partir das specs, incluir ponte dos quatro R4 ausentes e exigir proveniência por resultado. Fechar lacunas históricas por cenário, não pela contagem de linhas associadas a logs.

## 6. Preserve avanços e conclua também as 25 fatias herdadas

Execute B-R5-01…B-R5-25 do plano e tasks R5-06. Os 17 requisitos novos refinam riscos, não substituem v4/R2/R3/R4.

Preservar: JSON estrito, L1 privado, migração histórica restaurada, adapter registry REST, destino/representação congelados, FileRefs, produtos persistidos, OIDC/MFA, DTOs administrativos corrigidos e kind independente.

Serviços SYNC elegíveis continuam diretos com resposta final e UUIDv7; não converter silenciosamente para fila. ASYNC sobre provedor síncrono, polling/callback e GET local do resultado devem permanecer. Modalidade de produto deve ser elegível na publicação e API; não publicar capacidade que runtime recusa por surpresa.

Topologia, UNKNOWN, pressão adaptativa, financeiro, objetos, retenção, administração, isolamento e monitoramento têm requisitos herdados além dos probes desta rodada. Fechar cada fatia pelo contrato completo e oráculo correspondente, sem recomeçar trabalho já demonstrado.

## 7. Ciclo de execução por fatia

1. Ler implementação real e cenário, reproduzir defeito ou comprovar correção posterior.
2. Fixar teste/oráculo independente e decisão técnica compatível.
3. Implementar dados, API, worker, configuração e UI necessários ao fluxo.
4. Executar unitários, integração e falha/concorrência pertinentes.
5. Corrigir até satisfazer invariantes e demonstrar migração/rollback afetados.
6. Atualizar evidência/status e avançar para a próxima fatia.

Oráculos externos devem observar efeitos reais do fixture: contador por chave, bytes recebidos, status depois de queda, journal/hold, estado após restart e acesso negado com controle positivo. Um mock que retorna a expectativa do próprio código não prova a fronteira distribuída.

Após correções, executar gate final: Go normal/race/vet, frontend, OpenSpec strict, integrações obrigatórias, browser, Compose/kind, upgrade/restore/carga pertinentes. Teste documental pode ter inspeção verificável; cenário de runtime exige execução. Não ampliar testes sem risco concreto depois de prova suficiente.

## 8. Decisões externas e limites de conclusão

D-01…D-07/P-01…P-11 continuam com responsáveis reais. Reutilize defaults sintéticos R3 somente no escopo autorizado. Não invente preços, demanda, SLO, retenção ou homologação comercial para declarar concluído. Prepare configurações/fixtures e deixe gate externo específico identificado.

T-R2-01 requer prova da janela entre decisão e commit e compatibilidade normativa. Não substituir confirmação durável no prazo por checagem pré-commit sem decisão explícita. “Autonomia” não autoriza afirmar ausência absoluta de perda/indisponibilidade sob qualquer falha.

## 9. Critério de encerramento

A implementação integral só pode ser declarada quando o escopo técnico aplicável estiver comprovado, P0 fechado, zero skip obrigatório, jornadas reais funcionando, migração/restore e OpenSpec válidos, evidência coerente com o artefato e diff revisado.

Não terminar apenas com plano, helpers desconectados ou nova rodada de specs. Se houver interrupção de sessão, salvar checkpoint exato e continuar a missão ao retomar. Se restar impedimento incontornável, antes conclua todo trabalho independente e informe recurso/permissão/decisão exatos, tentativas e gate reexecutável. A conclusão será parcial e delimitada, nunca sucesso fabricado.

Entregue resumo das correções, provas executadas, limitações materiais e caminhos dos artefatos. Não faça push, merge ou implantação remota.
