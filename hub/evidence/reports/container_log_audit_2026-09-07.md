# Auditoria de logs dos containers — correção do healthcheck PostgreSQL

Gerado em UTC: 2026-09-07T11:12:43Z

## Escopo
Containers auditados: postgres, localstack, atlas, orbita, cometa, pulsar, libra, provider-sim, webhook-sink, swagger-ui e kong.

## Classificação da janela de 2 minutos (inclui o instante do restart)
Durante a recriação do PostgreSQL e a reconexão dos serviços foram observados 4
`ERROR` transitórios: Pulsar perdeu o alias DNS por um instante, e os loops de
outbox/deadline do Cometa/Órbita receberam `connection refused`. Após a estabilização,
uma janela limpa de 30 segundos apresentou `FATAL=0`, `ERROR=0`, `WARN=0`, `panic=0` e
`exception=0`.

## Causa raiz confirmada
O healthcheck anterior executava pg_isready -U hub. Sem -d, o PostgreSQL tentava a base hub, inexistente neste compose; isso gerava FATAL periódico embora o servidor estivesse saudável.

## Correção
O healthcheck agora executa pg_isready -U hub -d postgres, usando a base bootstrap definida por POSTGRES_DB.

## Verificação
{
  "status": "healthy",
  "failing_streak": 0,
  "latest": {
    "Start": "2026-09-07T11:12:42.284231853Z",
    "End": "2026-09-07T11:12:42.335479302Z",
    "ExitCode": 0,
    "Output": "/var/run/postgresql:5432 - accepting connections\n"
  }
}

## Resultado
O ruído fatal periódico foi eliminado pelo healthcheck corrigido. Os 4 erros de
reconexão são esperados no instante de uma substituição forçada do banco; não houve
erro persistente após os serviços serem reiniciados com o PostgreSQL saudável.
