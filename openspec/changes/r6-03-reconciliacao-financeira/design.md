# Design: Reconciliação financeira recuperável

## Context
Brownfield a540b40007fe6b8ed523e17afe00e96ff8f8ad50; responsabilidade Financeiro e Backend. O delta adiciona critérios de fechamento, não substitui regras anteriores.

## R6-FIN-01 — decisão e justificativa
Persistir outcome e tentativa/ator atomicamente ou com claim cercado. Não marcar original encerrado sem recibo da obrigação sucessora ou aplicação. Tratar legado sem payload e EconomicEvent versus Envelope por formato explícito. Invalid JSON/event_id ausente exige identidade de transporte/hash sem inventar evento de negócio. Escopo global da lista requer autorização global real, não somente query tenant de fachada.

### Contrato obrigatório
O Hub SHALL produzir disposição tipada para replay: aplicado, ainda em quarentena, conflito ou falha transitória. Uma obrigação não aplicada SHALL continuar visível e recuperável. Todo payload custodiado SHALL declarar formato/versão e ter vínculo auditável com bytes de origem, inclusive sem identidade de domínio válida.

### Superfícies afetadas
[hub/internal/libra/consumers.go:33](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/consumers.go#L33), [hub/internal/libra/handlers.go:419](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/handlers.go#L419), [hub/internal/libra/store.go:389](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L389), [hub/internal/libra/store.go:420](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L420)

### Provas de fechamento
- R6-FIN-01-S01: quarentena cujo payload continua inválido → operador solicita replay → permanece pendente com tentativa registrada, nunca REPLAYED por simples nil.
- R6-FIN-01-S02: fato tardio após período fechado → consultar e reconciliar pelo endpoint → formato é interpretado e ajuste vinculado ao período sem reemitir SUBMIT.
- R6-FIN-01-S03: crash após aplicação e antes da confirmação do replay → repetir comando com mesma identidade → um efeito econômico, estado consistente e auditoria por ator.

## R6-FIN-02 — decisão e justificativa
Definir marcador durável por produtor/partição/coorte com tratamento de lacunas e avanço monotônico, publicado por outbox no próprio fluxo. Tempo de parede não é prova de entrega. Conciliar receita do produto e custos de etapas; testar valor efetivo acima da reserva por política aprovada, sem fabricá-la. Duplicata semanticamente igual com escala decimal diferente não deve gerar conflito falso. Exportação fechada é imutável; ajustes posteriores são novos lançamentos.

### Contrato obrigatório
O Hub SHALL fechar período somente com prova de completude de todos os produtores e obrigações da coorte. Reserva, captura efetiva, liberação e ajustes SHALL conservar identidade e valores exatos sob reordenação, duplicação e resultados tardios.

### Superfícies afetadas
[hub/internal/libra/store.go:144](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L144), [hub/internal/libra/store.go:453](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/store.go#L453), [hub/internal/libra/settlement.go:217](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/libra/settlement.go#L217)

### Provas de fechamento
- R6-FIN-02-S01: fatos fora de ordem e um produtor atrasado → solicitar fechamento → período bloqueia até cobertura durável de todos os fatos elegíveis.
- R6-FIN-02-S02: reserva 10 e consumo efetivo 7, com duplicata → processar no consumer e consultar saldo/journal → captura 7, liberação 3 e nenhum lançamento duplicado.
- R6-FIN-02-S03: fato econômico chega depois de exportação → reconciliar e aprovar ajuste → exportação original permanece íntegra e ajuste é consultável e rastreável.

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
