# Matriz de risco

| Risco | Prioridade | Mitigação | Responsável | Evidência exigida |
|---|---|---|---|---|
| Provas novas executadas: null foi aceito como string; "123" como integer; e dois documentos concatenados retornaram sem erro. Decoder lê apenas o primeiro documento e, sem mapping, devolve input original. A preservação do inteiro grande anterior foi corrigida e não deve regredir. | P0 | R4-CTR-01 | Core e Catálogo | R4-CTR-01-S01/S02/S03 |
| Prova executada: minimum:0 é ignorado e -1 aceito. A validação de publicação verifica só type object/properties. O comentário de recusa a construções não qualificadas não corresponde a whitelist efetiva. integer rejeita representação com .eE e enum compara json.Number lexicalmente, divergindo da semântica usual de JSON Schema. | P1 | R4-CTR-02 | Core e Catálogo | R4-CTR-02-S01/S02/S03 |

Probabilidade quantitativa não medida. Evidência estática não é incidente observado.
