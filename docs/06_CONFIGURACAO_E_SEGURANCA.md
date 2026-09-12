# 06 · Configuração, segurança e manutenção

## CFG-01 · Portal e API administrativa — proprietário: Atlas / Produto

O plano de controle deve permitir cadastrar, buscar, validar, versionar, publicar, suspender e descontinuar clientes/aplicações, provedores/contas, serviços, produtos, planos, contratos, rotas, destinos, políticas de polling, limites e retenção. Cada tela/capacidade informa campos obrigatórios, unidade, faixa válida, dependências e impacto em pedidos novos/em voo.

Integração existente deve ser configurável sem deploy para endpoint homologado, mapeamento suportado, segredo referenciado, modo de obtenção do final, ativação de polling, intervalos dentro de limites, política de rota e preço. Nova semântica, protocolo, algoritmo ou transformação fora do conjunto seguro exige desenvolvimento e homologação; “configurável” não autoriza código arbitrário no portal.

## CFG-02 · Publicação governada — proprietário: Atlas / Segurança

Ciclo: rascunho → validação estrutural → simulação com dados sintéticos → aprovação conforme risco → publicação versionada → observação → rollback por nova publicação. A validação deve verificar schemas, grafo, conta, capacidades, contratos, preços, quotas, prazo e retenção. Publicar parcialmente um conjunto inconsistente é proibido.

Projeções nos componentes registram versão e data de sincronização. Pedido fixa uma versão completa; nó sem a versão necessária deve sincronizar ou recusar novas admissões, não misturar configuração antiga e nova. Atlas indisponível não impede execução com versão válida já conhecida. Revogação emergencial tem distribuição prioritária e independente da interface humana. A idade máxima da evidência de autorização é explícita por perfil, com 60 s como proposta inicial a qualificar; vencida essa validade sem atualização confiável, bloquear somente operações/escopos cuja autorização não possa ser comprovada. O prazo deve ser compatível com a continuidade aprovada em OPE-15, sem prolongamento silencioso durante falha.

Mudança de polling de operação em voo exige ação explícita, faixa segura e trilha do motivo. Revogação de segredo interrompe novos usos; rotação coordenada mantém a mesma conta e respeita SEG-05, inclusive operações em voo. Alteração de preço não acompanha automaticamente a troca de segredo. Rollback de configuração cria nova versão referenciando conteúdo anterior; não apaga eventos de publicação ou pedidos já executados.

## CFG-03 · Manutenção e diagnóstico — proprietário: Operações / Engenharia

O portal operacional deve pesquisar por protocolo, tenant, cliente, conta de provedor, provider_request_id, período, estado e código de erro, respeitando permissão. Exibir linha do tempo com versões, passos, tentativas, resultado, entregas, uso e divergências; payload sensível só por acesso excepcional auditado.

Ações distintas: retomar timer; republicar evento; reprocessar inbox em quarentena; reconciliar operação; reenviar webhook; revisar resultado; reexecutar serviço. Cada ação possui escopo, pré-condição, motivo, responsável e efeito financeiro visível. Reexecução cria novo protocolo; replay técnico preserva identidade. Não oferecer um botão genérico “reprocessar tudo” que possa multiplicar chamadas e custos.

## SEG-01 · Identidade e autorização — proprietário: Segurança

Clientes de máquina usam OAuth2 client credentials como padrão, com JWT de emissor/audiência/escopos homologados. Usuários administrativos usam OIDC e MFA; sessões e privilégios têm prazo. API keys só são admitidas em integração legada explicitamente aprovada, com rotação e escopo. Comunicação HTTPS entre aplicações usa identidade de workload e mTLS; broker, banco e S3 usam TLS e os mecanismos de autenticação/menor privilégio suportados por cada serviço.

### MFA do laboratório local R2

O usuário administrativo sintético `operadora-a` usa a senha de fixture definida
no Compose e TOTP de 6 dígitos, SHA-1, período de 30 segundos. Para cadastro
manual em Google Authenticator, Authy ou aplicativo compatível, use o segredo
Base32 público abaixo, sem espaços:

`IFES2SCVIIWVEMRNJVDECLKLIVMS2MBR`

