# Tasks: Qualificação operacional e promoção

## Execução incremental
Responsável: Plataforma e SRE. Dependências entre changes e ambiente estão no plano da revisão.

## R6-OPE-01
- [ ] R6-OPE-01-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-OPE-01-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/deploy/r2/tests/promotion-gate.sh, hub/deploy/r2/tests/validate-qualification-evidence.py, hub/deploy/r2/tests/validate-qualification-evidence.py.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-01-B — Implementar a regra no caminho real
  - Objetivo: Pipeline gera manifesto com commit/diff, digest de imagem, spec/scenario/log, toolchain, execução e oráculos. Validador conhece inventário obrigatório e origem autorizada, verifica hash de bytes e assinatura/atestação confiável quando aplicável; trust root não vem do mesmo input não confiável. Testar também caminho ALLOW legítimo. Desenvolvimento local com fixtures não equivale a autorização de prd.
  - Componentes: hub/deploy/r2/tests/promotion-gate.sh, hub/deploy/r2/tests/validate-qualification-evidence.py, hub/deploy/r2/tests/validate-qualification-evidence.py.
  - Depende de: R6-OPE-01-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-01-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL bloquear promoção quando qualquer evidência obrigatória não corresponder ao artefato candidato, ao cenário vigente e à execução autorizada. Status SHALL usar enum estrito; cobertura parcial e declarações autoatribuídas não qualificam promoção.
  - Componentes: hub/deploy/r2/tests/promotion-gate.sh, hub/deploy/r2/tests/validate-qualification-evidence.py, hub/deploy/r2/tests/validate-qualification-evidence.py.
  - Depende de: R6-OPE-01-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-01-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/deploy/r2/tests/promotion-gate.sh, hub/deploy/r2/tests/validate-qualification-evidence.py, hub/deploy/r2/tests/validate-qualification-evidence.py.
  - Depende de: R6-OPE-01-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-OPE-02
- [ ] R6-OPE-02-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-OPE-02-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-02-B — Implementar a regra no caminho real
  - Objetivo: Inventário de schema e obrigações atualizado por migração. Fixture possui produtos em execução, UNKNOWN, callbacks, webhook pendente, saldo e versões/pins/tombstones. Restaurar em alvo isolado, mapear IDs de versão quando não preserváveis e verificar bytes/hash. Suspender novas admissões até reconciliação. Ensaiar RTO/RPO sob perfil acordado; restaurar sem privilégios requer reaplicar roles/policies e provar acesso correto.
  - Componentes: hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh.
  - Depende de: R6-OPE-02-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-02-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL demonstrar restore não vazio de todas as autoridades e obrigações, incluindo versões de objetos referenciadas, seguido de retomada cercada e reconciliação de efeitos e valores. A ausência de dados ou a mera igualdade de contagens SHALL impedir aprovação de recuperação integral.
  - Componentes: hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh.
  - Depende de: R6-OPE-02-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-02-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh, hub/deploy/r2/tests/restore-reconciliation.sh.
  - Depende de: R6-OPE-02-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-OPE-03
- [ ] R6-OPE-03-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-OPE-03-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/deploy/r2/kind/render-independent-dependencies.py, hub/internal/queue/queue.go, hub/cmd/orbita/main.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-03-B — Implementar a regra no caminho real
  - Objetivo: Docker Compose oficial único; kind serve qualificação, Kubernetes/EKS alvo remoto conforme decisão D-05. Gerar HPA/KEDA, recursos, PDB, startup/readiness/liveness, afinidade, identidades, redes, secrets e volumes/serviços duráveis por ambiente. Autoscaling de pods depende de nós, conexões, broker, storage e quotas cloud; onboarding automatizado valida capacidade antes de aceitar contrato. Topologia gerenciada por autoridade única ou reconciliação idempotente contínua com prova de falha de assinatura. Risco de perda regional exige política RPO/RTO aprovada.
  - Componentes: hub/deploy/r2/kind/render-independent-dependencies.py, hub/internal/queue/queue.go, hub/cmd/orbita/main.go.
  - Depende de: R6-OPE-03-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-03-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL fornecer perfis reproduzíveis local/dev/hom/ppd/prd com dependências e limites explícitos, escalabilidade automática dentro do envelope qualificado e continuidade mensurável. Readiness e topologia de obrigações SHALL refletir capacidade real de admitir/processar com custódia, sem confundir laboratório efêmero com HA produtiva.
  - Componentes: hub/deploy/r2/kind/render-independent-dependencies.py, hub/internal/queue/queue.go, hub/cmd/orbita/main.go.
  - Depende de: R6-OPE-03-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-03-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/deploy/r2/kind/render-independent-dependencies.py, hub/internal/queue/queue.go, hub/cmd/orbita/main.go.
  - Depende de: R6-OPE-03-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

## R6-OPE-04
- [ ] R6-OPE-04-A — Reproduzir a fronteira observada
  - Objetivo: Fixar fixture e oráculo dos cenários R6-OPE-04-S01…S03; preservar a melhoria R5 descrita em explore.md.
  - Componentes: hub/internal/cometa/executor.go, hub/internal/cometa/executor.go, hub/internal/cometa/capacity.go.
  - Depende de: Nenhuma; baseline fixada.
  - Validação: Leitura e teste negativo.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-04-B — Implementar a regra no caminho real
  - Objetivo: Publicação recusa domínio ausente em rota que exige controle. Separar orçamento de transporte, SUBMIT, polling, pendências e reconciliação. Feedback de timeout/429/5xx reduz concorrência; janelas estáveis recuperam gradualmente até teto qualificado, nunca inferem capacidade infinita. Persistir settlement pendente e reconciliar por evidência, não por expiração de lease. Materializar contadores/índices do estado ativo e provar cardinalidade/custo.
  - Componentes: hub/internal/cometa/executor.go, hub/internal/cometa/executor.go, hub/internal/cometa/capacity.go.
  - Depende de: R6-OPE-04-A.
  - Validação: Unitário e integração pela API/worker.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-04-C — Qualificar falhas e concorrência
  - Objetivo: O Hub SHALL exigir política efetiva de capacidade para rotas ativas e conservar obrigações de feedback/settlement até resolução. Controle adaptativo SHALL respeitar teto seguro contratado e compartilhar capacidade de forma justa entre tenants, com custo estável em função do estado ativo.
  - Componentes: hub/internal/cometa/executor.go, hub/internal/cometa/executor.go, hub/internal/cometa/capacity.go.
  - Depende de: R6-OPE-04-B.
  - Validação: Executar todos os cenários e limites do domínio.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] R6-OPE-04-D — Integrar operação e evidência
  - Objetivo: Adicionar métricas/runbook, ensaiar migração/rollback quando aplicável e atualizar resultado individual sem promoção de evidência antiga.
  - Componentes: hub/internal/cometa/executor.go, hub/internal/cometa/executor.go, hub/internal/cometa/capacity.go.
  - Depende de: R6-OPE-04-C.
  - Validação: Artefato+spec+log com digests; revisão do resultado.
  - Conclusão: evidência verificável do critério, com esperado/observado e resultado no candidato.

- [ ] OPE-G1 — Validar OpenSpec strict e compatibilidade com todos os requisitos anteriores vinculados.
- [ ] OPE-G2 — Conferir tarefas contra evidências e registrar riscos remanescentes sem autodeclarar aprovação produtiva.
