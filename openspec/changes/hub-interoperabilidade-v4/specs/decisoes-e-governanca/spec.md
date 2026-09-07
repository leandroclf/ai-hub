# Decisões, Pendências e Referências — Delta de Especificação

## ADDED Requirements

### Requirement: DEC-01 — Decisões arquiteturais de referência
O processo de governança SHALL manter, sob propriedade de Arquitetura, um registro de decisões arquiteturais de referência (ADR-01 a ADR-10) que orientam a especificação v4, cada uma registrando a escolha adotada, a consequência resultante e a alternativa descartada com o motivo de não adoção. Este registro NÃO SHALL ser tratado como aprovação comercial nem como infraestrutura já existente. Conta, região, condições de dados, orçamento, licenciamento e modo gerenciado/autogerenciado SHALL permanecer fora do escopo destes ADRs e SHALL ser tratados como pendências de implantação distintas (DEC-02), de modo que Engenharia NÃO SHALL escolher tecnologia por ambiente silenciosamente.

#### Scenario: ADR-04 mantém a custódia do resultado no Hub
- GIVEN uma proposta de delegar a um provedor externo o papel de proxy do resultado final
- WHEN essa alternativa é avaliada contra ADR-04
- THEN o processo de governança SHALL manter a decisão de que o Hub conserva a versão final e atende ao GET local, e SHALL registrar que a alternativa de proxy no provedor foi descartada por violar a custódia e a independência solicitadas

#### Scenario: ADR-06 recusa promessa de exactly-once sem suporte externo
- GIVEN uma proposta de anunciar entrega exactly-once ponta a ponta sem suporte do provedor externo
- WHEN essa proposta é confrontada com ADR-06
- THEN o processo de governança SHALL manter outbox/inbox, chaves semânticas, reconciliação e o estado UNKNOWN como a decisão vigente, e SHALL registrar a promessa de exactly-once sem suporte externo como infundada e não adotada

#### Scenario: ADR-05 mantém deadline rígido e TTL sem renovação
- GIVEN uma proposta de permitir que um retorno tardio reabra um protocolo já expirado
- WHEN essa proposta é avaliada contra ADR-05
- THEN o processo de governança SHALL manter a máquina de estados durável com grafo limitado, deadline rígido e TTL sem renovação, e NÃO SHALL permitir que o retorno tardio reabra o protocolo expirado

#### Scenario: Pendências de implantação não decididas por ADR ficam registradas em DEC-02
- GIVEN uma decisão sobre conta, região, orçamento, licenciamento ou modo gerenciado/autogerenciado ainda não aprovada
- WHEN essa decisão é necessária para avançar a implantação
- THEN o processo de governança NÃO SHALL tratá-la como resolvida pelos ADR-01 a ADR-10, e SHALL exigir que ela seja registrada e aprovada como pendência de implantação distinta antes de Engenharia escolher a tecnologia do ambiente

### Requirement: DEC-02 — Registro de decisões pendentes (P-01 a P-11)
O processo de governança SHALL manter, sob liderança técnica, um registro de decisões pendentes (P-01 a P-11), cada uma com a informação necessária, o responsável final e a condição de bloqueio/critério de resolução explícitos. Cada pendência SHALL bloquear apenas o gate, a produção ou a oferta específica a que se refere, NÃO SHALL bloquear todo o trabalho de engenharia sintética; números de metas e retenção marcados como propostos MAY ser usados para projetar e ensaiar, mas NÃO SHALL substituir a decisão pendente correspondente.

#### Scenario: P-08 bloqueia a promessa de ausência de perda regional até desenho e prova de recuperação
- GIVEN a promessa de ausência de perda regional/prd depende do domínio de falhas, RPO zero contratado, confirmação entre regiões, RTO e capacidade por célula definidos em P-08
- WHEN essa promessa é avaliada antes da aprovação de P-08 pelo SRE
- THEN o processo de governança SHALL manter o gate de promessa de ausência de perda regional bloqueado enquanto o desenho e a prova de recuperação, o impacto de latência e o custo não estiverem aprovados

