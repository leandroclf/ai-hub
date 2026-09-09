# Delta for 06-dados-plataforma-e-continuidade

## ADDED Requirements

### Requirement: R3-OPE-01 — Objetos conectados ao fluxo e retenção
O Hub SHALL consumir referências imutáveis por streaming limitado, conservar resultado volumoso antes da publicação e integrar retenção/pins a protocolos, entregas e finanças. Classe/região restringem processamento e acesso.

#### Scenario: R3-OPE-01-S01 — entrada/saída volumosas
- GIVEN entrada/saída volumosas
- WHEN executar/consultar/baixar
- THEN hash/version/tenant preservados e memória limitada

#### Scenario: R3-OPE-01-S02 — falha entre upload e final ou limpeza concorrente
- GIVEN falha entre upload e final ou limpeza concorrente
- WHEN recuperar
- THEN objeto é recuperável ou coletado somente após prova de ausência de obrigação

#### Scenario: R3-OPE-01-S03 — outro tenant usa FileRef ou retenção vence com disputa
- GIVEN outro tenant usa FileRef ou retenção vence com disputa
- WHEN acessar/purgar
- THEN acesso é negado e pin impede remoção necessária

### Requirement: R3-OPE-02 — Kind completo e ambientes reprodutíveis
A entrega SHALL oferecer Compose local completo e kind completo com dependências no cluster, UI/gateway/identidade/dados/mensageria/cofre emulável/observabilidade. Serviços gerenciados em dev/hom/ppd/prd exigem IaC e referências explícitas; imagens e endpoints devem ser reproduzíveis sem IP efêmero.

#### Scenario: R3-OPE-02-S01 — máquina limpa com pré-requisitos
- GIVEN máquina limpa com pré-requisitos
- WHEN subir Compose e kind isolados
- THEN todos os componentes saudáveis e jornada autenticada funciona em ambos

#### Scenario: R3-OPE-02-S02 — recriação de containers/nós
- GIVEN recriação de containers/nós
- WHEN reiniciar
- THEN endpoints não dependem de IP anterior e custódia sobrevive

#### Scenario: R3-OPE-02-S03 — overlays dev/hom/ppd/prd
- GIVEN overlays dev/hom/ppd/prd
- WHEN renderizar e validar
- THEN imagens fixadas e configurações/secrets específicos sem modo local em produção

### Requirement: R3-OPE-03 — Escala e continuidade com envelope
O Hub SHALL automatizar onboarding/placement e escala de pods/nós/dados dentro de envelope aprovado, separar readiness por capacidade e drenar workers/leases. Saturação além do envelope requer admissão controlada; escala infinita e failover instantâneo não são promessas válidas.

#### Scenario: R3-OPE-03-S01 — crescimento dentro do envelope
- GIVEN crescimento dentro do envelope
- WHEN executar carga
- THEN placement e capacidade crescem sem chamado rotineiro

#### Scenario: R3-OPE-03-S02 — rollout ou perda de zona
- GIVEN rollout ou perda de zona
- WHEN medir continuidade
- THEN obrigações sobrevivem e SLO/RTO/RPO medidos são registrados

#### Scenario: R3-OPE-03-S03 — quota esgotada ou autoridade indisponível
- GIVEN quota esgotada ou autoridade indisponível
- WHEN admitir
- THEN recusa seletiva antecede promessa de custódia e leitura segura continua quando possível

### Requirement: R3-OPE-04 — Autoridade durável, isolamento e restore
O Hub SHALL preservar autoridade durável replicada, papéis mínimos e isolamento tenant/aplicação/célula, e qualificar restore reconciliando mensagens/efeitos/objetos/finanças. Redis é dispensável. Falha total da autoridade exige recusa segura; fallback durável alternativo exige consistência/fencing comprovados.

#### Scenario: R3-OPE-04-S01 — papel runtime sem ownership e pool reutilizado
- GIVEN papel runtime sem ownership e pool reutilizado
- WHEN tentar acesso cruzado API/SQL
- THEN fronteiras e trilhas se mantêm

#### Scenario: R3-OPE-04-S02 — Redis fora e manutenção de réplica PostgreSQL
- GIVEN Redis fora e manutenção de réplica PostgreSQL
- WHEN executar failover
- THEN Redis não bloqueia e obrigações confirmadas sobrevivem dentro da garantia qualificada

#### Scenario: R3-OPE-04-S03 — restore anterior a efeitos já realizados
- GIVEN restore anterior a efeitos já realizados
- WHEN reconciliar antes de SUBMIT
- THEN nenhum efeito é repetido cegamente e lacunas têm disposição auditável

### Requirement: R3-OPE-05 — SLA bilateral e telemetria verificável
O Hub SHALL emitir métricas reais de idade/custódia/lag/capacidade e SLA cliente-Hub/Hub-provedor, permitir consulta autorizada de violações e correlacionar logs/traces/protocolos sem segredos. Prometheus/Loki/Grafana devem ser testados com falhas injetadas e cardinalidade limitada.

#### Scenario: R3-OPE-05-S01 — um resultado tempestivo e outro tardio
- GIVEN um resultado tempestivo e outro tardio
- WHEN consultar API/painel
- THEN cada contrato mostra seu relógio e evidência de violação

#### Scenario: R3-OPE-05-S02 — obrigação parada e provedor degradado
- GIVEN obrigação parada e provedor degradado
- WHEN injetar falha
- THEN alerta usa série existente e leva à timeline/runbook

#### Scenario: R3-OPE-05-S03 — payload secreto e muitos protocolos
- GIVEN payload secreto e muitos protocolos
- WHEN observar telemetria
- THEN segredos são removidos e UUIDs não criam cardinalidade ilimitada
