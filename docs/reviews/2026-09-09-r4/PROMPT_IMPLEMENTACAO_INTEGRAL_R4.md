# Missão de implementação integral — AI Hub R4

Continue o repositório real e conclua o trabalho técnico executável v4/R2/R3/R4.
Esta é uma missão de implementação e qualificação; não entregue apenas novos documentos,
helpers desconectados ou relatório repetindo pendências.

## 1. Estado e instruções
Leia AGENTS.md e instruções aplicáveis, docs/openspec-docs, a baseline v4, os nove changes R2,
os sete R3 e os quatro R4. Leia todos os documentos de docs/reviews/2026-09-09-r4/.
Snapshot auditado: b9d0f90ce02aa0c27cad546745153d160ff5867f. Se HEAD mudou, revalide o delta e preserve correções.
Inspecione git status/diff e trabalho não commitado antes de editar. Não resetar, limpar
destrutivamente nem sobrescrever alterações preexistentes.

A R4 não substitui a arquitetura v4. Há 201 requisitos e 732 cenários na união.
Os 696 cenários herdados incluem 416 da v4 anteriormente omitidos da matriz.
Use os inventários entregues e regenere a partir das specs para detectar mudanças.
IDs V4:<requisito>:S<ordem> são identidade de auditoria, não alteração da norma original.

## 2. Regra operacional atual obrigatória
AGENTS determina no máximo um ecossistema Compose do AI Hub ativo por vez.
Esta regra prevalece sobre sugestões históricas de laboratórios Compose paralelos.
Antes de iniciar/trocar, execute docker compose ls e docker ps -a; identifique o projeto do Hub.
Use nome estável para o projeto oficial local. Ao substituir, pare apenas o ecossistema
identificado, confirme encerramento e só então suba o novo.
Preserve volumes/dados por padrão e não toque em componentes de terceiros.
Não use docker system prune nem curingas amplos.
Limpe somente recursos da execução identificados e dispensáveis; registre exceções no checkpoint.
No perfil kind completo, dependências devem estar no cluster; não abrir Compose auxiliar paralelo.

## 3. Autonomia e limites
Está autorizado a editar código, contratos, testes, migrações, UI, scripts e configurações.
Pode instalar ferramentas compatíveis em escopo local, preparar fixtures sintéticas e executar
testes/carga/falhas locais dentro da capacidade disponível. Tome decisões técnicas reversíveis
compatíveis com os requisitos e registre motivo/alternativas.
Não peça confirmação para cada tarefa já autorizada.

Não está autorizado a apagar dados preexistentes, usar operações comerciais reais,
provisionar recursos pagos, contornar permissões, expor segredos ou fazer push/merge/deploy remoto.
Não aprove decisões comerciais nem altere semântica normativa de SLA em nome dos responsáveis.

## 4. Preparação e desbloqueio
Classifique cada pendência como implementação, ferramenta, fixture, decisão técnica ou dependência externa.
Código ausente deve ser implementado; ferramenta ausente investigada/instalada; fixture corrigida.
Não pare apenas porque o relatório anterior disse “não executado”.

Inventarie Go/Node, Docker/Compose, PostgreSQL, LocalStack, OIDC, OpenSpec, kubectl/kind e browser.
OpenSpec 1.12.0 validou os 17 changes existentes nesta revisão sem config.yaml.
Instale versão fixada compatível ou use a já definida no projeto; não use latest sem qualificação.
Execute help e strict, corrigindo estrutura sem reduzir requisitos.

Verifique daemon/socket/permissões se Docker falhar. Não contorne restrição de plataforma.
Processos locais podem validar partes independentes, mas não equivalem a aprovação de Compose/kind.
Prepare OIDC reproduzível sem apagar volumes do usuário. Use provider-sim e webhook-sink como oráculos externos.
Registre SHA mais hash do working tree e digests de imagens: HEAD sozinho não identifica alterações locais.

## 5. Corrija primeiro os achados R4
A. Upgrade: 0002_provider_auth.sql histórica foi alterada. Runner checksum-guardado pode parar
antes da 0004. Preserve histórico e prove upgrade do banco R2, instalação limpa e variante R3 conhecida.
Não resolva apagando volume ou atualizando todos os checksums.

B. Callback: a URL devolvida carrega capability, mas rota segue sob JWT Hub. Feche o caminho real
com autenticação homologada da conta. Não basta testar handler sem middleware/gateway.

C. Inbox: autenticar origem antes de admitir órfão; escopo/quota/retention explícitos.
Deduplicação não pode permitir que tentativa inválida ocupe identidade de evento legítimo.

D. Recuperador: worker autônomo com lote/claim/lease/epoch e justiça por escopo.
Não depender de outro callback; não drenar backlog global no HTTP nem manter cursor ocupando
pool enquanto faz aplicações aninhadas. Isolar poison item.

E. Resultado: unificar schema/correlação/conta/snapshot em SUBMIT/poll/callback.
Não sobrescrever provider_request_id divergente nem truncar resposta para detail.
Recibo bruto protegido e resultado normalizado são registros distintos.

F. JSON: preservar precisão já corrigida, recusar null como string, número em string e
documentos concatenados; validar EOF com/sem mapping. Publicar dialeto explícito:
keywords não suportadas devem ser recusadas, nunca ignoradas.

G. Cache: L1 válido e autorizado antes do cofre, com revogação/expiração; coordenação
cancelável e limitada por binding/versão. Redis não armazena tokens.

