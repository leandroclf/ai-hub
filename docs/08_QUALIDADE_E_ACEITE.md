# 08 · Qualidade, rastreabilidade e aceite

## QUA-01 · Estratégia de engenharia — proprietário: QA / Engenharia

Cada requisito identificado em CAT, ARQ, COM, DAD, EXE, FIN, CFG, SEG e OPE deve ter cenário positivo, negativos relevantes e evidência verificável. A matriz anexa mapeia todos os requisitos, incluindo os de qualidade e decisão. Não é uma declaração de testes executados: o status inicial de todos os cenários é **NÃO EXECUTADO**.

Níveis: revisão de domínio e contratos; testes de invariantes transacionais; integração PostgreSQL/broker/objetos; contrato de cada provedor/cliente; ponta a ponta; concorrência/falhas; segurança; carga e longa duração; recuperação; reconciliação financeira. Cobertura percentual de código, quando existir implementação, é sinal auxiliar, não substituto da cobertura de riscos e requisitos.

Coleção sintética de referência deve conter: serviço síncrono; assíncrono só callback; assíncrono só polling; ambos; operação sem idempotência; callback antecipado; resultado grande; produto agregado; produto composto; provedor que cobra status; cliente com franquia; contrato com saldo estrito. Simulador local reproduz comportamento determinístico e falhas injetadas; homologação também usa sandbox real para capacidades e limites não simuláveis.

## QUA-02 · Casos críticos de falha e concorrência — proprietário: QA

| Cenário | Injeção | Invariante esperado |
| --- | --- | --- |
| Commit sem resposta ao cliente | Perder conexão após admissão durável | Mesma chave recupera mesmo protocolo, sem nova operação |
| Banco gravado, broker fora | Interromper publicação e restaurar | Outbox preservado e publicação posterior, sem pedido perdido |
| Publicação sem marca local | Crash após broker aceitar evento | Redelivery sem reaplicar efeito ou gerar nova receita |
| Consumidor antes/depois do ack | Crash nas fronteiras transacionais | Efeito ou intenção retomável; mensagem não confirmada prematuramente |
| Timeout externo ambíguo | Provedor processa mas resposta se perde | UNKNOWN, consulta segura, sem failover automático indevido |
| Callback antes do submit responder | Antecipar retorno externo | Recibo durável e associação posterior inequívoca |
| Callback e polling simultâneos | Entregar mesmo final em paralelo | Uma conclusão, uma entrega por destino e uma unidade de receita |
| Dois finais incompatíveis | Divergir conteúdo/estado | Evidências preservadas, conflito e política explícita |
| Lease expira com chamada em voo | Pausar worker e iniciar outro | Worker antigo não sobrescreve estado; risco externo não é ocultado |
| Quota de provedor compartilhada | Aumentar pods e tenants | Limite global respeitado, sem multiplicar quota pelo HPA |
| Polling e fetch pagos | Cobrar ambos em sandbox | Quantidades/custos discriminados sem dupla receita do produto |
| Saldo disputado | Pedidos simultâneos para última unidade | Nenhum efeito sem reserva; consumo não excede limite estrito |
| Objeto gravado sem commit | Falhar referência no banco | Órfão reconciliável, não sucesso com referência inválida |
| Webhook indisponível | Falhar durante janela e recuperar | EXHAUSTED/retentativa rastreada; GET segue local; sem reexecução |
| Expurgo de resultado | Vencer retenção autorizada | Consulta informa expurgo, sem chamada automática ao provedor |
| Restore com efeitos externos já feitos | Recuperar backup anterior a operações | Saída bloqueada até reconciliar; não repetir custos/efeitos às cegas |

Além desses casos, testar alteração de preço em voo, vencimento de contrato, evento antigo, incompatibilidade de schema, revogação, certificado vencido, troca de tenant, SSRF, SQL/RLS, XML malicioso quando SOAP for usado e consumo excessivo de recursos por arquivos.

## QUA-03 · Critérios de aceite e evidência — proprietário: QA / responsáveis por domínio

Um caso só passa com pré-condições, massa, versão de contrato/configuração, instantes, resultado observado, protocolo de teste e comparação com esperado. Evidência de desempenho registra hardware/cluster, réplicas, requests/limits, versões, latência do simulador/provedor, taxa, tamanho, fan-out, duração, erros e percentis. Relatório sem essas condições não qualifica um SLO.

