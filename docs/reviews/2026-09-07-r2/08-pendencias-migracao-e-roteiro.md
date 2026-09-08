# Pendências, migração e sequência da implementação

Não se presumiu que qualquer decisão pendente tenha sido resolvida. P-01 a P-11 são os IDs vigentes na baseline v4. D-01 a D-07 eram lembretes da proposta anterior; não devem substituir ou fechar os P-* por coincidência de assunto.

| Pendência vigente | Responsável funcional herdado | Evidência necessária | O que bloqueia |
|---|---|---|---|
| P-01 | Plataforma | Conta/região, referência EKS ou operação própria, orçamento, registro OCI, gateway/licenças | IaC remota e promoção cloud; não laboratório |
| P-02 | Produto | Portfólio, RPS/picos, fan-out, tamanhos, duração, mix, perfis legados e isolamento | Compromisso de capacidade/SLO e ofertas reais |
| P-03 | Comercial | Compra/venda, tarifas, faixas/franquias, marcos, parcialidade e failover | Ativação econômica; usar fixtures para engenharia |
| P-04 | Responsável pelos dados + Segurança | Classes, finalidades, retenção, região e custódia de resposta | Dados reais e expurgo/publicação dependentes |
| P-05 | Integrações | Idempotência, modalidades, polling/callback, credenciais, capacidade e equivalência por parceiro | Adapter/rota/failover homologados |
| P-06 | Financeiro | ERP/layout, contas gerenciais, moeda/arredondamento, competência e liquidação | Fechamento/exportação comercial |
| P-07 | Produto | SLA bilateral, TTL, enforcement, reserva final e entrega | Prometer/ativar SLA por oferta |
| P-08 | SRE | Domínio de falhas, autoridade, RPO/RTO, recuperação e custo/latência | Prd e promessa de RPO zero regional |
| P-09 | Liderança técnica | Nomes e concessões administrativas, responsáveis e aprovação do escopo | Aprovação/ativação da fatia; não criar conta compartilhada |
| P-10 | Produto e responsável pelo sistema consumidor | Máxima interrupção, contingência e criticidade por caso | Ofertas com impacto em vida/segurança |
| P-11 | Plataforma | Quotas autorizadas, envelope, N−1, horizonte/headroom e células | Qualificação de expansão integral em ppd |

Além das pendências herdadas, T-R2-01 a T-R2-05 estão no documento de arquitetura, incluindo prova da fronteira de commit/deadline e coordenação de capacidade. São resultados técnicos exigidos, não autorização para enfraquecer requisitos.

## Ordem de execução e contratos compartilhados

| Faixa | Changes | Entrega e dependências |
|---|---|---|
| Preparação | 09, tarefas de baseline/fixtures/gates | Iniciar imediatamente; não aguardar o final para construir provas |
| Fundação crítica | 01 e 02 | Identidade/autorização, custódia, posse, resposta real e prazo. Desenvolvimento pode usar interfaces acordadas, mas integração final exige ambas |
| Controle e integrações | 04 e 03 | 04 fixa schemas/versões/ofertas; 03 consome contratos e fecha adapter/credenciais/polling/callback/pressão. Pactuar DTOs primeiro; sem mocks como aceite |
| Produto administrativo | 05 | Shell, sessão e listas após 01; jornadas dependem das APIs de 02/03/04/06/08. Não bloquear toda UI pelo financeiro; não declarar tela indisponível como implementada |
| Custódia e econômico | 07 e 06 | Objetos/retenção dependem de 01/02; financeiro depende de snapshots de 04 e fatos de 02/03. Hold e efeito precisam teste integrado |
| Operação | 08 | Laboratório pode ser preparado cedo; qualificação final depende de runtime seguro/recuperável. Separar implementação de manifests de aprovação cloud |
| Fechamento | 09 | Reexecutar cenários afetados, reconciliar 98 IDs e gates pertinentes; só então concluir change/ativar perfil |

Não há sprint, prazo ou estimativa de esforço inventados. Dependências são técnicas. As nove changes formam um backlog de engenharia, não nove implantações independentes autorizadas em produção.

## Contratos que precisam ser acordados antes da codificação correspondente

- Identidade e escopo entre Portal e cada domínio; APIs de sessão/permissão da UI.
- DTO de snapshot e publicação Atlas → execução, incluindo versão de contrato, prazo, perfil e referências de binding.
- Command/Result duráveis e envelope, com command_id/operation_id/epoch e recuperação da resposta.
- Recibo externo e elegibilidade de final, inclusive T-R2-01; separar tempo do provedor e custódia do Hub.
- Unidade econômica e contrato de reserva/hold; settlement_party não depende de tipo de credencial.
- Representação final imutável e FileRef; retorno GET/webhook preserva perfil legado.
- APIs paginadas, detalhe/ETag, simulação/publicação, timeline, SLA e financeiro usadas pela UI.
- Contrato de capacity_domain/lease/placement e sinais consumidos por autoscalers.

## Migração compatível

Inventário antes de alterar dados; registrar quais protocolos foram gerados pelo simulador e quais são de integração real comprovada. Se só existem dados de ensaio e o proprietário autorizar reset, esse é procedimento separado e explícito; esta revisão não autoriza excluir dados. Por padrão, manter histórico e avançar por migrations adicionais.

Aplicar esquema novo compatível, backfill retomável com amostras/contagens/hashes, leituras compatíveis durante transição e cutover por feature flag/política. Preservar bytes finais entregues, identity keys, outbox/inbox e responsabilidade econômica. Se resposta externa antiga não foi persistida, não há como reconstruí-la com certeza do input: marcar qualidade histórica e reconciliar fonte permitida. Evitar reexecutar operação como “migração”.

Identidade: migrar consumidores para tokens reais; nunca liberar header arbitrário como fallback remoto. Credenciais: introduzir resolver real, invalidar cache legado por conta e qualificar identidade no parceiro. Catálogo: converter importados em staging/homologação; sem delete global. Financeiro: lote de abertura e proveniência dos fatos legados; sem repricing pelo contrato corrente. Infra: recursos de células recebem namespaces e um dono; cutover cerca antigo writer/consumidor antes de permitir nova autoridade.

Rollback de código não remove tabelas/campos/obrigações ainda usados; estorno financeiro é compensação, não DELETE. Para incompatibilidade de worker, parar novas admissões afetadas, drenar, preservar UNKNOWN/holds e escolher versão compatível. Nunca desabilitar segurança/fencing para “voltar rápido”.
