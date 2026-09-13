# Gates e evidências — R6
Auditoria em 13/09/2026, SHA a540b40007fe6b8ed523e17afe00e96ff8f8ad50.

## Execuções desta revisão
| Comando/prova | Resultado | Limite |
|---|---|---|
| go test -race -json ./... | Exit 0; 139 testes PASS / 70 SKIP | Contagem de eventos Test; não inclui eventos de pacote. Dependências reais ausentes aqui. |
| go vet ./... | Exit 0 | Análise estática; não demonstra concorrência SQL. |
| npm run build | Exit 0 | Dependências reutilizadas após igualdade de package-lock; sem browser. |
| Probe de precisão do fato | PASS | Mantém 9007199254740993 no roundtrip do consumidor; não qualifica toda cadeia. |
| Probe de produto/rota | FAIL esperado | BuildProductPlan aceitou route B / account A / binding B/A / command A. |
| promotion-gate com manifesto não confiável | ALLOW reproduzido | Somente script local, nenhuma implantação; defeito registrado em R6-OPE-01. |
| OpenSpec baseline strict | 27/27 aprovados | Validação estrutural, não implementação. |

PostgreSQL, S3/LocalStack, Redis, OIDC, Compose/kind, browser, carga, HA e restore **não foram executados nesta auditoria**. Relatos históricos não foram negados, mas seus logs não puderam ser reconstituídos a partir do scroll citado no repositório. Não misturar resultados deste ambiente com os 245/246 testes relatados historicamente.

## Reproduzir probes
Criar overlay JSON com Replace apontando nomes virtuais de teste em hub/internal/orbita/ para evidence/product_route_probe_test.go e evidence/precision_probe_test.go. Executar em hub/: `go test -overlay <overlay-absoluto.json> ./internal/orbita -run 'TestR6|TestR5' -v`. A falha do caso de produto é a reprodução do problema, não um teste da suíte original.

O manifesto evidence/untrusted-manifest.json foi utilizado com `R2_PROMOTION_ENV=prd`, `R2_QUALIFICATION_PROFILE=unverified`, `R2_ENVIRONMENT_ISOLATION_PROOF=PASS`, `R2_PROMOTION_APPROVALS=P-01,P-08,P-10` e `R2_QUALIFICATION_EVIDENCE_MANIFEST=<caminho-absoluto>` ao executar `bash hub/deploy/r2/tests/promotion-gate.sh`. Não usar esse manifesto como evidência real; é entrada adversarial sintética para testar o gate.

## Critérios de encerramento
- G0: inventário e instruções lidos, SHA/diff preservados, dependências/toolchains fixadas.
- G1: todos os cenários novos e herdados obrigatórios possuem cobertura identificada; sem exclusão silenciosa.
- G2: zero P0 sem prova específica; correção e prova integradas no caminho real.
- G3: testes de DB/broker/objetos/OIDC/browser necessários executados; ausência de dependência faz gate obrigatório falhar.
- G4: imagem candidata, migração populada, rollback/forward-fix, restore e perfis de escala qualificados conforme o escopo.
- G5: manifesto validado por identidade/cobertura/integridade e caminho ALLOW legítimo; entradas adulteradas bloqueadas.
- G6: relatório coerente com logs e matriz. Aprovação técnica local e aprovação produtiva são estados separados.

## Formato de evidência
scenario_id; requirement_id; resultado enum; SHA; hash do diff; digest da spec/cenário; imagem/toolchain; ambiente/fixture; comando; esperado; observado; log e digest; início/fim; restrições. Sanitizar segredos e dados reais. Hash comprova bytes, não veracidade de resultado: proveniência e oráculo continuam necessários.

Não inserir UUID de protocolo/tenant individual em labels de alta cardinalidade. Use logs/traces correlacionados e agregações controladas para SLO.
