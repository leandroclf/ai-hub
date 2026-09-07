# Hub de Interoperabilidade — Engenharia de Requisitos v4.0

**Nome de trabalho:** Constelação. **Data:** 6 de setembro de 2026. **Status:** especificação para revisão técnica e aprovação de negócio. **Escopo deste pacote:** requisitos, regras, responsabilidades, decisões arquiteturais, modelo lógico de dados e critérios de qualidade. Não contém implementação, scripts de implantação ou contratos executáveis.

## Objetivo

Oferecer um ecossistema de interoperabilidade entre clientes e provedores, com portfólio de serviços e produtos, processamento durável, resultados sob custódia do hub, contratos comerciais rastreáveis e operação configurável. Disponibilidade, escala e baixa latência são objetivos medidos por capacidade e SLOs; não decorrem automaticamente de microserviços ou da linguagem escolhida.

## Organização dos artefatos

| Arquivo | Conteúdo | Leitores principais |
| --- | --- | --- |
| 01_DOMINIO_E_PORTFOLIO.md | Domínio, catálogo, agregação/composição, contratos de produto e roteamento | Produto, Integrações, Comercial |
| 02_ARQUITETURA_E_COMUNICACAO.md | Componentes, tecnologias, interfaces e fronteiras de consistência | Arquitetura, Engenharia |
| 03_PERSISTENCIA_E_DADOS.md | Bancos, propriedade, integridade, resultados, arquivos e retenção | Dados, Engenharia, Segurança |
| 04_EXECUCAO_E_INTEGRACOES.md | Estados, modos de atendimento, polling, callbacks e recuperação | Integrações, Engenharia, QA |
| 05_CONTRATOS_CONSUMO_E_FINANCEIRO.md | Compra de provedores, planos de clientes, medição, preços e conciliação | Comercial, Financeiro, Produto |
| 06_CONFIGURACAO_E_SEGURANCA.md | Portal, publicação, credenciais, isolamento e auditoria | Operações, Segurança |
| 07_DESEMPENHO_E_OPERACAO.md | Capacidade, ambientes, HA, autoscaling, observabilidade e runbooks | Plataforma/SRE, Engenharia |
| 08_QUALIDADE_E_ACEITE.md | Estratégia de testes, critérios de promoção e evidências | QA e responsáveis por domínio |
| 09_DECISOES_E_REFERENCIAS.md | Decisões, registro de gaps, compatibilidade, pendências e fontes primárias | Liderança técnica |
| MATRIZ_RASTREABILIDADE.csv | Relação requisito → regra → responsável → cenário → evidência | QA, gestão técnica |

A edição consolidada HTML contém os mesmos capítulos para leitura contínua e impressão. Os arquivos Markdown são as fontes editoriais. Não há duas especificações normativas diferentes. O pacote reúne requisitos identificados, cenários de aceite e decisões pendentes; as quantidades e a cobertura constam da matriz e do relatório desta revisão.

## Como interpretar

“Deve” e “não pode” indicam obrigação da especificação. “Meta proposta” indica valor inicial a qualificar e aprovar. “Pendente” identifica informação de negócio ou infraestrutura ainda não fornecida; não autoriza o implementador a inventá-la. Exemplos monetários e de capacidade são ilustrativos, não preços ou volumes reais.

Esta v4 substitui a v3 como especificação vigente. Todos os capítulos, a matriz, o diagrama, a edição consolidada e o relatório foram revistos. O pacote v3 e seus arquivos foram recuperados e conferidos antes da alteração. Este trabalho é exclusivamente documental: não afirma implementação, teste de carga ou publicação no GitHub. Exemplos de campos e fórmulas descrevem regras, não código.

## Invariantes obrigatórios

