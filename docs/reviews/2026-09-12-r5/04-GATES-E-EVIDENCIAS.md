# Gates e evidências
Snapshot f87ce33034ae29c9431b1910dcc6a633b545e330. Auditoria de 12/09/2026.

## Executado aqui
- Go 1.24.13: go test -race -json ./... em hub/: exit 0; 137 testes pass, 69 skip. JSONL anexado; eventos de pacote não contam como teste.
- go vet ./...: exit 0.
- TypeScript/Vite build: exit 0. Dependências locais reutilizadas após igualdade de package-lock; não é npm ci limpo.
- OpenSpec 1.12.0 baseline: validate --all --strict --json --no-interactive, 21 changes. União R5 validada separadamente.
- Probes R4 JSON + token L1 adaptado: PASS. O instrumento antigo de token exigiu atualização por alteração do tipo privado cachedToken; falha de compilação do instrumento não foi classificada como defeito do Hub.
- Offer 403→cache: FAIL reproduzido. Inteiro grande no consumidor de fatos: FAIL reproduzido. Logs e testes por overlay anexados.
- promotion-gate.sh com profile não verificado/strings de aprovação/sem manifesto: ALLOW reproduzido; nenhuma implantação ocorreu.
- Helper antigo ExecuteDAG ainda inicia etapa após cancelamento no probe anterior. É auxiliar não ligado ao novo runtime persistido; não usar esse resultado para afirmar execução persistida ausente.

## Não executado aqui
PostgreSQL/S3/LocalStack/OIDC/Docker/Compose/kind/browser/carga/HA/restore. Ferramentas de runtime não estão disponíveis neste executor. Isso não nega subprovas históricas no repositório; delimita a revisão.
Dockerfile atual Go 1.26 e Nginx atualizado não foram construídos. Go 1.24.13 não é prova da imagem candidata.

## Reprodução dos probes
Criar overlay Go Replace com caminhos absolutos virtuais hub/internal/atlasclient/r5_offer_probe_test.go e hub/internal/orbita/r5_precision_probe_test.go apontando aos arquivos entregues. Executar em hub/ com go test -overlay <arquivo> ./internal/atlasclient ./internal/orbita -run TestR5 -v.
Token: mapear hub/internal/providerauth/r4_token_probe_test.go ao token_probe_test.go adaptado; JSON: hub/internal/atlas/r4_contract_probe_test.go ao contract_probe_test.go. Executar -run 'TestR4Contract|TestR4Previous|TestR4Warm'.
Probe de promoção: R2_PROMOTION_ENV=prd R2_QUALIFICATION_PROFILE=unverified R2_ENVIRONMENT_ISOLATION_PROOF=PASS R2_PROMOTION_APPROVALS=P-01,P-08,P-10 bash hub/deploy/r2/tests/promotion-gate.sh. Esse script avalia gate; não publica aplicação.

## Qualificação obrigatória da implementação
Dependência ausente deve falhar gate integrado obrigatório, nunca virar t.Skip contado como PASS. Unitários seguem separados.
Executar contratos/DB/broker/objetos/OIDC/browser, runtime Compose/kind, migração/restore e cenários de concorrência/expiração por risco.
Oráculos: efeito externo por chave, bytes/hash/versão de objetos e webhook, estado após reinício, saldo/journal exatos, negações com controle positivo e métricas sob carga.
Gate de promoção deve validar evidência assinada/íntegra/compatível com o artefato no pipeline autorizado, não variáveis autoafirmadas.
Evidência por cenário: SHA + hash do diff local + imagem/base/toolchain + comando/ambiente + esperado/observado + resultado + log/digest. Mudança material invalida resultado afetado.
