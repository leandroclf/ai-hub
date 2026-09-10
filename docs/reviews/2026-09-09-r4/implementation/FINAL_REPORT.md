# Relatório final da execução R4

## Atualização de qualificação — 10/09/2026

- OpenSpec estrito: 21/21 mudanças válidas.
- Kind: cluster recriado de forma controlada; três nós `Ready`, cinco
  deployments disponíveis, métricas/HPA funcionais e recuperação de pod
  validada.
- Portal: smoke Chromium autenticado passou com OIDC PKCE, senha+OTP,
  criação/leitura persistente, navegação, logout e viewport móvel sem overflow.
- Backend: `go test -race ./...` passou.
- Carga: autenticação passou, mas a execução funcional está bloqueada pela
  ausência de seed de catálogo publicado; os serviços responderam
  `offer_not_eligible` conforme o contrato.
- Browser Harness: integração opcional não executada porque o comando não está
  instalado.

## Resultado

Implementação parcial. Foram corrigidos contratos JSON estritos, preservação da migração histórica com reconciliação conhecida, caminho público de callback com autenticação de origem independente de JWT, cache L1/coordenação cancelável, recuperador autônomo da inbox, validação compartilhada de correlação/resultado e resolução indexada seletiva de ofertas. O portal recebeu proteção de sessão, preservação de rascunho e validação exploratória documentada. A suíte Go, o build do portal e o OpenSpec strict passam.

## Não concluído

Não há base para declarar conclusão integral: o Compose e restore foram executados, mas kind completo, browser autenticado atual, carga autorizada, HA, cenários externos de callback, revogação/crescimento de cache, projeção de ofertas e os cortes herdados continuam pendentes ou parciais. A auditoria detalhada está em `OPENSPEC-AUDIT-2026-09-10.md`.

O estado detalhado está em `REQUIREMENTS_STATUS.csv`, `SCENARIO_RESULTS.csv` e `CHECKPOINT.md`. Nenhum skip obrigatório foi convertido em aprovação.
