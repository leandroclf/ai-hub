# Gates de execução, migração e evidência

## Estado desta auditoria
- Go 1.24.13: race suite exit 0; 37 pass e 18 skip.
- TypeScript/Vite build: exit 0, dependências locais de lockfile idêntico; instalação limpa não qualificada.
- OpenSpec 1.12.0: 17 changes existentes passaram strict, sem config.yaml.
- Quatro violações JSON e uma falha de L1/cofre reproduzidas por overlays separados.
- Precisão grande/enum anteriores passaram. Root null também foi recusado corretamente.
- Banco/Compose/kind/browser/carga/HA não executados. Não inferir sua aprovação de logs históricos.

## Gates da implementação
| Gate | Critério | Evidência |
|---|---|---|
| G0 Proveniência | HEAD/diff/AGENTS e ferramentas fixados | Manifesto de conteúdo/versões e status Git. |
| G1 Migração | Banco R2 antigo, banco limpo e variante R3 conhecida | Checksums, schema, contagem/hash de dados e plano de convergência. |
| G2 Contratos | Tipos/EOF/dialeto/precisão e correlação externos | Contraexemplos agora verdes e negativos preservados. |
| G3 Custódia | Callback ingress real, inbox, worker, UNKNOWN e handoff | Queda antes/depois de commit/ACK/efeito; oráculo de contagem. |
| G4 Jornada | SYNC/ASYNC/poll+callback/DAG/GET/webhook/financeiro | Mesmo resultado, efeito único conforme garantia e conservação monetária. |
| G5 Console | Todas as jornadas B-R4 ligadas a APIs reais | Browser, reload, erros e autorização negativa. |
| G6 Operação | Um Compose, kind completo, dados, sinais, escala e restore | Runtime e ensaio; render/build não bastam. |
| G7 União | 201 requisitos e 732 cenários inventariados | Resultados por cenário, zero skip obrigatório, sem fechamento falso. |
| G8 Promoção | Contratos/SLO/quotas/DR e ambiente externo aprovados | Aprovações reais e qualificação específica; não implica deploy autorizado. |

A validação OpenSpec é estrutural; pode passar com toda implementação pendente.
Resultado de teste precisa ser atribuído ao requisito/cenário que demonstra, não a todos do pacote.

## Ensaios novos obrigatórios
1. Callback na URL real retornada no SUBMIT sem token de workload Hub.
2. Órfão com origem inválida seguido de legítima para mesmo ID/body: legítimo não é envenenado.
3. Órfão retomado sem novo callback, duas réplicas, pool SQL pequeno e item defeituoso.
4. Correlação/schema divergente no callback não substitui operação.
5. null/string, string/integer, EOF, minimum, número exato e dialeto.
6. L1 válido durante falha do cofre; expirado/revogado nunca é aceito.
7. Crescimento de bindings/versões e cancelamento de refresh sem memória/espera ilimitada.
8. Upgrade a partir de checksum antigo sem apagar volume ou adulterar ledger.
9. Catálogo grande com maioria de ofertas não relacionadas: medir plano/IO/round-trips/memória.
10. Matriz deriva das specs e detecta omissão de cenários v4 ou evidência de conteúdo anterior.

## Regras do ambiente
Antes de subir ou trocar: docker compose ls e docker ps -a.
Identificar projeto oficial local; manter nome estável e um ecossistema Hub ativo.
Parar somente Hub identificado antes da substituição; verificar encerramento.
Preservar volumes e dados. Não usar docker system prune ou curingas amplos.
Kind deve usar dependências próprias no cluster no perfil completo, sem Compose auxiliar paralelo.
Se runtime for restrito, registrar erro concreto e concluir todo trabalho independente.

## Evidência mínima por execução
Data, ambiente, comando, versão de ferramenta, SHA, hash de diff quando houver, digest de imagens,
IDs de cenários, pass/fail/skip, logs saneados e hashes.
Um working tree não commitado deve ter manifesto; usar somente HEAD produziria atribuição incompleta.
Atualizar hashes após correção e repetir somente validações materialmente afetadas, depois gate final.
Cenários normativos/documentais podem ter inspeção verificável; cenários de efeito exigem oráculo real.

## Tratamento de bloqueios
Ferramenta ausente: instalar localmente versão compatível quando permitido.
Fixture defeituosa: corrigir bootstrap e testar. Decisão técnica reversível: escolher e registrar ADR.
Credencial/custo externo: usar fixture sintética e deixar homologação específica pendente.
Mudança normativa: não aprovar por conta própria; registrar opções e concluir partes independentes.
