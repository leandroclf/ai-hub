# Revisão R4 — AI Hub
**O merge avançou correções específicas, mas não concluiu v4/R2/R3.**
Foram comparados os arquivos do commit [b9d0f90ce02a](https://github.com/leandroclf/ai-hub/commit/b9d0f90ce02aa0c27cad546745153d160ff5867f)
com a revisão anterior. Há 107 arquivos adicionados/alterados, dos quais 18 estão sob hub/;
nenhum arquivo do frontend administrativo foi alterado.

## Evidência versus conclusão anunciada
O título da PR #2 usa “conclui implementação e pacote de qualificação integral”.
O FINAL_REPORT versionado declara implementação parcial; CHECKPOINT registra alterações e validações posteriores
não refletidas no relatório. Conclusão técnica deve seguir comportamento e evidência, não título do commit.

[docs/reviews/2026-09-08-r3/implementation/FINAL_REPORT.md:1](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/docs/reviews/2026-09-08-r3/implementation/FINAL_REPORT.md#L1);
[docs/reviews/2026-09-08-r3/implementation/CHECKPOINT.md:1](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/docs/reviews/2026-09-08-r3/implementation/CHECKPOINT.md#L1).

## Avanços confirmados
- Inteiro grande preservado e enum simples rejeitado: provas anteriores passam.
- API_KEY atravessa conta, projeção, executor e polling.
- Bearer tokens não são mais gravados/lidos no Redis; coordenação é por chave.
- Leitura administrativa exige papel e MFA.
- Callback ganhou capability, retorno de erro de custódia e inbox.
- Oferta deixou de recusar portfólio acima de 100.
Esses avanços não fecham automaticamente os requisitos completos: veja a matriz dos 25 achados.

## Verificações executadas nesta auditoria
| Verificação | Resultado e alcance |
|---|---|
| go test -race -json ./... | Exit 0; 37 testes pass e 18 skip. Não é qualificação integrada. |
| Frontend: TypeScript e Vite | Exit 0; dependências locais reutilizadas após igualdade de package-lock. Instalação limpa não foi qualificada. |
| OpenSpec 1.12.0 strict --all | 17/17 changes válidos, sem config.yaml. A ausência desse arquivo não impediu a execução. |
| Provas adversariais JSON | Quatro defeitos reproduzidos: null/string, string/integer, documento concatenado e minimum ignorado. |
| Regressão anterior de precisão/enum | PASS; não atribuir novamente os defeitos antigos já corrigidos. |
| L1 com token válido e cofre indisponível | FAIL esperado; chamada tenta Resolve antes do cache. |
| Inventário das specs | 189 requisitos; 696 cenários, incluindo 416 v4 omitidos na matriz anterior. |
| Upgrade de migração | Hash histórico mudou e runner rejeita mismatch; não executado contra PostgreSQL nesta revisão. |
| DB/S3/OIDC/Compose/kind/browser/carga/HA | Não executados nesta sessão. Docker/psql não disponíveis no runtime inspecionado. |

Instrumentos por Go overlay não alteram os fontes do checkout. Logs e comandos estão no pacote.
O build não demonstra jornada de usuário, e um teste pulado não é aprovação.

## Severidade e natureza da evidência
P0: custódia, integridade, autenticação ou upgrade que impede qualificação segura.
P1: completude, contrato, desempenho e evidência necessários ao escopo.
Defeitos reproduzidos são explicitamente indicados. Riscos de concorrência/upgrade são inferências
do caminho lido, com cenário de confirmação exigido; não são incidentes observados em produção.

## Achados R4

### F-R4-01 — Rota de callback compatível com a autenticação do provedor (P0)

O handler ganhou capability por operação e custódia antes de 2xx. Entretanto, toda a rota /internal/ continua envolvida pelo JWT do Hub; a URL entregue ao provedor contém apenas capability. O simulador não envia JWT Hub. A correção do handler não fecha o caminho de rede. Isso é incompatibilidade de composição, não prova de endpoint publicamente desprotegido.

**Exigência:** O Hub SHALL expor uma rota de callback com autenticação explicitamente homologada por conta e independente da credencial de workload interna. O caminho público/gateway/middleware/handler deve ser qualificado de ponta a ponta. Capability de operação pode complementar correlação, nunca substituir silenciosamente política da conta; segredos não devem aparecer em URLs registradas, logs ou traces.

**Fontes:** [hub/cmd/cometa/main.go:104](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/cmd/cometa/main.go#L104), [hub/internal/cometa/handlers.go:116](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/handlers.go#L116), [hub/internal/cometa/executor.go:273](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/executor.go#L273), [hub/internal/platform/httpserver/httpserver.go:94](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/platform/httpserver/httpserver.go#L94).

**Rastreio:** R4-CBK-01 → R3-EXE-02, R2-SEG-04, R2-INT-04. Responsável: Integrações e Core.

### F-R4-02 — Admissão e deduplicação segura da inbox órfã (P0)

Quando a operação não existe, qualquer token não vazio que alcance o handler permite inserir body e receber 202. A chave única usa operação+hash do body, sem identidade autenticada/token: primeira tentativa com token incorreto pode ocupar a identidade de posterior recibo correto, que só incrementa occurrences. A tabela também não registra tenant/conta/célula, TTL ou quota. Exploração externa depende da rota/autenticação atual; o risco permanece ao corrigir essa rota.

**Exigência:** O Hub SHALL autenticar a origem antes de admitir órfão e vincular custódia a tenant/conta/célula/política e identidade do evento. Uma tentativa inválida não pode ocupar a chave de deduplicação de recibo legítimo. Quota, retenção, autorização e disposição devem ser explícitas; conflito de identidade mantém evidência sem descartar o evento correto.

**Fontes:** [hub/internal/cometa/handlers.go:98](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/handlers.go#L98), [hub/internal/cometa/custody.go:114](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/custody.go#L114), [hub/migrations/core/0032_callback_inbox.sql:14](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/migrations/core/0032_callback_inbox.sql#L14).

**Rastreio:** R4-CBK-02 → R3-EXE-02, R3-OPE-04, R2-SEG-04. Responsável: Integrações e Core.

### F-R4-03 — Recuperador autônomo limitado e independente do HTTP (P0)

ReconcileCallbackInbox só é chamado ao receber outro callback conhecido. A consulta percorre todos os RECEIVED associados, sem LIMIT, claim, lease ou escopo de célula. O cursor permanece aberto enquanto são feitas outras operações SQL e aplicação. O HTTP de uma operação pode depender do backlog/erro de outra, e sem novo callback não há gatilho autônomo.

**Exigência:** O Hub SHALL reconciliar inbox por worker durável com lote limitado, claim/lease/epoch e isolamento por célula/conta/tenant. Aceite de callback conhecido não deve aguardar drenagem global. A recuperação deve ocorrer após reinício/correlação sem depender de tráfego futuro; um item inválido deve ter disposição sem bloquear os seguintes.

**Fontes:** [hub/internal/cometa/custody.go:131](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/custody.go#L131), [hub/internal/cometa/handlers.go:157](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/handlers.go#L157).

**Rastreio:** R4-CBK-03 → R3-EXE-02, R3-EXE-03, R3-INT-01, R2-EXE-09. Responsável: Integrações e Core.

### F-R4-04 — Uma validação de resultado para SUBMIT, polling e callback (P0)

finalize valida OutputSchema e conserva OperationResult. ApplyExternalObservation transforma callback somente em detail, não chama a mesma validação e não confere provider_request_id contra correlação armazenada. ConserveObservation pode sobrescrever a correlação com qualquer ID não vazio recebido. Token de operação não torna o body semanticamente correto.

**Exigência:** O Hub SHALL aplicar uma única validação de resultado externo para todas as modalidades, verificando conta, operação, correlação, schema e snapshot antes de alterar o estado. Conservar recibo original protegido e resultado normalizado versionado; divergência recebe disposição inválida sem substituir correlação/final. Não truncar a resposta a campos de simulador.

**Fontes:** [hub/internal/cometa/executor.go:319](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/executor.go#L319), [hub/internal/cometa/executor.go:232](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/executor.go#L232), [hub/internal/cometa/custody.go:288](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/custody.go#L288).

**Rastreio:** R4-CBK-04 → R3-EXE-01, R3-EXE-02, R3-CAT-02, R2-EXE-04. Responsável: Integrações e Core.

### F-R4-05 — Documento JSON único e tipos sem coerção (P0)

Provas novas executadas: null foi aceito como string; "123" como integer; e dois documentos concatenados retornaram sem erro. Decoder lê apenas o primeiro documento e, sem mapping, devolve input original. A preservação do inteiro grande anterior foi corrigida e não deve regredir.

**Exigência:** O Hub SHALL aceitar exatamente um documento JSON completo e validar tipo JSON sem coerção por representação Go. null só é válido quando permitido explicitamente; strings numéricas não são números. Entrada deve ser totalmente consumida antes de retornar/persistir o payload, com mesma regra com/sem mapping.

**Fontes:** [hub/internal/atlas/offers.go:179](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L179), [hub/internal/atlas/offers.go:330](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L330).

**Rastreio:** R4-CTR-01 → R3-CAT-01, R2-CAT-05. Responsável: Core e Catálogo.

### F-R4-06 — Dialeto de schema publicado e semântica numérica (P1)

Prova executada: minimum:0 é ignorado e -1 aceito. A validação de publicação verifica só type object/properties. O comentário de recusa a construções não qualificadas não corresponde a whitelist efetiva. integer rejeita representação com .eE e enum compara json.Number lexicalmente, divergindo da semântica usual de JSON Schema.

**Exigência:** O Hub SHALL declarar dialeto/subconjunto e compilar/validar schemas na publicação, recusando keywords não suportadas em qualquer nível. Publicação e runtime usam a mesma semântica; equivalência numérica e tipo integer devem seguir o dialeto documentado com precisão exata. Limitar profundidade/custo sem ignorar regras.

**Fontes:** [hub/internal/atlas/offers.go:266](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L266), [hub/internal/atlas/offers.go:355](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L355), [hub/internal/atlas/catalog.go:327](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/catalog.go#L327).

**Rastreio:** R4-CTR-02 → R3-CAT-01, R2-CAT-01, R2-CAT-05. Responsável: Core e Catálogo.

### F-R4-07 — L1 utilizável na falha de cofre e coordenação limitada (P1)

Tokens deixaram Redis e lock passou a ser por chave. Contudo, Resolve é chamado antes do L1; prova com L1 válido e cofre indisponível falha. locks cresce sem remoção por binding/versão e mutex não respeita cancelamento durante espera.

**Exigência:** O Hub SHALL consultar cache válido por identidade/versionamento autorizado antes de buscar segredo remoto, respeitar expiração/revogação e coordenar renovação com espera cancelável por chave. Entradas de cache e estruturas de coordenação devem ter limites/evicção seguros sob rotação e crescimento, sem fallback para outro binding.

**Fontes:** [hub/internal/providerauth/client.go:74](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/providerauth/client.go#L74), [hub/internal/providerauth/client.go:144](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/providerauth/client.go#L144).

**Rastreio:** R4-OPE-01 → R3-INT-02, R3-OPE-04, R2-INT-03. Responsável: Plataforma, Dados e Integrações.

### F-R4-08 — Upgrade com migração histórica imutável (P0)

0002_provider_auth.sql já existente na R2 foi alterada para API_KEY e header. O runner checksum-guardado para em 0002 de um banco previamente migrado, antes de executar a nova 0004. Comparação dos bytes/hashes comprova alteração; falha SQL integrada ainda não foi executada nesta auditoria.

**Exigência:** A entrega SHALL preservar migrações aplicadas e implementar mudanças por migrações aditivas. Qualificar instalação limpa e upgrade de volume com checksum R2 anterior sem apagar dados ou desabilitar verificação. Bancos já inicializados com variante modificada precisam reconciliação explícita e restrita a hashes/estados conhecidos, não atualização cega do ledger.

**Fontes:** [hub/migrations/control/0002_provider_auth.sql:3](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/migrations/control/0002_provider_auth.sql#L3), [hub/migrations/control/0004_provider_api_key.sql:2](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/migrations/control/0004_provider_api_key.sql#L2), [hub/deploy/r2/scripts/migrate.sh:22](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/deploy/r2/scripts/migrate.sh#L22).

**Rastreio:** R4-OPE-02 → R3-OPE-02, R3-OPE-04, R2-OPE-08, R2-DAD-05. Responsável: Plataforma, Dados e Integrações.

### F-R4-09 — Consulta de oferta seletiva sem materializar o portfólio (P1)

Paginação eliminou recusa acima de 100, mas agrega todas as páginas em resources antes de filtrar aplicação/serviço. Portanto CPU/memória/round-trips continuam proporcionais ao total de ofertas do tenant e o control plane ainda é consultado por pedido.

**Exigência:** O Hub SHALL resolver pelo conjunto elegível indexado de tenant/aplicação/alvo/versão/vigência, sem materializar catálogo inteiro por pedido, e usar projeção válida conforme a baseline. A prova de crescimento deve medir round-trips/memória/latência e ambiguidade verdadeira, não apenas capacidade de encontrar registro 101.

**Fontes:** [hub/internal/atlas/offers.go:33](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/offers.go#L33), [hub/internal/atlas/catalog.go:419](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlas/catalog.go#L419), [hub/internal/atlasclient/client.go:22](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/atlasclient/client.go#L22).

**Rastreio:** R4-OPE-03 → R3-CAT-04, R3-INT-03, R2-CAT-06. Responsável: Plataforma, Dados e Integrações.

### F-R4-10 — Inventário integral e evidência coerente com o conteúdo (P1)

As specs têm 189 requisitos e 696 cenários (416 v4+205 R2+75 R3), mas matriz registra 280, omitindo cenários v4. FINAL_REPORT diz OpenSpec não executado; CHECKPOINT informa strict 17/17 e novas mudanças. Título do merge afirma conclusão, incompatível com gates ainda abertos.

**Exigência:** A engenharia SHALL gerar inventário diretamente de todas as specs, preservar os 696 cenários herdados e acrescentar R4 sem omissões. Relatório/checkpoint/matrizes devem ser consistentes e vinculados a commit+hash de conteúdo/digests. PASS exige resultado verificável; histórico e evidência do working tree devem ter proveniência distinguível.

**Fontes:** [docs/reviews/2026-09-08-r3/implementation/SCENARIO_RESULTS.csv:1](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/docs/reviews/2026-09-08-r3/implementation/SCENARIO_RESULTS.csv#L1), [docs/reviews/2026-09-08-r3/implementation/FINAL_REPORT.md:1](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/docs/reviews/2026-09-08-r3/implementation/FINAL_REPORT.md#L1), [docs/reviews/2026-09-08-r3/implementation/CHECKPOINT.md:1](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/docs/reviews/2026-09-08-r3/implementation/CHECKPOINT.md#L1).

**Rastreio:** R4-QUA-01 → R3-QUA-01, R2-QUA-01, R2-QUA-04. Responsável: Engenharia, Frontend e Qualidade.

### F-R4-11 — Entrega vertical do console e baseline remanescente (P1)

Nenhum arquivo de hub/admin-ui mudou entre os snapshots. Permanecem a escolha incorreta de delivery_id, SLA/reconcile sem rota efetiva e comandos financeiros incompatíveis. Domínios financeiro, DAG, capacidade, objetos e kind também não receberam fechamento funcional neste delta.

**Exigência:** A próxima implementação SHALL concluir o backlog remanescente v4/R2/R3 por fatias verticais, com contratos backend/UI, persistência, oráculos e browser. Cada jornada só termina quando seu efeito real e autorização são demonstrados; componentes auxiliares e documentação isolados não satisfazem a tarefa. Os requisitos herdados não são substituídos pelos novos R4.

**Fontes:** [hub/admin-ui/src/pages/OperationsPage.tsx:8](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/admin-ui/src/pages/OperationsPage.tsx#L8), [hub/admin-ui/src/pages/FinancePage.tsx:5](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/admin-ui/src/pages/FinancePage.tsx#L5), [hub/internal/orbita/admin.go:14](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/orbita/admin.go#L14).

**Rastreio:** R4-QUA-02 → R3-ADM-02, R3-ADM-03, R3-ADM-04, R3-QUA-01. Responsável: Engenharia, Frontend e Qualidade.

### F-R4-12 — Qualificação reproduzível com um ecossistema local (P1)

go test -race passa com 37 testes pass e 18 skip nesta revisão. OpenSpec strict 17/17 passa sem config.yaml; não é bloqueio técnico atual. Não há comprovação nova de DB/Compose/kind/browser/HA. AGENTS atualizado exige um ecossistema Compose por vez, contrariando prompts históricos que sugeriam laboratórios paralelos.

**Exigência:** A qualificação SHALL preparar ferramentas fixadas, executar gates integrados sem skip obrigatório e cumprir um único ecossistema Compose ativo do Hub por vez. Inventariar containers/projeto, preservar volumes e terceiros, trocar somente alvos identificados e registrar limpeza. Ausência real de runtime é bloqueio delimitado, não PASS nem motivo para abandonar implementação independente.

**Fontes:** [AGENTS.md:14](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/AGENTS.md#L14), [hub/internal/cometa/custody_test.go:21](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/internal/cometa/custody_test.go#L21), [hub/deploy/r2/kind/render-runtime.py:1](https://github.com/leandroclf/ai-hub/blob/b9d0f90ce02aa0c27cad546745153d160ff5867f/hub/deploy/r2/kind/render-runtime.py#L1).

**Rastreio:** R4-QUA-03 → R3-QUA-01, R3-OPE-02, R3-OPE-03, R3-OPE-05. Responsável: Engenharia, Frontend e Qualidade.

## Pendências herdadas
A R4 não repete como “novos” todos os achados anteriores. DAG, política efetiva, UNKNOWN/fencing,
controle adaptativo, financeiro, destinos congelados, objetos e plataforma continuam no backlog obrigatório.
A matriz de 25 achados detalha o delta e aponta as specs originais. Sem concluir esse backlog,
corrigir apenas os 12 achados R4 não entrega o objetivo do Hub.

## Limite da conclusão
Não há comprovação de disponibilidade para operações críticas, isolamento sob carga ou ausência de perda
no modelo de falhas requerido. Há custódia local implementada parcialmente e requisitos ainda desconectados.
A próxima execução deve produzir fluxos verticais reais e reconciliar a evidência com o conteúdo entregue.