#### Scenario: P-10 bloqueia oferta com impacto em vida/segurança até perfil de criticidade aprovado
- GIVEN uma oferta de uso com impacto em vida ou segurança depende do perfil de criticidade, do responsável do sistema consumidor, da máxima interrupção em segundos e da contingência definidos em P-10
- WHEN essa oferta é avaliada antes da aprovação de P-10 por Produto
- THEN o processo de governança SHALL manter essa oferta bloqueada enquanto riscos, SLO/RPO/RTO e provas não estiverem aprovados

#### Scenario: P-11 bloqueia o gate ppd de expansão automática até envelopes e quotas aprovados
- GIVEN o gate ppd de expansão automática depende dos envelopes de escala, quotas preautorizadas, reserva, horizonte de provisionamento e política de células definidos em P-11
- WHEN esse gate é avaliado antes da aprovação de P-11 por Plataforma
- THEN o processo de governança SHALL manter o gate ppd de expansão automática bloqueado enquanto um cenário sem chamados, com limites e custos demonstrados, não estiver aprovado

#### Scenario: P-01 bloqueia IaC remota e o gate ppd até a topologia de plataforma ser aprovada
- GIVEN a decisão de conta, região, Kubernetes gerenciado ou infraestrutura própria, orçamento, registro OCI e licenças ainda não foi aprovada em P-01
- WHEN IaC remota ou o gate ppd são avaliados
- THEN o processo de governança SHALL manter IaC remota e o gate ppd bloqueados até que a aprovação registrada da topologia e do custo esteja concluída

#### Scenario: P-09 bloqueia o gate G0 da fatia até responsáveis e aprovação da v4 registrados
- GIVEN o gate G0 da fatia depende dos responsáveis, dos perfis administrativos dos desenvolvedores e da aprovação da v4 definidos em P-09
- WHEN o gate G0 é avaliado antes dessa aprovação pela liderança técnica
- THEN o processo de governança SHALL manter o gate G0 bloqueado até que o registro de decisão e os responsáveis estejam formalizados

### Requirement: DEC-03 — Decisões herdadas da v3 e preservadas
O processo de governança SHALL preservar integralmente as decisões ADR-11 a ADR-18 herdadas da v3, sob propriedade conjunta de Arquitetura e Produto, mantendo os IDs existentes para rastreabilidade. O processo de governança SHALL manter as correções introduzidas na v3 sobre a v2 — representação unificada (200) para consulta de resultado pendente, proibição de retorno tardio publicar revisão de sucesso após encerramento por SLA, segmentação de bancos e filas por célula, e a antiga meta regional de RPO de 15 minutos deixando de ser padrão implícito — tratando-as como mudanças de especificação, não como alegação de migração de sistema implementado.

#### Scenario: ADR-12 impede que retorno tardio após EXPIRED por SLA vire sucesso
- GIVEN um protocolo já encerrado como EXPIRED por SLA recebe posteriormente uma resposta tardia do provedor
- WHEN essa resposta tardia é processada, mesmo que o HTTP de recebimento tenha retornado 2xx
- THEN o processo de governança SHALL manter a regra de que esse HTTP 2xx de recebimento NÃO SHALL ser tratado como aceitação de negócio, e o recibo SHALL permanecer apenas como evidência e eventual custo, sem reabrir o protocolo como sucesso

#### Scenario: ADR-18 mantém o compromisso regional de RPO zero bloqueado até desenho adequado
- GIVEN o objetivo de ausência de perda está ligado ao modelo de falhas testado conforme ADR-18
- WHEN um compromisso regional de RPO zero é proposto sem esse desenho qualificado
- THEN o processo de governança SHALL manter esse compromisso bloqueado até que o desenho adequado exista, na mesma condição referenciada por P-08

#### Scenario: Consulta de resultado pendente usa representação unificada (200) herdada da correção v2→v3
- GIVEN um cliente consulta um protocolo ainda pendente
- WHEN a resposta é construída conforme a correção de compatibilidade preservada da v3
- THEN o processo de governança SHALL manter a representação unificada com HTTP 200 para esse caso, sem reintroduzir a ambiguidade anterior da v2

#### Scenario: A antiga meta regional de RPO de 15 minutos não é mais padrão implícito
- GIVEN uma referência à antiga meta regional de RPO de 15 minutos usada na v2
- WHEN essa meta é invocada como padrão para uma decisão atual
- THEN o processo de governança NÃO SHALL tratar essa meta como padrão implícito vigente, remetendo a decisão de RPO regional para P-08 e ADR-18

