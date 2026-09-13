# Aceite operacional e jornadas

| Área | Jornada / falha | Oráculo obrigatório |
|---|---|---|
| Catálogo | Publicar contrato/serviço/oferta, alteração após aceite, busca acima de 1000 | versão/hash/elegibilidade e referência persistida |
| Cliente | SYNC/ASYNC/AUTO, repetir idempotency key, GET após final | UUIDv7, efeito único, resultado da autoridade local |
| Provedor | sync exposto async, polling+callback, timeout depois de efeito | estado externo por chave/correlação, recibo e reconciliação |
| Callback | órfão falso/autêntico, duplicata, rotação, timestamp inválido | rejeição antes de custódia indevida, aceite durável do válido |
| Produto | dois provedores, dependência, max_parallel=1 entre pods, crash RUNNING | snapshots por etapa, vaga única e retomada |
| Compensação | B depende A; expiração do cliente antes de compensar | ordem reversa, prazo próprio e obrigação financeira |
| Financeiro | reserva versus efetivo, franquia, watermark, tardio/estorno | conservação monetária, corte e disputa consultável |
| Objetos | multipart, FileRef de entrada/saída, pin, restore versões | bytes/hash/versão resolvidos e retenção sem perda |
| Console | timeout/reload, contrato inválido, autorização, teclado | intenção/recibo e efeito real; foco/erro/estado verificáveis |
| Mensageria | namespace vazio, consumidor atrasado, poison financeiro | topologia antes de publicar e conteúdo antes de ACK |
| HA | Redis/cofre/Atlas fora; DB/broker falha; nó/zona perde | budgets, custódia de aceites e retomada sem replay cego |
| Escala | tenants/provedores crescem e agressor satura | isolamento dentro de SLO, quota agregada e escalonamento automático |
| Promoção | manifesto ausente/trocado; imagem diferente; fixture em prd | BLOCK por gate e proveniência do artefato |

Perfis local/dev/hom/ppd/prd devem separar namespace, dados, identidade, segredo, egress e observabilidade. Kind independente atual preservado; banco/objetos usam emptyDir no laboratório e não demonstram durabilidade de produção. Qualificar PVC/backup/restore no perfil durável, mantendo um único Compose do Hub conforme AGENTS.
Não é necessário comprar serviços externos para implementar harness. Fixtures locais devem produzir evidência real do efeito sintético, sem serem chamadas de homologação comercial.
