# Tasks: Execução coerente de produtos e prazos

## Execução incremental
Responsável: Core e Integrações. Dependências entre changes e ambiente estão no plano da revisão.

## R6-EXE-01
- [ ] R6-EXE-01-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-EXE-01-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/orbita/product_plan.go, hub/internal/orbita/product_plan.go, hub/internal/cometa/executor.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-01-B — Implementar a regra no caminho real
  - Objetivo: Resolver grafo completo no Atlas durante publicação/admissão; não selecionar primeira rota preenchida ignorando elegibilidade. Incluir dependências versionadas no snapshot do produto; gerar comando a partir do filho completo, não de cópia parcial. Recalcular hash e verificar invariantes no executor. Venda do produto e custos das etapas usam chaves econômicas distintas. Preservar contrato final de consolidação.
  - Componentes: hub/internal/orbita/product_plan.go, hub/internal/orbita/product_plan.go, hub/internal/cometa/executor.go.
  - Depende de: R6-EXE-01-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-01-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL materializar cada etapa com rota, conta, binding, perfil, contrato de compra, prazo e incidência coerentes e congelados. Se a oferta não contém referências suficientes para resolver a etapa, a publicação ou admissão SHALL ser recusada antes de qualquer efeito.
  - Componentes: hub/internal/orbita/product_plan.go, hub/internal/orbita/product_plan.go, hub/internal/cometa/executor.go.
  - Depende de: R6-EXE-01-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-01-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/orbita/product_plan.go, hub/internal/orbita/product_plan.go, hub/internal/cometa/executor.go.
  - Depende de: R6-EXE-01-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-EXE-02
- [ ] R6-EXE-02-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-EXE-02-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/handlers.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-02-B — Implementar a regra no caminho real
  - Objetivo: Centralizar decisão persistida de elegibilidade imediatamente antes de I/O; distinguir SUBMIT de STATUS/consulta por chave. Consumir Expired em todo caminho e finalizar a intenção com disposição recuperável. Manter relógios cliente, provedor, tentativa e TTL desde primeira falha separados; não tratar polling saudável como indisponibilidade.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/handlers.go.
  - Depende de: R6-EXE-02-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-02-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL barrar nova submissão com possibilidade de efeito depois do prazo aplicável em DIRECT, QUEUED e recuperação. A reconciliação de efeito possivelmente já realizado SHALL continuar pela operação de consulta autorizada, sem se confundir com reenvio.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/handlers.go.
  - Depende de: R6-EXE-02-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-02-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/handlers.go.
  - Depende de: R6-EXE-02-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-EXE-03
- [ ] R6-EXE-03-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-EXE-03-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/product_store.go, hub/migrations/core/0048_compensation_independent_deadline.sql.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-03-B — Implementar a regra no caminho real
  - Objetivo: Criar elegibilidade de obrigação compensatória independente do terminal público e bloquear novas etapas produtivas após cancelamento. Configurar prazo/retry/UNKNOWN da compensação por versão. Cobrir ancestrais transitivos quando etapas intermediárias não tiverem compensação; falha de compensação requer disposição explícita e consulta administrativa.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/product_store.go, hub/migrations/core/0048_compensation_independent_deadline.sql.
  - Depende de: R6-EXE-03-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-03-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL separar estado de atendimento do cliente e estado de obrigações compensatórias. Encerrar protocolo não pode impedir compensação já devida; compensações SHALL respeitar causalidade reversa, política própria e reconciliação sem ressuscitar o final público.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/product_store.go, hub/migrations/core/0048_compensation_independent_deadline.sql.
  - Depende de: R6-EXE-03-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-03-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/product_store.go, hub/migrations/core/0048_compensation_independent_deadline.sql.
  - Depende de: R6-EXE-03-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-EXE-04
- [ ] R6-EXE-04-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-EXE-04-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/intents.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-04-B — Implementar a regra no caminho real
  - Objetivo: Definir ordem única de locks entre claim, CompleteIntent e ApplyProductFact; incluir epoch em escritas e verificar RowsAffected. Reconciliador detecta drift por dados autoritativos, sem GREATEST mascarando perda de invariante. Testar falha no ponto entre estados com DB real e barreiras, não apenas go -race.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/intents.go.
  - Depende de: R6-EXE-04-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-04-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL atualizar intent, lease da etapa e contador do plano atomicamente também na conclusão, falha e expiração. Takeover e fatos concorrentes SHALL preservar max_parallel, idempotência e recuperabilidade sem zerar contadores artificialmente.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/intents.go.
  - Depende de: R6-EXE-04-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-EXE-04-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/orbita/intents.go, hub/internal/orbita/intents.go, hub/internal/orbita/intents.go.
  - Depende de: R6-EXE-04-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] EXE-G1 — Validar OpenSpec strict e compatibilidade com todos os requisitos anteriores vinculados.
- [ ] EXE-G2 — Conferir tarefas contra evidências e registrar riscos remanescentes sem autodeclarar aprovação produtiva.
