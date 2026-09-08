# Delta for identidade-e-isolamento-operacional

## ADDED Requirements

### Requirement: R2-SEG-01 — Identidade autenticada e autorização por recurso

O Hub SHALL derivar tenant e aplicação de identidade autenticada e autorizada, validar emissor/audience/expiração/escopos e verificar autorização em cada domínio. Campos fornecidos pelo cliente NÃO SHALL ampliar seu tenant ou portfólio. Recursos alheios e inexistentes SHALL produzir resposta indistinguível ao cliente, sem consulta externa ao provedor.

Baseline relacionada: SEG-01, SEG-03, EXE-07, EXE-16.

#### Scenario: R2-SEG-01-S01 — Cliente autorizado

- GIVEN token válido de A e oferta de A
- WHEN A admite e consulta seu protocolo
- THEN o Hub autoriza somente a aplicação e os recursos concedidos

#### Scenario: R2-SEG-01-S02 — Tentativa entre tenants

- GIVEN token de A e protocolo de B
- WHEN A altera X-Tenant-Id, tenant_id, filtros ou cursor para B
- THEN o Hub não retorna dados de B e registra a negação sem revelar sua existência

#### Scenario: R2-SEG-01-S03 — Token inválido

- GIVEN token expirado, emissor ou audience incorretos
- WHEN uma API pública ou administrativa é chamada
- THEN o Hub responde 401 antes de qualquer efeito; token válido sem escopo recebe 403

### Requirement: R2-SEG-02 — Workloads com menor privilégio

APIs internas SHALL exigir identidade de workload e escopo restrito ao domínio/ação/célula. Acesso de rede NÃO SHALL conferir permissão. Papéis de aplicação SHALL diferir de migração e administração; controles por tenant SHALL resistir à reutilização de conexão.

Baseline relacionada: SEG-01, SEG-03, DAD-01, DAD-07.

#### Scenario: R2-SEG-02-S01 — Chamada interna permitida

- GIVEN Órbita autenticada com escopo de despacho da célula A
- WHEN chama Cometa na mesma célula
- THEN Cometa valida identidade e contexto antes de aceitar comando

#### Scenario: R2-SEG-02-S02 — Bypass pela porta interna

- GIVEN cliente ou workload sem escopo conhece rota internal
- WHEN tenta ler protocolos ou reservar saldo
- THEN a ação é negada mesmo dentro da rede do cluster

#### Scenario: R2-SEG-02-S03 — Pool reutilizado

- GIVEN conexão anteriormente usada por A é devolvida ao pool
- WHEN a aplicação atende B nessa conexão
- THEN políticas e contexto transacional impedem acesso residual a A

### Requirement: R2-SEG-03 — Leitura administrativa individual entre tenants

O perfil hub_protocol_reader SHALL permitir leitura administrativa entre tenants por usuário individual com MFA, concessão explícita e auditoria de finalidade. Esse perfil NÃO SHALL habilitar escrita, reprocessamento, alteração de resultado ou leitura de segredos. Consulta pública SHALL continuar restrita ao tenant mesmo para esse usuário.

Baseline relacionada: SEG-04, CFG-03.

#### Scenario: R2-SEG-03-S01 — Diagnóstico de desenvolvedor

- GIVEN usuário individual com MFA e hub_protocol_reader
- WHEN informa protocolo e justificativa na consulta administrativa
- THEN recebe visão permitida e mascarada com auditoria de autor, tenant, finalidade e horário

#### Scenario: R2-SEG-03-S02 — Sem elevação implícita

- GIVEN usuário só com hub_protocol_reader
- WHEN tenta publicar contrato, replay ou consultar segredo
- THEN o servidor nega a ação e o console não oferece permissão de escrita

#### Scenario: R2-SEG-03-S03 — Auditoria indisponível

- GIVEN leitura administrativa de dados sensíveis exige auditoria durável
- WHEN a trilha não pode ser registrada
- THEN essa leitura é negada de forma controlada sem interromper consultas públicas elegíveis

### Requirement: R2-SEG-04 — Destinos externos e callbacks protegidos

Destinos de provedor, token, upload remoto e webhook SHALL obedecer política de rede e TLS no cadastro e na conexão efetiva, incluindo resolução DNS, redirects e certificados. Endpoints privados SHALL exigir perfil explícito aprovado. Credenciais NÃO SHALL ser encaminhadas a outra origem por redirect.

Baseline relacionada: SEG-02, COM-02, CFG-05.

#### Scenario: R2-SEG-04-S01 — Destino homologado

- GIVEN endpoint HTTPS no perfil autorizado
- WHEN é validado e utilizado
- THEN TLS e destino efetivo são verificados e a política aplicada fica rastreável

#### Scenario: R2-SEG-04-S02 — SSRF e rebinding

- GIVEN URL aponta ou passa a resolver para rede não autorizada/metadados
- WHEN o envio é tentado ou redirecionado
- THEN o Hub bloqueia antes da conexão não permitida e não expõe segredo

#### Scenario: R2-SEG-04-S03 — Rede privada legítima

- GIVEN provedor privado tem perfil de rede autorizado
- WHEN o conector o acessa
- THEN a permissão é limitada ao destino/porta/identidade aprovados, sem exceção global a redes privadas

### Requirement: R2-SEG-05 — Sessão administrativa e trilha de alterações

Toda ação administrativa SHALL exigir sessão válida, permissão de domínio e registro durável de autor, alvo, versão anterior/nova, resultado e razão quando exigida. UI, logs e erros NÃO SHALL exibir segredos, tokens ou dados além da permissão. Encerramento de sessão SHALL limpar estado sensível.

Baseline relacionada: CFG-01, CFG-02, SEG-01, SEG-03.

#### Scenario: R2-SEG-05-S01 — Alteração atribuível

- GIVEN editor autorizado publica versão validada
- WHEN a alteração é confirmada
- THEN a trilha permite identificar autor, diff mascarado e versão publicada

#### Scenario: R2-SEG-05-S02 — Sessão termina

- GIVEN página tem dados sensíveis carregados
- WHEN sessão expira, usuário sai ou identidade muda
- THEN dados são removidos do estado da UI e novas ações exigem autenticação

#### Scenario: R2-SEG-05-S03 — Falha de backend

- GIVEN banco ou cofre retorna detalhe interno
- WHEN UI recebe erro
- THEN a resposta contém código tratável e correlação, sem DSN, token, stack ou segredo
