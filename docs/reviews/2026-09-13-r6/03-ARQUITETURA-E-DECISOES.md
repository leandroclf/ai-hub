# Arquitetura e decisões da rodada R6
Base analisada: a540b40007fe6b8ed523e17afe00e96ff8f8ad50. A proposta preserva as tecnologias existentes; seleção nova exige evidência de limitação da atual.

## Responsabilidade, tecnologia e justificativa
| Componente | Tecnologia/comunicação | Persistência/autoridade | Motivo e limite |
|---|---|---|---|
| Atlas / catálogo | Go; HTTP/JSON de administração e projeções | PostgreSQL hub_control; publicações versionadas | Mesma stack do backend, transações de publicação e contratos estruturados. Não colocar lookup remoto obrigatório em cada hit de oferta. |
| Órbita / protocolo e produtos | Go; HTTP direto SYNC e intents/fatos assíncronos | PostgreSQL hub_core: protocolo, plano, etapas, inbox/outbox | Go atende concorrência I/O com goroutines limitadas; SQL suporta CAS/leases/transações. Não confundir goroutine com ausência de espera ou com garantia de throughput. |
| Cometa / provedores | Go; adapters REST/HTTP e capacidades explicitamente homologadas | Operações/tentativas/callbacks/capacidade no core; segredos no cofre | Isola contrato externo, pooling de conexões, deadlines e UNKNOWN. Não presumir que todos os provedores possuem idempotência ou polling por chave. |
| Pulsar / webhook cliente | Go; HTTP webhook, fatos via SNS/SQS | Deliveries, destinos versionados, recibos e outbox | Preserva callback sem acoplar disponibilidade do cliente à finalização. Entrega ao menos uma vez requer dedupe no destino. |
| Libra / financeiro | Go; fatos via SNS/SQS, API administrativa HTTP | PostgreSQL hub_finance: fatos, ledger, reservas, períodos, quarentena | Decimal exato e transações para auditabilidade. Não usar float64 para dinheiro ou saldo em Redis. |
| Console administrativo | React/TypeScript/Vite; APIs HTTP por domínio | Autoridade no backend; estado de edição e intenção local sem segredos | Ecossistema já adotado; tipos TS não validam resposta em runtime. Jornadas precisam guards e testes browser. |
| Gateway | Kong configurado; HTTP/TLS/OIDC | Configuração versionada; identidade central | Centraliza borda e rate limiting, mas não substitui autorização por recurso no serviço. Edição/licença é decisão D-05. |
| Broker | SNS/SQS; LocalStack somente no laboratório | Filas/assinaturas e retenção por perfil | Mantém stack já implementada e fan-out por obrigação. Semântica depende de custódia antes do ACK, consumidores idempotentes e topologia real. |
| Objetos/cofre/cache | S3/Secrets Manager; Redis derivado e L1 | S3 versionado e referências duráveis; secrets por conta/binding/versão | Evita payload grande em fila/memória. Redis não pode ser requisito de correção. Cofre indisponível permite apenas credencial previamente válida dentro da política; não inventar segredo fallback. |
| Plataforma e observabilidade | Compose; kind; Kubernetes/EKS por perfil; Prometheus/Loki/Grafana | Volumes/serviços duráveis e configuração reproduzível | Compose para execução local, kind para ensaios de cluster; EKS depende de orçamento e D-05. Métricas/logs/traces verificam objetivos, não os garantem sozinhos. |

## Fronteiras de persistência
Control e Finance possuem autoridades distintas. Órbita, Cometa e Pulsar compartilham o domínio core hoje: documentar ownership por tabela e permissões, evitando declarar falsamente banco por microserviço. R6 não manda particionar bancos sem plano de custódia; exige isolamento, transações corretas e budgets de conexão.

Outbox precisa da mesma transação do estado que torna a obrigação devida. Redis nunca substitui essa transação. Se todas as autoridades duráveis elegíveis ficarem indisponíveis, não é possível prometer aceite sem perda e continuidade irrestrita ao mesmo tempo: recusar novas admissões sem confirmação de custódia, conservar obrigações anteriores e continuar consultas/resultados onde existir cópia autorizada e consistente. Qualquer journal alternativo exige desenho explícito de autoridade, replay e fencing; memória/cache não serve.

