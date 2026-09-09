# Delta for 03-capacidade-credenciais-e-latencia

## ADDED Requirements

### Requirement: R3-INT-01 — Controle adaptativo conectado a todo I/O
O Hub SHALL adquirir concessão global por domínio antes de SUBMIT/STATUS/FETCH, devolver métricas de resultado e adaptar concorrência com redução por erros, recuperação amortecida e Retry-After. Limites contratuais duros coexistem com adaptação e justiça entre tenants.

#### Scenario: R3-INT-01-S01 — provedor aumenta capacidade sob demanda
- GIVEN provedor aumenta capacidade sob demanda
- WHEN executar carga
- THEN janela cresce dentro do envelope sem intervenção

#### Scenario: R3-INT-01-S02 — provedor retorna 429/timeouts
- GIVEN provedor retorna 429/timeouts
- WHEN medir janela seguinte
- THEN concorrência cai sem tempestade de retries

#### Scenario: R3-INT-01-S03 — vários pods e tenants na conta compartilhada
- GIVEN vários pods e tenants na conta compartilhada
- WHEN saturar A e reiniciar owner
- THEN soma real respeita domínio global e B preserva parcela; UNKNOWN não libera concessão insegura

### Requirement: R3-INT-02 — Cache de autenticação isolado e sem tokens no Redis
O Hub SHALL manter segredos em cofre/cache de memória limitado com expiração/revogação, eliminar tokens do Redis e coordenar renovação por binding sem bloquear contas independentes. L1 válido deve dispensar chamada remota; falha não permite token expirado ou fallback de outro cliente.

#### Scenario: R3-INT-02-S01 — L1 válido e Redis/cofre temporariamente fora
- GIVEN L1 válido e Redis/cofre temporariamente fora
- WHEN autenticar
- THEN token ainda autorizado permite chamada sem dependência remota obrigatória

#### Scenario: R3-INT-02-S02 — token expirado/revogado e cofre indisponível
- GIVEN token expirado/revogado e cofre indisponível
- WHEN renovar
- THEN somente binding afetado falha com segurança

#### Scenario: R3-INT-02-S03 — OAuth da conta A lento e B saudável
- GIVEN OAuth da conta A lento e B saudável
- WHEN renovar ambas
- THEN B não espera mutex de A e Redis/logs não contêm segredos

### Requirement: R3-INT-03 — Pools HTTP e budgets de concorrência
O Hub SHALL reutilizar pools limitados por origem/identidade TLS, manter timeout por operação e drenar pools obsoletos. Controlador global limita concorrência; certificados/bindings incompatíveis não compartilham estado de autenticação.

#### Scenario: R3-INT-03-S01 — milhares de chamadas à origem
- GIVEN milhares de chamadas à origem
- WHEN medir sockets/handshakes
- THEN reuso amortiza handshakes e FDs ficam no budget

#### Scenario: R3-INT-03-S02 — certificado mTLS muda de versão
- GIVEN certificado mTLS muda de versão
- WHEN rotacionar
- THEN novo pool usa certificado novo e antigo é drenado

#### Scenario: R3-INT-03-S03 — provedor A lento e B saudável
- GIVEN provedor A lento e B saudável
- WHEN saturar A
- THEN B preserva budget e goroutines/filas permanecem limitadas
