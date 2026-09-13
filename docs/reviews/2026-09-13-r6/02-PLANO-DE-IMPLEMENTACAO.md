# Plano de implementação R6

| Ordem | Fatia | Dependência/saída |
|---|---|---|
| 0 | Fixar candidato e laboratório | Ler AGENTS, inventariar alterações, preservar dados; ferramentas e logs antes de editar. |
| 1 | SEG e FIN-01 | Custódia/isolamento corretos habilitam testes sem contaminação e replay seguro. |
| 2 | EXE-01…04 | Publicação de snapshots coerentes antes da qualificação dos fluxos, compensation e contadores. |
| 3 | FIN-02 e OPE-04 | Receber fatos corretos; qualificar liquidação, completude, capacidade e feedback. |
| 4 | UX-01/02 | Integrar contratos corrigidos e executar jornadas browser, sem substituir backend por mock. |
| 5 | OPE-02/03 | Restore populado, perfis Compose/kind e escala/HA dentro dos budgets. |
| 6 | OPE-01 e QUA-01 | Harness pode ser preparado cedo; fechamento depende de evidência real do candidato. |

## Organização do trabalho
Há 72 tarefas nos seis changes. Executar fatias verticais: reproduzir → corrigir caminho real → qualificar falha → registrar evidência. Dividir tarefas grandes por transação/endpoint sem renumerar requisitos. Não parar porque a alteração compila ou porque uma sessão atingiu um checkpoint.

## 25 verificações obrigatórias de continuidade
Cada item abaixo é trabalho a conferir; reutilizar evidência somente quando escopo e artefato corresponderem. Não reimplementar uma regra já comprovada.

- [ ] C-R6-01 — R3-EXE-01: Adapters REST/provedores externos: homologar capacidades e autenticação por contrato; preservar rest-json-v1.
- [ ] C-R6-02 — R3-EXE-02: Callbacks: preservar autenticação pré-custódia; fechar rotação/retomada de R6-SEG-02.
- [ ] C-R6-03 — R3-EXE-03: UNKNOWN/fencing: provar crash após envio antes de recibo, reconciliação por chave e ausência de reenvio cego.
- [ ] C-R6-04 — R3-EXE-04: Prazos: R6-EXE-02; observar polling saudável e callback tardio sem reabrir final.
- [ ] C-R6-05 — R3-EXE-05: Mensageria: preservar topologia inicial completa; qualificar ACK/quarentena recuperável e drift após startup.
- [ ] C-R6-06 — R3-CAT-01: TransformJSON e precisão do fato corrigidos; preservar regressões e validar contrato final completo.
- [ ] C-R6-07 — R3-CAT-02: Contratos de entrada/saída por cliente e perfil: validar snapshot e representação GET/webhook idêntica.
- [ ] C-R6-08 — R3-CAT-03: Produto persistido: R6-EXE-01/03/04; qualificar paralelo, parcialidade e consolidação publicada.
- [ ] C-R6-09 — R3-CAT-04: Oferta/cache: preservar cache-first e negação; comprovar janela de revogação e isolamento multi-réplica.
- [ ] C-R6-10 — R3-INT-01: Capacidade adaptativa: R6-OPE-04 com teto externo, múltiplos tenants e recuperação de permits.
- [ ] C-R6-11 — R3-INT-02: Credencial compartilhada/dedicada: provar rotação, ausência de fallback cruzado e token L1 sob cofre indisponível.
- [ ] C-R6-12 — R3-INT-03: Egress: qualificar SSRF, DNS/redirect, mTLS e limites de conexões/memória sob carga.
- [ ] C-R6-13 — R3-ADM-01: Administrador nominal/MFA/escopo: R6-SEG-01 e autorização dos recursos auxiliares.
- [ ] C-R6-14 — R3-ADM-02: Operações/SLA/deliveries: jornada browser com consulta/retry autorizado, conflito e frescor.
- [ ] C-R6-15 — R3-ADM-03: Finanças: IDs textuais preservados; qualificar períodos, contestação, dupla aprovação e intenção de mutação.
- [ ] C-R6-16 — R3-ADM-04: Catálogo: R6-UX-01/02 e formulários específicos de credenciais, serviços/produtos e vigência.
- [ ] C-R6-17 — R3-FIN-01: Captura efetiva ligada: provar reserva/franquia/hold/excedente com valores exatos e política contratada.
- [ ] C-R6-18 — R3-FIN-02: Incidências e fechamento: R6-FIN-01/02, sem transformar observação em cobrança indevida.
- [ ] C-R6-19 — R3-FIN-03: Webhook cliente: destinos e bytes congelados, retry idempotente, rotação e falha após envio.
- [ ] C-R6-20 — R3-OPE-01: Arquivos: upload/download autorizado, FileRefs, streaming, retenção/pins/versões e restore R6-OPE-02.
- [ ] C-R6-21 — R3-OPE-02: Kind independente existente: qualificar todos os componentes no perfil certo e documentar diferenças remotas.
- [ ] C-R6-22 — R3-OPE-03: HPA/KEDA: provar escala de pods/nós/dependências e justiça por cliente dentro do envelope.
- [ ] C-R6-23 — R3-OPE-04: RLS e restore: R6-SEG-01/OPE-02 com autoridades populadas e retomada.
- [ ] C-R6-24 — R3-OPE-05: SLA bilateral: coortes/tempo de provedor versus Hub/cliente, alertas e limites de cardinalidade.
- [ ] C-R6-25 — R3-QUA-01: Inventário completo: R6-QUA-01 e evidência por cenário sem skip obrigatório.

## Encerramento
Um requisito só fecha com cenários e oráculos correspondentes. Propriedade distribuída exige teste integrado; go -race verifica acesso concorrente em memória e não prova transação PostgreSQL, perda de mensagem ou HA.

Decisões externas bloqueiam apenas a promoção ou configuração que depende delas. Implementar mecanismos com fixtures explicitamente locais; não inventar aprovação nem interromper trabalho independente. Se existir bloqueio externo irredutível, concluir todo o restante e registrar causa, tentativas, evidência e próximo passo mínimo.