## Modalidades e resultado único
- Provedor SYNC → cliente SYNC: resposta final na mesma requisição dentro do orçamento contratual; custódia e UUIDv7 presentes. Filas não podem transformar silenciosamente esse modo em resposta 202. Se não concluir, retornar erro contratual e protocolo consultável; ASYNC/AUTO somente quando solicitados e elegíveis.
- Provedor SYNC → cliente ASYNC: aceite durável, worker executa, final é persistido e entregue por webhook ou GET.
- Provedor ASYNC: configuração de polling/callback por capacidade homologada; ambos podem concorrer pela mesma conclusão. Polling saudável não inicia TTL de indisponibilidade.
- Final tardio: conservar evidência interna para reconciliação/SLA/financeiro, sem substituir a resposta pública já encerrada.
- GET do protocolo e corpo enviado ao webhook usam a mesma representação de saída congelada para aquele cliente/versão; isso não significa que o input do GET seja igual ao body do webhook.

## Baixa latência e escalabilidade
Definir latência adicional do Hub separadamente da latência total do provedor. Medir p50/p95/p99 de admissão, despacho, conclusão e consulta por coorte. Não aprovar números de SLA sem D-01/D-02.

Usar projeções válidas L1, conexões reutilizadas, streaming de objetos, filas separadas por obrigação e concorrência limitada por tenant/produto/rota. Paralelizar somente etapas independentes. Evitar joins/consultas ao control plane no caminho quente quando snapshot basta. Reservar capacidade para STATUS, compensação e webhook para que SUBMIT não consuma todo o orçamento.

Escalar exige CPU/memória, limites de conexões PostgreSQL, IOPS, partições/filas, quotas externas, nós e limites financeiros. HPA de pods isoladamente não atende crescimento irrestrito. Onboarding automático atribui célula e budgets dentro do envelope provisionável; esgotamento produz contenção seletiva e alerta. O requisito operacional é ausência de intervenção por cliente dentro desse envelope, não capacidade infinita.

## Segurança administrativa
Clientes nunca consultam protocolos alheios. Administração global usa usuários nominais de desenvolvedores, MFA, papel/escopo específico, motivo e auditoria; não conta compartilhada. Permissão global deve abranger o recurso concreto, inclusive quarentena/capacidade, e não ser inferida de qualquer finance:read. Workers têm identidade distinta de humanos.

## Decisões externas
| ID | Definição necessária | Responsáveis | Regra de continuidade |
|---|---|---|---|
| D-01 | Volumes, picos, arquivos e duração | Negócio e Engenharia | Implementar gerador/capacidade parametrizados; fixture sintética não é demanda aprovada. |
| D-02 | SLAs, parcialidade, timeout e resultado tardio | Produto e Operações | Implementar políticas e relógios; aprovação dos valores antecede contrato produtivo. |
| D-03 | Receita/custo, franquias, estornos e saldo estrito | Comercial e Financeiro | Implementar mecanismos exatos e cenários locais; não fixar política comercial implicitamente. |
| D-04 | Retenção, dados permitidos e região | Responsáveis pelos dados e Segurança | Fixture sintética; parâmetros de dados reais dependem de definição formal. |
| D-05 | AWS/EKS, orçamento, gateway e continuidade regional | Arquitetura e Plataforma | Gerar IaC/perfis reviewáveis; não provisionar serviço pago nem declarar RPO zero sem prova. |
| D-06 | Equivalência, idempotência, polling e callback | Integrações e Produto | Adapter deve declarar capacidades; ausência exige contenção/reconciliação, não reenvio cego. |
| D-07 | Catálogo inicial e integração financeira | Produto e Financeiro | Catálogo de fixture e contratos de exportação versionados; homologação de produção continua separada. |
| T-R2-01 | Decisão técnica pendente registrada nas rodadas anteriores | Dono identificado no registro original | Recuperar texto e evidência original; não inventar significado nem aprovação. |

Não há evidência neste snapshot de aprovação integral dessas decisões. O agente deve verificar registros formais antes de mudar status.
