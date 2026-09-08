# Qualidade, cenários integrados e evidências

Os 205 cenários normativos nas nove changes são o aceite detalhado. Esta seção liga os riscos entre componentes e define oráculos; não substitui os cenários das specs.

## Verificações efetivamente feitas nesta revisão

| Verificação | Resultado e limite |
|---|---|
| Snapshot e instruções | Clone de main, SHA a39d394b0d87185ed4cc3861c12ec45f2c302d9e; AGENTS.md e metodologia em docs/openspec-docs lidos |
| Revisão estática | Cinco aplicações, clientes internos, outbox/fila, auth, UI, migrations, contratos, Compose/Kubernetes/Terraform/Crossplane e testes inspecionados |
| Frontend | npm ci com scripts de instalação desabilitados; npm run build passou: TypeScript sem emissão + Vite 5.4.21, 37 módulos, bundle JS 164,12 kB (50,34 kB gzip) |
| Backend | go test ./... passou com Go 1.24.13 baixado com checksum SHA-256 oficial verificado. Apenas internal/atlas e internal/platform/idgen tinham testes unitários executados; demais pacotes sem testes nesse comando |
| Screenshots | Inspecionada captura existente no repositório da tela de serviços. É evidência histórica incluída pelo projeto, não nova sessão E2E nesta revisão |
| Não executado | Testes com build tag e2e, integração com provedor real, Compose, kind, Terraform apply, HPA/KEDA, carga, HA, restauração e pentest; ambiente desta revisão não tinha Docker/kubectl/kind/Terraform inicialmente |
| Validação do pacote R2 | Ver relatório 10-validacao-do-pacote.md gerado após a montagem; validação documental não significa código implementado |

Nenhum arquivo de aplicação foi alterado. Instalações e build geraram apenas dependências/saídas locais ignoradas pelo Git. Nenhuma conta cloud foi provisionada e nenhuma PR foi aberta nesta rodada documental.

## Fixtures mínimas versionadas

Cliente A dedicado, B dedicado na mesma conta, C compartilhado e D sem credencial; leitor restrito, operador e hub_protocol_reader individual. Provedor SYNC com resposta distinta do input; provedor ASYNC com status+callback e principal externo verificável; provedor lento/429/503, resultado inválido, efeito ambíguo e idempotência consultável. Produto simples, agregado A+B e composto A/B→C; passo obrigatório/opcional e compensação. Planos por sucesso, por submit, pacote, franquia/faixa, saldo 2,00 e CLIENT_DIRECT. Arquivo sintético de tamanho parametrizado com checksum. Datas/controlador de relógio e falhas injetáveis sem dependência de sleep arbitrário.

## Testes integrados da próxima implementação

