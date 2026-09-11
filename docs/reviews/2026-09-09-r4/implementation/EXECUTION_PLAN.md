# Plano de execução R4 — continuação real

Estado: implementação parcial, sem promoção para homologação.

1. Revalidar proveniência, instruções, status Git, ferramentas e ecossistema local.
2. Fechar contratos JSON com regressões adversariais e dialeto explícito.
3. Preservar a migração histórica 0002 e adicionar reconciliação restrita da variante R3 conhecida.
4. Separar callback externo da autenticação JWT de workload e exigir prova de origem da conta.
5. Evoluir inbox/recuperador, resultado unificado, ofertas seletivas e console por fatias verticais.
6. Executar gates Go, banco, broker, objetos, OIDC, Compose/kind e navegador quando as dependências estiverem disponíveis.
7. Atualizar este diretório com proveniência, evidência e bloqueios; não marcar PASS por inspeção histórica.

## Política permanente de validação frontend

O Playwright permanece o gate determinístico de aceitação. O Browser Harness,
integrado em `hub/deploy/r2/tests/browser-harness/`, deve ser usado quando a
demanda exigir exploração com navegador real, diagnóstico visual/acessibilidade
ou validação de jornada autenticada já preparada pelo operador. Seu resultado
é `PASS-EXPLORATORY`, `FAIL-EXPLORATORY`, `BLOCKED-ENVIRONMENT` ou
`NOT-APPLICABLE`; nunca encerra requisito sozinho. Senhas, MFA, tokens,
cookies, gravações e Browser Use Cloud permanecem sujeitos às regras de
segurança e consentimento descritas no README do harness.

Na execução de 2026-09-11, os gates locais de código, OpenSpec, navegador,
carga autorizada, produto HTTP, RLS, entrega, financeiro e restore passaram;
kind passou após reinício dos workloads, incluindo ensaio de duas réplicas. O
Browser Harness passou como exploratório com `BU_CDP_URL` explícito; a
descoberta automática de CDP headless continua indisponível. As etapas 5–7
seguem abertas nos pontos que exigem implementação ou cobertura integral;
consulte `EXECUTION-2026-09-10-SEQUENCE.md` para a matriz de evidências e
pendências sem promoção indevida.
