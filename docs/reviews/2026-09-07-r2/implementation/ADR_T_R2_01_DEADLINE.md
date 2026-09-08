# T-R2-01 — Prova temporal, limite do modelo e caminho de implementação

Estado: decisão técnica proposta; contraexemplos executados; requisito **não qualificado**. Data: 2026-09-08. Responsáveis: Core/Dados e SRE. Rastreio: EXE-11, EXE-12, R2-EXE-06-S01 a S04, IT-08, F-09, tarefas 02:2.6/3.6.

## Resultado verificável

`Store.Finalize` consulta `clock_timestamp()` antes de atualizar protocolo/outbox e confirmar a transação. Isso rejeita resultados já atrasados, mas não impede que uma confirmação iniciada no prazo atravesse o limite. O experimento [deadline-commit-probe.json](../../../../hub/evidence/r2/execution/deadline-commit-probe.json) reproduziu essa janela em PostgreSQL 16.15 com `fsync=on`, `synchronous_commit=on` e tabelas persistentes registradas no WAL. A checagem ocorreu às 16:28:17.201480Z, o prazo era 16:28:17.351480Z e a observação após COMMIT ocorreu às 16:28:17.609117Z. Um trigger de constraint diferido executou 400 ms de atraso durante processamento do COMMIT. Não é uma simulação de latência física do disco: é um atraso concreto entre decisão SQL e conclusão do commit.

Reprodução, na raiz do repositório:

```sh
python3 hub/deploy/r2/tests/deadline-commit-probe.py
```

O script usa exclusivamente schema aleatório próprio, remove apenas esse schema ao terminar e não altera tabelas de domínio ou configurações globais. SHA e estado de diff estão na evidência. O resultado é `COUNTEREXAMPLES_REPRODUCED_NOT_REQUIREMENT_PASS`, nunca aprovação de cenário funcional. Falha precoce do script pode deixar seu schema próprio para inspeção.

## Obrigação que precisa ser preservada

Há três instantes diferentes: custódia de bytes de candidato, confirmação de elegibilidade terminal e disponibilização contratual do final no Hub. EXE-11 exige representação validada e durável estritamente antes do prazo e estabelece máximo entre aceite e disponibilização final. Apenas chamar um candidato privado de “final disponível” muda esse contrato. Eventos de sucesso não podem escapar antes da prova e um final já servido não pode virar EXPIRED após crash.

