# Riscos e mitigação

| Risco | Prioridade | Mitigação | Aceite |
|---|---|---|---|
| Snapshot de etapa coerente em todas as identidades | P0 | O Hub SHALL materializar cada etapa com rota, conta, binding, perfil, contrato de compra, prazo e incidência coerentes e congelados. Se a oferta não contém referências suficientes para resolver a etapa, a publicação ou admissão SHALL ser recusada antes de qualquer efeito. | R6-EXE-01-S01…S03 |
| Expiração bloqueia também recuperação DIRECT | P0 | O Hub SHALL barrar nova submissão com possibilidade de efeito depois do prazo aplicável em DIRECT, QUEUED e recuperação. A reconciliação de efeito possivelmente já realizado SHALL continuar pela operação de consulta autorizada, sem se confundir com reenvio. | R6-EXE-02-S01…S03 |
| Compensação progride após final público | P0 | O Hub SHALL separar estado de atendimento do cliente e estado de obrigações compensatórias. Encerrar protocolo não pode impedir compensação já devida; compensações SHALL respeitar causalidade reversa, política própria e reconciliação sem ressuscitar o final público. | R6-EXE-03-S01…S03 |
| Liberação de slot e intent na mesma transação | P0 | O Hub SHALL atualizar intent, lease da etapa e contador do plano atomicamente também na conclusão, falha e expiração. Takeover e fatos concorrentes SHALL preservar max_parallel, idempotência e recuperabilidade sem zerar contadores artificialmente. | R6-EXE-04-S01…S03 |