Para ausência de perda, contabilizar conjunto de entradas aceitas e confrontar com protocolos, obrigações, resultados e pendências; “fila vazia” não prova completude. Para faturamento, recomputar exemplos sintéticos por calculadora independente, verificar chaves econômicas, balanceamento por moeda, arredondamento e conciliação de extrato. Para consulta local, instrumentar contagem de chamadas ao provedor antes/depois e exigir zero novas chamadas atribuíveis ao GET. Comparar também o hash e o contrato da representação final entre GET e webhook, inclusive falha de SLA, payload legado e reentrega.

Critérios bloqueadores de release: perda de pedido aceito; acesso entre tenants; cobrança duplicada; resultado final sobrescrito sem revisão; retry inseguro de operação incerta; fechamento sem explicação de divergências; segredo exposto; protocolo sem possibilidade de recuperação. A avaliação de outros defeitos depende de impacto, sem reclassificar bloqueadores para cumprir prazo.

## QUA-04 · Gates e definição de pronto — proprietário: liderança técnica

**G0 — pronto para implementar:** domínio, estados, matrizes de atendimento, contratos econômicos, modelo de dados e critérios de aceite revisados; pendências que afetam a fatia selecionada resolvidas; responsáveis nominais definidos. Não exigir conta de produção para começar uma fatia sintética, mas não inventar condições comerciais.

**G1 — local/dev:** implementação da fatia, revisão de código futura, testes de contrato/integração, Compose e kind, probes e telemetria demonstrados. Emulação identificada como tal.

**G2 — hom:** cliente/provedor sandbox homologados, catálogo publicável, TTL/deadline e rejeição tardia demonstrados, contratos legados/GET-webhook equivalentes, polling/callback e apuração aprovados; acesso administrativo e testes negativos entre tenants aprovados.

**G3 — ppd:** infraestrutura representativa por célula, carga/longa duração, ruído entre tenants, adaptação de pressão, escala para cima/baixo, perda de zona e controlador, restore, aferição bilateral de SLA, dashboards e runbooks ensaiados. Mesmo artefato destinado à produção; RPO assumido precisa de prova correspondente. Incluir escala automática de nós/células, falta de Redis, writer/cofre indisponíveis, modo SYNC sem broker e manutenção sob carga.

**G4 — prd:** SLOs aprovados, contratos e retenção válidos, observabilidade e plantão ativos, segregação e credenciais verificadas, plano de rollback/restore aprovado e nenhuma falha bloqueadora. Oferta crítica exige perfil OPE-15, limites de interrupção e cenários QUA-06 aprovados. Não declarar plataforma pronta com endpoints de negócio retornando “não implementado”.

Fatia inicial recomendada: serviço síncrono de provedor disponível em SYNC direto e ASYNC, com UUIDv7, resultado local, credencial compartilhada/dedicada e uma regra simples de receita/custo; depois polling/callback concorrentes; depois produto composto e planos avançados. Qualificar ausência de Redis e falhas nas fronteiras de escrita desde a primeira fatia. É sequência de entrega, não exclusão de requisitos da v4. Não há prazo de calendário presumido.

## Revisão documental realizada nesta entrega

Verificar presença de IDs únicos, rastreabilidade de todos os requisitos, referências internas, consistência da política de resultado local, separação entre compra/venda e distinção de testes planejados versus executados. Nenhum teste de software, performance, segurança ou ambiente foi executado para aprovar a futura implementação: o entregável atual é documental.

## QUA-05 · Qualificação dos incrementos v3 — proprietário: QA / Arquitetura

