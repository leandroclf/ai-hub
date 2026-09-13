# Explore — Capacidade, ambientes e promoção verificáveis

Snapshot: f87ce33034ae29c9431b1910dcc6a633b545e330. Fontes lidas; hipóteses de concorrência são estáticas até ensaio.

## F-R5-11 (P1)
Controle adaptativo e pools agora estão ligados a SUBMIT/STATUS/reconciliação e webhook. Contudo domínio vazio/controller nil desabilita controle; permissões externas expiradas não são recicladas automaticamente e falha de settlement apenas gera log. Scripts históricos precisaram reconciliar permits por 404 do simulador. Contagens percorrem histórico de permits por domínio a cada Acquire sob lock global do domínio.

[hub/internal/cometa/executor.go:72](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/executor.go#L72), [hub/internal/cometa/capacity.go:250](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/capacity.go#L250), [hub/internal/cometa/capacity.go:346](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/capacity.go#L346), [hub/internal/cometa/executor.go:120](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/internal/cometa/executor.go#L120)

## F-R5-12 (P0)
Probe local do script devolveu ALLOW para prd com profile=unverified, três nomes de aprovação e isolation=PASS, sem manifesto de evidência. Script é gate, não deploy: nenhuma implantação foi feita. Validador exige formato de SHA, mas isso não vincula sozinho execução ao artefato promovido. HEAD altera bases Go/Nginx após evidências anteriores; Node build não está fixado por digest.

[hub/deploy/r2/tests/promotion-gate.sh:22](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/promotion-gate.sh#L22), [hub/deploy/r2/tests/validate-qualification-evidence.py:53](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/tests/validate-qualification-evidence.py#L53), [hub/deploy/Dockerfile:5](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/Dockerfile#L5), [hub/deploy/r2/Dockerfile.ui:1](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/Dockerfile.ui#L1)

## F-R5-13 (P1)
Kind independente agora inclui dependências/UI/gateway: o achado antigo de ausência deve ser encerrado nesse escopo. Overlays remotos continuam centrados nas réplicas de cinco serviços; laboratório independente não constitui IaC regional nem durabilidade/escala de dados. Recuperar dois pods não mede continuidade de negócio sob perda de nó/zona e backlog financeiro.

[hub/deploy/r2/kind/render-independent-dependencies.py:1](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/kind/render-independent-dependencies.py#L1), [hub/deploy/r2/kind/bootstrap-independent.sh:1](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/kind/bootstrap-independent.sh#L1), [hub/deploy/r2/k8s/overlays/prd/kustomization.yaml:6](https://github.com/leandroclf/ai-hub/blob/f87ce33034ae29c9431b1910dcc6a633b545e330/hub/deploy/r2/k8s/overlays/prd/kustomization.yaml#L6)
