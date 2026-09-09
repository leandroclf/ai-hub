# Relatório final da execução R3 — implementação parcial

## Identificação

- HEAD de entrada: `a4a876a9f8e875db882f7ca45cf7dece24d57aee`.
- HEAD de saída: o mesmo SHA; não houve commit, push, merge ou implantação remota.
- Branch: `codex/r3-implementacao-integral`.
- Alterações pré-existentes preservadas: remoção local de `IMPLEMENTATION_AUDIT.md` e pacote R3 recebido.

## O que foi implementado

- `TransformJSON` deixou de desserializar números em `float64`; inteiros grandes são preservados sem perda.
- Validação declarativa adicionada para `enum`, objetos aninhados, arrays, tipos e `additionalProperties:false`, sem execução de código, rede ou segredo.
- Resolução de ofertas passou a paginar o portfólio, removendo o teto artificial de 100 ofertas antes do filtro efetivo.
- `API_KEY` foi conectado ao modelo de conta, validação, projeção, migração, cliente, executor e polling, com header limitado a `X-API-Key` ou `Authorization`; o valor permanece no cofre.
- Bearer tokens deixaram de ser escritos/lidos no Redis; o cache L1 é vinculado à conta/binding/tenant/ambiente/versão e a renovação é serializada por chave.
- Diagnóstico administrativo de protocolos exige identidade nominal, MFA, papel `hub_protocol_reader` e autorização explícita para escopo cross-tenant.
- Regressões unitárias adicionadas para precisão, enum e schema aninhado.
- Artefatos de rastreabilidade R3 criados: plano, matriz de 189 requisitos, matriz de 280 cenários, achados, decisões, checkpoint e índice de evidências.

## Validações executadas

| comando | resultado |
|---|---|
| `cd hub && go test ./...` | PASS |
| `cd hub && go test -race ./...` | PASS |
| `cd hub && go vet ./...` | PASS |
| `cd hub/admin-ui && npm run build` | PASS |
| `git diff --check` | PASS |
| testes PostgreSQL/S3/LocalStack/OIDC/Compose/browser | PASS nas subprovas registradas em `EVIDENCE_INDEX.md`; browser autenticado ainda parcial |
| `docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml config --quiet` | PASS |
| `npx --yes @fission-ai/openspec@latest validate --all --strict` | PASS: 17 changes, 0 falhas |

## Atualização após reconstrução do laboratório

Após o reinício do Docker Desktop, as imagens dos serviços Go e do admin-ui foram reconstruídas a partir do working tree atual, com as bases pinned disponíveis localmente. O bootstrap reaplicou as migrações de forma idempotente, incluindo `core/0032_callback_inbox.sql`, sem remover volumes.

Passaram contra a instância `ai_hub_r3qual` reconstruída: admissão concorrente e custódia de fatos da Órbita; todos os testes PostgreSQL do Cometa; publicação/paginação do Atlas; custódia versionada do Pulsar; multipart PostgreSQL+LocalStack; todos os cenários financeiros R2-FIN-01 a R2-FIN-06; e resolução SigV4 de segredo sintético no LocalStack. O Kong foi limitado a um worker no Compose local, recriado sem o socket residual pós-reboot, e está saudável.

O OpenSpec estrito está validado, mas isso valida a estrutura das 17 mudanças; não converte automaticamente as tarefas ainda desmarcadas em implementação comprovada.

## Gotes executados nesta retomada

O bootstrap foi corrigido para o projeto Compose oficial `ai_hub_r3qual`, com reconciliação idempotente das fixtures Keycloak existentes e criação idempotente da credencial sintética do LocalStack. O fluxo Chromium autenticado passou com MFA, salvamento, recarga, novo login e leitura durável; o runner aceita explicitamente a instalação isolada de `playwright-core`.

Kind foi executado de fato com três nós e cinco workloads `1/1`, métricas de nós, HPA CPU real e KEDA pronto. Prometheus, Loki, Tempo, Alloy e Grafana foram executados; Prometheus observou os serviços, Alloy enviou entradas Docker ao Loki sem erro e Tempo retornou traces reais.

Após a reconstrução, a suíte integrada repetida passou para Órbita, Cometa, Atlas, Pulsar, objectstore/LocalStack, Libra e provider-auth. Houve reinicialização controlada da Órbita, com readiness 200 nas cinco APIs.

## Digests principais

Os digests reproduzíveis dos arquivos alterados nesta execução estão em `EVIDENCE_INDEX.md`; eles devem ser regenerados após qualquer alteração adicional e antes do commit.

## Gates que impedem conclusão integral

- P0 F-R3-01…F-R3-05, F-R3-13, F-R3-17 e F-R3-23 permanecem abertos ou sem prova integrada.
- F-R3-06 tem implementação e regressão unitária positiva, mas ainda não possui publicação/integração com oráculo externo.
- Recuperação SUBMITTING/UNKNOWN, callback autenticado com custódia durável, fencing antes de I/O e polling+callback têm subprovas integradas; ainda falta o ensaio completo de falha/reinício com oráculo externo independente.
- Kind, stack de observabilidade real e jornada administrativa autenticada de salvar/recarregar foram qualificados nesta retomada.
- Restore reconciliado com efeitos externos/financeiro, RLS com credencial de runtime não proprietária, executor DAG/composição além do planejador, política/planos administrativos avançados e fluxos restantes do console continuam sem prova integral.
- Os 42 achados históricos R2 permanecem em revalidação; evidência histórica não foi promovida para este SHA.
- D-01…D-07 e T-R2-01 permanecem decisões externas não aprovadas.
- Existem testes condicionais que fazem `t.Skip` quando dependências reais não estão configuradas; eles não foram contados como PASS.
- O console recebeu retry transitório com a mesma `Idempotency-Key`, e o realm passou a declarar o User Profile administrativo necessário para claims de tenant/aplicação/célula. A reconciliação reproduzível das credenciais OTP e a jornada completa de navegador passaram nesta retomada.

