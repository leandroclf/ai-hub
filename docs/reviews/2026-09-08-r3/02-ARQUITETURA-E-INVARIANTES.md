# Arquitetura, responsabilidade e motivo das escolhas

A R3 preserva a stack implementada; não propõe reescrita. Go/React/PostgreSQL não são a causa dos principais gaps.
Versões presentes no Compose são inventário do snapshot, não recomendação automática de versões atuais.
Antes de promover, fixar digests e qualificar compatibilidade/suporte/vulnerabilidades sem atualizar tudo às cegas.

| Componente | Linguagem/tecnologia | Responsabilidade e dados próprios | Motivo da escolha e limite |
|---|---|---|---|
| Atlas | Go + PostgreSQL control | Catálogo, ofertas, versões, contratos, bindings por referência e projeções | Transações/constraints protegem publicação; Go permite compartilhar contratos tipados. Não consultar control DB por toda requisição do cliente. |
| Orbita | Go + PostgreSQL core | Aceite, UUIDv7, idempotência, intenção, etapas, deadlines, representação final | Concorrência estruturada e custódia transacional suportam SYNC/ASYNC. Goroutines são concorrência, não persistência de workflow. |
| Cometa | Go + PostgreSQL de execução sob ownership próprio | Adapter, tentativa, recibo, correlação, polling, capacidade | net/http e context suportam I/O cancelável/pools; banco permite recuperar owner/epoch e efeitos incertos. Não colocar provider effect em transação longa. |
| Pulsar | Go + tabelas de entrega sob ownership próprio | Destino versionado, entrega, claim, tentativa e recibo HTTP | Workers independentes e bytes persistidos desacoplam cliente lento. At-least-once exige idempotência no destinatário; 2xx externo é a garantia que o contrato declarar. |
| Libra | Go + PostgreSQL financeiro | Fatos, ledger, reservas, franquias, disputas e fechamento | NUMERIC/aritmética exata e transações são adequados a invariantes monetários. Não usar float64, nem compensar apagando lançamento. |
| Console administrativo | React + TypeScript + Vite | Jornada de operador; nenhum dado de negócio autoritativo no navegador | Stack já adotada facilita UI tipada e build. TypeScript sozinho não valida JSON de runtime nem comprova jornada; testes browser são obrigatórios. |
| Gateway | Kong declarativo | Entrada, roteamento, autenticação suportada, limites de borda | Evita roteamento ad hoc; edição/plugins devem ser explicitamente qualificados. Autorização de recurso continua no serviço. |
| Identidade | Keycloak local / provedor OIDC homologado nos ambientes | Sessão, MFA, papéis, workload identity | OAuth/OIDC com PKCE evita credencial de usuário na aplicação. Fixture local deve ter bootstrap reproduzível e identidade nominal. |
| Mensageria | SNS/SQS; LocalStack somente para emulação local | Transporte durável e fan-out para filas obrigatórias | Preserva baseline AWS e operação gerenciada. Não é event store contábil; assinaturas, DLQ, retenção e reconciliação precisam ser explícitas. |
| Objetos | S3 compatível por ambiente | Bytes versionados, checksum e lifecycle | Upload direto/streaming evita payload volumoso no broker/banco. Autorização é do Hub; URL temporária não é identidade durável. |
| Cofre | AWS Secrets Manager no desenho AWS; emulação local | Segredos e versões | Evita credenciais em catálogo/log/Redis; binding referencia a versão e escopo. Disponibilidade exige cache de memória governado, não bypass de revogação. |
| Redis | Cache opcional de metadados não sensíveis | Nenhuma autoridade de aceite/efeito/saldo | Pode reduzir leituras; sua falha não deve bloquear caminho quente. Remover tokens observados no Redis atual. |
| Orquestração de containers | Compose local; kind para qualificação Kubernetes; EKS na opção AWS | Scheduling, serviços, volumes, escala, rollout | Compose simplifica reprodução; kind testa APIs Kubernetes e topologia; EKS é destino condicionado à decisão de plataforma. kind não comprova HA multizona real. |
| Observabilidade | Prometheus, Loki, Grafana, Tempo, Alloy | Métricas, logs, traces e visualização | Stack já presente cobre sinais complementares. Dashboards importados não provam séries emitidas, e não substituem auditoria de negócio durável. |

## Comunicação definida

