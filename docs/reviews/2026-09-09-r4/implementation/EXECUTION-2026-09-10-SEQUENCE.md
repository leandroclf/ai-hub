# Execução da sequência R4 — 2026-09-10

## Resultado desta iteração

Esta iteração fecha a correção de prazo do polling, conecta capacidade aos
transportes de execução e reconciliação, amplia o portal e requalifica os
gates locais disponíveis. O OpenSpec estrito permanece válido em 21/21
mudanças. As tarefas B-R4 continuam abertas quando o critério exige cobertura
integral dos cenários, integração ainda ausente ou validação de ambiente não
reprodutível.

## Evidências executadas

- Suíte Go completa com `-race`: PASS, sem pacotes ignorados, usando bancos
  isolados `*_r4test` e LocalStack local.
- `go vet ./...`: PASS.
- `hub/admin-ui`: `npm run build`: PASS.
- OpenSpec: `validate --all --strict --no-interactive --json`: PASS, 21/21.
- Seed do catálogo: PASS, recursos existentes e versionados.
- Playwright determinístico: PASS para OIDC Authorization Code + PKCE + OTP,
  persistência do admin, erro de JSON inválido sem falso sucesso, filtro local
  de referências, navegação autenticada, SLA bilateral, destinos versionados,
  ausência de tokens persistentes, viewport de 390px e logout.
- Editor de produto: PASS; o portal persiste e relê o mapeamento de entrada
  entre etapas dependentes.
- Produto/DAG via HTTP: PASS; duas etapas independentes geraram duas operações
  e dois efeitos no provider-sim, com consolidação `SUCCEEDED` e idempotência.
- Carga autorizada: PASS em SYNC sucesso/falha, ASYNC polling A/B, callback,
  AUTO, saldo estrito e credencial dedicada; idempotência observada quatro
  vezes para o mesmo protocolo.
- SLO de referência: PASS local com 30 admissões ASYNC e 30 GETs autenticados,
  concorrência 5 e payload de 59989 bytes; admissão p95=55,07 ms/p99=57,06 ms
  e GET p95=5,16 ms/p99=5,23 ms, dentro das metas propostas de R2-OPE-09-S01.
  A evidência não cobre ainda manutenção sob ruído nem converte a meta proposta
  em aprovação de produção.
- Broker fora: PASS local; `localstack` foi interrompido, Órbita e Cometa foram
  recriados sem dependências, permaneceram prontos e executaram SYNC/GET com
  `SUCCEEDED`. O broker foi restaurado ao final e a única fixture sintética
  sem efeito da primeira tentativa foi fechada pelo oráculo 404 local.
- Ambientes/promoção: PASS estrutural; os cinco overlays renderizaram namespaces
  distintos e referências de segredo sem valores materializados. O gate bloqueou
  `prd` sem perfil/aprovações P-01/P-08/P-10 e permitiu `dev`; o isolamento
  cross-environment com IdP e credenciais efetivas ainda é parcial.
- Dependência opcional: PASS local; Alloy ficou indisponível, Órbita e Cometa
  mantiveram readiness/liveness e os demais workloads não sofreram cascata de
  reinícios. O runner restaurou Alloy antes de terminar.
- Provedor fora: PASS local; o provider-sim foi interrompido, a tentativa SYNC
  retornou 504 sem anunciar sucesso, o GET preservou o protocolo e a concessão
  ficou fechada para transporte com pendência externa. Os demais containers não
  reiniciaram; o provider foi restaurado e a fixture reconciliada por 404.
- Quota/capacidade: PASS unitário PostgreSQL; 60 contendores em três identidades
  de réplica e duas células preservaram reservas A/B, cercaram owner stale e
  isolaram rate limit por tenant. Isso não qualifica autoscaling cloud ou
  provisionamento de novas células.
- Política de escala: PASS estrutural; KEDA declara sinal de backlog pendente,
  HPA/PDB/topology spread estão presentes e o overlay ppd fixa envelope 3→6.
  O ensaio de backlog real sob CPU baixa e o provisionamento cloud permanecem
  pendentes.
- Realocação de tenant: PASS integrado local; a consulta pública resolveu por
  tenant/UUID tanto o protocolo histórico da célula A quanto o novo da B após
  a mudança de placement, ocultou protocolo de outro tenant e manteve o fence
  da rota interna contra leitura pela célula errada.
- Fronteira de broker: PASS unitário/estrutural; o mesmo nome lógico de fila
  recebe prefixos distintos para ambiente/célula (`dev-cell-a`, `hom-cell-a` e
  `dev-cell-b`). O laboratório de célula única ainda não qualifica entrega
  runtime entre ambientes/células.
