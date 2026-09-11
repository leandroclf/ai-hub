# Runbooks operacionais do AI Hub R2

Runbooks mínimos referenciados pelos alertas do perfil local/Kind. Eles
orientam diagnóstico e recuperação sem incluir credenciais, tokens ou ações
destrutivas. Em ambientes remotos, substitua os comandos locais pelas
interfaces aprovadas da plataforma e registre o identificador da mudança.

## Responsabilidades e regras comuns

- SRE/Operações é o responsável primário por disponibilidade, filas,
  telemetria e escalonamento.
- Engenharia do domínio é consultada quando houver protocolo `UNKNOWN`,
  quarentena, outbox/inbox ou divergência financeira.
- Não confirme sucesso, faça replay de provedor ou remova mensagem/volume
  durante o diagnóstico. A autoridade durável e a evidência de negócio têm
  precedência sobre o alerta.
- Toda intervenção deve registrar ambiente, célula, componente, janela UTC,
  correlação, comando/automação usada e critério de recuperação.

## `capacidade-indisponivel`

### Gatilho

`HubWorkloadUnavailable` indica que um componente do job `hub` deixou de ser
coletado por 30 segundos.

### Diagnóstico seguro

1. Confirmar o ambiente e a célula nos labels do alerta; não assumir que a
   falha é global.
2. Verificar readiness/liveness e reinícios do componente no orquestrador.
3. Conferir o último `trace_id`/`protocol_id` autorizado e separar falha de
   HTTP, autoridade PostgreSQL, broker, cofre e provedor.
4. Verificar outbox, quarentena e leases antes de qualquer reprocessamento.

### Recuperação e evidência

Restaurar somente o workload dentro do procedimento de rollout aprovado,
aguardar readiness estável e comprovar que novas admissões não receberam
`202` sem custódia. Anexar estado do workload, janela de indisponibilidade,
RTO observado, métricas de erro e consulta de pendências duráveis.

### Escalonamento

Escalar para Plataforma quando a causa for nó, rede, controlador ou banco; para
Engenharia quando houver perda de autoridade, `UNKNOWN` ou obrigação sem dono.

## `outbox-e-quarentena`

### Gatilho

`HubOutboxDelayed` indica obrigação de outbox com idade acima de 30 segundos.
Quarentena ou DLQ pode acompanhar a mesma indisponibilidade, mas não deve ser
tratada como autorização para descartar ou reenviar efeitos externos.

### Diagnóstico seguro

1. Medir idade, quantidade, domínio, célula, tenant e estado da obrigação sem
   expor payload ou segredo.
2. Verificar worker, lease/epoch, conexão com a autoridade e disponibilidade
   do broker.
3. Conferir se há `UNKNOWN`, entrega `EXHAUSTED`, callback pendente ou fato
   financeiro ainda não custodiado.

### Recuperação e evidência

Corrigir a dependência ou retomar o worker por procedimento idempotente. A
mensagem só pode ser confirmada depois do commit local; reexecução externa só
é permitida quando a correlação durável e a política do domínio autorizarem.
Registrar antes/depois da outbox, inbox/quarentena, epoch do owner, efeitos
observados e ausência de cobrança duplicada.

### Escalonamento

Escalar imediatamente para o dono do domínio quando existir efeito externo
incerto, atraso de fechamento financeiro ou quarentena crescente.

## `telemetria`

### Gatilho

`HubTelemetryDropped` sinaliza perda do exportador/buffer de telemetria. A
telemetria é auxiliar: sua perda não pode bloquear o caminho de negócio nem
substituir auditoria, outbox, inbox ou ledger.

### Diagnóstico seguro

1. Confirmar o contador de descarte e o componente que perdeu exportação.
2. Verificar Alloy/OTel, Prometheus, Loki e Tempo sem usar o backend de
   observabilidade como autoridade de estado.
3. Confirmar que HTTP, admissão, custódia financeira e auditoria continuam
   respondendo conforme os probes e os bancos duráveis.

### Recuperação e evidência

Restaurar o exportador ou sua capacidade dentro do limite aprovado; não elevar
buffers indefinidamente nem bloquear requisições para recuperar amostras.
Registrar intervalo de perda, métrica `hub_telemetry_dropped_total`, tráfego
aceito, estado da auditoria/ledger e teste de recuperação.

### Escalonamento

Escalar para Observabilidade quando a perda persistir ou os backends estiverem
indisponíveis; escalar para o dono do domínio se houver impacto nos SLIs de
negócio ou divergência entre telemetria e custódia.

## Critério de encerramento

O incidente só é encerrado quando o alerta estiver resolvido pelo período de
estabilidade definido no perfil, a causa e o domínio de falha estiverem
registrados, as obrigações duráveis tiverem disposição explícita e a evidência
de recuperação estiver anexada ao incidente. Um alerta ausente por falha do
Prometheus não é considerado resolvido.
