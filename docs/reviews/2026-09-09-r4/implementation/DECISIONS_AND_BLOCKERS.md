# Decisões e bloqueios

## Decisões

- Callback externo usa `/callbacks/` público no servidor, chave de ingresso
  configurada por ambiente (`CALLBACK_INGRESS_KEY`) e capability por operação.
  A rota interna continua disponível apenas para compatibilidade de registro,
  mas o processo oficial não a expõe sob o mux JWT.
- API_KEY é aditiva em `0004_provider_api_key.sql`; `0002_provider_auth.sql`
  preserva o byte histórico R2.
- A única convergência automática de ledger aceita a SHA R3 conhecida e
  verifica colunas/constraint antes de auditar a reconciliação.
- L1 é indexado por tenant, binding, conta, ambiente e versão. Tokens nunca
  são gravados no Redis; locks de renovação têm limite de 1024 entradas
  rastreadas e evicção somente ociosa.
- O catálogo sintético é preparado via endpoints versionados `admin/v1`; os
  únicos inserts diretos são a qualificação e a célula do laboratório, porque
  esta baseline não oferece endpoint administrativo para esses dois registros.

## Resolvido ou avançado nesta rodada operacional

- A carga antes bloqueada por `offer_not_eligible` passou após o seed
  versionado, cobrindo oito caminhos e idempotência.
- O Playwright determinístico passou com sessão OIDC/OTP, mutações de
  catálogo e destinos, SLA, logout e viewport 390×844. O Browser Harness foi
  instalado externamente em versão `0.1.13` e passou em CDP explícito,
  percorrendo 16 rotas; a descoberta automática do daemon para Chrome
  headless continua indisponível.
- Kind/HA foi ensaiado com cinco deployments em duas réplicas; perda
  controlada de pods de Cometa e Pulsar foi recuperada sem perda de prontidão.
- Restore comparou contagens/digests dos bancos e objetos sem replay externo.
- A inbox de callback passou a separar a identidade da capability, limitar
  payloads/itens e remover processados de forma limitada; a validação terminal
  ocorre antes da custódia órfã.
- A capacidade foi conectada ao `SUBMIT`, ao `STATUS` de polling e à
  reconciliação. Nesta iteração, o Pulsar também passou a adquirir `FETCH` no
  domínio isolado `r4-webhook`; o lease precisa cobrir o timeout e uma margem
  de segurança antes do POST. Todos esses caminhos validam o fencing antes da
  autenticação e do I/O externo; uma lease expirada não autoriza novo efeito.
  Destinos de webhook agora são congelados na admissão e transportados no fato
  final.
- Resultados que excedem o limite inline agora atravessam a autoridade de
  objetos: upload streaming como `ORPHAN`, `FileRef` na representação final e
  vinculação de retenção após o commit. A falha pós-commit permanece
  reconciliável e não reabre o protocolo.
- A custódia financeira preserva a distinção entre estado operacional e
  incidência: `SUBMIT` aceito externamente continua `UNKNOWN` para a execução,
  mas chega ao Libra como `SUBMITTED`; todo `STATUS` conservado carrega seu
  `attempt_id`. O gate local confirmou fatos, inbox e partidas balanceadas sem
  quarentena de incidência inválida.
- A execução de produtos compostos usa o `command_id` como chave de
  idempotência externa de cada etapa, mantendo o `protocol_id` nas operações
  simples. O probe HTTP comprovou duas etapas independentes, dois efeitos
  distintos e duplicata sem novo efeito.
- O finalizador aceita o snapshot do protocolo quando um intent histórico não
  está disponível. A reconciliação de `UNKNOWN` sem `provider_request_id` é
  rejeitada e auditada com `REJECTED`/`no_provider_correlation`; não há replay
  cego.
- A fila de comandos local usa a autoridade interna do LocalStack após a
  criação idempotente, porque a URL externa retornada pelo simulador não é
  roteável a partir dos containers. O worker registra bytes originais em
  quarentena antes do ACK para envelopes inválidos, deadlines vencidos e
  rejeições determinísticas; capacidade/credencial continuam em redelivery.
- A submissão Cometa agora exige a posse durável da operação (tenant,
  aplicação, célula, epoch, lease e hash do comando) antes da autenticação e
  novamente imediatamente antes do POST. Perda da posse conserva `UNKNOWN`
  sem disparar um segundo efeito.

## Bloqueios e limites ainda reproduzíveis

- O perfil Kind independente agora materializa Postgres, LocalStack, IdP,
  telemetria, provider, sink, gateway e UI dentro do cluster, sem Endpoints
  apontados ao Compose. O perfil Compose-linked continua disponível para
  compatibilidade. Ambos são laboratórios locais: não constituem HA regional,
  volumes duráveis ou IaC dos ambientes remotos.
- O OpenSpec strict `21/21` valida a forma das changes, mas não encerra as
  tarefas funcionais nem os 25 itens herdados de R4-04.
- Permanecem sem homologação provedores comerciais, AWS regional, decisões
  comerciais/SLO externas e a matriz integral de 201 requisitos/732 cenários.
- A jornada HTTP do executor DAG com catálogo/provedor comercial, carga
  prolongada/expiração durante I/O, fencing positivo de efeitos cujo
  `provider_request_id` já foi perdido, políticas produtivas de retenção,
  projeções de catálogo em escala e telemetria bilateral continuam sem
  demonstração integral. O budget efetivo local já está ligado e testado.
- Gaps arquiteturais remanescentes estão listados em `EXECUTION-2026-09-10.md`
  e não foram reclassificados como PASS por inferência.
