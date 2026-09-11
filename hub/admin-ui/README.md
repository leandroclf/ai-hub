# AI Hub R4 — Portal administrativo

O `admin-ui` é o portal web administrativo do AI Hub. A aplicação é uma SPA
em Vite, React e TypeScript e é servida pelo Nginx da stack R2/R4. O portal
consome as APIs reais de Atlas, Órbita, Cometa, Pulsar e Libra por meio do
proxy same-origin configurado em `hub/deploy/r2/nginx.conf`.

## Execução local

O laboratório oficial usa um único projeto Compose:

```bash
docker compose ls
docker ps -a
docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml up -d
```

Com a stack ativa, o portal fica disponível em
`http://localhost:13000`. Para desenvolvimento isolado:

```bash
npm ci
npm run dev
```

O build de produção é validado por `npm run build`, que executa a checagem
TypeScript e o build Vite. A imagem oficial do Compose usa `Dockerfile.ui` e
mantém o proxy reverso no mesmo origin da SPA.

## Autenticação e segurança

O acesso usa Authorization Code + PKCE no realm OIDC configurado pelas
variáveis `VITE_OIDC_ISSUER` e `VITE_OIDC_CLIENT_ID`. O fixture local usa
senha e OTP; o portal recebe o access token somente em memória, remove a
transação PKCE após o callback e renova a sessão até logout ou expiração.

As APIs enviam `Authorization: Bearer` apenas quando há sessão válida. O
backend continua sendo a autoridade para tenant, papel, MFA e escopo. A UI
não deve ser exposta fora de um ambiente protegido apenas porque o build
passou.

Segredos de provedor e webhook nunca são digitados nem exibidos em claro.
As telas trabalham somente com referências versionadas, como `secret_ref` e
`secret_version`.

## Jornadas disponíveis

O menu é filtrado pelos escopos devolvidos pelo backend e cobre:

- catálogo versionado de clientes, aplicações, serviços, produtos, ofertas,
  importações, provedores, contas externas, vínculos, perfis técnicos,
  políticas e contratos;
- protocolos, entregas e timeline de protocolos;
- reconciliação autorizada de protocolo e redelivery de entrega esgotada;
- destinos de webhook versionados, por tenant ou aplicação, com URL,
  política de tentativas e referência de segredo persistidas;
- apuração de SLA bilateral baseada no snapshot persistido da operação;
- consultas e comandos financeiros para contas, fatos, journal, fechamentos,
  ajustes, divergências e recibos de exportação.

As ações de efeito operacional dependem do escopo apropriado, justificativa e,
quando exigido pelo contrato, MFA e segregação de funções. Paginação, filtros
de tenant e `If-Match`/idempotência são tratados pelas APIs; o frontend não
substitui essas verificações.

Para serviços com `adapter_id: rest-json-v1`, o editor exibe o campo JSON
`adapter_contract`. Preencha somente `submit_path` e `status_path`, por
exemplo `{"submit_path":"/analise","status_path":"/consulta/{id}"}`.
O backend rejeita host absoluto, query, fragmento, traversal, placeholders
adicionais e contrato parcial; a URL base continua sendo propriedade da conta
externa homologada.

## Validação de frontend

Use os gates nesta ordem quando a demanda envolver o portal:

1. `npm run build` para compilação e tipos.
2. Playwright determinístico:

   ```bash
   PLAYWRIGHT_MODULE=/home/leandro/IdeaProjects/lfsolucoes/ai-hub/hub/evidence/screenshots/node_modules/playwright-core/index.js \
     R2_COMPOSE_PROJECT=ai_hub_r3qual \
     node ../deploy/r2/tests/browser-smoke.mjs
   ```

   O smoke cobre OIDC/OTP, criação e readback durável, navegação autenticada,
   SLA, publicação/readback de destino, ausência de token persistido,
   logout e viewport de 390 px. A execução cria apenas dados sintéticos no
   laboratório; não use contas ou endpoints reais.

3. Browser Harness, quando a demanda exigir exploração visual, acessibilidade
   ou diagnóstico assistido. Ele é opcional, exploratório e não substitui o
   Playwright:

   ```bash
   ../deploy/r2/tests/browser-harness/run.sh
   ```

   Instalação, sessão, classificação `PASS-EXPLORATORY`/
   `BLOCKED-ENVIRONMENT` e regras para não registrar cookies, tokens ou OTP
   estão documentadas em
   [`hub/deploy/r2/tests/browser-harness/README.md`](../deploy/r2/tests/browser-harness/README.md).

O resultado do smoke é salvo em
`hub/evidence/r2/execution/browser-smoke.json`; revisar o status de cada
check, e não apenas o código de saída do processo, antes de promover a
evidência.

## Limites conhecidos

O portal não é uma aprovação de produção por si só. A qualificação integral
do AI Hub ainda exige cobertura dos requisitos e cenários herdados, testes de
provedores reais, matriz completa de negativas por papel, finanças e entregas
ponta a ponta, além dos gates de kind/HA, restore e observabilidade. Consulte
`docs/reviews/2026-09-09-r4/implementation/` para a situação atual e os
limites explicitamente mantidos como `OPEN`.
