# Checkpoint R4

Data: 2026-09-11 (America/Sao_Paulo)

HEAD do checkpoint: consultar `git log` após o commit de consolidação desta
rodada. A evidência mais recente desta rodada registra o ajuste de egress da
bridge Docker, a resolução dinâmica dos upstreams do portal e a reexecução
limpa da carga autorizada, do contrato legado e do financeiro.

Executado: OpenSpec strict `21/21` — PASS; `npm run build` no portal — PASS;
`go test -race -count=1 ./...`, `go vet ./...` e `git diff --check` — PASS. O
Compose `ai_hub_r3qual` passou por seed versionado, carga autorizada, callbacks,
migrações, RLS, restore e testes integrados. O restore comparou digests dos três
bancos e S3 com alvo novo e sem replay externo. O cluster kind tem três nós,
readiness, métricas/HPA/KEDA e recuperação de pods em duas réplicas. O smoke
Chromium autenticado passou também pelo editor de produto e seu mapeamento entre
etapas. O probe HTTP do produto passou com dois efeitos independentes e
idempotência; o teste de budget efetivo passou. O Browser Harness percorreu as
16 rotas com `BU_CDP_URL` explícito em viewport 390×844 e foi classificado
`PASS-EXPLORATORY`; a descoberta automática do daemon headless continua
indisponível.

Atualização posterior: o commit `0a57029` corrigiu a autoridade interna da
fila LocalStack, a quarentena durável de envelopes inválidos/comandos
expirados e a preservação de falhas recuperáveis para redelivery. A prova do
produto foi repetida após reconciliação positiva da fixture local e passou com
duas etapas, dois efeitos e idempotência; o runner fechou dez ausências
comprovadas (`closed=10 protected=0`) naquela execução. Nesta rodada, uma
reconciliação posterior fechou oito ausências confirmadas
(`closed=8 protected=0`). A carga autorizada limpa
usou `r4-authorized-20260911clean` e passou nos oito caminhos, com quatro
observações idempotentes; o financeiro passou com `outbox=27`, `facts=13`,
`cost=4`, `revenue=9`, `inbox=27`, sem lançamentos desbalanceados ou
quarentena inválida. O contrato legado passou com limpeza das ofertas
temporárias publicadas.

R2-04/4.1 também foi comprovado em ensaio cercado: a migration aditiva
materializou dois serviços legados como drafts rastreáveis, o recuo suspendeu
somente v2 e preservou v1, histórico de publicação e protocolo aceito com
snapshot v1; o replay foi idempotente. Os bancos temporários foram removidos
sem alterar o volume oficial.

R2-07 também avançou com a qualificação integral local de R2-DAD-03:
retenção por classe respeitou pin de obrigação, expurgo produziu tombstone e o
replay antigo foi deduplicado sem nova finalização/outbox; falha de storage
permaneceu reconciliável. R2-DAD-04/05 seguem limitados por continuidade de
leitura, writer alternativo, fencing e HA regional não comprovados.

Em R2-DAD-04, a indisponibilidade da autoridade de leitura agora retorna
`503/protocol_unavailable` em vez de `500/internal_error` ou `404`; a regressão
foi coberta com PostgreSQL fechado, e o RLS cross-tenant foi revalidado nas três
bases. A implementação local está corrigida, mas réplica atrasada e writer
alternativo regional continuam fora do perfil de laboratório.

No R2-08, a plataforma local persistente e o cenário de dependência opcional
desligada foram revalidados: o Compose oficial manteve migrations idempotentes,
volumes e workloads de negócio, enquanto Alloy/Redis fora não causaram cascata.
Os gates de cinco ambientes, escala efetiva e HA regional continuam abertos.

R2-DAD-05/2.5 foi promovido como comportamento local implementado: o restore
isolado compara contagens e digests, bloqueia egress/readmissão durante a
reconciliação e consulta o oráculo externo sem replay. A prova de falha
regional/fencing de R2-DAD-05-S03 permanece aberta.

Estado atual: sequência operacional local concluída, incluindo a conexão HTTP
do DAG, o editor administrativo de mapeamento, a fronteira `ProviderAdapter`
com `rest-json-v1` e a qualificação complementar seletiva de R2-03/R2-02. A
matriz agora possui 143 resultados associados e 589 cenários explicitamente não
qualificados. A prova estrutural de R3-OPE-02-S03 renderizou os cinco overlays
Kustomize e confirmou imagens fixadas, namespaces determinísticos e ausência de
modo local no `prd`; ela não promove os cenários de máquina limpa ou recriação
de nós. Permanecem explícitos os gaps
arquiteturais do backlog herdado, a reexecução independente do Kind com
dependências `cluster-owned`,
ausência de homologação de provedores/AWS reais, carga prolongada/expiração de
I/O e matriz integral ainda não fechada. O adapter REST foi qualificado somente
com endpoint HTTP local independente e não representa provedor comercial. Ver
`EXECUTION-2026-09-10.md` para os oráculos e limites.

Correção operacional desta rodada: as regras padrão de egress do Compose agora
consideram a bridge privada real (`172.16.0.0/12`) sem abrir host/porta além dos
destinos declarados. O Nginx do portal resolve os nomes Docker com TTL curto e
reescreve explicitamente os prefixos `/api/*`, evitando upstream obsoleto após
`--force-recreate`. A prova Chromium voltou a completar OIDC/PKCE, senha+OTP e
as 24 verificações do portal (23 `PASS`, 1 `OBSERVED`, nenhum `FAIL`).
