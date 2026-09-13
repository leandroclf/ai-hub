# Explore: Qualificação operacional e promoção

Snapshot a540b40007fe6b8ed523e17afe00e96ff8f8ad50. Leitura direta do delta R5→R6.

## F-R6-09 · P0 · PROBE_REPRODUZIDO
A ausência de manifesto agora bloqueia prd. Probe novo passou manifesto de um cenário, SHA de quarenta zeros, status PASS_WITHOUT_EXECUTION, command not executed e oráculos inventados iguais: o gate respondeu ALLOW. O validador exige forma e presença do ID no arquivo, não a identidade do candidato, integridade, cobertura obrigatória ou proveniência confiável.

Fontes: [hub/deploy/r2/tests/promotion-gate.sh:22](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/promotion-gate.sh#L22), [hub/deploy/r2/tests/validate-qualification-evidence.py:71](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/validate-qualification-evidence.py#L71), [hub/deploy/r2/tests/validate-qualification-evidence.py:81](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/validate-qualification-evidence.py#L81)

Vínculo anterior: R5-OPE-02.

## F-R6-10 · P0 · ANALISE_ESTATICA
O script de restore não mudou nesta rodada: compara lista parcial de tabelas, faz s3 sync/list-objects-v2 e consulta o oráculo somente antes de declarar PASS sem replay. Planos, etapas, novos settlements e versões históricas de objetos não têm cobertura explícita. O relatório da implementação reconhece o item não iniciado.

Fontes: [hub/deploy/r2/tests/restore-reconciliation.sh:21](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L21), [hub/deploy/r2/tests/restore-reconciliation.sh:57](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L57), [hub/deploy/r2/tests/restore-reconciliation.sh:63](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/tests/restore-reconciliation.sh#L63)

Vínculo anterior: R5-DAD-02.

## F-R6-11 · P1 · ANALISE_ESTATICA
Kind independente já existe e EnsureTopology foi ligado aos quatro processos antes dos consumidores/relays: preservar. Não houve evolução de deploy além do gate nesta rodada. O gerador de dependências kind declara volumes efêmeros; isso é adequado ao laboratório, mas não prova recuperação ou ambientes remotos duráveis. Bootstrap único também não qualifica remoção posterior de assinatura SNS.

Fontes: [hub/deploy/r2/kind/render-independent-dependencies.py:4](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/deploy/r2/kind/render-independent-dependencies.py#L4), [hub/internal/queue/queue.go:228](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/queue/queue.go#L228), [hub/cmd/orbita/main.go:68](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/cmd/orbita/main.go#L68)

Vínculo anterior: R5-OPE-03;R5-EXE-05.

## F-R6-12 · P1 · ANALISE_ESTATICA
O executor passou a recusar controller nil quando domínio existe. Domínio vazio ainda retorna capacidade desabilitada. Há resolução por evidência, porém settle com lease vencido pode falhar e falhas de resolução são logadas; contagens sob lock percorrem histórico do domínio. Não afirmar ausência do controlador adaptativo, que já está implementado.

Fontes: [hub/internal/cometa/executor.go:74](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L74), [hub/internal/cometa/executor.go:134](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/executor.go#L134), [hub/internal/cometa/capacity.go:202](https://github.com/leandroclf/ai-hub/blob/a540b40007fe6b8ed523e17afe00e96ff8f8ad50/hub/internal/cometa/capacity.go#L202)

Vínculo anterior: R5-OPE-01.
