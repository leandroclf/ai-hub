# Decisões e dependências R2

Registro em andamento. As condições comerciais/externas P-01 a P-11 não foram presumidas aprovadas. Fixtures locais não autorizam contratos reais ou produção.

- D-R2-01: conservar Go/PostgreSQL e as cinco autoridades. Sem troca de broker/stack. Base normativa v4 permanece aberta.
- D-R2-02: laboratório separado da stack já existente. Projeto/portas/volumes R2 próprios e migrações aditivas.
- D-R2-03: OIDC Authorization Code + PKCE na SPA; JWT verificado em cada API; credenciais próprias por workload. Tokens administrativos só em memória.
- T-R2-01: ainda não resolvido. Provar confirmação durável anterior ao prazo, com commit lento e incerteza de relógio; sem prova, não publicar sucesso. Não aceitar `now()` no início como solução.
- T-R2-02: homologação externa HivePlace depende de sandbox/credenciais/contratos verificáveis ainda não apresentados. Implementar adapter configurável e testes sintéticos antes de classificar a fronteira final de homologação.
- T-R2-03: fluxo OIDC em implementação; qualificação end-to-end pendente.
- T-R2-04: orçamento compartilhado por domínio de capacidade em PostgreSQL proposto; experimentos de concorrência e takeover pendentes.
- T-R2-05: preservar baseline e IDs; nenhuma cópia de specs será tratada como conclusão de software.

O limite de uso interrompeu agentes em 07/09 na sessão inicial. Retomada solicitada pelo usuário; trabalho preservado e execução continua. Não é justificativa para marcar requisito funcional como bloqueado ou aprovado.
