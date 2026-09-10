# Comandos e alcance
Snapshot b9d0f90ce02aa0c27cad546745153d160ff5867f.
- Go 1.24.13: go test -race -json ./... em hub/; exit 0, 37 pass/18 skip.
- Frontend: node node_modules/typescript/bin/tsc -b && node node_modules/vite/bin/vite.js build; exit 0.
  Dependências reutilizadas do checkout anterior após igualdade exata de package-lock.json.
  Não equivale a npm ci limpo. Vite 5.4.21; 38 módulos transformados.
- OpenSpec 1.12.0: validate --all --strict --json --no-interactive; 17/17.
  Variáveis OPENSPEC_TELEMETRY=0, DO_NOT_TRACK=1, OPENSPEC_NO_UPDATE_CHECK=1.
- Provas separadas: go test -overlay /caminho/overlay.json ./internal/atlas -run TestR4 -v;
  quatro falhas esperadas de contrato, controles positivos passam.
- Prova token: mesmo overlay, ./internal/providerauth -run TestR4 -v; falha esperada.
- git status --short: vazio ao conferir checkout após verificações.

## Reproduzir overlay sem alterar a aplicação
Criar JSON com Replace mapeando o caminho absoluto virtual
hub/internal/atlas/r4_contract_probe_test.go para contract_probe_test.go deste diretório;
e hub/internal/providerauth/r4_token_probe_test.go para token_probe_test.go.
Executar no diretório hub do SHA auditado. O overlay é instrumento de auditoria, não implementação.

Docker/psql não encontrados no ambiente desta auditoria. DB/Compose/kind/browser/carga/HA não executados.
Não houve runtime local do Hub iniciado; portanto não foi necessário parar ecossistema de usuário.
