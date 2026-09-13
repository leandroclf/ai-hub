# Validação do pacote R5

Revisão em 12/09/2026; código observado: f87ce33034ae29c9431b1910dcc6a633b545e330.

| Verificação | Resultado |
|---|---|
| OpenSpec 1.12.0, R5 isolada, all/strict | 6 changes aprovados; zero falhas |
| OpenSpec 1.12.0, baseline remota + R5, all/strict | 27 changes aprovados; zero falhas |
| Novos requisitos / cenários / tarefas | 17 / 51 / 88 |
| Inventário consolidado, IDs únicos | 218 requisitos / 783 cenários |
| Reavaliação histórica | 42 achados R2; 25 R3; 16 requisitos R4 (12 remotos + 4 da edição regenerada) |
| Código de produção nesta revisão | Sem alterações; git diff --check aprovado |
| Go race | Comando aprovado: 137 testes PASS, 69 SKIP |
| Go vet / build TypeScript-Vite | Aprovados, dentro dos limites do documento 04 |
| Probes de cache após 403 / precisão de fatos | Falhas reproduzidas e convertidas em requisitos |
| Probes JSON e token L1 | Aprovados |
| Gate de promoção com entradas não verificadas | ALLOW reproduzido; tratado como falha do gate |
| Integração real / browser / HA / carga / restore | Não executados nesta auditoria |

OpenSpec valida a estrutura dos changes; não certifica a implementação. O inventário não transforma cenários em testes aprovados. Cenários R5 têm linha e digest do bloco normalizado (strip; UTF-8; SHA-256); requisitos têm digest do arquivo spec completo. Logs anexos preservam o resultado observado, inclusive falhas e skips.

O manifesto SHA256SUMS.txt cobre cada arquivo do pacote, exceto ele próprio. O ZIP contém somente documentação da revisão e os seis changes novos. Não inclui código, node_modules, credenciais, volumes ou checkout.