#### Scenario: IDs de ADR-11 a ADR-18 permanecem estáveis para rastreabilidade entre capítulos
- GIVEN os capítulos existentes referenciam ADR-11 a ADR-18 herdados da v3
- WHEN novos requisitos são adicionados na v4
- THEN o processo de governança SHALL manter os IDs herdados sem renumeração, atribuindo IDs próprios apenas aos requisitos novos

### Requirement: DEC-04 — Decisões e compatibilidade introduzidas na v4
O processo de governança SHALL registrar, sob propriedade de Arquitetura, as decisões ADR-19 a ADR-25 introduzidas na v4, incluindo a alteração da escolha de transporte e a retirada do default de oito segundos para todos os caminhos síncronos. O processo de governança SHALL preservar integralmente, junto dessas novas decisões, os requisitos de TTL em segundos, deadline não renovável, recusa tardia, aferição bilateral, custódia de resultado, agregação/composição, pressão adaptativa, contratos legados, equivalência GET/webhook, compra/venda e acesso administrativo individual, mantendo os IDs herdados e ampliando a matriz apenas com os novos IDs.

#### Scenario: ADR-19 retira a fila obrigatória do caminho SYNC sem fallback silencioso
- GIVEN uma execução SYNC direta entre Órbita e Cometa conforme ADR-19
- WHEN essa execução conclui dentro do caminho síncrono
- THEN o processo de governança SHALL manter a ausência de fila obrigatória nesse caminho, preservando estado, intenção, deadline e fatos, e NÃO SHALL permitir fallback silencioso de SYNC para 202

#### Scenario: ADR-20 recupera protocolo perdido por HTTP via idempotência com UUIDv7 universal
- GIVEN o protocolo foi persistido com UUIDv7 conforme ADR-20 antes de uma perda de resposta HTTP
- WHEN o cliente ou o sistema tenta recuperar esse protocolo
- THEN o processo de governança SHALL manter a recuperação via idempotência baseada nesse identificador persistido, e uma recusa anterior ao aceite NÃO SHALL virar protocolo fictício

#### Scenario: ADR-21 proíbe fallback implícito de credencial entre SHARED_HUB e TENANT_DEDICATED
- GIVEN uma operação exige resolução de credencial entre os perfis SHARED_HUB e TENANT_DEDICATED
- WHEN essa resolução ocorre com vínculo/conta/tenant e rotação definidos conforme ADR-21
- THEN o processo de governança SHALL manter a proibição de fallback implícito entre esses perfis, tratando a definição do pagador como regra contratual separada

#### Scenario: ADR-22 dispensa broker/Atlas/cofre por pedido quando material válido já existe
- GIVEN um SYNC simples cujas quatro fronteiras lógicas duráveis já possuem material válido carregado
- WHEN esse caminho é executado conforme ADR-22
- THEN o processo de governança SHALL manter L1/projeções e Redis como opcionais e NÃO SHALL exigir chamada a broker, Atlas ou cofre por pedido nessas condições, sem adotar escrita alternativa sem autoridade

#### Scenario: ADR-25 exige prova específica de RTO e perda de região para uso com impacto em vida/segurança
- GIVEN uma oferta com impacto em vida ou segurança avalia manutenção, RTO e perda de região conforme ADR-25
- WHEN percentual mensal de disponibilidade e réplica multi-AZ são apresentados como evidência
- THEN o processo de governança SHALL manter a regra de que manutenção conta no SLI e de que RTO e perda de região exigem prova específica, e NÃO SHALL aceitar percentual mensal e réplica multi-AZ isoladamente como prova suficiente para esse uso

### Requirement: DEC-05 — Registro de gaps, riscos e mitigação
O processo de governança SHALL manter, sob propriedade conjunta de Arquitetura e liderança técnica, um registro de gaps (GAP-01 a GAP-15) com severidade, risco identificado, mitigação e requisitos de desenho associados, e responsável final com condição de encerramento. "Resolvido no desenho" SHALL significar apenas que a regra foi definida e a contradição documental removida, NÃO SHALL significar que a mitigação foi implementada ou testada; todo item SHALL exigir evidência antes da liberação do escopo indicado, e um limite arquitetural NÃO SHALL ser encerrado apenas por revisão de texto.

