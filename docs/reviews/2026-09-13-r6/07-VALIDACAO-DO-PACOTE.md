# Validação do pacote R6

Data: 13/09/2026. Snapshot: a540b40007fe6b8ed523e17afe00e96ff8f8ad50.

| Verificação | Resultado |
|---|---|
| OpenSpec 1.12.0 baseline, all/strict | 27/27 PASS |
| OpenSpec 1.12.0 R6 isolada, all/strict | 6/6 PASS |
| OpenSpec 1.12.0 união, all/strict | 33/33 PASS |
| Inventário com IDs únicos | 233 requisitos / 828 cenários |
| Novos requisitos/cenários | 15 / 45 |
| Achados | 15: 10 P0 e 5 P1 |
| Trabalho planejado | 72 tarefas nos changes + 25 verificações herdadas |
| Reavaliação R5 | 17 requisitos reconciliados individualmente |
| Go race / vet / frontend build | Exit 0; race: 139 PASS e 70 SKIP |
| Probe precisão do fato | PASS |
| Probe produto/rota | FAIL reproduzido, convertido em R6-EXE-01 |
| Probe manifesto não confiável | ALLOW indevido reproduzido, convertido em R6-OPE-01 |
| Git diff --check e estado do checkout | Aprovado; nenhuma alteração versionável |
| DB/Compose/kind/browser/carga/HA/restore nesta auditoria | NOT_RUN |

OpenSpec verifica estrutura e sintaxe normativa; não valida implementação nem execução. Os 828 cenários do inventário não são 828 testes executados. Digests dos cenários usam bloco normalizado por strip, UTF-8 e SHA-256; digest do requisito é o do arquivo spec completo.

SHA256SUMS.txt cobre os arquivos entregues, exceto ele próprio. O ZIP contém apenas documentação e changes R6, sem código de produção, node_modules, segredos ou volumes. Logs da revisão se referem ao candidato analisado, não à implementação futura solicitada ao agente.