H. Oferta: não juntar todas as páginas em memória para filtrar depois.
Resolver conjunto elegível por índice/versão/vigência e projeção válida, com prova de custo.

I. Evidência: corrigir omissão dos 416 cenários v4 e inconsistência relatório/checkpoint.
Mensagem de merge não é prova de conclusão.

## 6. Conclua também o backlog herdado
Execute todas as 25 fatias B-R4 de 02-BACKLOG-DE-CONCLUSAO.md e tasks do change r4-04.
Não encerre a rodada após apenas os 12 requisitos novos.

Obrigatoriamente:
- Adapter real e capacidade homologada, além de synthetic-provider.
- UNKNOWN/SUBMITTING recuperáveis e fencing antes de efeito.
- Topologia, outbox/inbox e quarentena financeira recuperável.
- SYNC direto com final na mesma chamada e UUIDv7 após aceite.
- ASYNC sobre provedor síncrono e polling+callback combináveis.
- TTL desde primeira falha transitória sem renovação; relógios separados.
- Snapshot efetivo, versão solicitada, OutputMapping e igualdade GET/webhook.
- DAG/agregação recuperável, paralelismo limitado, parcialidade e compensação.
- Controle adaptativo global ligado a SUBMIT/STATUS/FETCH e pools reutilizados.
- Saldo/franquia pré-efeito, captura efetiva, unidades de incidência e fechamento.
- Destinos congelados e redelivery sem novo SUBMIT.
- Objetos/streaming/pins/retention e restore reconciliado.
- Console operacional completo.
- Compose/kind completos, ambientes, probes/drenagem, escala e sinais reais.

## 7. Console: provar jornadas reais
Nenhum arquivo do frontend mudou no delta auditado; os problemas anteriores permanecem.
Corrigir delivery_id/protocol_id, SLA/reconcile, DTO financeiro, tenant/ator/datas,
chave idempotente da intenção, multimodalidade, auth condicional, fuso e lookups.
Validar JSON em runtime e não transformar resposta inválida em sucesso.
Implementar APIs ausentes antes de considerar botão funcional.
Testar navegador, salvar/reabrir/recarregar e confirmar efeito no backend/provedor/ledger.
API é autoridade de acesso: consumidor contra admin, tenants/aplicações distintos e
operador nominal global com/sem MFA devem ter provas negativas.
Não reabrir bypass já corrigido; preserve MFA/papel atuais e refine escopos/mascaramento.

## 8. Engenharia de validação
Para cada fatia:
1. Reproduzir defeito ou confirmar lacuna no HEAD atual.
2. Fixar contrato e teste de regressão independente.
3. Implementar o caminho real, dados e recuperação.
4. Executar testes pertinentes e falhas/concorrência.
5. Corrigir até cumprir invariantes.
6. Registrar evidência e avançar.

Use probes fornecidos por overlay como regressões, sem chamar falha esperada de PASS.
Separe unitários de integração obrigatória; dependência ausente deve falhar o gate integrado.
Não remover testes, enfraquecer assertions ou marcar t.Skip como aprovação.
Não exigir arbitrariamente um teste automatizado para cada cenário documental:
registre inspeção verificável quando apropriada; cenários de runtime exigem execução/oráculo.

Execute Go test/race/vet, frontend, OpenSpec, DB, broker/objetos/cofre/OIDC,
fluxos externos, browser, Compose/kind e experimentos de carga/restore pertinentes.
Oráculos: contador de efeitos no provedor, bytes/hash de webhook, estado pós-reinício,
conservação monetária, acesso negado e métricas globais de capacidade/custódia.
Não faça testes extras sem risco concreto depois de qualificar a mudança; faça o gate final integral.

## 9. Decisões externas e T-R2-01
D-01…D-07/P-01…P-11 não são aprovação implícita por existir código.
Fixtures sintéticas permitem engenharia; não são contrato comercial ou SLO aprovado.
T-R2-01 exige continuar investigação e reproduzir pausa entre decisão SQL e commit.
Não substituir confirmação durável no prazo por checagem anterior ao commit.
Se solução exigir alteração normativa, documentar opções/prova/impacto e manter o gate específico
aberto, concluindo os demais. “Implementar tudo” não autoriza inventar garantia impossível.

## 10. Registros e conclusão
Criar docs/reviews/2026-09-09-r4/implementation com:
EXECUTION_PLAN.md, REQUIREMENTS_STATUS.csv, SCENARIO_RESULTS.csv, FINDINGS_STATUS.md,
DECISIONS_AND_BLOCKERS.md, CHECKPOINT.md, EVIDENCE_INDEX.md e FINAL_REPORT.md.
Preservar relatórios históricos como históricos, sem promover evidência antiga para conteúdo novo.
Manter relação requisito→cenário→tarefa→código→comando→resultado→hash/digest.

Concluir integralmente só quando obrigações aplicáveis forem comprovadas, P0 estiverem fechados,
zero skip obrigatório, jornadas reais funcionando, upgrade/restore demonstrados, OpenSpec strict
válido e diff revisado. Relatório e checkpoint devem concordar.
Se houver interrupção, checkpoint preciso permite retomar; interrupção não é conclusão.

Se restar bloqueio incontornável, primeiro concluir todo trabalho independente, nomear recurso/
permissão/decisão exatos, tentativas e erro, e deixar gate reproduzível. Informar implementação parcial.
Não criar outra rodada de specs para substituir código pendente.
Não fazer push, merge ou implantação remota.
