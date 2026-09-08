# Checkpoint catálogo e console R2

SHA base/HEAD inicial: a39d394b0d87185ed4cc3861c12ec45f2c302d9e, branch main. Alterações locais prévias preservadas. Responsabilidade: agente catalog_console; raiz integra runtime e identidade.

Leitura realizada: instruções AGENTS/CLAUDE, prompt integral, metodologia, especificações CAT/ADM, tasks e desenho 04/05, contratos/jornadas e baseline catálogo/segurança. Achados F17–22 confirmados no código inicial.

Em implementação: migration aditiva 0020_catalog_versions.sql (recursos por versão, publicação/auditoria e staging, trigger imutável para legado) e catalog.go (validação limitada de schemas/DAG/contratos e persistência com revisão). Código ainda sem qualificação; nenhum cenário marcado aprovado.

Interfaces acordadas: auth.FromContext e auth.Authorize(ctx,scope,tenant), DTO DAG enviado à raiz que implementa execução durável. Admin /admin/v1/{kind}, detalhe /{id}/{version}, ações /validate,/publish,/suspend,/simulate. UI OIDC Authorization Code PKCE Keycloak realm ai-hub-r2, cliente ai-hub-admin. Raiz fornecerá me, protocolos, timeline, SLA, Pulsar entregas. Agente finance_objects fornece /admin/v1/finance/{accounts,facts,journal,periods,exports,adjustments,disputes}.

Próximas ações: handlers com auth e paginação vinculada ao escopo; snapshot oferta + mapeamentos tipados; staging sanitizado; UI URL/lista/edição/jornadas; PostgreSQL integração e navegador real. Não usar mocks como evidência.

Bloqueios externos: P02/P03/P05/P09 continuam sem aprovação comercial/identidades reais. Capacidade/qualificação sandbox usam fixtures declaradas. Não há bloqueio técnico confirmado que autorize abandonar implementação.
