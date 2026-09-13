# PROMPT — Implementação integral da rodada R6 do AI Hub

Você é o agente responsável pela implementação e qualificação da próxima rodada do AI Hub. Trabalhe no checkout local de https://github.com/leandroclf/ai-hub. A revisão R6 observou o SHA `a540b40007fe6b8ed523e17afe00e96ff8f8ad50`. O repositório já contém a arquitetura v4 e rodadas R2/R3/R4/R5. **Implemente os deltas R6 e conclua as pendências normativas herdadas com evidência real**, preservando correções existentes.

## 1. Objetivo e autonomia
Esta é uma instrução para executar, corrigir, integrar e testar, não para produzir apenas outro plano. Prossiga pelas etapas autorizadas sem pedir confirmações de rotina e sem parar após build, testes unitários ou checkpoint. Resolva setup local reversível, ferramentas, fixtures, configuração e erros de implementação. Mantenha atualizações curtas de progresso e continue até cumprir os critérios abaixo ou restar impedimento externo tecnicamente irredutível.

Não fabrique aprovação de decisão externa, contrato comercial, resultado de teste ou homologação. Não burle política de acesso, proteção de branch ou bloqueio de ferramenta. Não use prazo de sessão como justificativa para abandonar o próximo trabalho disponível; mantenha checkpoint que permita continuidade em caso de interrupção real.

## 2. Leia antes de alterar
1. AGENTS.md e instruções aplicáveis em subdiretórios.
2. docs/openspec-docs/ e convenções reais dos changes.
3. docs/reviews/2026-09-13-r6/00-LEIA-PRIMEIRO.md, relatório, plano, arquitetura/decisões, gates, reavaliação e jornadas.
4. Os seis openspec/changes/r6-* completos: explore, proposal, design, tasks, specs, traceability, riscos e checklist.
5. Inventários R6, implementação R5 e registros de decisões/achados das rodadas anteriores.

Fixe HEAD/branch/status/diff de entrada. Se HEAD diferir de `a540b40007fe6b8ed523e17afe00e96ff8f8ad50`, leia o delta e revalide o achado antes de corrigir. Preserve alterações locais preexistentes, inclusive remoções intencionais e os pacotes recebidos. Não faça reset/checkout destrutivo. Não altere checksums de migrações antigas para esconder divergência.

## 3. Escopo integral e rastreabilidade
Baseline revisada: 218 requisitos, 783 cenários, 27 changes. R6 adiciona 15 requisitos, 45 cenários e seis changes: total esperado 233 requisitos, 828 cenários, 33 changes, salvo evolução posterior documentada. Os changes R6 têm 72 tarefas; o plano inclui 25 verificações herdadas C-R6, total 97 itens de trabalho.

Extraia os inventários das specs vigentes e compare com os CSVs; não escreva contagens à mão. Cada cenário terá resultado próprio, mesmo NOT_RUN/BLOCKED. Não confunda cenário, teste, subteste e pacote. Resultados históricos só podem ser reutilizados com escopo e proveniência suficientes para o candidato. Os 42 achados R2, 25 R3 e 17 requisitos R5 continuam rastreáveis; não declare fechamento em lote.

## 4. Preserve avanços existentes
Não reintroduza a antiga perda float64 já corrigida no consumidor. Preserve HMAC antes de callback órfão, cache-first com validade e negações, claim de produto com lock do plano, DAG reverso, captura efetiva ligada a ApplyEvent, payload de quarentena, EnsureTopology no startup e manifesto obrigatório em prd. A R6 fecha fronteiras restantes desses mecanismos.

## 5. Prioridades obrigatórias de implementação
### R6-SEG-01 — RLS no caminho real e prova com dados existentes
O Hub SHALL aplicar escopo autenticado por transação e por item de trabalho com roles runtime sem bypass; tabelas sem tenant direto devem ter autorização por relação ou autoridade operacional específica. Testes de isolamento SHALL afirmar existência e acesso próprio, invisibilidade alheia e recusa de escrita cruzada antes de remover fixtures.