`JBSWY3DPEHPK3PXP` é um segredo de exemplo do TOTP e não funciona nesta
fixture. Para conferir o código vigente no terminal, sem emitir token, use:

```bash
python3 hub/deploy/r2/scripts/token.py --otp
```

Digite o campo `otp` imediatamente no formulário; `window_seconds` indica o
tempo restante da janela atual.

Se o realm tiver sido reiniciado com volume persistente ou uma tentativa manual
tiver alterado a credencial, reconcilie a fixture antes de abrir o portal:

```bash
docker compose -p ai_hub_r3qual -f hub/deploy/r2/compose.yaml \
  run --rm identity-reconcile
```

Depois abra uma janela anônima ou faça logout completo, acesse
`http://localhost:13000/` e inicie uma transação nova. Não submeta o mesmo OTP
novamente depois de uma rejeição: o código é válido por janela e não deve ser
reutilizado.

O valor Base32 é a representação de cadastro; o Keycloak armazena a chave ASCII
equivalente para manter compatibilidade com a credencial importada do fixture.
O gerador de token e a prova de compatibilidade derivam os mesmos bytes a partir
do segredo Base32, enquanto os smokes de navegador usam essa chave equivalente,
para que o acesso manual e os gates usem exatamente o mesmo algoritmo.

Após reiniciar o Compose, sempre comece em `http://localhost:13000/` e clique em
**Entrar**. Não reutilize uma URL antiga de `login-actions`; ela pertence a uma
transação OIDC expirada e causa `Invalid authenticator code` mesmo com um OTP
correto. Se a conta tiver sido alterada localmente, execute a reconciliação da
fixture para restaurar a credencial sintética antes de testar novamente.

Gateway elimina cabeçalhos de identidade fornecidos pelo cliente e transmite contexto confiável. Serviços validam identidade de workload e autorização de tenant/recurso; não confiam só na rede interna. Permissões separam administrar catálogo, contrato, credenciais, executar, consultar, diagnosticar, ajustar financeiro, aprovar e confirmar liquidação. Operador de runtime não pode editar ledger publicado.

## SEG-02 · Destinos, callbacks e dados — proprietário: Segurança / Integrações

URLs de provedores e clientes devem ser registradas com domínio, finalidade e métodos permitidos. Prevenir SSRF: bloquear destinos de metadados e redes internas não autorizadas, revalidar resolução DNS, controlar egress, impedir redirecionamento para destino não validado e não seguir URLs retornadas pelo provedor sem política. Certificados TLS são validados; mTLS por contrato quando exigido. A recomendação segue o modelo de proteção de [SSRF da OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html).

Segredos ficam em cofre, nunca em eventos, tabelas de configuração em claro, logs ou arquivos versionados. Chaves têm dono, escopo, expiração, rotação e revogação; assinaturas de callback suportam key_id e janela de sobreposição controlada. Dado pessoal não vira label de Prometheus. Conteúdo XML deve impedir resolução de entidades externas; uploads recebem controles por tipo/risco.

## SEG-03 · Isolamento e auditoria — proprietário: Segurança / QA

Separar tenants em consulta, objeto, contrato, consumo, dashboards e ações operacionais. Separar ambientes por identidades, bancos, filas, buckets, credenciais e domínios; dados reais em não produção exigem autorização e proteção compatível. Testar autorização negativa por troca de IDs, não apenas por navegação do portal.

Auditar publicação, suspensão, alteração de destino, uso excepcional de dados, replay, reexecução, conciliação, ajustes e exportações. Registro inclui ator, instante UTC, motivo, correlação e diferença segura de configuração. Auditoria de negócio durável é separada de logs Loki; perda ou expurgo de logs não pode eliminar evidência de cobrança.

## CFG-04 · Configuração dos novos limites — proprietário: Atlas / Produto / SRE

