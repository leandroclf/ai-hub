# Variáveis da topologia de referência AWS do Hub (ppd/prd).
#
# Todas as variáveis marcadas "PLACEHOLDER — pendente de aprovação em
# P-01" usam valores de exemplo genéricos, nunca conta/região/ID real.
# Ver openspec/changes/hub-interoperabilidade-v4/specs/decisoes-e-governanca/spec.md
# (DEC-02, P-01) e specs/arquitetura-e-comunicacao/spec.md (ARQ-03).

variable "aws_region" {
  description = "Região AWS de referência. PLACEHOLDER — pendente de aprovação em P-01 (conta/região/orçamento ainda não aprovados)."
  type        = string
  default     = "us-east-1" # PLACEHOLDER — pendente de aprovação em P-01
}

variable "aws_account_id" {
  description = "Conta AWS de destino. PLACEHOLDER — pendente de aprovação em P-01. Não preencher com conta real antes da aprovação formal."
  type        = string
  default     = "000000000000" # PLACEHOLDER — pendente de aprovação em P-01
}

variable "monthly_budget_usd" {
  description = "Orçamento mensal de referência (USD) usado apenas para dimensionar o exemplo. PLACEHOLDER — pendente de aprovação em P-01; não é um compromisso financeiro."
  type        = number
  default     = 0 # PLACEHOLDER — pendente de aprovação em P-01 (sem orçamento aprovado)
}

variable "environment" {
  description = "Ambiente de implantação (OPE-04: local/dev/hom/ppd/prd). Este Terraform é referência para ppd/prd; não há promoção real enquanto P-01 estiver pendente."
  type        = string
  default     = "ppd"

  validation {
    condition     = contains(["dev", "hom", "ppd", "prd"], var.environment)
    error_message = "environment deve ser um de: dev, hom, ppd, prd (local não usa esta topologia AWS)."
  }
}

variable "project_name" {
  description = "Prefixo de nomenclatura dos recursos (não é nome de conta/organização real)."
  type        = string
  default     = "hub-constelacao"
}

variable "vpc_cidr" {
  description = "Bloco CIDR da VPC de referência."
  type        = string
  default     = "10.42.0.0/16"
}

variable "az_count" {
  description = "Quantidade de zonas de disponibilidade usadas pela topologia multi-AZ (OPE-04 exige ao menos 3)."
  type        = number
  default     = 3

  validation {
    condition     = var.az_count >= 3
    error_message = "az_count deve ser >= 3 (multi-AZ, conforme OPE-04 e ARQ-03/ARQ-06)."
  }
}

# ---------------------------------------------------------------------------
# Células (ARQ-01/ARQ-02/DAD-01/DAD-11): cada célula tem seu próprio par de
# bancos hub_core/hub_finance. hub_control é único e global (Atlas), fora do
# conceito de célula. A implementação Go atual (hub/cmd/*/main.go) só roda
# uma célula implícita localmente; aqui modelamos "cell-01" explicitamente
# para refletir a topologia de referência da v4 e alimentar o módulo
# Crossplane de exemplo (../crossplane).
# ---------------------------------------------------------------------------
variable "cells" {
  description = "Lista de células de negócio, cada uma com bancos hub_core e hub_finance próprios (DAD-01/ADR-03). Ao menos uma célula de referência ('cell-01') é modelada nesta revisão."
  type        = list(string)
  default     = ["cell-01"]

  validation {
    condition     = length(var.cells) >= 1
    error_message = "cells deve conter ao menos uma célula (ex.: \"cell-01\")."
  }
}

# ---------------------------------------------------------------------------
# PostgreSQL (ARQ-03/ARQ-05, DAD-07: backups automáticos + PITR de 35 dias,
# multi-AZ)
# ---------------------------------------------------------------------------
variable "db_engine_version" {
  description = "Versão do engine PostgreSQL gerenciado (RDS). Versão suportada fixada conforme ARQ-05 (exige plano de atualização antes de produção)."
  type        = string
  default     = "16.4"
}

variable "db_instance_class" {
  description = "Classe de instância RDS de referência (dimensionamento real depende de P-01/orçamento aprovado)."
  type        = string
  default     = "db.r6g.large"
}

variable "db_backup_retention_days" {
  description = "Retenção de backup automático / janela de PITR (DAD-07: 35 dias)."
  type        = number
  default     = 35

  validation {
    condition     = var.db_backup_retention_days >= 35
    error_message = "db_backup_retention_days deve ser >= 35 (DAD-07: PITR de 35 dias)."
  }
}

variable "db_allocated_storage_gb" {
  description = "Armazenamento inicial (GB) de cada banco RDS de referência."
  type        = number
  default     = 100
}

variable "db_max_allocated_storage_gb" {
  description = "Teto de autoscaling de armazenamento (GB) de cada banco RDS de referência."
  type        = number
  default     = 500
}

# ---------------------------------------------------------------------------
# EKS (ARQ-03: Kubernetes, referência EKS multi-AZ)
# ---------------------------------------------------------------------------
variable "eks_cluster_version" {
  description = "Versão do control plane EKS. Versão suportada fixada conforme ARQ-05."
  type        = string
  default     = "1.30"
}

variable "eks_node_instance_types" {
  description = "Tipos de instância do node group gerenciado de base (capacidade mínima; expansão real via Karpenter/NodePools, fora deste Terraform)."
  type        = list(string)
  default     = ["m6i.large"]
}

variable "eks_node_group_desired_size" {
  description = "Tamanho desejado do node group gerenciado de base (piso; Karpenter cobre picos, ARQ-06)."
  type        = number
  default     = 3
}

variable "eks_node_group_min_size" {
  type    = number
  default = 3
}

variable "eks_node_group_max_size" {
  type    = number
  default = 6
}

# ---------------------------------------------------------------------------
# S3 (DAD-05)
# ---------------------------------------------------------------------------
variable "objects_bucket_force_destroy" {
  description = "Nunca deve ser true em ppd/prd (ARQ-06: proteção contra exclusão automática). Existe apenas para permitir destruição de um bucket de teste isolado, se necessário."
  type        = bool
  default     = false
}
