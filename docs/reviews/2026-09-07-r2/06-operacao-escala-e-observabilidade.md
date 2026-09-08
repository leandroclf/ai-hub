# Operação, ambientes, escala e observabilidade

Status: especificação para implementação; nenhuma infraestrutura remota foi aplicada. HPA, probes e Terraform existentes foram inspecionados, não executados nesta revisão. O laboratório solicitado inclui a pilha completa e deve ser reprodutível sem arquivos privados de uma estação.

## Ambientes e dependências

| Ambiente | Execução e finalidade | Dados/dependências | Gate de saída |
|---|---|---|---|
| local | Docker Compose; desenvolvimento e jornada integral | PostgreSQL por domínio, LocalStack para SNS/SQS/S3, IdP e cofre de ensaio, Redis opcional, UI/gateway, Prometheus/Grafana/Loki/Alloy/OTel/Tempo; simulador isolado | Uma subida documentada; fixtures próprias; recriação preserva volume; reset separado |
| local-kind | Cluster kind com topologia laboratorial explícita | Mesmos contratos, imagens e manifests; dependências de teste e controllers necessários; nós com labels de zonas simuladas quando pertinente | Rotas, probes, HPA/KEDA, shutdown, migrations e falhas de pod testados; não prova perda de zona física |
| dev | Kubernetes remoto de referência, separado de hom | Dados sintéticos, identidade/segredos próprios; contratos de dependência iguais aos promovidos | Integração/contract tests e smoke completo; imagens identificáveis por digest |
| hom | Kubernetes; homologação funcional e parceiros sandbox | Dados autorizados/sanitizados; contas externas homologadas; observabilidade completa | Jornadas, cenários negativos, identidade externa real e contratos aprovados |
| ppd | Topologia representativa de prd; multi-AZ na referência | Mesma arquitetura, políticas, versões e classe de dados permitida; carga e N−1 qualificados | Latência, ruído, failover, restore, manutenção, segurança e capacidade demonstrados |
| prd | Kubernetes gerenciado de referência e dependências qualificadas | Identidades, domínios de dados, backups e recursos isolados; capacidade aquecida | Pendências aplicáveis resolvidas, observabilidade e contingência operáveis, gates aprovados |

local-kind é uma modalidade de laboratório do ambiente local, não um sexto ambiente comercial. A palavra kind não substitui Kubernetes gerenciado em prd. Se a opção for toda infraestrutura autogerenciada, P-01 exige decisão distinta de disponibilidade/custo/operação; não inventar equivalência entre LocalStack e AWS real.

## Papéis de runtime e políticas operacionais

| Aplicação/dependência | Papéis e escala | Readiness/startup | Liveness e desligamento |
|---|---|---|---|
| Portal | Réplicas/HPA; APIs e administração com limites independentes | Rotas, TLS e material de validação de identidade elegíveis | Saúde local; drenar conexões; não depender de todo provedor |
| Atlas | API/ publicação; HPA e pools próprios | writer do controle e versão/políticas válidas para publicação | Falha de outro domínio não reinicia Atlas |
| Órbita admissão | HPA, piso aquecido, conexões e byte budget | writer, políticas/rotas válidas; reserva estrita por operação quando necessária | Drenar HTTP; impedir novos despachos sem custódia; conservar aceites em voo |
| Órbita leitura | HPA e pool/reserva separados da admissão | fonte autorizada de consulta elegível; final imutável pode usar réplica/cache | Writer fora não remove automaticamente essa capacidade |
| Cometa direto/callback | HPA ou política explícita para API; piso aquecido | persistência, projeções elegíveis; callback exige custódia de recibo | Draining impede novos claims; recibo válido não recebe ACK sem commit |
| Cometa workers/polling | KEDA/métricas de idade, pools por domínio/tenant; concorrência limitada | claim store, projeções e caminho de trabalho pertinentes | Lease/visibility renovados; UNKNOWN conservado ao interromper operação |
| Pulsar workers | KEDA com sinal de obrigações de entrega, não apenas fila de entrada | agenda/recibos e acesso ao final por versão | Clientes lentos não prendem demais; drenar tentativa/lease |
| Libra reserva | HTTP com piso e escala por latência/concorrência; um writer financeiro | autoridade financeira, moeda/limite e esquema elegíveis | Reserva em voo é resolvida por identidade, sem liberar por timeout de transporte |
| Libra apuração | KEDA por lag/idade financeira e backlogs; pool separado | inbox/journal e snapshots | Não ACK antes do efeito; drenar transações curtas |
| PostgreSQL | Redundância e capacidade por célula; pooler se demonstrar benefício | failover/conexão, replicas e lag medidos | Backup/PITR não substituem replicação; nunca HPA para writers |
| SNS/SQS/S3/cofre | Serviços gerenciados remotos, emulados em laboratório | Verificação por capacidade; clientes AWS não criam recursos em runtime remoto | Retry limitado/circuito; SDK usa workload identity |
| Redis | Opcional, L2 de dados permitidos | Não é gate para subir Hub ou autorizar segredo | Bypass rápido; limites de memória/tempo; nenhum token compartilhado |
| Telemetria | Stack provisionada e com quotas próprias | Saída de observabilidade não é readiness de negócio | Buffers limitados; auditoria durável não depende de Loki |

