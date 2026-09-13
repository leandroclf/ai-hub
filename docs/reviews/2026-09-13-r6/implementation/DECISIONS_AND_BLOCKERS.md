# Decisões e impedimentos — R6

## D-LOCAL-01 — Não cortar a credencial de runtime antes de converter os call sites
**Decisão (técnica, local, não externa):** implementar e provar o mecanismo de escopo por transação (`WithTenantTx`/`WithAuditedScopeTx` + policies `audited_scope`) sem ainda trocar `CORE_DSN`/`CONTROL_DSN`/`FINANCE_DSN` de `hub` (owner, superuser) para `hub_runtime`.
**Motivo:** o role `hub` usado hoje por todos os 5 serviços é `SUPERUSER` (confirmado via `pg_roles` no Postgres real do `hub-local`) — RLS é ignorado incondicionalmente por superusers, então nenhuma policy jamais foi de fato exercida em produção, apesar de `FORCE ROW LEVEL SECURITY` estar habilitado desde R5. Trocar a credencial sem antes converter todo call site que usa `s.db.Query/Exec/Begin` diretamente faria toda query não convertida devolver 0 linhas (SELECT/UPDATE) silenciosamente ou falhar (INSERT) — uma regressão em produção, não uma correção.
**Impacto:** R6-SEG-01 permanece **parcialmente implementado**: o mecanismo existe, está provado com Postgres real (não simulado, não com ROLLBACK mascarando o oráculo), mas não está no caminho real do runtime ainda. F-R6-01 não pode ser declarado fechado.
**Alternativa técnica local:** converter serviço por serviço, começando por `orbita` (não compartilha `hub_core` com segurança até `cometa`/`pulsar` também migrarem), com suíte de testes real (Postgres) verde antes de cada corte de credencial. Call sites diretos hoje: orbita 25, cometa 45, libra 23, atlas 23, pulsar 16 (total 132).
**Não é aprovação externa pendente** — é trabalho de engenharia sequenciado; nenhuma decisão de negócio bloqueia o avanço.

## Achado adicional não catalogado em F-R6-01 original
`hub_control` (`migrations/control/0034_runtime_rls_migration_compat.sql`) e `hub_finance` (`migrations/finance/0034_runtime_rls_migration_compat.sql`) ainda tinham a cláusula `current_user = 'hub' OR ...` nas suas policies — o mesmo bypass nominal que a 0049 já havia removido de `hub_core` na R5. Corrigido nesta sessão em `migrations/control/0038_strict_runtime_rls.sql` e `migrations/finance/0035_strict_runtime_rls.sql`. Mesma causa raiz de F-R6-01, superfície mais ampla do que o relatório original mapeou.

## Impedimentos externos irredutíveis (nenhum nesta sessão, até agora)
- Docker: o socket do Docker Desktop (`~/.docker/desktop/docker.sock`) não respondia; contornado usando o daemon do sistema via `DOCKER_HOST=unix:///var/run/docker.sock`, que já mantinha o ecossistema `hub-local` ativo e saudável (Postgres real usado para toda a prova desta sessão). Não é um impedimento — apenas uma variável de ambiente ausente por padrão neste shell.
- Nenhum outro impedimento externo (SLA, volumetria, orçamento cloud, aprovação de contrato) foi necessário para o trabalho desta sessão.

## Decisões externas D-01…D-07 e T-R2-01
Não revisitadas nesta sessão — nenhum trabalho desta sessão dependeu delas. Permanecem como registradas nas rodadas anteriores.
