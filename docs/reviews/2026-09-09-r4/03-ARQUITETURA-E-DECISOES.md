# Arquitetura e critérios técnicos da R4

A revisão preserva o desenho e corrige integração. Não há evidência de que trocar linguagem ou broker
resolveria os defeitos atuais. As escolhas abaixo são continuidade da stack, sujeitas à qualificação de versões.

| Componente | Tecnologia | Responsabilidade | Motivo e limite |
|---|---|---|---|
| Atlas | Go + PostgreSQL control | Catálogo, versão, contrato, binding por referência e projeção | Constraints/transações protegem publicação; resolução precisa índice/projeção, não varredura por pedido. |
| Orbita | Go + PostgreSQL core | Aceite, UUIDv7, intenção, DAG, deadline e final do cliente | Concorrência e custódia local; goroutine não é workflow recuperável. |
| Cometa | Go + tabelas de execução com ownership | Adapter, tentativa, grants, correlação, recibo, polling/callback | context/net/http suportam I/O limitado; precisa pools e controle global realmente conectados. |
| Pulsar | Go + estado durável de entrega | Destino/versionamento, bytes, claim e retry de webhook | Desacopla destinatário lento; não decide novamente resultado nem refaz serviço. |
| Libra | Go + PostgreSQL financeiro | Compra/venda, reservas, ledger, disputas e fechamento | Valores exatos/invariantes transacionais; evento técnico não é automaticamente unidade cobrável. |
| Console | React/TypeScript/Vite | Jornada do operador | Stack existente e DTOs tipados; validação runtime e browser continuam obrigatórios. |
| Gateway | Kong | Ingresso, roteamento e política de borda | Menos roteamento ad hoc; autorização de recurso permanece no domínio e callback requer política própria. |
| Identidade/cofre | OIDC/Keycloak e Secrets Manager no desenho AWS | Identidade/MFA e segredos versionados | Referências em catálogo; L1 limitado não dispensa expiração/revogação. |
| Transporte/objetos | SNS/SQS e S3, emulação LocalStack | Mensagens e bytes | Preserva opção AWS; topologia, DLQ/retention e custódia não são inferidas do nome do serviço. |
| Cache | Redis opcional | Metadados não sensíveis | Ausência não muda autoridade; token saiu do Redis, mas L1 ainda precisa ficar independente do cofre quando válido. |
| Plataforma | Compose, kind e destino AWS/EKS condicionado | Instalação, scheduling, escala e operação | Um Compose do Hub por vez; kind completo testa Kubernetes, não comprova HA regional AWS. |
| Observabilidade | Prometheus/Loki/Grafana/Tempo/Alloy | Métrica/log/trace/consulta | Sinais devem ser emitidos e alertados de fato, sem cardinalidade/segredos irrestritos. |

## Comunicação e persistência
Cliente→Gateway→Orbita usa HTTPS e contrato publicado. SYNC Orbita→Cometa permanece HTTP direto
idempotente, sem fila obrigatória no tempo de resposta. ASYNC usa intenção/outbox e fila com handoff
durável. Cometa→provedor usa capacidade do adapter homologado; importar OpenAPI não implementa
automaticamente REST/SOAP/gRPC/legado.

Callback autentica conta antes de receber custódia; polling e callback convergem na mesma validação.
Cometa conserva recibo e publica fato; Orbita decide final do atendimento. GET e webhook reutilizam
representação final congelada. Pulsar entrega; Libra calcula incidência por snapshot/unidade.
Object storage guarda bytes; DB guarda identidade/versionamento/hash e obrigações de retenção.

Não há fallback seguro genérico que elimine toda persistência e mantenha aceite durável.
Redis fora deve ser dispensável. Control plane/cofre fora pode admitir caminho quente válido dentro
da política. Autoridade durável totalmente indisponível exige recusa antes de prometer custódia,
salvo protocolo alternativo de autoridade/fencing explicitamente qualificado.

## Baixa latência e escala
Medir separadamente autenticação, projeção, custódia, espera de grant, rede, provedor e finalização.
Resolver antes de buscar remotamente quando houver projeção/cache válido; limitar locks/filas/pools.
Paginar todas as ofertas e juntá-las em memória não é resolução seletiva.
Conta compartilhada exige limite global independente do número de pods; adaptação usa métricas e
respeita teto contratual. Tenant agressor deve ser ensaiado junto de tenant saudável com orçamento definido.

Escalabilidade automatizada vale dentro de quotas/envelope e inclui pods, nós, dados, placement e backpressure.
Não prometer crescimento infinito, latência zero, failover instantâneo ou ausência absoluta de interferência.
Prometer custódia de aceites no modelo de falhas e SLO qualificados, não sucesso de negócio de todo pedido.

## Prazos e finanças
Distinguir deadline do cliente, SLA do provedor, timeout por tentativa, TTL desde primeira falha transitória,
horizonte de reconciliação e política de entrega. Nenhum retry renova indefinidamente TTL.
Polling saudável não consome janela de retry de indisponibilidade.
Final tardio fica como evidência externa, sem reabrir protocolo; custo pode existir mesmo após expiração.
UNKNOWN retém orçamento até prova positiva e exige consulta segura antes de reenvio.

## Decisões ainda não presumidas
D-01: volumes/envelope; D-02: SLA/parcialidade/tardio; D-03: incidência/franquia/saldo;
D-04: retenção/dados/região; D-05: AWS/EKS/orçamento/gateway/DR; D-06: equivalência/idempotência/callback;
D-07: catálogo e integração financeira. Reconciliar com P-01…P-11 da v4 sem renumerar para apagar pendência.
Fixtures explícitas permitem trabalho técnico independente.
T-R2-01 exige solução/prova ou decisão normativa explícita. Agente não pode aprovar relaxamento de SLA.

## Referências técnicas verificadas em 2026-09-09
JSON Schema trata 1 e 1.0 como o mesmo inteiro e recusa números representados como strings.
O validador deve seguir o dialeto declarado: [JSON Schema — Numeric types](https://json-schema.org/understanding-json-schema/reference/numeric).
O decoder Go não deve ser confundido com validador de schema:
[encoding/json](https://pkg.go.dev/encoding/json#Unmarshal).
A seleção do projeto Compose é explícita; a restrição de apenas um ativo vem de AGENTS do Hub:
[Docker — project name](https://docs.docker.com/compose/how-tos/project-name/).
