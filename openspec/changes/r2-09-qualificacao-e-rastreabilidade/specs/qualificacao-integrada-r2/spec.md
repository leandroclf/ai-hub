# Delta for qualificacao-integrada-r2

## ADDED Requirements

### Requirement: R2-QUA-01 — Rastreabilidade do alvo até evidência

Toda conclusão de implementação SHALL ligar requisito baseline/R2, cenário, tarefa, arquivos, comando/ambiente/SHA e resultado verificável. Documentação, compilação, mock, screenshot e validação client-side SHALL ser distinguidos de teste de integração/HA. Requisito parcialmente demonstrado NÃO SHALL ser marcado concluído integralmente.

Baseline relacionada: QUA-03, QUA-04, DEC-02, DEC-05.

#### Scenario: R2-QUA-01-S01 — Novo requisito implementado

- GIVEN cenários normativos têm provas de execução
- WHEN revisor avalia conclusão
- THEN status registra o escopo comprovado, links e limites sem usar quantidade de arquivos como evidência

#### Scenario: R2-QUA-01-S02 — Evidência incompatível

- GIVEN teste verifica apenas status e requisito exige corpo/identidade
- WHEN relatório declara atendimento integral
- THEN gate reprova conclusão e mantém gap aberto

#### Scenario: R2-QUA-01-S03 — Decisão ainda pendente

- GIVEN documentos evoluíram mas responsável não aprovou P-01 a P-11
- WHEN revisão atualiza status
- THEN pendência permanece aberta; não presume avanço por commit ou lembrete

### Requirement: R2-QUA-02 — Fixtures e oráculos independentes reproduzíveis

Suite SHALL criar fixtures sintéticas isoladas por execução e validar identidade externa, payload real, saldo/ledger, bytes completos, efeitos e custódia com oráculos independentes. SHALL cobrir falha nas fronteiras duráveis, duplicatas/concorrência, deadlines e dependências. Collection privada ou banco de sessão anterior NÃO SHALL ser pré-requisito implícito.

Baseline relacionada: QUA-01, QUA-02, QUA-05, QUA-06.

#### Scenario: R2-QUA-02-S01 — Ambiente limpo

- GIVEN clone não contém collection privada nem seed histórico
- WHEN pipeline sobe o ambiente de teste
- THEN fixtures locais habilitam todos os cenários obrigatórios sem segredos reais

#### Scenario: R2-QUA-02-S02 — Regressão do eco

- GIVEN conector devolve input em vez de resposta
- WHEN suite executa sucesso SYNC e ASYNC
- THEN falha apesar de status SUCCEEDED

#### Scenario: R2-QUA-02-S03 — Regressão financeira

- GIVEN saldo capturado deixa de contar no limite
- WHEN suite executa duas capturas e terceira tentativa acima do saldo
- THEN falha se terceira chamada externa ocorrer

### Requirement: R2-QUA-03 — Gates de contrato, experiência e operação

Promoção SHALL exigir gates proporcionais ao perfil: schema/contratos, testes de domínio, integração concorrente/falhas, jornadas UI reais e acessibilidade, laboratório completo, carga/isolamento, recuperação e segurança. Exceção SHALL ter responsável, risco e validade explícitos e NÃO SHALL autorizar mentira sobre atendimento. CI SHALL bloquear regressões em cenários P0.

Baseline relacionada: QUA-04, QUA-06, COM-04, OPE-15.

#### Scenario: R2-QUA-03-S01 — PR de backend

- GIVEN mudança altera corpo público ou envelope
- WHEN pipeline compara contratos e roda consumidores
- THEN incompatibilidade não negociada bloqueia merge/promoção

#### Scenario: R2-QUA-03-S02 — UI sem backend

- GIVEN jornada funciona só com mocks
- WHEN equipe tenta marcar pronta
- THEN gate de integração real não passa, embora build de UI possa passar

#### Scenario: R2-QUA-03-S03 — P0 reaberto

- GIVEN ensaio de callback perde recibo confirmado
- WHEN há release candidato
- THEN ativação do fluxo afetado é bloqueada até corrigir ou manter explicitamente indisponível

### Requirement: R2-QUA-04 — Evolução OpenSpec e baseline sem falso arquivamento

Revisão SHALL preservar os 98 IDs v4 e distinguir baseline normativo da implementação. Deltas R2 SHALL usar IDs próprios, cenários, dependências e tarefas abertas. Reconciliação da árvore canônica SHALL ter plano explícito sem arquivar a change v4 inteira como implementada. Fechamento de tarefa SHALL depender de evidência e atualização de rastreabilidade, mantendo histórico. Fronteiras e escolhas de linguagem da baseline SHALL permanecer rastreadas; troca de stack SHALL exigir decisão e evidência de benefício.

Baseline relacionada: DEC-01, DEC-03, DEC-04, DEC-05, QUA-03, ARQ-04.

#### Scenario: R2-QUA-04-S01 — Importação deste pacote

- GIVEN repositório contém change v4 aberta e não contém openspec/specs
- WHEN arquivos R2 são adicionados
- THEN não substituem specs históricas nem marcam tarefas concluídas; validação trata changes R2 separadamente

#### Scenario: R2-QUA-04-S02 — Mudança R2 concluída

- GIVEN todos os cenários da change têm prova e gates pertinentes passaram
- WHEN equipe consolida a especificação
- THEN preserva IDs e histórico, resolve baseline/deltas sem esconder requisitos v4 ainda abertos

#### Scenario: R2-QUA-04-S03 — Regressão posterior

- GIVEN nova versão invalida garantia antes demonstrada
- WHEN auditoria recebe evidência
- THEN status reabre com referência ao novo SHA, sem apagar prova histórica nem reduzir requisito
