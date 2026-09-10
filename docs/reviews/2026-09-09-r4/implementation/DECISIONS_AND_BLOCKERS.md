# Decisões e bloqueios

## Decisões

- Callback externo usa `/callbacks/` público no servidor, uma chave de ingresso configurada por ambiente (`CALLBACK_INGRESS_KEY`) e capability por operação. A rota interna continua disponível apenas para compatibilidade de registro, mas o processo oficial não a expõe sob o mux JWT.
- API_KEY é aditiva em `0004_provider_api_key.sql`; `0002_provider_auth.sql` volta ao byte histórico R2.
- A única convergência automática de ledger aceita a SHA R3 conhecida e verifica colunas/constraint antes de auditar a reconciliação.
- L1 é indexado por tenant, binding, conta, ambiente e versão; Redis não recebe tokens.

## Bloqueios reproduzíveis

- PostgreSQL/Compose e restore foram executados; kind completo permanece bloqueado por estado residual de DNS/rede no cluster existente, browser autenticado atual falhou por timeout e carga autorizada ainda não foi executada.
- O OpenSpec strict atual passou em 21/21 changes, mas isso não encerra tarefas funcionais nem os 25 itens herdados de R4-04.
- `go test` local usa Go 1.22.6; a baseline documenta Go 1.24.13. A diferença precisa ser resolvida no ambiente fixado antes do gate final.
- Não há decisão comercial/SLO externa autorizada para fechar os gates de homologação.
