# Prompt — Implementação integral R3 do AI Hub

Você é o agente responsável por concluir a evolução do repositório AI Hub. Trabalhe no código real, com rigor de engenharia,
até concluir todo o trabalho executável da baseline v4, da R2 e desta R3. Não pare após planejar, criar stubs ou implementar helpers isolados.
Não declare “tudo concluído” se houver testes obrigatórios pulados, P0 aberto, jornada desconectada ou decisão normativa crítica não resolvida.

## 1. Localização e autoridade
O pacote foi extraído na raiz do repositório:
- docs/reviews/2026-09-08-r3/
- openspec/changes/r3-01-custodia-e-integracao-externa/
- openspec/changes/r3-02-portfolio-e-contratos-efetivos/
- openspec/changes/r3-03-capacidade-credenciais-e-latencia/
- openspec/changes/r3-04-console-e-autorizacao/
- openspec/changes/r3-05-financeiro-e-entregas/
- openspec/changes/r3-06-dados-plataforma-e-continuidade/
- openspec/changes/r3-07-qualificacao-integral/

O snapshot auditado foi a4a876a9f8e875db882f7ca45cf7dece24d57aee. Antes de alterar, obtenha HEAD/status/diff atuais.
Se HEAD mudou, revalide cada achado no código atual e preserve melhorias existentes. Não retorne o repositório ao snapshot antigo.
Leia AGENTS.md de cada diretório aplicável, openspec/config.yaml se existir, e todas as instruções de docs/openspec-docs/.
Leia a baseline hub-interoperabilidade-v4, os nove changes r2-*, relatórios de implementação R2, o ADR T-R2-01 e todo o pacote R3.
Siga as instruções de maior prioridade; conflito concreto deve ser registrado, nunca ocultado.

Os 25 requisitos R3 complementam 164 herdados: total de 189 obrigações na matriz de trabalho.
Os 75 cenários novos NÃO substituem cenários v4/R2. Não reescreva requisitos para que a implementação existente pareça correta.
Não arquive changes ou crie baseline consolidada afirmando conformidade antes de executar as provas.
As pastas históricas e seus relatórios devem permanecer reconhecíveis; registre nova execução em implementation/ da R3.

## 2. Modo de execução
Crie branch apropriada, preserve alterações do usuário e registre plano operacional com dependências e checkpoints.
Implemente fatias verticais curtas: contrato → persistência → domínio → transporte → UI → teste real → evidência.
Mantenha o escopo completo e avance sem pedir confirmação para cada decisão técnica reversível já coberta pelas especificações.
Não implemente tudo de uma vez sem verificar. Um teste passando não autoriza encerrar tarefas de outra camada.
Se contexto/tempo interromper execução, grave checkpoint verificável e retome da próxima tarefa; não apresente interrupção como conclusão.
Não faça push, merge ou implantação remota sem autorização aplicável. Nunca apague volumes/dados de usuário para limpar testes.
Não use credenciais comerciais reais em logs, commits, fixtures ou prompts.

## 3. Primeira entrega interna
Crie docs/reviews/2026-09-08-r3/implementation/ com:
- EXECUTION_PLAN.md: sequência, dependências, HEAD e estado inicial.
- REQUIREMENTS_STATUS.csv: 189 IDs, requisito, cenários, tarefas, status e evidência.
- SCENARIO_RESULTS.csv: TODOS os cenários v4/R2/R3, comando, ambiente, pass/fail/skip e evidência.
- FINDINGS_STATUS.md: 25 achados R3 e reavaliação dos 42 históricos, com prova por encerramento.
- DECISIONS_AND_BLOCKERS.md: decisão/recurso ausente, responsável, impacto e trabalho independente.
- CHECKPOINT.md: último passo validado, processos/ambientes, próximo passo e pendências.
- EVIDENCE_INDEX.md: SHA, digests, logs saneados e hashes.

Status de auditoria não é resultado de execução. Inicialize como pendente e atualize somente com evidência.
Não copie o relatório R2 como prova nova.

## 4. Prioridades obrigatórias
1. Preparar harness de dependências reais, identidade OIDC limpa/reproduzível e imagens do SHA. Gate integrado sem fixture deve falhar, não pular.
2. Fechar fronteira consumidor/admin/aplicação e custódia de mensagens/callback/topologia/quarentena.
3. Corrigir perda numérica/schema e ligar adapters/autenticação real, preservando resposta do provedor.
4. Recuperar SUBMITTING/UNKNOWN, aplicar fencing antes de I/O e idempotência externa homologada.
5. Separar relógios e habilitar polling+callback; tratar o contraexemplo de confirmação durável sem relaxar o SLA.
6. Ligar controlador adaptativo a SUBMIT/STATUS/FETCH; corrigir pools e cache de tokens por binding.
7. Executar DAG/agregação, snapshot efetivo e representação customizada com igualdade GET/webhook.
8. Integrar reserva/franquia pré-efeito, incidência de compra/venda, reconciliação/fechamento e destinos congelados.
9. Concluir jornadas administrativas reais, incluindo formulários, fuso, seleção de entrega, SLA e comandos financeiros.
10. Ligar objetos/retention, concluir Compose/kind/ambientes, escala/telemetria/restore e qualificar a união completa.

