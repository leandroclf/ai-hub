# Design: Identidade, isolamento de tenants e administração segura

## Context

Baseline `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. [Explore](explore.md) distingue fatos de premissas; [spec](specs/identidade-e-isolamento-operacional/spec.md) é o contrato verificável. A solução abaixo é proposta de engenharia, não código implementado.

## Goals and Constraints

### Goals

Fechar a cadeia identidade → autorização por recurso → auditoria, da borda até os domínios internos.

### Constraints

Cinco autoridades existentes, Go no backend, TypeScript/React na UI; Redis opcional; SYNC direto; UUIDv7/custódia; nenhuma alteração retroativa de publicado ou efeito externo por replay inseguro. Gates comerciais/infra permanecem explícitos.

## Proposed Architecture

Fechar a cadeia identidade → autorização por recurso → auditoria, da borda até os domínios internos. Componentes e razão de tecnologia/linguagem constam em [arquitetura](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md), parte integrante deste design. Não duplicar autoridade em cache, gateway ou frontend.

## Technical Decisions

### Decision 1: Autorizar em cada fronteira

Decision: Kong autentica a borda e serviços validam identidade/escopo do recurso; wrappers compartilhados em Go reduzem divergência sem retirar a decisão do domínio.

Rationale: Fecha confiança no header e acesso apenas pela rede.

Trade-offs: Mais verificações locais, compensadas por JWKS/políticas válidas em cache; não chamar IdP remotamente em toda consulta.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 2: Leitor global separado de operador

Decision: Rota administrativa auditada com hub_protocol_reader individual, MFA e dados mascarados; grants específicos para comandos.

Rationale: Atende desenvolvedores sem conceder poderes a clientes ou usar conta compartilhada.

Trade-offs: Auditoria dessa leitura é obrigatória; se indisponível, negar essa ação em vez de abrir acesso irrestrito.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.

### Decision 3: Sessão de frontend

Decision: OIDC Authorization Code + PKCE para SPA e token curto em memória; modelo cookie somente se aprovado e protegido contra CSRF.

Rationale: Evita client_secret e tokens persistidos no navegador.

Trade-offs: Depende de compatibilidade/edição gateway/IdP em T-R2-03; teste de logout e revogação é parte do aceite.

Consequences: refletir nos contratos, migrations, métricas e cenários desta change; não declarar a decisão qualificada antes de executar a prova.


## Alternatives Considered

Manter somente os placeholders atuais é insuficiente para os cenários e riscos de [risk-matrix.md](risk-matrix.md). Troca geral de linguagem/broker/orquestrador foi descartada nesta rodada por não corrigir as invariantes por si só. Alternativas específicas e T-R2-01 a T-R2-05 estão no documento arquitetural, com trade-offs e provas exigidas.

## Affected Components

| Arquivo/componente existente | Mudança e motivo |
|---|---|
| hub/deploy/kong/kong.yml | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/platform/httpserver/httpserver.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/orbita/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/atlas/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/libra/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |
| hub/internal/pulsar/handlers.go | Evoluir dentro da autoridade definida; ver tarefa 2.x vinculada ao requisito e seus cenários. |

Arquivos novos sugeridos nas tarefas são marcados como propostos. Os diretórios acima existem no snapshot; migrations novas não substituem arquivos já aplicados.

## Main Flows

1. Autenticar na borda e limpar cabeçalhos de identidade fornecidos pelo cliente.
2. Resolver grant/política válida e contexto de tenant/aplicação/célula.
3. Autorizar recurso no domínio e registrar trilha exigida.
4. Executar ação permitida ou responder erro sanitizado e auditável.

## Error Flows

- R2-SEG-01: Token inválido — o Hub responde 401 antes de qualquer efeito; token válido sem escopo recebe 403.
- R2-SEG-02: Pool reutilizado — políticas e contexto transacional impedem acesso residual a A.
- R2-SEG-03: Auditoria indisponível — essa leitura é negada de forma controlada sem interromper consultas públicas elegíveis.
- R2-SEG-04: Rede privada legítima — a permissão é limitada ao destino/porta/identidade aprovados, sem exceção global a redes privadas.
- R2-SEG-05: Falha de backend — a resposta contém código tratável e correlação, sem DSN, token, stack ou segredo.

## API / Contract Design

Usar os contratos detalhados em [dados e estados](../../../docs/reviews/2026-09-07-r2/05-contratos-dados-e-estados.md) e [console/API](../../../docs/reviews/2026-09-07-r2/04-console-jornadas-e-api.md). Preservar versão pública v1 por perfil durante a migração; novos comandos têm identidade/idempotência e erros estáveis. OpenAPI/AsyncAPI devem acompanhar a implementação e ser validados contra consumidores reais. Nenhum endpoint administrativo proposto é apresentado como já existente.

## Data Model and Persistence

Principais administrativos individuais, grants por aplicação/tenant, versões de política e trilha append-only de decisões de acesso; credenciais nunca na trilha.

Campos, chaves, consultas/índices e autoridades estão na tabela do modelo lógico compartilhado. Persistência do domínio confirma estado e intenção/inbox na mesma transação quando exigido; nenhuma chamada de rede externa fica dentro de transação aberta. Escritas/joins entre autoridades são proibidos; projeções e IDs conservam contexto.

## Authentication and Authorization

Aplicar R2-SEG-01 a R2-SEG-05 a cada superfície envolvida. Escopo de tenant/aplicação é autenticado; privilégio administrativo individual não se transforma em acesso público global. Segredos e tokens não aparecem em retorno, evento ou logs. Papéis de aplicação/migração são distintos.

## Security and Privacy

Validar URLs e conexão efetiva, limites de tamanho/complexidade e dados permitidos. Auditoria durável de ações sensíveis com mascaramento e classe de retenção. Projeção ou cache inválido não amplia autorização. Usuário e workload só acessam recursos próprios da concessão/célula.

## Observability

### Logs

Eventos estruturados com correlação e estado/causa; não corpo sensível/segredo. Registrar falhas sem perder a obrigação de negócio.

### Metrics

Instrumentar as decisões/tempos relevantes aos cenários: contagem de recusas/erros, idade de obrigação, conflitos/duplicatas, execução/final/entrega ou mudança administrativa pertinente. Usar rota normalizada e dimensões de cardinalidade orçada, sem UUID como label.

### Tracing

Propagar contexto de trace e causation/event/command correlacionados; protocolo permanece identidade de negócio. Traces não são ledger nem prova única de custódia.

### Alerts

Alertar violação de invariante, obrigação envelhecida e dependência bloqueadora com runbook; acionar ensaio para provar detecção. Detalhamento em [operação](../../../docs/reviews/2026-09-07-r2/06-operacao-escala-e-observabilidade.md).

## Testing Strategy

Executar cenários `R2-SEG-01` a `R2-SEG-05` e integração pertinente da matriz IT-01 a IT-24. Unitários para decisões puras; DB/broker/HTTP reais de ensaio para transação e identidade; concorrência multiworker; contrato e E2E da UI quando afetada. Evidência precisa comparar esperado/obtido e apontar SHA. A suíte atual verde é baseline de compilação, não prova desses requisitos novos.

## Migration Strategy

Criar rotas e identidades novas; migrar consumidores por perfil autenticado. O header legado pode continuar como metadado de correlação, mas nunca como autoridade. Isolar explicitamente o simulador local.

## Rollback Strategy

Rollback mantém autenticação e autorização; não voltar à confiança no header. Reverter política para versão assinada anterior ainda válida; sessão revogada não ressuscita.

## Compatibility

Manter corpos finais já entregues e perfis v1, chaves de idempotência, snapshots e responsabilidades econômicas. Corrigir segurança não significa manter bypass inseguro. Implementação consumidora nova só é habilitada quando contratos requeridos estiverem disponíveis; coexistência é testada e limitada por janela explícita.

## Operational Considerations

Segurança e Engenharia de Plataforma responde pela implementação e prova; SRE/Qualidade revisam gates pertinentes. Dependências: Sem pré-requisito de implementação para iniciar; integrações específicas descritas abaixo. Laboratório e ambientes usam a mesma semântica com dados/identidades isolados. Não executar DDL/IaC remoto ou pagamento a partir desta documentação.

## Remaining Risks

Os achados desta change permanecem abertos até implementação/ensaio. Custódia total sem writer, ausência de impacto sob recursos ilimitadamente compartilhados e RPO zero regional sem topologia adequada não são garantias válidas.

## Open Questions

P-09 define responsáveis e perfis nominais; P-04 define dados permitidos. Essas pendências não impedem eliminar confiança em headers arbitrários. Questões técnicas adicionais T-R2-* seguem responsáveis e prova especificados; não enfraquecer cenário para encerrá-las.
