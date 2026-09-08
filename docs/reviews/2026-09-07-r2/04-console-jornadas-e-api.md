# Console administrativo — jornadas, telas e contratos necessários

Status: proposta funcional/técnica R2. É evolução do produto administrativo, sem implementação nesta entrega. Todas as rotas abaixo são propostas, exceto as oito rotas Atlas de criação/consulta exata existentes e POST/GET de protocolo identificados no relatório.

## Atores e alcance

| Papel proposto | Pode fazer | Limite |
|---|---|---|
| Administrador de identidade | Gerir concessões individuais e escopos | Não recebe segredo de provedor ou poder financeiro automaticamente |
| Gestor de catálogo | Editar rascunhos, simular e publicar conforme aprovação | Não altera publicado ou resultado passado |
| Gestor de integrações | Contas, adapters, referências, homologação e política de pressão | Não escolhe preço nem expõe tokens; teste externo exige sandbox autorizado |
| Operações | Protocolos, SLA, entrega, reconciliação e replay permitido | Não modifica corpo final; UNKNOWN exige pré-condição de segurança |
| hub_protocol_reader | Leitura administrativa entre tenants para desenvolvedores | Individual, MFA, auditado, mascarado; sem comandos operacionais |
| Financeiro | Planos, extratos, ajustes autorizados, fechamento/conciliação | Segregação entre preparação e aprovação de ajustes conforme P-09/P-06 |
| Leitor do tenant | Somente recursos e relatórios da aplicação/cliente autorizado | Não herda a leitura global do desenvolvedor |

Papéis são funções propostas; não presumem nomes nem concessões já existentes. Backend aplica RBAC e escopo por recurso. Desabilitar botão não é autorização.

## Navegação proposta

Rotas agrupadas por trabalho: visão operacional; clientes/aplicações; catálogo (serviços, produtos, ofertas, importações); integrações (provedores, contas, adapters, vínculos, capacidade); contratos/políticas; protocolos; entregas/quarentena; financeiro; SLA/relatórios; auditoria; plataforma/onboarding. Ambiente e escopo ficam visíveis na barra superior. O uso entre tenants exige contexto administrativo explícito. Não transformar o console em editor de SQL, shell, Terraform ou segredos.

