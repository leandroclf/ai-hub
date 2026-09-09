# Decisões, ordem de implementação e migração

## Decisões não presumidas
As sete decisões iniciais D-01…D-07 e o registro P-01…P-11 da v4 precisam ser reconciliados nominalmente.
Código existente não prova aprovação comercial/operacional. Ler registro original; não renumerar pendência para apagá-la.

| Decisão | Confirmação necessária | Responsável funcional | Trabalho independente permitido |
|---|---|---|---|
| D-01 | Demanda real, picos, arquivos, duração e envelope | Negócio/Engenharia | Harness e perfis sintéticos explícitos. |
| D-02 | SLA, parcialidade, timeout e tardio | Produto/Operações | Relógios distintos, provas de corrida e políticas versionadas. |
| D-03 | Compra/venda, incidência, franquia, estorno e saldo estrito | Comercial/Financeiro | Ledger, reservas, contratos de fixture e conciliação. |
| D-04 | Retenção, classe, dados permitidos e região | Donos dos dados/Segurança | Controle por classe, pins e testes com dados sintéticos. |
| D-05 | AWS/EKS, orçamento, gateway/edição e continuidade regional | Arquitetura/Plataforma | Compose/kind e IaC validável sem provisionar custo não autorizado. |
| D-06 | Equivalência, idempotência externa, polling/callback | Integrações/Produto | Matriz de capacidade e simuladores adversariais. |
| D-07 | Catálogo inicial e formatos financeiros | Produto/Financeiro | Contratos de import/export e jornada homologável por fixture. |
| T-R2-01 | Semântica de prazo com confirmação durável | Core/Dados/SRE | Continuar investigação técnica e contraexemplos; não mudar SLA para tempo de decisão. |

O ADR existente T-R2-01 é evidência de um limite do algoritmo atual, não autorização para enfraquecer EXE-11.
Se uma solução exigir mudança normativa, documentar opções/consequências e manter o gate aberto; concluir trabalho independente.
Não prometer que qualquer agente poderá resolver matematicamente um requisito incompatível com o modelo de falhas.

## Baseline adotada para execução — 2026-09-09

As decisões abaixo foram adotadas pelo responsável da execução para viabilizar o planejamento no laboratório local e no perfil de homologação. Elas são defaults versionados, não contratos comerciais nem autorização de produção. Uma mudança exige nova versão de contrato, migração compatível e requalificação dos cenários afetados.

| ID | Decisão adotada | Motivo e trade-off | Critério de revisão |
|---|---|---|---|
| D-01 | Perfil sintético inicial: 100 tenants, 10 aplicações/tenant, 20 provedores, 200 RPS de admissão, 50 RPS por provedor, payload JSON até 1 MiB, arquivo até 1 GiB, operação até 60 s, concorrência de 500 operações por célula | Envelope pequeno e reproduzível para Compose/Kind; não representa demanda comercial real | Substituir pelos percentis observados de 30 dias e pelo pico contratado antes da ativação comercial |
| D-02 | SLA de cliente padrão de 30 s; reserva de finalização de 5 s; timeout por tentativa de 10 s; `Retry-After` respeitado; primeira falha transitória fixa a janela de retry; `UNKNOWN` não é sucesso; resultado tardio vira evidência sem reabrir protocolo | Mantém prazos separados e evita retry inseguro; pode reduzir disponibilidade percebida em provedor ambíguo | Aprovação por Produto/Operações após ensaio T-R2-01 e definição por perfil de criticidade |
| D-03 | Cobrança por unidade econômica efetiva, deduplicada por chave semântica; reserva antes do efeito; captura pelo valor efetivo; `UNKNOWN` mantém retenção; ledger decimal em BRL com 4 casas e estorno por lançamento compensatório | Conserva dinheiro e evita cobrança por duplicata técnica; não define impostos ou preço comercial final | Aprovação Comercial/Financeiro e contrato de tabela de preços, franquias, impostos e moedas |
| D-04 | Dados sintéticos no laboratório; classificação `SYNTHETIC`/`CONFIDENTIAL`; PostgreSQL e S3 na mesma região lógica; retenção online de 90 dias, uploads órfãos por 24 h, logs 30 dias, traces 7 dias, métricas 90 dias; obrigações abertas sem expurgo | Reproduzível e conservador para teste; prazos não são decisão jurídica | Aprovação dos donos dos dados, Segurança e Jurídico; substituir por política regulatória/contratual |
| D-05 | Kubernetes/EKS como alvo de produção, Compose para local, Kind para qualificação; PostgreSQL gerenciado por célula; S3 versionado; Kong na borda; HPA/KEDA para escala; OpenTelemetry + Prometheus/Loki/Grafana/Tempo; sem provisionamento pago nesta execução | Padrão operacional comum e compatível com os artefatos existentes; custo e RPO/RTO de produção continuam condicionados ao orçamento | Aprovação de Arquitetura/Plataforma/FinOps com região, orçamento, RPO/RTO, edição e topologia |
| D-06 | Cada operação recebe chave estável; reenviar só com idempotência homologada ou prova de ausência de efeito; callback e polling podem coexistir; callback é recebido em inbox durável; conflito gera reconciliação; conta/binding permanecem congelados por operação | Maximiza interoperabilidade e segurança; exige provider-sim/adapters que exponham status e idempotência | Homologação individual de cada provedor e aprovação da matriz de capacidades |
| D-07 | Catálogo inicial versionado com JSON Schema, perfis de entrada/saída, produtos e ofertas imutáveis após publicação; import/export em JSON/CSV versionado; planos suportam unidade, franquia, volume total e faixas marginais | Facilita evolução compatível e jornada administrativa; requer governança de versão e validação financeira | Aprovação de Produto/Financeiro do catálogo, layouts, preços e contratos reais |
| T-R2-01 | Adotar operacionalmente a regra conservadora: confirmação durável antes de anunciar sucesso; em dúvida, `UNKNOWN`/reconciliação; nunca usar comparação de relógio anterior ao commit como prova | É segura sob falha entre efeito externo e commit, mas pode converter casos ambíguos em indisponibilidade | Core/Dados/SRE devem concluir contraexemplo com relógios separados, fencing e commit; a regra normativa não é considerada matematicamente encerrada sem esse ensaio |

