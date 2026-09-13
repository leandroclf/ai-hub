# Reavaliação da R5

| Requisito | Situação | Fundamento |
|---|---|---|
| R5-SEG-01 | PARCIAL | Assinatura pré-custódia corrigida; rotação/retomada/isolamento do lock em R6-SEG-02. |
| R5-SEG-02 | IMPLEMENTADO_LOCAL_PARCIAL | Cache-first/validade/negação/limite verificados pela suíte; invalidação emergencial multi-réplica ainda exige prova da janela publicada. Não repetir probe antigo que pressupunha consulta remota em todo hit. |
| R5-SEG-03 | PARCIAL | Migração aditiva existe; uso runtime e oráculo continuam R6-SEG-01. |
| R5-EXE-01 | PARCIAL | QUEUED trata Expired; DIRECT precisa R6-EXE-02. |
| R5-EXE-02 | PARCIAL | Claim atômico avançou; conclusão/counter exige R6-EXE-04. |
| R5-EXE-03 | FALHA_REPRODUZIDA | Snapshot incoerente por etapa: R6-EXE-01. |
| R5-EXE-04 | PARCIAL | DAG reverso adicionado; status público bloqueia: R6-EXE-03. |
| R5-EXE-05 | IMPLEMENTADO_NAO_QUALIFICADO_INTEGRAL | EnsureTopology ligado no startup; ensaiar consumidor tardio/assinatura removida em R6-OPE-03. |
| R5-DAD-01 | PROBE_CORRIGIDO | Precisão do consumidor PASS; representação completa GET/webhook/browser segue qualificação herdada. |
| R5-DAD-02 | ABERTO | Restore sem mudança; R6-OPE-02. |
| R5-DAD-03 | PARCIAL | Captura ligada; completude e ajustes em R6-FIN-02. |
| R5-DAD-04 | PARCIAL | Payload/endpoints ligados; semântica de replay R6-FIN-01. |
| R5-OPE-01 | PARCIAL | Controller nil com domínio é recusado; domínio vazio/feedback em R6-OPE-04. |
| R5-OPE-02 | FALHA_REPRODUZIDA | BLOCK sem manifesto corrigido; ALLOW indevido com manifesto em R6-OPE-01. |
| R5-OPE-03 | ABERTO | Deploy sem alteração relevante nesta rodada; R6-OPE-03. |
| R5-UX-01 | PARCIAL | Busca e guards parciais; R6-UX-01/02. |
| R5-QUA-01 | PARCIAL | Inventário original e logs históricos incompletos; R6-QUA-01. |

## Continuidade normativa
Os 218 requisitos/783 cenários presentes antes desta rodada continuam no inventário. As quatro exigências RUN da R4 regenerada já foram reconciliadas na R5 incorporada; não adicionar novamente o quinto change antigo. R6 não cria uma nova versão de arquitetura nem redefine marcos comerciais.

Todos os 42 achados R2 e 25 achados R3 devem manter a rastreabilidade existente. Ausência de novo delta em um arquivo não comprova fechamento histórico. Ao qualificar, vincular resultado ao cenário do inventário em vez de criar IDs ad hoc como S_BLOCK fora das specs.
