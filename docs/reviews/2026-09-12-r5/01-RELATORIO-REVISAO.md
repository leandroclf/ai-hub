# Revisão detalhada R5 — AI Hub
Snapshot [f87ce33034ae29c9431b1910dcc6a633b545e330](https://github.com/leandroclf/ai-hub/commit/f87ce33034ae29c9431b1910dcc6a633b545e330), commit de 12/09/2026 01:47:26 UTC.
Comparação com d93d86b: 404 arquivos alterados, 27.669 inserções e 1.116 remoções (inclui documentos e evidências, não só aplicação).

## Conclusão técnica
Há avanço funcional substancial. R4 original foi implementada em várias fatias; a edição regenerada de 10/09 não foi incorporada integralmente.
A próxima rodada fecha fronteiras de autorização, custódia, concorrência e operação, preservando correções comprovadas.
Não há base para declarar conformidade integral nem homologação produtiva. Isso não invalida as subprovas locais existentes.

## Avanços reconhecidos
- Probes anteriores JSON (null, tipo, EOF, minimum e precisão na transformação) passam.
- L1 válido antes do cofre e coordenação cancelável/limitada; probe adaptado confirma fallback de token.
- Migração 0002 restaurada e convergência restrita para hash conhecido.
- Callback público, HMAC por conta, inbox/worker, receipts e correlação têm implementação nova.
- Adapter rest-json-v1 e provas com endpoint HTTP distinto do provider-sim.
- Produtos com plano e etapas persistidos, duas operações HTTP e compensação registrada.
- Oferta seletiva, perfil efetivo, FileRefs, destinos e representação final avançaram.
- Console ganhou rotas, DTOs financeiros, SLA, destinos, multimodalidade, OIDC e smokes.
- Kind independente inclui dependências/gateway/UI; sinais e subprovas HA locais registrados.

## Método e limites
Leitura de instruções, delta, fluxos e evidências do snapshot. Fonte por arquivo/linha fixada abaixo.
Probes próprios separados da suíte do repositório. Achados estáticos exigem cenários de reprodução; não são incidentes de produção.
Go race/vet/build/OpenSpec executados; integrações PostgreSQL/Docker/kind/browser não reexecutadas por ausência desses runtimes aqui.
Go instalado nesta auditoria é 1.24.13, enquanto Dockerfile atualizado usa Go 1.26; o teste local não qualifica a imagem nova.
Evidências históricas são reconhecidas por escopo, sem importação automática para PASS deste SHA.

## Achados

### F-R5-01 — Autenticação antes de custódia de callback órfão (P0)

A rota pública e HMAC por conta foram implementados. Porém AuthenticateAccountCallback consulta operations antes de Verify. Para operação inexistente, o handler chama StoreOrphanAccountCallback, que só valida campos não vazios e grava via storeOrphan sem Verify. Assim, uma assinatura arbitrária com timestamp recente pode atingir a custódia órfã e 202, se o banco estiver disponível. A quota é global; variação da assinatura muda token_hash e identidade de dedupe. A reconciliação só seleciona órfãos com operação existente e prune não remove RECEIVED. Evidência estática de caminho; não foi executada exploração externa.

**Exigência:** O Hub SHALL autenticar a origem antes de confirmar ou reservar custódia órfã. Cada recibo deve possuir escopo de origem autenticada, identidade estável de evento e disposição recuperável. Tentativas não autenticadas não podem consumir quota durável de obrigações válidas. Quotas, retenção e claims devem isolar contas/células; eventos irrelacionáveis devem chegar a disposição auditável sem bloqueio global.

**Fontes:** [hub/internal/cometa/handlers.go:206](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/handlers.go#L206), [hub/internal/cometa/custody.go:228](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/custody.go#L228), [hub/internal/cometa/custody.go:240](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/custody.go#L240).

**Rastreio:** R5-SEG-01 → R4-CBK-01, R4-CBK-02, R4-CBK-03; dono: Segurança e Core.

### F-R5-02 — Fallback de oferta respeita negação e revogação (P0)

Offer agora tem cache de fallback, mas chama Atlas antes dele em toda requisição e usa cache após qualquer erro. Probe em httptest reproduziu retorno de oferta antiga após HTTP 403 offer_not_eligible. Suspensão/revogação explícita pode ser ocultada até vencer o cache; o caminho quente ainda paga a consulta remota.

**Exigência:** O Hub SHALL distinguir indisponibilidade transitória de negação autoritativa ao resolver ofertas. Negação, revogação, conflito ou resposta inválida não podem ser convertidos em autorização por cache. Projeções válidas devem atender o caminho quente dentro de janela de autorização explicitamente publicada, com invalidação e limites de armazenamento mensuráveis.

**Fontes:** [hub/internal/atlasclient/client.go:23](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlasclient/client.go#L23), [hub/internal/atlas/offers.go:202](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlas/offers.go#L202).

**Rastreio:** R5-SEG-02 → R4-OPE-03, R3-CAT-04, R2-SEG-01; dono: Segurança e Core.

### F-R5-03 — Adoção real de RLS e autorização de dados auxiliares (P0)

WithTenantTx permanece sem chamada nos stores; RuntimeDSN é opt-in de tenant fixo. Script RLS continua revertendo fixtures antes da assertion negativa e apenas imprime a contagem dentro da transação. Zero após rollback não prova isolamento. Migração de compatibilidade permite hub e grants abrangem tabelas auxiliares sem cobertura completa. Endpoint capacity-domains consulta todos os domínios para integrations:read sem escopo por recurso explícito.

**Exigência:** O Hub SHALL aplicar identidade autenticada por transação/work item com roles runtime sem propriedade nem bypass, incluindo tabelas auxiliares e caminhos administrativos. Escopo global exige autorização nominal específica e auditoria. Provas de isolamento devem demonstrar dados próprios existentes, dados alheios existentes e invisíveis, escrita cruzada recusada e ausência de vazamento ao reutilizar conexões.

**Fontes:** [hub/internal/platform/pg/pg.go:39](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/platform/pg/pg.go#L39), [hub/deploy/r2/tests/rls-runtime-proof.sh:17](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/rls-runtime-proof.sh#L17), [hub/migrations/core/0034_runtime_rls_migration_compat.sql:5](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/migrations/core/0034_runtime_rls_migration_compat.sql#L5), [hub/internal/cometa/handlers.go:65](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/handlers.go#L65).

**Rastreio:** R5-SEG-03 → R3-OPE-04, R3-ADM-01, R2-DAD-05; dono: Segurança e Core.

### F-R5-04 — Horizonte de retry e SLA verificados antes do despacho (P0)

A primeira falha/horizonte persistem, mas ClaimIntent e ClaimDirectIntent ainda não excluem retry_until vencido; CompleteIntent avalia depois de I/O e DELIVERED prevalece. O polling já foi corrigido para priorizar StepDeadline em vez do TTL inicial e essa correção deve ser preservada. UI ainda explica TTL desde aceite, divergindo da regra de primeira falha. Fencing de submissão foi implementado e deve ser preservado; ele não elimina a diferença entre esses relógios.

**Exigência:** O Hub SHALL impedir novos despachos com possibilidade de efeito após o horizonte aplicável e conservar separadamente prazo do cliente, provedor, tentativa e retry de indisponibilidade desde primeira falha. Pendência assíncrona legítima não inicia TTL de falha. Takeover não renova prazos; observação tardia conserva evidência sem alterar final fechado.

**Fontes:** [hub/internal/orbita/intents.go:135](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L135), [hub/internal/orbita/intents.go:255](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L255), [hub/internal/cometa/polling_custody.go:55](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/polling_custody.go#L55), [hub/admin-ui/src/pages/CatalogPage.tsx:6](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/CatalogPage.tsx#L6).

**Rastreio:** R5-EXE-01 → R3-EXE-04, R2-EXE-05, R2-EXE-06; dono: Core e Integrações.

### F-R5-05 — Limite de execução de produto atômico e retomável (P0)

Produtos agora têm plano/etapas e execução HTTP comprovada em laboratório. Porém o claim conta RUNNING/WAITING_PROVIDER sem cercar a linha do plano e atualiza etapa em outra transação: dois publishers podem observar vaga simultaneamente. Se crash ocorre após RUNNING antes de publicar, re-claim depende da mesma contagem e pode ficar bloqueado no próprio max_parallel. O helper antigo ExecuteDAG segue com falha de cancelamento, mas não é o runtime persistido; não confundir as implementações.

**Exigência:** O Hub SHALL reservar vaga de execução e posse da etapa atomicamente por produto, respeitando max_parallel entre réplicas. Uma etapa com posse expirada deve ser recuperável sem disputar uma segunda vaga nem repetir efeito confirmado. Cancelamento impede novo efeito e estado terminal não é ressuscitado.

**Fontes:** [hub/internal/orbita/intents.go:135](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L135), [hub/internal/orbita/product_store.go:55](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_store.go#L55), [hub/internal/atlas/executor.go:34](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlas/executor.go#L34).

**Rastreio:** R5-EXE-02 → R3-CAT-03, R2-EXE-01, R2-CAT-02; dono: Core e Integrações.

### F-R5-06 — Contrato, provedor e incidência próprios por etapa (P0)

BuildProductPlan copia o snapshot do produto, troca Target e preserva SelectedRoute/Account/Binding e EconomicSnapshot de base. Isso executa duas etapas no mesmo provedor sintético, mas não demonstra composição de serviços com provedores, credenciais e contratos de compra distintos. Hash do filho é esvaziado. O plano guarda consolidation, porém consolidação final usa formato fixo de steps.

**Exigência:** O Hub SHALL congelar por etapa serviço/versão, rota homologada, conta, binding, contrato de compra, perfil técnico, prazo e identidade econômica coerentes. A venda do produto e os custos das etapas devem manter escopos próprios sem duplicação. Consolidação deve seguir política publicada e preservar proveniência do plano.

**Fontes:** [hub/internal/orbita/product_plan.go:42](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_plan.go#L42), [hub/internal/orbita/product_plan.go:77](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_plan.go#L77), [hub/internal/atlas/offers.go:35](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/atlas/offers.go#L35).

**Rastreio:** R5-EXE-03 → R3-CAT-02, R3-CAT-03, R3-FIN-02; dono: Core e Integrações.

### F-R5-07 — Compensação tem ordem causal e prazo próprio (P0)

Compensações são persistidas, avanço sobre a R3. Porém são inseridas READY com depends_on vazio e herdam deadlines do comando original; o publisher exige protocolo não terminal e client_deadline futura. Assim, ordem reversa causal não está representada e compensação de obrigação externa pode deixar de executar depois de encerrar o atendimento.

**Exigência:** O Hub SHALL conservar e executar compensações em ordem causal reversa com prazo, retry e autorização próprios, independentemente do encerramento da resposta ao cliente. Falha ou incerteza de compensação deve permanecer reconciliável sem ser convertida em sucesso nem reabrir protocolo.

**Fontes:** [hub/internal/orbita/product_store.go:221](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_store.go#L221), [hub/internal/orbita/product_plan.go:116](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_plan.go#L116), [hub/internal/orbita/intents.go:48](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/intents.go#L48).

**Rastreio:** R5-EXE-04 → R3-CAT-03, R2-EXE-01; dono: Core e Integrações.

### F-R5-08 — Precisão do resultado preservada em todas as fronteiras (P0)

TransformJSON agora passa todos os probes anteriores. A perda reaparece depois: operationFact.ResponseBody é any e json.Unmarshal usa float64. Probe reproduziu 9007199254740993→9007199254740992 na desserialização/serialização desse fato. Consolidação de produto também usa any; frontend converte fact_id com Number. Correção do validador não protege essas fronteiras.

**Exigência:** O Hub SHALL preservar valor e tipo de números/identificadores em toda cadeia provedor→fato→estado→produto→representação→GET/webhook/console, sem coerção imprecisa. IDs inteiros fora da faixa segura do consumidor devem ter contrato textual explícito; resultados inválidos não são persistidos como sucesso.

**Fontes:** [hub/internal/orbita/factconsumer.go:25](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/factconsumer.go#L25), [hub/internal/orbita/factconsumer.go:35](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/factconsumer.go#L35), [hub/internal/orbita/product_store.go:289](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/orbita/product_store.go#L289), [hub/admin-ui/src/pages/FinancePage.tsx:58](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/FinancePage.tsx#L58).

**Rastreio:** R5-DAD-01 → R4-CTR-01, R4-CTR-02, R3-CAT-01; dono: Dados e Financeiro.

### F-R5-09 — Restore populado e reconciliação antes de retomada (P0)

Harness agora evita reutilizar alvo e compara digests SQL; permanece s3 sync/list-objects-v2 e uma única consulta ao oráculo antes de PASS. Não comprova versões históricas referenciadas, pins/tombstones, retomada nem reconciliação financeira/externa após replay. Lista de tabelas comparadas não inclui operation_plans/steps. Evidência histórica de cópia não encerra restore do produto atualizado.

**Exigência:** O Hub SHALL demonstrar restore não vazio e retomada cercada de todas as autoridades e obrigações, incluindo planos, etapas, compensações, recibos, versões de objetos e financeiro. Antes de liberar tráfego deve reconciliar identidades/bytes/efeitos/valores; ausência de fixture ou divergência impede aprovação.

**Fontes:** [hub/deploy/r2/tests/restore-reconciliation.sh:57](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/restore-reconciliation.sh#L57), [hub/deploy/r2/tests/restore-reconciliation.sh:63](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/restore-reconciliation.sh#L63), [hub/deploy/r2/tests/restore-reconciliation.sh:21](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/restore-reconciliation.sh#L21).

**Rastreio:** R5-DAD-02 → R3-OPE-04, R3-OPE-01, R2-OPE-08; dono: Dados e Financeiro.

### F-R5-10 — Liquidação por valor efetivo e completude demonstrável (P0)

Incidências SUBMITTED/STATUS, dedupe e ledger balanceado avançaram. Captura ainda muda estado da reserva sem apurar liberação da diferença para o valor real; SetWatermark aparece chamado em teste, sem produtor integrado de completude. Evento tardio insere finance_quarantine, apesar do comentário prometer disputa. São lacunas de fechamento operacional, não ausência de ledger.

**Exigência:** O Hub SHALL liquidar saldo pelo valor contratado efetivo, preservando reserva/hold de incerteza e liberando excedente de forma auditável. Fechamento exige evidência durável de completude dos produtores, sem watermarks fabricados. Fatos tardios devem gerar disposição financeira consultável e ajustável sem alterar exportação fechada.

**Fontes:** [hub/internal/libra/store.go:126](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L126), [hub/internal/libra/store.go:337](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L337), [hub/internal/libra/settlement.go:231](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/settlement.go#L231), [hub/internal/libra/store.go:300](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L300).

**Rastreio:** R5-DAD-03 → R3-FIN-01, R3-FIN-02, R2-FIN-03, R2-FIN-05; dono: Dados e Financeiro.

### F-R5-11 — Capacidade sem bypass e recuperação de permits (P1)

Controle adaptativo e pools agora estão ligados a SUBMIT/STATUS/reconciliação e webhook. Contudo domínio vazio/controller nil desabilita controle; permissões externas expiradas não são recicladas automaticamente e falha de settlement apenas gera log. Scripts históricos precisaram reconciliar permits por 404 do simulador. Contagens percorrem histórico de permits por domínio a cada Acquire sob lock global do domínio.

**Exigência:** O Hub SHALL exigir política de capacidade qualificada para cada rota ativa e recuperar concessões pendentes com evidência durável de transporte/efeito, sem reciclar apenas por timeout. Crescimento de histórico e número de tenants não deve violar orçamento de concessão publicado; isolamento e limite agregado devem valer entre réplicas.

**Fontes:** [hub/internal/cometa/executor.go:72](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/executor.go#L72), [hub/internal/cometa/capacity.go:250](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/capacity.go#L250), [hub/internal/cometa/capacity.go:346](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/capacity.go#L346), [hub/internal/cometa/executor.go:120](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/executor.go#L120).

**Rastreio:** R5-OPE-01 → R3-INT-01, R3-INT-03, R2-INT-05; dono: Plataforma e Operações.

### F-R5-12 — Promoção exige evidência vinculada ao artefato (P0)

Probe local do script devolveu ALLOW para prd com profile=unverified, três nomes de aprovação e isolation=PASS, sem manifesto de evidência. Script é gate, não deploy: nenhuma implantação foi feita. Validador exige formato de SHA, mas isso não vincula sozinho execução ao artefato promovido. HEAD altera bases Go/Nginx após evidências anteriores; Node build não está fixado por digest.

**Exigência:** O processo de promoção SHALL recusar artefato sem evidência íntegra e compatível com SHA/conteúdo/imagens efetivos e sem aprovações verificáveis do ambiente aplicável. Uma string PASS ou nome de aprovação não constitui prova. Mudança de toolchain/base exige requalificação material antes da promoção.

**Fontes:** [hub/deploy/r2/tests/promotion-gate.sh:22](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/promotion-gate.sh#L22), [hub/deploy/r2/tests/validate-qualification-evidence.py:53](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/validate-qualification-evidence.py#L53), [hub/deploy/Dockerfile:5](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/Dockerfile#L5), [hub/deploy/r2/Dockerfile.ui:1](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/Dockerfile.ui#L1).

**Rastreio:** R5-OPE-02 → R3-QUA-01, R2-QUA-04, R2-OPE-08; dono: Plataforma e Operações.

### F-R5-13 — Ambientes elásticos com dados duráveis e isolamento completo (P1)

Kind independente agora inclui dependências/UI/gateway: o achado antigo de ausência deve ser encerrado nesse escopo. Overlays remotos continuam centrados nas réplicas de cinco serviços; laboratório independente não constitui IaC regional nem durabilidade/escala de dados. Recuperar dois pods não mede continuidade de negócio sob perda de nó/zona e backlog financeiro.

**Exigência:** O Hub SHALL disponibilizar perfis local/dev/hom/ppd/prd reproduzíveis com identidade, dados, segredos, rede, observabilidade e capacidade isolados. Escala automática deve abranger pods/nós/placement e budgets das dependências dentro de quotas explícitas; continuidade é aferida por requisições/obrigações reconciliadas, não somente readiness.

**Fontes:** [hub/deploy/r2/kind/render-independent-dependencies.py:1](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/kind/render-independent-dependencies.py#L1), [hub/deploy/r2/kind/bootstrap-independent.sh:1](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/kind/bootstrap-independent.sh#L1), [hub/deploy/r2/k8s/overlays/prd/kustomization.yaml:6](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/k8s/overlays/prd/kustomization.yaml#L6).

**Rastreio:** R5-OPE-03 → R3-OPE-02, R3-OPE-03, R3-OPE-05, R2-OPE-01; dono: Plataforma e Operações.

### F-R5-14 — Console preserva intenção e escala com catálogo (P1)

Delivery ID, SLA/reconcile, DTO financeiro, destinos, OIDC e multimodalidade foram melhorados e têm smokes. Persistem lookup que carrega até 1000 e falha acima disso, validação runtime apenas isRecord e nova chave a cada nova chamada da ação após falha/reload. Capacidade é consulta somente leitura; onboarding de capacidade/qualificação segue seed direto no laboratório. UI TTL ainda descreve aceite (rastreado em R5-EXE-01).

**Exigência:** O console SHALL permitir selecionar referências por busca/paginação de servidor sem teto funcional de 1000, validar contratos de resposta por jornada e preservar identidade de intenção após resposta incerta. Configuração operacional suportada deve ter fluxo autenticado, versionado e auditável; o usuário deve distinguir pendente, confirmado e recusado.

**Fontes:** [hub/admin-ui/src/pages/CatalogPage.tsx:25](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/CatalogPage.tsx#L25), [hub/admin-ui/src/api/admin.ts:26](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/api/admin.ts#L26), [hub/admin-ui/src/api/admin.ts:29](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/api/admin.ts#L29), [hub/admin-ui/src/pages/CapacityDomainsPage.tsx:7](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/admin-ui/src/pages/CapacityDomainsPage.tsx#L7).

**Rastreio:** R5-UX-01 → R3-ADM-02, R3-ADM-03, R3-ADM-04, R4-QUA-02; dono: Frontend e Produto.

### F-R5-15 — Baseline reconciliada e resultados com proveniência (P1)

Repositório contém R4 original (12 requisitos/36 cenários), não o quinto change da edição regenerada (quatro requisitos adicionais). Baseline real é 201/732; matriz reconhece 231 linhas associadas e 501 sem qualificação. Gerador associa resultado por scenario_id, sem exigir digest do código/spec da execução na linha de origem; vínculo textual não prova atualidade nem PASS. R4 adicional precisa ponte explícita, sem copiar 744 como contagem remota.

**Exigência:** A engenharia SHALL inventariar specs reais e reconciliar revisões não incorporadas com rastreio explícito, sem perder requisito nem duplicar identidade. Resultado deve carregar proveniência de cenário/conteúdo/artefato e oráculo, distinguindo PASS, FAIL, NOT_RUN e bloqueio externo. Associação a arquivo não é conformidade; mudança material invalida a prova afetada.

**Fontes:** [docs/reviews/2026-09-09-r4/implementation/FINAL_REPORT.md:131](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/docs/reviews/2026-09-09-r4/implementation/FINAL_REPORT.md#L131), [hub/deploy/r2/tests/generate-openspec-results.py:103](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/generate-openspec-results.py#L103), [hub/deploy/r2/tests/generate-openspec-inventory.py:1](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/generate-openspec-inventory.py#L1).

**Rastreio:** R5-QUA-01 → R4-QUA-01, R4-QUA-03, R3-QUA-01; dono: Engenharia e Qualidade.

### F-R5-16 — Quarentena financeira conserva mensagem recuperável (P0)

ProcessEnvelope valida e chama Quarantine para envelope inválido; Quarantine persiste apenas payload_hash. O consumidor apaga a mensagem após retorno nil. Portanto o conteúdo inválido pode deixar de ser recuperável depois do ACK. O avanço do worker Cometa, que conserva bytes, não foi aplicado à autoridade financeira.

**Exigência:** O Hub SHALL conservar conteúdo recuperável e proveniência de mensagens financeiras não aplicáveis antes de confirmar sua remoção do transporte. Reprocessamento exige autorização, idempotência e trilha de disposição; hash isolado não constitui custódia do conteúdo.

**Fontes:** [hub/internal/libra/store.go:331](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/store.go#L331), [hub/internal/libra/consumers.go:58](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/libra/consumers.go#L58).

**Rastreio:** R5-DAD-04 → R3-EXE-05, R2-DAD-01, R2-FIN-01; dono: Dados e Financeiro.

### F-R5-17 — Topologia de mensagens pronta antes de publicar obrigações (P0)

Bootstrap assíncrono preserva disponibilidade HTTP durante falha de broker, mas cada processo confirma apenas sua parte. Órbita cria tópico de fatos finais e inicia relay sem comprovar assinaturas de Pulsar/Libra; Cometa publica fatos externos sem barreira comum de todas as assinaturas obrigatórias. Em ambiente limpo com startup fora de ordem, publicação SNS pode ser confirmada antes de assinaturas necessárias. Risco estático a ensaiar; tópicos já provisionados escondem essa janela.

**Exigência:** O Hub SHALL confirmar a topologia e políticas de todas as assinaturas obrigatórias antes de liberar publicação de fatos duráveis. Indisponibilidade da topologia deve manter obrigações no outbox sem bloquear rotas independentes. Alteração ou recriação de recurso exige revalidação antes de descartar custódia local.

**Fontes:** [hub/cmd/orbita/main.go:67](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/cmd/orbita/main.go#L67), [hub/cmd/cometa/main.go:72](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/cmd/cometa/main.go#L72), [hub/cmd/libra/main.go:58](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/cmd/libra/main.go#L58), [hub/internal/queue/bootstrap.go:12](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/queue/bootstrap.go#L12).

**Rastreio:** R5-EXE-05 → R3-EXE-05, R2-EXE-09; dono: Core e Integrações.
