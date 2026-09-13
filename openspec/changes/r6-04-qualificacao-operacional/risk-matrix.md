# Riscos e mitigação

| Risco | Prioridade | Mitigação | Aceite |
|---|---|---|---|
| Gate de promoção vinculado ao artefato e à cobertura | P0 | O Hub SHALL bloquear promoção quando qualquer evidência obrigatória não corresponder ao artefato candidato, ao cenário vigente e à execução autorizada. Status SHALL usar enum estrito; cobertura parcial e declarações autoatribuídas não qualificam promoção. | R6-OPE-01-S01…S03 |
| Restore populado com versões e retomada reconciliada | P0 | O Hub SHALL demonstrar restore não vazio de todas as autoridades e obrigações, incluindo versões de objetos referenciadas, seguido de retomada cercada e reconciliação de efeitos e valores. A ausência de dados ou a mera igualdade de contagens SHALL impedir aprovação de recuperação integral. | R6-OPE-02-S01…S03 |
| Ambientes e elasticidade com orçamento de dependências | P1 | O Hub SHALL fornecer perfis reproduzíveis local/dev/hom/ppd/prd com dependências e limites explícitos, escalabilidade automática dentro do envelope qualificado e continuidade mensurável. Readiness e topologia de obrigações SHALL refletir capacidade real de admitir/processar com custódia, sem confundir laboratório efêmero com HA produtiva. | R6-OPE-03-S01…S03 |
| Capacidade obrigatória e feedback recuperável | P1 | O Hub SHALL exigir política efetiva de capacidade para rotas ativas e conservar obrigações de feedback/settlement até resolução. Controle adaptativo SHALL respeitar teto seguro contratado e compartilhar capacidade de forma justa entre tenants, com custo estável em função do estado ativo. | R6-OPE-04-S01…S03 |
