# Browser Harness — validação exploratória do console

Esta pasta integra o [browser-harness](https://github.com/browser-use/browser-harness)
como uma camada opcional de validação exploratória e diagnóstico do frontend do AI
Hub R4.

O harness controla um navegador real por CDP e é útil quando a demanda exige
interação visual, inspeção de acessibilidade, reprodução de uma jornada autenticada
ou descoberta de uma regressão que ainda não possui um helper determinístico.
Ele não substitui o `browser-smoke.mjs`: o Playwright continua sendo o gate de
aceitação automatizado, repetível e bloqueador de CI.

## Quando usar

Use este harness quando a demanda envolver:

- exploração de uma jornada administrativa que acabou de mudar;
- diagnóstico de navegação, foco, teclado, viewport ou comportamento visual;
- confirmação de que uma sessão já autenticada consegue concluir uma jornada;
- criação de um helper reutilizável para um fluxo recorrente do console;
- coleta de screenshot ou gravação local explicitamente solicitada.

Não use para:

- substituir testes de contrato, unidade, integração ou Playwright;
- automatizar produção, contas reais, CAPTCHA ou ações irreversíveis;
- armazenar senha, OTP, access token, refresh token ou cookies no repositório;
- concluir um gate apenas porque um agente conseguiu “encontrar” um elemento;
- executar em paralelo no mesmo navegador local.

## Pré-requisitos

1. Compose oficial do AI Hub ativo, sem subir um segundo ecossistema:

   ```bash
   docker compose ls
   docker ps -a
   ```

2. Console disponível em `http://localhost:13000`.
3. Chrome/Chromium com CDP habilitado.
4. `browser-harness` instalado opcionalmente fora do repositório:

   ```bash
   uv tool install --python 3.12 --upgrade --force browser-harness
   browser-harness --doctor
   ```

   A instalação oficial e o procedimento de conexão estão no
   [guia upstream](https://github.com/browser-use/browser-harness/blob/main/install.md).

5. Uma sessão de teste autenticada no navegador. Senha, MFA e consentimento
   devem ser fornecidos interativamente pelo operador; o cenário não os captura
   nem os escreve em arquivo.

## Execução rápida

Com uma sessão de teste já autenticada no console:

```bash
hub/deploy/r2/tests/browser-harness/run.sh
```

O runner usa o daemon local padrão, mantém uma única aba e executa o cenário
`scenarios/admin-console.py`. O cenário percorre as 16 rotas administrativas
da fixture, aplica efetivamente um viewport CDP de 390×844, confirma que cada
tela contém conteúdo útil e verifica ausência de overflow horizontal. Ele não
cria, publica, suspende ou exclui recursos. Fixe `browser-harness 0.1.13` fora
do repositório quando a equipe precisar reproduzir o ambiente; nesta máquina
o executável está instalado, mas o daemon fica `BLOCKED-ENVIRONMENT` quando o
Chrome não expõe `DevToolsActivePort` ou uma sessão `BU_CDP_WS`.

Para uma execução contra outro endereço:

```bash
R2_ADMIN_URL=http://localhost:13000 browser-harness \
  < hub/deploy/r2/tests/browser-harness/scenarios/admin-console.py
```

## Preparação e reconciliação da fixture local

Quando uma imagem do laboratório for reconstruída, use o bootstrap oficial
antes dos gates de frontend e carga:

```bash
R2_COMPOSE_PROJECT=ai_hub_r3qual bash hub/deploy/r2/scripts/bootstrap.sh
```

O script constrói as imagens antes de descobrir os IPs das fixtures e faz a
subida final sem `--build`. Isso mantém as regras CIDR de egress do Cometa
alinhadas aos containers efetivamente em execução. Não altere manualmente
essas regras em um ambiente com provedor real.

Uma falha de transporte pode deixar uma concessão da fixture em
`pending_external`. Antes de repetir a carga, a reconciliação local pode ser
executada somente com confirmação explícita:

```bash
R2_LOCAL_RECONCILIATION_CONFIRM=I_UNDERSTAND_LOCAL_FIXTURE \
  bash hub/deploy/r2/tests/reconcile-local-pending.sh
```

O script consulta apenas domínios `r4-*`, pergunta ao oráculo sintético
`/__qualification/protocols/<protocol_id>` e fecha a concessão somente diante
de `404` (ausência comprovada de efeito). Uma resposta `200` protege a
obrigação e interrompe a execução; respostas de erro também interrompem. Esse
oráculo é exclusivo do laboratório local e nunca deve ser usado para concluir
uma reconciliação de provedor comercial. O log gerado em
`hub/evidence/r2/execution/capacity-reconciliation-latest.log` deve acompanhar
a rodada.

## Evidência e classificação

O resultado textual deve ser anexado à rodada em
`hub/evidence/r2/execution/` com data UTC, SHA do repositório, URL, perfil de
ambiente e indicação de sessão de teste. Screenshots e gravações podem conter
dados sensíveis; gravações ficam desabilitadas por padrão e só devem ser
habilitadas após consentimento explícito.

Um resultado do harness deve ser classificado como:

- `PASS-EXPLORATORY`: jornada observada, sem substituir o gate automatizado;
- `FAIL-EXPLORATORY`: comportamento reproduzido e evidência anexada;
- `BLOCKED-ENVIRONMENT`: daemon, navegador, OIDC ou ambiente indisponível;
- `NOT-APPLICABLE`: a demanda não exige interação real de navegador.

Nunca converter automaticamente `PASS-EXPLORATORY` em aprovação de requisito.
Para aprovação, atualizar também o cenário determinístico correspondente,
executar o `browser-smoke.mjs` e registrar o vínculo na matriz de rastreabilidade.

## Segurança operacional

- usar somente realm, tenant e usuário de fixture;
- não habilitar gravação em telas com dados reais;
- não usar Browser Use Cloud para esta qualificação local sem autorização e
  controle de custo;
- não imprimir URLs de sessão, cookies, headers Authorization ou payloads
  sensíveis;
- fechar abas criadas ao terminar, salvo quando a inspeção manual exigir mantê-las;
- interromper diante de uma ação irreversível ou de uma escolha de conta ambígua.

## Relação com os demais gates

| Camada | Objetivo | Bloqueia CI? |
|---|---|---:|
| Build TypeScript/Vite | compilação do frontend | Sim |
| Playwright `browser-smoke.mjs` | aceitação determinística integrada | Sim quando dependências disponíveis |
| Browser Harness | exploração, acessibilidade e diagnóstico assistido | Não por padrão |
| Revisão humana | decisão sobre risco visual/operacional | Conforme o requisito |

O harness está classificado como Alpha no projeto upstream. Fixe a versão da
ferramenta no ambiente de trabalho quando a equipe precisar reproduzir uma
execução; não adicione a ferramenta como dependência de runtime do AI Hub.
