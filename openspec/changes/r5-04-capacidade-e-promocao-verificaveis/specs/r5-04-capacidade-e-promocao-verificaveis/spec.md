# Delta for r5-04-capacidade-e-promocao-verificaveis

## ADDED Requirements

### Requirement: R5-OPE-01 — Capacidade sem bypass e recuperação de permits
O Hub SHALL exigir política de capacidade qualificada para cada rota ativa e recuperar concessões pendentes com evidência durável de transporte/efeito, sem reciclar apenas por timeout. Crescimento de histórico e número de tenants não deve violar orçamento de concessão publicado; isolamento e limite agregado devem valer entre réplicas.

#### Scenario: R5-OPE-01-S01 — rota ativa sem política ou controller
- GIVEN rota ativa sem política ou controller
- WHEN admitir envio
- THEN recusa antes de efeito, exceto fixture local explicitamente segregada

#### Scenario: R5-OPE-01-S02 — queda após efeito ou durante settlement de permit
- GIVEN queda após efeito ou durante settlement de permit
- WHEN reiniciar reconciliador
- THEN obrigação é resolvida por evidência e capacidade não fica perdida nem excedida

#### Scenario: R5-OPE-01-S03 — histórico grande, tenant agressor e provedor que escala
- GIVEN histórico grande, tenant agressor e provedor que escala
- WHEN medir controle sob carga
- THEN custo de Acquire limitado, justiça e adaptação respeitam teto e budgets

### Requirement: R5-OPE-02 — Promoção exige evidência vinculada ao artefato
O processo de promoção SHALL recusar artefato sem evidência íntegra e compatível com SHA/conteúdo/imagens efetivos e sem aprovações verificáveis do ambiente aplicável. Uma string PASS ou nome de aprovação não constitui prova. Mudança de toolchain/base exige requalificação material antes da promoção.

#### Scenario: R5-OPE-02-S01 — prd sem manifesto e variáveis textuais PASS
- GIVEN prd sem manifesto e variáveis textuais PASS
- WHEN executar gate
- THEN BLOCK com evidência ausente

#### Scenario: R5-OPE-02-S02 — manifesto válido pertence a outro SHA ou digest
- GIVEN manifesto válido pertence a outro SHA ou digest
- WHEN avaliar imagem candidata
- THEN BLOCK por incompatibilidade de proveniência

#### Scenario: R5-OPE-02-S03 — artefato qualificado e aprovações autênticas do ambiente
- GIVEN artefato qualificado e aprovações autênticas do ambiente
- WHEN avaliar promoção
- THEN ALLOW auditável sem imprimir credenciais

### Requirement: R5-OPE-03 — Ambientes elásticos com dados duráveis e isolamento completo
O Hub SHALL disponibilizar perfis local/dev/hom/ppd/prd reproduzíveis com identidade, dados, segredos, rede, observabilidade e capacidade isolados. Escala automática deve abranger pods/nós/placement e budgets das dependências dentro de quotas explícitas; continuidade é aferida por requisições/obrigações reconciliadas, não somente readiness.

#### Scenario: R5-OPE-03-S01 — novo cliente e carga dentro do envelope aprovado
- GIVEN novo cliente e carga dentro do envelope aprovado
- WHEN provisionar e escalar
- THEN onboarding/placement automatizado sem editar infraestrutura manual por tenant

#### Scenario: R5-OPE-03-S02 — perda de nó/zona com tráfego e obrigações em trânsito
- GIVEN perda de nó/zona com tráfego e obrigações em trânsito
- WHEN recuperar ambiente qualificado
- THEN SLO/RPO/RTO medidos e aceites/resultados reconciliados

#### Scenario: R5-OPE-03-S03 — credencial/fixture local em configuração prd
- GIVEN credencial/fixture local em configuração prd
- WHEN validar release
- THEN recusa por fronteira de ambiente e segredo
