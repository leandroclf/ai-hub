# Execução R2

Estado: execução em andamento; nenhuma conclusão global. Base e HEAD inicial: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`, branch inicial `main`, trabalho em `r2-implementation`.

O snapshot é o mesmo da revisão. Foram preservados o diff anterior em `hub/evidence/EVIDENCE.md`, a documentação R2, suas nove changes, o resultado antispoofing e `preview.html` não rastreados. Relatório e matriz históricos não serão reescritos.

| Frente | Responsável de execução | Dependências | Implementação e prova |
| --- | --- | --- | --- |
| R2-09 | principal + operations_quality | desde o início | Inventário, fixtures, matriz por cenário; gates G1–G6 integrados ao final; G7 externo quando aplicável |
| R2-01 | principal | contratos de identidade | JWT/OIDC, workload, tenant/recurso, MFA, auditoria, SSRF; testes negativos e DB real |
| R2-02 | principal | 01, snapshot 04 | Aceite/intenção, posse, resposta, final, deadline e recuperação; crash/concorrência/ACK com PostgreSQL/SQS |
| R2-04 | catalog_console | 01, interface 02 | Atlas versionado/ofertas/contratos e DAG; root integra scheduler em Órbita |
| R2-03 | principal | 01/02/04 | Cofre/binding, adapter, polling/callback, quotas e entregas; identidade externa e falhas sintéticas |
| R2-05 | catalog_console | APIs proprietárias | OIDC PKCE, listas/detalhes/edição/publicação e operações reais; browser e confirmação no backend |
| R2-07 | finance_objects | 01/02 | FileRef/streaming/retention/restore; root integra admissão e resposta |
| R2-06 | finance_objects | 01/02/04 | Decimal exato, snapshot econômico, reservas/holds/journal/reconciliação; concorrência real |
| R2-08 | operations_quality | laboratório cedo, runtime integrado ao final | Compose isolado, kind/controladores, observabilidade, carga/recuperação e runbooks |

Cada tarefa 2.x conserva a ligação à sua prova 3.x no OpenSpec. Uma tarefa só fecha após resultados concretos de todos os cenários pertinentes. `SCENARIO_RESULTS.csv` não atribui PASS por nome de teste. Matriz baseline e R2 é acompanhada em `REQUIREMENTS_STATUS.csv`.

Interfaces acordadas: `internal/platform/auth` fornece principal verificado e autorização por scope/tenant; SPA Authorization Code + PKCE com token em memória; Atlas entrega snapshot JSON versionado; fatos incluem `economic_snapshot` com montantes string decimal e identidade da unidade; FileRef é estável e autorização de URL é separada; APIs administrativas permanecem em cada autoridade.

Laboratório: projeto Compose `ai-hub-r2`, PostgreSQL 15432, LocalStack 14566, APIs 18080–18084, IdP 18085, UI 13000. A stack anterior `hub-local` permanece preservada. Dados, credenciais e limites locais são fixtures; não decisões comerciais.

Skill selecionada: `golang-pro`, catálogo AAS 15.3.0 (digest `sha256-23ef3d92463bcdbc0b3e1156514b10ac19c1e6af21ba87b889062bb78dab0640`). Sem override local. Aplicada a concorrência/erros/testes Go; DBOS não selecionada por exigir stack diferente. As specs do projeto governam os demais domínios. Ferramentas e versões efetivas serão registradas nas evidências.
