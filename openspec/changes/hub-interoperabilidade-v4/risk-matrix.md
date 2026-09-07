# Matriz de riscos consolidada — hub-interoperabilidade-v4

Esta matriz consolida, para fins de acompanhamento desta mudança, os 15 gaps de risco registrados em `decisoes-e-governanca/spec.md` (DEC-05, originados em `docs/09_DECISOES_E_REFERENCIAS.md`) e as 11 decisões pendentes (P-01 a P-11). Nenhum item aqui é encerrado por esta mudança: "Resolvido no desenho" significa que a regra está definida e a contradição documental foi removida — não que a mitigação foi implementada ou testada.

## Gaps de risco (DEC-05)

| Gap | Severidade | Risco identificado | Mitigação de desenho (capability) | Responsável final | Condição de encerramento |
|---|---|---|---|---|---|
| GAP-01 | Alta | Fila obrigatória compromete contrato SYNC | Transporte direto com commits (`arquitetura-e-comunicacao` COM-01/06; `execucao-e-integracoes` EXE-14) | Engenharia | Final na mesma conexão com broker fora; resolvido no desenho, qualificação pendente |
| GAP-02 | Crítica | Caminhos direto e fila duplicam efeitos | Intenção, posse, identidade e recuperação comuns (`execucao-e-integracoes` EXE-15) | Engenharia | Crash/duplicata/UNKNOWN sem reexecução indevida; qualificação pendente |
| GAP-03 | Crítica | Credencial de outro cliente ou conta compartilhada usada indevidamente | Resolução explícita sem fallback implícito (`configuracao-e-seguranca` CFG-05, SEG-05) | Segurança | Testes cruzados, rotação e logs sem segredo; resolvido no desenho |
| GAP-04 | Alta | Redis vira ponto único ou causa avalanche na origem | Cache dispensável, timeout/bypass, origem dimensionada (`persistencia-e-dados` DAD-10) | SRE | Pico N−1 com Redis desligado e frio; qualificação pendente |
| GAP-05 | Crítica | "DB opcional" confirma pedido/final sem prova de custódia | Fronteiras duráveis e continuidade restrita (`persistencia-e-dados` DAD-09; `desempenho-e-operacao` OPE-13) | Arquitetura | Limite assumido; nenhuma aprovação pode chamar escrita sem autoridade durável |
| GAP-06 | Alta | HPA aumenta pods, mas nó/banco/célula exigem chamado | Automação de camadas e ativação com capacidade (`configuracao-e-seguranca` CFG-06; `desempenho-e-operacao` OPE-12) | Plataforma | Expansão completa sem intervenção no envelope aprovado; depende de P-11 |
| GAP-07 | Alta | Células/credenciais novas multiplicam quota externa | Orçamento global, concessões e pressão hierárquica (`desempenho-e-operacao` OPE-07/12) | Integrações | Múltiplas células, dono perdido e slots incertos; qualificação pendente |
| GAP-08 | Alta | Cofre/Atlas/identidade fora causam indisponibilidade ou autorização velha | Pré-aquecimento e validade de material/projeções (`configuracao-e-seguranca` CFG-02, SEG-05; `desempenho-e-operacao` OPE-13) | Segurança | Máxima defasagem aprovada e ensaios quente/frio; risco residual explícito |
| GAP-09 | Crítica | Migração perde idempotência ou cria duas autoridades | Placement/epoch, cópia e drenagem com corte exclusivo (`persistencia-e-dados` DAD-11) | Dados | Interrupção em cada fase sem split-brain; limite de corte dentro do SLA |
| GAP-10 | Crítica | Final no limite do SLA é aceito por relógio ou commit incorreto | Arbitragem terminal com prova temporal (`execucao-e-integracoes` EXE-11/14) | Engenharia | D−ε/D/D+ε e commit lento; garantia ainda não demonstrada |
| GAP-11 | Crítica | Failover de banco excede tolerância do consumidor | Perfil de criticidade e qualificação física (`desempenho-e-operacao` OPE-14/15) | Produto | P-10 e SRE comprovam máximo de interrupção; oferta crítica bloqueada se incompatível |
| GAP-12 | Crítica | Perda regional destrói dados aceitos ou cria split-brain | Topologia regional adicional só com autoridade e cópias adequadas (`desempenho-e-operacao` OPE-11/15) | Arquitetura | Depende de P-08; referência atual não qualifica RPO zero regional |
| GAP-13 | Crítica | Provedor executa, resposta some e não oferece reentrega/status/idempotência | UNKNOWN, contrato de recuperação e restrição de oferta (`execucao-e-integracoes` EXE-09; `desempenho-e-operacao` OPE-13/15) | Integrações | Depende de P-05; não vender garantia externa que o parceiro não sustenta |
| GAP-14 | Alta | Conta dedicada altera cobrança/pagador por inferência | Vínculo econômico explícito e fatos completos (`contratos-e-financeiro` FIN-11) | Financeiro | Três modelos reconciliados, incluindo polls e retorno tardio |
| GAP-15 | Crítica | Texto completo é confundido com disponibilidade comprovada | Gates, cenários, matriz e estado NÃO EXECUTADO (`qualidade-e-aceite` QUA-03/04/06) | Liderança técnica | Evidências da fatia/perfil, riscos residuais aceitos e responsáveis nominais |

