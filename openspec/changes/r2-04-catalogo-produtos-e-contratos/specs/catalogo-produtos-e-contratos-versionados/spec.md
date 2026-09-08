# Delta for catalogo-produtos-e-contratos-versionados

## ADDED Requirements

### Requirement: R2-CAT-01 — Catálogo versionado com publicação governada

Serviços e produtos SHALL ter identidade, versão, schemas, modalidades, classificação de dados e ciclo de vida explícitos. Versão publicada SHALL ser imutável. Publicação SHALL exigir validação, diff, compatibilidade e permissões; rascunhos SHALL usar controle de concorrência. Suspensão SHALL impedir novas admissões preservando protocolos e snapshots existentes.

Baseline relacionada: CAT-01, CAT-02, CFG-02, CFG-04.

#### Scenario: R2-CAT-01-S01 — Publicação válida

- GIVEN rascunho completo passou validações e aprovação exigida
- WHEN editor publica
- THEN versão/hash e autor ficam registrados e projeções passam a indicar a publicação

#### Scenario: R2-CAT-01-S02 — Sobrescrita de publicado

- GIVEN versão publicada já tem consumidores
- WHEN POST ou edição tenta alterar seu conteúdo
- THEN servidor retorna conflito e orienta nova versão; UI não anuncia alteração como publicada

#### Scenario: R2-CAT-01-S03 — Edição concorrente

- GIVEN dois operadores abriram mesma revisão
- WHEN o segundo salva após o primeiro
- THEN servidor detecta versão desatualizada e oferece diff/recarregamento sem sobrescrever trabalho alheio

### Requirement: R2-CAT-02 — Ofertas por cliente e roteamento autorizado

Cliente/aplicação SHALL acessar apenas ofertas vigentes que vinculam serviço/produto, perfil técnico, compra/venda, credencial, capacidade e SLA compatíveis. Hub SHALL escolher provedor por elegibilidade e política publicada. provider_account_id do contrato legado NÃO SHALL permitir escapar da oferta. Failover SHALL exigir equivalência e segurança de efeito comprovadas.

Baseline relacionada: CAT-02, CAT-06, CAT-10, CAT-11, FIN-02.

#### Scenario: R2-CAT-02-S01 — Seleção elegível

- GIVEN oferta tem dois provedores equivalentes homologados
- WHEN entra pedido de cliente autorizado
- THEN rota seleciona conta elegível com motivo e versão registrados

#### Scenario: R2-CAT-02-S02 — Conta escolhida pelo cliente

- GIVEN cliente fornece conta fora da sua oferta
- WHEN admissão avalia o pedido
- THEN nega antes do efeito, mesmo que a conta exista e possua credencial compartilhada

#### Scenario: R2-CAT-02-S03 — Sem combinação de SLA viável

- GIVEN SLA, modalidade, credencial ou capacidade contradiz perfil homologado
- WHEN operador tenta publicar oferta
- THEN recebe bloqueios específicos e oferta não é ativada

### Requirement: R2-CAT-03 — Agregação paralela limitada

Produto agregado SHALL executar passos independentes em paralelo dentro de quotas e orçamento de prazo, conservando identidade e resultado de cada passo. Política versionada SHALL definir obrigatoriedade, coleta, parcialidade, consolidação e custos; atraso de um passo NÃO SHALL serializar indevidamente os demais.

Baseline relacionada: CAT-03, CAT-05, CAT-08, EXE-13.

#### Scenario: R2-CAT-03-S01 — A e B independentes

- GIVEN produto agrega A e B sem dependência
- WHEN operação é executada sob capacidade reservada
- THEN traces mostram sobreposição e final combina respostas reais dos dois passos

#### Scenario: R2-CAT-03-S02 — Passo opcional falha

- GIVEN B é opcional e contrato autoriza parcialidade
- WHEN A conclui e B falha
- THEN final PARTIALLY_SUCCEEDED descreve falha conforme perfil e cobra somente marcos elegíveis

#### Scenario: R2-CAT-03-S03 — Fan-out abusivo

- GIVEN produto excede fan-out/concorrência autorizado
- WHEN é validado ou admitido
- THEN configuração/pedido é recusado antes de exceder recursos de outros clientes

### Requirement: R2-CAT-04 — Composição por DAG e compensações

Produto composto SHALL declarar DAG acíclico de passos versionados, dependências e mapeamentos seguros. Passos SHALL executar apenas com pré-condições duráveis satisfeitas. Falha SHALL aplicar política de interrupção/compensação; compensação SHALL ser operação de negócio rastreada, sem supor rollback externo completo. Deadline SHALL impedir novos passos produtivos.

Baseline relacionada: CAT-04, CAT-05, CAT-08, EXE-03, EXE-13.

#### Scenario: R2-CAT-04-S01 — Dependência C

- GIVEN A e B rodam em paralelo e C depende de ambos
- WHEN A termina antes de B
- THEN C só inicia após pré-condições duráveis de A e B; reinício retoma o grafo

#### Scenario: R2-CAT-04-S02 — Ciclo ou referência inválida

- GIVEN rascunho liga A a B e B a A ou usa saída inexistente
- WHEN publicação valida o grafo
- THEN publicação é bloqueada com os passos envolvidos

#### Scenario: R2-CAT-04-S03 — Compensação incompleta

- GIVEN A produziu efeito e passo obrigatório posterior falha
- WHEN compensação de A falha ou é impossível
- THEN protocolo informa estado contratado e gera obrigação de reconciliação, sem declarar rollback integral fictício

