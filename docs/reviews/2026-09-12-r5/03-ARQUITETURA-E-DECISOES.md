# Arquitetura preservada e decisões de engenharia

As tecnologias são mantidas por continuidade e adequação de responsabilidade. Não há prova de que reescrever a stack resolveria os defeitos auditados. Versões efetivamente construídas precisam ser qualificadas; não há benchmark comparativo de linguagens nesta revisão.

| Componente | Tecnologia/linguagem | Responsabilidade e motivo | Limite e ação |
|---|---|---|---|
| Portal | Kong + HTTPS | Ingresso/roteamento/política de borda já integrados | Autorização de recurso no domínio; callback com política própria |
| Atlas | Go + PostgreSQL control | Catálogo versionado, elegibilidade, contas, bindings e contratos; transações/constraints | Projeções com lease/revogação, consulta seletiva e administração escalável |
| Órbita | Go + PostgreSQL core | UUIDv7, admissão, intenção, plano/etapas, prazos e resposta final | Estado durável, slots atômicos e compensação fora do prazo do cliente |
| Cometa | Go + net/http + estado de operações | Adapter REST, autenticação, capacidade, tentativa, receipts e observação | Pools/budgets/fencing reais; adapter define protocolo suportado, não importação mágica |
| Pulsar | Go + estado de entrega | Webhook com destino e bytes congelados, retries independentes | Nunca repetir SUBMIT para reenviar notificação; permissão/custódia verificadas |
| Libra | Go + PostgreSQL finance | Compra/venda, valores exatos, reserva, journal e corte | Captura efetiva, watermark e quarentena recuperável são necessários |
| Console | React + TypeScript + Vite | Administração de catálogo/operação/financeiro no stack existente | Types não validam payload runtime; busca paginada e intenção recuperável |
| Identidade | OIDC/Keycloak local | Fluxos de identidade nominal, MFA e workload já integrados | Identidade de ambiente e autorização global auditada; fixtures não são produção |
| Cofre | Secrets Manager; LocalStack em ensaio | Segredos versionados por binding compartilhado/dedicado | L1 não ultrapassa revogação/validade; nunca fallback para outro tenant |
| Mensagem | SNS/SQS; LocalStack em ensaio | Fan-out e filas duráveis conforme desenho atual AWS | Topologia/DLQ e ACK qualificados; at-least-once exige idempotência |
| Dados pesados | S3 + FileRef | Bytes externos ao banco, hash/versão/pins/retention | Restore de versão e acesso externo ao objeto precisam contrato do adapter |
| Cache | Redis opcional + L1 | Reduzir consultas sem assumir autoridade | Redis fora não pode impedir serviço; negação explícita não vira cache autorizado |
| Runtime | Docker Compose / kind / destino Kubernetes | Laboratório reproduzível e evolução por perfis | Kind agora independente; HA regional e escala de nós/dados exigem IaC/ensaio |
| Observabilidade | Prometheus/Loki/Grafana/Tempo/Alloy | Métricas, logs e traces existentes | Sinais bilaterais e alertas comprovados; cardinalidade limitada |

## Comunicação definida
Cliente→gateway→Órbita: HTTPS REST/JSON com contrato publicado por cliente/aplicação.
SYNC de serviço elegível: Órbita→Cometa via HTTP interno direto, final na mesma requisição quando possível dentro do contrato, sem fila obrigatória; UUIDv7 permite consulta futura.
ASYNC/AUTO: intenção/outbox→SNS/SQS→Cometa; receipts/fatos retornam por transporte durável.
Provedor síncrono pode ser exposto assíncrono; observação de provedor assíncrono usa polling/callback configurados e contrato de adapter homologado.
rest-json-v1 é um contrato REST específico com POST/GET; SOAP/gRPC/arquivos legados exigem adapter e qualificação próprios, não estão automaticamente disponíveis.
GET final é servido da custódia do Hub; webhook entrega a mesma representação congelada.

## Autoridade, disponibilidade e latência
Reduzir dependência no caminho quente com projeção autorizada e pools reutilizados. DB durável não pode ser substituído por memória/Redis indisponíveis mantendo a mesma garantia de aceite.
Quando a autoridade necessária não está disponível, recusar antes de confirmar custódia ou usar protocolo alternativo de autoridade explicitamente qualificado. Aceites anteriores continuam obrigações recuperáveis.
Separar latência de autenticação/cache, admissão SQL, espera de permit, rede/provedor, recibo e finalização. Definir budgets/SLO por contrato e evidência D-01/D-02, não inventar um p95 global.
Isolamento sob carga precisa teste de tenant agressor e saudável, filas/concorrência/quotas/pools limitados e autoscaling de dependências. Crescimento automático é dentro de quotas e envelope, nunca infinito.

## Decisões abertas
D-01 demanda; D-02 SLA/parcialidade/tardio; D-03 preços/franquias/estornos/saldo; D-04 dados/retenção/localização; D-05 plataforma/orçamento/região/gateway; D-06 provedores/contratos/capacidades; D-07 catálogo e financeiro.
Reconciliar P-01…P-11 com esses IDs sem apagar pendências. Defaults técnicos R3 valem apenas no escopo registrado. T-R2-01 exige solução/prova ou decisão normativa explícita, jamais declaração absoluta não demonstrada.
