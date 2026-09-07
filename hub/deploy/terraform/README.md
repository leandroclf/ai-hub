# Terraform de referência — topologia AWS (ppd/prd)

## O que isto é, e o que NÃO é

Este diretório contém **Terraform de referência** para a topologia AWS
que o Hub de Interoperabilidade "Constelação" usaria em ppd/prd,
conforme:

- `openspec/changes/hub-interoperabilidade-v4/design.md`
- `openspec/changes/hub-interoperabilidade-v4/specs/arquitetura-e-comunicacao/spec.md`
  (ARQ-03, ARQ-05, ARQ-06)
- `openspec/changes/hub-interoperabilidade-v4/specs/decisoes-e-governanca/spec.md`
  (DEC-02, especialmente **P-01**)

**Nada aqui foi aplicado a uma conta AWS real.** `terraform apply` (e
`terraform plan` contra uma conta real) **NÃO foi executado** nesta
revisão, e não deve ser executado até que:

1. **P-01** seja aprovado — conta, região, Kubernetes gerenciado ou
   infraestrutura própria, orçamento, registro OCI e licenças. Hoje
   isso **não está decidido** (ver DEC-02, "P-01 bloqueia IaC remota e
   o gate ppd até a topologia de plataforma ser aprovada").
2. Existam credenciais AWS válidas para a conta aprovada — não há
   nenhuma configurada neste ambiente de desenvolvimento.

Todo valor de conta/região/orçamento neste diretório é
**PLACEHOLDER genérico** (ex.: `us-east-1`, conta `000000000000`,
orçamento `0`), nunca um valor real. Ver os comentários
`# PLACEHOLDER — pendente de aprovação em P-01` em `variables.tf`.

Este Terraform serve para:

- Revisão de arquitetura/sintaxe por Engenharia e Arquitetura.
- Base de discussão para quando P-01 for aprovado.
- Ensaiar `terraform init -backend=false` e `terraform validate`
  localmente (sem apply), conforme demonstrado abaixo.

## Estrutura

| Arquivo/diretório | Conteúdo |
|---|---|
| `providers.tf` | Providers (`aws`, `random`, `tls`) e bloco `backend "s3"` **comentado** (referência futura, não usado agora) |
| `variables.tf` | Variáveis, incluindo os placeholders de conta/região/orçamento |
| `main.tf` | VPC multi-AZ, EKS multi-AZ, RDS PostgreSQL (`hub_control` global + `hub_core`/`hub_finance` por célula), SNS/SQS (nomes replicados de `hub/internal/queue/queue.go` e `hub/cmd/*/main.go`), S3, KMS, Secrets Manager, papéis IAM (IRSA) de Karpenter/KEDA/Crossplane |
| `outputs.tf` | Saídas de referência (endpoints, ARNs, URLs) |
| `crossplane/` | XRD + Composition + claim de exemplo demonstrando o provisionamento declarativo de uma nova célula (ARQ-06, CFG-06, OPE-12) |

## Mapeamento para a implementação Go real

Os nomes de filas/tópicos em `main.tf` foram lidos diretamente do
código, não inventados:

| Recurso Terraform | Nome AWS | Origem no código |
|---|---|---|
| `aws_sqs_queue.this["cometa-commands"]` | `cometa-commands` | `hub/cmd/orbita/main.go`, `hub/cmd/cometa/main.go` — comando ASYNC/AUTO Órbita→Cometa (nunca usado no despacho DIRECT de SYNC) |
| `aws_sns_topic.facts["hub-operation-facts"]` | `hub-operation-facts` | `hub/cmd/cometa/main.go` — fatos de operação publicados por Cometa |
| `aws_sqs_queue.this["orbita-operation-facts"]` | `orbita-operation-facts` | `hub/cmd/orbita/main.go` — consumo de fatos de operação por Órbita |
| `aws_sqs_queue.this["libra-cost-facts"]` | `libra-cost-facts` | `hub/cmd/libra/main.go` — consumo de fatos de operação por Libra (custo) |
| `aws_sns_topic.facts["hub-protocol-facts"]` | `hub-protocol-facts` | `hub/cmd/orbita/main.go` — fatos de protocolo publicados por Órbita |
| `aws_sqs_queue.this["pulsar-protocol-facts"]` | `pulsar-protocol-facts` | `hub/cmd/pulsar/main.go` — consumo de fatos de protocolo por Pulsar (webhook) |
| `aws_sqs_queue.this["libra-revenue-facts"]` | `libra-revenue-facts` | `hub/cmd/libra/main.go` — consumo de fatos de protocolo por Libra (receita) |

Os três bancos lógicos (`hub_control`, `hub_core`, `hub_finance`) vêm
de `hub/deploy/postgres-init/01-init.sh` (DAD-01); em `main.tf`,
`hub_control` é uma única instância global (Atlas) e `hub_core`/
`hub_finance` são um par por célula (`var.cells`, hoje só `cell-01`,
já que a implementação Go local também só roda uma célula implícita).

**Diferença deliberada de segurança:** `queue.go` usa
`AllowSNSDelivery` com `Principal: "*"` como atalho de bootstrap para
o LocalStack local. Este Terraform usa uma `aws_sqs_queue_policy` com
`Principal: sns.amazonaws.com` restrito por `aws:SourceArn` ao tópico
específico — o padrão de referência para uma conta real, não o atalho
de desenvolvimento local.

## Como validar (sem aplicar nada)

Terraform não vem pré-instalado neste ambiente; o binário usado nesta
revisão foi baixado avulso (sem privilégio de root, sem tocar em
gerenciador de pacotes do sistema) só para rodar `init`/`validate`:

```bash
# (opcional, se `terraform` não estiver no PATH)
curl -fsSL -o /tmp/terraform.zip \
  https://releases.hashicorp.com/terraform/1.9.8/terraform_1.9.8_linux_amd64.zip
unzip -o -q /tmp/terraform.zip -d /tmp/tfbin
export PATH="/tmp/tfbin:$PATH"

cd hub/deploy/terraform
terraform fmt -check          # verifica formatação canônica
terraform init -backend=false # baixa apenas os providers (aws/random/tls); NÃO configura backend nem credenciais
terraform validate            # verifica sintaxe/referências, sem tocar em nenhuma API AWS
```

`terraform validate` não faz nenhuma chamada de rede à AWS e não
requer credenciais — ele só verifica que a configuração é
sintaticamente válida e internamente consistente (tipos, referências
entre recursos, blocos obrigatórios). **Nunca rode `terraform plan`
ou `terraform apply` aqui** — isso exigiria credenciais de uma conta
aprovada, que não existe (P-01 pendente).

### Resultado desta revisão

```
terraform fmt -check   → sem diferenças (arquivos já formatados)
terraform init -backend=false → sucesso (providers hashicorp/aws 5.100.0,
                                  hashicorp/random 3.9.0, hashicorp/tls 4.4.0)
terraform validate     → "Success! The configuration is valid." (sem warnings)
```

O arquivo `.terraform.lock.hcl` gerado por este `init` está commitado
(prática recomendada); o diretório `.terraform/` (cache de plugins) é
ignorado via `.gitignore` deste diretório.

## Módulo Crossplane (`crossplane/`)

`crossplane/xrd.yaml` e `crossplane/composition.yaml` demonstram, em
YAML simplificado (não production-grade), como uma **nova célula**
seria provisionada declarativamente pelo plano de gestão do Crossplane
a partir do placement/demanda registrado por Atlas (ARQ-06, CFG-06,
OPE-12): banco `hub_core` + `hub_finance`, filas/tópicos e papel IAM
por célula. `crossplane/claim-example.yaml` mostra o pedido
(`HubCell`) que a Plataforma/Atlas emitiria para uma célula `cell-02`
hipotética.

Este módulo nunca foi aplicado a um cluster Kubernetes/Crossplane real
— não há cluster nem provider Crossplane instalado neste ambiente. Os
arquivos foram validados apenas como YAML sintaticamente correto
(`yaml.safe_load_all` via Python), não como CRDs aceitas por um
`kubectl apply --dry-run` contra um cluster com os CRDs do Crossplane
e do provider AWS instalados (o que exigiria um cluster fora do
escopo desta tarefa). Ver o cabeçalho de cada arquivo para as
simplificações assumidas (nomes de fila/tópico por célula são
ilustrativos; a implementação Go atual não os namespacea por célula).

## Regras que este Terraform respeita deliberadamente

- Nenhuma conta/região/ID real da AWS aparece em nenhum arquivo.
- Nenhum `terraform apply`/`plan` foi executado contra uma conta real.
- RDS: `multi_az = true`, `backup_retention_period >= 35` (DAD-07),
  `storage_encrypted = true` via KMS, `deletion_protection = true` e
  `prevent_destroy` no lifecycle (ARQ-06: bancos protegidos contra
  exclusão automática).
- S3: versionamento habilitado, bloqueio total de acesso público,
  criptografia SSE-KMS, `prevent_destroy`.
- Segredos: apenas referências (`aws_secretsmanager_secret`); o valor
  em si é gerado por `random_password` no momento do (futuro) apply,
  nunca hardcoded neste repositório.
- SQS/SNS: nomes exatos replicados do código Go real, política de fila
  restrita por `aws:SourceArn` (não o atalho `Principal: "*"` do
  bootstrap local).
