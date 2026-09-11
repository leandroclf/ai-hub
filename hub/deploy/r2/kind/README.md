# Laboratório Kind do AI Hub

Este diretório contém dois perfis locais para o AI Hub R2/R4:

- `bootstrap.sh`: perfil de compatibilidade que reutiliza as dependências do
  Compose oficial quando o cenário precisa compartilhar a fixture existente;
- `bootstrap-independent.sh`: perfil independente, com Postgres, LocalStack,
  Keycloak, observabilidade, gateway, provider-sim, webhook-sink e UI
  materializados dentro do Kind.

## Perfil independente

Pré-requisitos: Docker, Python 3, `kubectl`, acesso ao binário `kind` e o
Compose oficial do AI Hub já qualificado ou parado conforme a política do
repositório. O bootstrap é offline para imagens: quando a imagem de terceiros
não está disponível no formato aceito pelo containerd do Kind, ele a
reempacota localmente e a carrega nos três nós. Antes de aplicar os workloads,
ele instala as versões fixadas do metrics-server e do KEDA e aguarda o CRD
`ScaledObject`; assim, o perfil não depende de um cluster previamente preparado
para aceitar o autoscaling do Pulsar.

```bash
cd hub
bash deploy/r2/kind/bootstrap-independent.sh
KUBECONFIG=/tmp/ai-hub-r2-tools/kubeconfig \
  bash deploy/r2/tests/continuity-runtime-proof.sh
```

O segundo comando verifica nós Ready, cinco workloads de negócio, dependências
cluster-owned, PDB/HPA/KEDA, distribuição em mais de um nó e a recuperação
controlada de um pod de Cometa e de um pod de Pulsar. O gate imprime o RTO
observado de cada recuperação e encerra com `CONTINUITY_RUNTIME_PROOF=PASS`.

Para inspeção manual, use portas alternativas quando o Compose oficial ocupar
as portas padrão:

```bash
kubectl --kubeconfig /tmp/ai-hub-r2-tools/kubeconfig \
  -n ai-hub-local-kind port-forward svc/admin-ui 23000:8080
kubectl --kubeconfig /tmp/ai-hub-r2-tools/kubeconfig \
  -n ai-hub-local-kind port-forward svc/identity 28085:8080
kubectl --kubeconfig /tmp/ai-hub-r2-tools/kubeconfig \
  -n ai-hub-local-kind port-forward svc/kong 28000:8000
```

O perfil é um laboratório: os dados usam volumes efêmeros, a imagem do banco é
carregada localmente e não há promessa de HA regional, backup/PITR, KMS,
Secrets Manager ou IaC remoto. Para continuidade de dados, a evidência
oficial continua sendo o restore reconciliado em
`hub/deploy/r2/tests/restore-reconciliation.sh`; a prova Kind cobre
recuperação de workloads e isolamento de dependências, não substitui o perfil
operacional contratado.

Ao trocar entre perfis, siga a regra do `AGENTS.md`: inspecione
`docker compose ls`/`docker ps -a`, encerre o ecossistema AI Hub anterior,
confirme os containers e preserve volumes por padrão. Não use `docker system
prune`.
