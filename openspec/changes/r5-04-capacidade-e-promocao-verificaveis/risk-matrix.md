# Riscos

| Risco observado | Prioridade | Mitigação | Dono |
|---|---|---|---|
| Controle adaptativo e pools agora estão ligados a SUBMIT/STATUS/reconciliação e webhook. Contudo domínio vazio/controller nil desabilita controle; permissões externas expiradas não são recicladas automaticamente e falha de settlement apenas gera log. Scripts históricos precisaram reconciliar permits por 404 do simulador. Contagens percorrem histórico de permits por domínio a cada Acquire sob lock global do domínio. | P1 | R5-OPE-01 | Plataforma e Operações |
| Probe local do script devolveu ALLOW para prd com profile=unverified, três nomes de aprovação e isolation=PASS, sem manifesto de evidência. Script é gate, não deploy: nenhuma implantação foi feita. Validador exige formato de SHA, mas isso não vincula sozinho execução ao artefato promovido. HEAD altera bases Go/Nginx após evidências anteriores; Node build não está fixado por digest. | P0 | R5-OPE-02 | Plataforma e Operações |
| Kind independente agora inclui dependências/UI/gateway: o achado antigo de ausência deve ser encerrado nesse escopo. Overlays remotos continuam centrados nas réplicas de cinco serviços; laboratório independente não constitui IaC regional nem durabilidade/escala de dados. Recuperar dois pods não mede continuidade de negócio sob perda de nó/zona e backlog financeiro. | P1 | R5-OPE-03 | Plataforma e Operações |
