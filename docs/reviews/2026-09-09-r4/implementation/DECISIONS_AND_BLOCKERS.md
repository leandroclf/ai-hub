# Decisões e bloqueios

## Decisões

- Callback externo usa `/callbacks/` público no servidor, uma chave de ingresso configurada por ambiente (`CALLBACK_INGRESS_KEY`) e capability por operação. A rota interna continua disponível apenas para compatibilidade de registro, mas o processo oficial não a expõe sob o mux JWT.
- API_KEY é aditiva em `0004_provider_api_key.sql`; `0002_provider_auth.sql` volta ao byte histórico R2.
- A única convergência automática de ledger aceita a SHA R3 conhecida e verifica colunas/constraint antes de auditar a reconciliação.
- L1 é indexado por tenant, binding, conta, ambiente e versão; Redis não recebe tokens.

## Bloqueios reproduzíveis

- Em 10/09/2026, a sessão OIDC com senha+OTP e a validação do token passaram,
  mas a carga autorizada recebeu `403 offer_not_eligible`: o banco atual está
  sem `catalog_resources` e `catalog_publications`. O endpoint legado de
  cadastro também respondeu `410 use_versioned_catalog`. O gerador agora aceita
  `R2_BEARER_TOKEN` sem imprimir o segredo. Não foi feito seed direto por SQL;
  é necessário fornecer uma carga versionada de catálogo e repetir o gate.
- O Browser Harness opcional não está instalado (`run.sh` retorna código 2).
  O smoke Playwright nativo passou integralmente e a integração permanece
  documentada para execução após provisionamento da ferramenta.

- PostgreSQL/Compose e restore foram executados; o laboratório kind foi
  recriado e seus gates locais passaram. Permanecem pendentes a qualificação
  de HA, dependências completas dentro do cluster, carga funcional com
  catálogo publicado e os cenários externos de produção.
- O OpenSpec strict atual passou em 21/21 changes, mas isso não encerra tarefas funcionais nem os 25 itens herdados de R4-04.
- `go test` local usa Go 1.22.6; a baseline documenta Go 1.24.13. A diferença precisa ser resolvida no ambiente fixado antes do gate final.
- Não há decisão comercial/SLO externa autorizada para fechar os gates de homologação.
