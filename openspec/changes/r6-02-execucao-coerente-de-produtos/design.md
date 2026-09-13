# Design: Execução coerente de produtos e prazos

## Context
Brownfield a540b40007fe6b8ed523e17afe00e96ff8f8ad50; responsabilidade Core e Integrações. O delta adiciona critérios de fechamento, não substitui regras anteriores.

## R6-EXE-01 — decisão e justificativa
Resolver grafo completo no Atlas durante publicação/admissão; não selecionar primeira rota preenchida ignorando elegibilidade. Incluir dependências versionadas no snapshot do produto; gerar comando a partir do filho completo, não de cópia parcial. Recalcular hash e verificar invariantes no executor. Venda do produto e custos das etapas usam chaves econômicas distintas. Preservar contrato final de consolidação.

### Contrato obrigatório
O Hub SHALL materializar cada etapa com rota, conta, binding, perfil, contrato de compra, prazo e incidência coerentes e congelados. Se a oferta não contém referências suficientes para resolver a etapa, a publicação ou admissão SHALL ser recusada antes de qualquer efeito.

### Superfícies afetadas
[hub/internal/orbita/product_plan.go:126](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_plan.go#L126), [hub/internal/orbita/product_plan.go:83](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_plan.go#L83), [hub/internal/cometa/executor.go:199](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L199)

### Provas de fechamento
- R6-EXE-01-S01: produto com serviços de A e B e credenciais distintas → admitir e executar duas etapas → cada endpoint observa sua credencial, conta, contrato e custo corretos.
- R6-EXE-01-S02: referência da conta B ausente ou inelegível → publicar ou admitir produto → falha antes de custódia executável e nenhum envio a A como fallback silencioso.
- R6-EXE-01-S03: produto aceito e catálogo alterado → reiniciar e continuar etapas e compensação → snapshots e contratos congelados permanecem coerentes e verificáveis.

## R6-EXE-02 — decisão e justificativa
Centralizar decisão persistida de elegibilidade imediatamente antes de I/O; distinguir SUBMIT de STATUS/consulta por chave. Consumir Expired em todo caminho e finalizar a intenção com disposição recuperável. Manter relógios cliente, provedor, tentativa e TTL desde primeira falha separados; não tratar polling saudável como indisponibilidade.

### Contrato obrigatório
O Hub SHALL barrar nova submissão com possibilidade de efeito depois do prazo aplicável em DIRECT, QUEUED e recuperação. A reconciliação de efeito possivelmente já realizado SHALL continuar pela operação de consulta autorizada, sem se confundir com reenvio.

### Superfícies afetadas
[hub/internal/orbita/intents.go:208](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L208), [hub/internal/orbita/intents.go:363](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L363), [hub/internal/orbita/handlers.go:337](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/handlers.go#L337)

### Provas de fechamento
- R6-EXE-02-S01: DIRECT READY com retry_until vencido e SLA cliente futuro → recuperador assume após crash → zero novos SUBMIT e intenção expirada auditada.
- R6-EXE-02-S02: DIRECT perdeu resposta depois de efeito externo → vencer retry e reconciliar por chave idempotente → efeito é observado sem duplicação e final público não é reaberto.
- R6-EXE-02-S03: QUEUED, DIRECT e polling saudável sob relógio controlado → atravessar fronteiras dos prazos → mesma regra pré-I/O e pendência saudável usa prazo do provedor.

## R6-EXE-03 — decisão e justificativa
Criar elegibilidade de obrigação compensatória independente do terminal público e bloquear novas etapas produtivas após cancelamento. Configurar prazo/retry/UNKNOWN da compensação por versão. Cobrir ancestrais transitivos quando etapas intermediárias não tiverem compensação; falha de compensação requer disposição explícita e consulta administrativa.

### Contrato obrigatório
O Hub SHALL separar estado de atendimento do cliente e estado de obrigações compensatórias. Encerrar protocolo não pode impedir compensação já devida; compensações SHALL respeitar causalidade reversa, política própria e reconciliação sem ressuscitar o final público.

### Superfícies afetadas
[hub/internal/orbita/intents.go:50](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L50), [hub/internal/orbita/product_store.go:264](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/product_store.go#L264), [hub/migrations/core/0048_compensation_independent_deadline.sql:4](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/migrations/core/0048_compensation_independent_deadline.sql#L4)

### Provas de fechamento
- R6-EXE-03-S01: A→B concluídos e C falha, cliente expira → executar recuperador após final EXPIRED → compensar B antes de A e conservar final EXPIRED.
- R6-EXE-03-S02: cancelamento após efeito conhecido → reiniciar workers → nenhuma etapa produtiva nova; compensação devida continua.
- R6-EXE-03-S03: compensação apresenta UNKNOWN ou falha definitiva → esgotar sua política e consultar operação → obrigação permanece reconciliável e falha não é promovida a sucesso.

## R6-EXE-04 — decisão e justificativa
Definir ordem única de locks entre claim, CompleteIntent e ApplyProductFact; incluir epoch em escritas e verificar RowsAffected. Reconciliador detecta drift por dados autoritativos, sem GREATEST mascarando perda de invariante. Testar falha no ponto entre estados com DB real e barreiras, não apenas go -race.

### Contrato obrigatório
O Hub SHALL atualizar intent, lease da etapa e contador do plano atomicamente também na conclusão, falha e expiração. Takeover e fatos concorrentes SHALL preservar max_parallel, idempotência e recuperabilidade sem zerar contadores artificialmente.

### Superfícies afetadas
[hub/internal/orbita/intents.go:295](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L295), [hub/internal/orbita/intents.go:336](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L336), [hub/internal/orbita/intents.go:180](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/orbita/intents.go#L180)

### Provas de fechamento
- R6-EXE-04-S01: intent de produto falha quando retry expira → interromper entre atualização do intent e liberação da etapa → após recuperação contador e etapas concordam e nenhuma vaga fica perdida.
- R6-EXE-04-S02: dois publishers e um fato final concorrentes → disputar última vaga e repetir conclusão → running_count nunca excede máximo nem sofre decremento duplo.
- R6-EXE-04-S03: lease RUNNING expira após crash pré-publicação → takeover com época nova e dono antigo retorna → mesma vaga é recuperada e dono antigo não altera estado.

## Persistência e atomicidade
A autoridade do domínio conserva transição, inbox/outbox e recibos em transação local. Cache é derivado; queda de Redis não altera a decisão durável. Para efeitos externos, lease sozinho não é prova de ausência de efeito: usar fencing reconhecido, idempotência do provedor ou reconciliação por chave. Estado UNKNOWN conserva obrigação.
## Comunicação e paralelismo
HTTP/JSON nos contratos existentes; SNS/SQS para fatos/obrigações; HTTP direto preservado para SYNC. Workers paralelos limitados por tenant/rota/produto, sem goroutine ilimitada. Timeout de transporte não encerra automaticamente o estado econômico.
## Segurança e privacidade
Identidade nominal/MFA para administração, autorização por recurso, logs sem payload sensível ou segredo. Testes com identidades A/B e controles positivos. Não registrar credenciais em evidências.
## Observabilidade
Métricas de aceites, finais, idade de obrigação, retries, violações de prazo, drift, quarentenas e ações administrativas. Labels limitadas por agregação; UUID/protocolo em logs/traces, não em séries ilimitadas. Cada erro de custódia tem alerta e disposição.
## Estratégia de migração e rollback
Aplicar migrações novas, manter checksums antigos e ensaiar upgrade populado. Rollback binário só se schema/dados forem compatíveis; caso contrário bloquear downgrade e aplicar forward fix. Não apagar recibos, históricos, saldos ou volumes para passar teste.
## Alternativas
Reescrita da stack acrescentaria risco sem resolver os invariantes. Mocks isolados são úteis para contrato puro, insuficientes para concorrência/DB/restore. Adoção de componente adicional requer ADR com benefício medido.
## Estratégia de validação
Reproduzir caso negativo antes da correção; confirmar chamada real, sucesso, falha, concorrência, crash e idempotência conforme cenário. DB/broker/objetos/OIDC reais em laboratório quando necessários. Browser para jornadas humanas. Medir carga/HA apenas com perfil de capacidade identificado.
