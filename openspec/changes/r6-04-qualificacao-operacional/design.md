# Design: Qualificação operacional e promoção

## Context
Brownfield a540b40007fe6b8ed523e17afe00e96ff8f8ad50; responsabilidade Plataforma e SRE. O delta adiciona critérios de fechamento, não substitui regras anteriores.

## R6-OPE-01 — decisão e justificativa
Pipeline gera manifesto com commit/diff, digest de imagem, spec/scenario/log, toolchain, execução e oráculos. Validador conhece inventário obrigatório e origem autorizada, verifica hash de bytes e assinatura/atestação confiável quando aplicável; trust root não vem do mesmo input não confiável. Testar também caminho ALLOW legítimo. Desenvolvimento local com fixtures não equivale a autorização de prd.

### Contrato obrigatório
O Hub SHALL bloquear promoção quando qualquer evidência obrigatória não corresponder ao artefato candidato, ao cenário vigente e à execução autorizada. Status SHALL usar enum estrito; cobertura parcial e declarações autoatribuídas não qualificam promoção.

### Superfícies afetadas
[hub/deploy/r2/tests/promotion-gate.sh:22](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/promotion-gate.sh#L22), [hub/deploy/r2/tests/validate-qualification-evidence.py:71](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/validate-qualification-evidence.py#L71), [hub/deploy/r2/tests/validate-qualification-evidence.py:81](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/validate-qualification-evidence.py#L81)

### Provas de fechamento
- R6-OPE-01-S01: manifesto de um cenário, SHA falso e status prefixado PASS → avaliar promoção prd → BLOCK por identidade, enum e cobertura insuficientes.
- R6-OPE-01-S02: manifesto completo de execução autorizada e artefato idêntico → avaliar gate e depois alterar log/imagem/spec → ALLOW original e BLOCK de cada adulteração.
- R6-OPE-01-S03: manifesto contém skip obrigatório ou aprovação não autenticada → avaliar promoção → BLOCK com motivo específico sem executar deploy.

## R6-OPE-02 — decisão e justificativa
Inventário de schema e obrigações atualizado por migração. Fixture possui produtos em execução, UNKNOWN, callbacks, webhook pendente, saldo e versões/pins/tombstones. Restaurar em alvo isolado, mapear IDs de versão quando não preserváveis e verificar bytes/hash. Suspender novas admissões até reconciliação. Ensaiar RTO/RPO sob perfil acordado; restaurar sem privilégios requer reaplicar roles/policies e provar acesso correto.

### Contrato obrigatório
O Hub SHALL demonstrar restore não vazio de todas as autoridades e obrigações, incluindo versões de objetos referenciadas, seguido de retomada cercada e reconciliação de efeitos e valores. A ausência de dados ou a mera igualdade de contagens SHALL impedir aprovação de recuperação integral.

### Superfícies afetadas
[hub/deploy/r2/tests/restore-reconciliation.sh:21](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L21), [hub/deploy/r2/tests/restore-reconciliation.sh:57](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L57), [hub/deploy/r2/tests/restore-reconciliation.sh:63](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L63)

### Provas de fechamento
- R6-OPE-02-S01: backup com etapas pendentes, UNKNOWN e duas versões referenciadas → restaurar em alvo isolado → todas as referências resolvem para bytes corretos e obrigações permanecem.
- R6-OPE-02-S02: restore completo e fronteira de I/O cercada → retomar workers e comparar oráculo antes/depois → efeitos e saldos exatos sem duplicação e protocolos recuperados.
- R6-OPE-02-S03: backup vazio ou versão de objeto ausente → executar qualificação → BLOCK, sem anunciar restore reconciliado.

## R6-OPE-03 — decisão e justificativa
Docker Compose oficial único; kind serve qualificação, Kubernetes/EKS alvo remoto conforme decisão D-05. Gerar HPA/KEDA, recursos, PDB, startup/readiness/liveness, afinidade, identidades, redes, secrets e volumes/serviços duráveis por ambiente. Autoscaling de pods depende de nós, conexões, broker, storage e quotas cloud; onboarding automatizado valida capacidade antes de aceitar contrato. Topologia gerenciada por autoridade única ou reconciliação idempotente contínua com prova de falha de assinatura. Risco de perda regional exige política RPO/RTO aprovada.

### Contrato obrigatório
O Hub SHALL fornecer perfis reproduzíveis local/dev/hom/ppd/prd com dependências e limites explícitos, escalabilidade automática dentro do envelope qualificado e continuidade mensurável. Readiness e topologia de obrigações SHALL refletir capacidade real de admitir/processar com custódia, sem confundir laboratório efêmero com HA produtiva.

### Superfícies afetadas
[hub/deploy/r2/kind/render-independent-dependencies.py:4](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/kind/render-independent-dependencies.py#L4), [hub/internal/queue/queue.go:228](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/queue/queue.go#L228), [hub/cmd/orbita/main.go:68](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/cmd/orbita/main.go#L68)

### Provas de fechamento
- R6-OPE-03-S01: aumento de tenants dentro do envelope de capacidade → executar onboarding e carga → escala automática sem ticket por cliente, com budgets e isolamento mensurados.
- R6-OPE-03-S02: rolling update e falha de pod/nó com obrigações aceitas → substituir componentes → custódia preservada, readiness correta e recuperação dentro do SLO do perfil.
- R6-OPE-03-S03: assinatura obrigatória removida após bootstrap → publicar e observar reconciliação → obrigação não é silenciosamente perdida; alerta, contenção e recuperação verificáveis.

## R6-OPE-04 — decisão e justificativa
Publicação recusa domínio ausente em rota que exige controle. Separar orçamento de transporte, SUBMIT, polling, pendências e reconciliação. Feedback de timeout/429/5xx reduz concorrência; janelas estáveis recuperam gradualmente até teto qualificado, nunca inferem capacidade infinita. Persistir settlement pendente e reconciliar por evidência, não por expiração de lease. Materializar contadores/índices do estado ativo e provar cardinalidade/custo.

### Contrato obrigatório
O Hub SHALL exigir política efetiva de capacidade para rotas ativas e conservar obrigações de feedback/settlement até resolução. Controle adaptativo SHALL respeitar teto seguro contratado e compartilhar capacidade de forma justa entre tenants, com custo estável em função do estado ativo.

### Superfícies afetadas
[hub/internal/cometa/executor.go:74](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L74), [hub/internal/cometa/executor.go:134](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L134), [hub/internal/cometa/capacity.go:202](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/capacity.go#L202)

### Provas de fechamento
- R6-OPE-04-S01: rota sem domínio de capacidade ou política vencida → publicar e executar → negação seletiva antes de I/O sem bypass.
- R6-OPE-04-S02: provedor reduz capacidade e depois escala → medir feedback com dois tenants → controlador reduz pressão, recupera de forma estável e respeita teto e justiça.
- R6-OPE-04-S03: settlement falha e lease vence → reiniciar e fornecer evidência terminal → permit é resolvido sem perda de obrigação nem reciclagem cega por TTL.

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