**Bloqueadores priorizados de escopo:** GAP-02, GAP-03, GAP-05, GAP-09, GAP-10, GAP-11, GAP-12, GAP-13, GAP-15.

## Decisões pendentes (DEC-02)

| ID | Responsável | Informação necessária | Condição de bloqueio |
|---|---|---|---|
| P-01 | Plataforma | Conta, região, Kubernetes gerenciado ou infraestrutura própria, orçamento, registro OCI, licenças | Antes de IaC remota e ppd |
| P-02 | Produto | Catálogo real, volumes, fan-out, perfis legados, mix real, classes de isolamento | Antes de assumir SLO/capacidade |
| P-03 | Comercial | Contratos de compra/venda, tarifas, planos, marcos, faixas, parcialidade, riscos de failover | Antes de habilitar oferta comercial |
| P-04 | Responsável pelos dados | Retenção, finalidade, região permitida, direitos de custódia | Antes de dados reais/publicação |
| P-05 | Integrações | Idempotência, SYNC/polling/callback, credenciais compartilhadas/dedicadas, recuperação, equivalência | Antes de habilitar vínculo/failover |
| P-06 | Financeiro | ERP, layout/API, plano de contas gerencial, arredondamento, competência, liquidação | Antes de fechamento/exportação |
| P-07 | Produto | SLA bilateral, TTL em segundos, enforcement do provedor, reserva final | Antes de vender SLA |
| P-08 | SRE | Domínio de falhas, RPO zero contratado, confirmação entre regiões, RTO por célula | Antes de prometer ausência de perda regional/prd |
| P-09 | Liderança técnica | Responsáveis, perfis administrativos, aprovação da v4 | Antes do gate G0 da fatia |
| P-10 | Produto | Perfil de criticidade, responsável do sistema consumidor, interrupção máxima em segundos | Antes de ofertar uso com impacto em vida/segurança |
| P-11 | Plataforma | Envelopes de escala, quotas preautorizadas, reserva, horizonte de provisionamento | Antes do gate ppd de expansão automática |

## Como usar esta matriz

- Antes de avançar qualquer tarefa de `tasks.md` que dependa de um gap crítico ou de uma decisão pendente listada acima, confirmar que a condição de encerramento/bloqueio foi satisfeita ou aceitar o risco residual com responsável nominal.
- Esta matriz não substitui `MATRIZ_RASTREABILIDADE.csv` (98 requisitos/cenários da especificação-fonte); ela é um recorte de gestão de risco para esta mudança específica.
- Nenhum item desta matriz deve ser marcado como encerrado sem evidência (ensaio, ADR aprovado ou decisão registrada com responsável), conforme QUA-03/QUA-04.