| Família | Configuração obrigatória | Validação de publicação |
| --- | --- | --- |
| Cliente/oferta | contrato técnico entrada/saída, client_sla_seconds, versão, criticidade e isolamento | Prazo e representação homologados; perfil legado inclui falha e expiração |
| Modalidade | modos autorizados, sync_wait_seconds, margem de retorno e reserva final | SYNC nativo qualificado, deadline compatível com conexão; não converter para AUTO |
| Credencial | modo, binding, tenant dedicado quando aplicável, conta, secret_ref, validade e settlement_party | Escolha inequívoca, conta autorizada, sem fallback implícito |
| Expansão | perfil de célula, capacidade reservada, horizonte, orçamento e quotas autorizadas | Fluxo de ativação automático e recursos prontos antes de publicar tráfego |
| Retry | retry_ttl_seconds, max_attempts, backoff/jitter e retry_budget | Inteiros/unidades explícitos, zero sem retry, sem relógio renovado |
| Passo/provedor | provider_sla_seconds, marco de início, enforcement, request_timeout_seconds | Compatível com deadline do pai, margem de finalização e idempotência |
| Polling | modo, primeiro agendamento, min/max intervalo, fallback, deadline e quota | Primeiro poll/fallback não pode estar depois do prazo útil; consultas pagas previstas |
| Pressão | domínio de capacidade, rate_min/start/max, concurrency_min/start/max, limite async pendente, janelas, erro alvo, latência alvo, cooldown | Sem limites negativos, máximos ≥ mínimos; teto contratual prevalece; versão publicada e trilha de alteração |
| Isolamento | cell_id, classe compartilhada/dedicada, quotas e reservas, autoridade financeira | Um tenant/contrato tem rota/autoridade ativa definida, sem duplicação na migração |
| Administração | papel individual, ambiente, acesso entre tenants, validade e restrição de dados | Nenhuma credencial de cliente consegue elevar-se a administradora |

Todos os prazos configuráveis aqui são inteiros em segundos; métricas podem manter precisão subsegundo. client_sla_seconds e provider_sla_seconds devem ser positivos; retry_ttl_seconds pode ser zero; max_attempts inclui a primeira tentativa e deve ser ≥1; reserva de finalização é não negativa e menor que o prazo do cliente. Taxas e janelas de erro têm unidades e faixas explícitas; concorrência usa inteiros. Circuito aberto pode zerar tráfego produtivo mantendo apenas sondagens autorizadas. O portal mostra valor resolvido e sua origem: contrato do cliente, produto, serviço, conta/provedor ou teto da plataforma. Não permitir override silencioso do cliente que amplie sua autorização ou prazo acima do contrato. Alteração em voo não estende deadline; exceções operacionais exigem novo protocolo/contrato, não reescrita histórica.

Antes de publicar, simular caminho crítico, retry, polling, SLA e incidência financeira com casos sintéticos. Simulação declara premissas e não comprova capacidade de produção. Estado do controlador adaptativo é observável, mas não é edição contínua do contrato comercial. Permitir pausa e limite emergencial menor com motivo/auditoria; retomar começa conservadoramente, não restaura instantaneamente pico antigo.

## SEG-04 · Administração entre tenants para desenvolvedores — proprietário: Segurança / liderança técnica

Clientes permanecem estritamente limitados ao próprio tenant em protocolo, resultado, objeto, SLA, consumo e entrega. Conhecer UUID, alterar header, usar URL de outra célula ou indicar tenant_id no corpo não confere acesso. Operação inexistente e operação de outro tenant devem ter resposta que não revele existência.

Criar perfil administrativo **hub_protocol_reader**, atribuído a identidades humanas individuais dos desenvolvedores autorizados do hub. Esse perfil pode localizar e consultar protocolos de qualquer tenant, inclusive histórico, etapas, erros, entregas e fatos de SLA, por API/rota administrativa própria. Não se cria conta humana compartilhada: cada acesso precisa identificar seu autor. OIDC, MFA, grupos aprovados, expiração/revisão de concessão e separação entre produção/não produção são obrigatórios.

Consultar todos os tenants não concede automaticamente segredos, payload integral, edição, replay, reexecução ou poderes financeiros. Por padrão, dados sensíveis são mascarados; papel adicional de leitura de conteúdo, com justificativa e prazo, atende investigação autorizada. Download, busca e visualização são auditados com identidade, tenant alvo, protocolo, horário, finalidade e ambiente. Não aceitar função administrativa por flag em JWT emitido ao cliente ou header arbitrário.