| Caso obrigatório | Procedimento de qualificação | Critério de aceite |
| --- | --- | --- |
| TTL contínuo | Primeiro erro, retry, reinício de pod e troca de worker; tentar renovar janela | Mesmo first_transient_at/retry_until e fim previsível em segundos |
| TTL zero e erro permanente | Serviço sem retry e serviço com erro de contrato | Sem repetição indevida; resposta final prevista pelo cliente |
| Provedor lento sem falhas HTTP | Submit/polls 200, mas final depois do deadline | EXPIRED/SLA_EXCEEDED e rejeição do final tardio |
| Fronteira temporal | Callback/poll/finalização em D−ε, D e D+ε; atrasar transação e scheduler | Um único final, igualdade tratada como atraso; sem sucesso sem elegibilidade durável comprovada |
| Recebimento cedo, consolidação tarde | Persistir observação externa no prazo, atrasar representação do cliente | Quebra atribuída corretamente ao hub; sem usar hora externa para esconder atraso |
| Retorno tardio e financeiro | Provedor cobrado por aceite conclui após EXPIRED | Zero receita de sucesso, custo/contestação preservados, sem reabertura |
| Callback/poll concorrentes | Mesmo final em canais diferentes cruzando o deadline | Uma decisão terminal, um corpo de saída e uma obrigação de entrega por destino |
| Pressão variável | Provedor simulado dobra capacidade e depois cai para metade; adicionar 429/timeouts | Limite efetivo cai e recupera gradualmente, sem ultrapassar teto nem amplificar retries |
| Múltiplas réplicas | Escalar Cometa, perder controlador e provocar posse antiga | Quota agregada respeitada; fencing; nenhuma multiplicação do orçamento por pod |
| Métrica ausente | Cortar telemetria do controlador com fila crescente | Sem crescimento cego; modo conservador/fail-closed e alerta |
| Async acumulado | Submit rápido, final lento e polls pagos | Pendência externa limita novas submissões; polls e custo entram no orçamento |
| Cliente ruidoso | Tenant A excede quota, B mantém carga contratada; saturar CPU, DB, logs e provedor de A | B mantém SLO dentro do domínio qualificado; A não toma reserva de B |
| Contrato legado por cliente | Mesmo produto com nomes/tipos/erros distintos para dois tenants | Entrada correta vira canônico; saída corresponde à versão de cada cliente |
| GET = webhook | Capturar corpo final em GET e em cada tentativa de callback; mudar configuração atual | Mesmo hash/schema/media type por protocolo/versão; somente headers variáveis permitidos |
| Administração | Cliente tenta acessar outro tenant; dev com e sem perfil global; revogar concessão | Cliente sempre negado; administrador autorizado lê e é auditado; revogado é negado |
| Preservação após falha | Crash após commits, falha de broker e zona, restore e reconciliação | Nenhum fato confirmado perdido dentro do modelo; lacunas identificadas e compromisso RPO documentado |

Cenários devem incluir a matriz dos contratos, passos paralelos/sequenciais, tempos precisos e histórico da decisão adaptativa. Comparar controlador adaptativo com limite fixo conservador no mesmo provedor simulado: throughput útil, erros, latência, oscilação, tempo de recuperação e consumo faturável. Mais requisições enviadas não é melhoria se final útil e SLA piorarem.

Todos os testes seguem NÃO EXECUTADO até implementação e evidência. A revisão documental garante cobertura descrita e coerência, não prova isolamento, arbitragem de commit ou RPO. Casos de deadline, acesso entre tenants, duplicação econômica e perda de fato confirmado bloqueiam promoção.

## QUA-06 · Qualificação integrada da v4 — proprietário: QA / Arquitetura / SRE