| Caso | Estímulo/falha | Oráculo obrigatório | Changes |
|---|---|---|---|
| IT-01 Identidade | Token A e header/cursor/UUID de B; API interna direto | Nenhum dado/efeito de B; recusa e auditoria; workload indevido também negado | 01/05 |
| IT-02 Principal externo | A/B dedicados compartilham conta; rotação concorrente | Provedor registra identidade externa correta por pedido; nenhum token compartilhado | 01/03/04 |
| IT-03 SYNC sem broker | Broker fora antes do restart e durante chamada | Final real na conexão com UUID, persistência e outbox; GET independente | 02/08 |
| IT-04 202 recuperável | Crash depois de commit e antes de publicação | Intent recuperada, um efeito externo, mesmo UUID; sem abandono silencioso | 02/08 |
| IT-05 ACK seguro | Falhar commit em cada consumidor e repetir evento | Sem Delete antes do commit; uma unidade econômica/delivery/terminal; quarentena quando inválido | 02/03/06 |
| IT-06 Posse concorrente | Dois pods, DIRECT/queue/reconciliador e lease antiga | Uma tentativa autorizada; epoch vencida não atua; UNKNOWN não reenvia | 02/03 |
| IT-07 Final externo | Resposta do provedor contém marcador ausente do input | Marcador chega a POST/GET/webhook e replay; JSON inválido nunca produz sucesso | 02/03 |
| IT-08 Prazo duro | Timer parado; igualdade; commit começa antes e termina depois | Nenhum sucesso tardio publicado, um terminal e prova da ordenação; T-R2-01 resolvido | 02 |
| IT-09 TTL | TTL 0 e TTL 10 s sintético; restart no segundo 7; Retry-After 20 s | Deadline não renova; nenhum retry fora do prazo; incerteza recebe reconciliação | 03 |
| IT-10 Callback/poll | Callback antes da associação e dois finais conflitantes | Recibos preservados, ACK só após custódia, um final válido e conflito visível | 02/03 |
| IT-11 Corpo completo | Upgrade, perfil legado, duas entregas e GET | Bytes completos, hash, versão e HMAC iguais; URL temporária separada | 02/03/04 |
| IT-12 DAG | A/B paralelos, C dependente; B falha opcional/obrigatória | Ordem/overlap corretos; parcialidade e compensação; custo por passo, sem duplicata | 02/04/06 |
| IT-13 Saldo | 2,00 iniciais, duas capturas de 1,00, terceira tentativa; corrida pelo último saldo | Terceiro efeito não ocorre; soma exata; limite ausente nega | 06 |
| IT-14 Hold tardio | Cliente EXPIRED, provedor UNKNOWN e depois custo confirmado | Resposta continua expirada, reserva não liberada cedo, custo único sem receita de sucesso | 02/03/06 |
| IT-15 Preço/snapshot | Publicar tarifa v2 antes do consumo de fato v1 | Valor v1, moeda/contrato/unidade rastreáveis; CLIENT_DIRECT sem passivo indevido | 04/06 |
| IT-16 Objetos | Multipart incompleto, checksum ruim, file_ref alheia, S3 falha antes do final | Sem final com link inválido; sem acesso cruzado; memória limitada | 01/02/07 |
| IT-17 Redis/Atlas | Cache frio, Redis lento/desligado, Atlas fora com projeção válida/expirada | Bypass no material válido; negativa seletiva no expirado; nenhum token Redis; latência medida | 03/04/08 |
| IT-18 Pressão/ruído | Provedor degrada e depois escala; A usa 10× quota, B carga normal | Limite agregado cai/recupera com estabilidade; B mantém SLO e reserva; não multiplicar quota por pod | 03/08 |
| IT-19 Ciclo de deploy | Clone limpo, subida, rollout com tráfego, recriação e retorno | Todos componentes presentes; volumes preservados; intents/leases/UNKNOWN recuperados | 08 |
| IT-20 Expansão | Demanda supera headroom, recurso provisiona com falha parcial | Retry idempotente de onboarding; célula só READY após canário; histórico consultável na origem | 04/08 |
| IT-21 Restore | PITR com efeitos externos posteriores e objetos versionados | Egress bloqueado até reconciliação; ledger/idempotência íntegros, sem replay em massa | 06/07/08 |
| IT-22 Jornada administrativa | Login → cliente → serviço/produto → binding → oferta → protocolo → SLA → extrato | Dados persistem após refresh; permissões e estados reais; zero dependência de mock no aceite | 01/04/05/06/08 |
| IT-23 Exportação | Lote com pendência, depois fechado; recibo ERP perdido | Fechamento bloqueia pendência; export_id/hash estáveis; sem duplo pagamento | 06 |
| IT-24 Recibo indisponível | Callback autêntico e writer fora | Nenhum 2xx sem custódia; retransmissão reconcilia uma vez | 02/03 |

## Gates e definição de pronto

G0: escopo da change, contratos e responsável; decisões abertas registradas. G1: tipos/schemas/OpenSpec e compilação. G2: testes de domínio com oráculos e DB real onde transação importa. G3: integração/falhas/concorrência e segurança multi-tenant. G4: jornadas UI reais, acessibilidade e contrato público legado. G5: laboratório completo e promoção reprodutível. G6: carga/ruído/escala/recuperação no perfil e infraestrutura representativos. G7: aprovação das pendências aplicáveis e ativação comercial/crítica.

G1 nunca substitui G2–G7. Changes 01–04/06–08 entregam suas provas ao gate integrado; change 09 prepara a infraestrutura de testes desde o início e fecha a qualificação ao final. Cada tarefa só recebe [x] com evidência, commit e revisor. Falta de dado comercial não bloqueia corrigir bug genérico; bloqueia vender/ativar perfil dependente.

## Registro mínimo de execução

Para cada cenário: ID, baseline/R2, SHA, data/hora UTC, executor, ambiente/versões, fixture e hash, carga/mix/tamanho, pré-condições, falha injetada, observações de DB/broker/parceiro, outputs sanitizados, esperado/obtido, pass/fail/blocked, causa, links e risco residual. Ensaios de carga informam número de amostras, aquecimento, duração, percentis ponta a ponta e recursos; não somar percentis de etapas. Ensaios de preservação comparam contagens e identidades aceitas, concluídas, pendentes e reconciliadas, não apenas HTTP 200.