A aplicação autoriza explicitamente a consulta entre tenants e registra o alvo. Não usar credencial de superusuário PostgreSQL nem desativar globalmente RLS para disponibilizar o portal. O modelo de banco/serviço deve oferecer política de leitura administrativa controlada; todos os testes negativos de cliente continuam obrigatórios. Revogar o grupo deve interromper novas consultas administrativas dentro da política de atualização de identidade.

Relatórios globais usam projeções de leitura e paginação com limites; não disparam varreduras irrestritas nem fan-out ilimitado nos bancos de execução. Busca entre células tem concorrência limitada e evidencia fonte/atualidade. Se apenas parte das células responde, apresentar resultado parcial administrativo com indicação de incompletude, sem declarar protocolo inexistente globalmente.

## CFG-05 · Gestão de credenciais Hub–cliente–provedor — proprietário: Atlas / Segurança / Integrações

O portal deve administrar metadados e ciclo de vida do vínculo de credencial. O segredo é escrito no cofre por fluxo restrito, sem leitura posterior em claro pela interface. O gestor visualiza identificador, tipo de autenticação, titular, conta externa, escopo, ambiente, validade, último teste, estado de rotação e uso recente autorizado. Máscara não pode revelar parte suficiente para reutilizar o segredo.

| Campo/regra | Definição obrigatória |
| --- | --- |
| credential_mode | SHARED_HUB ou TENANT_DEDICATED; não derivado da ausência de campo |
| Relacionamento | provider_id, provider_account_id, serviço/vínculo, ambiente; tenant_id obrigatório e único no modo dedicado |
| Autenticação | API key, OAuth2 de máquina, certificado mTLS ou mecanismo homologado; audience, escopos e destino permitidos |
| Referência | credential_binding_id e versão de metadados, secret_ref, política de versão e rotação; valores secretos só no cofre |
| Estado | RASCUNHO, EM_VALIDACAO, ATIVO, EM_ROTACAO, SUSPENSO, REVOGADO, EXPIRADO |
| Continuidade | refresh, idade/validade máxima em cache, expiração externa, sobreposição e ação após perda do cofre |
| Capacidade | domínios sobrepostos da conta/provedor, capacidade compartilhada ou dedicada; não uma quota nova por versão de chave |
| Economia | titular contratual, settlement_party, contrato aplicável e escopo de extrato |

O cadastro valida que o operador pode administrar aquele tenant/provedor, que a conta pertence ao escopo e que não existe resolução ambígua. Teste de credencial usa endpoint sem efeito, se fornecido; um teste que execute serviço precisa de operação de homologação explícita e de seu custo identificado. Falta de endpoint de teste não autoriza executar pedido real invisível.

Ativação exige credencial válida, contrato coerente, isolamento de segredo, política de rotação e teste de acesso. O portal deve permitir suspender apenas o vínculo afetado, localizar protocolos em voo e visualizar repercussão em SLA, polling e custo. Troca de segredo não troca automaticamente de conta externa. Cliente pode fornecer/rotacionar sua credencial por permissão específica, mas não escolher secret_ref arbitrário nem ler cofre ou chave compartilhada do Hub.

## CFG-06 · Onboarding e capacidade como serviço interno — proprietário: Atlas / Plataforma

Fluxo declarativo: RASCUNHO → VALIDADO → RESERVANDO_CAPACIDADE → PROVISIONANDO quando necessário → QUALIFICANDO → ATIVO; falha fica PENDENTE_COM_MOTIVO e preserva recursos/progresso. Para cliente em célula pronta, provisionamento novo pode ser dispensado. Validações comerciais e de segurança continuam existindo; aprovação de contrato não equivale a abrir chamado para criar pod/banco manualmente.

Atlas calcula placement usando perfil, região permitida, criticidade, carga prevista, capacidade já reservada e limites de conta de provedor. Reserva capacidade de forma concorrente e idempotente no controle para impedir duas ativações de prometerem a mesma folga. Políticas aprovadas disparam reconciliação de infraestrutura por Crossplane e de réplicas/nós por seus controladores próprios. Credenciais e configurações são pré-distribuídas somente aos workloads elegíveis.