| Origem → destino | Tipo | Contrato/garantia |
|---|---|---|
| Cliente → Gateway → Orbita | HTTPS REST/JSON ou formato legado homologado | Autorização tenant/application; idempotency key; UUIDv7 após aceite; limite de payload. |
| Orbita → Cometa SYNC | HTTP direto idempotente protegido, deadline propagado | Retorno final na mesma requisição dentro do budget; broker não é hop obrigatório. |
| Orbita → Cometa ASYNC | Intenção durável/outbox → fila de comando | Ack após custódia; owner/epoch e recuperador; sem depender da memória do processo. |
| Cometa → provedor | Adapter REST/SOAP/gRPC/legado conforme capacidade efetivamente implementada | SUBMIT/STATUS/FETCH distintos, binding congelado, timeout, grant, evidência. Catálogo não implica suporte automático a todos os protocolos. |
| Provedor → Cometa | Callback autenticado da conta; polling de STATUS opcional e combinável | Recibo antes de 2xx; convergência e retenção de tardio; consulta não repete SUBMIT. |
| Cometa → Orbita/Libra | Fato durável por filas obrigatórias | Envelope versionado, identidade integral, inbox, outbox e quarentena recuperável. |
| Orbita → Pulsar/Libra | Final durável + representação/hash e snapshot econômico | Final único; entrega e incidência independentes e idempotentes. |
| Cliente → consulta de protocolo | HTTPS leitura local do Hub | Não chama provedor quando há final; aplica isolamento e reutiliza representação congelada. |
| Pulsar → cliente | HTTPS webhook | Corpo idêntico ao GET final; envelope de assinatura em headers; destino congelado e política de retry própria. |
| Console → APIs administrativas | HTTPS/OIDC | DTO específico, autorização por recurso, MFA onde exigido, motivo e chave de intenção para comandos. |
| Atlas → data plane | Projeção versionada com atualização/reconciliação | Validade e revogação limitam fallback; não confiar em cache vencido. |

Não introduzir gRPC, Kafka, MongoDB, Cassandra, Rust ou Java apenas por popularidade.
Uma alternativa requer problema medido, ADR, migração e prova de ganho superior ao custo operacional.

## Invariantes de persistência
1. Nenhum aceite positivo antes de custódia do protocolo e intenção.
2. Nenhum efeito externo antes de tentativa identificada e autorização de capacidade/contrato.
3. Nenhum ACK de mensagem/callback antes de recibo ou obrigação recuperável.
4. Cada final é único e tem representação persistida; GET não busca novamente no provedor.
5. Estado externo UNKNOWN pode sobreviver ao atendimento EXPIRED; não reabre resultado, não libera saldo sem prova.
6. Duplicata de mensagem não significa nova unidade econômica. Unidade contratada tem chave própria.
7. Restore exige reconciliação com efeitos externos; backup do banco sozinho não prova exatamente uma execução.
8. Redis fora não muda autoridade. Banco autoritativo totalmente fora não permite prometer aceite durável em memória.

## Relógios e estados
Separar protocolo de atendimento, operação externa, tentativa, entrega e obrigação financeira.
Protocolo termina uma vez em sucesso/falha/expiração conforme contrato; operação externa pode continuar em reconciliação.
Definir no snapshot client_deadline, provider_deadline, first_transient_failure_at, retry_deadline,
timeout de tentativa, política de polling e prazo de entrega/reconciliação.
Retry usa min(janela restante, budget de tentativa e limites aplicáveis); TTL zero não autoriza retry.
Polling saudável não é retry de falha. Resultado tardio é conservado como evidência e recusado para atendimento.
O ADR T-R2-01 demonstra risco entre decisão SQL e commit. Não declarar essa decisão encerrada com teste de relógio anterior ao commit.

## Baixa latência e isolamento
O budget deve medir: borda/autorização + resolução local de contrato + custódia + espera de grant +
interconexão + tempo do provedor + persistência final. Publicar p50/p95/p99 de cada parcela.
Retirar chamadas remotas desnecessárias ao control plane/cofre no caminho quente, reutilizar pools e usar buffers limitados.
Não manter mutex compartilhado durante rede. Não criar goroutine/fila ilimitada por entrada.
Agregação paraleliza apenas etapas independentes, com budgets por tenant/produto/provedor.
Bulkheads e justiça limitam interferência; prova exige workload A saturado enquanto B é observado.
Nenhuma arquitetura compartilhada garante impacto literalmente zero em qualquer carga: estabelecer envelope,
reservas, SLO por classe e recusa controlada além da capacidade.

## Modelo de disponibilidade
Garantia é de custódia das requisições aceitas no modelo de falhas qualificado, não de sucesso de negócio garantido.
Replica/quorum, retenção, backup, DR e capacidade devem ter RPO/RTO/SLO medidos e responsáveis.
Falha Redis: continuar caminho quente. Falha broker: SYNC direto elegível continua e obrigações ficam duráveis;
ASYNC depende de espaço/custódia disponível e políticas de admissão. Falha total do banco autoritativo:
recusar nova promessa durável; oferecer apenas leitura segura onde comprovada. Não inventar fallback que divida autoridade.
