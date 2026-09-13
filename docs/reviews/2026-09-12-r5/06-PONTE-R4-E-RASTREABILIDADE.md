# Ponte entre revisões

Não confundir arquitetura v4 com rodada R4. A baseline remota é a fonte normativa existente; quatro obrigações da edição regenerada são mantidas explicitamente pela R5.

| R4 | Estado | Avanço/limite | Destino |
|---|---|---|---|
| R4-CBK-01 | IMPLEMENTADO_COM_NOVA_LACUNA | Rota pública sem JWT Hub corrigida; origem órfã ainda sem verificação | R5-SEG-01 |
| R4-CBK-02 | PARCIAL | Dedupe/limites novos; HMAC órfão não verificado, quota global e retenção RECEIVED | R5-SEG-01 |
| R4-CBK-03 | PARCIAL | Worker autônomo e lote existem; escopo/órfão permanente/rotação faltam | R5-SEG-01 |
| R4-CBK-04 | IMPLEMENTADO_REQUALIFICAR_FLUXOS | Validação comum/correlação/bytes avançaram; homologar entrada externa completa | R5-SEG-01,R5-DAD-01 |
| R4-CTR-01 | CORRECAO_REPRODUZIDA | Todos os contraexemplos JSON anteriores agora passam | R5-DAD-01: perda em outra fronteira |
| R4-CTR-02 | CORRECAO_REPRODUZIDA | Dialeto/keywords/limites/número exato implementados | Preservar regressões existentes |
| R4-OPE-01 | CORRECAO_REPRODUZIDA | Probe adaptado de L1 válido/cofre indisponível passa; locks/revogação têm testes | Requalificar rotação por requisito herdado |
| R4-OPE-02 | CORRIGIDO_NO_CODIGO_PROVA_HISTORICA | 0002 restaurada; reconciliação restrita. Upgrade integrado não reexecutado aqui | Preservar checksums/ensaios |
| R4-OPE-03 | PARCIAL_COM_REGRESSAO | Seleção indexada existe; fallback 403 indevido reproduzido | R5-SEG-02 |
| R4-QUA-01 | PARCIAL | Inventário 732 existe, sem inventar integralidade; digest de resultado falta | R5-QUA-01 |
| R4-QUA-02 | PARCIAL_COM_AVANCO | Console muito alterado e smokes reais registrados | R5-UX-01 |
| R4-QUA-03 | PARCIAL_COM_AVANCO | Kind independente e integração registrados; gates atuais limitados | R5-OPE-02,R5-OPE-03 |
| R4-RUN-01 | NAO_PRESENTE_COMO_SPEC_REMOTA | Primeira falha existe; claim não cerca retry_until | R5-EXE-01 |
| R4-RUN-02 | NAO_PRESENTE_COMO_SPEC_REMOTA | DAG persistido avançou; slots/rotas/compensação a fechar | R5-EXE-02,R5-EXE-03,R5-EXE-04 |
| R4-RUN-03 | NAO_PRESENTE_COMO_SPEC_REMOTA | Helper ainda não adotado e prova RLS insuficiente | R5-SEG-03 |
| R4-RUN-04 | NAO_PRESENTE_COMO_SPEC_REMOTA | Restore ainda não qualifica retomada/versões/financeiro | R5-DAD-02 |

## Regras de contagem
Baseline remota: 201 requisitos/732 cenários. R5: 17/51. União: 218/783.
Os quatro RUN ausentes são mapeados semanticamente; não somar novamente 4/12 como se estivessem no repositório.
Não importar status antigo por título: comparar cenário, código, hash e oráculo. As matrizes R2/R3 preservam achados históricos; não são novos requisitos.