### Requirement: R2-CAT-05 — Perfis técnicos de entrada e saída por cliente

Oferta SHALL selecionar perfil técnico versionado por cliente/aplicação com schemas, media_type, tradução de entrada, final e erros. Transformações SHALL ser determinísticas, limitadas em CPU/memória/tamanho e sem acesso arbitrário a rede/segredos. Perfil publicado SHALL ser testável com amostras sanitizadas e preservado no protocolo.

Baseline relacionada: CAT-09, COM-02, COM-04, COM-05.

#### Scenario: R2-CAT-05-S01 — Dois contratos legados

- GIVEN A usa nomes/campos diferentes de B para o mesmo produto
- WHEN ambos executam pedidos equivalentes
- THEN normalização mantém semântica e cada cliente recebe seu perfil; GET e webhook do cliente têm mesmo corpo final

#### Scenario: R2-CAT-05-S02 — Transformação inválida

- GIVEN campo obrigatório não pode ser produzido ou expansão excede limite
- WHEN simulação/publicação ou execução valida contrato
- THEN erro é explícito e não expõe payload de outro cliente nem produz sucesso inválido

#### Scenario: R2-CAT-05-S03 — Upgrade de schema

- GIVEN perfil v2 remove campo usado em v1
- WHEN publicação avalia compatibilidade
- THEN mudança incompatível exige nova versão e migração; protocolos v1 conservam seu resultado

### Requirement: R2-CAT-06 — Snapshot e projeções independentes de consulta por pedido

Aceite SHALL fixar versões de produto/serviços/rota/adapter/perfis/contratos/prazos e referências de credencial/capacidade aplicáveis. Projeções locais SHALL ter versão, hash, validade e política de atualização/revogação. Mudança de controle NÃO SHALL reinterpretar protocolo antigo. Falha do Atlas SHALL permitir operar apenas com material ainda válido e autorizado.

Baseline relacionada: FIN-03, CFG-02, DAD-08, DAD-10, ARQ-02.

#### Scenario: R2-CAT-06-S01 — Contrato muda durante processamento

- GIVEN pedido foi aceito com compra/venda e perfil v1
- WHEN operador publica v2 antes do final
- THEN execução, resultado e financeiro usam snapshot v1

#### Scenario: R2-CAT-06-S02 — Atlas parado

- GIVEN workload possui projeção válida e cache limitado
- WHEN chegam novos pedidos elegíveis
- THEN não consulta Atlas a cada pedido e opera dentro da validade; itens não conhecidos não são inventados

#### Scenario: R2-CAT-06-S03 — Revogação de segurança

- GIVEN binding ou permissão é revogado
- WHEN há snapshot histórico e trabalho pendente
- THEN política de revogação de segurança prevalece para novos acessos ao segredo; preserva histórico sem reutilizar credencial proibida

### Requirement: R2-CAT-07 — Importação segura e inventário executável distinto

Importação de OpenAPI/collection SHALL criar lote em staging, sanitizar valores sensíveis, conservar proveniência e apresentar diff por identidade estável. NÃO SHALL apagar cadastros/contratos/vínculos alheios nem ativar endpoints automaticamente. Inventário importado e catálogo executável homologado SHALL ter estados e contagens distintos.

Baseline relacionada: CAT-01, CAT-02, CFG-02, COM-02, QUA-01.

#### Scenario: R2-CAT-07-S01 — Reimportação idempotente

- GIVEN mesmo arquivo sanitizado já foi importado
- WHEN operador reimporta
- THEN não duplica identidade nem apaga cadastros; mostra itens iguais, alterados e novos

#### Scenario: R2-CAT-07-S02 — Falha no meio do lote

- GIVEN lote contém erro ou execução interrompe
- WHEN importação é retomada
- THEN catálogo publicado anterior permanece íntegro e lote incompleto não se apresenta como ativado

#### Scenario: R2-CAT-07-S03 — Credenciais na collection

- GIVEN arquivo contém tokens, headers e environments sensíveis
- WHEN preview é solicitado
- THEN valores secretos não aparecem no lote, logs ou UI; somente referências e metadados permitidos são conservados

### Requirement: R2-CAT-08 — Consulta administrativa persistente e onboarding

Atlas SHALL disponibilizar listagens paginadas e filtros autorizados para clientes, aplicações, provedores, contas, serviços, produtos, ofertas, perfis e contratos. Onboarding SHALL registrar validação, capacidade/placement, pendências e estado de ativação; criação de cliente NÃO SHALL exigir novo deploy ou ticket de capacidade dentro do envelope aprovado.

Baseline relacionada: CFG-01, CFG-06, CAT-11, DAD-11.

#### Scenario: R2-CAT-08-S01 — Milhares de registros

- GIVEN operador tem escopo sobre catálogo grande
- WHEN filtra por estado/provedor/cliente com cursor
- THEN obtém página limitada, ordenação estável e contexto de autorização verificado no servidor

#### Scenario: R2-CAT-08-S02 — Refresh e outro operador

- GIVEN cadastro foi salvo e publicado
- WHEN outro operador abre sua URL após refresh
- THEN consulta fonte do servidor e visualiza mesma versão conforme permissão

#### Scenario: R2-CAT-08-S03 — Capacidade insuficiente

- GIVEN novo cliente excede headroom disponível
- WHEN onboarding solicita colocação
- THEN registra provisionamento automático e só ativa após qualificação; não aponta tráfego para célula incompleta