Mapear tabela→owner→role→policy→caminho. Separar migrador, requests e workers globais; estes precisam claims limitados e auditados, não tenant arbitrário fornecido pelo cliente. Adotar contexto LOCAL e provar limpeza em commit/rollback. A migração isolada não autoriza trocar o DSN e derrubar todos os workers. Qualificar migração com bases populadas e credenciais runtime efetivas.

Execute R6-SEG-01-S01, S02 e S03. Fonte e reprodução: r6-01-isolamento-e-custodia/explore.md.

### R6-SEG-02 — Callback aceito sobrevive à rotação e à indisponibilidade
O Hub SHALL preservar a atestação de autenticação obtida no ingresso e a disposição recuperável de cada callback aceito. Rotação e falha transitória não podem invalidar custódia legítima; esgotamento de retry deve manter obrigação em quarentena reprocessável, distinta de rejeição de contrato.

Persistir key-id/versão/conta/hash/instante de verificação e proteger sua integridade. Segredo em cofre, jamais no recibo. Reconciliar correlação sem exigir assinatura contra nova chave. Particionar locks/quotas por origem e célula; retenção de órfãos e quarentena com responsável, prazo, alertas e replay autorizado. Evitar descarte automático de dados aceitos.

Execute R6-SEG-02-S01, S02 e S03. Fonte e reprodução: r6-01-isolamento-e-custodia/explore.md.

### R6-EXE-01 — Snapshot de etapa coerente em todas as identidades
O Hub SHALL materializar cada etapa com rota, conta, binding, perfil, contrato de compra, prazo e incidência coerentes e congelados. Se a oferta não contém referências suficientes para resolver a etapa, a publicação ou admissão SHALL ser recusada antes de qualquer efeito.

Resolver grafo completo no Atlas durante publicação/admissão; não selecionar primeira rota preenchida ignorando elegibilidade. Incluir dependências versionadas no snapshot do produto; gerar comando a partir do filho completo, não de cópia parcial. Recalcular hash e verificar invariantes no executor. Venda do produto e custos das etapas usam chaves econômicas distintas. Preservar contrato final de consolidação.

Execute R6-EXE-01-S01, S02 e S03. Fonte e reprodução: r6-02-execucao-coerente-de-produtos/explore.md.

### R6-EXE-02 — Expiração bloqueia também recuperação DIRECT
O Hub SHALL barrar nova submissão com possibilidade de efeito depois do prazo aplicável em DIRECT, QUEUED e recuperação. A reconciliação de efeito possivelmente já realizado SHALL continuar pela operação de consulta autorizada, sem se confundir com reenvio.

Centralizar decisão persistida de elegibilidade imediatamente antes de I/O; distinguir SUBMIT de STATUS/consulta por chave. Consumir Expired em todo caminho e finalizar a intenção com disposição recuperável. Manter relógios cliente, provedor, tentativa e TTL desde primeira falha separados; não tratar polling saudável como indisponibilidade.

Execute R6-EXE-02-S01, S02 e S03. Fonte e reprodução: r6-02-execucao-coerente-de-produtos/explore.md.

### R6-EXE-03 — Compensação progride após final público
O Hub SHALL separar estado de atendimento do cliente e estado de obrigações compensatórias. Encerrar protocolo não pode impedir compensação já devida; compensações SHALL respeitar causalidade reversa, política própria e reconciliação sem ressuscitar o final público.

Criar elegibilidade de obrigação compensatória independente do terminal público e bloquear novas etapas produtivas após cancelamento. Configurar prazo/retry/UNKNOWN da compensação por versão. Cobrir ancestrais transitivos quando etapas intermediárias não tiverem compensação; falha de compensação requer disposição explícita e consulta administrativa.

Execute R6-EXE-03-S01, S02 e S03. Fonte e reprodução: r6-02-execucao-coerente-de-produtos/explore.md.

### R6-EXE-04 — Liberação de slot e intent na mesma transação
O Hub SHALL atualizar intent, lease da etapa e contador do plano atomicamente também na conclusão, falha e expiração. Takeover e fatos concorrentes SHALL preservar max_parallel, idempotência e recuperabilidade sem zerar contadores artificialmente.