Consulte 05-DECISOES-E-MIGRACAO.md para dependências detalhadas. Faça testes desde o primeiro fluxo.
Não trocar linguagem/broker/framework para evitar corrigir a integração existente.

## 5. Regras que não podem ser violadas
- Cliente não consulta protocolo de outro tenant/aplicação fora de autorização explícita.
- Administrador global é identidade nominal autorizada, com MFA/auditoria; não conta compartilhada de desenvolvedores.
- UUIDv7 é recuperável após aceite tanto SYNC quanto ASYNC. Pedido rejeitado antes da custódia não pode fingir protocolo persistido.
- SYNC de provedor síncrono retorna final na mesma chamada e não depende de round-trip em fila.
- Resultado final consultado vem do Hub; GET e corpo webhook reutilizam mesma representação congelada.
- Antes de efeito, persistir intenção/tentativa e autorização; antes de ACK, conservar recibo/obrigação.
- Duplicata não repete efeito/lançamento. UNKNOWN exige evidência, não retry cego.
- TTL não renova a cada tentativa; polling saudável e reconciliação não se confundem com retry de indisponibilidade.
- Resposta tardia não reabre protocolo expirado; evidência externa continua conservada para reconciliação financeira.
- Redis não é autoridade e não recebe bearer tokens/segredos. Fallback não usa memória volátil para prometer custódia.
- Saldo/franquia estritos são autorizados antes do efeito; valores exatos e captura conciliada.
- Produtos têm estado por etapa, paralelismo limitado e compensação explícita.
- Controle de capacidade é global no domínio real do provedor e justo por tenant.
- Migração/rollback não apaga ledger, outbox, recibo ou obrigação pendente.

## 6. Engenharia e testes
Execute o OpenSpec CLI qualificado pelo repositório com validação estrita, sem atualizações arbitrárias de ferramenta.
Use os testes de auditoria em evidence/precision_probe_test.go como ponto de partida para regressões de contrato.
Para os testes de banco, utilize PostgreSQL real com configuração durável e papéis runtime mínimos.
Use provider-sim/webhook-sink externos como oráculos de contagem/bytes; mocks internos não demonstram efeito único.
Execute go test, race, vet, build frontend, contratos, integração e browser; trate skip obrigatório como falha.
Em browser, clique, salve, recarregue e confirme efeito no backend/runtime. Não considerar menu visível como funcionalidade.
Teste erro 401/403/409/422/503, resposta inválida, timeout após commit, perda de lease, retry e acesso cruzado.
Qualifique Compose completo e kind completo com dependências no cluster. Renderização YAML não é execução.
Teste Prometheus/Loki/Grafana/Tempo com sinais reais e alerta injetado.
Execute carga/isolamento/falha/restore conforme 04-QUALIFICACAO-E-OPERACAO.md. Não invente volumes comerciais ou SLO aprovado.

## 7. Decisões e bloqueios
Não presuma que D-01…D-07/P-01…P-11 foram aprovadas por existir código.
Use fixtures sintéticas explícitas para avançar sem contrato comercial e registre o gate externo pendente.
T-R2-01 permanece aberto enquanto não houver prova válida para a semântica normativa. Repetir o contraexemplo é obrigatório.
Se a solução exigir mudança de requisito, registre opções, consequência e decisão necessária; não altere silenciosamente.
Esgote o trabalho independente antes de parar por bloqueio. Nomeie exatamente o que falta, sem declaração genérica de impossibilidade.

## 8. Conclusão verificável
Antes de declarar concluído:
- Todas as tarefas aplicáveis v4/R2/R3 têm implementação conectada e evidência atual.
- Todos os cenários obrigatórios passam; zero skip obrigatório.
- Nenhum P0 está aberto e riscos P1/decisões impeditivas estão resolvidos ou explicitamente impedem a conclusão integral.
- Frontend e API têm contratos coerentes e jornadas demonstradas no navegador.
- Compose/kind, dados, observabilidade e recuperação foram realmente executados.
- Git diff foi revisado, sem segredos, artefatos locais de IDE ou alterações alheias.
- Migração e rollback foram ensaiados sem perder obrigações.
- OpenSpec strict e matriz de rastreabilidade estão consistentes.

Gere FINAL_REPORT.md com HEAD/digests, o que mudou, testes/contagens, evidências, riscos e pendências reais.
Se houver bloqueio, diga “implementação parcial” e identifique o gate; jamais “100%” por contagem de arquivos ou caixas marcadas.
A resposta final deve permitir ao mantenedor reproduzir e avaliar a entrega sem depender da narrativa do agente.
