# Delta for operacao-elastica-e-observavel

## ADDED Requirements

### Requirement: R2-OPE-01 — Plataforma local completa e persistente

Ambiente local SHALL subir por procedimento versionado todos os componentes necessários: cinco aplicações, UI, gateway, identidade de ensaio, PostgreSQL por domínio, broker/objetos emulados, cofre de ensaio, cache opcional e observabilidade. Dados SHALL sobreviver à recriação documentada dos containers por volumes explícitos. Fixtures SHALL ser isoladas e nenhuma subida SHALL apagar dados por padrão.

Baseline relacionada: OPE-04, QUA-01, DAD-07.

#### Scenario: R2-OPE-01-S01 — Primeira instalação

- GIVEN máquina limpa possui pré-requisitos declarados
- WHEN executa bootstrap documentado
- THEN UI e APIs autenticadas, broker/objetos e painéis ficam acessíveis com verificações de saúde

#### Scenario: R2-OPE-01-S02 — Recriação sem reset

- GIVEN há protocolos e cadastros de ensaio persistidos
- WHEN containers são recriados no procedimento normal
- THEN dados e referências continuam consultáveis; reset destrutivo é ação separada explícita

#### Scenario: R2-OPE-01-S03 — Cache desligado

- GIVEN operador não habilita perfil Redis
- WHEN sobe plataforma e executa jornada padrão
- THEN serviços não falham por dependência do cache opcional

### Requirement: R2-OPE-02 — Kubernetes e cinco ambientes coerentes

Deploy SHALL definir local, dev, hom, ppd e prd e laboratório local-kind, com contratos de configuração consistentes e isolamento de dados/credenciais. kind SHALL ser laboratório de integração Kubernetes, não prova de HA física. Imagens, rotas, UI, identidade, dependências, migrations, probes e controllers SHALL existir de forma aplicável, sem placeholders ativos em promoção.

Baseline relacionada: OPE-04, OPE-05, ARQ-06, ARQ-01.

#### Scenario: R2-OPE-02-S01 — Overlay aplicável

- GIVEN dependências e CRDs da versão fixada foram instalados
- WHEN overlay de laboratório é renderizado e aplicado
- THEN pods ficam prontos, rotas funcionam e não há referência não resolvida

#### Scenario: R2-OPE-02-S02 — Separação ambiental

- GIVEN mesmo tenant existe em dev e hom
- WHEN configuração/credencial de dev tenta operar hom
- THEN isolamento impede cruzamento; promoção não copia segredos entre ambientes

#### Scenario: R2-OPE-02-S03 — Produção sem perfil

- GIVEN P-01/P-08/P-10 relevantes não estão qualificados
- WHEN pipeline tenta promover oferta crítica
- THEN gate bloqueia ativação e mostra pendências sem impedir desenvolvimento local

### Requirement: R2-OPE-03 — Probes por capacidade e drenagem

Readiness SHALL representar capacidades efetivas de cada papel (admissão, leitura, direto, worker), startup SHALL validar bootstrap e liveness SHALL depender de saúde local sem causar reinício por falha externa. SIGTERM SHALL retirar admissão, drenar trabalho e conservar incerteza/leases antes de terminar. Falha exclusiva de escrita NÃO SHALL derrubar leitura final comprovadamente elegível.

Baseline relacionada: OPE-05, OPE-13, OPE-14.

#### Scenario: R2-OPE-03-S01 — Dependência opcional falha

- GIVEN Redis ou telemetria fica indisponível
- WHEN probes são avaliadas
- THEN workloads elegíveis continuam prontos e falha é observável sem cascata de reinícios

#### Scenario: R2-OPE-03-S02 — Rolling update

- GIVEN há SYNC em voo, poll e delivery agendados
- WHEN pod recebe SIGTERM
- THEN deixa de admitir, drena dentro do grace e conserva obrigação não concluída para takeover cercado

#### Scenario: R2-OPE-03-S03 — Provedor fora

- GIVEN um parceiro responde timeout
- WHEN liveness e readiness geral são verificadas
- THEN não reinicia todos os pods; circuito/quotas contêm apenas a capacidade afetada

### Requirement: R2-OPE-04 — Escala de pods, nós e dados sem chamados rotineiros