Definir ordem única de locks entre claim, CompleteIntent e ApplyProductFact; incluir epoch em escritas e verificar RowsAffected. Reconciliador detecta drift por dados autoritativos, sem GREATEST mascarando perda de invariante. Testar falha no ponto entre estados com DB real e barreiras, não apenas go -race.

Execute R6-EXE-04-S01, S02 e S03. Fonte e reprodução: r6-02-execucao-coerente-de-produtos/explore.md.

### R6-FIN-01 — Replay financeiro distingue aplicação de nova quarentena
O Hub SHALL produzir disposição tipada para replay: aplicado, ainda em quarentena, conflito ou falha transitória. Uma obrigação não aplicada SHALL continuar visível e recuperável. Todo payload custodiado SHALL declarar formato/versão e ter vínculo auditável com bytes de origem, inclusive sem identidade de domínio válida.

Persistir outcome e tentativa/ator atomicamente ou com claim cercado. Não marcar original encerrado sem recibo da obrigação sucessora ou aplicação. Tratar legado sem payload e EconomicEvent versus Envelope por formato explícito. Invalid JSON/event_id ausente exige identidade de transporte/hash sem inventar evento de negócio. Escopo global da lista requer autorização global real, não somente query tenant de fachada.

Execute R6-FIN-01-S01, S02 e S03. Fonte e reprodução: r6-03-reconciliacao-financeira/explore.md.

### R6-FIN-02 — Completude e liquidação verificáveis no fluxo real
O Hub SHALL fechar período somente com prova de completude de todos os produtores e obrigações da coorte. Reserva, captura efetiva, liberação e ajustes SHALL conservar identidade e valores exatos sob reordenação, duplicação e resultados tardios.

Definir marcador durável por produtor/partição/coorte com tratamento de lacunas e avanço monotônico, publicado por outbox no próprio fluxo. Tempo de parede não é prova de entrega. Conciliar receita do produto e custos de etapas; testar valor efetivo acima da reserva por política aprovada, sem fabricá-la. Duplicata semanticamente igual com escala decimal diferente não deve gerar conflito falso. Exportação fechada é imutável; ajustes posteriores são novos lançamentos.

Execute R6-FIN-02-S01, S02 e S03. Fonte e reprodução: r6-03-reconciliacao-financeira/explore.md.

### R6-OPE-01 — Gate de promoção vinculado ao artefato e à cobertura
O Hub SHALL bloquear promoção quando qualquer evidência obrigatória não corresponder ao artefato candidato, ao cenário vigente e à execução autorizada. Status SHALL usar enum estrito; cobertura parcial e declarações autoatribuídas não qualificam promoção.

Pipeline gera manifesto com commit/diff, digest de imagem, spec/scenario/log, toolchain, execução e oráculos. Validador conhece inventário obrigatório e origem autorizada, verifica hash de bytes e assinatura/atestação confiável quando aplicável; trust root não vem do mesmo input não confiável. Testar também caminho ALLOW legítimo. Desenvolvimento local com fixtures não equivale a autorização de prd.

Execute R6-OPE-01-S01, S02 e S03. Fonte e reprodução: r6-04-qualificacao-operacional/explore.md.

### R6-OPE-02 — Restore populado com versões e retomada reconciliada
O Hub SHALL demonstrar restore não vazio de todas as autoridades e obrigações, incluindo versões de objetos referenciadas, seguido de retomada cercada e reconciliação de efeitos e valores. A ausência de dados ou a mera igualdade de contagens SHALL impedir aprovação de recuperação integral.

Inventário de schema e obrigações atualizado por migração. Fixture possui produtos em execução, UNKNOWN, callbacks, webhook pendente, saldo e versões/pins/tombstones. Restaurar em alvo isolado, mapear IDs de versão quando não preserváveis e verificar bytes/hash. Suspender novas admissões até reconciliação. Ensaiar RTO/RPO sob perfil acordado; restaurar sem privilégios requer reaplicar roles/policies e provar acesso correto.