### Aprovação necessária antes de produção

Os defaults acima autorizam o avanço técnico local e de homologação. Antes de qualquer ativação comercial, devem ser substituídos ou ratificados por responsáveis nominais: demanda real (D-01), SLA (D-02), contratos financeiros (D-03), retenção e região (D-04), plataforma/orçamento/RPO-RTO (D-05), matriz de capacidades de cada provedor (D-06), catálogo e formatos oficiais (D-07) e resultado técnico do contraexemplo T-R2-01.

## Sequência e dependências
1. G0/G1, fixture OIDC e harness G2/G3: tornam as correções verificáveis desde o começo.
2. EXE-05 e ADM-01: topologia/custódia e fronteira de acesso antes de expor jornadas.
3. CAT-01/CAT-02 e EXE-01/EXE-02/EXE-03: contratos exatos, adapter, callback e recuperação.
4. EXE-04 e INT-01/02/03: prazos, pressão e autenticação/pools; não fechar SYNC antes da integração.
5. CAT-03/CAT-04 e FIN-01/02/03: DAG, projeções, orçamento/incidência e entregas.
6. ADM-02/03/04: jornadas ligadas aos contratos já efetivos; começar UI/contrato antes, fechar só depois do runtime.
7. OPE-01…05: objetos/ambientes/escala/restore/telemetria, com ensaios desde os primeiros fluxos.
8. QUA-01: gate integral dos 189 requisitos (164 herdados + 25 R3), não apenas dos 75 cenários novos.
Esta ordem é um grafo de dependência de entregas; não exige terminar uma pasta inteira antes de testar outra.

## Migração
Criar branch de trabalho. Registrar HEAD inicial e mudanças existentes. Migrações aditivas, idempotentes quando pertinente,
com índices/backfill em lotes e contagens/hashes antes/depois. Constraints novas devem considerar dados históricos.
Versionar contratos, destinos e snapshots; não reescrever decisão histórica sob catálogo atual.
Habilitar canário com admissão controlada e métricas; manter leitura/recuperação para protocolos antigos.

## Rollback
Reverter ativação/imagem compatível com schema expandido, parar novas admissões da capacidade e drenar.
Não apagar protocolo/outbox/receipt/quarentena/ledger/reserva para obter estado limpo.
Dados com versão nova precisam leitor compatível ou plano de forward-fix; downgrade de schema destrutivo não é rollback seguro.
Registrar runbook e ensaio de restauração reconciliada antes de promoção.

## Critério de conclusão
Uma tarefa está concluída com código conectado, migração, testes e evidência. Um requisito só fecha com todos os cenários/invariantes.
Bloqueio deve nomear recurso/decisão ausente, tentativa feita, impacto, itens independentes concluídos e próximo passo.
Nenhum pedido de “implementar tudo” autoriza inventar evidência, apagar pendência ou usar credencial comercial inexistente.
