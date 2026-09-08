# Relatório de execução R2 — em andamento

Este relatório é um registro de execução, não uma declaração de conclusão. A rodada permanece aberta porque há critérios de produção local ainda sem qualificação, especialmente T-R2-01, RLS com papéis de runtime, fluxo integrado atual com broker, DAG, recuperação/restore e ensaios de kind/carga.

## Estado comprovado nesta retomada

O código compilou com `go test ./...` e `go vet ./...`; o frontend compilou com `npm run build`. As suítes PostgreSQL reais e sanitizadas estão em `hub/evidence/r2/execution/`.

Foram comprovados: admissão/idempotência e pin atômico de FileRef; custódia de operação Cometa; inbox/quarentena e ACK seguro de fatos da Órbita; aceitação `PENDING` atômica; consulta de operação com escopo por workload; espera limitada AUTO; polling com política congelada, lease/epoch, takeover, retry e deadline absoluto; capacidade agregada PostgreSQL com concorrência por tenant/célula, rate rolling, pendências UNKNOWN e feedback adaptativo; aritmética/financeiro e objetos conforme os artefatos já indexados.

Os testes usam apenas fixtures sintéticas locais em PostgreSQL 16, LocalStack 3.8, Keycloak 26.7.3 e simuladores locais. Não são aprovação comercial, de provedor ou de produção.

Após a integração, `go test ./...`, `go vet ./...`, `git diff --check` e `npm run build` passaram. Os cinco serviços foram reconstruídos e `/healthz/ready` respondeu 200 nas portas locais. A subida completa de observabilidade foi interrompida por timeout TLS ao baixar Loki do Docker Hub; esse componente permanece sem prova atual. O Keycloak do volume persistente rejeitou a senha da fixture `operadora-a`, então o ensaio autenticado atual não foi repetido e não deve ser confundido com a evidência anterior de imagem antiga.

O diagnóstico confirmou que a conta existe e está habilitada no Keycloak, mas o grant da fixture continua devolvendo `invalid_user_credentials`; a tentativa de redefinição via API administrativa retornou 204 e não resolveu o grant. O volume não foi apagado. Esse bloqueio é reproduzível nos artefatos `http-current-auth-me.json` e nos logs sanitizados do container.

## Limitações ainda abertas

O teste de deadline atual rejeita resultado observado após o prazo, mas não prova que uma confirmação durável cujo commit cruza o limite nunca possa escapar; T-R2-01 segue sem solução aceita. A capacidade agregada possui API e prova de domínio, mas ainda não está conectada ao executor de efeitos externos.

Ainda falta executar e evidenciar o fluxo HTTP completo com as imagens reconstruídas nesta retomada, incluindo catálogo qualificado, Cometa, finalização, Pulsar e Libra. Permanecem pendentes as jornadas completas do console administrativo, callbacks autenticados, RLS/papéis não proprietários, DAG composto, restore, kind com runtime aplicado, carga/elasticidade prolongada e validação estrita das nove changes.

Consulte [CHECKPOINT.md](CHECKPOINT.md), [EVIDENCE_INDEX.md](../../../../hub/evidence/r2/execution/EVIDENCE_INDEX.md), [REQUIREMENTS_STATUS.csv](REQUIREMENTS_STATUS.csv) e [SCENARIO_RESULTS.csv](SCENARIO_RESULTS.csv) para rastreabilidade e limites. Nenhum cenário deve ser marcado como aprovado apenas por correspondência de nome ou existência de teste unitário.
