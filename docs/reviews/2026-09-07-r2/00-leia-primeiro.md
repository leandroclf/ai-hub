# AI Hub — Pacote OpenSpec de evolução R2

**Leitura principal:** [relatório da revisão](01-relatorio-de-revisao.md). **Snapshot:** `a39d394b0d87185ed4cc3861c12ec45f2c302d9e`. **Conclusão:** 2026-09-08. Este pacote contém propostas documentais para a próxima implementação; não contém mudanças de aplicação ou infraestrutura executadas.

São 42 achados, 9 changes, 66 requisitos complementares e 205 cenários, ligados aos 98 requisitos da baseline v4. Os principais riscos estão no núcleo de custódia, autenticação, resposta, concorrência, deadline e financeiro; o console precisa evoluir por jornadas completas, não só aparência.

## Como usar

1. Extrair o ZIP e adicionar os diretórios `openspec/changes/r2-*` e `docs/reviews/2026-09-07-r2` ao repositório, preservando os arquivos existentes. `LEIA_PRIMEIRO_R2.md` é uma entrada opcional na raiz.
2. Comparar HEAD ao SHA revisado. Se mudou, revalidar apenas os achados/cenários afetados, sem descartar a baseline.
3. Ler relatório, matriz e roteiro; escolher a change pela dependência e prioridade. Cada pasta tem explore/proposal/design/spec/tasks/risks/traceability/evaluator.
4. Enviar [o prompt de implementação](09-prompt-para-proxima-implementacao.md) junto dos artefatos. Não marcar tasks concluídas antes de prova.
5. Qualificar contratos/dados/segurança, integração e operação conforme perfil; somente depois promover/ativar.

## As nove changes

| Change | Responsável funcional | Requisitos / cenários | Dependências de base |
|---|---|---|---|
| [r2-01-identidade-e-isolamento](../../../openspec/changes/r2-01-identidade-e-isolamento/proposal.md) | Segurança e Engenharia de Plataforma | 5 / 15 | Pode iniciar já |
| [r2-02-execucao-duravel-e-resultados](../../../openspec/changes/r2-02-execucao-duravel-e-resultados/proposal.md) | Engenharia de Core e Dados | 9 / 29 | 01 |
| [r2-03-integracoes-credenciais-e-pressao](../../../openspec/changes/r2-03-integracoes-credenciais-e-pressao/proposal.md) | Integrações e Engenharia de Core | 8 / 28 | 01, 02 |
| [r2-04-catalogo-produtos-e-contratos](../../../openspec/changes/r2-04-catalogo-produtos-e-contratos/proposal.md) | Produto, Atlas e Engenharia de Core | 8 / 24 | 01, 02 |
| [r2-05-console-administrativo](../../../openspec/changes/r2-05-console-administrativo/proposal.md) | Produto, Frontend e UX | 12 / 36 | 01 |
| [r2-06-financeiro-auditavel](../../../openspec/changes/r2-06-financeiro-auditavel/proposal.md) | Financeiro, Comercial e Engenharia Libra | 6 / 19 | 01, 02, 04 |
| [r2-07-objetos-dados-e-retencao](../../../openspec/changes/r2-07-objetos-dados-e-retencao/proposal.md) | Dados, Segurança e Plataforma | 5 / 15 | 01, 02 |
| [r2-08-operacao-elastica-e-observavel](../../../openspec/changes/r2-08-operacao-elastica-e-observavel/proposal.md) | Plataforma, SRE e Engenharia | 9 / 27 | 01, 02 |
| [r2-09-qualificacao-e-rastreabilidade](../../../openspec/changes/r2-09-qualificacao-e-rastreabilidade/proposal.md) | Qualidade, Liderança técnica e SRE | 4 / 12 | Pode iniciar já |

A change 09 inicia com fixtures/gates e fecha qualificação depois das demais. A change 05 pode começar sessão/listas após 01, mas suas jornadas dependem das APIs por domínio; ver roteiro. Dependências de interface não autorizam concluir telas com mocks.

## Documentos complementares

- [03-arquitetura-e-decisoes-tecnicas.md](03-arquitetura-e-decisoes-tecnicas.md)
- [04-console-jornadas-e-api.md](04-console-jornadas-e-api.md)
- [05-contratos-dados-e-estados.md](05-contratos-dados-e-estados.md)
- [06-operacao-escala-e-observabilidade.md](06-operacao-escala-e-observabilidade.md)
- [07-qualidade-e-cenarios-integrados.md](07-qualidade-e-cenarios-integrados.md)
- [08-pendencias-migracao-e-roteiro.md](08-pendencias-migracao-e-roteiro.md)
- [09-prompt-para-proxima-implementacao.md](09-prompt-para-proxima-implementacao.md)
- [11-fontes-e-metodo.md](11-fontes-e-metodo.md)
- [02-rastreabilidade-98.csv](02-rastreabilidade-98.csv): cada ID baseline ligado aos incrementos.
- [achados-42.csv](achados-42.csv): backlog de riscos com evidências.
- [cenarios-r2.csv](cenarios-r2.csv): 205 casos para planejamento/execução de QA.
- [10-validacao-do-pacote.md](10-validacao-do-pacote.md): validações realmente realizadas no pacote.
- [manifesto.json](manifesto.json): contagens, SHA e arquivos gerados; [checksums.sha256](checksums.sha256): integridade dos arquivos.

## Compatibilidade com OpenSpec existente

O repo mantém 98 requisitos em uma change v4 aberta e não tem specs canônicas. R2 usa **ADDED Requirements em capabilities de escopo próprio**, com IDs R2, para evitar MODIFIED contra uma base canônica inexistente e colisão de títulos. Não remove/reescreve a v4 nem a arquiva. O termo “complementar” refina aceite e fecha gaps; não substitui regras mais amplas dos 98 IDs. IDs normativos e tarefas abertas são preservados.

Consolidar árvore canônica é T-R2-05 / R2-QUA-04: fazê-lo com rastreabilidade e aprovação de especificação, sem confundir com software implementado. Validar changes é diferente de certificar que seus cenários já foram executados.
