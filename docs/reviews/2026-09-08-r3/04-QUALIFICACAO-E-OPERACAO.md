# Plano de qualificação e operação

## Gates da próxima implementação
| Gate | Ambiente | Exigência |
|---|---|---|
| G0 | checkout limpo | Instruções, SHA, status Git, ferramentas fixadas e OpenSpec strict. |
| G1 | Go/Node | Unitários, race, vet, build e regressões de precisão/schema. |
| G2 | Dependências reais locais | PostgreSQL com fsync/synchronous_commit, mensageria, objetos/cofre emulados e OIDC. Sem skip obrigatório. |
| G3 | Runtime completo | Imagens do SHA, provider-sim e webhook-sink externos, jornadas SYNC/ASYNC/AUTO e consumidor/operador. |
| G4 | Browser | Todas as jornadas administrativas da matriz, negativas de autorização e reload. |
| G5 | Kind completo | Componentes/dependências no cluster; probes, DNS, volumes, TLS/políticas, HPA e telemetria funcionais. |
| G6 | Falha/carga/restore | Matriz abaixo, oráculos independentes e SLO/envelope explicitado. |
| G7 | Candidato remoto | Compatibilidade AWS real, quotas/IAM/zonas/DR, contratos comerciais e segurança da promoção. Não executar ações comerciais sem autorização. |

## Matriz de experimentos obrigatórios
1. SYNC provedor síncrono: resposta real na mesma chamada, UUIDv7 e GET posterior iguais; broker desligado no caminho direto.
2. ASYNC sobre provedor síncrono: aceite durável, callback final com mesmo body de GET.
3. Provedor assíncrono com polling+callback: corrida, duplicata, callback órfão e final tardio.
4. Quedas antes/depois de intenção, antes/depois de efeito e antes/depois de ACK/commit; contador externo comprova efeitos.
5. Pausa real entre decisão SQL e commit no deadline, conforme ADR R2; não esconder contraexemplo.
6. Redis/cofre/Atlas degradados: quente/frio, token válido/vencido/revogado e outro tenant saudável.
7. Conta compartilhada com vários pods: grants globais, 429/timeouts, capacidade variável, recuperação amortecida e justiça.
8. Portfólio acima de 100 ofertas, modos múltiplos, DAG paralelo e contrato legado com grandes inteiros.
9. Saldo/franquia concorrentes, tarifa diferente da reserva, UNKNOWN e reconciliação, incidência STATUS/FETCH e fechamento.
10. Arquivos grandes, hashes/versionamento, acesso cruzado, pins, orphan recovery, purge e restore.
11. Dois tenants/aplicações, token consumidor contra admin, papel global com/sem MFA e revogação.
12. Rollout, perda de pod/nó/zona na plataforma apropriada, scale-out, backlog e drenagem.
13. Métrica de obrigação, scrape, alerta, log Loki, trace Tempo e consulta Grafana/API com correlação.
14. Inicialização do broker em ordem adversa: publicar antes de assinatura obrigatória, topologia ausente e DLQ/retention.
15. Restore que antecede efeito externo: reconciliar antes de autorizar novo SUBMIT.

## Perfil de carga
Enquanto D-01 não for aprovado, criar perfis SINTÉTICOS nomeados, nunca chamá-los de demanda real.
Publicar tenants ativos, ofertas por tenant, provedores/contas, taxa/concorrência, duração, tamanho de payload,
percentis do provedor, política econômica e envelope de CPU/memória/conexões/IOPS/filas.
Comparar base versus tenant agressor e repetir com falhas. Dados comerciais aprovados substituem fixtures posteriormente.
Não declarar p99 ou SLO não medido. Não declarar carga encerrada só porque CPU está baixa.

## Formato da evidência
Guardar em hub/evidence/r3/ ou caminho aprovado nas instruções atuais:
SHA, digest de cada imagem, versões, comando exato, timestamp, ambiente, parâmetros, resultado,
pass/fail/skip, IDs de cenários e hashes de logs saneados. Não publicar tokens ou payload pessoal.
Contadores do provider-sim e webhook-sink são oráculos de efeito e bytes. Consultas financeiras verificam conservação monetária.
A evidência deve corresponder às imagens atuais, não a código editado depois da execução.

## Obrigatoriedade e bloqueios
Unitários podem ter casos opcionais, mas gate integrado obrigatório falha se não houver dependência/fixture.
Corrigir bootstrap OIDC com ambiente isolado e reproduzível; não apagar volumes de usuário para limpar fixture.
Credencial comercial ausente bloqueia apenas homologação dessa integração. O resto deve continuar com fixture real local.
Ao final, nenhum skip obrigatório, P0 aberto ou decisão normativa crítica pendente é compatível com declaração de pronto.