| Jornada / requisito | Lista e filtros mínimos | Detalhe/formulário e ações | API/dono propostos | Aceite e dependências |
|---|---|---|---|---|
| ADM-01 / R2-ADM-01 | Identidade, ambiente e escopo | Login, retorno à URL, expiração e logout | IdP + Portal; GET /admin/v1/me | Sem sessão não carrega dados; change 01 |
| ADM-02 / R2-ADM-02 | Busca, estado, versão, período; cursor | Detalhe por ID, editar rascunho, conflito, histórico | GET /admin/v1/{recurso}; ETag/If-Match | Refresh consulta servidor; 409/412 sem sobrescrita; changes 01/04 |
| ADM-03 / R2-ADM-03 | Clientes por nome/ID/estado/classe/célula | Aplicações, grants, ofertas, pendências e ativação/suspensão | Atlas: /clients, /applications, /offers, /onboardings | Novo cliente ativa só com elegibilidade e capacidade; 04/08 |
| ADM-04 / R2-ADM-04 | Serviços, versão, estado, provedor, adapter; importados separados | Schemas, modalidades, classificação, diff e homologação | Atlas: /services, /imports, /publications | Quantidade importada não significa disponível; 03/04 |
| ADM-05 / R2-ADM-05 | Produtos por versão/estado/tipo | DAG e tabela equivalente, passos/dependências, parcialidade, compensação, simulação | Atlas /products e /simulations; Órbita executa publicado | A/B paralelos, C dependente e ciclo bloqueado; 02/04/06 |
| ADM-06 / R2-ADM-06 | Provedor, conta, ambiente externo, modo, estado do binding | Capacidade/saúde, secret_ref mascarada, key_id, rotação e teste homologado | Atlas /providers, /provider-accounts, /credential-bindings; Cometa /capacity-domains | Dois clientes na mesma conta preservam identidade; 03 |
| ADM-07 / R2-ADM-07 | Contrato/perfil por cliente/aplicação/oferta/vigência | Schema/transformação entrada/saída/erro; TTL, SLA, polling, callback; valores efetivos | Atlas /technical-profiles e /policies; /simulations | Sem campos livres para enum e sem alterar snapshot; 04 |
| ADM-08 / R2-ADM-08 | UUID, tenant, aplicação, produto, provedor, status, período e quebra de SLA | Timeline, passos, tentativas, recibos, prazo, final, custos permitidos | Órbita /admin/v1/protocols e /{id}/timeline; projeções de Cometa/Libra/Pulsar | Sem chamada externa para buscar final; leitura global auditada; 02/03/06 |
| ADM-09 / R2-ADM-09 | Entregas por destino/estado/idade; quarentena por tipo/causa | Reentrega, retry de mensagem, reconciliação, justificativa e recibo de ação | Pulsar /deliveries; donos de fila /quarantine; Órbita /reconciliations | Reentregar não executa produto; reader não comanda; 02/03 |
| ADM-10 / R2-ADM-10 | Coorte, contrato/versão, tenant, serviço/provedor, janela | População elegível, pendentes, vencidos, percentis, excedente, evidências e exportação | Projeção operacional: /sla-reports; Grafana autorizado | Mostrar watermark/atualidade, sem viés de rápidos; 03/08 |
| ADM-11 / R2-ADM-11 | Compra/venda, moeda, competência, origem, estado de fechamento | Planos, franquias, saldo/holds, journal, ajuste, conciliação, exportação | Atlas contratos/planos; Libra /admin/v1/finance/* | Decimal/contrato/origem; sem apagar lançamento; 06 |
| ADM-12 / R2-ADM-12 | Todas as jornadas | Teclado, foco, validação, estados acessíveis e retorno seguro | Frontend + contratos tipados das APIs | WCAG 2.2 AA proposta, ensaio integrado, viewport 360px/desktop |

## Campos e validações de maior risco

| Formulário | Campos/valores | Validação cruzada e comportamento |
|---|---|---|
| Serviço | code, version, input_schema, output_schema, data_class, modes, SLA e retry TTL | version inteira positiva; SLA >0; TTL >=0 inteiro. Publicado exige schema e adapter compatíveis |
| Produto | tipo agregação/composição, step_id, service_version, dependências, obrigatório, mapeamentos, consolidação, compensação | DAG sem ciclo; referências existentes; fan-out e complexidade dentro do perfil; resultado parcial apenas quando contratado |
| Conta | provider_id, provider_account_id, ambiente externo, base URL, capabilities, capacity_domain | Ambiente externo pode ser sandbox/HML do parceiro; não confundir com local/dev/hom/ppd/prd do Hub. URLs seguras e modo homologado |
| Binding | SHARED_HUB/TENANT_DEDICATED, tenant/aplicação pertinente, conta, secret_ref, secret_version, vigência, estado, settlement_party | Dedicado exige tenant; compartilhado não herda tenant arbitrário; unicidade por escopo; um ativo elegível, sem LIMIT 1 ambíguo |
| Autenticação externa | NONE somente em perfil permitido; BASIC/API_KEY/OAUTH/MTLS conforme adapter; refs de certificado/segredo; audience/scope | Referência nunca é senha. Não testar conta real automaticamente ao editar; teste exige destino/permissão de sandbox |
| Oferta | cliente/aplicação, produto/serviço, perfil técnico, compra/venda, rota, modalidade, classe, prazo e vigência | Publicação valida capacidade, segredo, SLA e compatibilidade. Campo provider_account_id legado só pode restringir rota já autorizada |
| SLA/retry | client_sla_seconds, sync_http_budget_seconds, finalization_reserve_seconds, provider_sla_seconds, provider_sla_start, enforcement, retry_ttl_seconds, auto_wait_seconds | Explicar menor deadline efetivo. TTL não renova SLA. auto_wait não ultrapassa prazo. Polling saudável não impede expiração |
| Polling | enabled, interval_seconds, max_interval_seconds, jitter, operações status/fetch, política de erro | Validade por versão; claims e quota incluem STATUS/FETCH; não ignorar HTTP status, Retry-After ou resposta inválida |
| Webhook | destination_id/version, URL verificada, perfil, key_id, retry/deadline e suspensão | Versão congelada na obrigação. Timestamp/HMAC. Alterar URL não redireciona entrega antiga silenciosamente |
| Plano | moeda, preço decimal, vigência, medidores, franquia/faixas, pacote/soma, parcialidade, custo/receita | Não usar Number/float para apuração. Mostrar exemplos calculados no servidor e regra de arredondamento |

## Contratos administrativos transversais

Prefixo proposto `/admin/v1` roteado pelo Portal para a autoridade do recurso; não supõe um novo microserviço central. Listas usam `limit` com teto de perfil e cursor opaco vinculado a filtro/ordenação/escopo; ordenação estável por chave única secundária. Campos de filtro permitidos são allowlist. Relatórios extensos viram job assíncrono autorizado com arquivo protegido.

POST de criação/comando administrativo usa Idempotency-Key quando produzir efeito repetível. GET de detalhe retorna ETag da revisão; PATCH de rascunho exige If-Match (412 se revisão desatualizada). Imutabilidade de publicado ou transição inválida é 409. Validação de campos/relacionamentos é 422 com `field_errors`; JSON inválido 400; autenticação 401; autorização 403; recurso alheio/inexistente 404 no escopo público; cota 429; dependência 503. Todos possuem código estável e correlação sanitizada; não enviar SQL/stack.

Criação de nova versão é ação diferente de editar publicado. Publicação contém `target_id`, `draft_revision`, `validation_report_id` e razão; resposta identifica versão/hash e estado de distribuição. Simulação contém apenas amostras sintéticas/sanitizadas, referências de versões e cenário; retorno separa validação, plano de execução e estimativa econômica, sem efeito externo real.

Consulta de protocolo administrativo retorna visão operacional com dados mascarados e evidências por escopo; **não altera o contrato público nem precisa ter o corpo idêntico ao webhook**. A igualdade requerida aplica-se ao resultado público do cliente e payload enviado ao cliente.

## Estados de tela e qualidade

Carregamento inicial usa indicação anunciável; vazio traz ação permitida, erro traz causa tratável e retry; resultado antigo fica marcado com horário. Submit desabilita somente ação concorrente relevante. Abort/cancel de request evita resposta antiga substituir novo filtro. Validação servidor é autoridade; UI antecipa erros sem divergir. Busca por ID aceita colar UUID; chips de estado não dependem só de cor. Tabela larga tem área navegável e alternativa de detalhe; grafo também tem lista de dependências. Foco retorna após diálogo; não exigir arrastar para compor produto. Confirmação de publicação/suspensão/replay informa impacto real, razão e versão, sem checklists genéricos.

A [WCAG 2.2](https://www.w3.org/TR/WCAG22/) é referência para critérios de acessibilidade; esta revisão propõe AA como gate das jornadas entregues, sem alegar conformidade já obtida. Testes automatizados devem ser combinados com teclado, foco e leitor de tela nas rotas principais.