Conclusão: a entrega permanece **implementação parcial qualificada em laboratório**. Kind, observabilidade e navegador autenticado agora têm evidência real; continuam impedindo a declaração integral os gates de restore reconciliado, RLS/continuidade, DAG/composição e console avançado, reavaliação individual dos 42 achados, cenários completos de falha/restore e as decisões externas D-01…D-07/T-R2-01. Não houve push, merge, implantação remota ou remoção de volumes.

## Atualização técnica posterior — 2026-09-09

Foram implementados e verificados os seguintes incrementos, sem promover artificialmente seus gates:

- Executor DAG bounded em `hub/internal/atlas/executor.go`, com entradas derivadas das dependências, paralelismo limitado, execução parcial e compensação reversa; testes unitários com race passaram. A integração ao DAG persistido e aos fluxos de composição ainda não foi comprovada.
- Provider-sim com deduplicação por protocolo e oráculo independente de efeitos. O ensaio de duas submissões iguais produziu um único efeito; a fixture é em memória e não substitui uma prova de recuperação do provedor.
- RLS efetivo para a credencial sintética não proprietária `hub_runtime`, com `FORCE ROW LEVEL SECURITY`, migrações de compatibilidade, flags de papel não privilegiado e prova negativa de acesso cruzado em control/core/finance. Os serviços ainda não foram convertidos integralmente para essa credencial.
- Harness de restore em bancos e bucket isolados, com comparação de contagens, objetos e observação de não-replay durante a janela cercada. A reconciliação populada entre efeitos externos e financeiro ainda falta.
- A validação individual dos 42 achados históricos foi documentada em `REAVALIACAO-42-INDIVIDUAL.md`; nenhum foi encerrado sem evidência própria.
- O migrador agora interrompe divergência de checksum em vez de ignorá-la.
- O retry transitório agora grava a primeira falha e uma janela absoluta em `command_intents`; a prova PostgreSQL real confirmou que takeover não renova o TTL e que a intenção termina em `EXPIRED` após o horizonte.
- Observações finais divergentes de callback/polling agora são classificadas como `*_CONFLICT` em receipts duráveis; o resultado vencedor e o único outbox terminal permanecem inalterados, comprovado no PostgreSQL real.

Assim, o estado correto continua sendo **implementação parcial qualificada**, não implementação integral. Permanecem abertos: restore populado reconciliado, adoção RLS no runtime dos serviços, conexão do executor ao DAG real, planos comerciais e jornadas administrativas restantes, ensaios completos de falha/reinício com oráculos independentes, fechamento individual dos achados quando houver prova, e as decisões D-01…D-07/T-R2-01. As alterações desta atualização ainda estão no working tree; não houve push, merge, implantação remota ou remoção de volumes.

## Decisões adotadas para viabilizar a execução

Os defaults técnicos de D-01…D-07 e T-R2-01 foram registrados em `docs/reviews/2026-09-08-r3/05-DECISOES-E-MIGRACAO.md` e estão valendo para local/homologação. Eles não fecham os gates de produção: os valores reais de demanda, SLA, preço, retenção, plataforma/orçamento, capacidades dos provedores, catálogo oficial e a semântica normativa de T-R2-01 ainda exigem ratificação e evidência próprias.

## Atualização técnica — worker de retenção

O runtime de Órbita passou a executar o expurgo de objetos por lote durável quando `RETENTION_TENANTS` contém uma lista explícita. O plano de candidatos é persistido antes do efeito externo; falhas individuais deixam `PURGING` para retry e pins continuam impedindo expurgo de obrigações. A configuração vazia mantém o worker desabilitado por segurança. A regressão de normalização de tenants e a compilação dos pacotes passaram; a qualificação integrada de retenção com objetos expirados e falha/retry do S3 ainda é um gate aberto.

O provider-sim agora mantém seu oráculo de efeitos em volume exclusivo, com gravação atômica. A submissão sintética repetida após reinício preservou o mesmo identificador externo e um único efeito. Essa prova melhora a qualificação de não-reexecução, mas não substitui o restore completo reconciliado com a autoridade financeira.

A prova de retenção PostgreSQL+LocalStack também cobre execução efetiva: o pin exclui a obrigação do plano, o `UNPIN` gera ação auditável, o lote cria tombstone e remove a versão, e uma leitura posterior é negada. Falhas de S3 continuam retomáveis pelo estado `PURGING`.

O harness de restore foi fortalecido com digestes determinísticos de linhas e correção de nomes de bucket S3. A execução isolada `autonomous_digest2` passou e conservou o oráculo externo sem replay; como não havia objetos no bucket fonte nessa janela, a prova ainda não encerra o restore populado de objetos, efeitos externos e financeiro.

Foi adicionada uma injeção controlada de queda entre efeito externo e resposta HTTP. O provider-sim persiste a chave/efeito antes de fechar a conexão e o Cometa encaminha `force_drop_after_effect`; a implementação está coberta por teste unitário, mas o cenário integrado com protocolo real, recuperação e oráculo financeiro permanece aberto.

No ensaio HTTP real após reconstrução da imagem atual, a primeira chamada perdeu a resposta e a repetição retornou `SUCCEEDED` com o mesmo identificador externo; o oráculo permaneceu em um único efeito. A prova ainda é do provider-sim; a integração completa com protocolo do Hub e oráculo financeiro permanece aberta.
