# Design: Evidência consolidada da implementação

## Context
Brownfield a540b40007fe6b8ed523e17afe00e96ff8f8ad50; responsabilidade Engenharia e Qualidade. O delta adiciona critérios de fechamento, não substitui regras anteriores.

## R6-QUA-01 — decisão e justificativa
Preservar v4/R2/R3/R4/R5 e incorporar deltas R6, com semântica distinta para achado, requisito, cenário e teste. Registrar SHA/diff, toolchain, imagem, fixture, comando, esperado/observado, logs saneados e digests. Revalidar só o que mudança material afeta, mas completar cenários obrigatórios sem skip. CLI OpenSpec fixada, sem @latest em evidência reproduzível. Decisões externas não aprovadas permanecem pendentes sem impedir implementação de fixture técnica.

### Contrato obrigatório
A engenharia SHALL manter inventário integral extraído das specs e resultado individual com evidência por cenário. Contagens SHALL ser calculadas dos logs, com subtestes/pacotes diferenciados. Um resultado histórico sem vínculo verificável com artefato SHALL permanecer histórico e não virar PASS do candidato.

### Superfícies afetadas
[docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md:29](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/EVIDENCE_INDEX.md#L29), [docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv:11](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/SCENARIO_RESULTS.csv#L11), [docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md:48](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/docs/reviews/2026-09-12-r5/implementation/FINAL_REPORT.md#L48)

### Provas de fechamento
- R6-QUA-01-S01: specs vigentes e resultados parciais antigos → gerar inventário e matriz → cada cenário consta exatamente uma vez e lacunas ficam NOT_RUN.
- R6-QUA-01-S02: teste altera subcasos e SHA muda → regenerar evidências e relatório → contagens derivadas e aprovação ligada ao candidato correto.
- R6-QUA-01-S03: dependência obrigatória indisponível → executar gate integrado → FAIL/BLOCK com causa, nunca skip contado como aprovação.

## Persistência e atomicidade
A autoridade do domínio conserva transição, inbox/outbox e recibos em transação local. Cache é derivado; queda de Redis não altera a decisão durável. Para efeitos externos, lease sozinho não é prova de ausência de efeito: usar fencing reconhecido, idempotência do provedor ou reconciliação por chave. Estado UNKNOWN conserva obrigação.
## Comunicação e paralelismo
HTTP/JSON nos contratos existentes; SNS/SQS para fatos/obrigações; HTTP direto preservado para SYNC. Workers paralelos limitados por tenant/rota/produto, sem goroutine ilimitada. Timeout de transporte não encerra automaticamente o estado econômico.
## Segurança e privacidade
Identidade nominal/MFA para administração, autorização por recurso, logs sem payload sensível ou segredo. Testes com identidades A/B e controles positivos. Não registrar credenciais em evidências.
## Observabilidade
Métricas de aceites, finais, idade de obrigação, retries, violações de prazo, drift, quarentenas e ações administrativas. Labels limitadas por agregação; UUID/protocolo em logs/traces, não em séries ilimitadas. Cada erro de custódia tem alerta e disposição.
## Estratégia de migração e rollback
Aplicar migrações novas, manter checksums antigos e ensaiar upgrade populado. Rollback binário só se schema/dados forem compatíveis; caso contrário bloquear downgrade e aplicar forward fix. Não apagar recibos, históricos, saldos ou volumes para passar teste.
## Alternativas
Reescrita da stack acrescentaria risco sem resolver os invariantes. Mocks isolados são úteis para contrato puro, insuficientes para concorrência/DB/restore. Adoção de componente adicional requer ADR com benefício medido.
## Estratégia de validação
Reproduzir caso negativo antes da correção; confirmar chamada real, sucesso, falha, concorrência, crash e idempotência conforme cenário. DB/broker/objetos/OIDC reais em laboratório quando necessários. Browser para jornadas humanas. Medir carga/HA apenas com perfil de capacidade identificado.