1. **Aceite durável:** o hub só confirma aceitação depois de persistir protocolo, pedido, versão de configuração e intenção de execução.
2. **Autoridade do resultado:** a consulta do cliente lê o hub; não dispara consulta nem reexecução no provedor, inclusive quando o protocolo ainda está pendente.
3. **Identidades distintas:** protocolo, passo, operação externa, tentativa física, evento, resultado, entrega e fato econômico têm identidades próprias e vínculos verificáveis.
4. **Convergência:** callback, polling e respostas imediatas concorrem para a mesma operação; somente transições válidas consolidam seu resultado.
5. **Entrega repetível:** repetição de mensagem ou webhook não implica repetir efeito de negócio nem cobrar novamente.
6. **Economia separada:** faturamento do cliente, custo do provedor e confirmação de entrega são dimensões independentes.
7. **Configuração histórica:** cada execução fixa as versões do produto e dos contratos aplicáveis. Alterações posteriores não reescrevem a história.
8. **Falha rastreável:** um pedido aceito terá conclusão contratual ou pendência operacional explícita. Não será abandonado porque expirou uma mensagem.
9. **Isolamento:** nenhum cliente acessa dados, credenciais, consumo ou objetos de outro cliente por conhecer um identificador.
10. **Limite da garantia:** processamento externo exatamente uma vez depende das capacidades do provedor. Resultado desconhecido bloqueia reexecução insegura e exige reconciliação.
11. **Prazo absoluto:** o aceite fixa deadline; retry, polling, fila, reinício e mudança de provedor não renovam o SLA do cliente.
12. **Final por SLA:** protocolo encerrado por prazo não é reaberto por retorno tardio. A evidência é preservada e segregada do resultado entregue.
13. **Contrato uniforme por cliente:** GET de protocolo e corpo enviado ao webhook usam a mesma representação contratada e a mesma versão final.
14. **Pressão adaptativa:** taxa e concorrência efetivas por domínio de capacidade do provedor variam com métricas, dentro dos limites técnicos e contratuais.
15. **Isolamento:** tráfego de um tenant não pode consumir a capacidade reservada de outro; a classe de isolamento e seu domínio de falha devem ser explícitos e testados.
16. **Administração:** clientes não cruzam tenants; desenvolvedores autorizados podem consultar protocolos de qualquer tenant por perfil administrativo individual, de leitura e auditado.

17. **SYNC real:** serviço nativamente síncrono elegível pode devolver seu final na mesma requisição; filas de execução e de resultado não são uma espera obrigatória desse caminho.
18. **Protocolo universal:** toda admissão, SYNC ou ASYNC, recebe UUIDv7 persistido; consultar novamente não repete a operação no provedor.
19. **Credencial vinculada:** toda operação resolve explicitamente conta, modo compartilhado ou dedicado ao tenant, vínculo, versão de segredo e responsável econômico. Credencial dedicada não cai silenciosamente para a compartilhada.
20. **Cache dispensável:** Redis não guarda a única cópia de protocolo, resultado, idempotência, saldo, outbox, agenda, autorização ou quota global. O runtime deve ser qualificado também com Redis desligado.
21. **Escala sem chamado rotineiro:** admissão de novos clientes/provedores e expansão usam perfis homologados e automação de capacidade. Cresce a quantidade de recursos, preservando a arquitetura; quotas físicas e contratuais continuam explícitas.
22. **Continuidade responsável:** fallback só usa informação válida, autorizada e durável quando a obrigação exigir. Não confirmar operação sem custódia, inventar resposta ou repetir efeito incerto para aparentar disponibilidade.

## Resultado e limites desta revisão

A referência passa a ter dois transportes de execução sobre a mesma máquina de estados: chamada interna direta para SYNC; agenda/filas duráveis para ASYNC. Os fatos de domínio, a entrega ao cliente e a apuração pós-paga continuam desacoplados. O plano de controle administra credenciais e provisionamento automático; não participa de cada pedido.

Disponibilidade é requisito crítico, inclusive durante manutenção. Isso exige capacidade de reserva, isolamento, redundância, recuperação automática e ensaios por perfil de criticidade. Redis e consultas repetitivas de configuração saem das dependências obrigatórias do atendimento. A custódia transacional do pedido e do resultado permanece necessária. Se todas as cópias autoritativas exigidas ficarem inacessíveis, a arquitetura não pode oferecer simultaneamente novo aceite irrestrito e prova de ausência de perda. OPE-13 define continuidade por operação; OPE-15 define o gate para consumidores com impacto em vida/segurança.

Os 98 requisitos e cenários mínimos constam da matriz; a cobertura documental não certifica operação em produção. DEC-05 registra os gaps, a mitigação definida, a evidência ainda necessária e o responsável. O desenho regional de referência não deve ser comercializado como tolerante a qualquer falha ou como plataforma crítica já qualificada.

## Responsabilidade pela aprovação

Produto aprova semântica dos serviços e resultados; Comercial aprova condições de venda/compra; Financeiro aprova apuração, conciliação e liquidação; Arquitetura aprova fronteiras e tecnologias; Engenharia implementa; QA reúne evidências; Plataforma/SRE responde por operação e recuperação; Segurança e responsáveis pelos dados aprovam acesso e retenção. São papéis, sem atribuição nominal presumida. Cada decisão pendente tem um único responsável final indicado no capítulo 09.
