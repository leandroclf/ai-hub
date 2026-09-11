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
  instalado externamente em versão `0.1.13`, mas o daemon não encontrou um
  Chrome com `DevToolsActivePort` utilizável e permanece bloqueado pelo
  ambiente.
- Kind/HA foi ensaiado com cinco deployments em duas réplicas; perda
  controlada de pods de Cometa e Pulsar foi recuperada sem perda de prontidão.
- Restore comparou contagens/digests dos bancos e objetos sem replay externo.
- A inbox de callback passou a separar a identidade da capability, limitar
  payloads/itens e remover processados de forma limitada; a validação terminal
  ocorre antes da custódia órfã.
- A capacidade foi conectada ao `SUBMIT`, ao `STATUS` de polling e à
  reconciliação. Nesta iteração, o Pulsar também passou a adquirir `FETCH` no
  domínio isolado `r4-webhook`; o lease precisa cobrir o timeout e uma margem
  de segurança antes do POST. Destinos de webhook agora são congelados na
  admissão e transportados no fato final.
- A custódia financeira preserva a distinção entre estado operacional e
  incidência: `SUBMIT` aceito externamente continua `UNKNOWN` para a execução,
  mas chega ao Libra como `SUBMITTED`; todo `STATUS` conservado carrega seu
  `attempt_id`. O gate local confirmou fatos, inbox e partidas balanceadas sem
  quarentena de incidência inválida.

## Bloqueios e limites ainda reproduzíveis

- O kind local ainda resolve Postgres, LocalStack, IdP, telemetria, provider e
  sink por Endpoints apontados para containers do Compose; isso é um laboratório
  integrado, não um cluster independente de produção.
- O OpenSpec strict `21/21` valida a forma das changes, mas não encerra as
  tarefas funcionais nem os 25 itens herdados de R4-04.
- Permanecem sem homologação provedores comerciais, AWS regional, decisões
  comerciais/SLO externas e a matriz integral de 201 requisitos/732 cenários.
- O executor DAG ligado ao atendimento, fencing geral de efeitos UNKNOWN,
  budgets integrais de I/O, resultados de FileRefs no adapter real, projeções
  de catálogo em escala e telemetria bilateral continuam sem demonstração
  integral.
- Gaps arquiteturais remanescentes estão listados em `EXECUTION-2026-09-10.md`
  e não foram reclassificados como PASS por inferência.
