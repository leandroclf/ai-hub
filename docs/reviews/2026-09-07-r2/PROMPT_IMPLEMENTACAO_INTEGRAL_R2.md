# Prompt de execução integral — AI Hub / Constelação — R2

> Salve este arquivo em `docs/reviews/2026-09-07-r2/PROMPT_IMPLEMENTACAO_INTEGRAL_R2.md`. O conteúdo a partir de “Missão e autorização” é a instrução para o agente implementador.

## Missão e autorização

Atue como responsável técnico pela implementação, integração e qualificação da rodada R2 do AI Hub / Constelação. Execute as nove changes OpenSpec desta rodada até concluir o escopo tecnicamente executável e comprovar seu funcionamento. Trabalhe no código, frontend, contratos, migrations, infraestrutura declarativa, testes e documentação necessários. Esta solicitação autoriza a implementação e os ensaios locais isolados; não é apenas um pedido de análise ou planejamento.

Prossiga autonomamente entre etapas: não encerre a execução após produzir plano, primeira change, build, testes unitários ou relatório parcial. Corrija falhas encontradas e continue nas demais frentes. Não solicite confirmação para decisões rotineiras, alterações reversíveis, instalação de ferramentas de desenvolvimento compatíveis ou execução de testes locais isolados que já estejam no escopo.

Não confunda persistência com repetição indefinida: se uma ação continuar falhando, investigue a causa, altere a abordagem e registre o diagnóstico. Bloqueio real de uma capacidade não deve paralisar trabalho independente. Se o ambiente impuser limite de execução ou contexto, conserve um checkpoint suficiente para retomar sem refazer trabalho nem declarar conclusão.

“Concluído” significa comportamento implementado e comprovado no ambiente e perfil declarados. Não prometa ausência absoluta de falhas ou perda de dados fora do modelo de falhas testado. Nunca enfraqueça requisitos, suprima testes ou invente evidências para obter um resultado verde.

## 1. Fontes de verdade e preparação

Antes de editar, leia:

1. `AGENTS.md`, instruções aplicáveis às subpastas e arquivos de orientação do agente existentes no repositório.
2. `docs/openspec-docs/`, especialmente o prompt metodológico, template de change e checklist avaliador.
3. `LEIA_PRIMEIRO_R2.md` e `docs/reviews/2026-09-07-r2/00-leia-primeiro.md`.
4. O relatório `01-relatorio-de-revisao.md`, a matriz `02-rastreabilidade-98.csv` e os documentos compartilhados de arquitetura, console, contratos/dados/estados, operação, qualidade, pendências e método nessa mesma pasta.
5. A baseline `openspec/changes/hub-interoperabilidade-v4/`, os documentos de engenharia em `docs/` e todas as nove changes `openspec/changes/r2-*`.
6. Em cada change, leia integralmente `explore.md`, `proposal.md`, `design.md`, `specs/`, `tasks.md`, `risk-matrix.md`, `TRACEABILITY.md` e `evaluator-checklist.md` antes da implementação correspondente.