Escala SHALL usar sinais de capacidade/latência e backlog/idade, um controlador por Deployment, piso aquecido para rotas síncronas e orçamento global de conexões. Provisionamento de nós/células SHALL antecipar horizonte de expansão e N−1 dentro de quotas aprovadas; ativação SHALL ocorrer somente após qualificação. Banco transacional NÃO SHALL ser tratado como replicável por HPA.

Baseline relacionada: OPE-12, OPE-08, CFG-06, DAD-11, ARQ-06.

#### Scenario: R2-OPE-04-S01 — Fila cresce com CPU baixa

- GIVEN trabalho I/O acumula e idade aumenta
- WHEN scaler avalia métricas
- THEN réplicas sobem respeitando quota externa, conexões e reserva de capacidade

#### Scenario: R2-OPE-04-S02 — Célula próxima do envelope

- GIVEN demanda prevista excede headroom antes do prazo de provisionamento
- WHEN política de expansão atua
- THEN nova célula é criada/qualificada e placement muda para novas admissões sem ticket rotineiro

#### Scenario: R2-OPE-04-S03 — Quota esgotada

- GIVEN expansão cloud não consegue provisionar recurso autorizado
- WHEN onboarding precisa de capacidade
- THEN estado registra impedimento e novas admissões são limitadas; não rouba reserva de clientes existentes

### Requirement: R2-OPE-05 — Placement, fencing e recursos por ambiente e célula

Roteamento SHALL fixar célula/epoch de execução no aceite e manter consulta de protocolos antigos após realocação do tenant. Filas, tópicos, dados e permissões SHALL estar alinhados ao ambiente/célula. Cada recurso SHALL ter um único controlador proprietário; Terraform e Crossplane NÃO SHALL reconciliar simultaneamente o mesmo objeto. Promoção de autoridade SHALL cercar o escritor anterior.

Baseline relacionada: DAD-11, ARQ-06, CFG-06, OPE-08.

#### Scenario: R2-OPE-05-S01 — Cliente realocado

- GIVEN protocolos antigos pertencem à célula A e novos à B
- WHEN cliente consulta ambos por UUID
- THEN diretório autorizado resolve a origem correta sem busca em provedores ou vazamento entre tenants

#### Scenario: R2-OPE-05-S02 — Dois ambientes/células

- GIVEN dev/hom e células A/B usam broker
- WHEN eventos são publicados
- THEN namespaces e políticas impedem colisão de nome/entrega cruzada

#### Scenario: R2-OPE-05-S03 — Falha parcial de provisionamento

- GIVEN Crossplane criou banco mas não políticas/filas
- WHEN Atlas avalia readiness da célula
- THEN não atribui tráfego até recursos e canário estarem qualificados; retry de reconcile não duplica dono

### Requirement: R2-OPE-06 — Fallback seletivo e dependências mínimas

Caminho SYNC e GET SHALL dispensar broker, Atlas e Redis por pedido quando projeções/autorização/material válidos e autoridades necessárias estiverem disponíveis. Bootstrap SHALL permitir operar essas capacidades sem criar recursos de broker. Ausência de writer, credencial válida ou reserva estrita necessária SHALL impedir apenas a operação dependente, com resposta explícita, nunca aceite em armazenamento alternativo sem autoridade.

Baseline relacionada: OPE-13, DAD-09, DAD-10, ARQ-02.

#### Scenario: R2-OPE-06-S01 — Reinício com broker fora

- GIVEN DB e projeções recuperáveis estão válidos
- WHEN Órbita/Cometa reiniciam durante indisponibilidade SNS/SQS
- THEN rotas SYNC e GET elegíveis sobem; fatos ficam recuperáveis sem bloquear servidor HTTP

#### Scenario: R2-OPE-06-S02 — Redis lenta

- GIVEN latência Redis aumenta acima do orçamento
- WHEN workload usa cache de metadados opcional
- THEN bypass rápido/circuito e cache L1 limitado mantêm o perfil qualificado

#### Scenario: R2-OPE-06-S03 — Todas cópias autoritativas indisponíveis

- GIVEN não há writer/quorum válido
- WHEN chega nova admissão
- THEN Hub informa indisponibilidade sem 202 fictício; preserva repetição pela mesma chave quando voltar

### Requirement: R2-OPE-07 — Telemetria e SLA com investigação acionável

