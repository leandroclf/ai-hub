# Delta for 04-console-e-autorizacao

## ADDED Requirements

### Requirement: R3-ADM-01 — Fronteira administrativa e aplicação
O Hub SHALL distinguir consumidor de operador e aplicar escopo por tenant/aplicação em listagens, detalhe e payload. Leitura entre tenants exige identidade nominal autorizada, MFA, auditoria e mascaramento; desenvolvedor não usa conta compartilhada nem herda escrita financeira.

#### Scenario: R3-ADM-01-S01 — operador nominal global com MFA
- GIVEN operador nominal global com MFA
- WHEN consultar protocolo de outro tenant
- THEN leitura autorizada é auditada e mascarada

#### Scenario: R3-ADM-01-S02 — token consumidor aplicação A
- GIVEN token consumidor aplicação A
- WHEN chamar admin ou consultar B
- THEN acesso é negado sem revelar payload/existência

#### Scenario: R3-ADM-01-S03 — papel revogado durante paginação
- GIVEN papel revogado durante paginação
- WHEN repetir consulta
- THEN backend revalida autorização sem depender da UI

### Requirement: R3-ADM-02 — Ações operacionais com API efetiva
O console SHALL usar contratos validados por operação, delivery_id para entregas e endpoints efetivos para reconciliação e SLA. Ação não implementada não pode simular disponibilidade. Redelivery repete bytes da entrega e não executa novamente provedor.

#### Scenario: R3-ADM-02-S01 — duas entregas do mesmo protocolo
- GIVEN duas entregas do mesmo protocolo
- WHEN abrir cada detalhe
- THEN IDs distintos selecionam recursos corretos

#### Scenario: R3-ADM-02-S02 — API retorna 401/403/409/503 ou JSON inválido
- GIVEN API retorna 401/403/409/503 ou JSON inválido
- WHEN acionar comando
- THEN UI mantém contexto e mostra erro sem falso sucesso

#### Scenario: R3-ADM-02-S03 — operador reconcilia e consulta SLA
- GIVEN operador reconcilia e consulta SLA
- WHEN executar jornada real
- THEN rotas funcionam com autorização e timeline auditável

### Requirement: R3-ADM-03 — Comandos financeiros compatíveis
O console SHALL emitir contrato financeiro válido, propagar tenant autorizado, converter períodos com fuso explícito e preservar chave da intenção em retries. Autor deriva da identidade; aprovação, disputa e exportação devem completar jornadas com segregação.

#### Scenario: R3-ADM-03-S01 — operador autorizado seleciona tenant B
- GIVEN operador autorizado seleciona tenant B
- WHEN preparar ajuste/período
- THEN comando persiste no tenant correto com ator autenticado

#### Scenario: R3-ADM-03-S02 — resposta se perde após commit
- GIVEN resposta se perde após commit
- WHEN repetir intenção
- THEN mesmo recurso é recuperado sem segundo ajuste

#### Scenario: R3-ADM-03-S03 — autor tenta autoaprovação ou altera tenant
- GIVEN autor tenta autoaprovação ou altera tenant
- WHEN executar
- THEN API recusa e UI não contorna segregação

### Requirement: R3-ADM-04 — Formulários completos e tempo local
O console SHALL preservar multimodalidade, oferecer campos condicionais de autenticação, lookup pesquisável/paginado e conversão correta de fuso. Respostas devem ser validadas em runtime e jornadas acessíveis por teclado com erros vinculados aos campos.

#### Scenario: R3-ADM-04-S01 — oferta SYNC+ASYNC+AUTO e contas OAuth/API Key/mTLS
- GIVEN oferta SYNC+ASYNC+AUTO e contas OAuth/API Key/mTLS
- WHEN editar/salvar/reabrir
- THEN modos e campos preservados sem expor segredo

#### Scenario: R3-ADM-04-S02 — UTC aberto em America/Sao_Paulo e referência além da centena
- GIVEN UTC aberto em America/Sao_Paulo e referência além da centena
- WHEN editar e pesquisar
- THEN instante não muda e referência é selecionável

#### Scenario: R3-ADM-04-S03 — JSON inválido em 200 ou erro de campo
- GIVEN JSON inválido em 200 ou erro de campo
- WHEN salvar
- THEN não há falso sucesso e foco/dados pendentes são preservados
