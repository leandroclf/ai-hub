# Verificação do snapshot
a4a876a9f8e875db882f7ca45cf7dece24d57aee. Data da sessão: 2026-09-08.

| Verificação | Resultado |
|---|---|
| go test ./... | exit 0 |
| go test -race -json ./... | exit 0; 35 pass; 18 skip |
| npm ci --ignore-scripts | exit 0 |
| npm run build | exit 0 |
| Prova de inteiro exato | FAIL esperado: 9007199254740993 → 9007199254740992 |
| Prova de enum | FAIL esperado: INVALID aceito |
| Integração Docker/DB/browser/kind/HA | NÃO EXECUTADO nesta revisão |


## Testes pulados
- ai-hub/hub/internal/atlas:TestCatalogPostgresConcurrencyPublicationAndIsolation
- ai-hub/hub/internal/atlas:TestCatalogPostgresPaginationAndStaging
- ai-hub/hub/internal/cometa:TestPostgresCapacityAggregateReplicasCells
- ai-hub/hub/internal/cometa:TestPostgresCapacityAdaptiveAndPending
- ai-hub/hub/internal/cometa:TestPostgresCapacityRollingRateIsolation
- ai-hub/hub/internal/cometa:TestPostgresSubmissionAndObservationCustody
- ai-hub/hub/internal/cometa:TestPostgresPendingCustodyAtomic
- ai-hub/hub/internal/cometa:TestPostgresOperationResourceScope
- ai-hub/hub/internal/cometa:TestPostgresPollingClaimsFenceAndAbsoluteDeadline
- ai-hub/hub/internal/cometa:TestPostgresPollingCallbackConflictRetainsBoth
- ai-hub/hub/internal/cometa:TestPostgresPollingAuthenticatedHTTP
- ai-hub/hub/internal/libra:TestFinancePostgresScenarios
- ai-hub/hub/internal/objectstore:TestPostgresS3MultipartFileRef
- ai-hub/hub/internal/orbita:TestAdmissionPostgresAtomicIdempotency
- ai-hub/hub/internal/orbita:TestAutoPostgresBoundedWait
- ai-hub/hub/internal/orbita:TestOperationFactPostgresCustody
- ai-hub/hub/internal/providerauth:TestAWSVaultLocalStackVersion
- ai-hub/hub/internal/pulsar:TestPostgresVersionedWebhookCustody
## Reprodução das provas
O arquivo evidence/precision_probe_test.go é um instrumento de auditoria; não faz parte da implementação entregue.
Em checkout do SHA revisado, criar overlay JSON com Replace apontando
hub/internal/atlas/r3_precision_probe_test.go para o caminho absoluto desse instrumento.
Executar em hub/: go test -overlay /caminho/overlay.json ./internal/atlas -run TestR3 -v.
As duas falhas demonstram defeitos atuais. Na implementação, convertê-las em regressões positivas de contrato.
Não atribuir as falhas ao go test original: o teste novo foi executado separadamente.