Plataforma SHALL provisionar métricas, logs, traces, dashboards e alertas com Prometheus, Grafana, Loki/Alloy e OpenTelemetry/Tempo conforme referência. SHALL medir latência ponta a ponta e por etapa, lag/idade de obrigação, UNKNOWN, outbox/inbox/quarentena, pressão, pools, custódia, financeiro e SLA bilateral. IDs individuais NÃO SHALL ser labels de métricas; trilha de auditoria SHALL ter custódia própria.

Baseline relacionada: OPE-03, OPE-06, OPE-10.

#### Scenario: R2-OPE-07-S01 — Fluxo rastreável

- GIVEN requisição percorre SYNC ou ASYNC
- WHEN operador usa protocolo/trace autorizado
- THEN correlaciona etapas e logs, visualiza percentis e recebe links de evidência sem dados sensíveis nos labels

#### Scenario: R2-OPE-07-S02 — Alerta exercitado

- GIVEN outbox envelhece, DLQ recebe item ou SLA vence
- WHEN condição excede limiar do perfil
- THEN alerta aponta componente/runbook e teste registra detecção e recuperação

#### Scenario: R2-OPE-07-S03 — Backend de observabilidade fora

- GIVEN Loki/Tempo/Prometheus está temporariamente indisponível
- WHEN Hub recebe tráfego elegível
- THEN buffer/export limitado não bloqueia thread de negócio; perda de telemetria é sinalizada e não altera ledger/auditoria duráveis

### Requirement: R2-OPE-08 — Segurança e reprodutibilidade de infraestrutura

Ambiente remoto SHALL usar identidade de workload, segredo externo, TLS, políticas mínimas e recursos previamente provisionados. Aplicações NÃO SHALL criar filas/buckets ou sobrescrever políticas em bootstrap remoto. Build/deploy SHALL fixar versões/digests e separar configuração por ambiente, com migrations versionadas e validação de compatibilidade de controllers/CRDs.

Baseline relacionada: ARQ-05, SEG-01, OPE-04, QUA-04.

#### Scenario: R2-OPE-08-S01 — AWS real

- GIVEN workload possui IAM adequado e URLs/ARNs configurados
- WHEN sobe em ambiente remoto
- THEN usa cadeia de identidade autorizada sem credencial local/local nem SetQueueAttributes/CreateQueue

#### Scenario: R2-OPE-08-S02 — IAM insuficiente

- GIVEN papel não tem permissão para recurso de outra célula
- WHEN workload tenta acessá-lo
- THEN acesso é negado sem política permissiva de fallback

#### Scenario: R2-OPE-08-S03 — Atualização de dependência

- GIVEN imagem, runtime ou controller muda de versão
- WHEN pipeline valida promoção
- THEN registra suporte/licença, lock/digest, scan e ensaios de contrato; referência de versão não vira aprovação automática

### Requirement: R2-OPE-09 — SLO, isolamento e continuidade qualificados

Cada perfil SHALL declarar carga, payload/fan-out, concorrência, SLO, interrupção máxima, RTO/RPO por domínio e falhas cobertas. Cumprimento SHALL ser medido sob carga e manutenção; 429 por falta de capacidade do Hub e timeouts SHALL compor indisponibilidade conforme v4. Promessas de zero perda/impacto SHALL estar limitadas ao domínio demonstrado. Perfil crítico NÃO SHALL herdar automaticamente RTO geral.

Baseline relacionada: OPE-01, OPE-02, OPE-03, OPE-09, OPE-11, OPE-15.

#### Scenario: R2-OPE-09-S01 — Perfil de referência

- GIVEN ensaio usa JSON até 64 KiB e mix/fan-out/carga declarados
- WHEN mede admissão ASYNC e GET
- THEN compara às metas propostas v4: aceite p95≤250 ms/p99≤500 ms, GET p95≤100 ms/p99≤250 ms, registrando resultado sem declarar metas já atingidas

#### Scenario: R2-OPE-09-S02 — Manutenção e ruído

- GIVEN cliente B opera dentro da reserva e A excede quota
- WHEN ocorre rollout e falha de dependência no ensaio
- THEN SLIs incluem as falhas e B mantém limites do perfil; variação p95≤10% é alvo adicional proposto, sujeito a qualificação

#### Scenario: R2-OPE-09-S03 — Uso crítico

- GIVEN oferta pode afetar vida/segurança
- WHEN não há ensaio compatível com interrupção máxima e domínio de falhas
- THEN ativação é bloqueada para esse perfil, com contingência e pendências explicitadas
