# Risk matrix

| Risco | Prioridade | Mitigação | Prova | Responsável |
|---|---|---|---|---|
| FileRefs/pins/upload existem; Execute não consome FileRefs e persistência de resultado volumoso não está conectada à execução real. | P1 | R3-OPE-01 | R3-OPE-01-S01/S02/S03 | Plataforma, Dados e SRE |
| Kubernetes contém cinco serviços de negócio; local-kind aponta dependências a IPs de containers Compose. Overlays remotos não materializam sozinhos toda configuração/dependências. | P1 | R3-OPE-02 | R3-OPE-02-S01/S02/S03 | Plataforma, Dados e SRE |
| Probes/réplicas/autoscalers são avanços declarativos, ainda sem comprovação de escala de nós/dados, placement automático e drenagem integrada. | P1 | R3-OPE-03 | R3-OPE-03-S01/S02/S03 | Plataforma, Dados e SRE |
| Custódia transacional avançou; RLS com papel real não proprietário e restore reconciliado continuam sem qualificação integrada. Fallback não pode transformar aceite em memória volátil. | P0 | R3-OPE-04 | R3-OPE-04-S01/S02/S03 | Plataforma, Dados e SRE |
| Telemetria HTTP/manifests existem, mas API de SLA não funciona no painel e alertas de obrigação exigem produtor real. Render não prova scrape/log/trace/alerta. | P1 | R3-OPE-05 | R3-OPE-05-S01/S02/S03 | Plataforma, Dados e SRE |

P0 bloqueia qualificação de uso; P1 bloqueia conclusão do escopo. Probabilidade quantitativa não foi medida.