As cinco aplicações lógicas continuam as mesmas; papéis separados são decisões de isolamento/escala dentro delas. Não escalar Cometa direto exclusivamente por fila, nem Pulsar exclusivamente pelo tópico recebido se as obrigações já estão na agenda. Não permitir dois HPAs controlarem o mesmo Deployment. Os mínimos, máximos e requests/limits reais dependem de P-02/P-11; o laboratório terá perfil sintético próprio explicitamente não comercial.

## Fallback e suas fronteiras

| Falha | Continua quando | Limite seguro e comportamento observável |
|---|---|---|
| Redis | Todo o fluxo deve suportar bypass, com L1/dado autoritativo válido | Nada importante existe somente em Redis; não fazer chamada lenta a cache antes de toda operação |
| Atlas | Projeção assinada/versionada ainda válida e revogação atendida | Sem configuração conhecida e elegível, recusar apenas oferta afetada; cache expirado não vira autorização eterna |
| Broker | SYNC/GET com writer/material válido; ASYNC pode aceitar dentro do backlog durável autorizado | Outbox preserva fatos/intents; alerta por idade e admissão limitada pelo envelope; bootstrap de HTTP não cria recursos de broker |
| Cofre/IdP externo | Material privado não expirado/revogado ainda atende política | Sem identidade/segredo válido não há acesso; nunca fallback para outra credencial; demais contas continuam |
| Writer core | GET de final confirmado em cópia confiável pode continuar | Não admitir em memória, Redis ou disco de pod; respostas não confirmadas não viram sucesso. Readiness por papel |
| Writer financeiro | Pós-pago com intenção/fato custodiado pode prosseguir conforme contrato | Oferta estrita precisa reserva confirmada. Não “emprestar” saldo de cache; limite ausente não é infinito |
| S3 | Operações sem arquivos e GET inline confirmado | Fluxo que necessita custódia de objeto não anuncia sucesso com referência quebrada |
| Provedor | Outros provedores/domínios/clientes com isolamento permanecem elegíveis | Controlar pressão, retry seguro/TTL e eventual failover homologado; UNKNOWN impede efeito duplicado |
| Região | Somente perfil de continuidade cuja autoridade e dados estejam preservados | Cercar escritor antigo e provar recuperação. Replicação assíncrona não permite prometer RPO zero regional |

Os fallbacks buscam reduzir indisponibilidade, mas não permitem chamar recusa 503 de “zero downtime”. SLIs contabilizam falha percebida. “Sem impacto perceptível” é critério a ensaiar no perfil sem Redis, sob carga comparável, incluindo cache frio e reinício.

## Controle adaptativo e prevenção de competição

1. **Domínio de capacidade:** chave publicada que representa limite real do parceiro (conta, contrato ou conjunto de contas), independente da rotação de segredo e da réplica do Hub. Atlas conserva o mapeamento; Cometa é autoridade dos créditos/pendências externas.
2. **Sinais:** taxa elegível, inflight HTTP, operações assíncronas ainda abertas, 429/Retry-After, 5xx, timeout, latência e idade de reconciliação. Separar erro funcional 4xx de sobrecarga; 401/403 bloqueia credencial afetada, não precisa reduzir todos os parceiros.
3. **Algoritmo proposto:** AIMD com janela, mínimo de amostras, redução multiplicativa após sinal de congestão, recuperação aditiva após estabilidade, piso/teto e circuit breaker. Parâmetros são versão de política. Exemplo de ensaio, não default comercial: janela 5 s, mínimo 20 amostras, redução ×0,7, aumento de 1 unidade após três janelas estáveis. Ajustar taxa e concorrência sem liberar artificialmente pendências externas desconhecidas.
4. **Quota agregada:** autoridade concede leases de créditos limitados por validade/epoch a executores. Somar créditos ainda válidos e inflight antes de conceder novos. Durante partição, executor só usa reserva que ainda tem autoridade; renovação falha não reinicia orçamento. Horário/expiração exigem margem de incerteza; T-R2-04 qualifica a solução entre células.
5. **Fairness:** filas lógicas por tenant/oferta/classe, escalonamento ponderado dentro de grants e bulkheads de conexão/CPU/memória. Contas compartilhadas compartilham quota externa, mas não podem consumir a reserva de outro cliente sem política explícita de capacidade ociosa recuperável.
6. **Reconciliação:** reservar capacidade para status/fetch/cancel e callbacks/GET; não deixar novos submits impedirem conclusão de obrigações já aceitas. ACK de submit assíncrono libera conexão HTTP, não a pendência externa.
7. **Escala:** HPA/KEDA adicionam executores, não elevam contrato do provedor. Provedor que melhora recebe sondagem gradual e pode aumentar capacidade até teto autorizado. Sem teto real, definir envelope de risco aprovado, não crescer sem limite.