Execute R6-OPE-02-S01, S02 e S03. Fonte e reprodução: r6-04-qualificacao-operacional/explore.md.

### R6-OPE-03 — Ambientes e elasticidade com orçamento de dependências
O Hub SHALL fornecer perfis reproduzíveis local/dev/hom/ppd/prd com dependências e limites explícitos, escalabilidade automática dentro do envelope qualificado e continuidade mensurável. Readiness e topologia de obrigações SHALL refletir capacidade real de admitir/processar com custódia, sem confundir laboratório efêmero com HA produtiva.

Docker Compose oficial único; kind serve qualificação, Kubernetes/EKS alvo remoto conforme decisão D-05. Gerar HPA/KEDA, recursos, PDB, startup/readiness/liveness, afinidade, identidades, redes, secrets e volumes/serviços duráveis por ambiente. Autoscaling de pods depende de nós, conexões, broker, storage e quotas cloud; onboarding automatizado valida capacidade antes de aceitar contrato. Topologia gerenciada por autoridade única ou reconciliação idempotente contínua com prova de falha de assinatura. Risco de perda regional exige política RPO/RTO aprovada.

Execute R6-OPE-03-S01, S02 e S03. Fonte e reprodução: r6-04-qualificacao-operacional/explore.md.

### R6-OPE-04 — Capacidade obrigatória e feedback recuperável
O Hub SHALL exigir política efetiva de capacidade para rotas ativas e conservar obrigações de feedback/settlement até resolução. Controle adaptativo SHALL respeitar teto seguro contratado e compartilhar capacidade de forma justa entre tenants, com custo estável em função do estado ativo.

Publicação recusa domínio ausente em rota que exige controle. Separar orçamento de transporte, SUBMIT, polling, pendências e reconciliação. Feedback de timeout/429/5xx reduz concorrência; janelas estáveis recuperam gradualmente até teto qualificado, nunca inferem capacidade infinita. Persistir settlement pendente e reconciliar por evidência, não por expiração de lease. Materializar contadores/índices do estado ativo e provar cardinalidade/custo.

Execute R6-OPE-04-S01, S02 e S03. Fonte e reprodução: r6-04-qualificacao-operacional/explore.md.

### R6-UX-01 — Seletores completos para produtos e rotas
O console SHALL permitir localizar e selecionar qualquer referência elegível por ID/versão, inclusive em etapas e rotas, sem carregar todo o catálogo nem ocultar a referência atualmente selecionada. Paginação e busca SHALL preservar tenant, filtros e estado de edição.

Componente comum de seleção remota com debounce, cancelamento, cursor e resolução pontual do valor selecionado. Passar consulta por editor e não refazer todos os lookups por tecla. Diferenciar vazio de erro e impedir resposta antiga de outro tenant. Contratos de elegibilidade vêm da API, não de filtros locais parciais.

Execute R6-UX-01-S01, S02 e S03. Fonte e reprodução: r6-05-console-operavel/explore.md.

### R6-UX-02 — Validação e intenção preservadas em toda mutação
O console SHALL validar contratos de sucesso e erro por operação e bloquear mutações enquanto qualquer editor apresenta entrada inválida. Uma intenção de mutação com resultado desconhecido SHALL conservar identidade, payload e versão até reconciliação, inclusive após nova tentativa ou recarga.

Propagar validade do editor ao formulário pai sem apagar texto inválido. Guards ou schemas gerados para DTOs de catálogo, finanças, protocolos e operações; erro de contrato não altera draft. Journal local da intenção sem tokens/segredos, escopado por usuário/tenant/ambiente, reconciliado com autoridade antes de gerar outra chave. Testes browser de timeout após commit, duplo clique, sessão expirada, ETag e permissões.

Execute R6-UX-02-S01, S02 e S03. Fonte e reprodução: r6-05-console-operavel/explore.md.

### R6-QUA-01 — Evidência completa e vinculada ao código revisado
A engenharia SHALL manter inventário integral extraído das specs e resultado individual com evidência por cenário. Contagens SHALL ser calculadas dos logs, com subtestes/pacotes diferenciados. Um resultado histórico sem vínculo verificável com artefato SHALL permanecer histórico e não virar PASS do candidato.

