# Terraform de REFERÊNCIA para a topologia AWS de ppd/prd do Hub de
# Interoperabilidade "Constelação" (ARQ-03/ARQ-05/ARQ-06).
#
# IMPORTANTE (ver README.md deste diretório e DEC-02/P-01):
#   - Nenhum destes recursos foi aplicado a uma conta AWS real.
#   - Não há conta, região ou orçamento aprovados (P-01 continua
#     pendente); todo valor de conta/região/orçamento aqui é
#     PLACEHOLDER genérico, sujeito a aprovação.
#   - Este diretório existe para revisão de sintaxe/arquitetura
#     (terraform init -backend=false + terraform validate), nunca
#     para terraform plan/apply contra uma conta real.

terraform {
  required_version = ">= 1.7.0, < 2.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  # Backend remoto (S3 + DynamoDB/lock) é o alvo de referência para
  # ppd/prd, mas SHALL NOT ser configurado/usado antes de P-01 aprovar
  # conta/região/orçamento. Por isso este bloco permanece comentado —
  # o estado local só existe para permitir `terraform init
  # -backend=false` e `terraform validate` neste ambiente de revisão.
  #
  # backend "s3" {
  #   bucket         = "PLACEHOLDER-hub-terraform-state"      # pendente P-01
  #   key            = "hub/constelacao/terraform.tfstate"
  #   region         = "us-east-1"                             # pendente P-01
  #   dynamodb_table = "PLACEHOLDER-hub-terraform-locks"       # pendente P-01
  #   encrypt        = true
  # }
}

# NENHUMA credencial AWS é configurada aqui. Em ppd/prd, o provider
# resolveria credenciais via identidade de workload (OIDC/IRSA de
# pipeline), nunca chave estática em arquivo. Neste ambiente de
# referência não existe sequer uma conta para autenticar.
provider "aws" {
  region = var.aws_region # PLACEHOLDER — pendente de aprovação em P-01

  default_tags {
    tags = {
      Project     = "hub-constelacao"
      ManagedBy   = "terraform"
      Environment = var.environment
      Reference   = "openspec/changes/hub-interoperabilidade-v4"
      Pending     = "P-01"
    }
  }
}

provider "random" {}

provider "tls" {}
