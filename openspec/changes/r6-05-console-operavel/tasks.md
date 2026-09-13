# Tasks: Console administrativo operável

## Execução incremental
Responsável: Frontend e Produto. Dependências entre changes e ambiente estão no plano da revisão.

## R6-UX-01
- [ ] R6-UX-01-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-UX-01-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-UX-01-B — Implementar a regra no caminho real
  - Objetivo: Componente comum de seleção remota com debounce, cancelamento, cursor e resolução pontual do valor selecionado. Passar consulta por editor e não refazer todos os lookups por tecla. Diferenciar vazio de erro e impedir resposta antiga de outro tenant. Contratos de elegibilidade vêm da API, não de filtros locais parciais.
  - Componentes: hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: R6-UX-01-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-UX-01-C — Qualificar falhas e concorrência
  - Objetivo: O console SHALL permitir localizar e selecionar qualquer referência elegível por ID/versão, inclusive em etapas e rotas, sem carregar todo o catálogo nem ocultar a referência atualmente selecionada. Paginação e busca SHALL preservar tenant, filtros e estado de edição.
  - Componentes: hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: R6-UX-01-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-UX-01-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: R6-UX-01-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-UX-02
- [ ] R6-UX-02-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-UX-02-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-UX-02-B — Implementar a regra no caminho real
  - Objetivo: Propagar validade do editor ao formulário pai sem apagar texto inválido. Guards ou schemas gerados para DTOs de catálogo, finanças, protocolos e operações; erro de contrato não altera draft. Journal local da intenção sem tokens/segredos, escopado por usuário/tenant/ambiente, reconciliado com autoridade antes de gerar outra chave. Testes browser de timeout após commit, duplo clique, sessão expirada, ETag e permissões.
  - Componentes: hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: R6-UX-02-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-UX-02-C — Qualificar falhas e concorrência
  - Objetivo: O console SHALL validar contratos de sucesso e erro por operação e bloquear mutações enquanto qualquer editor apresenta entrada inválida. Uma intenção de mutação com resultado desconhecido SHALL conservar identidade, payload e versão até reconciliação, inclusive após nova tentativa ou recarga.
  - Componentes: hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: R6-UX-02-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-UX-02-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/api/admin.ts, hub/admin-ui/src/pages/CatalogPage.tsx, hub/admin-ui/src/pages/CatalogPage.tsx.
  - Depende de: R6-UX-02-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] UX-G1 — Validar OpenSpec strict e compatibilidade com todos os requisitos anteriores vinculados.
- [ ] UX-G2 — Conferir tarefas contra evidências e registrar riscos remanescentes sem autodeclarar aprovação produtiva.
