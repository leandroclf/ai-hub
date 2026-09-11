# Checkpoint R4

Data: 2026-09-11 (America/Sao_Paulo)

HEAD do checkpoint: consultar `git log` após o commit de consolidação desta
rodada. O código do adapter REST está em `ac9b346` e as evidências integradas
mais recentes foram consolidadas em `94afa58`.

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
comprovadas (`closed=10 protected=0`). A carga autorizada mais recente usou
`r4-authorized-1789127281502`; restore, RLS, financeiro, webhook e continuidade
também foram repetidos com resultado PASS.

Estado atual: sequência operacional local concluída, incluindo a conexão HTTP
do DAG, o editor administrativo de mapeamento, a fronteira `ProviderAdapter`
com `rest-json-v1` e a qualificação complementar seletiva de R2-03/R2-02. A
matriz agora possui 142 resultados associados e 590 cenários explicitamente não
qualificados. A prova estrutural de R3-OPE-02-S03 renderizou os cinco overlays
Kustomize e confirmou imagens fixadas, namespaces determinísticos e ausência de
modo local no `prd`; ela não promove os cenários de máquina limpa ou recriação
de nós. Permanecem explícitos os gaps
arquiteturais do backlog herdado, dependências Compose no renderer do kind,
ausência de homologação de provedores/AWS reais, carga prolongada/expiração de
I/O e matriz integral ainda não fechada. O adapter REST foi qualificado somente
com endpoint HTTP local independente e não representa provedor comercial. Ver
`EXECUTION-2026-09-10.md` para os oráculos e limites.
