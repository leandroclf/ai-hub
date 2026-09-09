# Reavaliação individual dos 42 achados R2

Base: `main` em `f31389a` mais o working tree desta retomada. `PARCIAL` não é encerramento; significa que existe prova atual de uma subinvariante, mas ainda falta o cenário integral original. Nenhum achado é fechado por inferência.

| achado | estado atual | evidência atual | condição para encerramento |
|---|---|---|---|
| F-01 | PARCIAL | OIDC nominal/MFA/browser e auth unitário | prova de identidade em todas as rotas administrativas |
| F-02 | PARCIAL | workload auth e escopo HTTP Cometa | negar acesso cruzado em cada domínio com credencial runtime |
| F-03 | PARCIAL | callback inbox/capability e testes de custódia | callback real órfão/tardio após restart com sink externo |
| F-04 | PARCIAL | `command_intents` e recuperação direta | queda real entre handoff e confirmação |
| F-05 | PARCIAL | inbox/quarentena e consumer de fatos | bytes recuperáveis e ACK condicionado no broker real |
| F-06 | PARCIAL | provider-sim e transformação de output | adapter real do fluxo completo com hash externo |
| F-07 | PARCIAL | `PrepareSubmission`, epoch e concorrência | prova de efeito externo único sob crash simultâneo |
| F-08 | PARCIAL | `ConserveObservation` transacional | callback e polling no mesmo cenário de falha |
| F-09 | ABERTO | testes de deadline/fencing isolados | contraexemplo T-R2-01 com relógios/commit separados |
| F-10 | PARCIAL | TTL/agenda e Retry-After persistidos | ensaio temporal completo com primeira falha transitória |
| F-11 | PARCIAL | SYNC/AUTO/UUID e browser | orçamento, erro e consulta futura em cada modalidade |
| F-12 | PARCIAL | binding/secret version pinning | prova por tenant de que a credencial usada é a congelada |
| F-13 | PARCIAL | SigV4/OAuth fixture LocalStack | homologação de adapter e mTLS fora do simulador |
| F-14 | PARCIAL | SYNC direto sem fila obrigatória | Redis/broker/cofre degradados durante execução real |
| F-15 | PARCIAL | lease, epoch, auth e Retry-After | pressão adaptativa ligada a SUBMIT/STATUS/FETCH real |
| F-16 | PARCIAL | controlador de capacidade testado | grants/feedback observáveis em todas as chamadas externas |
| F-17 | PARCIAL | catálogo versionado e snapshot | política efetiva completa por cliente/aplicação/oferta/perfil |
| F-18 | PARCIAL | `PlanDAG` e executor bounded adicionados | integração do executor com produto/adapters e compensação durável |
| F-19 | PARCIAL | importação/catalogação validada | ativação real de adapter com rollback |
| F-20 | PARCIAL | console com navegação ampliada e browser | concluir todas as jornadas com backend real |
| F-21 | PARCIAL | páginas de operações/financeiro presentes | ações de reconciliação, entrega e financeiro ponta a ponta |
| F-22 | PARCIAL | browser responsivo e DTO corrigido | multimodalidade, timezone, lookups e validação runtime |
| F-23 | PARCIAL | snapshots econômicos e testes Libra | incidência por unidade/tentativa em fluxo externo |
| F-24 | PARCIAL | reserva estrita e UNKNOWN hold | captura pelo efetivo e franquia no caminho real |
| F-25 | PARCIAL | ledger imutável e exact decimal | fechamento, disputa, ajuste e exportação operacionais |
| F-26 | PARCIAL | destino versionado no Pulsar | destino congelado no aceite por aplicação/protocolo |
| F-27 | PARCIAL | claim/redelivery/custódia Pulsar | ciclo completo de retry com bytes e orçamento externo |
| F-28 | PARCIAL | representação persistida e readback | OutputMapping por aplicação no webhook real |
| F-29 | PARCIAL | FileRef, multipart e pins | FileRef consumido no adapter e resultado volumoso publicado |
| F-30 | PARCIAL | migration RLS runtime + retenção | DSNs dos serviços com contexto tenant e restore real |
| F-31 | PARCIAL | Compose oficial completo executado | recriação integrada e persistência de todos os componentes |
| F-32 | PARCIAL | Kind 3 nós/5 workloads/HPA | ambientes remotos e dependências completas fora do Compose |
| F-33 | ABERTO | HPA/KEDA declarados e observados | placement/IaC/nós/dados reconciliados sob escala |
| F-34 | PARCIAL | endpoints/credential chain configuráveis | bootstrap dev/hom/ppd/prd e contratos de ferramenta |
| F-35 | PARCIAL | probes e restart controlado | drenagem de workers/leases sob SIGTERM e saturação |
| F-36 | PARCIAL | Prometheus/Loki/Tempo/Grafana reais | SLA bilateral, pressão e idade de obrigações consultáveis |
| F-37 | ABERTO | subprovas de isolamento e reinício | HA, restore e falha regional com oráculos independentes |
| F-38 | PARCIAL | Go/race/vet e gate integrado sem skip | cobertura integral dos cenários obrigatórios v4/R2/R3 |
| F-39 | PARCIAL | relatório/checkpoint atualizado nesta retomada | matriz requisito→cenário→código→comando→evidência completa |
| F-40 | PARCIAL | versões pinned e OpenSpec 17/17 | compatibilidade operacional dos ambientes exigidos |
| F-41 | PARCIAL | egress/SSRF e destino validado | redirect/DNS/adapter e rotação qualificados integralmente |
| F-42 | PARCIAL | capacity/lease/limits locais | orçamento global por tenant e pool por chamada sob carga |

## Resultado

Nenhum dos 42 achados é promovido para `RESOLVIDO` nesta etapa. A lista deixa de ser um bloco genérico pendente e passa a registrar, individualmente, a prova atual e o critério objetivo restante.