Preservar v4/R2/R3/R4/R5 e incorporar deltas R6, com semântica distinta para achado, requisito, cenário e teste. Registrar SHA/diff, toolchain, imagem, fixture, comando, esperado/observado, logs saneados e digests. Revalidar só o que mudança material afeta, mas completar cenários obrigatórios sem skip. CLI OpenSpec fixada, sem @latest em evidência reproduzível. Decisões externas não aprovadas permanecem pendentes sem impedir implementação de fixture técnica.

Execute R6-QUA-01-S01, S02 e S03. Fonte e reprodução: r6-06-evidencia-consolidada/explore.md.

## 6. Regras operacionais e de segurança que não podem regredir
- Custódia durável antes de aceite/ACK; UUIDv7 para operações aceitas SYNC e ASYNC. Resultado local consultável sem depender de nova busca no provedor.
- SYNC bem-sucedido devolve final na mesma requisição; não converter silenciosamente para 202. Respeitar elegibilidade de produto e orçamento HTTP publicados.
- Resposta final pública imutável; resultado tardio conserva evidência interna, reconciliação e impactos financeiros legítimos.
- Compensação possui autoridade/prazo próprios. UNKNOWN nunca autoriza reenvio cego; lease não prova ausência de efeito externo.
- Outbox/inbox em transação local com estado; Redis/L1 são derivados, nunca única autoridade. Se todas as autoridades duráveis falham, não prometer aceite sem perda usando memória.
- Clientes isolados por tenant/aplicação/ambiente. Desenvolvedor global usa identidade nominal, MFA, permissão explícita e auditoria; não usuário compartilhado.
- Contratos por cliente, credenciais compartilhadas ou dedicadas e versões congeladas por etapa. Segredos no cofre; nenhum segredo em logs, fixtures versionadas ou browser storage.
- Representação final de GET e webhook semanticamente/contratualmente idêntica, com bytes congelados quando o contrato assim exige. Não confundir parâmetros do GET com body do webhook.
- Paralelismo limitado e justo; quotas de transporte, pendência, polling, compensação e webhook. Escala automática não ultrapassa capacidade contratual/financeira.

## 7. Prepare laboratório completo e seguro
Antes de subir Docker, execute docker compose ls e docker ps -a. Mantenha um único ecossistema Compose do AI Hub ativo conforme AGENTS.md, nome estável e recursos identificados. Preserve containers de terceiros e volumes/dados. Nunca use docker system prune ou remoções amplas. Identifique recursos de execução interrompida e limpe somente os descartáveis do próprio trabalho.

Use fixtures sintéticas e segredos locais explicitamente não produtivos. Suba PostgreSQL e dependências reais exigidas pelos testes, com migrações em base vazia e populada. Verifique serviços habilitados no LocalStack, OIDC, S3 versionado e broker antes do gate. Instale ferramentas locais ausentes por mecanismo permitido com versão fixada e registro de origem; não conte ausência de CLI como conclusão da tarefa.

Execute o perfil kind independente do repositório para provar todos os componentes dentro do cluster. Não apresentar dependências via IP do Compose como qualificação do perfil independente. Perfil efêmero continua útil para smoke; restore/HA requerem perfil de dados adequado.

Não provisionar cloud paga, fazer push/merge ou implantação remota nesta instrução. Entregue alterações locais revisáveis e comandos de continuidade. Caso outra instrução vigente autorize explicitamente essas ações, respeite os gates e o escopo dessa autorização; não deduza aprovação produtiva da execução local.

## 8. Teste pelo risco e pelo caminho real
Primeiro reproduza os probes da revisão em overlay ou adapte-os para regressões dentro da suíte, mantendo a intenção do oráculo. O probe de produto deve passar por rejeitar configuração incompleta ou produzir snapshot integralmente coerente, nunca por afrouxar a assertion. O manifesto adversarial deve bloquear; crie também evidência legítima para qualificar ALLOW.

