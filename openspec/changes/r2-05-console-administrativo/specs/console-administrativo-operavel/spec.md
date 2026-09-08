# Delta for console-administrativo-operavel

## ADDED Requirements

### Requirement: R2-ADM-01 — Sessão, contexto e navegação administrativa

Console SHALL exigir sessão autenticada, mostrar ambiente e escopo efetivo, oferecer navegação por URL retomável e disponibilizar ações conforme permissões do servidor. Contexto administrativo entre tenants SHALL ser explícito e não alterar a autorização pública. Dados da sessão anterior SHALL ser limpos ao mudar identidade/tenant.

Baseline relacionada: CFG-01, SEG-01, SEG-04.

#### Scenario: R2-ADM-01-S01 — Entrada autorizada

- GIVEN usuário tem papel de Operações e escopo A
- WHEN abre diretamente URL de protocolo de A
- THEN após login retorna à rota permitida e vê ambiente/escopo e ações autorizadas

#### Scenario: R2-ADM-01-S02 — URL sem permissão

- GIVEN operador de A abre URL de B ou tela financeira sem papel
- WHEN servidor avalia a sessão
- THEN exibe acesso negado sem dados, e ocultar menu não é o único controle

#### Scenario: R2-ADM-01-S03 — Logout

- GIVEN usuário visualizou dados restritos
- WHEN sai e outro usuário entra no mesmo navegador
- THEN não reaparecem dados, filtros sensíveis ou permissões do anterior

### Requirement: R2-ADM-02 — Listagens reais e detalhe editável

Console SHALL consultar listagens persistentes com busca/filtros/paginação e detalhe por URL. Recursos SHALL ter estados carregando, vazio, erro, acesso negado e desatualizado distinguíveis. Edição SHALL carregar versão atual e tratar conflito de revisão; memória da aba NÃO SHALL substituir listagem do servidor.

Baseline relacionada: CFG-01, CAT-01, CFG-06.

#### Scenario: R2-ADM-02-S01 — Catálogo populoso

- GIVEN servidor tem mais registros que uma página
- WHEN operador busca e avança cursor
- THEN filtro/ordenação se mantêm, volumes por página são limitados e refresh recompõe a consulta

#### Scenario: R2-ADM-02-S02 — Lista falha

- GIVEN API responde timeout após mostrar dados antigos
- WHEN usuário solicita atualização
- THEN UI sinaliza erro e desatualização, oferece retry seguro e não apresenta lista vazia como ausência de registros

#### Scenario: R2-ADM-02-S03 — Conflito de edição

- GIVEN dois operadores editam uma mesma versão
- WHEN segundo tenta salvar
- THEN UI apresenta conflito e diff/recarregamento sem sobrescrita silenciosa

### Requirement: R2-ADM-03 — Clientes, aplicações e ofertas contratadas

Console SHALL permitir criar e consultar cliente/aplicação, grants e ofertas com vigência, ambiente, classe de isolamento e estado do onboarding. Ativação SHALL expor validações pendentes e depender de capacidade/credenciais/contratos elegíveis. Suspensão SHALL mostrar impacto em novos pedidos e conservar histórico.

Baseline relacionada: CFG-01, CFG-06, CAT-02, CAT-11.

#### Scenario: R2-ADM-03-S01 — Novo cliente

- GIVEN editor autorizado informa cliente, aplicações e oferta validada
- WHEN solicita ativação
- THEN acompanha validação/placement até ativo sem ticket de capacidade dentro do envelope

#### Scenario: R2-ADM-03-S02 — Pendência visível

- GIVEN oferta não tem vínculo de credencial válido
- WHEN operador tenta ativar
- THEN UI mostra bloqueio específico e link para corrigir o vínculo, sem sucesso fictício

#### Scenario: R2-ADM-03-S03 — Suspensão

- GIVEN cliente tem protocolos em andamento
- WHEN operador suspende novas admissões
- THEN confirma impacto e razão; timeline histórica continua acessível ao papel autorizado

