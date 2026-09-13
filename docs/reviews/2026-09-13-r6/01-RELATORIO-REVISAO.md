# Revisão detalhada R6 — implementação atual do AI Hub
Snapshot [a540b40007fe6b8ed523e17afe00e96ff8f8ad50](https://github.com/leandroclf/ai-hub/commit/a540b40007fe6b8ed523e17afe00e96ff8f8ad50), de 13/09/2026 12:36:24 UTC.
Comparação com f87ce33034ae29c9431b1910dcc6a633b545e330: **118 arquivos, 7.600 inserções, 177 remoções**, incluindo docs/evidências.

## Parecer técnico
A implementação avançou e o próprio relatório R5 delimita a entrega como parcial. A próxima rodada deve fechar invariantes entre componentes e a qualidade das provas. Build e suíte verde não demonstram automaticamente SLA, isolamento, custódia ou recuperação distribuída.

## Melhorias confirmadas por leitura e testes locais
- Callback órfão verifica assinatura antes da gravação; dedupe por corpo/conta/operação e quota por conta foram adicionados.
- Cache de oferta atende caminho quente, expira projeção e não mascara 403/409 após resolução autoritativa; três testes de oferta passam na suíte.
- Inteiro grande no consumidor de fatos agora é preservado: probe anterior PASS. IDs financeiros passam a ter representação textual.
- Claim de produto usa transação, lock do plano, contador e lease; snapshots derivados têm hash.
- Compensações têm dependências reversas e prazo próprio, ainda com incompatibilidade com final público.
- Captura financeira foi ligada a ApplyEvent; quarentena passou a guardar payload e tem endpoints de replay.
- EnsureTopology é chamado pelos quatro componentes antes de liberar seus fluxos de mensageria.
- Gate prd recusa manifesto ausente; catálogo introduziu busca remota e guards em parte dos DTOs.

## Evidência histórica e limites desta auditoria
EVIDENCE_INDEX da implementação relata 245 testes sem skip com PostgreSQL/LocalStack/Redis, mas informa que logs brutos não foram salvos. O commit menciona 246; a diferença deve ser reconciliada por logs, sem inferir fraude ou falha dos testes. O script RLS executado historicamente ainda contém a falha de oráculo descrita abaixo.
Nesta auditoria: Go 1.24.13, race com 139 PASS/70 SKIP, vet e build frontend aprovados. PostgreSQL, Docker/kind, OIDC, browser, carga, HA e restore não executados aqui; ferramentas de runtime ausentes. Go local não qualifica Dockerfile Go 1.26. Build frontend reutilizou dependências cujo package-lock é idêntico; não foi npm ci limpo.

## Achados atuais
### F-R6-01 — RLS no caminho real e prova com dados existentes (P0)

**Natureza:** ANALISE_ESTATICA.

A migração 0049 remove o bypass nominal nas tabelas core contempladas. WithTenantTx continua sem chamada produtiva, os DSNs Compose usam hub e o contexto opcional continua fixo por workload. A prova rls-runtime-proof.sh continua executando ROLLBACK das fixtures antes das assertions negativas; a contagem dentro da transação só é impressa. Portanto, a execução histórica do script não prova isolamento de dados existentes nem adoção pelo runtime.

**Motivo da correção:** O Hub SHALL aplicar escopo autenticado por transação e por item de trabalho com roles runtime sem bypass; tabelas sem tenant direto devem ter autorização por relação ou autoridade operacional específica. Testes de isolamento SHALL afirmar existência e acesso próprio, invisibilidade alheia e recusa de escrita cruzada antes de remover fixtures.

**Proposta:** Mapear tabela→owner→role→policy→caminho. Separar migrador, requests e workers globais; estes precisam claims limitados e auditados, não tenant arbitrário fornecido pelo cliente. Adotar contexto LOCAL e provar limpeza em commit/rollback. A migração isolada não autoriza trocar o DSN e derrubar todos os workers. Qualificar migração com bases populadas e credenciais runtime efetivas.

**Fontes:** [hub/migrations/core/0049_strict_runtime_rls.sql:5](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/migrations/core/0049_strict_runtime_rls.sql#L5), [hub/internal/platform/pg/pg.go:39](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/platform/pg/pg.go#L39), [hub/deploy/r2/tests/rls-runtime-proof.sh:17](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/rls-runtime-proof.sh#L17), [hub/deploy/r2/compose.yaml:203](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/compose.yaml#L203)

**Rastreabilidade:** R6-SEG-01 ← R5-SEG-03; responsável Segurança e Core.

### F-R6-02 — Callback aceito sobrevive à rotação e à indisponibilidade (P0)

**Natureza:** ANALISE_ESTATICA.

A assinatura é agora verificada antes da custódia órfã e a identidade não depende do timestamp: preservar a correção. A reconciliação ainda revalida com CALLBACK_INGRESS_KEY atual, não com uma versão de chave congelada. O lock advisory de storeOrphan continua global, e três falhas de apply podem converter recibo em REJECTED. Órfãos sem operação não são selecionados pelo JOIN de reconciliação.

**Motivo da correção:** O Hub SHALL preservar a atestação de autenticação obtida no ingresso e a disposição recuperável de cada callback aceito. Rotação e falha transitória não podem invalidar custódia legítima; esgotamento de retry deve manter obrigação em quarentena reprocessável, distinta de rejeição de contrato.

**Proposta:** Persistir key-id/versão/conta/hash/instante de verificação e proteger sua integridade. Segredo em cofre, jamais no recibo. Reconciliar correlação sem exigir assinatura contra nova chave. Particionar locks/quotas por origem e célula; retenção de órfãos e quarentena com responsável, prazo, alertas e replay autorizado. Evitar descarte automático de dados aceitos.

**Fontes:** [hub/internal/cometa/custody.go:228](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L228), [hub/internal/cometa/custody.go:195](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L195), [hub/internal/cometa/custody.go:385](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L385), [hub/internal/cometa/custody.go:411](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L411)

**Rastreabilidade:** R6-SEG-02 ← R5-SEG-01; responsável Segurança e Core.

### F-R6-03 — Snapshot de etapa coerente em todas as identidades (P0)

**Natureza:** PROBE_REPRODUZIDO.

BuildProductPlan agora gera hash do filho, mas stepSnapshot só substitui SelectedRoute; Account, Binding, contratos e perfil permanecem da oferta base, e command := base conserva ProviderAccountID e EconomicSnapshot. Probe reproduziu rota B com conta A, binding B/A e comando A. Hash válido certifica bytes, não coerência semântica.

**Motivo da correção:** O Hub SHALL materializar cada etapa com rota, conta, binding, perfil, contrato de compra, prazo e incidência coerentes e congelados. Se a oferta não contém referências suficientes para resolver a etapa, a publicação ou admissão SHALL ser recusada antes de qualquer efeito.

**Proposta:** Resolver grafo completo no Atlas durante publicação/admissão; não selecionar primeira rota preenchida ignorando elegibilidade. Incluir dependências versionadas no snapshot do produto; gerar comando a partir do filho completo, não de cópia parcial. Recalcular hash e verificar invariantes no executor. Venda do produto e custos das etapas usam chaves econômicas distintas. Preservar contrato final de consolidação.

**Fontes:** [hub/internal/orbita/product_plan.go:126](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_plan.go#L126), [hub/internal/orbita/product_plan.go:83](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_plan.go#L83), [hub/internal/cometa/executor.go:199](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L199)

**Rastreabilidade:** R6-EXE-01 ← R5-EXE-03; responsável Core e Integrações.

### F-R6-04 — Expiração bloqueia também recuperação DIRECT (P0)

**Natureza:** ANALISE_ESTATICA.

claimIntentMode retorna Expired e RunIntentPublisher evita envio quando true. RunDirectRecovery chama o mesmo claim, mas executa DispatchDirect sem verificar Expired; ClaimDirectIntent também devolve o campo sem consumidor no handler. O guard do modo QUEUED não fecha a fronteira DIRECT.

**Motivo da correção:** O Hub SHALL barrar nova submissão com possibilidade de efeito depois do prazo aplicável em DIRECT, QUEUED e recuperação. A reconciliação de efeito possivelmente já realizado SHALL continuar pela operação de consulta autorizada, sem se confundir com reenvio.

**Proposta:** Centralizar decisão persistida de elegibilidade imediatamente antes de I/O; distinguir SUBMIT de STATUS/consulta por chave. Consumir Expired em todo caminho e finalizar a intenção com disposição recuperável. Manter relógios cliente, provedor, tentativa e TTL desde primeira falha separados; não tratar polling saudável como indisponibilidade.

**Fontes:** [hub/internal/orbita/intents.go:208](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L208), [hub/internal/orbita/intents.go:363](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L363), [hub/internal/orbita/handlers.go:337](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/handlers.go#L337)

**Rastreabilidade:** R6-EXE-02 ← R5-EXE-01; responsável Core e Integrações.

### F-R6-05 — Compensação progride após final público (P0)

**Natureza:** ANALISE_ESTATICA.

O DAG reverso e um prazo próprio foram adicionados. Os claims ainda exigem p.status NOT IN estados terminais mesmo quando continue_after_client_deadline=true. Portanto, a exceção de deadline não autoriza continuar após EXPIRED/CANCELLED. O TTL mínimo de compensação também é fixado em 30 segundos, sem política independente demonstrada.

**Motivo da correção:** O Hub SHALL separar estado de atendimento do cliente e estado de obrigações compensatórias. Encerrar protocolo não pode impedir compensação já devida; compensações SHALL respeitar causalidade reversa, política própria e reconciliação sem ressuscitar o final público.

**Proposta:** Criar elegibilidade de obrigação compensatória independente do terminal público e bloquear novas etapas produtivas após cancelamento. Configurar prazo/retry/UNKNOWN da compensação por versão. Cobrir ancestrais transitivos quando etapas intermediárias não tiverem compensação; falha de compensação requer disposição explícita e consulta administrativa.

**Fontes:** [hub/internal/orbita/intents.go:50](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L50), [hub/internal/orbita/product_store.go:264](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_store.go#L264), [hub/migrations/core/0048_compensation_independent_deadline.sql:4](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/migrations/core/0048_compensation_independent_deadline.sql#L4)

**Rastreabilidade:** R6-EXE-03 ← R5-EXE-04; responsável Core e Integrações.

### F-R6-06 — Liberação de slot e intent na mesma transação (P0)

**Natureza:** ANALISE_ESTATICA.

O claim ganhou lock do plano e running_count: não repetir o antigo achado de contagem desprotegida. CompleteIntent ainda atualiza command_intents em uma transação implícita e, depois, redefine a etapa e decrementa o plano em outro Exec. Uma interrupção entre essas escritas pode deixar estado e contador divergentes, sobretudo se o intent já ficou EXPIRED e não poderá ser reclamado.

**Motivo da correção:** O Hub SHALL atualizar intent, lease da etapa e contador do plano atomicamente também na conclusão, falha e expiração. Takeover e fatos concorrentes SHALL preservar max_parallel, idempotência e recuperabilidade sem zerar contadores artificialmente.

**Proposta:** Definir ordem única de locks entre claim, CompleteIntent e ApplyProductFact; incluir epoch em escritas e verificar RowsAffected. Reconciliador detecta drift por dados autoritativos, sem GREATEST mascarando perda de invariante. Testar falha no ponto entre estados com DB real e barreiras, não apenas go -race.

**Fontes:** [hub/internal/orbita/intents.go:295](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L295), [hub/internal/orbita/intents.go:336](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L336), [hub/internal/orbita/intents.go:180](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L180)

**Rastreabilidade:** R6-EXE-04 ← R5-EXE-02; responsável Core e Integrações.

### F-R6-07 — Replay financeiro distingue aplicação de nova quarentena (P0)

**Natureza:** ANALISE_ESTATICA.

Payload e endpoints de replay foram adicionados. ProcessEnvelope retorna nil também quando Quarantine grava com sucesso; o handler então marca REPLAYED. Se o payload ainda inválido recria a mesma identidade, ON CONFLICT DO NOTHING conserva a linha e MarkQuarantineReplayed a retira da lista de pendências sem aplicação. Fato tardio é armazenado como EconomicEvent, enquanto o endpoint espera queue.Envelope.

**Motivo da correção:** O Hub SHALL produzir disposição tipada para replay: aplicado, ainda em quarentena, conflito ou falha transitória. Uma obrigação não aplicada SHALL continuar visível e recuperável. Todo payload custodiado SHALL declarar formato/versão e ter vínculo auditável com bytes de origem, inclusive sem identidade de domínio válida.

**Proposta:** Persistir outcome e tentativa/ator atomicamente ou com claim cercado. Não marcar original encerrado sem recibo da obrigação sucessora ou aplicação. Tratar legado sem payload e EconomicEvent versus Envelope por formato explícito. Invalid JSON/event_id ausente exige identidade de transporte/hash sem inventar evento de negócio. Escopo global da lista requer autorização global real, não somente query tenant de fachada.

**Fontes:** [hub/internal/libra/consumers.go:33](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/consumers.go#L33), [hub/internal/libra/handlers.go:419](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/handlers.go#L419), [hub/internal/libra/store.go:389](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L389), [hub/internal/libra/store.go:420](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L420)

**Rastreabilidade:** R6-FIN-01 ← R5-DAD-04; responsável Financeiro e Backend.

### F-R6-08 — Completude e liquidação verificáveis no fluxo real (P0)

**Natureza:** ANALISE_ESTATICA.

A captura efetiva agora é chamada por ApplyEvent e registra diferença da reserva. SetWatermark permanece chamado apenas por teste; assim o mecanismo de fechamento carece de produtor durável de completude. Fatos tardios ainda entram em quarentena, não demonstram disputa/ajuste completo. A correção de captura deve ser preservada e testada com múltiplos fatos/ordem/escala decimal.

**Motivo da correção:** O Hub SHALL fechar período somente com prova de completude de todos os produtores e obrigações da coorte. Reserva, captura efetiva, liberação e ajustes SHALL conservar identidade e valores exatos sob reordenação, duplicação e resultados tardios.

**Proposta:** Definir marcador durável por produtor/partição/coorte com tratamento de lacunas e avanço monotônico, publicado por outbox no próprio fluxo. Tempo de parede não é prova de entrega. Conciliar receita do produto e custos de etapas; testar valor efetivo acima da reserva por política aprovada, sem fabricá-la. Duplicata semanticamente igual com escala decimal diferente não deve gerar conflito falso. Exportação fechada é imutável; ajustes posteriores são novos lançamentos.

**Fontes:** [hub/internal/libra/store.go:144](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L144), [hub/internal/libra/store.go:453](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L453), [hub/internal/libra/settlement.go:217](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/settlement.go#L217)

**Rastreabilidade:** R6-FIN-02 ← R5-DAD-03; responsável Financeiro e Backend.

### F-R6-09 — Gate de promoção vinculado ao artefato e à cobertura (P0)

**Natureza:** PROBE_REPRODUZIDO.

A ausência de manifesto agora bloqueia prd. Probe novo passou manifesto de um cenário, SHA de quarenta zeros, status PASS_WITHOUT_EXECUTION, command not executed e oráculos inventados iguais: o gate respondeu ALLOW. O validador exige forma e presença do ID no arquivo, não a identidade do candidato, integridade, cobertura obrigatória ou proveniência confiável.

**Motivo da correção:** O Hub SHALL bloquear promoção quando qualquer evidência obrigatória não corresponder ao artefato candidato, ao cenário vigente e à execução autorizada. Status SHALL usar enum estrito; cobertura parcial e declarações autoatribuídas não qualificam promoção.

**Proposta:** Pipeline gera manifesto com commit/diff, digest de imagem, spec/scenario/log, toolchain, execução e oráculos. Validador conhece inventário obrigatório e origem autorizada, verifica hash de bytes e assinatura/atestação confiável quando aplicável; trust root não vem do mesmo input não confiável. Testar também caminho ALLOW legítimo. Desenvolvimento local com fixtures não equivale a autorização de prd.

**Fontes:** [hub/deploy/r2/tests/promotion-gate.sh:22](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/promotion-gate.sh#L22), [hub/deploy/r2/tests/validate-qualification-evidence.py:71](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/validate-qualification-evidence.py#L71), [hub/deploy/r2/tests/validate-qualification-evidence.py:81](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/validate-qualification-evidence.py#L81)

**Rastreabilidade:** R6-OPE-01 ← R5-OPE-02; responsável Plataforma e SRE.

### F-R6-10 — Restore populado com versões e retomada reconciliada (P0)

**Natureza:** ANALISE_ESTATICA.

O script de restore não mudou nesta rodada: compara lista parcial de tabelas, faz s3 sync/list-objects-v2 e consulta o oráculo somente antes de declarar PASS sem replay. Planos, etapas, novos settlements e versões históricas de objetos não têm cobertura explícita. O relatório da implementação reconhece o item não iniciado.

**Motivo da correção:** O Hub SHALL demonstrar restore não vazio de todas as autoridades e obrigações, incluindo versões de objetos referenciadas, seguido de retomada cercada e reconciliação de efeitos e valores. A ausência de dados ou a mera igualdade de contagens SHALL impedir aprovação de recuperação integral.

**Proposta:** Inventário de schema e obrigações atualizado por migração. Fixture possui produtos em execução, UNKNOWN, callbacks, webhook pendente, saldo e versões/pins/tombstones. Restaurar em alvo isolado, mapear IDs de versão quando não preserváveis e verificar bytes/hash. Suspender novas admissões até reconciliação. Ensaiar RTO/RPO sob perfil acordado; restaurar sem privilégios requer reaplicar roles/policies e provar acesso correto.

**Fontes:** [hub/deploy/r2/tests/restore-reconciliation.sh:21](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L21), [hub/deploy/r2/tests/restore-reconciliation.sh:57](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L57), [hub/deploy/r2/tests/restore-reconciliation.sh:63](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L63)

**Rastreabilidade:** R6-OPE-02 ← R5-DAD-02; responsável Plataforma e SRE.

### F-R6-11 — Ambientes e elasticidade com orçamento de dependências (P1)

**Natureza:** ANALISE_ESTATICA.

Kind independente já existe e EnsureTopology foi ligado aos quatro processos antes dos consumidores/relays: preservar. Não houve evolução de deploy além do gate nesta rodada. O gerador de dependências kind declara volumes efêmeros; isso é adequado ao laboratório, mas não prova recuperação ou ambientes remotos duráveis. Bootstrap único também não qualifica remoção posterior de assinatura SNS.

**Motivo da correção:** O Hub SHALL fornecer perfis reproduzíveis local/dev/hom/ppd/prd com dependências e limites explícitos, escalabilidade automática dentro do envelope qualificado e continuidade mensurável. Readiness e topologia de obrigações SHALL refletir capacidade real de admitir/processar com custódia, sem confundir laboratório efêmero com HA produtiva.

**Proposta:** Docker Compose oficial único; kind serve qualificação, Kubernetes/EKS alvo remoto conforme decisão D-05. Gerar HPA/KEDA, recursos, PDB, startup/readiness/liveness, afinidade, identidades, redes, secrets e volumes/serviços duráveis por ambiente. Autoscaling de pods depende de nós, conexões, broker, storage e quotas cloud; onboarding automatizado valida capacidade antes de aceitar contrato. Topologia gerenciada por autoridade única ou reconciliação idempotente contínua com prova de falha de assinatura. Risco de perda regional exige política RPO/RTO aprovada.

**Fontes:** [hub/deploy/r2/kind/render-independent-dependencies.py:4](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/kind/render-independent-dependencies.py#L4), [hub/internal/queue/queue.go:228](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/queue/queue.go#L228), [hub/cmd/orbita/main.go:68](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/cmd/orbita/main.go#L68)

**Rastreabilidade:** R6-OPE-03 ← R5-OPE-03;R5-EXE-05; responsável Plataforma e SRE.

### F-R6-12 — Capacidade obrigatória e feedback recuperável (P1)

**Natureza:** ANALISE_ESTATICA.

O executor passou a recusar controller nil quando domínio existe. Domínio vazio ainda retorna capacidade desabilitada. Há resolução por evidência, porém settle com lease vencido pode falhar e falhas de resolução são logadas; contagens sob lock percorrem histórico do domínio. Não afirmar ausência do controlador adaptativo, que já está implementado.

**Motivo da correção:** O Hub SHALL exigir política efetiva de capacidade para rotas ativas e conservar obrigações de feedback/settlement até resolução. Controle adaptativo SHALL respeitar teto seguro contratado e compartilhar capacidade de forma justa entre tenants, com custo estável em função do estado ativo.

**Proposta:** Publicação recusa domínio ausente em rota que exige controle. Separar orçamento de transporte, SUBMIT, polling, pendências e reconciliação. Feedback de timeout/429/5xx reduz concorrência; janelas estáveis recuperam gradualmente até teto qualificado, nunca inferem capacidade infinita. Persistir settlement pendente e reconciliar por evidência, não por expiração de lease. Materializar contadores/índices do estado ativo e provar cardinalidade/custo.

**Fontes:** [hub/internal/cometa/executor.go:74](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L74), [hub/internal/cometa/executor.go:134](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L134), [hub/internal/cometa/capacity.go:202](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/capacity.go#L202)

**Rastreabilidade:** R6-OPE-04 ← R5-OPE-01; responsável Plataforma e SRE.

### F-R6-13 — Seletores completos para produtos e rotas (P1)

**Natureza:** ANALISE_ESTATICA.

Lookup agora busca no servidor com limite 50, resolvendo o antigo teto de 1000 para campos com pesquisa. StepsEditor e RoutesEditor recebem apenas essa primeira lista e não expõem pesquisa/cursor próprios; serviços, contas e bindings fora dos primeiros 50 ficam inalcançáveis nesses editores. Referência selecionada fora da página também não é carregada pontualmente.

**Motivo da correção:** O console SHALL permitir localizar e selecionar qualquer referência elegível por ID/versão, inclusive em etapas e rotas, sem carregar todo o catálogo nem ocultar a referência atualmente selecionada. Paginação e busca SHALL preservar tenant, filtros e estado de edição.

**Proposta:** Componente comum de seleção remota com debounce, cancelamento, cursor e resolução pontual do valor selecionado. Passar consulta por editor e não refazer todos os lookups por tecla. Diferenciar vazio de erro e impedir resposta antiga de outro tenant. Contratos de elegibilidade vêm da API, não de filtros locais parciais.

**Fontes:** [hub/admin-ui/src/pages/CatalogPage.tsx:25](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L25), [hub/admin-ui/src/pages/CatalogPage.tsx:59](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L59), [hub/admin-ui/src/pages/CatalogPage.tsx:76](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L76)

**Rastreabilidade:** R6-UX-01 ← R5-UX-01; responsável Frontend e Produto.

### F-R6-14 — Validação e intenção preservadas em toda mutação (P1)

**Natureza:** ANALISE_ESTATICA.

Guards runtime foram introduzidos, mas são opcionais e usados sobretudo nas listagens; detalhe, save e command ainda fazem cast sem guard. mappingErrors é local ao StepsEditor e não participa do bloqueio save do pai: JSON de mapping inválido pode deixar salvo o valor anterior. command mantém chave por duas tentativas, mas uma nova ação/reload cria outra identidade.

**Motivo da correção:** O console SHALL validar contratos de sucesso e erro por operação e bloquear mutações enquanto qualquer editor apresenta entrada inválida. Uma intenção de mutação com resultado desconhecido SHALL conservar identidade, payload e versão até reconciliação, inclusive após nova tentativa ou recarga.

**Proposta:** Propagar validade do editor ao formulário pai sem apagar texto inválido. Guards ou schemas gerados para DTOs de catálogo, finanças, protocolos e operações; erro de contrato não altera draft. Journal local da intenção sem tokens/segredos, escopado por usuário/tenant/ambiente, reconciliado com autoridade antes de gerar outra chave. Testes browser de timeout após commit, duplo clique, sessão expirada, ETag e permissões.

**Fontes:** [hub/admin-ui/src/api/admin.ts:18](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/api/admin.ts#L18), [hub/admin-ui/src/api/admin.ts:32](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/api/admin.ts#L32), [hub/admin-ui/src/pages/CatalogPage.tsx:62](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L62), [hub/admin-ui/src/pages/CatalogPage.tsx:41](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/admin-ui/src/pages/CatalogPage.tsx#L41)

**Rastreabilidade:** R6-UX-02 ← R5-UX-01; responsável Frontend e Produto.

### F-R6-15 — Evidência completa e vinculada ao código revisado (P1)

**Natureza:** ANALISE_ESTATICA.

O relatório R5 é explicitamente parcial. EVIDENCE_INDEX diz que os logs brutos ficaram no scroll, não foram persistidos. O commit menciona 246 testes, enquanto FINAL_REPORT/EVIDENCE_INDEX mencionam 245. SCENARIO_RESULTS contém apenas subconjunto e IDs de gate ad hoc. As matrizes originais da auditoria não foram regeneradas. Isto limita verificabilidade, não prova que testes históricos falharam.

**Motivo da correção:** A engenharia SHALL manter inventário integral extraído das specs e resultado individual com evidência por cenário. Contagens SHALL ser calculadas dos logs, com subtestes/pacotes diferenciados. Um resultado histórico sem vínculo verificável com artefato SHALL permanecer histórico e não virar PASS do candidato.

**Proposta:** Preservar v4/R2/R3/R4/R5 e incorporar deltas R6, com semântica distinta para achado, requisito, cenário e teste. Registrar SHA/diff, toolchain, imagem, fixture, comando, esperado/observado, logs saneados e digests. Revalidar só o que mudança material afeta, mas completar cenários obrigatórios sem skip. CLI OpenSpec fixada, sem @latest em evidência reproduzível. Decisões externas não aprovadas permanecem pendentes sem impedir implementação de fixture técnica.

**Fontes:** [docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md:29](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md#L29), [docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv:11](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv#L11), [docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md:48](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md#L48)

**Rastreabilidade:** R6-QUA-01 ← R5-QUA-01; responsável Engenharia e Qualidade.
