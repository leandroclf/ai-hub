# Explore — Contratos JSON estritos

Snapshot b9d0f90ce02aa0c27cad546745153d160ff5867f; revisão incremental v4/R2/R3.

## F-R4-05
Provas novas executadas: null foi aceito como string; "123" como integer; e dois documentos concatenados retornaram sem erro. Decoder lê apenas o primeiro documento e, sem mapping, devolve input original. A preservação do inteiro grande anterior foi corrigida e não deve regredir.

[hub/internal/atlas/offers.go:179](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L179), [hub/internal/atlas/offers.go:330](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L330)

## F-R4-06
Prova executada: minimum:0 é ignorado e -1 aceito. A validação de publicação verifica só type object/properties. O comentário de recusa a construções não qualificadas não corresponde a whitelist efetiva. integer rejeita representação com .eE e enum compara json.Number lexicalmente, divergindo da semântica usual de JSON Schema.

[hub/internal/atlas/offers.go:266](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L266), [hub/internal/atlas/offers.go:355](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L355), [hub/internal/atlas/catalog.go:327](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/catalog.go#L327)

Preservar correções anteriores. Nenhum achado estático é relatado como incidente de produção.
