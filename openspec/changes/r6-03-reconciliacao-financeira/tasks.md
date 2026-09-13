# Tasks: Reconciliação financeira recuperável

## Execução incremental
Responsável: Financeiro e Backend. Dependências entre changes e ambiente estão no plano da revisão.

## R6-FIN-01
- [ ] R6-FIN-01-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-FIN-01-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/libra/consumers.go, hub/internal/libra/handlers.go, hub/internal/libra/store.go, hub/internal/libra/store.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-FIN-01-B — Implementar a regra no caminho real
  - Objetivo: Persistir outcome e tentativa/ator atomicamente ou com claim cercado. Não marcar original encerrado sem recibo da obrigação sucessora ou aplicação. Tratar legado sem payload e EconomicEvent versus Envelope por formato explícito. Invalid JSON/event_id ausente exige identidade de transporte/hash sem inventar evento de negócio. Escopo global da lista requer autorização global real, não somente query tenant de fachada.
  - Componentes: hub/internal/libra/consumers.go, hub/internal/libra/handlers.go, hub/internal/libra/store.go, hub/internal/libra/store.go.
  - Depende de: R6-FIN-01-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-FIN-01-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL produzir disposição tipada para replay: aplicado, ainda em quarentena, conflito ou falha transitória. Uma obrigação não aplicada SHALL continuar visível e recuperável. Todo payload custodiado SHALL declarar formato/versão e ter vínculo auditável com bytes de origem, inclusive sem identidade de domínio válida.
  - Componentes: hub/internal/libra/consumers.go, hub/internal/libra/handlers.go, hub/internal/libra/store.go, hub/internal/libra/store.go.
  - Depende de: R6-FIN-01-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-FIN-01-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/libra/consumers.go, hub/internal/libra/handlers.go, hub/internal/libra/store.go, hub/internal/libra/store.go.
  - Depende de: R6-FIN-01-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-FIN-02
- [ ] R6-FIN-02-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-FIN-02-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/libra/store.go, hub/internal/libra/store.go, hub/internal/libra/settlement.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-FIN-02-B — Implementar a regra no caminho real
  - Objetivo: Definir marcador durável por produtor/partição/coorte com tratamento de lacunas e avanço monotônico, publicado por outbox no próprio fluxo. Tempo de parede não é prova de entrega. Conciliar receita do produto e custos de etapas; testar valor efetivo acima da reserva por política aprovada, sem fabricá-la. Duplicata semanticamente igual com escala decimal diferente não deve gerar conflito falso. Exportação fechada é imutável; ajustes posteriores são novos lançamentos.
  - Componentes: hub/internal/libra/store.go, hub/internal/libra/store.go, hub/internal/libra/settlement.go.
  - Depende de: R6-FIN-02-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-FIN-02-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL fechar período somente com prova de completude de todos os produtores e obrigações da coorte. Reserva, captura efetiva, liberação e ajustes SHALL conservar identidade e valores exatos sob reordenação, duplicação e resultados tardios.
  - Componentes: hub/internal/libra/store.go, hub/internal/libra/store.go, hub/internal/libra/settlement.go.
  - Depende de: R6-FIN-02-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-FIN-02-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/libra/store.go, hub/internal/libra/store.go, hub/internal/libra/settlement.go.
  - Depende de: R6-FIN-02-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] FIN-G1 — Validar OpenSpec strict e compatibilidade com todos os requisitos anteriores vinculados.
- [ ] FIN-G2 — Conferir tarefas contra evidências e registrar riscos remanescentes sem autodeclarar aprovação produtiva.