Execute go test, go test -race e go vet com a toolchain do candidato. Construa a imagem Docker candidata e identifique seu digest; Go instalado localmente pode divergir do Dockerfile. Frontend: instalação reproduzível pelo lockfile, build e browser real com OIDC/autoridades do laboratório.

Testes obrigatórios condicionados à infraestrutura devem falhar no gate integrado quando faltar dependência. Pode manter suíte unitária rápida separada, mas não promover t.Skip para PASS. Capture logs desde o início, saneados, com exit code. Não altere expectativas para aceitar defeito nem troque integração por mock para zerar falhas.

Use DB real e barreiras/fault injection para corrida de publishers, atualização parcial, takeover, callback tardio e replay. go -race não prova essas propriedades. Em restore, use dados não vazios, versões de objetos e oráculo externo antes/depois de retomar. Em carga/HA, publique perfil e budgets; avalie contenção por tenant/provedor, p95/p99, filas, memória, conexões, mensagens e valores reconciliados.

Execute OpenSpec com versão fixada compatível com o repositório. Nesta auditoria, OpenSpec 1.12.0 validou a baseline; validar todo o conjunto com `validate --all --strict --json --no-interactive`. Não usar @latest em comando que se pretende reproduzível. Validar spec não qualifica implementação.

## 9. Decisões externas e impossibilidades
Leia D-01…D-07 e T-R2-01 nos registros originais. Não há autorização para inventar volumes reais, SLA, preços, política de saldo, residência de dados, orçamento cloud ou garantia externa de provedor. Implemente configuração, validação, fixtures e testes técnicos parametrizados enquanto essas decisões não chegam. Só bloquear promoção/configuração dependente; continuar todo trabalho independente.

Quando ambiente bloquear uma ação, diagnostique causa concreta e tente alternativa permitida que preserve o oráculo. Não afirmar sucesso sem execução. Uma limitação irredutível deve trazer comando/erro, ações tentadas, artefato afetado e passo mínimo externo. Não abrir aprovação para escolha rotineira de implementação nem contornar controle de acesso.

## 10. Artefatos de execução a manter
Em docs/reviews/2026-09-13-r6/implementation/, crie e atualize:
- EXECUTION_PLAN.md: ordem, dependências e próximo item executável.
- REQUIREMENTS_STATUS.csv e SCENARIO_RESULTS.csv: inventário integral, resultado por item e evidência.
- FINDINGS_STATUS.md: cada F-R6 e vínculos R5/R3/R2, sem promover build a fechamento.
- DECISIONS_AND_BLOCKERS.md: decisão, autoridade, impacto e alternativa técnica local.
- EVIDENCE_INDEX.md e manifesto de evidências: SHA/diff, imagem, spec/cenário/log e digests, comando, ambiente, esperado e observado.
- CHECKPOINT.md: estado real, alterações preservadas, recursos locais e comando seguinte.
- FINAL_REPORT.md: implementado, testado, falhas, limites, contagens extraídas e aptidão técnica/produtiva separadas.

Não guardar logs somente no scroll. Não copiar PASS histórico para preencher matriz. Evitar evidências com hashes autorreferentes; o manifesto não precisa conter seu próprio digest. Regerar digests depois de qualquer alteração relevante e indicar exatamente que snapshot foi qualificado.

## 11. Definition of Done
Concluir todas as tarefas e verificações aplicáveis, com zero P0 aberto sem prova e zero skip obrigatório. APIs/consumidores/UI devem chamar os mecanismos novos; não basta helper sem caller. Todos os cenários têm resultado verificável, migrações e dados existentes preservados, build/imagem/integração/browser/restore/perfis operacionais qualificados conforme seu escopo. Os ambientes remotos têm artefatos completos e gates, mesmo sem implantação autorizada.

Antes de encerrar, compare tasks, findings, matriz, logs e git diff. Se houver trabalho executável pendente, continue. Se restar impedimento externo irredutível após concluir o restante, declare implementação parcial com impedimentos precisos. Nunca rotule a entrega integral por pressão desta instrução. Aprovação de produção depende também dos contratos e perfis formalmente aprovados.

Comece pela inspeção do checkout e prossiga com a execução real.
