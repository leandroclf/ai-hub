# FINDINGS_STATUS — R5

Cada achado do `01-RELATORIO-REVISAO.md` (F-R5-01..17), estado real observado
nesta sessão sobre o SHA `f87ce33` + working tree.

| Achado | Requisito | Estado antes desta sessão | Ação nesta sessão | Estado agora |
|---|---|---|---|---|
| F-R5-01 | R5-SEG-01 | Auth de callback órfão já implementada no working tree, mas `storeOrphan` tinha SQL inválido (`ON CONFLICT DO UPDATE` duplicado) que quebraria em runtime para qualquer callback com `provider_account_id`. | Corrigido o alvo de conflito para `(operation_id, body_sha256, token_hash)`. `go build`/`go vet` limpos. | Corrigido; sem teste de cenário S01/S02/S03 dedicado novo. |
| F-R5-02 | R5-SEG-02 | Client de oferta já respeitava janela de autorização e invalidava em 403/409; cache sem limite de tamanho; `OfferAuthoritative`/`InvalidateOffer` sem chamador. | Cache limitado (`maxCacheEntries`); 2 testes novos provando S01 e o limite. | Regressão específica coberta e passando. |
| F-R5-03 | R5-SEG-03 | Migration 0049 (RLS estrita, aditiva) já presente não commitada. | Executado `rls-runtime-proof.sh` contra Postgres real desta sessão — PASS. | Prova real obtida; inventário completo tabela→policy não refeito. |
| F-R5-04 | R5-EXE-01 | `ClaimDirectIntent`/`RecoverOrphanedIntent` já cercam `retry_until`/`continue_after_client_deadline` no working tree. | Nenhuma (já íntegro); confirmado via `go test ./... -race` sem falha. | Íntegro, sem nova prova de cenário isolado. |
| F-R5-05 | R5-EXE-02 | `claimIntentMode` já reescrito com lock explícito de `operation_plans`/`running_count`. | Nenhuma; confirmado por `-race` sem falha (não é prova do cenário exato de 2 publishers). | Íntegro pelo teste existente; cenário do achado não replicado isoladamente. |
| F-R5-06 | R5-EXE-03 | `stepSnapshot` já resolve rota/serviço por etapa com hash próprio. | Nenhuma. | Íntegro pelo build/testes existentes. |
| F-R5-07 | R5-EXE-04 | `product_store.go` já gera compensação com DAG reverso e deadline próprio (migration 0048). | Nenhuma. | Íntegro pelo build/testes existentes. |
| F-R5-08 | R5-DAD-01 | `UnmarshalJSON`/`json.Number` já presentes em `dispatch/contract.go`, `orbita/factconsumer.go`, `libra/settlement.go`. | Nenhuma. | Íntegro; não reproduzi manualmente o probe do inteiro `9007199254740993` ponta a ponta nesta sessão (coberto por testes existentes). |
| F-R5-09 | R5-DAD-02 | Sem mudança de restore no working tree. | Não trabalhado — falta ambiente de restore. | Aberto. |
| F-R5-10 | R5-DAD-03 | `CaptureEffective`/`SetWatermark` existiam como dead code (nenhum chamador). | `ApplyEvent` agora chama `captureEffectiveTx` com valor efetivo somado dos fatos REVENUE; libera diferença como `released_amount`. `SetWatermark` continua sem produtor. | Parcialmente fechado (captura efetiva ligada; watermark ainda pendente). |
| F-R5-11 | R5-OPE-01 | `acquireCapacity` já recusa domínio sem policy (`ErrCapacityPolicy`). | Nenhuma. | Parcialmente íntegro; reciclagem de permits expirados não revisada. |
| F-R5-12 | R5-OPE-02 | `promotion-gate.sh` já exige manifesto em `prd`. | Executado o script real: BLOCK confirmado sem manifesto. | Verificado com execução real. |
| F-R5-13 | R5-OPE-03 | Sem mudança. | Não trabalhado — falta kind/k8s. | Aberto. |
| F-R5-14 | R5-UX-01 | `CatalogPage.tsx`/`admin.ts` já reescritos (busca server-side, validação de contrato). | Build TS/Vite confirmado limpo. | Íntegro pelo build; sem E2E de browser. |
| F-R5-15 | R5-QUA-01 | Sem mudança nos CSVs de inventário. | OpenSpec strict confirmado 27/27; CSVs não regenerados. | Parcial. |
| F-R5-16 | R5-DAD-04 | `Quarantine`/`ReplayQuarantine`/`MarkQuarantineReplayed` gravavam payload mas sem endpoint HTTP. | Endpoints `GET/POST /admin/v1/finance/quarantine[...]` adicionados; replay reprocessa via `ProcessEnvelope`. | Fechado no nível de wiring; sem teste HTTP dedicado. |
| F-R5-17 | R5-EXE-05 | Sem mudança (bootstrap `EnsureTopology` já centralizado nos 4 `cmd/*`). | Nenhuma. | Íntegro pelo build/testes existentes; prova de "consumidor tardio"/"assinatura removida" não isolada nesta sessão. |

## Regressões
Nenhuma regressão introduzida: `go build ./...`, `go vet ./...` e
`go test ./... -race` (245 testes, 0 falhas, 0 skips) permanecem limpos após
todas as edições desta sessão, com a suíte inteira rodando contra
Postgres/Redis/LocalStack reais (não mocks).
