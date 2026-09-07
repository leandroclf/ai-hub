# Relatório de revisão documental v4.0

**Data:** 6 de setembro de 2026. **Natureza:** revisão completa de engenharia, sem implementação, provisionamento ou teste de software.

## Resultado da revisão

Foram revistos dez capítulos, 31,095 palavras aproximadamente, 98 requisitos únicos com responsáveis e 98 cenários mínimos de aceite: 78 requisitos herdados/revisados e 20 novos. A matriz mantém os IDs anteriores, títulos atualizados, resultado esperado, evidência e status NÃO EXECUTADO. O catálogo de QUA-06 detalha combinações adicionais de falha e concorrência.

O caminho SYNC agora usa comando/resposta internos diretos, com registros duráveis antes do efeito e antes da resposta pública. Filas deixam de ser espera obrigatória desse modo; ASYNC/AUTO e fatos posteriores mantêm mensageria. Todas as admissões usam UUIDv7 e consulta futura local. A revisão especifica credenciais compartilhadas ou dedicadas, escala de pods/nós/células, Redis dispensável, continuidade seletiva e critérios de criticidade.

## Rastreabilidade do incremento

| Pedido | Requisitos e decisão |
| --- | --- |
| Crescer clientes/provedores sem chamado rotineiro | CAT-11, CFG-06, DAD-11, OPE-12, ADR-23 |
| Credencial do Hub ou dedicada ao cliente | CFG-05, SEG-05, FIN-11, ADR-21 |
| Resposta SYNC na mesma chamada, sem fila obrigatória | COM-01/06, EXE-02/14/15, OPE-09, ADR-19 |
| UUIDv7 em SYNC e ASYNC; consulta futura | EXE-01/16, DAD-04, COM-05, ADR-20 |
| Menor dependência de bancos e Redis | DAD-09/10, ARQ-06, OPE-13, ADR-22/24 |
| Manutenção e continuidade operacional | OPE-14/15, QUA-06, ADR-25 |
| Gaps e compatibilidade com regras anteriores | DEC-04/05, matriz completa, QUA-05/06 |

## Compatibilidade preservada

Permanecem: contratos de entrada/saída por cliente; mesma representação final em GET/webhook; agregação/composição; deadline rígido; TTL de retry em segundos sem renovação; rejeição de retorno tardio; SLA bilateral; pressão adaptativa por domínio externo; isolamento de tenants; administrador individual auditado; custo, receita e contraparte separados; objetos sob custódia do Hub; ambientes local/dev/hom/ppd/prd e requisitos de probes/métricas/logs.

A v4 substitui explicitamente a fila obrigatória do caminho SYNC da v3, o timeout de espera universal de oito segundos e a meta genérica anterior de 99,95%. A proposta de 99,99% e os limites por criticidade continuam sujeitos a qualificação, não são resultados medidos. SYNC não vira 202 quando o prazo termina; retorno final sem custódia não é anunciado como sucesso.

## Gaps e limites que permanecem relevantes

O registro DEC-05 contém 15 gaps, com severidade, mitigação, responsável e condição de encerramento. A definição documental resolve ambiguidades, mas todos os mecanismos ainda precisam de implementação e evidência. Pontos bloqueadores incluem exclusividade DIRECT/QUEUED, isolamento de credenciais, arbitragem de deadline/commit, migração com idempotência, perda regional e recuperação de efeito externo desconhecido.

Redis pode ser removido do caminho obrigatório e sua queda deve ser ensaiada sem perda de SLO. Autoridade durável de escrita não é opcional para confirmar novo protocolo/final. Se ela ficar inacessível, a arquitetura preserva dados/efeitos conhecidos e limita novas operações; isso é uma restrição real de disponibilidade, não uma garantia de funcionamento irrestrito.

Failover automático não prova interrupção zero. A tolerância em segundos dos consumidores com impacto em vida/segurança precisa ser definida em P-10 e comprovada para banco, aplicação, rede, provedor, credencial e domínio regional correspondente. A referência regional não está qualificada para RPO zero regional ou qualquer RTO crítico presumido. Escala automática preserva a arquitetura lógica e aumenta recursos dentro de quotas e orçamento; não significa capacidade infinita nem criação instantânea de bancos.

## Verificação documental

Conferidos IDs únicos, preservação dos 78 requisitos anteriores, cobertura dos 98 requisitos na matriz, status de testes, referências internas, estrutura de tabelas e ausência de arquivos de aplicação/deploy no pacote. Diagrama SVG e HTML consolidado foram regenerados a partir dos capítulos desta revisão. O relatório de validação editorial verifica consistência e integridade dos arquivos; não executa nem aprova o software futuro.

## Estado de entrega

O pacote v3 e seus arquivos foram recuperados e conferidos antes da alteração. Esta revisão contém dez capítulos Markdown, matriz CSV, diagrama SVG, edição HTML, este relatório e um ZIP completo. Fontes primárias e suas limitações constam do capítulo 09. Não houve alteração de código, deploy local/remoto, teste de carga, commit ou publicação de PR nesta entrega. As decisões P-01 a P-11 e gates G0–G4 permanecem explícitos por escopo.