#### Scenario: GAP-02 (Crítica) permanece com qualificação pendente até crash/duplicata/UNKNOWN comprovados
- GIVEN os caminhos direto e por fila (EXE-15) compartilham intenção, posse, identidade e recuperação comuns conforme o desenho de mitigação de GAP-02
- WHEN a liberação do escopo afetado por essa duplicação potencial de efeitos é avaliada
- THEN o processo de governança SHALL manter GAP-02 como resolvido apenas no desenho, com Engenharia responsável por comprovar ausência de reexecução indevida em cenários de crash, duplicata e UNKNOWN antes de encerrar a qualificação

#### Scenario: GAP-09 (Crítica) exige limite de corte dentro do SLA sem split-brain antes de fechar migração de célula
- GIVEN uma migração de célula usa placement/epoch, cópia e drenagem com corte exclusivo conforme DAD-11 para mitigar GAP-09
- WHEN a interrupção de cada fase da migração é avaliada
- THEN o processo de governança SHALL exigir que Dados comprove ausência de split-brain em cada fase e que o corte permaneça dentro do limite do SLA antes de considerar o gap encerrado

#### Scenario: GAP-12 (Crítica) mantém a referência atual sem qualificar RPO zero regional
- GIVEN uma topologia regional adicional é proposta apenas com autoridade e cópias adequadas para mitigar GAP-12
- WHEN essa mitigação é avaliada contra a promessa de ausência de perda regional
- THEN o processo de governança SHALL manter, sob responsabilidade de Arquitetura e vínculo com P-08, que a referência atual NÃO SHALL ser tratada como qualificação de RPO zero regional enquanto a evidência não existir

#### Scenario: GAP-05 (Crítica) impede que "banco opcional" confirme pedido ou final sem prova de custódia
- GIVEN fronteiras duráveis e continuidade restrita (DAD-09, OPE-13) mitigam GAP-05
- WHEN uma aprovação avalia se um caminho pode confirmar pedido ou resultado final sem banco disponível
- THEN o processo de governança SHALL manter, sob responsabilidade de Arquitetura, que nenhuma aprovação SHALL chamar de escrita válida uma confirmação sem autoridade de durável, tratando esse limite como assumido e não encerrado por texto

#### Scenario: GAP-13 (Crítica) restringe oferta de garantia externa que o provedor não sustenta
- GIVEN um provedor pode executar, perder a resposta e não oferecer reentrega, status ou idempotência, mitigado por UNKNOWN e contrato de recuperação (EXE-09, OPE-13/15) vinculados a P-05
- WHEN uma oferta comercial é avaliada para esse provedor
- THEN o processo de governança SHALL manter, sob responsabilidade de Integrações, a restrição de não vender garantia externa que o parceiro não sustenta, enquanto a evidência de homologação de P-05 não estiver aprovada

#### Scenario: GAP-15 (Crítica) impede confundir texto completo com disponibilidade comprovada
- GIVEN gates, cenários, matriz e o estado NÃO EXECUTADO (QUA-03/04/06) mitigam GAP-15
- WHEN a conclusão do texto normativo é usada como argumento de disponibilidade da fatia ou do perfil
- THEN o processo de governança SHALL manter, sob responsabilidade da liderança técnica, a exigência de evidências da fatia/perfil, riscos residuais aceitos e responsáveis nominais antes de tratar qualquer escopo como comprovadamente disponível, e a elaboração da engenharia documental SHALL continuar mesmo com P-01 a P-11 ainda pendentes, sem autorizar hipóteses comerciais ocultas

## Notas de origem

Este delta deriva integralmente do capítulo `docs/09_DECISOES_E_REFERENCIAS.md` da especificação de engenharia de requisitos v4.0 do Hub de Interoperabilidade (Constelação), cobrindo os requisitos DEC-01 a DEC-05 — decisões arquiteturais de referência (ADR-01 a ADR-10), registro de decisões pendentes (P-01 a P-11), decisões herdadas da v3 (ADR-11 a ADR-18), decisões introduzidas na v4 (ADR-19 a ADR-25) e o registro de gaps, riscos e mitigação (GAP-01 a GAP-15) — conforme texto normativo de origem. Nenhum código, manifesto, provisionamento ou teste de aplicação foi produzido nesta revisão.