### Requirement: R2-ADM-04 — Catálogo importado e serviços executáveis

Console SHALL distinguir inventário importado de serviço executável, expor método/path/auth declarada sem segredos, schemas, adapter, versão, modalidades e estado de homologação. Publicação SHALL mostrar diff e relatório de validação. Campos enumerados SHALL usar escolhas válidas fornecidas pelo contrato.

Baseline relacionada: CAT-01, CAT-02, CFG-02, COM-02.

#### Scenario: R2-ADM-04-S01 — Importação por preview

- GIVEN arquivo sanitizado contém endpoints existentes e novos
- WHEN operador revisa o lote
- THEN vê diferenças e estados, sem exclusão global ou publicação automática

#### Scenario: R2-ADM-04-S02 — Serviço validado

- GIVEN adapter/schema/modo foram homologados
- WHEN operador publica a versão
- THEN UI apresenta versão publicada imutável e qualificação usada

#### Scenario: R2-ADM-04-S03 — Sem adapter

- GIVEN endpoint importado ainda não pode executar
- WHEN operador consulta catálogo
- THEN estado deixa claro que não está disponível para consumo

### Requirement: R2-ADM-05 — Produtos agregados e compostos

Console SHALL editar produto como passos, dependências, obrigatoriedade, mapeamentos e política de consolidação/compensação. SHALL oferecer representação de grafo e alternativa tabular acessível, validar ciclos e simular cenários sem efeitos em provedores reais. Publicação SHALL conservar versão e mostrar impacto comercial/operacional.

Baseline relacionada: CAT-03, CAT-04, CAT-05, CAT-07, CAT-08.

#### Scenario: R2-ADM-05-S01 — Editor de composição

- GIVEN produto possui A e B paralelos e C dependente
- WHEN editor configura e simula
- THEN grafo/tabela e resultado de simulação mostram ordem, dependências e política de parcialidade

#### Scenario: R2-ADM-05-S02 — Ciclo detectado

- GIVEN dependência fecha ciclo
- WHEN editor valida
- THEN UI identifica os passos envolvidos e bloqueia publicação

#### Scenario: R2-ADM-05-S03 — Simulação sem efeitos

- GIVEN usuário usa dados sintéticos
- WHEN clica simular
- THEN nenhuma chamada faturável real é feita; relatório identifica fixture e versões

### Requirement: R2-ADM-06 — Provedores, vínculos e saúde de integração

Console SHALL separar provedor, conta, ambiente externo, binding, modo e pagador; apresentar apenas referências mascaradas e metadados de segredo. SHALL permitir gerir ciclo/rotação via fluxo autorizado, verificar integração em sandbox homologado e consultar domínio de capacidade, pressão atual, motivos de redução e pendências.

Baseline relacionada: CFG-05, SEG-05, OPE-07, CAT-06.

#### Scenario: R2-ADM-06-S01 — Diagnóstico de binding

- GIVEN cliente dedicado está usando determinada conta
- WHEN operador autorizado consulta resolução
- THEN vê binding/version/pagador elegíveis sem revelar token ou segredo e sem confundir conta com tenant

#### Scenario: R2-ADM-06-S02 — Rotação

- GIVEN gestor autorizado agenda nova versão
- WHEN validação de conexão conclui
- THEN UI mostra estado e vigência, preservando histórico e não oferecendo leitura do segredo

#### Scenario: R2-ADM-06-S03 — Pressão reduzida

- GIVEN provedor responde timeouts
- WHEN operador abre saúde da integração
- THEN vê limite efetivo, teto, latência, erro, backlog e motivo da redução, sem tratá-lo como limite fixo eterno

### Requirement: R2-ADM-07 — Contratos técnicos e políticas temporais

