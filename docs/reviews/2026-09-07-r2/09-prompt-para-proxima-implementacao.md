# Instrução de handoff para a próxima implementação

O texto abaixo pode acompanhar o pacote no agente que implementará a rodada.

---

Revise AGENTS.md e docs/openspec-docs do repositório antes de agir. Use como baseline os 98 requisitos de openspec/changes/hub-interoperabilidade-v4, os documentos de engenharia em docs e as nove changes r2-01 a r2-09 deste pacote. O snapshot revisado é a39d394b0d87185ed4cc3861c12ec45f2c302d9e; se HEAD mudou, faça diff primeiro e revalide os achados afetados. Não trate observações desse SHA como fatos de uma versão posterior sem verificar.

Leia primeiro docs/reviews/2026-09-07-r2/00-leia-primeiro.md, 01-relatorio-de-revisao.md, 02-rastreabilidade-98.csv e o desenho da change que será implementada. A revisão é documental: nenhuma tarefa R2 foi implementada ou autorizada em produção por esse pacote.

Execute na ordem de dependências descrita em 08-pendencias-migracao-e-roteiro.md. Prepare testes/fixtures de r2-09 desde o início. Priorize P0 de identidade, custódia, resultado externo, concorrência, deadline, credencial e saldo antes de ampliar exposição. Corrija o núcleo e entregue jornadas reais do console; não encerre a rodada apenas com novos formulários, mocks, comentários ou manifests de referência.

Preserve as cinco autoridades Atlas/Órbita/Cometa/Pulsar/Libra. Implemente as capacidades em papéis de runtime isolados quando exigido; não adicione microserviços ou troque stack para mascarar gaps. Contratos e migrations precedem UI consumidora. Faça testes de falha/concorrência com DB/broker reais de ensaio. Compilar não comprova requisito operacional.

Invariantes que não podem ser relaxados:

- tenant vem de identidade autenticada; leitor administrativo entre tenants é individual, auditado e somente leitura;
- todo aceite tem UUIDv7 e intenção durável; SYNC devolve final na conexão quando elegível e não muda silenciosamente para 202;
- resposta é a do provedor, validada e conservada; GET e webhook usam os mesmos bytes completos da versão do cliente;
- ACK somente depois de efeito/obrigação e inbox duráveis; duplicata não duplica efeito nem cobrança;
- posse/epoch cercam execução e recuperação; UNKNOWN exige reconciliação, sem novo submit cego;
- TTL não renova prazo; commit depois do deadline não vira sucesso; recibo tardio fica conservado sem reabrir final;
- credencial efetiva é a do binding/cliente; mTLS/OAuth reais, sem referência usada como senha; Redis é dispensável;
- contratos de compra/venda e perfis são versionados; saldo considera capturas/holds; dinheiro não usa float;
- payload grande usa FileRef validada; dados/obrigações não são apagados por importação ou reset implícito;
- escalabilidade respeita capacidade externa, pools, células e recursos contratados; métricas/HA precisam ensaio.

T-R2-01 exige prova técnica do ponto de confirmação durável no deadline, incluindo commit lento; não considerar resolvido por now() no UPDATE ou timer a cada segundo. T-R2-04 exige quota externa agregada entre réplicas/células. Se a tecnologia escolhida não satisfizer o contrato, registre o bloqueio e proponha decisão explícita; não reduza a especificação para fazer os testes passarem.

Não invente aprovação de P-01 a P-11, não use credenciais de produção, não execute importação destrutiva nem aplique infraestrutura remota sem autorização correspondente. Use fixtures sintéticas isoladas. Ausência de informação comercial bloqueia ativação do perfil dependente, mas não interrompe correções genéricas já especificadas.

Para concluir uma tarefa: atualizar seu checkbox somente com evidência; registrar baseline/R2, cenário, SHA, comando, ambiente, resultado e link. Atualizar TRACEABILITY/IMPLEMENTATION_AUDIT com diferenças reais e manter risco residual explícito. Não arquivar a change v4 inteira como implementada. Valide OpenSpec conforme versão de ferramenta registrada e rode os gates pertinentes à change, incluindo jornadas UI com backend real. Entregue resumo de mudanças, testes, migração/rollback e pendências restantes.

---
