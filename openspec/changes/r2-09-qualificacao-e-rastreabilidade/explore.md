# Explore: Qualificação por requisito e evidência reproduzível

## Fatos observados

Base fixa: `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`; evidência estática do código, não ataque executado ou ensaio de HA.

| Achado | Prioridade | Problema | Evidência |
|---|---|---|---|
| F-37 | P1 | Ensaios de HA, isolamento e recuperação não foram demonstrados | [hub/deploy/k8s/README.md:106](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/deploy/k8s/README.md#L106) |
| F-38 | P1 | Cobertura existente não prova as garantias que os nomes dos testes sugerem | [hub/test/e2e/e2e_test.go:78](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/test/e2e/e2e_test.go#L78) |
| F-39 | P2 | Rastreabilidade histórica contém conclusões que não refletem o snapshot | [openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md:68](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/openspec/changes/hub-interoperabilidade-v4/TRACEABILITY.md#L68) |
| F-40 | P2 | Versões e contratos de ferramenta carecem de qualificação operacional | [hub/go.mod:3](https://github.com/leandroclf/ai-hub/blob/a39d394b0d87185ed4cc3861c12ec45f2c302d9e/hub/go.mod#L3) |

## Premissas explícitas

- Brownfield: conservar a implementação e o alvo v4; não assumir que os relatórios de conclusão são prova de todas as regras.
- Alterações descritas são futuras. Todas as tarefas deste pacote permanecem abertas.
- Valores comerciais, perfis de carga e credenciais reais não foram fornecidos; usar fixtures e gates por perfil.

## Decisão proposta

Converter o alvo de 98 requisitos em entregas demonstráveis e impedir que teste superficial, manifesto de referência ou screenshot se torne prova de produção.

Dependências: Sem pré-requisito de implementação para iniciar; integrações específicas descritas abaixo.

## Riscos e perguntas técnicas

P-01 a P-11 permanecem com responsáveis de função, sem nomes inventados. Falta de decisão limita ativação específica; não é justificativa para declarar funcionalidade incompleta como pronta.

Ver [arquitetura e trade-offs](../../../docs/reviews/2026-09-07-r2/03-arquitetura-e-decisoes-tecnicas.md) e [risks](risk-matrix.md). A correção descrita em design/spec é mitigação proposta, não risco já eliminado.