A revisão R2 analisou o commit `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. Registre o HEAD atual, a branch e alterações locais existentes. Compare com esse SHA e revalide achados afetados: um gap já corrigido exige prova e atualização de rastreabilidade, não nova implementação redundante. Não trate a auditoria histórica como observação automática do HEAD atual.

Preserve trabalho do usuário. Use branch ou worktree isolado quando necessário; não execute reset destrutivo, limpeza indiscriminada ou rebase de trabalho alheio. Verifique ferramentas, recursos e dependências disponíveis. Prepare fixtures sintéticas e identifique claramente ambientes de ensaio.

A baseline v4 continua sendo o alvo normativo; R2 complementa seus critérios e fecha lacunas. Não arquive a v4 inteira nem descarte requisitos que não foram repetidos literalmente nas changes R2. Resolva contradições explícitas com rastreabilidade. Se a solução exigir mudar uma obrigação de negócio, mantenha esse ponto bloqueado até decisão autorizada e avance no restante.

## 2. Planejamento executável e sequência

Crie um plano de execução rastreado por change, requisito, tarefa, dependência e prova. Mantenha as tarefas OpenSpec como referência de conclusão. A rodada contém 98 requisitos de baseline, 42 achados, 66 requisitos R2, 205 cenários e 186 tarefas inicialmente abertas; os totais ajudam na conferência, mas não substituem leitura e cobertura semântica.

Organize o trabalho nesta sequência de dependências, sem transformar toda a rodada em uma fila artificialmente serial:

| Frente | Change | Condição de execução |
|---|---|---|
| Qualidade desde o início | `r2-09-qualificacao-e-rastreabilidade` | Preparar fixtures, oráculos, execução reproduzível e coleta de evidências imediatamente; fechar a qualificação ao final |
| Identidade e isolamento | `r2-01-identidade-e-isolamento` | Fundação para APIs, console e integração segura |
| Custódia e execução | `r2-02-execucao-duravel-e-resultados` | Corrigir aceite, intenção, posse, resposta real, idempotência e deadline; integrar com identidade |
| Catálogo e contratos | `r2-04-catalogo-produtos-e-contratos` | Fixar contratos versionados, ofertas, perfis e DAG antes de concluir consumidores |
| Integrações reais | `r2-03-integracoes-credenciais-e-pressao` | Integrar execução durável, contratos, credenciais, polling, callbacks e controle de pressão |
| Console administrativo | `r2-05-console-administrativo` | Iniciar sessão/navegação/listas cedo; concluir cada jornada com APIs reais do domínio correspondente |
| Dados e arquivos | `r2-07-objetos-dados-e-retencao` | Integrar custódia, autorização, referências de objetos e recuperação |
| Financeiro | `r2-06-financeiro-auditavel` | Usar contratos congelados e fatos duráveis de execução; qualificar concorrência e reconciliação |
| Operação completa | `r2-08-operacao-elastica-e-observavel` | Preparar laboratório cedo; qualificar runtime integrado, escala, manutenção e recuperação antes do fechamento |

Priorize os 13 P0 antes de ampliar exposição dos fluxos afetados. Reduza retrabalho acordando interfaces produtor/consumidor antes de avançar sobre elas. Se o ambiente permitir trabalho delegado, use-o apenas para frentes independentes e delimitadas, com responsabilidade de integração e revisão final mantida por você.

## 3. Regras de implementação que não podem ser relaxadas

### Identidade e isolamento

- Derive tenant e permissões de identidade autenticada. Cabeçalhos, corpo, UUID, filtros ou cursores fornecidos pelo chamador não concedem identidade ou acesso.
- Valide autorização de recurso no serviço proprietário, inclusive em chamadas internas. Autenticação no gateway não substitui essa autorização.
- Clientes não consultam protocolos, objetos, entregas, contratos ou informações financeiras de outros clientes.
- Implemente acesso administrativo global individual e auditado, com MFA e perfil de leitura entre tenants separado das permissões de escrita. Não crie uma conta compartilhada para desenvolvedores.
- Use credenciais de workload e proteção de egress/SSRF; não confunda acesso à rede com autorização.

### Aceite, execução e resposta

- Todo aceite conhecido, síncrono ou assíncrono, possui UUIDv7 recuperável, identidade, snapshot e intenção duráveis. Rejeições anteriores ao aceite não devem inventar protocolo.
- SYNC elegível recebe o final contratual na mesma requisição e não depende obrigatoriamente do broker; não converter silenciosamente SYNC em 202. ASYNC e AUTO seguem os contratos distintos da spec.
- Desconexão do cliente não elimina a obrigação de reconciliar efeitos já enviados. Timeout de transporte não prova que o provedor deixou de executar.
- Persistir tentativa e reivindicar posse válida antes do efeito externo. Idempotência, lease e fencing precisam funcionar sob concorrência, restart e takeover; `ON CONFLICT DO NOTHING` isolado não basta.
- Preservar atomicidade local de efeito, inbox/outbox e obrigação. Confirmar mensagem somente após custódia comprovada; mensagens inválidas têm quarentena durável e diagnóstico.
- Conservar a resposta real e validada do provedor. Nunca usar eco do input, corpo vazio ou HTTP 200 sem semântica validada como sucesso.
- Materializar a representação final completa conforme o perfil do cliente. POST final, GET final e corpo do webhook usam a mesma representação imutável; não reconstruir partes que possam divergir após upgrade.
- GET consulta o ecossistema do Hub, com autorização e tratamento de réplica atrasada; não provoca nova consulta externa para recuperar um final já custodiado.

### Prazos, polling e entregas

- O TTL de reprocessamento é configurável em segundos, tem referência temporal explícita e não reinicia a cada tentativa. Orçamentos de tentativa, fila, provedor e cliente devem ser compatíveis.
- Deadline expirado fecha o protocolo conforme contrato. Resultado tardio não reabre sucesso nem modifica a resposta final entregue; conservar evidência e eventual obrigação financeira.
- Resolver tecnicamente T-R2-01: confirmação durável estritamente anterior ao prazo precisa de prova, inclusive com commit lento e relógio incerto. Comparação com `now()` no início da transação ou timer periódico não basta.
- Polling deve ser configurável, autenticado, persistido e coordenado entre réplicas. Polling e callback concorrem sobre a mesma identidade de operação, sem duplicar efeito ou final.
- ACK de callback confirma custódia do recibo, não aceitação do resultado como sucesso dentro do SLA. Proteger também callback recebido antes da correlação estar pronta.
- Entregas ao cliente possuem destino e chave versionados, tentativas reivindicadas, recibos e política de retry. Um receptor lento não bloqueia os demais. Sucesso HTTP não prova processamento de negócio pelo receptor.

### Provedores, pressão e contratos

- Resolver e utilizar efetivamente a credencial compartilhada do Hub ou dedicada ao cliente conforme binding autorizado. Não trocar credencial dedicada por padrão compartilhado em caso de falha.
- Segredos são resolvidos em cofre; referências não são senhas. OAuth, API keys, Basic e mTLS só são declarados suportados quando implementados e homologados de fato.
- Aplicar amortecimento adaptativo, limites de concorrência, feedback, fairness e orçamento por domínio de capacidade. Reduzir pressão sob degradação e recuperar gradualmente com sinais de estabilidade.
- Resolver T-R2-04: o consumo agregado entre réplicas/células respeita a capacidade compartilhada. Acrescentar pods não pode multiplicar indevidamente a quota do provedor. Redis não será a única autoridade desse controle.
- Selecionar provedor por elegibilidade e equivalência homologada antes de otimizar prioridade/custo/latência. Incerteza após envio não autoriza failover cego ou hedging.
- Serviços e produtos possuem versões, contratos de entrada/saída por cliente, ofertas e snapshots. Agregação/composição usa DAG, dependências, paralelismo limitado, parcialidade e compensações especificadas.
- Importação de coleção produz staging, diff e homologação. Cadastro de endpoint não prova adapter executável; importação não apaga configurações locais indiscriminadamente.

### Persistência, financeiro e disponibilidade

- Preserve as autoridades Atlas, Órbita, Cometa, Pulsar e Libra e justifique ajustes de tecnologia por evidência. Separe papéis e recursos de runtime quando necessário, sem criar um domínio por adaptador.
- Redis é cache dispensável: falha ou reinício não pode torná-lo dependência obrigatória do caminho elegível. Testar operação contínua e inicialização sem Redis.
- Reduza dependências síncronas de controle com projeções válidas e caches limitados, mantendo segurança, validade e revogação. Não use estado expirado como autorização.
- Falha de persistência não autoriza aceite fictício. Fallback deve preservar autoridade e custódia; onde isso não for possível, recusar a admissão afetada explicitamente e preservar capacidades independentes.
- Arquivos grandes usam transferência direta/streaming e FileRef autorizada, validada e imutável; não carregar corpos sem limite em memória. URLs temporárias não devem tornar mutável o corpo final persistido.
- Compra do provedor e venda ao cliente têm regras, partes e marcos próprios, congelados no contrato da operação. Tipo de credencial não determina automaticamente quem paga.
- Usar precisão monetária exata, unidade econômica idempotente, journal balanceado, reservas, capturas, holds, estornos e reconciliação. UNKNOWN não libera saldo por conveniência.
- Retenção, restauração e migração preservam obrigações, evidências e idempotência. Não reconstruir resultado histórico ausente a partir do input nem reexecutar operação externa como backfill.

## 4. Console administrativo: entregar jornadas completas

Siga `04-console-jornadas-e-api.md` e todos os requisitos R2-ADM. Implemente sessão, contexto de ambiente/tenant, navegação por URL, listas reais paginadas, detalhe, edição, validação, publicação, conflitos de concorrência e permissões por ação.

Cubra clientes/aplicações/ofertas; serviços e importação; produtos compostos; provedores e vínculos de credenciais; perfis técnicos e contratos legados; protocolos e timeline; entregas e reprocessamento autorizado; SLA bilateral; financeiro e conciliação.

Mostre estados de carregamento, vazio, erro, indisponibilidade, sucesso e conflito. Valide formulários com regras e dados do domínio, evitando exigir que o operador conheça IDs internos. Inclua acessibilidade, navegação por teclado e responsividade conforme os critérios da change.

Telas devem recuperar seu estado real após refresh e funcionar com múltiplos operadores. Não concluir jornada com dados estáticos, sucesso fictício ou mock. Se faltar API, implemente o contrato e backend correspondente antes de declarar a jornada pronta. Capturas de tela ajudam a revisão, mas não substituem teste integrado de ação, persistência e autorização.

## 5. Deploy, elasticidade e observabilidade

Entregue execução local reproduzível por Docker Compose com os componentes e dependências requeridos, incluindo frontend, identidade, persistência, mensageria, objetos e observabilidade. Documente ordem de inicialização, migrations, bootstrap idempotente, volumes e recuperação após recriação. Use isolamento de portas, nomes e volumes do ensaio.

Entregue laboratório kind e configurações coerentes para local, dev, hom, ppd e prd. kind é laboratório de Kubernetes, não prova de disponibilidade física multi-AZ. Não trate manifests de referência, placeholders ou comentários de KEDA como instalação funcional.

Inclua startup/readiness/liveness apropriadas por capacidade, requests/limits, shutdown e drenagem, distribuição de réplicas, políticas de interrupção, HPA/KEDA, expansão de nós/células conforme arquitetura e um controlador responsável por cada recurso. Trate indisponibilidade de métricas e saturação com limites seguros.

Autoscaling precisa considerar fila, concorrência, capacidade externa, conexões SQL, recursos disponíveis e orçamento. Crescimento dentro do envelope aprovado deve ocorrer sem chamados rotineiros; não prometa capacidade infinita ou eliminação de quotas externas.

Disponibilize Prometheus, Loki, Alloy, Grafana e instrumentação OpenTelemetry/Tempo conforme desenho, com dashboards, alertas e runbooks provisionados. Meça latência por trecho, SLA cliente–Hub e Hub–provedor, atraso de filas, custódia, retries, UNKNOWN, resultados tardios, entregas e pressão adaptativa. Controle cardinalidade e não exponha segredos ou payloads sensíveis em logs/métricas.

## 6. Método de execução e comprovação

Para cada requisito/fatia:

1. Revalidar o código e a regra, identificar produtores/consumidores e contratos afetados.
2. Definir oráculo verificável e casos de falha relevantes antes de implementar.
3. Implementar a menor fatia completa, com migrations aditivas e compatibilidade quando necessária.
4. Executar verificações de domínio, contrato e integração apropriadas.
5. Corrigir falhas e regressões; revisar riscos de concorrência, segurança e recuperação.
6. Integrar com demais componentes e jornada administrativa pertinente.
7. Registrar evidência e somente então concluir a tarefa correspondente.

Use os 205 cenários R2 e os cenários integrados IT do documento de qualidade. Reaproveite testes quando comprovarem a mesma obrigação, mas não substitua a prova de comportamento por correspondência de nome ou cobertura de linhas.

Inclua testes negativos entre tenants, credenciais dedicadas, resultado diferente do input, duplicatas, perda de ACK, queda entre persistência e publicação, crash após efeito externo, callback/polling/timer concorrentes, commit atravessando deadline, retry até TTL, saldo concorrente, destinos webhook compartilhando URL, indisponibilidade de Redis/broker, arquivos grandes e reentrega após upgrade.

Use PostgreSQL e mensageria de ensaio onde a garantia depende deles. Um mock pode apoiar teste unitário, mas não comprova transação, lock, lease, redelivery ou recuperação. Provedores simulados devem reproduzir falhas e retornar marcadores e identidade verificáveis; não devem mascarar ausência de integração real homologada.

Qualifique cada camada de escala: carga controlada, saturação de um tenant/serviço e efeito nos demais, capacidade adaptativa do provedor, comportamento do scaler, preservação de conexões e recuperação do backlog. Declare perfil, volume, duração, recursos e limite efetivamente ensaiados.

Exercite migrations em base limpa e existente, reinício/recriação, rollback compatível e restore cercado. Nunca restaure backup e libere automaticamente egress que possa repetir efeitos antigos.

Para cada evidência, registre: IDs de requisito/cenário/tarefa, SHA e eventual diff ainda não commitado, ambiente, versões, fixture, comando/procedimento, esperado, obtido, status, data e caminho do artefato sanitizado. Para teste manual de UI, descreva passos, dados, ação e confirmação no backend.

## 7. Decisões e bloqueios: continuar sem inventar aprovação

P-01 a P-11 e T-R2-01 a T-R2-05 não são presumidos resolvidos. Procure decisões e evidências já existentes; resolva questões técnicas no seu escopo com experimento e ADR. Valores de teste devem estar identificados como fixtures, não contratos aprovados.

Ausência de credencial de provedor, licença, definição comercial, política de dados ou acesso cloud impede apenas homologação/ativação dependente. Implemente o suporte configurável, teste com fixtures adequadas, conserve o bloqueio específico e avance em trabalho independente.

Não execute produção, alteração remota onerosa, concessão de privilégios reais, rotação de segredo real, importação destrutiva ou migração irreversível sem autorização específica aplicável. Não reduza proteção nem use credenciais de produção para contornar limitação local. Respeite mecanismos de aprovação do ambiente.

Cada bloqueio deve conter requisito afetado, evidência da limitação, tentativas realizadas, alternativa avaliada, impacto, responsável funcional e ação necessária para desbloquear. Não marque cenário bloqueado como aprovado. Só encerre por bloqueio quando não restar trabalho independente viável.

## 8. Critério de conclusão e entrega final

Feche esta rodada apenas quando todos os itens executáveis estiverem implementados, integrados e qualificados e os restantes, se houver, estiverem identificados como bloqueios reais. Diferencie explicitamente:

- **Implementado e comprovado:** comportamento e cenários pertinentes passaram com evidência no perfil declarado.
- **Implementado, não qualificado:** código existe, mas a prova exigida não foi executada ou depende de ambiente externo; tarefa de qualificação continua aberta.
- **Bloqueado:** dependência real impede avanço; indicar ação necessária.
- **Não implementado:** trabalho pendente; não apresentar como bloqueado sem causa demonstrada.

Antes de concluir:

1. Validar as nove changes pelo OpenSpec em modo estrito, com versão registrada.
2. Executar lint/análise estática/build e testes pertinentes, além dos gates de integração, segurança, UI, operação e recuperação aplicáveis.
3. Resolver erros introduzidos e regressões; não omitir falhas preexistentes relevantes ao escopo.
4. Conferir a rastreabilidade dos 98 requisitos anteriores e dos 66 R2, os 42 achados e os 205 cenários, mantendo a distinção entre implementação e prova.
5. Atualizar tasks, TRACEABILITY, avaliação, auditoria e instruções de execução/migração/rollback de acordo com a evidência real.
6. Conferir que o console usa backend real, que ambientes não dependem de placeholders não declarados e que não existem segredos, dados reais ou artefatos transitórios indevidamente versionados.
7. Preservar o relatório e a matriz originais da revisão como evidência do SHA anterior. Registrar evolução em artefatos novos, sem reescrever retrospectivamente os achados como se nunca tivessem existido.

Mantenha em `docs/reviews/2026-09-07-r2/implementation/`:

- `EXECUTION_PLAN.md`: plano e dependências atualizados.
- `CHECKPOINT.md`: estado retomável, próxima ação, comandos e bloqueios; sem credenciais.
- `REQUIREMENTS_STATUS.csv`: requisito, tarefa, situação e evidência, incluindo baseline e R2.
- `SCENARIO_RESULTS.csv`: cada cenário R2 com resultado e referência de prova.
- `FINDINGS_STATUS.md`: situação dos 42 achados e riscos residuais.
- `DECISIONS_AND_BLOCKERS.md`: decisões técnicas tomadas e pendências reais.
- `FINAL_REPORT.md`: mudanças, testes, limites, instruções de operação e próximos passos necessários.

Guarde evidências sanitizadas na estrutura `hub/evidence/` do projeto, referenciando-as pelos relatórios. Não versione caches, dependências instaladas, arquivos locais de IDE ou artefatos volumosos desnecessários. Se criar commits, siga a identidade e regras Git do repositório. Push/PR e implantação remota dependem da autorização e acesso aplicáveis à sessão; não são pré-requisitos para concluir a implementação local solicitada.

Na resposta final, entregue a situação concreta do projeto: o que funciona, em qual ambiente foi comprovado, quais testes passaram/falharam/não foram executados, como reproduzir, onde estão os artefatos e quais bloqueios restam. Não use “100% pronto”, “sem perda garantida” ou “produção validada” sem escopo e evidência correspondentes.

Comece agora pela leitura e revalidação do snapshot. Em seguida, implemente, integre, teste e corrija continuamente até atingir os critérios acima. Não pare apenas para oferecer continuar.
