# Jornadas integradas obrigatórias
As jornadas complementam os 45 cenários R6 e os 783 anteriores; não reduzem o inventário a smoke tests.

| Jornada | Atores/dados | Oráculo e falhas |
|---|---|---|
| Cadastro até consumo | Cliente A/B, aplicação, serviço, conta, binding, compra/venda, perfil e oferta | Publicação valida referências; consumo usa exatamente versões e credenciais elegíveis. Testar suspensão e expiração em múltiplas réplicas. |
| SYNC e recuperação | Serviço SYNC, chave idempotente, timeout após efeito | Final na mesma chamada quando bem-sucedido; UUIDv7 consultável; crash/UNKNOWN sem SUBMIT duplicado ou conversão implícita para ASYNC. |
| ASYNC callback + polling | Provedor com operação demorada e callback duplicado | Uma conclusão pública; polling saudável até prazo; resultado tardio custodiado internamente e SLA aferível sem alterar final. |
| Produto composto | A e B com contratos distintos; DAG com falha parcial | Paralelismo só independente; slots corretos; custos por etapa e venda do produto; consolidação conforme oferta; compensação depois de final público. |
| Webhook cliente e GET | Contrato legado customizado e destino versionado | Comparar bytes/schema da mesma representação final; falha após envio e retry não mudam destino nem resultado. |
| Finanças e recuperação | Reserva, consumo, franquia, fechamento, fato tardio e quarentena | Valores exatos; completude antes de fechar; replay inválido continua pendente; ajuste não altera exportação fechada. |
| Console com catálogo grande | Mais de 50 referências e item selecionado em outra página | Criar/editar produto e rotas, pesquisa, versão, conflito ETag, mapping inválido, timeout após commit/reload e sessão expirada. |
| Segurança | Cliente A/B e desenvolvedor nominal global com MFA | Controles positivos e negativos por endpoint/tabela; sem acesso cruzado por cursor, ID, cache, jobs ou recursos auxiliares. |
| Arquivos/restore | Objetos grandes, versões, pins e protocolo pendente | Uploads/downloads escopados, streaming, retenção e restore de bytes/versionamento; reconciliar efeito externo antes/depois da retomada. |
| Escala e operação | Dois tenants, um provedor degradado, HPA/KEDA e observabilidade | Contenção seletiva, métricas bilaterais, recuperação de permits, falha de pod/nó/assinatura SNS e evidência sem perda de custódia. |