Console SHALL configurar por oferta/cliente os perfis de entrada/saída/erro, modalidades, TTL em segundos, SLA bilateral, enforcement, polling/callback e entrega. SHALL mostrar precedência e valores efetivos, validar incompatibilidades e simular deadline/retry antes de publicar. Edição de versão vigente NÃO SHALL afetar protocolo aceito.

Baseline relacionada: CAT-09, CFG-04, FIN-03, EXE-10, EXE-11.

#### Scenario: R2-ADM-07-S01 — Contrato legado

- GIVEN cliente usa saída personalizada e polling mais callback
- WHEN editor simula e publica perfil
- THEN UI exibe entrada/saída e prova de mesmo corpo final em GET/webhook

#### Scenario: R2-ADM-07-S02 — TTL versus SLA

- GIVEN retry_ttl_seconds excede prazo do cliente
- WHEN editor salva configuração
- THEN UI explica limite efetivo mínimo; nenhum preview sugere renovação do deadline

#### Scenario: R2-ADM-07-S03 — SYNC incompatível

- GIVEN SLA ou natureza do provedor não permite SYNC
- WHEN editor tenta habilitar modalidade
- THEN relatório aponta incompatibilidade e bloqueia publicação

### Requirement: R2-ADM-08 — Busca e timeline operacional de protocolos

Console SHALL pesquisar protocolos por identificador e filtros autorizados de cliente/aplicação/oferta/provedor/estado/período, sem query ilimitada. Detalhe SHALL mostrar timeline de aceite, snapshots, passos, tentativas, recibos, decisões de SLA, resultado e entregas. hub_protocol_reader SHALL poder consultar outros tenants somente pela superfície administrativa auditada.

Baseline relacionada: CFG-03, SEG-04, EXE-07, DAD-04.

#### Scenario: R2-ADM-08-S01 — Diagnóstico completo

- GIVEN protocolo possui tentativas e resultado tardio
- WHEN operador autorizado abre detalhe
- THEN timeline diferencia erro do cliente, final do provedor, custo e delivery e exibe causa de encerramento

#### Scenario: R2-ADM-08-S02 — Resultado custodiado

- GIVEN provedor está offline e final está no Hub
- WHEN operador consulta o resultado
- THEN nenhuma chamada ao provedor ocorre; resultado provém da versão materializada

#### Scenario: R2-ADM-08-S03 — Filtro entre tenants

- GIVEN usuário sem papel global manipula filtros
- WHEN solicita lista de outro cliente
- THEN API e UI não revelam protocolos alheios

### Requirement: R2-ADM-09 — Entregas e reprocessamento com segurança de efeito

Console SHALL distinguir reentregar webhook, reconciliar operação, reprocessar mensagem e executar novo pedido. Cada ação SHALL exigir permissão própria, pré-condições, justificativa e registro de resultado. Ações sobre UNKNOWN NÃO SHALL oferecer novo submit como retry genérico. Estado final do cliente NÃO SHALL ser editável manualmente.

Baseline relacionada: CFG-03, EXE-08, EXE-09, SEG-04.

#### Scenario: R2-ADM-09-S01 — Webhook esgotado

- GIVEN resultado final existe e delivery está EXHAUSTED
- WHEN operador autorizado reentrega
- THEN cria ação auditada ligada à entrega e aos mesmos bytes sem chamada ao provedor

#### Scenario: R2-ADM-09-S02 — UNKNOWN externo

- GIVEN não há prova de ausência de efeito
- WHEN operador abre ações
- THEN UI permite diagnóstico/reconciliação elegível e explica por que novo submit está bloqueado

#### Scenario: R2-ADM-09-S03 — Leitor administrativo

- GIVEN hub_protocol_reader abre protocolos de vários clientes
- WHEN tenta comando de replay por chamada manual
- THEN servidor nega; possuir leitura global não concede operação

### Requirement: R2-ADM-10 — SLA bilateral e painéis operacionais consultáveis

