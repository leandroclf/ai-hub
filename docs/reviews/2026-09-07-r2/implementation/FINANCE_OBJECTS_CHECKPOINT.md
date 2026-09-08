# Checkpoint R2-06 / R2-07

SHA de origem: a39d394b0d87185ed4cc3861c12ec45f2c302d9e, branch main com alterações locais originais preservadas. Leituras: instruções, prompt integral, metodologia, changes 06/07, contratos e baseline financeiro/dados. Root coordena R2-01/02/03 e integração.

Implementação iniciada: `hub/internal/libra/contracts.go`, tipos Snapshot/PricingRule/EconomicEvent e decimal textual exato usando big.Rat, preço unitário/pacote/franquia/faixas explícitas. Ainda não testado. Próximos passos: migrations 0030 finance/core; Store com reserva estrita cumulativa, inbox+fato+journal atômicos, ajustes/fechamento; consumidores ACK posterior; APIs autorizadas; catálogo S3 por tenant e retenção; testes PostgreSQL/S3 reais.

Interfaces acordadas: eventos incluem `economic_snapshot` e identidade tenant/protocolo/operação/tentativa/evidência; root congela no aceite. Auth root fornece `auth.FromContext`, `auth.Authorize`; escopos finance:read/write/reserve/approve. Catálogo FileRef registrado por root em Órbita. Planos/preços/retenção das fixtures não são aprovação comercial. P-03/P-04/P-06/P-08 permanecem pendentes para ativação externa.

Laboratório: Postgres localhost:15432, bases hub_core/hub_finance; S3/SQS localhost:14566. Configuração de fixture em `hub/deploy/r2`; sem credencial real. Reexecutar migrations por compose com coordenação de operations. Nenhum cenário declarado PASS até evidência produzida.
