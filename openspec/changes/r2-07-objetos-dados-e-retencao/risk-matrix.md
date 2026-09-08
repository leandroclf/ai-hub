# Riscos: Arquivos, propriedade de dados, retenção e recuperação

Estado de todos os riscos: **aberto; mitigação especificada, não implementada/qualificada nesta revisão**. P0 bloqueia exposição/ativação do fluxo afetado; P1 bloqueia completude da capacidade; P2 qualifica usabilidade/governança. Classificação está vinculada ao código e ao uso proposto, sem afirmar exploração em ambiente remoto.

| ID/prioridade | Risco e impacto | Mitigação/prova exigida | Responsável | Evidência |
|---|---|---|---|---|
| F-29 / P1 | Payload grande pode esgotar memória; sucesso do teste S3 não prova custódia/atomicidade de arquivo referenciado por protocolo. | Upload direto e finalização validam proprietário/tamanho/checksum; referência imutável antes de sucesso; broker leva só metadados; falha de S3 afeta apenas operações que dele dependem. | Dados, Segurança e Plataforma | [hub/internal/objectstore/objectstore.go:43](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/internal/objectstore/objectstore.go#L43) |
| F-30 / P1 | Permissões excessivas e exclusão sem retenção de obrigações comprometem isolamento e futura recuperação. | Papéis de aplicação/migração separados; RLS testada com reutilização de pool; expurgo protege obrigações; restore com egress bloqueado reconcilia core/finance/objetos antes de retomar efeitos. | Dados, Segurança e Plataforma | [hub/deploy/postgres-init/01-init.sh:8](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/postgres-init/01-init.sh#L8) |

Risco de integração: dependências r2-01-identidade-e-isolamento, r2-02-execucao-duravel-e-resultados. Ver também riscos transversais de identidade, custódia e rollback no design.
