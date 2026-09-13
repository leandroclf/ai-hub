# Riscos

| Risco observado | Prioridade | Mitigação | Dono |
|---|---|---|---|
| Delivery ID, SLA/reconcile, DTO financeiro, destinos, OIDC e multimodalidade foram melhorados e têm smokes. Persistem lookup que carrega até 1000 e falha acima disso, validação runtime apenas isRecord e nova chave a cada nova chamada da ação após falha/reload. Capacidade é consulta somente leitura; onboarding de capacidade/qualificação segue seed direto no laboratório. UI TTL ainda descreve aceite (rastreado em R5-EXE-01). | P1 | R5-UX-01 | Frontend e Produto |
