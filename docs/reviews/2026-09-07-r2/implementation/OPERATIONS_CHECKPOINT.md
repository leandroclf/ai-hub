# Checkpoint de operação R2

Estado em 2026-09-08, HEAD inicial a39d394b0d87185ed4cc3861c12ec45f2c302d9e, diff ainda não commitado.

Stack original `hub-local` preservada. Ensaio independente `ai-hub-r2` definido em `hub/deploy/r2/compose.yaml`: PostgreSQL porta 15432, LocalStack porta 14566, volumes postgres-data/aws-data próprios. Dados exclusivamente sintéticos. Migrations usam checksum e advisory lock em `scripts/migrate.sh`; leitura de migrations novas deve ser coordenada com agentes de domínio.

Em implementação: Compose com UI/Keycloak/telemetria; httpserver com hook AuthMiddleware, drenagem HTTP e histogramas por rota; laboratório kind e gates de qualificação. Nenhuma capacidade R2-OPE integralmente qualificada neste checkpoint. HTTP draining não implica drenagem de workers: domínios devem cancelar aquisição de leases em seu contexto.

Próximas ações: finalizar Compose e realm PKCE, aplicar migrations estáveis, subir serviços integrados, testar persistência/recriação e observabilidade; gerar/aplicar kind com controllers ou registrar falha real. Registrar comandos/resultados em hub/evidence/r2/operations. Não usar down -v; não tocar containers hub-local. Perfis cloud continuam dependentes de P-01/P-08/P-10/P-11.

Atualização: Keycloak 26.7.3, Prometheus 3.1.0, Grafana 11.4.0, Loki 3.3.2, Alloy 1.5.1 e Tempo 2.6.1 ativos. Password+OTP com AMR real e workload client_credentials testados (`identity-claims.json`, `workload-claims.json`). kind 0.27.0/Kubernetes 1.32.2 criado com três nós e kubeconfig próprio `/tmp/ai-hub-r2-tools/kubeconfig`; metrics-server 0.7.2 e KEDA 2.17.2 instalados e pods Running. Runtime kind ainda não aplicado. Provas unitárias `go test -race ./internal/platform/httpserver` passaram para middleware/probes, cardinalidade, drenagem e fila limitada de telemetria.
