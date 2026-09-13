# Design: Isolamento e custódia autenticada

## Context
Brownfield a540b40007fe6b8ed523e17afe00e96ff8f8ad50; responsabilidade Segurança e Core. O delta adiciona critérios de fechamento, não substitui regras anteriores.

## R6-SEG-01 — decisão e justificativa
Mapear tabela→owner→role→policy→caminho. Separar migrador, requests e workers globais; estes precisam claims limitados e auditados, não tenant arbitrário fornecido pelo cliente. Adotar contexto LOCAL e provar limpeza em commit/rollback. A migração isolada não autoriza trocar o DSN e derrubar todos os workers. Qualificar migração com bases populadas e credenciais runtime efetivas.

### Contrato obrigatório
O Hub SHALL aplicar escopo autenticado por transação e por item de trabalho com roles runtime sem bypass; tabelas sem tenant direto devem ter autorização por relação ou autoridade operacional específica. Testes de isolamento SHALL afirmar existência e acesso próprio, invisibilidade alheia e recusa de escrita cruzada antes de remover fixtures.

### Superfícies afetadas
[hub/migrations/core/0049_strict_runtime_rls.sql:5](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/migrations/core/0049_strict_runtime_rls.sql#L5), [hub/internal/platform/pg/pg.go:39](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/platform/pg/pg.go#L39), [hub/deploy/r2/tests/rls-runtime-proof.sh:17](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/rls-runtime-proof.sh#L17), [hub/deploy/r2/compose.yaml:203](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/compose.yaml#L203)

### Provas de fechamento
- R6-SEG-01-S01: fixtures A e B confirmadas e role runtime → consultar pela API de A e SQL escopado → A existe e é visível, B existe e é invisível, escrita em B falha.
- R6-SEG-01-S02: pool de uma conexão e A/B alternados → executar sucesso, rollback e reconexão → não sobra identidade na conexão nem acesso owner.
- R6-SEG-01-S03: worker de múltiplos tenants e administrador nominal → processar claims e consultar escopo global → obrigações progridem com autorização específica e auditoria sem acesso global implícito.

## R6-SEG-02 — decisão e justificativa
Persistir key-id/versão/conta/hash/instante de verificação e proteger sua integridade. Segredo em cofre, jamais no recibo. Reconciliar correlação sem exigir assinatura contra nova chave. Particionar locks/quotas por origem e célula; retenção de órfãos e quarentena com responsável, prazo, alertas e replay autorizado. Evitar descarte automático de dados aceitos.

### Contrato obrigatório
O Hub SHALL preservar a atestação de autenticação obtida no ingresso e a disposição recuperável de cada callback aceito. Rotação e falha transitória não podem invalidar custódia legítima; esgotamento de retry deve manter obrigação em quarentena reprocessável, distinta de rejeição de contrato.

### Superfícies afetadas
[hub/internal/cometa/custody.go:228](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L228), [hub/internal/cometa/custody.go:195](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L195), [hub/internal/cometa/custody.go:385](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L385), [hub/internal/cometa/custody.go:411](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/custody.go#L411)

### Provas de fechamento
- R6-SEG-02-S01: callback legítimo aceito antes da correlação → rotacionar chave, reiniciar e criar operação → recibo original é aplicado uma vez com autenticação preservada.
- R6-SEG-02-S02: três falhas transitórias no apply de callback aceito → restaurar dependência e reprocessar → obrigação continua recuperável e resultado chega sem nova emissão do provedor.
- R6-SEG-02-S03: muitos órfãos de uma conta e conta saudável → executar ingestão e expirar janela de correlação → conta saudável progride e órfãos recebem disposição auditável sem lock global.

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
