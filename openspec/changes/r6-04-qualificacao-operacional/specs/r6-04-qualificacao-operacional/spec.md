# Delta for r6-04-qualificacao-operacional

## ADDED Requirements

### Requirement: R6-OPE-01 — Gate de promoção vinculado ao artefato e à cobertura
O Hub SHALL bloquear promoção quando qualquer evidência obrigatória não corresponder ao artefato candidato, ao cenário vigente e à execução autorizada. Status SHALL usar enum estrito; cobertura parcial e declarações autoatribuídas não qualificam promoção.

#### Scenario: R6-OPE-01-S01 — manifesto de um cenário, SHA falso e status prefixado PASS
- GIVEN manifesto de um cenário, SHA falso e status prefixado PASS
- WHEN avaliar promoção prd
- THEN BLOCK por identidade, enum e cobertura insuficientes

#### Scenario: R6-OPE-01-S02 — manifesto completo de execução autorizada e artefato idêntico
- GIVEN manifesto completo de execução autorizada e artefato idêntico
- WHEN avaliar gate e depois alterar log/imagem/spec
- THEN ALLOW original e BLOCK de cada adulteração

#### Scenario: R6-OPE-01-S03 — manifesto contém skip obrigatório ou aprovação não autenticada
- GIVEN manifesto contém skip obrigatório ou aprovação não autenticada
- WHEN avaliar promoção
- THEN BLOCK com motivo específico sem executar deploy

### Requirement: R6-OPE-02 — Restore populado com versões e retomada reconciliada
O Hub SHALL demonstrar restore não vazio de todas as autoridades e obrigações, incluindo versões de objetos referenciadas, seguido de retomada cercada e reconciliação de efeitos e valores. A ausência de dados ou a mera igualdade de contagens SHALL impedir aprovação de recuperação integral.

#### Scenario: R6-OPE-02-S01 — backup com etapas pendentes, UNKNOWN e duas versões referenciadas
- GIVEN backup com etapas pendentes, UNKNOWN e duas versões referenciadas
- WHEN restaurar em alvo isolado
- THEN todas as referências resolvem para bytes corretos e obrigações permanecem

#### Scenario: R6-OPE-02-S02 — restore completo e fronteira de I/O cercada
- GIVEN restore completo e fronteira de I/O cercada
- WHEN retomar workers e comparar oráculo antes/depois
- THEN efeitos e saldos exatos sem duplicação e protocolos recuperados

#### Scenario: R6-OPE-02-S03 — backup vazio ou versão de objeto ausente
- GIVEN backup vazio ou versão de objeto ausente
- WHEN executar qualificação
- THEN BLOCK, sem anunciar restore reconciliado

### Requirement: R6-OPE-03 — Ambientes e elasticidade com orçamento de dependências
O Hub SHALL fornecer perfis reproduzíveis local/dev/hom/ppd/prd com dependências e limites explícitos, escalabilidade automática dentro do envelope qualificado e continuidade mensurável. Readiness e topologia de obrigações SHALL refletir capacidade real de admitir/processar com custódia, sem confundir laboratório efêmero com HA produtiva.

#### Scenario: R6-OPE-03-S01 — aumento de tenants dentro do envelope de capacidade
- GIVEN aumento de tenants dentro do envelope de capacidade
- WHEN executar onboarding e carga
- THEN escala automática sem ticket por cliente, com budgets e isolamento mensurados

#### Scenario: R6-OPE-03-S02 — rolling update e falha de pod/nó com obrigações aceitas
- GIVEN rolling update e falha de pod/nó com obrigações aceitas
- WHEN substituir componentes
- THEN custódia preservada, readiness correta e recuperação dentro do SLO do perfil

#### Scenario: R6-OPE-03-S03 — assinatura obrigatória removida após bootstrap
- GIVEN assinatura obrigatória removida após bootstrap
- WHEN publicar e observar reconciliação
- THEN obrigação não é silenciosamente perdida; alerta, contenção e recuperação verificáveis

### Requirement: R6-OPE-04 — Capacidade obrigatória e feedback recuperável
O Hub SHALL exigir política efetiva de capacidade para rotas ativas e conservar obrigações de feedback/settlement até resolução. Controle adaptativo SHALL respeitar teto seguro contratado e compartilhar capacidade de forma justa entre tenants, com custo estável em função do estado ativo.

#### Scenario: R6-OPE-04-S01 — rota sem domínio de capacidade ou política vencida
- GIVEN rota sem domínio de capacidade ou política vencida
- WHEN publicar e executar
- THEN negação seletiva antes de I/O sem bypass

#### Scenario: R6-OPE-04-S02 — provedor reduz capacidade e depois escala
- GIVEN provedor reduz capacidade e depois escala
- WHEN medir feedback com dois tenants
- THEN controlador reduz pressão, recupera de forma estável e respeita teto e justiça

#### Scenario: R6-OPE-04-S03 — settlement falha e lease vence
- GIVEN settlement falha e lease vence
- WHEN reiniciar e fornecer evidência terminal
- THEN permit é resolvido sem perda de obrigação nem reciclagem cega por TTL
