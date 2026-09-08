# Delta for objetos-dados-e-retencao

## ADDED Requirements

### Requirement: R2-DAD-01 — Upload direto e referência imutável por tenant

Arquivos SHALL ingressar por sessão autorizada de upload direto/multipart com tamanho/tipo/checksum e prazo definidos. FileRef SHALL apontar objeto/versão imutável cujo proprietário e integridade foram validados. Protocolo NÃO SHALL ser aceito com referência não elegível; URL pré-assinada SHALL ser obtida separadamente da representação final e ter validade curta do perfil.

Baseline relacionada: DAD-05, COM-02, SEG-02.

#### Scenario: R2-DAD-01-S01 — Arquivo elegível

- GIVEN cliente A concluiu upload e validação
- WHEN submete pedido com file_ref
- THEN Hub verifica vínculo/versão/hash sem carregar arquivo inteiro na API

#### Scenario: R2-DAD-01-S02 — Objeto de outro cliente

- GIVEN A indica file_ref de B
- WHEN admissão ou download é solicitado
- THEN acesso é negado indistinguivelmente de referência inexistente para A

#### Scenario: R2-DAD-01-S03 — Multipart incompleto

- GIVEN faltam partes ou checksum diverge
- WHEN cliente pede confirmação do objeto
- THEN referência não fica READY; sessão pode ser retomada/expirada conforme política

### Requirement: R2-DAD-02 — Custódia de resultado volumoso

Sucesso com resultado volumoso SHALL exigir objeto durável, versão/checksum e direito de custódia confirmados antes de consolidar referência final. Streaming SHALL limitar buffers e broker SHALL transportar apenas referências/metadados permitidos. Falha entre storage e banco SHALL ser reconciliável; objeto órfão NÃO SHALL ser confundido com protocolo concluído.

Baseline relacionada: DAD-05, DAD-04, DAD-09, OPE-11.

#### Scenario: R2-DAD-02-S01 — Resultado grande

- GIVEN provedor entrega arquivo superior ao limite inline
- WHEN conector conserva a resposta
- THEN streaming produz objeto validado e só então final aponta file_ref

#### Scenario: R2-DAD-02-S02 — S3 falha

- GIVEN resultado externo exige custódia em objeto
- WHEN put ou verificação falha
- THEN nenhum sucesso com link quebrado é emitido; obrigação de captura/reconciliação permanece

#### Scenario: R2-DAD-02-S03 — Objeto sem commit do protocolo

- GIVEN upload final completou e commit posterior falhou
- WHEN reconciliador executa
- THEN liga objeto somente à obrigação correta ou marca órfão para política segura, sem reexecutar provedor

### Requirement: R2-DAD-03 — Retenção por classe e obrigações abertas

Retenção SHALL ser definida por classe/contrato/região e aplicada a entradas, resultados, recibos, chaves de idempotência, eventos, auditoria e financeiro. Expurgo SHALL respeitar pins de entrega/reconciliação/contestação/legal hold, registrar tombstone e não permitir ressurreição de efeitos por replay antigo. Ausência causada por expurgo SHALL seguir contrato público sem revelar outro tenant.

Baseline relacionada: DAD-06, DAD-08, OPE-11.

#### Scenario: R2-DAD-03-S01 — Expurgo elegível

- GIVEN classe aprovada atingiu retenção e não há obrigação aberta
- WHEN job executa após dry-run validado
- THEN expurgo é auditado e deixa metadado mínimo/tombstone permitido

#### Scenario: R2-DAD-03-S02 — Entrega pendente

- GIVEN resultado excedeu retenção padrão mas delivery/contestação precisa dele
- WHEN job avalia elegibilidade
- THEN objeto/final permanece protegido ou migra segundo política contratada sem perder obrigação

#### Scenario: R2-DAD-03-S03 — Mensagem antiga

- GIVEN evento fora da janela tenta reaplicar efeito já encerrado
- WHEN consumer recebe replay
- THEN tombstone/dedup impede novo efeito e produz diagnóstico

### Requirement: R2-DAD-04 — Propriedade, RLS e continuidade de leitura

Cada domínio SHALL ter autoridade de escrita e papéis próprios. Consultas de final imutável MAY usar réplica/cache autorizado com prova de versão e integridade; ausência em cópia atrasada NÃO SHALL virar 404 conclusivo. Sem cópia confiável ou autorização válida SHALL haver indisponibilidade explícita. Redis NÃO SHALL ser fonte única de dado/lock financeiro.

Baseline relacionada: DAD-01, DAD-07, DAD-10, SEG-03.

#### Scenario: R2-DAD-04-S01 — Writer indisponível com final válido

- GIVEN existe réplica/cache confiável do final e autorização vigente
- WHEN cliente consulta
- THEN recebe mesmo corpo/hash sem acessar provedor; novas admissões continuam exigindo writer

#### Scenario: R2-DAD-04-S02 — Réplica atrasada

- GIVEN réplica ainda não conhece protocolo recentemente aceito
- WHEN GET tenta recuperar resultado
- THEN consulta autoridade quando disponível ou informa indisponibilidade/estado não confirmado, sem falso inexistente

#### Scenario: R2-DAD-04-S03 — Isolamento de domínio

- GIVEN usuário de Pulsar tenta escrever em saldo ou protocolo
- WHEN banco avalia permissão
- THEN nega escrita fora da autoridade mesmo que aplicação tenha bug

### Requirement: R2-DAD-05 — Restauração reconciliada sem duplicar efeito

Restauração SHALL recuperar conjunto reconciliável de core, controle, financeiro, objetos, idempotência e obrigações, com egress produtivo cercado até validação. Backup assíncrono NÃO SHALL ser apresentado como prova de RPO zero regional. Reativação SHALL exigir autoridade única e tratamento de operações enviadas depois do ponto restaurado.

Baseline relacionada: DAD-07, OPE-11, OPE-15.

#### Scenario: R2-DAD-05-S01 — Restore controlado

- GIVEN backup/PITR e versões de objetos estão disponíveis
- WHEN ambiente isolado é restaurado
- THEN egress permanece bloqueado até confrontar comandos, fatos, saldos e referências

#### Scenario: R2-DAD-05-S02 — Operação após backup

- GIVEN provedor executou depois do ponto restaurado
- WHEN sistema encontra lacuna histórica
- THEN reconcilia com parceiro/idempotência; não reenfileira todos os pedidos do período

#### Scenario: R2-DAD-05-S03 — Falha regional

- GIVEN perfil contratado exige perda regional sem perda de fatos
- WHEN topologia possui somente confirmação regional
- THEN perfil não é qualificado nem ativado até prova de custódia e fencing compatíveis