`synchronous_commit=on` faz o ACK esperar o flush local do WAL; não constitui garantia de limite superior do tempo de commit. A topologia configurada determina se há também espera por réplica. [Documentação PostgreSQL 16 sobre WAL](https://www.postgresql.org/docs/16/runtime-config-wal.html).

O timestamp do registro de commit não deve ser tratado como hora do flush físico: `RecordTransactionCommit` cria o registro e somente depois chama `XLogFlush`, seguido pela marcação CLOG e eventual espera por réplica. Consequentemente `track_commit_timestamp` não substitui um testemunho posterior à durabilidade. [Implementação PostgreSQL 16, xact.c](https://github.com/postgres/postgres/blob/REL_16_STABLE/src/backend/access/transam/xact.c).

## Modelo explicitamente considerado

- PostgreSQL permanece autoridade e recebe transações SQL normais.
- Processo, rede e armazenamento podem sofrer pausa finita, sem limite superior conhecido; o processo pode cair em qualquer fronteira.
- Não existe certificado configurado de erro máximo do relógio nem garantia de continuidade entre reinícios, troca de host ou failover.
- GET, timer, outbox, callback e reconciliação podem executar simultaneamente em réplicas diferentes.

Neste modelo, uma comparação anterior ao último commit que torna o sucesso visível não basta: duas execuções têm o mesmo prefixo até a última comparação. Em uma, o commit termina antes do prazo; na outra, a pausa ocorre depois da comparação e ele termina depois. O estado escrito e a decisão anteriores à pausa são iguais. Sem testemunho posterior ou limite de pausa comprovado, o algoritmo não distingue os casos e não pode permitir o primeiro sem também permitir o segundo. Isso é um argumento sobre esta classe de algoritmos, não uma alegação de impossibilidade de todos os sistemas de tempo real.

Adicionar margem fixa, timeout de cliente/statement, `SERIALIZABLE`, lock exclusivo, timer mais frequente ou trigger diferido com nova comparação apenas desloca a última comparação. Nenhum deles limita o atraso restante de flush/commit. Cancelamento pode deixar o resultado do commit incerto e não autoriza tratá-lo como rollback.

## Candidato durável + certificado posterior: útil, mas insuficiente sozinho

Uma implementação parcial segura pode conservar candidato sem emitir sucesso:

1. Sob lock/epoch do protocolo, gravar representação completa, hash, perfil, versão, observação e origem em candidato imutável. Nenhum evento de sucesso nesta transação.
2. Confirmar com a política de durabilidade requerida. ACK perdido deixa candidato incerto; não reenviar provedor.
3. Após ACK, obter amostra da mesma autoridade temporal, com intervalo `[lower, upper]`, epoch, validade e limite de deriva. Somente `upper < client_deadline_at` pode certificar a custódia anterior ao prazo. Limite desconhecido, igualdade, regressão ou epoch incompatível recusam a certificação.
4. Persistir certificado vinculado a candidato/hash/protocolo/tenant/cell/epoch e ao nível de durabilidade realmente observado. A transação de promoção disputa com timer sob a mesma versão e grava terminal e outbox atomicamente. Candidato ou certificado de epoch antigo não autoriza promoção.
5. Candidato sem certificado válido conserva evidência e obrigação externa/financeira. Timer não apaga essa custódia; recuperação não deduz sucesso nem ausência de efeito.

Esse protocolo prova que os bytes já estavam duráveis quando a amostra foi obtida, se o relógio estiver qualificado. Não prova que o primeiro final contratual ficou disponível antes do prazo: a persistência do certificado/promoção pode ser lenta. O segundo caso do experimento confirma ACK de candidato anterior ao prazo e promoção posterior. O campo de indisponibilidade pública nesse caso explicita a hipótese do desenho; o ensaio é SQL, não teste de API.

Expor sucesso a partir de certificado apenas em memória evita a espera da promoção, mas introduz outro contraexemplo: resposta servida, crash antes da persistência do certificado, recuperação sem prova e timer expirando. Evitar a expiração para sempre elimina progresso; inferir sucesso elimina a proteção temporal. Assim, não adotar essa alternativa como resolução de T-R2-01 sem uma prova recuperável adicional.

## Proteção executável recomendada agora

Implementar recusa conservadora no caminho de sucesso, mantendo as demais frentes ativas:

- Capacidade temporal começa `UNQUALIFIED`; readiness global pode continuar, mas oferta que exige sucesso com deadline rígido não é anunciada como qualificada para essa capacidade.
- Antes de qualquer `SUCCEEDED`/`PARTIALLY_SUCCEEDED`, exigir prova da autoridade temporal e do mecanismo completo qualificado. Na ausência, conservar candidato/observação e auditoria `TEMPORAL_PROOF_UNAVAILABLE`; não gravar outbox de sucesso ou receita de sucesso.
- Admissões que dependem exclusivamente dessa capacidade e ainda não existem devem ser recusadas antes de efeito. Protocolos já aceitos continuam consultáveis por UUID e encerram por contrato, preservando reconciliação de efeitos/custos.
- GET, relays e consumidores financeiros precisam da mesma regra: corrigir somente o retorno HTTP de `Finalize` depois do commit não impede o relay de publicar estado inválido já confirmado.
- Não fornecer flag `assume_clock_safe=true` que torne uma fixture aprovação. Configuração de epsilon sem monitor/fonte/epoch é hipótese de teste, não atestado temporal.

Essa proteção resolve a emissão de sucesso **não comprovado**, prevista expressamente em EXE-11; não entrega a capacidade positiva SYNC/ASYNC com sucesso no prazo e não fecha R2-EXE-06 inteiro. Sua implementação de domínio permanece a cargo do integrador; este ADR não alterou esses arquivos.

## Trabalho necessário para habilitar sucesso

Construir e qualificar uma autoridade que conserve prova recuperável de durabilidade **e** disponibilização antes do limite, ou um mecanismo de commit condicionado ao deadline com limites de falha realmente certificados. A prova deve incluir observação posterior ao flush, ordenação com o terminal/expiração, recuperação após perda de ACK, política de epoch e intervalo de incerteza do relógio. Uma extensão de armazenamento/testemunho durável pode ser investigada; não presumir que um hook de commit SQL tenha essas propriedades. Usar relógio monotônico apenas local não cobre restart/failover nem concede relógio civil autoritativo.

Gates mínimos: commit cruza prazo; último instante igual ao limite; relógio recua/avança/perde saúde; crash antes/depois de candidato/certificado/promoção; perda de ACK; timer/callback/poll/GET/relay concorrentes; réplica atrasada; prova com hash ou epoch divergente; sucesso já observado sobrevive restart com os mesmos bytes; nenhum terminal duplo ou captura de receita por candidato rejeitado. Além disso, comprovar quando o final passou a ser recuperável pela API, sem substituir esse oráculo por timestamp gravado antes do commit.

Não há credencial comercial necessária para investigar essa parte; ainda existe trabalho técnico independente. O limite aqui demonstrado impede declarar o algoritmo SQL atual suficiente. Uma alteração autorizada para SLA medido na decisão, ou para custódia privada anterior com publicação posterior, seria mudança explícita de EXE-11, não está autorizada implicitamente e não foi adotada.