## Modelo de capacidade e orçamento de latência

Medir RPS admitidos, picos e duração de rajada; mix SYNC/ASYNC/AUTO; duração externa p50/p95/p99; fan-out/passos, retries/status/fetch, bytes de entrada/final/objeto; distribuição por tenant/provedor; conexões e CPU por transformação. Derivar demanda externa aproximadamente como RPS admitido × passos médios × tentativas elegíveis, adicionando polling/reconciliação — não dimensionar só pelo número de clientes. Pendências estimadas seguem taxa × duração; capacidade de DB inclui custo de outbox/inbox, final, financeiro e consultas.

Somatório dos máximos de conexões de todos os pods em uma célula deve caber no orçamento do banco, com reserva de manutenção/replicação. Aumentar réplicas sem ajustar pools pode piorar disponibilidade. Cache L1 tem limite por bytes/itens, validade, evicção e métricas; não manter toda a população indefinidamente em maps.

Metas v4 continuam propostas para perfil declarado: admissão ASYNC p95 250 ms/p99 500 ms; GET local de JSON até 64 KiB p95 100 ms/p99 250 ms; final ASYNC observado no Cometa até final Órbita p95 2 s/p99 5 s; primeira tentativa webhook p95 5 s; atualidade financeira p95 60 s. SYNC é medido de borda a final durável dentro do orçamento negociado; 202 rápido não conta como SYNC. Separar latência do Hub e do provedor, sem subtrair percentis independentes.

## Instrumentação mínima

| Sinal proposto | Dimensões controladas | Alerta / investigação |
|---|---|---|
| hub_http_duration_seconds | ambiente, célula, papel, rota normalizada, modalidade, classe | percentis/erro por perfil; nunca URL com UUID |
| hub_obligation_age_seconds | tipo, estado, célula/classe | aceite sem despacho; outbox final sem entrega; reconciliação velha |
| hub_provider_request_duration_seconds / errors_total | domínio/provedor/adapter e tipo de operação, com cardinalidade orçada | congestão, falha de autenticação e degradação de contrato |
| hub_capacity_limit / inflight / external_pending | domínio/célula autorizados | limite efetivo, teto, créditos/leases, capacidade protegida |
| hub_sla_verdict_total | relação, serviço/classe, verdict | violações bilaterais; relatório detalhado vem de dados autorizados |
| hub_outbox_pending / inbox_conflicts / quarantine_total | domínio, tipo, célula | ACK prematuro detectado por invariantes, lag e poison message |
| hub_financial_lag_seconds / uncertain_holds / reconciliation_pending | célula, moeda, natureza | não fechar período com pendência, investigar saldo/ledger |
| hub_dependency_state / pool_wait / cache_bypass | componente, dependência, papel | diferenciar fallback saudável de perda de capacidade |

Alta cardinalidade de tenants/contratos deve ficar em logs/relatórios protegidos e projeções, não criar série por cliente ilimitada. Trace e protocolo são campos de investigação, não labels. Auditoria de leitura global/publicação/ajuste possui store próprio e retenção aprovada.

## Runbooks a implementar e exercitar

- **Aceite sem despacho / outbox atrasada:** localizar intenção por UUID, inspecionar idade/claim/epoch/broker, retomar mesma obrigação ou expirar por prazo; nunca criar novo pedido como reparo.
- **UNKNOWN externo:** bloquear novo efeito, recuperar status por identidade homologada, preservar hold e evidência; se parceiro não permite prova, acionar exceção rastreada.
- **DLQ/quarentena:** identificar schema/causa e corrigir consumidor/configuração; reprocessar com origem/idempotência; conferir efeitos e custódia antes/depois.
- **Provedor degradado:** inspecionar domínio, erros e limite efetivo; reduzir pressão e verificar credencial/rede; failover só se equivalente e seguro.
- **Manutenção/restore:** drenar, cercar egress, restaurar dados/objetos e reconciliar, testar canário, reativar autoridade única; não perder pendências financeiras.
- **Saldo/fechamento divergente:** impedir ajuste destrutivo, rastrear unidades/snapshots/lançamentos, corrigir por compensação e aprovação; exportação mantém identidade.

Cada runbook deve definir responsável funcional, comando diagnóstico seguro, pré-condição, evidência, condição de recuperação e escalonamento. Não inclui credenciais ou comandos destrutivos prontos para executar neste pacote.