Antes de ATIVO: dependências prontas, schema compatível, probes, quotas, credenciais, conectividade, objetos, telemetry/alertas, capacidade N−1 exigida e ensaio sintético do perfil. Os testes de prontidão não substituem a qualificação prévia do template. O processo deve ser retomável após crash; não criar contas/filas duplicadas ou publicar rota antes de persistir o estado aprovado.

Novo provedor com protocolo já homologado usa cadastro/adaptador e capacidade próprios. Protocolo/SDK ou efeito novo continua exigindo engenharia, não infraestrutura manual de rotina. Teto de nuvem, ausência de região permitida, restrição comercial e orçamento esgotado devem aparecer antes da saturação como impedimentos explícitos. Clientes ativos conservam sua reserva enquanto novas ativações aguardam. Desativar cliente drena/reconcilia obrigações e aplica retenção; não destrói automaticamente recursos compartilhados.

## SEG-05 · Resolução, rotação e isolamento de segredo — proprietário: Segurança / Cometa

Resolução determinística: obter tenant da identidade confiável; fixar contrato técnico/comercial e rota; resolver vínculo elegível de conta/serviço/ambiente; exigir o modo previsto; validar estado/validade e permissão; obter a versão de segredo autorizada; registrar referência da versão efetivamente utilizada na tentativa. O cliente não escolhe outra conta por parâmetro. Binding dedicado sem tenant correspondente é inválido; binding compartilhado só atende tenants autorizados pela política.

Se o modo é TENANT_DEDICATED e a credencial está ausente, expirada ou revogada, o Hub não usa SHARED_HUB nem credencial de outro tenant como fallback. Suspende o vínculo/recusa antes do envio ou conserva a pendência segundo prazo, sem alterar pagador. Failover entre provedores exige também um vínculo de credencial explicitamente elegível para o mesmo cliente. A indisponibilidade de uma credencial dedicada não deve desligar conectores de outros vínculos saudáveis.

Cache de segredo/tokens é local ao workload autorizado, limitado e separado por ambiente, provedor/conta, binding, versão, audience e escopos. Não armazenar no Redis comum, eventos, backups de configuração, logs, traces ou interface. Proteger memória, dumps e acesso do processo; a chave precisa existir transitoriamente em memória para uso e não se promete eliminação perfeita de todas as cópias por garbage collection. Pools mTLS separam certificado/conta; token OAuth não pode ser reutilizado em binding diferente. Não persistir token temporário como chave de correlação.

Rotação: preparar versão nova na mesma conta; testar; pré-aquecer workloads; publicar política/versionamento; usar nova versão em novas tentativas; manter sobreposição somente quando provedor e segurança permitirem; encerrar a anterior após validação das operações em voo. Polling e fetch conservam provider_request_id, conta e operação; podem usar nova versão autorizada da mesma conta. Se a operação exige chave antiga e ela for revogada, registrar impedimento e reconciliar, sem apontar a consulta para outra conta. Revogação emergencial prevalece sobre conveniência operacional e não espera terminar o TTL de retry.

Cofre fora do ar: material já obtido pode ser usado apenas até o menor limite de expiração externa, validade local e política de revogação. Atualização é antecipada e com jitter; refresh único evita tempestade. Instância sem material válido não atende aquele binding; preservar réplicas aquecidas e não retirar todas durante rotação/manutenção. A documentação do cache Go do Secrets Manager informa que ele não inclui invalidação automática nem endurecimento de segurança; a política acima precisa ser implementada/qualificada pelo Hub. [AWS — cache de segredos em Go](https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_cache-go.html).

Callbacks externos são autenticados/correlacionados na conta/binding original, inclusive durante sobreposição. Chaves de assinatura de webhook enviado ao cliente pertencem a outro vínculo e ciclo de rotação. Administrador hub_protocol_reader pode ver metadados permitidos e diagnósticos; não obtém valor secreto. Ensaios obrigatórios: tentativa cruzada de tenant, rotação em voo, cofre indisponível com processo frio/quente, revogação e troca indevida de conta.
