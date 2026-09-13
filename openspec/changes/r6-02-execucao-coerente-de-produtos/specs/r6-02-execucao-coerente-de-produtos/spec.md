# Delta for r6-02-execucao-coerente-de-produtos

## ADDED Requirements

### Requirement: R6-EXE-01 — Snapshot de etapa coerente em todas as identidades
O Hub SHALL materializar cada etapa com rota, conta, binding, perfil, contrato de compra, prazo e incidência coerentes e congelados. Se a oferta não contém referências suficientes para resolver a etapa, a publicação ou admissão SHALL ser recusada antes de qualquer efeito.

#### Scenario: R6-EXE-01-S01 — produto com serviços de A e B e credenciais distintas
- GIVEN produto com serviços de A e B e credenciais distintas
- WHEN admitir e executar duas etapas
- THEN cada endpoint observa sua credencial, conta, contrato e custo corretos

#### Scenario: R6-EXE-01-S02 — referência da conta B ausente ou inelegível
- GIVEN referência da conta B ausente ou inelegível
- WHEN publicar ou admitir produto
- THEN falha antes de custódia executável e nenhum envio a A como fallback silencioso

#### Scenario: R6-EXE-01-S03 — produto aceito e catálogo alterado
- GIVEN produto aceito e catálogo alterado
- WHEN reiniciar e continuar etapas e compensação
- THEN snapshots e contratos congelados permanecem coerentes e verificáveis

### Requirement: R6-EXE-02 — Expiração bloqueia também recuperação DIRECT
O Hub SHALL barrar nova submissão com possibilidade de efeito depois do prazo aplicável em DIRECT, QUEUED e recuperação. A reconciliação de efeito possivelmente já realizado SHALL continuar pela operação de consulta autorizada, sem se confundir com reenvio.

#### Scenario: R6-EXE-02-S01 — DIRECT READY com retry_until vencido e SLA cliente futuro
- GIVEN DIRECT READY com retry_until vencido e SLA cliente futuro
- WHEN recuperador assume após crash
- THEN zero novos SUBMIT e intenção expirada auditada

#### Scenario: R6-EXE-02-S02 — DIRECT perdeu resposta depois de efeito externo
- GIVEN DIRECT perdeu resposta depois de efeito externo
- WHEN vencer retry e reconciliar por chave idempotente
- THEN efeito é observado sem duplicação e final público não é reaberto

#### Scenario: R6-EXE-02-S03 — QUEUED, DIRECT e polling saudável sob relógio controlado
- GIVEN QUEUED, DIRECT e polling saudável sob relógio controlado
- WHEN atravessar fronteiras dos prazos
- THEN mesma regra pré-I/O e pendência saudável usa prazo do provedor

### Requirement: R6-EXE-03 — Compensação progride após final público
O Hub SHALL separar estado de atendimento do cliente e estado de obrigações compensatórias. Encerrar protocolo não pode impedir compensação já devida; compensações SHALL respeitar causalidade reversa, política própria e reconciliação sem ressuscitar o final público.

#### Scenario: R6-EXE-03-S01 — A→B concluídos e C falha, cliente expira
- GIVEN A→B concluídos e C falha, cliente expira
- WHEN executar recuperador após final EXPIRED
- THEN compensar B antes de A e conservar final EXPIRED

#### Scenario: R6-EXE-03-S02 — cancelamento após efeito conhecido
- GIVEN cancelamento após efeito conhecido
- WHEN reiniciar workers
- THEN nenhuma etapa produtiva nova; compensação devida continua

#### Scenario: R6-EXE-03-S03 — compensação apresenta UNKNOWN ou falha definitiva
- GIVEN compensação apresenta UNKNOWN ou falha definitiva
- WHEN esgotar sua política e consultar operação
- THEN obrigação permanece reconciliável e falha não é promovida a sucesso

### Requirement: R6-EXE-04 — Liberação de slot e intent na mesma transação
O Hub SHALL atualizar intent, lease da etapa e contador do plano atomicamente também na conclusão, falha e expiração. Takeover e fatos concorrentes SHALL preservar max_parallel, idempotência e recuperabilidade sem zerar contadores artificialmente.

#### Scenario: R6-EXE-04-S01 — intent de produto falha quando retry expira
- GIVEN intent de produto falha quando retry expira
- WHEN interromper entre atualização do intent e liberação da etapa
- THEN após recuperação contador e etapas concordam e nenhuma vaga fica perdida

#### Scenario: R6-EXE-04-S02 — dois publishers e um fato final concorrentes
- GIVEN dois publishers e um fato final concorrentes
- WHEN disputar última vaga e repetir conclusão
- THEN running_count nunca excede máximo nem sofre decremento duplo

#### Scenario: R6-EXE-04-S03 — lease RUNNING expira após crash pré-publicação
- GIVEN lease RUNNING expira após crash pré-publicação
- WHEN takeover com época nova e dono antigo retorna
- THEN mesma vaga é recuperada e dono antigo não altera estado
