# Atlas — Console administrativo (admin-ui)

Interface administrativa web do Atlas (plano de controle do Hub de
Interoperabilidade), implementando a decisão arquitetural **ARQ-04**: "Atlas
interface — TypeScript + React, aplicação web administrativa com formulários
tipados, visualização de contratos e timeline".

Stack: **Vite + React + TypeScript**. Não usa Next.js/SSR — o `design.md` do
projeto já registra explicitamente que "SSR/Next.js não é dependência
necessária" para esta interface.

## Pré-requisitos

- Node.js 18+ e npm.
- O backend Atlas real precisa estar no ar, escutando em `http://localhost:8081`
  (rota HTTP definida em `hub/cmd/atlas/main.go` e `hub/internal/atlas/handlers.go`).
  Suba-o via a stack docker compose do repositório:

  ```bash
  cd ../deploy
  docker compose up -d atlas
  ```

  (ou suba a stack inteira, se preferir: `docker compose up -d`).

## Rodando em desenvolvimento

```bash
npm install
npm run dev
```

O Vite abre em `http://localhost:5173` (padrão). O `vite.config.ts` configura
um proxy de `/api` para `http://localhost:8081`, então **todas** as chamadas
do frontend usam o prefixo relativo `/api/v1/...` (ver
`src/api/atlasClient.ts`) — nunca uma URL absoluta. Isso é necessário porque o
backend Go do Atlas **não configura CORS** (e este scaffold não altera o
backend para isso); em dev, o proxy do Vite contorna a questão fazendo a
chamada parecer same-origin do ponto de vista do navegador.

Esse mesmo padrão de prefixo relativo também permite, no futuro, servir esta
UI atrás do gateway Kong já existente (`hub/deploy/kong/kong.yml`), bastando
reescrever `/api` → `http://atlas:8081` na borda em vez de no Vite.

## Build de produção

```bash
npm run build
```

Roda `tsc -b` (checagem de tipos) seguido de `vite build`, gerando os
artefatos estáticos em `dist/`. Servir esses artefatos em produção exigiria
decidir como o prefixo `/api` será roteado até o Atlas (proxy reverso próprio,
ou uma rota Kong análoga à existente para `/admin/atlas`) — isso não está
implementado neste scaffold.

## O que esta interface cobre

Quatro telas (abas simples dentro de `App.tsx`, sem `react-router` — mantido
deliberadamente simples), cada uma cobrindo exatamente as rotas HTTP reais do
Atlas (nenhum endpoint foi inventado):

1. **Catálogo de serviços** — publicar um serviço novo (`POST /v1/services`) e
   consultar por código+versão (`GET /v1/services/{code}/{version}`).
2. **Contas de provedor** — cadastrar (`POST /v1/provider-accounts`) e
   consultar por ID (`GET /v1/provider-accounts/{id}`).
3. **Vínculos de credencial** — cadastrar vínculo (`POST /v1/credential-bindings`)
   e resolver credencial por tenant+conta (`GET /v1/credentials/resolve`).
4. **Contratos** — cadastrar/atualizar (`POST /v1/contracts`) e consultar por
   tenant (`GET /v1/contracts/{tenantId}`).

O Atlas não expõe endpoints de "listar todos" (só busca por chave exata), então
as tabelas de cada tela mostram um histórico local da sessão do navegador
(itens publicados/consultados), não uma fonte de verdade persistida.

### Segredos de credencial (CFG-05)

A tela de vínculo de credencial **nunca** tem um campo para digitar ou exibir
o segredo em si (ex.: `client_secret`, API key em claro). Apenas o campo
`secret_ref` (referência/caminho no cofre) é cadastrado, refletindo CFG-05:
"o segredo é escrito no cofre por fluxo restrito, sem leitura posterior em
claro pela interface". Isso está documentado também em comentários no código
(`src/api/atlasClient.ts` e `src/pages/CredentialBindingsPage.tsx`).

## Fora do escopo

- **Autenticação real / OIDC (SEG-01)**: não implementada. A UI hoje não exige
  login e não envia nenhum header de identidade/tenant nas chamadas — é um
  placeholder de desenvolvimento local, sinalizado com um aviso fixo no topo
  da aplicação (`App.tsx`). Não exponha este scaffold fora de um ambiente de
  desenvolvimento confiável antes de resolver SEG-01.
- **CORS no backend**: propositalmente não alterado — o Atlas continua sem
  CORS; o frontend depende do proxy do Vite (dev) ou de um gateway comum
  (produção) para funcionar.
- **Visualização de timeline** (mencionada na ARQ-04) e outras visualizações
  mais ricas de contrato: não implementadas neste scaffold mínimo — o Atlas
  hoje não expõe dados de timeline/eventos, apenas o snapshot atual de cada
  recurso.
- Rota Kong dedicada para esta UI (`/admin/atlas-ui`): não adicionada neste
  scaffold porque não há serviço/container desta UI no
  `hub/deploy/docker-compose.yml` para o Kong apontar; o essencial pedido
  (o frontend em si) está completo e documentado aqui.