Console SHALL oferecer consulta de SLA Hub–provedor e cliente–Hub com coorte, contrato/versão, elegíveis, excluídos, abertos no prazo, vencidos, cumpridos, expirados, percentis e excedente. SHALL indicar atualização dos dados, linkar evidências permitidas e separar tempo do provedor, processamento do Hub e entrega externa.

Baseline relacionada: OPE-06, OPE-10, CFG-03.

#### Scenario: R2-ADM-10-S01 — Apuração de quebra

- GIVEN coorte contém lentos ainda abertos e finais rápidos
- WHEN analista filtra contrato/período
- THEN denominador e estados impedem viés dos concluídos; atraso aparece com trilha do prazo

#### Scenario: R2-ADM-10-S02 — Dados desatualizados

- GIVEN projeção operacional está atrasada
- WHEN painel é aberto
- THEN mostra horário/watermark e atraso, não anuncia ausência de violações

#### Scenario: R2-ADM-10-S03 — Consulta limitada

- GIVEN cliente/operador tem escopo A
- WHEN exporta relatório de SLA
- THEN arquivo contém somente registros permitidos e exportação é auditada

### Requirement: R2-ADM-11 — Financeiro de compra, venda e conciliação

Console SHALL separar planos de venda, contratos de compra, consumo, reservas/holds, extratos/ledger, ajustes, fechamento e exportação. Valores SHALL ser exibidos com moeda, precisão e versão contratual, sem aritmética financeira autoritativa no navegador. Ajustes SHALL requerer papel próprio e razão; parcelas não conciliadas SHALL ser visíveis.

Baseline relacionada: FIN-01, FIN-02, FIN-06, FIN-07, FIN-08, FIN-10.

#### Scenario: R2-ADM-11-S01 — Extrato explicável

- GIVEN protocolo gerou receita, custo e reserva
- WHEN analista abre origem dos valores
- THEN vê contrato/medidor/unidade, saldo antes/depois e lançamentos, distinguindo custo de cliente direto

#### Scenario: R2-ADM-11-S02 — Fechamento incompleto

- GIVEN há fatos pendentes ou UNKNOWN financeiros
- WHEN financeiro solicita fechamento
- THEN UI mostra bloqueios e não marca período como conciliado/exportado

#### Scenario: R2-ADM-11-S03 — Estorno

- GIVEN usuário autorizado corrige cobrança
- WHEN registra ajuste com razão e origem
- THEN lançamento compensatório aparece no histórico; original não desaparece

### Requirement: R2-ADM-12 — Qualidade de uso, acessibilidade e integração real

Jornadas SHALL ser utilizáveis por teclado, com rótulos, foco previsível, erros associados a campos e mensagens de estado anunciadas, e SHALL suportar viewport de 360 px e desktop com tabelas navegáveis. A conformidade proposta é WCAG 2.2 AA nas jornadas entregues. UI SHALL validar limites coerentes com servidor, preservar entrada em falha e separar sucesso de rascunho/publicação. Aceite NÃO SHALL usar mocks como evidência de integração real.

Baseline relacionada: CFG-01, QUA-03, QUA-04.

#### Scenario: R2-ADM-12-S01 — Teclado e erro

- GIVEN operador usa somente teclado e envia formulário inválido
- WHEN validação falha
- THEN foco alcança resumo/primeiro erro, campos são identificados e dados preenchidos permanecem

#### Scenario: R2-ADM-12-S02 — Falha de rede ou 409

- GIVEN editor perde rede ou versão ficou obsoleta
- WHEN tenta salvar
- THEN recebe estado recuperável sem duplo envio, perda de entrada ou falso sucesso

#### Scenario: R2-ADM-12-S03 — Jornada integrada

- GIVEN build passou e backend real de ensaio está disponível
- WHEN QA cria, publica, consulta e revisita recurso após refresh
- THEN estado persiste no servidor; evidência inclui papel, SHA, requests e screenshot sanitizado, não só renderização