- RLS runtime: PASS nos domínios control/core/finance com prova negativa
  cross-tenant.
- Restore/reconciliation com sufixo `r4_sequence_20260910c`: PASS para bancos
  control/core/finance, objetos e oracle de efeitos externos sem replay. O
  script também recusa alvos de banco/bucket já existentes antes de iniciar.
- Kind existente `ai-hub-r2`: workloads atlas, orbita, cometa, pulsar e libra
  em `Running`; escala manual para duas réplicas por workload: PASS; PDBs,
  HPA/KEDA e placement em dois workers observados.
- Browser Harness: `PASS-EXPLORATORY` com `BU_CDP_URL` explícito; percorreu 16
  rotas em viewport 390×844. A descoberta automática do daemon headless ainda
  requer CDP explícito e não substitui o gate Playwright.
- Reconciliação da fixture de capacidade: `reconcile-local-pending.sh` fechou
  somente 12 concessões `r4-*` cujo oráculo sintético retornou `404`; nenhuma
  obrigação com efeito presente foi fechada. A rotina exige a confirmação
  `I_UNDERSTAND_LOCAL_FIXTURE` e é exclusiva do provider-sim.
- Carga autorizada pós-rebuild (`r4-authorized-1789087778741`): PASS após a
  conexão dos pools HTTP; os oito cenários terminaram conforme esperado e a
  mesma chave produziu quatro observações do mesmo protocolo.
- Reconciliação administrativa de protocolo: PASS no teste de integração com
  PostgreSQL; a solicitação exige MFA e `protocols:reconcile`, é idempotente
  enquanto aberta, fica auditada na mesma transação e retorna explicitamente
  `no_provider_replay`. O worker Cometa também foi exercitado com banco,
  catálogo e provedor HTTP reais de teste: uma consulta GET terminal produziu
  evidência, resolveu a solicitação e não gerou nenhum POST de replay.
- Identidade do laboratório: PASS na reconciliação idempotente do realm; o
  escopo `protocols:reconcile` está declarado nos clientes administrativos,
  mapeado somente para `hub_admin` e entregue aos usuários do console após
  MFA.
- Requalificação posterior: `go test ./... -count=1`, `go test -race ./...
  -count=1`, `npm run build`, Playwright, RLS runtime, carga autorizada e
  restore com sufixo `r4_20260911qual2` passaram novamente. A suíte de admissão
  usa uma célula sintética gerada por execução para não ser consumida pelo
  worker Orbita do laboratório.

## Correção implementada

`ScheduleAcceptedPollTx` passa a usar `StepDeadline` como prazo absoluto de
observação de um efeito externo já aceito, usando `RetryDeadline` somente como
fallback de compatibilidade. O teste PostgreSQL de custódia verifica o prazo
persistido. Isso evita que o TTL de recuperação de transporte expire um
polling saudável antes do SLA do passo.

Também foi criado um pool limitado de transports por origem em `egress`. Cada
consumidor obtém um cliente leve com timeout próprio, enquanto conexões ociosas
são reutilizadas pelo transport compartilhado; certificados mTLS continuam
clonando o transport antes da mutação. O teste de egress comprova a
reutilização por origem e a independência dos clientes.

A reconciliação administrativa agora tem custódia de lease/epoch em uma
migração aditiva, seleção por tenant/célula, consulta somente de status com o
snapshot de conta/vínculo da operação e resolução após a evidência terminal
ser gravada. O bootstrap do Cometa inicia esse worker junto do polling e da
inbox de callbacks; a cobertura PostgreSQL comprova takeover seguro no fluxo
normal e zero submissão durante a reconciliação.

A inbox de callback passou a validar payload terminal antes da custódia órfã,
separar duplicatas pela identidade da capability, impor limite de 512 KiB e
10.000 itens recebidos e remover processados de forma limitada e periódica.
O snapshot de destinos por aplicação/tenant é capturado na admissão e
transportado até o fato final, de onde o Pulsar entrega URL e política sem
consultar uma versão posterior.

## Pendências que impedem declaração integral

Ainda não há evidência suficiente para fechar a missão completa de 201
requisitos/732 cenários. Permanecem no OpenSpec os itens de provedor comercial,
carga e budgets completos em todo I/O, operações administrativas e financeiras
completas além das jornadas locais, FileRefs/resultados no adapter real,
fencing geral de efeito incerto, projeções escaláveis, telemetria bilateral e
matriz integral sem lacunas.

O resultado desta execução deve ser lido como `PASS` dos gates listados e
`OPEN` dos requisitos não demonstrados, nunca como promoção automática para
homologação ou produção.
