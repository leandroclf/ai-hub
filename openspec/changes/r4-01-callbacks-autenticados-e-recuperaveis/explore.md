# Explore — Callbacks autenticados e recuperáveis

Snapshot b9d0f90ce02aa0c27cad546745153d160ff5867f; revisão incremental v4/R2/R3.

## F-R4-01
O handler ganhou capability por operação e custódia antes de 2xx. Entretanto, toda a rota /internal/ continua envolvida pelo JWT do Hub; a URL entregue ao provedor contém apenas capability. O simulador não envia JWT Hub. A correção do handler não fecha o caminho de rede. Isso é incompatibilidade de composição, não prova de endpoint publicamente desprotegido.

[hub/cmd/cometa/main.go:104](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/cmd/cometa/main.go#L104), [hub/internal/cometa/handlers.go:116](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/handlers.go#L116), [hub/internal/cometa/executor.go:273](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/executor.go#L273), [hub/internal/platform/httpserver/httpserver.go:94](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/platform/httpserver/httpserver.go#L94)

## F-R4-02
Quando a operação não existe, qualquer token não vazio que alcance o handler permite inserir body e receber 202. A chave única usa operação+hash do body, sem identidade autenticada/token: primeira tentativa com token incorreto pode ocupar a identidade de posterior recibo correto, que só incrementa occurrences. A tabela também não registra tenant/conta/célula, TTL ou quota. Exploração externa depende da rota/autenticação atual; o risco permanece ao corrigir essa rota.

[hub/internal/cometa/handlers.go:98](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/handlers.go#L98), [hub/internal/cometa/custody.go:114](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/custody.go#L114), [hub/migrations/core/0032_callback_inbox.sql:14](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/migrations/core/0032_callback_inbox.sql#L14)

## F-R4-03
ReconcileCallbackInbox só é chamado ao receber outro callback conhecido. A consulta percorre todos os RECEIVED associados, sem LIMIT, claim, lease ou escopo de célula. O cursor permanece aberto enquanto são feitas outras operações SQL e aplicação. O HTTP de uma operação pode depender do backlog/erro de outra, e sem novo callback não há gatilho autônomo.

[hub/internal/cometa/custody.go:131](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/custody.go#L131), [hub/internal/cometa/handlers.go:157](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/handlers.go#L157)

## F-R4-04
finalize valida OutputSchema e conserva OperationResult. ApplyExternalObservation transforma callback somente em detail, não chama a mesma validação e não confere provider_request_id contra correlação armazenada. ConserveObservation pode sobrescrever a correlação com qualquer ID não vazio recebido. Token de operação não torna o body semanticamente correto.

[hub/internal/cometa/executor.go:319](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/executor.go#L319), [hub/internal/cometa/executor.go:232](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/executor.go#L232), [hub/internal/cometa/custody.go:288](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/custody.go#L288)

Preservar correções anteriores. Nenhum achado estático é relatado como incidente de produção.
