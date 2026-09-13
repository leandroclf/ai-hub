# Tasks: Isolamento e custódia autenticada

## Execução incremental
Responsável: Segurança e Core. Dependências entre changes e ambiente estão no plano da revisão.

## R6-SEG-01
- [ ] R6-SEG-01-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-SEG-01-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/migrations/core/0049_strict_runtime_rls.sql, hub/internal/platform/pg/pg.go, hub/deploy/r2/tests/rls-runtime-proof.sh, hub/deploy/r2/compose.yaml.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-SEG-01-B — Implementar a regra no caminho real
  - Objetivo: Mapear tabela→owner→role→policy→caminho. Separar migrador, requests e workers globais; estes precisam claims limitados e auditados, não tenant arbitrário fornecido pelo cliente. Adotar contexto LOCAL e provar limpeza em commit/rollback. A migração isolada não autoriza trocar o DSN e derrubar todos os workers. Qualificar migração com bases populadas e credenciais runtime efetivas.
  - Componentes: hub/migrations/core/0049_strict_runtime_rls.sql, hub/internal/platform/pg/pg.go, hub/deploy/r2/tests/rls-runtime-proof.sh, hub/deploy/r2/compose.yaml.
  - Depende de: R6-SEG-01-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-SEG-01-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL aplicar escopo autenticado por transação e por item de trabalho com roles runtime sem bypass; tabelas sem tenant direto devem ter autorização por relação ou autoridade operacional específica. Testes de isolamento SHALL afirmar existência e acesso próprio, invisibilidade alheia e recusa de escrita cruzada antes de remover fixtures.
  - Componentes: hub/migrations/core/0049_strict_runtime_rls.sql, hub/internal/platform/pg/pg.go, hub/deploy/r2/tests/rls-runtime-proof.sh, hub/deploy/r2/compose.yaml.
  - Depende de: R6-SEG-01-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-SEG-01-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/migrations/core/0049_strict_runtime_rls.sql, hub/internal/platform/pg/pg.go, hub/deploy/r2/tests/rls-runtime-proof.sh, hub/deploy/r2/compose.yaml.
  - Depende de: R6-SEG-01-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-SEG-02
- [ ] R6-SEG-02-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-SEG-02-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-SEG-02-B — Implementar a regra no caminho real
  - Objetivo: Persistir key-id/versão/conta/hash/instante de verificação e proteger sua integridade. Segredo em cofre, jamais no recibo. Reconciliar correlação sem exigir assinatura contra nova chave. Particionar locks/quotas por origem e célula; retenção de órfãos e quarentena com responsável, prazo, alertas e replay autorizado. Evitar descarte automático de dados aceitos.
  - Componentes: hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go.
  - Depende de: R6-SEG-02-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-SEG-02-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL preservar a atestação de autenticação obtida no ingresso e a disposição recuperável de cada callback aceito. Rotação e falha transitória não podem invalidar custódia legítima; esgotamento de retry deve manter obrigação em quarentena reprocessável, distinta de rejeição de contrato.
  - Componentes: hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go.
  - Depende de: R6-SEG-02-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-SEG-02-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go, hub/internal/cometa/custody.go.
  - Depende de: R6-SEG-02-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] SEG-G1 — Validar OpenSpec strict e compatibilidade com todos os requisitos anteriores vinculados.
- [ ] SEG-G2 — Conferir tarefas contra evidências e registrar riscos remanescentes sem autodeclarar aprovação produtiva.