| Caso obrigatório | Estímulo/condição | Resultado esperado e evidência |
| --- | --- | --- |
| SYNC nativo | Provedor síncrono, core saudável, broker cortado, capacidade de outbox suficiente | Final na mesma conexão; traço sem espera de fila; todos os commits e fatos recuperáveis |
| SYNC versus AUTO | Mesmo serviço nos dois modos com provedor lento | SYNC nunca vira 202; AUTO respeita espera e protocolo; erro/prazo não é sucesso fictício |
| UUID em todo aceite | SYNC/ASYNC/AUTO, sucesso/falha, resposta de criação perdida | UUIDv7 persistido em cada protocolo, mesma chave recupera mesmo UUID; IDs não concedem acesso |
| Fronteira DIRECT/QUEUED | Perder resposta interna, pausar dono, duplicar comando e acionar watchdog | Uma operação lógica e posse; nenhuma duplicação externa sem idempotência homologada |
| Dupla observação | Final direto e evento da mesma operação chegam simultaneamente | Uma transição Órbita, uma unidade econômica e corpo final único |
| Prazo SYNC e TTL | TTL maior que conexão, retries seguros e provider final atrasado | Prazo efetivo limitado; EXPIRED durável quando possível; retorno tardio rejeitado; custo preservado |
| Paralelismo SYNC | Dois passos independentes e sucessor dependente | Independentes paralelos, sem transação em espera; grafo/slots/deadline respeitados |
| Redis totalmente desligado | Pico contratado, cache L1 frio e perda de uma zona | Operações válidas mantêm SLO qualificado; sem dependência oculta de lock/idempotência/quota |
| Redis intermitente | Injetar latência, falhas e recuperação em massa | Bypass limitado e aquecimento gradual; sem avalanche no banco |
| Writer indisponível antes do aceite | Cortar autoridade transacional e enviar criação | Nenhum efeito externo nem falso aceite; leitura comprovada independente continua quando possível |
| Commit incerto / retorno externo sem custódia | Falhar conexão de commit e writer após efeito | Recuperação por chave, UNKNOWN/reconciliação, sem retry inseguro ou final inventado; lacunas explícitas |
| Réplica atrasada | Final novo ausente, final revisável antigo, final imutável disponível | Sem 404/pending incorreto; só informação comprovadamente adequada e autorizada é servida |
| Cofre quente/frio | Perder Secrets Manager/KMS com e sem material previamente obtido | Uso apenas dentro da validade; cold start bloqueia binding afetado, sem trocar conta |
| Dedicada ausente | Tenant A dedicado inválido, compartilhada saudável e B dedicado saudável | A não usa outra credencial; B continua; trilha sem segredo |
| Rotação em voo | Submit com versão antiga, poll/fetch/callback durante rotação | Conta/operação preservadas; versão usada rastreada; chave revogada não reaparece |
| Conta e pagador | Três modelos FIN-11 com polls pagos e final tardio | Custo/receita/conta a pagar corretos e segregados; nenhum débito presumido pelo tipo da chave |
| Escala sem chamado | Aumentar clientes, provedores e mix até exigir pods, nós e célula nova | Ativação por workflow, capacidade pronta antes da rota; nenhuma operação manual exigida no cenário aprovado |
| Quota negada / controle fora | Negar criação em nuvem ou parar Crossplane/Karpenter | Reserva ativa preservada; nova ativação espera com motivo; alarme antes de esgotar folga |
| Conta global em várias células | Escalar Cometa e perder dono do orçamento | Soma das concessões respeita teto; slots UNKNOWN não liberados cegamente; sem multiplicar capacidade por segredo/pod |
| Realocação interrompida | Falhar cada fase de DAD-11 e repetir chave antiga após corte | Uma autoridade e idempotência preservadas; sem efeito duplicado; barreira dentro do orçamento ou migração abortada |
| Manutenção sob carga | Rolling APIs, drenagem, failover writer e rotação de chave | Tempo observado comparado ao perfil; perda zero no domínio qualificado; erro não é omitido do SLI |
| Redis versus autorização | Cache possui resultado de outro tenant ou acesso revogado | Leitura negada mesmo em fallback; nenhum vazamento por bypass |
| Falha correlacionada crítica | Perda de zona + cache frio + indisponibilidade temporária de controle, com carga crítica | Reserva e perfil de continuidade demonstrados; não apenas ensaio de falhas isoladas |
| Perda regional / partição | Exercício no perfil contratado | Autoridade única, RPO/RTO comprovados; perfil regional não qualificado permanece bloqueado |

Massa deve cruzar modalidades, classes de criticidade, conta compartilhada/dedicada, contrato legado, agregação/composição e saldo estrito/pós-pago. O baseline de performance inclui núcleo durável completo, autenticação, transformação, provedor simulado controlável e depois sandbox real. Desligar Redis é requisito de aceite, não apenas opção de benchmark. Provar ausência de chamada à SQS no percurso obrigatório de SYNC não significa desligar a outbox de fatos.

A análise de falhas relaciona cada risco à ação preventiva, detecção, fallback, recuperação, dependência e responsável. Ensaios registram números antes/depois: latência final SYNC, aceite ASYNC, GET, throughput útil, erros, backlog, instante do efeito e dos commits, uso de conexões/memória, conta/credencial sem segredo e completude econômica. Quando não houver dado suficiente, o resultado é NÃO QUALIFICADO, não aprovação por ausência de erro observado.

Cobertura documental não mede confiabilidade. Todos os casos desta seção e as linhas da matriz permanecem NÃO EXECUTADOS. Falta de teste crítico bloqueia apenas a oferta/perfil afetado, mas a produção dessa oferta não é liberada por aceite de uma fatia menos exigente.
